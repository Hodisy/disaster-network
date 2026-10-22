import { policyHash as evidencePolicyHash } from '../hotline/evidence.js'
const fail = () => { throw Object.assign(new Error('INVALID_PROPOSAL'), { code: 'INVALID_PROPOSAL' }) }
export const services = ['accommodation', 'electrical_safety', 'cleaning_drying', 'moving_storage', 'damage_assessment']
export const benefits = ['accommodation', 'essential_items', 'electrical_safety', 'cleaning_drying', 'moving_storage', 'contents', 'building']
export function distanceKm(a, b) {
  if ([a.latitude, a.longitude, b.latitude, b.longitude].some(value => typeof value !== 'number' || !Number.isFinite(value))) return null
  const radians = n => n * Math.PI / 180
  const x = Math.sin(radians(b.latitude - a.latitude) / 2) ** 2 + Math.cos(radians(a.latitude)) * Math.cos(radians(b.latitude)) * Math.sin(radians(b.longitude - a.longitude) / 2) ** 2
  return Math.round(6371 * 2 * Math.asin(Math.sqrt(Math.min(1, x))) * 10) / 10
}
export function availableTeams(bundle, service, reservations = new Map()) {
  return bundle.providers.filter(p => p.service_types.includes(service)).map(provider => {
    const start = provider.availability.starts_at, end = provider.availability.ends_at
    const validWindow = Number.isFinite(Date.parse(start)) && Number.isFinite(Date.parse(end)) && Date.parse(end) > Date.parse(start)
    const used = bundle.assignments.filter(a => a.provider_id === provider.id && ['approved_in_demo', 'in_progress_in_demo'].includes(a.status)
      && Date.parse(a.starts_at) < Date.parse(end) && Date.parse(a.ends_at) > Date.parse(start)).length
    const remaining = Math.max(0, provider.capacity - used - (reservations.get(provider.id) || 0))
    const reasons = []
    if (service !== 'accommodation' && bundle.claim.access_status !== 'permitted_in_demo') reasons.push('Physical access has not been permitted.')
    if (!provider.service_area.includes(bundle.policy.insee_code)) reasons.push('The municipality is outside this provider service area.')
    if (provider.availability.status !== 'available' || !validWindow) reasons.push('Provider availability is not confirmed.')
    if (!remaining) reasons.push('No capacity is available in this service window.')
    if (distanceKm(bundle.policy, provider) === null) reasons.push('A location is missing; provider proximity cannot be assessed.')
    if (bundle.assignments.some(a => a.claim_id === bundle.claim.id && a.service_type === service && ['approved_in_demo', 'in_progress_in_demo', 'completed_in_demo'].includes(a.status))) reasons.push('This service already has a recorded assignment.')
    if (!Number.isSafeInteger(provider.pricing.estimated_unit_cents) || provider.pricing.estimated_unit_cents < 0) reasons.push('A service estimate is missing.')
    return { provider_id: provider.id, name: provider.name, service, remaining, capacity: provider.capacity, unit: provider.pricing.unit,
      coordinates: [provider.longitude, provider.latitude], distance_km: distanceKm(bundle.policy, provider), start, end,
      estimated_cents: provider.pricing.estimated_unit_cents, eligible: reasons.length === 0, reasons, simulated: true }
  }).sort((a, b) => Number(b.eligible) - Number(a.eligible) || (a.distance_km ?? Infinity) - (b.distance_km ?? Infinity) || a.estimated_cents - b.estimated_cents)
}
export function validateProposal(bundle, proposal) {
  if (!proposal || !['A', 'B', 'C', 'D'].includes(proposal.group) || !['P1', 'P2', 'P3'].includes(proposal.priority)) fail()
  for (const field of ['summary', 'rationale', 'reimbursement_explanation']) if (typeof proposal[field] !== 'string' || !proposal[field].trim() || proposal[field].length > 3000) fail()
  for (const field of ['coverage_lines', 'operations', 'missing_information', 'evidence_document_ids', 'clause_references']) if (!Array.isArray(proposal[field]) || proposal[field].length > 12) fail()
  if (proposal.coverage_lines.length !== benefits.length) fail()
  if (proposal.missing_information.some(s => typeof s !== 'string' || s.length > 1000)) fail()
  const documentIds = new Set(bundle.documents.map(d => d.id))
  const clauses = new Set(bundle.policy.sections.map(s => s.clause_reference))
  if (proposal.evidence_document_ids.some(id => !documentIds.has(id)) || proposal.clause_references.some(ref => !clauses.has(ref))) fail()
  const guardrails = [], reservations = new Map(), seenServices = new Set(), seenBenefits = new Set()
  const operations = proposal.operations.map(op => {
    if (!services.includes(op.service) || seenServices.has(op.service) || typeof op.reason !== 'string' || op.reason.length > 2000 || (op.provider_id !== null && typeof op.provider_id !== 'string')) fail()
    seenServices.add(op.service)
    const team = availableTeams(bundle, op.service, reservations).find(p => p.provider_id === op.provider_id)
    if (team?.eligible) reservations.set(team.provider_id, (reservations.get(team.provider_id) || 0) + 1)
    return { ...op, ...(team || {}), eligible: !!team?.eligible, blocked_reasons: team?.reasons || ['No eligible provider was selected.'], reason: op.reason }
  })
  const eventDate = new Date(bundle.claim.incident_at).toISOString().slice(0, 10)
  const activePolicy = eventDate >= bundle.policy.valid_from && eventDate <= bundle.policy.valid_to
  const coverage_lines = proposal.coverage_lines.map(line => {
    if (!benefits.includes(line.benefit) || seenBenefits.has(line.benefit) || !['proposed', 'needs_review', 'not_applicable'].includes(line.status)
      || !Array.isArray(line.document_ids) || line.document_ids.some(id => !documentIds.has(id)) || !clauses.has(line.clause_reference) || line.clause_reference !== `HP-${line.benefit.toUpperCase()}`
      || typeof line.reason !== 'string' || line.reason.length > 2000 || (line.amount_cents !== null && (!Number.isSafeInteger(line.amount_cents) || line.amount_cents < 0))) fail()
    seenBenefits.add(line.benefit)
    const ceiling = bundle.policy.coverage_data.benefits.find(b => b.code === line.benefit)?.limit_cents
    // The initial extractor can corroborate accommodation invoices. Other loss values remain estimates for handler review.
    let supported = 0
    if (line.benefit === 'accommodation' && activePolicy && bundle.claim.habitability === 'uninhabitable' && bundle.intake?.accommodationNeeded === true) {
      const unique = new Set()
      for (const doc of bundle.documents.filter(d => line.document_ids.includes(d.id))) {
        const analysis = doc.analysis
        if (!analysis || analysis.sourceSha256 !== doc.sha256 || unique.has(doc.sha256) || doc.review_status === 'unusable') continue
        const f = analysis.facts, checks = analysis.comparison?.checks || []
        const invoiceKey = JSON.stringify([f.invoice_total_cents, f.currency, f.stay_start, f.stay_end, f.holder_first_name])
        if (unique.has(invoiceKey) || f.document_type !== 'accommodation_invoice' || f.currency !== 'EUR' || !Number.isSafeInteger(f.invoice_total_cents) || f.invoice_total_cents < 0
          || analysis.comparison?.policyHash !== evidencePolicyHash(bundle.policy)
          || analysis.comparison?.claimVersion !== bundle.claim.version_token || checks.length < 7 || checks.some(c => c.status !== 'consistent')) continue
        unique.add(doc.sha256); unique.add(invoiceKey); supported += f.invoice_total_cents
      }
    }
    const maximum = Number.isSafeInteger(ceiling) ? Math.min(ceiling, supported) : 0
    const amount = activePolicy && maximum > 0 && line.amount_cents !== null ? Math.min(line.amount_cents, maximum) : null
    if (line.amount_cents !== null && amount !== line.amount_cents) guardrails.push('Unsupported reimbursement amounts were held for review.')
    const clause = bundle.policy.sections.find(s => s.clause_reference === line.clause_reference)
    return { ...line, amount_cents: amount, status: line.status === 'not_applicable' ? 'not_applicable' : activePolicy ? line.status : 'needs_review',
      reimbursement_status: amount !== null && line.status === 'proposed' ? 'proposed' : 'needs_review', limit_cents: ceiling ?? null, pdf_page: clause.pdf_page }
  })
  let group = proposal.group, priority = proposal.priority
  if (group === 'A' && bundle.claim.habitability !== 'uninhabitable') { group = 'D'; guardrails.push('Habitability requires confirmation before an uninhabitable classification.') }
  if (group === 'C' && bundle.claim.habitability !== 'habitable') { group = 'D'; guardrails.push('Habitability requires confirmation before a habitable classification.') }
  if (bundle.claim.needs.some(n => n.code === 'accommodation' && n.status === 'identified') && bundle.claim.habitability !== 'habitable') priority = 'P1'
  // No contents/building amount is payable from an estimate alone; the single combined deductible is therefore not silently waived.
  return { ...proposal, group, priority, coverage_lines, operations, guardrails,
    reimbursement_cents: coverage_lines.some(l => l.reimbursement_status === 'proposed') ? coverage_lines.filter(l => l.reimbursement_status === 'proposed').reduce((a, l) => a + l.amount_cents, 0) : null,
    reimbursement_incomplete: coverage_lines.some(l => l.status !== 'not_applicable' && l.reimbursement_status === 'needs_review'), deductible_cents: bundle.policy.coverage_data.contents_building_deductible_cents,
    deductible_status: 'Contents and building require an assessed loss before applying the single combined deductible.',
    requires_human_approval: true, simulated_operations: true, simulated_payments: true, paid_cents: 0 }
}
