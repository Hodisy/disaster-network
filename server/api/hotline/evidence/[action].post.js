import { checkOrigin, ownerFor, assertEvidenceAccess, limit } from '../../../core/hotline/sessions.js'
import { uploadEvidence, evidenceAction } from '../../../core/hotline/evidence.js'

export default defineEventHandler(async event => {
  setResponseHeader(event, 'Cache-Control', 'no-store')
  setResponseHeader(event, 'X-Content-Type-Options', 'nosniff')
  try {
    const config = useRuntimeConfig(event)
    checkOrigin(config, getHeader(event, 'origin'), getHeader(event, 'sec-fetch-site'))
    if (getHeader(event, 'x-hotline-request') !== '1') throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 })
    const owner = ownerFor(config, getCookie(event, 'disaster-hotline'))
    assertEvidenceAccess(owner)
    const action = getRouterParam(event, 'action')
    const ip = getRequestIP(event) || 'unknown'
    limit(`evidence:${ip}`, 100, 60000)
    if (['upload', 'analyze', 'source', 'contract'].includes(action)) limit(`evidence-work:${ip}`, 30, 60000)
    const contentType = getHeader(event, 'content-type') || ''
    const max = action === 'upload' ? 10 * 1024 * 1024 + 65536 : 4096
    if (Number(getHeader(event, 'content-length') || 0) > max) throw Object.assign(new Error(), { code: 'DOCUMENT_INVALID_SIZE', statusCode: 413 })
    const chunks = []; let size = 0
    for await (const chunk of event.node.req) {
      size += Buffer.byteLength(chunk)
      if (size > max) throw Object.assign(new Error(), { code: 'DOCUMENT_INVALID_SIZE', statusCode: 413 })
      chunks.push(Buffer.from(chunk))
    }
    const raw = Buffer.concat(chunks)
    let result
    if (action === 'upload') {
      if (!contentType.startsWith('multipart/form-data;')) throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 })
      let form
      try { form = await new Request('http://localhost', { method: 'POST', headers: { 'content-type': contentType }, body: raw }).formData() }
      catch { throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 }) }
      if (form.getAll('file').length !== 1 || typeof form.get('file')?.arrayBuffer !== 'function' || [...form.keys()].some(key => !['file', 'kind', 'locale'].includes(key))) throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 })
      result = await uploadEvidence(config, form.get('file'), form.get('kind'), form.get('locale') === 'fr' ? 'fr' : 'en')
    } else {
      if (!contentType.startsWith('application/json')) throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 })
      let body
      try { body = JSON.parse(raw.toString('utf8')) } catch { throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 }) }
      if (!body || Array.isArray(body) || typeof body !== 'object') throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 })
      result = await evidenceAction(config, action, body, body.locale === 'fr' ? 'fr' : 'en')
    }
    // A lease expiring during processing cannot disclose document contents to the browser.
    assertEvidenceAccess(owner)
    if (result.binary) {
      setResponseHeader(event, 'Content-Type', result.mime)
      setResponseHeader(event, 'Content-Disposition', `attachment; filename="supporting-document${result.mime === 'application/pdf' ? '.pdf' : ''}"; filename*=UTF-8''${encodeURIComponent(result.filename).replace(/['()*]/g, character => `%${character.charCodeAt(0).toString(16).toUpperCase()}`)}`)
      return result.binary
    }
    return { ok: true, ...result }
  } catch (error) {
    const permitted = new Set(['HOTLINE_ACCESS_DENIED', 'HOTLINE_DISABLED', 'HOTLINE_NOT_CONFIGURED', 'HOTLINE_RATE_LIMITED', 'CLAIM_ACCESS_EXPIRED', 'INVALID_REQUEST', 'CLAIM_CLOSED', 'DATABASE_NOT_CONFIGURED', 'DATABASE_UNAVAILABLE', 'DEMO_SCHEMA_MISSING', 'DEMO_RECORD_MISSING', 'DOCUMENT_NOT_FOUND', 'DOCUMENT_LIMIT_REACHED', 'DOCUMENT_INVALID_SIZE', 'DOCUMENT_FORMAT_UNSUPPORTED', 'PHOTO_IMAGE_REQUIRED', 'DOCUMENT_PDF_PAGE_LIMIT', 'DOCUMENT_PDF_UNREADABLE', 'EVIDENCE_CONSENT_REQUIRED', 'DOCUMENT_ANALYSIS_BUSY', 'EVIDENCE_ANALYSIS_UNAVAILABLE', 'OPENAI_NOT_CONFIGURED', 'OPENAI_RATE_LIMITED', 'POLICY_PDF_UNAVAILABLE'])
    setResponseStatus(event, error.statusCode || 503)
    return { ok: false, code: permitted.has(error.code) ? error.code : 'EVIDENCE_FAILED' }
  }
})
