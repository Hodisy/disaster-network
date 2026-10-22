import { createHash, scryptSync } from 'node:crypto'
import { DEMO_CUSTOMER_ID, DEMO_POLICY_ID, DEMO_CLAIM_ID, DEMO_CLAIM_REFERENCE, DEMO_CONTRACT_NUMBER, DEMO_PRODUCT_NAME } from '../app/config/demo.js'

// Entirely fictional fixtures. No remote service, real identity or real policy is used.
const timestamp = '2026-10-22T08:00:00+02:00'
const id = (group, number) => `${group}0000000-0000-4000-8000-${String(number).padStart(12, '0')}`
export const TABLE_ORDER = ['customers', 'documents', 'policies', 'claims', 'document_sections', 'messages', 'claim_events', 'providers', 'assignments']
export const DEMO_SCHEMA = 'disaster_demo'

const benefits = [
  ['accommodation', 'Emergency accommodation', 105000, 'Up to seven nights at EUR 150 per household per night, at most EUR 1,050. Confirm household size, unusability of the main residence, dates and current accommodation need. Extensions require separate review.'],
  ['essential_items', 'Essential items', 25000, 'Up to EUR 250 per household for identified essential expenditure. Confirm need and itemised costs.'],
  ['electrical_safety', 'Electrical inspection and safety', 50000, 'Up to EUR 500 for assessed electrical inspection and immediate safety work by a qualified professional. Physical work requires permitted access.'],
  ['cleaning_drying', 'Cleaning and drying', 300000, 'Up to EUR 3,000 for assessed cleaning and drying. Confirm damage, access, work scope and resource availability.'],
  ['moving_storage', 'Moving and temporary storage', 100000, 'Up to EUR 1,000 to protect recoverable belongings. Do not count the same belongings as both a total loss and intact stored property.'],
  ['contents', 'Damaged contents', 3000000, 'Up to EUR 30,000 for assessed eligible contents damage. Require an inventory, ownership or value evidence and assessment. No automatic new-for-old settlement.'],
  ['building', 'Insured building repairs', 20000000, 'Up to EUR 200,000 for assessed eligible damage to the declared owner-occupied main residence. Verify cause and repair scope.']
]
const exclusions = 'Scope and exclusions. This policy excludes personal injury, vehicles, commercial interruption, undeclared property and unrelated or pre-existing damage. Uncertain causation or ground instability requires human review. A single EUR 300 excess applies to combined eligible contents and building damage per household and event, after limits and eligibility assessment. No excess applies to the five assistance benefits. The contractual flood route does not require a CatNat order; CatNat recognition remains unverified and must not be inferred. An assistance request does not confirm a booking. Cover and reimbursement remain subject to review of the declared facts and supporting evidence.'

function verificationHash(customerId) {
  // Public evaluation answer: olivier. Never use this fixture mechanism for real users.
  const salt = createHash('sha256').update(`fictional:${customerId}`).digest('hex').slice(0, 32)
  const digest = scryptSync('olivier', salt, 32).toString('hex')
  return `scrypt:${salt}:${digest}`
}

function document(documentId, kind, filename, text) {
  const content = Buffer.from(text, 'utf8')
  return {
    id: documentId, claim_id: null, kind, filename, mime_type: 'text/plain',
    size_bytes: content.length, content_hex: content.toString('hex'),
    sha256: createHash('sha256').update(content).digest('hex'),
    language: 'en', source_url: null, provenance: 'synthetic',
    extraction_status: 'complete', review_status: 'received', created_at: timestamp
  }
}

