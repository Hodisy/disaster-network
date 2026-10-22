import { createHash, randomUUID } from 'node:crypto'
import { transaction } from './hotlineClaims.js'
import { policyPdfProjection, policyDocumentHash } from '../utils/policyPdfReferences.js'

const fail = code => { throw Object.assign(new Error(code), { code }) }
export const bundleHash = bundle => createHash('sha256').update(JSON.stringify(bundle, (_key, value) => value && typeof value === 'object' && !Array.isArray(value)
  ? Object.fromEntries(Object.keys(value).sort().map(key => [key, value[key]])) : value)).digest('hex')
export async function loadInsurerBundle(db, reference) {
  const claim = (await db.query(`SELECT c.*, c.updated_at::text AS version_token, u.first_name, u.household_size
    FROM disaster_demo.claims c JOIN disaster_demo.policies p ON p.id = c.policy_id JOIN disaster_demo.customers u ON u.id = p.customer_id
    WHERE c.reference = $1 AND c.reference ~ '^CAT-2026-[0-9]{3}$' FOR UPDATE OF c`, [reference])).rows[0]
  if (!claim) fail('INSURER_RECORD_MISSING')
  const source = (await db.query('SELECT *, valid_from::text AS valid_from, valid_to::text AS valid_to FROM disaster_demo.policies WHERE id = $1', [claim.policy_id])).rows[0]
  const sections = (await db.query(`SELECT s.id, s.document_id, s.title, s.clause_reference, s.text, s.page_start, s.page_end, d.filename, d.provenance
    FROM disaster_demo.document_sections s JOIN disaster_demo.documents d ON d.id = s.document_id
    WHERE s.document_id = ANY($1::uuid[]) ORDER BY s.document_id, s.position`, [[source.general_terms_document_id, source.specific_terms_document_id]])).rows
  const policy = policyPdfProjection(source, sections)
  const pdf = (await db.query(`SELECT d.id, d.filename, d.sha256, d.size_bytes FROM disaster_demo.claim_events e JOIN disaster_demo.documents d ON d.id::text = e.data->>'document_id'
    WHERE e.claim_id = $1 AND e.type = 'policy_pdf_stored' AND e.data->>'source_hash' = $2 AND e.data->>'sha256' = d.sha256 AND d.mime_type = 'application/pdf' ORDER BY e.created_at DESC LIMIT 1`, [claim.id, policyDocumentHash(policy)])).rows[0] || null
  const documents = (await db.query(`SELECT d.id, d.kind, d.filename, d.sha256, d.mime_type, d.extraction_status, d.review_status,
    (SELECT e.data->'result' FROM disaster_demo.claim_events e WHERE e.claim_id = $1 AND e.type = 'evidence_analysis_completed'
      AND e.data->>'document_id' = d.id::text AND e.data->>'sha256' = d.sha256 ORDER BY e.created_at DESC LIMIT 1) AS analysis
    FROM disaster_demo.documents d WHERE d.claim_id = $1 ORDER BY d.id`, [claim.id])).rows
  const messages = (await db.query('SELECT id, role, content, language, created_at FROM disaster_demo.messages WHERE claim_id = $1 ORDER BY created_at DESC LIMIT 15', [claim.id])).rows
  const intake = (await db.query(`SELECT data->'draft' AS draft FROM disaster_demo.claim_events WHERE claim_id = $1 AND type = 'hotline_report_saved' ORDER BY created_at DESC LIMIT 1`, [claim.id])).rows[0]?.draft || null
  const officialSource = (await db.query(`SELECT data FROM disaster_demo.claim_events WHERE claim_id = $1 AND type = 'official_source_checked' ORDER BY created_at DESC LIMIT 1`, [claim.id])).rows[0]?.data || null
  const providers = (await db.query('SELECT * FROM disaster_demo.providers ORDER BY id')).rows
  const assignments = (await db.query(`SELECT * FROM disaster_demo.assignments WHERE status <> 'cancelled' ORDER BY id`)).rows
  const reviews = (await db.query("SELECT data FROM disaster_demo.claim_events WHERE claim_id = $1 AND type = 'insurer_human_decision' ORDER BY created_at DESC LIMIT 5", [claim.id])).rows.map(row => row.data)
  return { claim, policy, pdf, documents, messages, intake, officialSource, providers, assignments, reviews }
}
export const getInsurerBundle = (config, reference) => transaction(config, db => loadInsurerBundle(db, reference))
export function getInsurerPortfolio(config) {
  return transaction(config, async db => {
    const claims = (await db.query(`SELECT c.*, p.contract_number, p.product_name, p.address, p.city, p.longitude, p.latitude, p.occupancy,
      u.first_name, u.household_size, p.coverage_data AS policy_coverage_data FROM disaster_demo.claims c JOIN disaster_demo.policies p ON p.id = c.policy_id JOIN disaster_demo.customers u ON u.id = p.customer_id
      WHERE c.reference >= 'CAT-2026-001' AND c.reference <= 'CAT-2026-080' ORDER BY c.reference`)).rows
    return { claims, providers: (await db.query('SELECT * FROM disaster_demo.providers ORDER BY id')).rows,
      assignments: (await db.query("SELECT * FROM disaster_demo.assignments WHERE status <> 'cancelled' ORDER BY id")).rows }
  })
}
export function latestInsurerProposal(config, reference) {
  return transaction(config, async db => {
    const { rows } = await db.query(`SELECT e.id, e.data, e.created_at,
      (SELECT h.data FROM disaster_demo.claim_events h WHERE h.claim_id = c.id AND h.type = 'insurer_human_decision'
        AND h.data->>'proposal_id' = e.id::text ORDER BY h.created_at DESC LIMIT 1) AS decision
      FROM disaster_demo.claims c JOIN disaster_demo.claim_events e ON e.claim_id = c.id
      WHERE c.reference = $1 AND e.type = 'insurer_ai_proposal' ORDER BY e.created_at DESC LIMIT 1`, [reference])
    return rows[0] || null
  })
}
export function saveInsurerProposal(config, reference, baseHash, proposal, trace, model, locale) {
  return transaction(config, async db => {
    const current = await loadInsurerBundle(db, reference)
    if (bundleHash(current) !== baseHash) fail('PROPOSAL_STALE')
    const id = randomUUID()
    await db.query(`INSERT INTO disaster_demo.claim_events (id, claim_id, type, actor, data) VALUES ($1, $2, 'insurer_ai_proposal', 'agent', $3::jsonb)`,
      [id, current.claim.id, JSON.stringify({ baseHash, proposal, trace, model, locale, simulated_operations: true, simulated_payments: true })])
    return { id, data: { baseHash, proposal, trace, model, locale }, decision: null, created_at: new Date().toISOString() }
  })
}
export function storeInsurerSource(config, reference, result) {
  return transaction(config, async db => {
    const claim = (await db.query('SELECT id FROM disaster_demo.claims WHERE reference = $1', [reference])).rows[0]
    if (!claim) fail('INSURER_RECORD_MISSING')
    await db.query(`INSERT INTO disaster_demo.claim_events (id, claim_id, type, actor, data) VALUES ($1, $2, 'official_source_checked', 'agent', $3::jsonb)`, [randomUUID(), claim.id, JSON.stringify(result)])
  })
}
export function readInsurerDocument(config, reference, documentId) {
  return transaction(config, async db => {
    const bundle = await loadInsurerBundle(db, reference)
    if (!bundle.documents.some(d => d.id === documentId) && bundle.pdf?.id !== documentId) fail('DOCUMENT_NOT_FOUND')
    const row = (await db.query('SELECT id, content, mime_type, filename, sha256 FROM disaster_demo.documents WHERE id = $1', [documentId])).rows[0]
    if (!row) fail('DOCUMENT_NOT_FOUND')
    if (createHash('sha256').update(row.content).digest('hex') !== row.sha256) fail('DOCUMENT_NOT_FOUND')
    return row
  })
}
export function storeInsurerExtraction(config, reference, baseHash, document, analysis) {
  return transaction(config, async db => {
    const before = await loadInsurerBundle(db, reference)
    if (bundleHash(before) !== baseHash || !before.documents.some(d => d.id === document.id && d.sha256 === document.sha256)) fail('PROPOSAL_STALE')
    await db.query(`UPDATE disaster_demo.documents SET extraction_status = 'complete',
      review_status = CASE WHEN review_status = 'received' THEN 'needs_review' ELSE review_status END WHERE id = $1`, [document.id])
    await db.query(`INSERT INTO disaster_demo.claim_events (id, claim_id, type, actor, data) VALUES ($1, $2, 'evidence_analysis_completed', 'agent', $3::jsonb)`,
      [randomUUID(), before.claim.id, JSON.stringify({ document_id: document.id, sha256: document.sha256, operation_id: randomUUID(), result: analysis })])
    return loadInsurerBundle(db, reference)
  })
}
export function reviewInsurerProposal(config, reference, proposalId, decision, humanNote, validator) {
  return transaction(config, async db => {
    const bundle = await loadInsurerBundle(db, reference)
    const event = (await db.query("SELECT * FROM disaster_demo.claim_events WHERE id = $1 AND claim_id = $2 AND type = 'insurer_ai_proposal'", [proposalId, bundle.claim.id])).rows[0]
    if (!event) fail('INSURER_RECORD_MISSING')
    const previous = (await db.query("SELECT data FROM disaster_demo.claim_events WHERE claim_id = $1 AND type = 'insurer_human_decision' AND data->>'proposal_id' = $2", [bundle.claim.id, proposalId])).rows[0]
    if (previous) {
      if (previous.data.decision === decision) return { decision: previous.data, replayed: true }
      fail('PROPOSAL_ALREADY_REVIEWED')
    }
    if (bundle.claim.closed_at) fail('CLAIM_CLOSED')
    if (decision === 'approve' && bundleHash(bundle) !== event.data.baseHash) fail('PROPOSAL_STALE')
    const proposal = event.data.proposal
    const checked = decision === 'approve' ? validator(bundle, proposal) : proposal
    const assignmentIds = []
    if (decision === 'approve') {
      for (const operation of checked.operations.filter(op => op.eligible)) {
        const id = randomUUID()
        await db.query(`INSERT INTO disaster_demo.assignments (id, claim_id, provider_id, service_type, starts_at, ends_at, status, estimated_cost_cents, currency, instructions)
          VALUES ($1, $2, $3, $4, $5, $6, 'approved_in_demo', $7, 'EUR', $8)`,
        [id, bundle.claim.id, operation.provider_id, operation.service, operation.start, operation.end, operation.estimated_cents, operation.reason])
        assignmentIds.push(id)
      }
      await db.query(`UPDATE disaster_demo.claims SET priority = $2, priority_reasons = $3::jsonb, assessment = $4::jsonb, updated_at = now() WHERE id = $1`,
        [bundle.claim.id, checked.priority, JSON.stringify([checked.rationale]), JSON.stringify({ ...bundle.claim.assessment,
          status: 'human_reviewed_proposal', group: checked.group, proposal_id: proposalId, coverage_proposal: checked.coverage_lines,
          reimbursement_proposal_cents: checked.reimbursement_cents, summary: checked.summary, missing_information: checked.missing_information,
          simulated_operations: true, simulated_payments: true, human_reviewed_at: new Date().toISOString() })])
      // Estimates, approved financial commitments, paid money, stage and site permission stay independent.
    }
    const recorded = { proposal_id: proposalId, decision, note: humanNote, assignment_ids: assignmentIds,
      reviewed_at: new Date().toISOString(), simulated_operations: true, simulated_payments: true, paid_cents: 0 }
    await db.query(`INSERT INTO disaster_demo.claim_events (id, claim_id, type, actor, data) VALUES ($1, $2, 'insurer_human_decision', 'handler', $3::jsonb)`, [randomUUID(), bundle.claim.id, JSON.stringify(recorded)])
    return { decision: recorded, replayed: false }
  })
}
