import { demoBenefits, claimStages } from './useClaimScenario.js'

export const insurerServiceLabels = { accommodation: 'Temporary accommodation', electrical_safety: 'Electrical safety', cleaning_drying: 'Cleaning and drying', moving_storage: 'Moving and storage', damage_assessment: 'Damage assessment', essential_items: 'Essential items', contents: 'Damaged contents', building: 'Insured building repairs' }
const codes = { accommodation: 'accommodation', essentials: 'essential_items', electrical: 'electrical_safety', drying: 'cleaning_drying', storage: 'moving_storage', contents: 'contents', building: 'building' }
export const insurerAgentInitial = () => ({ authorized: false, live: false, claims: [], providers: [], assignments: [], proposals: {} })
const errors = {
  INSURER_AGENT_DISABLED: 'Enable the insurer agent in the server configuration.', INSURER_NOT_CONFIGURED: 'Configure the private insurer access key on the server.',
  INSURER_ACCESS_DENIED: 'Unlock insurer access in Settings to continue.', OPENAI_NOT_CONFIGURED: 'Configure the existing OpenAI server key before using the insurer agent.',
  OPENAI_RATE_LIMITED: 'The AI service is busy. Try again later.', INSURER_AI_UNAVAILABLE: 'The insurer AI service is unavailable. No substitute proposal was generated.',
  INSURER_TOOL_CONTEXT_MISSING: 'The agent did not read all required sources. No proposal was saved.', INSURER_TOOL_LIMIT: 'The analysis reached its tool limit. No proposal was saved.',
  INVALID_PROPOSAL: 'The proposal failed the server checks. No decision was applied.', PROPOSAL_STALE: 'The dossier or resources changed. Run a new analysis before approving.',
  INSURER_REPORT_REQUIRED: 'Save the client report before requesting an insurer decision.', POLICY_PDF_UNAVAILABLE: 'The current policy PDF is missing. Prepare the stored policy documents first.',
  PROPOSAL_ALREADY_REVIEWED: 'This proposal has already been reviewed.', CLAIM_CLOSED: 'This dossier is closed. It cannot receive a new decision.',
  DATABASE_UNAVAILABLE: 'The insurer database operation could not be confirmed.', EVIDENCE_CONSENT_REQUIRED: 'Agree to send the dossier and requested files to OpenAI before analysis.'
}
export function projectInsurerClaim(row) {
  const group = row.assessment.group || row.assessment.scenario_group || (row.habitability === 'uninhabitable' ? 'A' : row.habitability === 'habitable' ? 'C' : row.access_status === 'restricted' ? 'B' : 'D')
  const stage = Math.max(0, ['collecting_documents', 'documents_complete', 'coverage_proposal', 'assistance_arranged', 'claim_closed'].indexOf(row.stage))
  const eventDate = new Date(row.incident_at).toISOString().slice(0, 10)
  return { id: row.reference, databaseId: row.id, policyId: row.policy_id, name: row.first_name, nameParameters: {}, initials: row.first_name.slice(0, 1),
    householdSize: row.household_size, address: row.address, municipality: row.city, cluster: row.city, coordinates: [row.longitude, row.latitude],
    addressNote: 'Approximate household location.', contractNumber: row.contract_number, productName: row.product_name, group,
    situation: ({ A: 'Home uninhabitable', B: 'Evacuated or inaccessible', C: 'Damaged but habitable', D: 'Limited or unconfirmed' })[group],
    stage, stageLabel: claimStages[stage], event: 'Flooding', eventDate, date: eventDate, reportedAt: row.reported_at,
    opened: row.intake_status === 'reported', priority: row.priority || 'P3', priorityReason: row.priority_reasons[0] || 'Situation requires review.',
    summary: row.assessment.summary || row.summary || '', summaryIsUserContent: true,
    needs: row.needs.filter(n => n.status === 'identified').map(n => insurerServiceLabels[n.code]).filter(Boolean),
    accommodationNeeded: row.needs.some(n => n.code === 'accommodation' && n.status === 'identified'), accommodationPlaced: false,
    currentAccommodation: 'Accommodation arrangements require confirmation.',
    access: ({ unknown: 'Unknown', restricted: 'Restricted', permitted_in_demo: 'Permitted' })[row.access_status],
    coverage: row.assessment.status === 'human_reviewed_proposal' ? 'Proposal' : 'Not reviewed', catNat: 'Not verified',
    waitingHours: row.reported_at ? Math.max(0, Math.floor((Date.parse('2026-10-22T12:00:00+02:00') - Date.parse(row.reported_at)) / 3600000)) : 0,
    moneyLines: demoBenefits.map(benefit => {
      const line = row.cost_estimate.lines?.find(l => l.code === codes[benefit.id])
      return { ...benefit, amount: line?.estimated_cents == null ? null : line.estimated_cents / 100,
        limit: (row.policy_coverage_data.benefits.find(b => b.code === codes[benefit.id])?.limit_cents || 0) / 100,
        approved: (line?.approved_cents || 0) / 100, paid: (line?.paid_cents || 0) / 100, review: 'Not reviewed' }
    }), source: 'postgresql', chronology: [...(row.reported_at ? [{ label: 'Report received', at: row.reported_at }] : []),
      ...(row.assessment.human_reviewed_at ? [{ label: 'Human review recorded', at: row.assessment.human_reviewed_at }] : [])] }
}
export function projectInsurerProviders(data) {
  return data.providers.map(p => {
    const used = data.assignments.filter(a => a.provider_id === p.id && ['approved_in_demo', 'in_progress_in_demo'].includes(a.status)
      && Date.parse(a.starts_at) < Date.parse(p.availability.ends_at) && Date.parse(a.ends_at) > Date.parse(p.availability.starts_at)).length
    return { id: p.id, name: p.name, skill: insurerServiceLabels[p.service_types[0]], capacity: p.capacity, used, remaining: Math.max(0, p.capacity - used),
      available: p.availability.status === 'available', status: p.availability.status !== 'available' ? 'Unavailable' : used ? 'Active' : 'Available',
      unit: p.pricing.unit === 'household_night' ? 'Household accommodation slots' : p.service_types[0] === 'cleaning_drying' ? 'Concurrent sites' : 'Visits per day', coordinates: [p.longitude, p.latitude] }
  })
}
export function useInsurerAgent() {
  const data = useState('insurer-agent-data', insurerAgentInitial)
  const busy = useState('insurer-agent-busy', () => ''), error = useState('insurer-agent-error', () => '')
  const accessKey = ref('')
  const { locale } = useAppI18n()
  const phrase = failure => errors[failure?.code] || 'The insurer operation could not be confirmed. Refresh before retrying.'
  async function request(action, body = {}, binary = false) {
    const response = await fetch(`/api/insurer/${action}`, { method: 'POST', credentials: 'same-origin', headers: { 'Content-Type': 'application/json', 'X-Insurer-Request': '1' }, body: JSON.stringify({ ...body, locale: locale.value }) })
    if (binary && response.ok && !response.headers.get('content-type')?.includes('application/json')) return response.blob()
    const result = await response.json()
    if (!response.ok || !result.ok) throw Object.assign(new Error(), { code: result.code })
    return result
  }
  async function refresh() {
    try { const result = await request('portfolio'); Object.assign(data.value, { claims: result.claims, providers: result.providers, assignments: result.assignments }); return true }
    catch (failure) { error.value = phrase(failure); return false }
  }
  async function restore() {
    try { await request('access'); data.value.authorized = true; if (await refresh()) data.value.live = true }
    catch { data.value.authorized = false }
  }
  async function unlock() {
    if (busy.value) return
    busy.value = 'unlock'; error.value = ''
    try { await request('authorize', { token: accessKey.value }); data.value.authorized = true; if (await refresh()) data.value.live = true }
    catch (failure) { error.value = phrase(failure) }
    finally { accessKey.value = ''; busy.value = '' }
  }
  async function loadProposal(reference) {
    try { const result = await request('proposal', { reference }); data.value.proposals[reference] = result.proposal }
    catch (failure) { error.value = phrase(failure) }
  }
  async function analyze(reference, consent) {
    if (busy.value) return
    busy.value = 'analyze'; error.value = ''
    try { const result = await request('analyze', { reference, consent }); data.value.proposals[reference] = result.proposal }
    catch (failure) { error.value = phrase(failure) }
    finally { busy.value = '' }
  }
  async function review(reference, proposalId, decision, note, confirmed) {
    if (busy.value) return
    busy.value = 'review'; error.value = ''
    try { await request('review', { reference, proposalId, decision, note, confirmed }); await loadProposal(reference); await refresh() }
    catch (failure) { error.value = phrase(failure) }
    finally { busy.value = '' }
  }
  async function downloadPolicy(reference, documentId = null, filename = null) {
    if (busy.value) return
    busy.value = 'download'; error.value = ''
    try {
      // The proposal trace identifies the actual stored PDF the agent read.
      const step = data.value.proposals[reference]?.data.trace.find(s => s.tool === 'read_policy_pdf' && s.document_id)
      if (!step && !documentId) throw Object.assign(new Error(), { code: 'POLICY_PDF_UNAVAILABLE' })
      const blob = await request('document', { reference, documentId: documentId || step.document_id }, true)
      const url = URL.createObjectURL(blob), anchor = document.createElement('a')
      anchor.href = url; anchor.download = filename || step?.filename || `policy-${reference}.pdf`; anchor.click(); setTimeout(() => URL.revokeObjectURL(url), 1000)
    } catch (failure) { error.value = phrase(failure) }
    finally { busy.value = '' }
  }
  return { data, busy, error, accessKey, restore, unlock, refresh, loadProposal, analyze, review, downloadPolicy }
}
