import { Client } from 'pg'
import { scrypt, timingSafeEqual, randomUUID, createHash } from 'node:crypto'
import { promisify } from 'node:util'
import { connectionOptions } from './demoDatabase.js'
import { DEMO_CUSTOMER_ID, DEMO_POLICY_ID, DEMO_CLAIM_ID, DEMO_CLAIM_REFERENCE, normalizeContractNumber } from '../../app/config/demo.js'
import { DEMO_LOCK_NAME, DEMO_SCHEMA_MARKER } from '../../database/operations.js'
import { policyPdfProjection, policyDocumentHash } from '../utils/policyPdfReferences.js'

const deriveKey = promisify(scrypt)
const fail = code => Object.assign(new Error(code), { code })

// Every operation shares the maintenance lock, so reset cannot interleave a write.
export async function transaction(config, work) {
  const client = new Client({ ...await connectionOptions(config), application_name: 'disaster-network-hotline' })
  try {
    await client.connect()
    await client.query('BEGIN')
    await client.query('SELECT pg_advisory_xact_lock(hashtext($1))', [DEMO_LOCK_NAME])
    const marker = await client.query("SELECT obj_description(oid, 'pg_namespace') AS marker FROM pg_namespace WHERE nspname = 'disaster_demo'")
    if (marker.rows[0]?.marker !== DEMO_SCHEMA_MARKER) throw fail('DEMO_SCHEMA_MISSING')
    const result = await work(client)
    await client.query('COMMIT')
    return result
  } catch (error) {
    await client.query('ROLLBACK').catch(() => {})
    const known = ['DEMO_SCHEMA_MISSING', 'DEMO_RECORD_MISSING', 'CLAIM_CLOSED', 'DRAFT_STALE', 'DOCUMENT_NOT_FOUND', 'DOCUMENT_LIMIT_REACHED', 'DOCUMENT_ANALYSIS_BUSY', 'INSURER_RECORD_MISSING', 'PROPOSAL_STALE', 'PROPOSAL_ALREADY_REVIEWED', 'PROVIDER_CAPACITY_CHANGED', 'POLICY_PDF_UNAVAILABLE', 'INVALID_PROPOSAL']
    throw fail(known.includes(error.code) ? error.code : 'DATABASE_UNAVAILABLE')
  } finally { await client.end().catch(() => {}) }
}

// Preflight checks only the configured record scope, without supplying identity facts to a model.
export function assertHotlineReady(config) {
  return transaction(config, async db => {
    const { rows } = await db.query(`SELECT c.id FROM disaster_demo.customers c
      JOIN disaster_demo.policies p ON p.customer_id = c.id
      JOIN disaster_demo.claims cl ON cl.policy_id = p.id
      WHERE c.id = $1 AND p.id = $2 AND cl.id = $3 AND c.verification_question IS NOT NULL`,
    [DEMO_CUSTOMER_ID, DEMO_POLICY_ID, DEMO_CLAIM_ID])
    if (!rows[0]) throw fail('DEMO_RECORD_MISSING')
  })
}

export function verificationQuestion(config, firstName, contractNumber) {
  return transaction(config, async db => {
    const { rows } = await db.query(`SELECT c.verification_question FROM disaster_demo.customers c
      JOIN disaster_demo.policies p ON p.customer_id = c.id
      WHERE c.id = $1 AND p.id = $2 AND lower(c.first_name) = lower($3) AND regexp_replace(p.contract_number, '[[:space:]-]', '', 'g') = $4`,
    [DEMO_CUSTOMER_ID, DEMO_POLICY_ID, firstName.trim(), normalizeContractNumber(contractNumber)])
    return rows[0] ? { question: rows[0].verification_question, fictional: true } : null
  })
}

