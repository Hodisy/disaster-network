import { mkdir, writeFile } from 'node:fs/promises'
import { createDemoDataset } from '../database/fixtures.js'
import { DEMO_POLICY_ID } from '../app/config/demo.js'
import { renderPolicyPdf } from '../server/services/policyPdf.js'
import { policyPdfProjection } from '../server/utils/policyPdfReferences.js'

// Offline artifact generation only; no environment secrets, database or AI calls.
const fixture = createDemoDataset()
const policy = fixture.policies.find(item => item.id === DEMO_POLICY_ID)
const sections = fixture.document_sections.filter(item => [policy.general_terms_document_id, policy.specific_terms_document_id].includes(item.document_id))
const source = policyPdfProjection(policy, sections)
const folder = new URL('../output/pdf/', import.meta.url)
await mkdir(folder, { recursive: true })
await writeFile(new URL(source.pdf_filename, folder), await renderPolicyPdf(source))
console.log(`Created output/pdf/${source.pdf_filename} from the offline fixture. The app renders current database clauses separately.`)
