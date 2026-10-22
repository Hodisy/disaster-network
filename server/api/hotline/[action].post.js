import { authorize, ownerFor, checkOrigin, limit, createSession, snapshot, confirmReport, readAuthorizedClaim, closeSession, interrupt, changeLanguage } from '../../core/hotline/sessions.js'

export default defineEventHandler(async event => {
  setResponseHeader(event, 'Cache-Control', 'no-store')
  setResponseHeader(event, 'X-Content-Type-Options', 'nosniff')
  try {
    const config = useRuntimeConfig(event)
    const ip = getRequestIP(event) || 'unknown'
    checkOrigin(config, getHeader(event, 'origin'), getHeader(event, 'sec-fetch-site'))
    limit(`http:${ip}`, 240, 60000)
    if (!getHeader(event, 'content-type')?.startsWith('application/json') || getHeader(event, 'x-hotline-request') !== '1') throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 })
    const chunks = []
    let size = 0
    for await (const chunk of event.node.req) {
      size += Buffer.byteLength(chunk)
      if (size > 65536) throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 413 })
      chunks.push(Buffer.from(chunk))
    }
    let body
    try { body = JSON.parse(Buffer.concat(chunks).toString('utf8')) }
    catch { throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 400 }) }
    const action = getRouterParam(event, 'action')
    if (action === 'authorize') {
      const capability = authorize(config, body?.token, ip)
      setCookie(event, 'disaster-hotline', capability, { httpOnly: true, secure: config.hotlineOrigin.startsWith('https://'), sameSite: 'strict', path: '/api/hotline', maxAge: 3600 })
      return { ok: true }
    }
    const owner = ownerFor(config, getCookie(event, 'disaster-hotline'))
    let result
    if (action === 'access') result = { authorized: true }
    else if (action === 'session') result = await createSession(owner, config, body, ip)
    else if (action === 'status') result = snapshot(owner)
    else if (action === 'claim') result = await readAuthorizedClaim(owner, config)
    else if (action === 'confirm') result = await confirmReport(owner, config, body)
    else if (action === 'close') result = await closeSession(owner, config, body?.sessionId)
    else if (action === 'interrupt') result = interrupt(owner)
    else if (action === 'language') result = changeLanguage(owner, body?.locale)
    else throw Object.assign(new Error(), { code: 'INVALID_REQUEST', statusCode: 404 })
    return { ok: true, ...result }
  } catch (error) {
    // Never return framework/SQL/upstream exception messages or stack traces.
    const code = /^[A-Z_]+$/.test(error.code || '') ? error.code : 'HOTLINE_FAILED'
    const permitted = new Set(['HOTLINE_ACCESS_DENIED', 'HOTLINE_DISABLED', 'HOTLINE_NOT_CONFIGURED', 'HOTLINE_RATE_LIMITED', 'HOTLINE_ALREADY_ACTIVE', 'INVALID_REQUEST', 'SESSION_CLOSED', 'CUSTOMER_VERIFICATION_REQUIRED', 'CLAIM_ACCESS_EXPIRED', 'DRAFT_STALE', 'CLAIM_CLOSED', 'BROWSER_CONFIRMATION_REQUIRED', 'DATABASE_NOT_CONFIGURED', 'DATABASE_UNAVAILABLE', 'DEMO_SCHEMA_MISSING', 'DEMO_RECORD_MISSING', 'OPENAI_NOT_CONFIGURED', 'OPENAI_CONNECTION_FAILED', 'OPENAI_LIVE_ACCESS_UNAVAILABLE', 'OPENAI_RATE_LIMITED', 'OPENAI_PROTOCOL_ERROR', 'OPENAI_SIDEBAND_FAILED', 'OPENAI_DISCONNECTED'])
    setResponseStatus(event, error.statusCode || 503)
    return { ok: false, code: permitted.has(code) ? code : 'HOTLINE_FAILED' }
  }
})
