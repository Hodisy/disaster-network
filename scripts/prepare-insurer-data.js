import { createHash, randomUUID } from 'node:crypto'
import { createInsurerDataset } from '../database/insurerFixture.js'
import { buildDemoSeed } from '../database/operations.js'
import { transaction } from '../server/repositories/hotlineClaims.js'
import { policyPdfProjection, policyDocumentHash } from '../server/utils/policyPdfReferences.js'
import { renderPolicyPdf } from '../server/services/policyPdf.js'

// Explicit user-authorized additive preparation. No reset, AI call or secret output.
try {
  const result = await transaction({}, async db => {
    const protectedRows = await db.query(`SELECT 'claims' AS kind, id, md5(to_jsonb(c)::text) AS hash FROM disaster_demo.claims c
      UNION ALL SELECT 'messages', id, md5(to_jsonb(m)::text) FROM disaster_demo.messages m
      UNION ALL SELECT 'assignments', id, md5(to_jsonb(a)::text) FROM disaster_demo.assignments a
      UNION ALL SELECT 'documents', id, md5(to_jsonb(d)::text) FROM disaster_demo.documents d`)
    await db.query(buildDemoSeed(createInsurerDataset()))
    const policies = (await db.query('SELECT *, valid_from::text AS valid_from, valid_to::text AS valid_to FROM disaster_demo.policies ORDER BY contract_number')).rows
    let insertedPdfs = 0
    for (const policy of policies) {
      const sections = (await db.query('SELECT * FROM disaster_demo.document_sections WHERE document_id = ANY($1::uuid[]) ORDER BY document_id, position', [[policy.general_terms_document_id, policy.specific_terms_document_id]])).rows
      const source = policyPdfProjection(policy, sections)
      const sourceHash = policyDocumentHash(source)
      const claim = (await db.query('SELECT id FROM disaster_demo.claims WHERE policy_id = $1 ORDER BY incident_at LIMIT 1', [policy.id])).rows[0]
      if (!claim) continue
      const previous = await db.query(`SELECT id FROM disaster_demo.claim_events WHERE claim_id = $1 AND type = 'policy_pdf_stored' AND data->>'source_hash' = $2`, [claim.id, sourceHash])
      if (previous.rows.length) continue
      const content = await renderPolicyPdf(source), documentId = randomUUID(), sha = createHash('sha256').update(content).digest('hex')
      await db.query(`INSERT INTO disaster_demo.documents (id, claim_id, kind, filename, mime_type, size_bytes, content, sha256, language, provenance, extraction_status)
        VALUES ($1, NULL, 'endorsement', $2, 'application/pdf', $3, $4, $5, 'en', 'synthetic', 'complete')`, [documentId, source.pdf_filename, content.length, content, sha])
      await db.query(`INSERT INTO disaster_demo.claim_events (id, claim_id, type, actor, data) VALUES ($1, $2, 'policy_pdf_stored', 'demo_setup', $3::jsonb)`,
      [randomUUID(), claim.id, JSON.stringify({ policy_id: policy.id, document_id: documentId, source_hash: sourceHash, sha256: sha, fictional: true })])
      insertedPdfs++
    }
    for (const table of ['claims', 'messages', 'assignments', 'documents']) {
      const expected = protectedRows.rows.filter(row => row.kind === table)
      if (!expected.length) continue
      const rows = (await db.query(`SELECT id, md5(to_jsonb(t)::text) AS hash FROM disaster_demo.${table} t WHERE id = ANY($1::uuid[])`, [expected.map(row => row.id)])).rows
      if (rows.length !== expected.length || rows.some(row => expected.find(before => before.id === row.id)?.hash !== row.hash)) throw new Error('Existing work changed')
    }
    const counts = (await db.query(`SELECT (SELECT count(*)::int FROM disaster_demo.claims) AS claims,
      (SELECT sum(household_size)::int FROM disaster_demo.customers) AS people,
      (SELECT count(*)::int FROM disaster_demo.providers) AS providers,
      (SELECT count(*)::int FROM disaster_demo.documents WHERE mime_type = 'application/pdf' AND provenance = 'synthetic') AS policy_pdfs`)).rows[0]
    return { ...counts, insertedPdfs, existingWorkPreserved: true, resetPerformed: false }
  })
  console.log(JSON.stringify(result))
} catch { console.error('Insurer preparation failed; the transaction was rolled back. No credentials or SQL details are printed.'); process.exitCode = 1 }