export function verifyCustomer(config, answer) {
  return transaction(config, async db => {
    const { rows } = await db.query('SELECT verification_answer_hash FROM disaster_demo.customers WHERE id = $1', [DEMO_CUSTOMER_ID])
    if (!rows[0]) throw fail('DEMO_RECORD_MISSING')
    const [scheme, salt, hash] = rows[0].verification_answer_hash.split(':')
    if (scheme !== 'scrypt' || !/^[a-f0-9]{64}$/.test(hash || '') || !/^[a-f0-9]{32}$/.test(salt || '')) throw fail('DEMO_RECORD_MISSING')
    const actual = await deriveKey(answer.trim().toLowerCase(), salt, 32)
    return { verified: timingSafeEqual(actual, Buffer.from(hash, 'hex')), fictional: true }
  })
}

export function readCustomer(config) {
  return transaction(config, async db => {
    const { rows } = await db.query('SELECT id, first_name, household_size FROM disaster_demo.customers WHERE id = $1', [DEMO_CUSTOMER_ID])
    if (!rows[0]) throw fail('DEMO_RECORD_MISSING')
    return rows[0]
  })
}

export function readPolicy(config) {
  return transaction(config, async db => {
    const { rows } = await db.query(`SELECT id, customer_id, contract_number, version, product_name, valid_from::text,
      valid_to::text, address, city, insee_code, property_type, occupancy, coverage_data, general_terms_document_id, specific_terms_document_id
      FROM disaster_demo.policies WHERE id = $1 AND customer_id = $2`, [DEMO_POLICY_ID, DEMO_CUSTOMER_ID])
    if (!rows[0]) throw fail('DEMO_RECORD_MISSING')
    const policy = rows[0]
    const sections = await db.query(`SELECT s.id, s.document_id, s.title, s.clause_reference, s.text, s.language,
      s.page_start, s.page_end, d.filename, d.provenance FROM disaster_demo.document_sections s
      JOIN disaster_demo.documents d ON d.id = s.document_id WHERE s.document_id = ANY($1::uuid[]) ORDER BY s.document_id, s.position`,
    [[policy.general_terms_document_id, policy.specific_terms_document_id]])
    if (!sections.rows.length) throw fail('DEMO_RECORD_MISSING')
    // PDF exports place the schedule on page one, then up to three clauses per page.
    const projection = policyPdfProjection(policy, sections.rows)
    const pdf = await db.query(`SELECT d.id, d.filename, d.sha256, d.size_bytes FROM disaster_demo.claim_events e
      JOIN disaster_demo.documents d ON d.id::text = e.data->>'document_id'
      WHERE e.claim_id = $1 AND e.type = 'policy_pdf_stored' AND e.data->>'source_hash' = $2 AND e.data->>'sha256' = d.sha256 AND d.mime_type = 'application/pdf'
      ORDER BY e.created_at DESC LIMIT 1`, [DEMO_CLAIM_ID, policyDocumentHash(projection)])
    return { ...projection, stored_pdf: pdf.rows[0] || null }
  })
}
export async function readPolicyPdf(config) {
  const policy = await readPolicy(config)
  if (!policy.stored_pdf) throw fail('POLICY_PDF_UNAVAILABLE')
  return transaction(config, async db => {
    const { rows } = await db.query('SELECT filename, content, mime_type, sha256 FROM disaster_demo.documents WHERE id = $1 AND sha256 = $2', [policy.stored_pdf.id, policy.stored_pdf.sha256])
    if (!rows[0]) throw fail('POLICY_PDF_UNAVAILABLE')
    if (createHash('sha256').update(rows[0].content).digest('hex') !== rows[0].sha256) throw fail('POLICY_PDF_UNAVAILABLE')
    return rows[0]
  })
}

