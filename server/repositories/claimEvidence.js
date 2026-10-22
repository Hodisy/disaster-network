import { randomUUID } from 'node:crypto'
import { transaction } from './hotlineClaims.js'
import { DEMO_CLAIM_ID, DEMO_POLICY_ID, DEMO_CUSTOMER_ID } from '../../app/config/demo.js'

const fail = code => { throw Object.assign(new Error(code), { code }) }
const metadata = 'id, kind, filename, mime_type, size_bytes, sha256, extraction_status, review_status, created_at'

async function claimFor(db, writable = false) {
  const { rows } = await db.query(`SELECT c.* FROM disaster_demo.claims c JOIN disaster_demo.policies p ON p.id = c.policy_id
    WHERE c.id = $1 AND p.id = $2 AND p.customer_id = $3 FOR UPDATE OF c`, [DEMO_CLAIM_ID, DEMO_POLICY_ID, DEMO_CUSTOMER_ID])
  if (!rows[0] || rows[0].intake_status !== 'reported') fail('DEMO_RECORD_MISSING')
  if (writable && rows[0].closed_at) fail('CLAIM_CLOSED')
  return rows[0]
}
async function event(db, type, actor, data) {
  await db.query('INSERT INTO disaster_demo.claim_events (id, claim_id, type, actor, data) VALUES ($1, $2, $3, $4, $5::jsonb)',
    [randomUUID(), DEMO_CLAIM_ID, type, actor, JSON.stringify(data)])
}
export function listEvidence(config) {
  return transaction(config, async db => {
    await claimFor(db)
    const { rows } = await db.query(`SELECT d.id, d.kind, d.filename, d.mime_type, d.size_bytes, d.sha256,
      d.extraction_status, d.review_status, d.created_at, a.data->'result' AS analysis
      FROM disaster_demo.documents d LEFT JOIN LATERAL (
        SELECT data FROM disaster_demo.claim_events WHERE claim_id = $1 AND type = 'evidence_analysis_completed'
        AND data->>'document_id' = d.id::text AND data->>'sha256' = d.sha256 ORDER BY created_at DESC LIMIT 1
      ) a ON true WHERE d.claim_id = $1 ORDER BY d.created_at, d.id`, [DEMO_CLAIM_ID])
    const source = await db.query(`SELECT data FROM disaster_demo.claim_events WHERE claim_id = $1
      AND type = 'official_source_checked' ORDER BY created_at DESC LIMIT 1`, [DEMO_CLAIM_ID])
    return { documents: rows, officialSource: source.rows[0]?.data || null }
  })
}
export function storeEvidence(config, file, locale) {
  return transaction(config, async db => {
    await claimFor(db, true)
    const prior = await db.query(`SELECT ${metadata} FROM disaster_demo.documents WHERE claim_id = $1 AND sha256 = $2 AND kind = $3`,
      [DEMO_CLAIM_ID, file.sha256, file.kind])
    if (prior.rows[0]) return { document: prior.rows[0], replayed: true }
    const totals = await db.query('SELECT count(*)::int AS count, COALESCE(sum(size_bytes), 0)::int AS bytes FROM disaster_demo.documents WHERE claim_id = $1', [DEMO_CLAIM_ID])
    if (totals.rows[0].count >= 12 || totals.rows[0].bytes + file.content.length > 30 * 1024 * 1024) fail('DOCUMENT_LIMIT_REACHED')
    const id = randomUUID()
    const { rows } = await db.query(`INSERT INTO disaster_demo.documents
      (id, claim_id, kind, filename, mime_type, size_bytes, content, sha256, language, provenance)
      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'customer_provided') RETURNING ${metadata}`,
    [id, DEMO_CLAIM_ID, file.kind, file.filename, file.mime, file.content.length, file.content, file.sha256, locale])
    await event(db, 'evidence_received', 'client', { document_id: id, sha256: file.sha256, category: file.kind, coverage_confirmed: false })
    return { document: rows[0], replayed: false }
  })
}
export function readEvidence(config, id) {
  return transaction(config, async db => {
    await claimFor(db)
    const { rows } = await db.query(`SELECT ${metadata}, content FROM disaster_demo.documents WHERE id = $1 AND claim_id = $2`, [id, DEMO_CLAIM_ID])
    if (!rows[0]) fail('DOCUMENT_NOT_FOUND')
    return rows[0]
  })
}
export function beginAnalysis(config, documentId, operationId) {
  return transaction(config, async db => {
    await claimFor(db, true)
    const current = await db.query('SELECT id FROM disaster_demo.documents WHERE id = $1 AND claim_id = $2', [documentId, DEMO_CLAIM_ID])
    if (!current.rows[0]) fail('DOCUMENT_NOT_FOUND')
    const recent = await db.query(`SELECT started.id FROM disaster_demo.claim_events started WHERE started.claim_id = $1 AND started.type = 'evidence_analysis_started'
      AND started.data->>'document_id' = $2 AND started.created_at > now() - interval '2 minutes'
      AND NOT EXISTS (SELECT 1 FROM disaster_demo.claim_events e WHERE e.claim_id = $1
        AND e.type IN ('evidence_analysis_completed', 'evidence_analysis_failed')
        AND e.data->>'operation_id' = started.data->>'operation_id')`, [DEMO_CLAIM_ID, documentId])
    if (recent.rows.length) fail('DOCUMENT_ANALYSIS_BUSY')
    await event(db, 'evidence_analysis_started', 'client', { document_id: documentId, operation_id: operationId })
  })
}
export function finishAnalysis(config, document, operationId, result) {
  return transaction(config, async db => {
    await claimFor(db, true)
    const updated = await db.query(`UPDATE disaster_demo.documents SET extraction_status = $3,
      review_status = CASE WHEN review_status = 'received' THEN 'needs_review' ELSE review_status END
      WHERE id = $1 AND claim_id = $2 AND sha256 = $4 RETURNING id`,
    [document.id, DEMO_CLAIM_ID, result ? 'complete' : 'failed', document.sha256])
    if (!updated.rows.length) fail('DOCUMENT_NOT_FOUND')
    await event(db, result ? 'evidence_analysis_completed' : 'evidence_analysis_failed', 'agent', {
      document_id: document.id, sha256: document.sha256, operation_id: operationId, ...(result ? { result } : {})
    })
    // No claim stage, entitlement, payment, booking or access status changes.
  })
}
export function storeOfficialSource(config, result) {
  return transaction(config, async db => {
    await claimFor(db, true)
    await event(db, 'official_source_checked', 'client', result)
    return result
  })
}
