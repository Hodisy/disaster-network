import { createHash } from 'node:crypto'
import { createDemoDataset } from './fixtures.js'
import { createPortfolio, createProviders } from '../app/composables/useInsurerWorkspace.js'

// The same synthetic scenario as the UI, inserted only by explicit preparation.
export const scenarioId = (group, n) => `${group}0000000-0000-4000-8000-${String(n).padStart(12, '0')}`
export const serviceCodes = { 'Temporary accommodation': 'accommodation', 'Electrical safety': 'electrical_safety', 'Cleaning and drying': 'cleaning_drying', 'Moving and storage': 'moving_storage', 'Damage assessment': 'damage_assessment' }
export const moneyCodes = { accommodation: 'accommodation', essentials: 'essential_items', electrical: 'electrical_safety', drying: 'cleaning_drying', storage: 'moving_storage', contents: 'contents', building: 'building' }
const towns = { Istres: ['13047', '13800'], 'Saint-Chamas': ['13092', '13250'], 'Saint-Mitre-les-Remparts': ['13098', '13920'], 'Martigues': ['13056', '13500'] }
export function createInsurerDataset() {
  const dataset = createDemoDataset()
  const portfolio = createPortfolio()
  const template = dataset.policies[0], customerTemplate = dataset.customers[0], claimTemplate = dataset.claims[0]
  for (let index = 3; index < portfolio.length; index++) {
    const source = portfolio[index], number = index + 1
    const customerId = scenarioId(1, number), policyId = scenarioId(2, number), documentId = scenarioId(4, number + 1)
    const text = `${source.productName.toUpperCase()} - POLICY SCHEDULE - ${source.contractNumber}\nNamed policyholder: ${source.name}. Household: ${source.householdSize} ${source.householdSize === 1 ? 'person' : 'people'}.\nDeclared owner-occupied main residence: ${source.address}.\nPolicy period: 1 January to 31 December 2026. The linked general terms and benefit limits apply.\nEvaluation copy: invented insurer, policyholder and approximate location. No binding insurance cover.`
    const content = Buffer.from(text)
    dataset.customers.push({ ...customerTemplate, id: customerId, first_name: source.name, email: `${source.name.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase()}-${number}@policyholder.invalid`, household_size: source.householdSize })
    dataset.documents.push({ ...dataset.documents[0], id: documentId, kind: 'specific_terms', filename: `allstorm-home-protect-${source.contractNumber}.txt`, content_hex: content.toString('hex'), size_bytes: content.length, sha256: createHash('sha256').update(content).digest('hex') })
    dataset.document_sections.push({ id: scenarioId(5, number + 8), document_id: documentId, position: 0, title: 'Declared household and insured residence', clause_reference: 'HP-SCHEDULE', page_start: null, page_end: null, text, language: 'en' })
    dataset.policies.push({ ...template, id: policyId, customer_id: customerId, contract_number: source.contractNumber, specific_terms_document_id: documentId,
      address: source.address, city: source.municipality, insee_code: towns[source.municipality][0], postal_code: towns[source.municipality][1], longitude: source.coordinates[0], latitude: source.coordinates[1] })
    dataset.claims.push({ ...claimTemplate, id: scenarioId(3, number), reference: source.id, policy_id: policyId,
      reported_at: source.reportedAt, intake_status: 'reported', description: source.summary, summary: source.summary,
      habitability: source.group === 'A' ? 'uninhabitable' : source.group === 'C' ? 'habitable' : 'unknown',
      access_status: source.access === 'Permitted' ? 'permitted_in_demo' : source.access.toLowerCase(), priority: source.priority, priority_reasons: [source.priorityReason],
      needs: source.needs.map(label => ({ code: serviceCodes[label], status: 'identified', quantity: 1 })),
      assessment: { status: 'not_reviewed', fictional: true, scenario_group: source.group, location_precision: 'approximate' },
      cost_estimate: { ...claimTemplate.cost_estimate, gross_cents: source.moneyLines.some(l => l.amount === null) ? null : source.moneyLines.reduce((a, l) => a + l.amount * 100, 0),
        lines: source.moneyLines.map(l => ({ code: moneyCodes[l.id], estimated_cents: l.amount === null ? null : l.amount * 100, approved_cents: null, paid_cents: 0 })) },
      stage: source.stage === 4 ? 'claim_closed' : 'collecting_documents', closed_at: source.stage === 4 ? '2026-10-22T11:45:00+02:00' : null })
  }
  const providers = createProviders()
  const mapping = { 'dry-1': 1, 'accommodation-1': 2, 'electrical-1': 3, 'dry-2': 4, 'dry-3': 5, 'dry-4': 6, 'dry-5': 7, 'dry-6': 8, 'assessment-1': 9, 'storage-1': 10 }
  for (const source of providers) {
    const id = scenarioId(8, mapping[source.id])
    if (dataset.providers.some(p => p.id === id)) continue
    dataset.providers.push({ id, name: source.name, service_types: [serviceCodes[source.skill]], latitude: source.coordinates[1], longitude: source.coordinates[0],
      service_area: Object.values(towns).map(t => t[0]), capacity: source.capacity,
      availability: { status: 'available', starts_at: '2026-10-23T08:00:00+02:00', ends_at: '2026-10-23T18:00:00+02:00', fictional: true },
      pricing: { currency: 'EUR', unit: source.skill === 'Temporary accommodation' ? 'household_night' : 'intervention', estimated_unit_cents: source.skill === 'Moving and storage' ? 35000 : source.skill === 'Damage assessment' ? 20000 : 54000 } })
  }
  return dataset
}