async function claimProjection(db) {
  const { rows } = await db.query(`SELECT c.*, c.updated_at::text AS version_token, p.address, p.contract_number, p.product_name, u.first_name, u.household_size,
    (SELECT data->'draft' FROM disaster_demo.claim_events WHERE claim_id = c.id AND type = 'hotline_report_saved' ORDER BY created_at DESC LIMIT 1) AS intake_details
    FROM disaster_demo.claims c JOIN disaster_demo.policies p ON p.id = c.policy_id
    JOIN disaster_demo.customers u ON u.id = p.customer_id
    WHERE c.id = $1 AND c.reference = $2 AND p.id = $3 AND u.id = $4 FOR UPDATE OF c`,
  [DEMO_CLAIM_ID, DEMO_CLAIM_REFERENCE, DEMO_POLICY_ID, DEMO_CUSTOMER_ID])
  if (!rows[0]) throw fail('DEMO_RECORD_MISSING')
  return rows[0]
}
export const readClaim = config => transaction(config, claimProjection)

// Deterministic operation identity lives in claim_events; no schema migration is needed.
// This updates the reserved record only and never inserts a second household/claim.
export function commitReport(config, draft, operationId, conversationId, language) {
  return transaction(config, async db => {
    const current = await claimProjection(db)
    const prior = await db.query(`SELECT id FROM disaster_demo.claim_events
      WHERE claim_id = $1 AND type = 'hotline_report_saved' AND data->>'operation_id' = $2`, [DEMO_CLAIM_ID, operationId])
    if (prior.rows.length) return { claim: current, replayed: true }
    if (current.closed_at) throw fail('CLAIM_CLOSED')
    if (current.version_token !== draft.baseVersion) throw fail('DRAFT_STALE')
    const needs = [...current.needs.filter(item => !draft.needs.includes(item.code)).map(item => item.status === 'identified' ? { ...item, status: 'to_confirm' } : item),
      ...draft.needs.map(code => ({ code, status: 'identified', quantity: code === 'accommodation' ? 1 : null }))]
    const required = [...new Set([...current.required_documents, ...draft.documents])]
    const accommodation = draft.needs.includes('accommodation') && draft.accommodationNeeded
    const urgent = accommodation && draft.habitability !== 'habitable'
    const priority = urgent ? 'P1' : current.priority || 'P2'
    const priorityReasons = urgent ? [draft.habitability === 'uninhabitable'
      ? 'The household cannot return home and has no lasting accommodation solution.'
      : 'An accommodation need was reported; habitability and access require review.'] : current.priority_reasons
    // Access can remain unknown or become restricted from intake; a voice call never authorizes site access.
    const access = draft.access === 'restricted' ? 'restricted' : current.access_status
    await db.query(`UPDATE disaster_demo.claims SET intake_status = 'reported',
      reported_at = COALESCE(reported_at, now()),
      description = $2, summary = $2, habitability = $3, access_status = $4,
      needs = $5::jsonb, required_documents = $6::jsonb, priority = $7,
      priority_reasons = $8::jsonb, updated_at = now() WHERE id = $1`,
    [DEMO_CLAIM_ID, draft.description, draft.habitability, access, JSON.stringify(needs), JSON.stringify(required), priority,
      JSON.stringify(priorityReasons)])
    await db.query(`INSERT INTO disaster_demo.claim_events (id, claim_id, type, actor, data)
      VALUES ($1, $2, 'hotline_report_saved', 'client', $3::jsonb)`,
    [randomUUID(), DEMO_CLAIM_ID, JSON.stringify({ operation_id: operationId, fictional: true, draft,
      confirmation: 'explicit_browser_review', customer_id: DEMO_CUSTOMER_ID, policy_id: DEMO_POLICY_ID })])
    await db.query(`INSERT INTO disaster_demo.messages (id, customer_id, claim_id, conversation_id, channel, role, content, language)
      VALUES ($1, $2, $3, $4, 'voice', 'client', $5, $6)`,
    [randomUUID(), DEMO_CUSTOMER_ID, DEMO_CLAIM_ID, conversationId, draft.description, language])
    return { claim: await claimProjection(db), replayed: false }
  })
}
