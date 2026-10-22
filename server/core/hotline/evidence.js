import { createHash, randomUUID } from 'node:crypto'
import { PDFDocument } from 'pdf-lib'
import * as evidence from '../../repositories/claimEvidence.js'
import { readClaim, readPolicy, readPolicyPdf } from '../../repositories/hotlineClaims.js'
import { extractEvidence } from '../../services/evidenceAnalysis.js'
import { lookupCatNat, CATNAT_SOURCE_URL } from '../../services/georisques.js'
import { normalizeContractNumber } from '../../../app/config/demo.js'

const fail = (code, statusCode = 400) => { throw Object.assign(new Error(code), { code, statusCode }) }
const uuid = value => typeof value === 'string' && /^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i.test(value)
export const policyHash = policy => createHash('sha256').update(JSON.stringify({ version: policy.version, coverage: policy.coverage_data,
  validity: [policy.valid_from, policy.valid_to], sections: policy.sections.map(s => [s.id, s.text]) })).digest('hex')

export async function uploadEvidence(config, file, kind, locale) {
  if (!file || !['photos', 'address', 'inventory', 'invoice', 'other'].includes(kind)) fail('INVALID_REQUEST')
  const content = Buffer.from(await file.arrayBuffer())
  if (!content.length || content.length > 10 * 1024 * 1024) fail('DOCUMENT_INVALID_SIZE', 413)
  let mime, extension
  if (content.subarray(0, 5).toString() === '%PDF-') { mime = 'application/pdf'; extension = 'pdf' }
  else if (content.subarray(0, 8).equals(Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]))) { mime = 'image/png'; extension = 'png' }
  else if (content[0] === 255 && content[1] === 216 && content[2] === 255) { mime = 'image/jpeg'; extension = 'jpg' }
  else if (content.subarray(0, 4).toString() === 'RIFF' && content.subarray(8, 12).toString() === 'WEBP') { mime = 'image/webp'; extension = 'webp' }
  else fail('DOCUMENT_FORMAT_UNSUPPORTED')
  if (kind === 'photos' && mime === 'application/pdf') fail('PHOTO_IMAGE_REQUIRED')
  if (mime === 'application/pdf') {
    try {
      const pdf = await PDFDocument.load(content, { updateMetadata: false })
      if (pdf.getPageCount() < 1 || pdf.getPageCount() > 12) fail('DOCUMENT_PDF_PAGE_LIMIT')
    } catch (error) { fail(error.code === 'DOCUMENT_PDF_PAGE_LIMIT' ? error.code : 'DOCUMENT_PDF_UNREADABLE') }
  }
  // Do not trust path components, supplied MIME types or executable filename extensions.
  const base = String(file.name || 'supporting-document').split(/[\\/]/).pop().replace(/[\u0000-\u001f\u007f]/g, '').replace(/\.[^.]*$/, '').slice(0, 120)
  const filename = `${base || 'supporting-document'}.${extension}`
  return evidence.storeEvidence(config, { content, mime, filename, kind, sha256: createHash('sha256').update(content).digest('hex') }, locale)
}

export function compareEvidence(policy, claim, facts) {
  const eventDate = new Date(claim.incident_at).toISOString().slice(0, 10)
  const policyActive = eventDate >= policy.valid_from && eventDate <= policy.valid_to
  const checks = [{ label: 'Policy active on the reported event date', status: policyActive ? 'consistent' : 'needs_review' }]
  const section = policy.sections.find(item => item.clause_reference === 'HP-ACCOMMODATION')
  if (facts.document_type === 'damage_photo') {
    checks.push({ label: 'Visible damage described; cause, date and location require review', status: 'needs_review' })
  } else if (facts.document_type === 'accommodation_invoice') {
    const coverage = policy.coverage_data
    const ceiling = coverage.benefits?.find(item => item.code === 'accommodation')?.limit_cents
    const validMoney = facts.currency === 'EUR' && facts.invoice_total_cents !== null
    const datedNights = facts.stay_start && facts.stay_end ? (Date.parse(facts.stay_end) - Date.parse(facts.stay_start)) / 86400000 : null
    const nights = facts.nights ?? datedNights
    const validNights = Number.isInteger(nights) && nights > 0 && (facts.nights === null || datedNights === null || facts.nights === datedNights)
    const knownLimit = Number.isSafeInteger(ceiling) && ceiling >= 0
    checks.push({ label: 'Invoice total compared with the accommodation limit',
      status: validMoney && knownLimit && facts.invoice_total_cents <= ceiling ? 'consistent' : 'needs_review',
      amountCents: facts.invoice_total_cents, limitCents: knownLimit ? ceiling : null })
    checks.push({ label: 'Stay length compared with the policy limit',
      status: validNights && nights <= coverage.accommodation_max_nights ? 'consistent' : 'needs_review', nights, maxNights: coverage.accommodation_max_nights })
    checks.push({ label: 'Nightly cost compared with the household limit',
      status: validMoney && validNights && Number.isSafeInteger(coverage.accommodation_max_nightly_cents) && facts.invoice_total_cents <= nights * coverage.accommodation_max_nightly_cents ? 'consistent' : 'needs_review' })
    const declaredStart = claim.intake_details?.accommodationStart
    const declaredNights = claim.intake_details?.accommodationNights
    checks.push({ label: 'Stay dates reconciled with the declared request',
      status: facts.stay_start && facts.stay_end && facts.stay_start >= eventDate && facts.stay_end > facts.stay_start && facts.stay_start === declaredStart && validNights && nights === declaredNights ? 'consistent' : 'needs_review' })
    checks.push({ label: 'Main residence unusability and accommodation need declared',
      status: claim.habitability === 'uninhabitable' && claim.intake_details?.accommodationNeeded === true ? 'consistent' : 'needs_review' })
    checks.push({ label: 'Invoice recipient reconciled with the policyholder',
      status: facts.holder_first_name?.trim().toLowerCase() === claim.first_name?.toLowerCase() ? 'consistent' : 'needs_review' })
  } else checks.push({ label: 'Document requires a handler review', status: 'needs_review' })
  if (facts.policy_reference !== null) checks.push({ label: 'Policy reference reconciled with the dossier', status: normalizeContractNumber(facts.policy_reference) === policy.contract_number ? 'consistent' : 'needs_review' })
  // A consistent numerical comparison is never a claim eligibility or authenticity decision.
  return { status: 'needs_review', coverageConfirmed: false, checks,
    citation: section && facts.document_type === 'accommodation_invoice'
      ? { clause: section.clause_reference, sectionId: section.id, page: section.pdf_page, filename: policy.pdf_filename } : null,
    policyVersion: policy.version, policyHash: policyHash(policy), claimVersion: claim.version_token }
}

