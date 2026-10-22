import { createReferenceClaim } from './useClaimScenario.js'
import { DEMO_CUSTOMER_ID, DEMO_POLICY_ID, DEMO_CLAIM_ID, DEMO_CLAIM_REFERENCE } from '../config/demo.js'

export async function hotlineRequest(action, body = {}, options = {}) {
  const response = await fetch(`/api/hotline/${action}`, { method: 'POST', credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', 'X-Hotline-Request': '1' }, body: JSON.stringify(body), ...options })
  const result = await response.json()
  if (!response.ok || !result.ok) throw Object.assign(new Error(result.code || 'HOTLINE_FAILED'), { code: result.code || 'HOTLINE_FAILED' })
  return result
}
const stages = ['collecting_documents', 'documents_complete', 'coverage_proposal', 'assistance_arranged', 'claim_closed']
const needLabels = { accommodation: 'Temporary accommodation', essential_items: 'Essential items', electrical_safety: 'Electrical safety', cleaning_drying: 'Cleaning and drying', moving_storage: 'Moving and storage', contents: 'Damaged contents', building: 'Insured building repairs' }

export function useHotlineClaim() {
  const { state } = useDemo()
  const source = useState('hotline-mode', () => 'live')
  const claimError = useState('hotline-claim-error', () => '')
  async function refreshLiveClaim({ quiet = false } = {}) {
    if (source.value !== 'live') return
    const revision = state.value.revision
    try {
      const { claim, readGrantExpiresAt } = await hotlineRequest('claim')
      if (source.value !== 'live' || state.value.revision !== revision) return
      if (claim.id !== DEMO_CLAIM_ID || claim.policy_id !== DEMO_POLICY_ID || claim.reference !== DEMO_CLAIM_REFERENCE || claim.intake_status !== 'reported') return
      const previous = state.value.claim
      const hasAccommodation = claim.intake_details?.accommodationNeeded === true
      state.value.claim = { ...createReferenceClaim(), id: DEMO_CLAIM_REFERENCE, databaseId: DEMO_CLAIM_ID, policyId: DEMO_POLICY_ID, customerId: DEMO_CUSTOMER_ID,
        contractNumber: claim.contract_number, productName: claim.product_name, name: claim.first_name, initials: 'J', address: claim.address, event: 'Flooding', date: '2026-10-20', eventDate: '2026-10-20', reportedAt: '2026-10-22',
        stage: Math.max(0, stages.indexOf(claim.stage)), summary: claim.summary || '', summaryIsUserContent: true,
        needs: claim.needs.filter(need => need.status === 'identified').map(need => needLabels[need.code]).filter(Boolean),
        householdSize: claim.household_size, declaredHouseholdSize: claim.intake_details?.householdSize, currentAccommodation: claim.intake_details?.currentAccommodation || '', currentAccommodationIsUserContent: true,
        intakeDetails: claim.intake_details, priorityReason: claim.priority_reasons[0] || 'Situation requires review.', opened: true, priority: claim.priority || 'P2',
        access: ({ unknown: 'Unknown', restricted: 'Restricted', permitted_in_demo: 'Permitted' })[claim.access_status],
        situation: ({ habitable: 'Home habitable', uninhabitable: 'Home uninhabitable', unknown: 'Situation unconfirmed' })[claim.habitability],
        accommodationNeeded: hasAccommodation, coverage: claim.assessment?.status === 'human_reviewed_proposal' ? 'Proposal' : 'Not reviewed', catNat: 'Not verified',
        requiredDocuments: claim.required_documents, source: 'postgresql', readGrantExpiresAt, updatedAt: claim.updated_at }
      if (!previous || previous.source !== 'postgresql') {
        // Discard scripted local previews. Live evidence is fetched through its scoped API.
        state.value.files.forEach(file => URL.revokeObjectURL(file.url))
        state.value.files = []; state.value.messages = []
      }
      claimError.value = ''
    } catch (error) {
      if (!quiet || state.value.claim?.source === 'postgresql') claimError.value = error.code || 'HOTLINE_FAILED'
    }
  }
  return { source, claimError, refreshLiveClaim }
}