export function createDemoDataset() {
  const generalId = id(4, 1)
  const customers = [
    { id: DEMO_CUSTOMER_ID, first_name: 'Jules', last_name: null, household_size: 3 },
    { id: id(1, 2), first_name: 'Camille', last_name: null, household_size: 1 },
    { id: id(1, 3), first_name: 'Sam', last_name: null, household_size: 2 }
  ].map((customer, index) => ({
    ...customer, email: `${customer.first_name.toLowerCase()}@policyholder.invalid`, phone: null,
    verification_question: 'What is your security word?',
    verification_answer_hash: verificationHash(customer.id), created_at: timestamp
  }))
  const generalText = [
    'ALLSTORM HOME PROTECT - GENERAL TERMS - 2026 EDITION',
    'Policy period: 1 January to 31 December 2026. These general terms describe flood-related assistance, contents and building cover for a declared main residence.',
    ...benefits.map(([code, title, , text]) => `${code}: ${title}\n${text}`),
    exclusions,
    'Evaluation policy. Allstorm Insurance and the policyholders are fictional; no insurance contract or external service is created.'
  ].join('\n\n')
  const documents = [document(generalId, 'general_terms', 'allstorm-home-protect-general-terms-2026.txt', generalText)]
  const sections = benefits.map(([code, title, , text], index) => ({
    id: id(5, index + 1), document_id: generalId, position: index,
    title, clause_reference: `HP-${code.toUpperCase()}`, page_start: null, page_end: null,
    text, language: 'en'
  }))
  sections.push({ id: id(5, 8), document_id: generalId, position: 7, title: 'Scope, exclusions and deductible', clause_reference: 'HP-SCOPE', page_start: null, page_end: null, text: exclusions, language: 'en' })
  const policies = customers.map((customer, index) => {
    const specificId = id(4, index + 2)
    const contractNumber = String(Number(DEMO_CONTRACT_NUMBER) + index)
    const town = 'Istres'
    const address = index === 0 ? '12 rue des Oliviers, 13800 Istres' : index === 1 ? 'Entressen, Istres' : 'Les Heures Claires, Istres'
    const specificText = `${DEMO_PRODUCT_NAME.toUpperCase()} - POLICY SCHEDULE - ${contractNumber}\nNamed policyholder: ${customer.first_name}. Household: ${customer.household_size} ${customer.household_size === 1 ? 'person' : 'people'}.\nDeclared owner-occupied main residence: ${address}.\n${DEMO_PRODUCT_NAME}, 2026 edition. Policy period: 1 January to 31 December 2026.\nThe general terms and benefit limits apply. Individual requests are subject to the policy conditions and assessment.\nEvaluation copy: fictional insurer and policyholder; no binding insurance contract.`
    documents.push(document(specificId, 'specific_terms', `allstorm-home-protect-${contractNumber}.txt`, specificText))
    sections.push({ id: id(5, 9 + index), document_id: specificId, position: 0, title: 'Declared household and insured residence', clause_reference: 'HP-SCHEDULE', page_start: null, page_end: null, text: specificText, language: 'en' })
    return {
      id: index === 0 ? DEMO_POLICY_ID : id(2, index + 1), customer_id: customer.id,
      contract_number: contractNumber, version: 1, product_name: DEMO_PRODUCT_NAME,
      valid_from: '2026-01-01', valid_to: '2026-12-31', address, city: town,
      postal_code: '13800', insee_code: '13047',
      // Town-scale illustrative points, not real household coordinates.
      latitude: index === 1 ? 43.599 : index === 2 ? 43.491 : 43.5148, longitude: index === 1 ? 4.939 : index === 2 ? 4.984 : 4.9784,
      property_type: 'house', occupancy: 'owner_occupier', general_terms_document_id: generalId,
      specific_terms_document_id: specificId,
      coverage_data: {
        fictional: true, currency: 'EUR', catnat_status: 'not_verified',
        benefits: benefits.map(([code, , limit]) => ({ code, limit_cents: limit })),
        accommodation_max_nights: 7, accommodation_max_nightly_cents: 15000,
        contents_building_deductible_cents: 30000
      }, created_at: timestamp
    }
  })
  const claims = policies.map((policy, index) => ({
    id: index === 0 ? DEMO_CLAIM_ID : id(3, index + 1),
    reference: index === 0 ? DEMO_CLAIM_REFERENCE : `CAT-2026-${String(index + 1).padStart(3, '0')}`,
    policy_id: policy.id, incident_type: 'flood', incident_at: '2026-10-20T10:00:00+02:00',
    reported_at: index === 0 ? null : '2026-10-22T08:15:00+02:00',
    intake_status: index === 0 ? 'awaiting_report' : 'reported',
    description: index === 0 ? null : index === 1 ? 'Water damaged belongings. The home remains habitable.' : 'The home is inaccessible. Physical damage has not been confirmed.',
    summary: index === 0 ? null : index === 1 ? 'Collect an inventory and photos; assess cleaning needs.' : 'Confirm access and temporary accommodation needs; do not infer physical damage.',
    stage: 'collecting_documents', habitability: index === 1 ? 'habitable' : 'unknown',
    access_status: index === 2 ? 'restricted' : 'unknown',
    priority: index === 0 ? null : index === 1 ? 'P3' : 'P2',
    priority_reasons: index === 0 ? [] : index === 1 ? ['Home remains habitable; damage assessment pending.'] : ['Home inaccessible; assistance needs require confirmation.'],
    needs: index === 0 ? [] : [{ code: index === 1 ? 'cleaning_drying' : 'accommodation', status: 'to_confirm', quantity: 1 }],
    required_documents: ['photos', 'address', 'inventory'],
    assessment: { status: 'not_reviewed', source_section_ids: [], missing_information: [], fictional: true },
    cost_estimate: { currency: 'EUR', gross_cents: null, approved_cents: null, paid_cents: 0, lines: [], fictional: true },
    updated_at: index === 0 ? timestamp : '2026-10-22T08:16:00+02:00', closed_at: null
  }))
  const messages = [1, 2].flatMap(index => [
    { id: id(6, index * 2), customer_id: customers[index].id, claim_id: claims[index].id, conversation_id: id(9, index), channel: 'chat', role: 'client', content: claims[index].description, language: 'en', created_at: '2026-10-22T08:15:00+02:00' },
    { id: id(6, index * 2 + 1), customer_id: customers[index].id, claim_id: claims[index].id, conversation_id: id(9, index), channel: 'chat', role: 'assistant', content: 'Your report has been recorded. Add any photos or supporting documents already available to you. Do not enter a restricted property to collect evidence.', language: 'en', created_at: '2026-10-22T08:16:00+02:00' }
  ])
  const events = claims.map((claim, index) => ({
    id: id(7, index + 1), claim_id: claim.id, type: index === 0 ? 'demo_record_reserved' : 'report_received',
    actor: 'demo_setup', data: { fictional: true, source: 'small-demo-fixture', coverage_confirmed: false },
    created_at: index === 0 ? timestamp : '2026-10-22T08:15:00+02:00'
  }))
  const providers = [
    { id: id(8, 1), name: 'Delta Drying', service_types: ['cleaning_drying'], capacity: 2, pricing: { currency: 'EUR', unit: 'intervention', estimated_unit_cents: 180000 } },
    { id: id(8, 2), name: 'Lagoon Family Accommodation', service_types: ['accommodation'], capacity: 4, pricing: { currency: 'EUR', unit: 'household_night', estimated_unit_cents: 12500 } },
    { id: id(8, 3), name: 'Lagoon Electrical', service_types: ['electrical_safety'], capacity: 1, pricing: { currency: 'EUR', unit: 'intervention', estimated_unit_cents: 35000 } }
  ].map(provider => ({ ...provider, latitude: 43.51, longitude: 4.99, service_area: ['13047', '13092'], availability: { status: 'available', starts_at: '2026-10-23T08:00:00+02:00', ends_at: '2026-10-23T18:00:00+02:00', fictional: true } }))
  return {
    customers, documents, policies, claims, document_sections: sections, messages,
    claim_events: events, providers, assignments: []
  }
}
