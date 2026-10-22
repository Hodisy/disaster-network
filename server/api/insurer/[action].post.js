import { checkOrigin, limit } from '../../core/hotline/sessions.js'
import { authorizeInsurer, requireInsurer } from '../../core/insurer/access.js'
import { insurerAction } from '../../core/insurer/agent.js'

export default defineEventHandler(async event => {
  setResponseHeader(event, 'Cache-Control', 'no-store')
  setResponseHeader(event, 'X-Content-Type-Options', 'nosniff')
  try {
    const config = useRuntimeConfig(event), ip = getRequestIP(event) || 'unknown'
    checkOrigin(config, getHeader(event, 'origin'), getHeader(event, 'sec-fetch-site'))
    if (!getHeader(event, 'content-type')?.startsWith('application/json') || getHeader(event, 'x-insurer-request') !== '1') throw Object.assign(new Error(), { code: 'INVALID_REQUEST' })
    limit(`insurer:${ip}`, 120, 60000)
    const chunks = []; let bytes = 0
    for await (const chunk of event.node.req) { bytes += Buffer.byteLength(chunk); if (bytes > 8192) throw Object.assign(new Error(), { code: 'INVALID_REQUEST' }); chunks.push(Buffer.from(chunk)) }
    let body
    try { body = JSON.parse(Buffer.concat(chunks).toString()) } catch { throw Object.assign(new Error(), { code: 'INVALID_REQUEST' }) }
    if (!body || Array.isArray(body) || typeof body !== 'object') throw Object.assign(new Error(), { code: 'INVALID_REQUEST' })
    const action = getRouterParam(event, 'action')
    if (action === 'authorize') {
      limit(`insurer-auth:${ip}`, 8, 600000)
      const cookie = authorizeInsurer(config, body.token)
      setCookie(event, 'disaster-insurer', cookie, { httpOnly: true, secure: config.hotlineOrigin.startsWith('https://'), sameSite: 'strict', path: '/api/insurer', maxAge: 1800 })
      return { ok: true }
    }
    requireInsurer(config, getCookie(event, 'disaster-insurer'))
    if (action === 'access') return { ok: true, authorized: true }
    if (action === 'analyze') limit(`insurer-analysis:${ip}`, 6, 600000)
    const result = await insurerAction(config, action, body)
    requireInsurer(config, getCookie(event, 'disaster-insurer'))
    if (result.binary) {
      setResponseHeader(event, 'Content-Type', result.mime)
      setResponseHeader(event, 'Content-Disposition', `attachment; filename="policy-document.pdf"; filename*=UTF-8''${encodeURIComponent(result.filename).replace(/['()*]/g, c => `%${c.charCodeAt(0).toString(16)}`)}`)
      return result.binary
    }
    return { ok: true, ...result }
  } catch (error) {
    const allowed = ['INSURER_AGENT_DISABLED', 'INSURER_NOT_CONFIGURED', 'INSURER_ACCESS_DENIED', 'HOTLINE_ACCESS_DENIED', 'HOTLINE_RATE_LIMITED', 'DATABASE_NOT_CONFIGURED', 'DATABASE_UNAVAILABLE', 'DEMO_SCHEMA_MISSING', 'INSURER_RECORD_MISSING', 'INSURER_REPORT_REQUIRED', 'POLICY_PDF_UNAVAILABLE', 'OPENAI_NOT_CONFIGURED', 'OPENAI_RATE_LIMITED', 'INSURER_AI_UNAVAILABLE', 'INSURER_TOOL_CONTEXT_MISSING', 'INSURER_TOOL_LIMIT', 'INVALID_PROPOSAL', 'PROPOSAL_STALE', 'PROPOSAL_ALREADY_REVIEWED', 'PROVIDER_CAPACITY_CHANGED', 'CLAIM_CLOSED', 'EVIDENCE_CONSENT_REQUIRED', 'DOCUMENT_NOT_FOUND', 'DOCUMENT_FORMAT_UNSUPPORTED', 'INVALID_REQUEST']
    setResponseStatus(event, error.statusCode || 503)
    return { ok: false, code: allowed.includes(error.code) ? error.code : 'INSURER_FAILED' }
  }
})