export async function analyzeEvidence(config, documentId, locale, consent) {
  if (!uuid(documentId) || consent !== true) fail('EVIDENCE_CONSENT_REQUIRED')
  const current = await evidence.listEvidence(config)
  const stored = current.documents.find(item => item.id === documentId)
  if (!stored) fail('DOCUMENT_NOT_FOUND', 404)
  const policy = await readPolicy(config)
  const currentClaim = await readClaim(config)
  if (stored.analysis?.comparison?.policyHash === policyHash(policy) && stored.analysis?.comparison?.claimVersion === currentClaim.version_token && stored.analysis?.locale === locale) return { analysis: stored.analysis, replayed: true }
  const operationId = randomUUID()
  await evidence.beginAnalysis(config, documentId, operationId)
  const document = await evidence.readEvidence(config, documentId)
  let analysis
  try {
    const facts = await extractEvidence(config, document, locale)
    const claim = await readClaim(config)
    analysis = { facts, locale, checkedAt: new Date().toISOString(), sourceSha256: document.sha256,
      comparison: compareEvidence(policy, claim, facts) }
  } catch (error) {
    await evidence.finishAnalysis(config, document, operationId, null).catch(() => {})
    throw error
  }
  await evidence.finishAnalysis(config, document, operationId, analysis)
  return { analysis, replayed: false }
}
export async function checkOfficialSource(config) {
  const previous = await evidence.listEvidence(config)
  const policy = await readPolicy(config)
  const claim = await readClaim(config)
  const incidentDate = new Date(claim.incident_at).toISOString().slice(0, 10)
  const old = previous.officialSource
  if (old && old.inseeCode === policy.insee_code && old.incidentDate === incidentDate && Date.now() - Date.parse(old.checkedAt) < 300000) return old
  let result
  try { result = await lookupCatNat(policy.insee_code, incidentDate) }
  catch {
    result = { provider: 'Géorisques · GASPAR', sourceUrl: CATNAT_SOURCE_URL, inseeCode: policy.insee_code,
      incidentDate, checkedAt: new Date().toISOString(), status: 'unavailable', records: [], complete: false, coverageConfirmed: false }
  }
  return evidence.storeOfficialSource(config, result)
}
export async function readClaimEvidence(config) {
  const result = await evidence.listEvidence(config)
  const policy = await readPolicy(config)
  const claim = await readClaim(config)
  return { ...result, documents: result.documents.map(document => ({ ...document, analysis: document.analysis ? {
    ...document.analysis, stale: document.analysis.comparison.policyHash !== policyHash(policy) || document.analysis.comparison.claimVersion !== claim.version_token
  } : null })) }
}
export async function evidenceAction(config, action, body, locale) {
  if (action === 'list') return readClaimEvidence(config)
  if (action === 'analyze') return analyzeEvidence(config, body.documentId, locale, body.consent)
  if (action === 'source') return { officialSource: await checkOfficialSource(config) }
  if (action === 'contract') {
    const pdf = await readPolicyPdf(config)
    return { binary: pdf.content, mime: pdf.mime_type, filename: pdf.filename }
  }
  if (action === 'download') {
    if (!uuid(body.documentId)) fail('INVALID_REQUEST')
    const document = await evidence.readEvidence(config, body.documentId)
    return { binary: document.content, mime: document.mime_type, filename: document.filename }
  }
  fail('INVALID_REQUEST', 404)
}
