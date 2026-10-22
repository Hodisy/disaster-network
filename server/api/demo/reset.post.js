import { resetDemoDatabase, DemoResetError } from '../../core/resetDemoDatabase.js'

export default defineEventHandler(async event => {
  setResponseHeader(event, 'Cache-Control', 'no-store')
  if (getHeader(event, 'sec-fetch-site') === 'cross-site') {
    setResponseStatus(event, 403)
    return { ok: false, code: 'INVALID_REQUEST' }
  }
  if (!getHeader(event, 'content-type')?.startsWith('application/json')) {
    setResponseStatus(event, 400)
    return { ok: false, code: 'INVALID_REQUEST' }
  }
  let body
  try { body = await readBody(event) }
  catch {
    setResponseStatus(event, 400)
    return { ok: false, code: 'INVALID_REQUEST' }
  }
  try {
    const result = await resetDemoDatabase({
      config: useRuntimeConfig(event),
      maintenanceKey: getHeader(event, 'x-demo-reset-key'),
      confirmation: body?.confirmation
    })
    return { ok: true, ...result }
  } catch (error) {
    const known = ['DATABASE_NOT_CONFIGURED', 'DEMO_SCHEMA_MISSING', 'DEMO_SCHEMA_MISMATCH', 'DATABASE_RESET_FAILED']
    const code = error instanceof DemoResetError || known.includes(error.code) ? error.code : 'DATABASE_RESET_FAILED'
    setResponseStatus(event, error instanceof DemoResetError ? error.statusCode : 503)
    return { ok: false, code }
  }
})
