// Shared export layout references, never original-source PDF page numbers.
import { createHash } from 'node:crypto'
export function policyDocumentHash(policy) {
  return createHash('sha256').update(JSON.stringify({ id: policy.id, contract: policy.contract_number,
    product: policy.product_name, address: policy.address, version: policy.version, validity: [policy.valid_from, policy.valid_to],
    coverage: policy.coverage_data, sections: policy.sections.map(s => [s.id, s.document_id, s.clause_reference, s.text, s.pdf_page]) })).digest('hex')
}
export function policyPdfProjection(policy, sections) {
  const ordered = [...sections].sort((a, b) => Number(b.document_id === policy.specific_terms_document_id) - Number(a.document_id === policy.specific_terms_document_id))
  return { ...policy, pdf_filename: `allstorm-home-protect-${policy.contract_number}.pdf`, sections: ordered.map((section, index) => ({ ...section, pdf_page: index === 0 ? 1 : 2 + Math.floor((index - 1) / 3) })) }
}
