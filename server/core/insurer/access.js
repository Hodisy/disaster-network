import { createHash, createHmac, randomBytes, timingSafeEqual } from 'node:crypto'
const fail = code => { throw Object.assign(new Error(code), { code, statusCode: 403 }) }
const hash = value => createHash('sha256').update(value).digest()
function secret(config) {
  if (![true, 'true'].includes(config.insurerAgentEnabled)) fail('INSURER_AGENT_DISABLED')
  if (typeof config.insurerAccessToken !== 'string' || config.insurerAccessToken.length < 32 || config.insurerAccessToken.length > 256) fail('INSURER_NOT_CONFIGURED')
  return config.insurerAccessToken
}
export function authorizeInsurer(config, token) {
  const key = secret(config)
  if (typeof token !== 'string' || token.length > 256 || !timingSafeEqual(hash(token), hash(key))) fail('INSURER_ACCESS_DENIED')
  const payload = Buffer.from(JSON.stringify({ role: 'insurer', expires: Date.now() + 1800000, nonce: randomBytes(16).toString('hex') })).toString('base64url')
  return `${payload}.${createHmac('sha256', key).update(payload).digest('hex')}`
}
export function requireInsurer(config, cookie) {
  const key = secret(config)
  if (typeof cookie !== 'string' || cookie.length > 1024) fail('INSURER_ACCESS_DENIED')
  const [payload, signature, extra] = cookie.split('.')
  if (extra || !/^[a-f0-9]{64}$/.test(signature || '') || !timingSafeEqual(Buffer.from(signature, 'hex'), createHmac('sha256', key).update(payload).digest())) fail('INSURER_ACCESS_DENIED')
  let grant
  try { grant = JSON.parse(Buffer.from(payload, 'base64url').toString()) } catch { fail('INSURER_ACCESS_DENIED') }
  if (grant.role !== 'insurer' || !(grant.expires > Date.now())) fail('INSURER_ACCESS_DENIED')
  return grant
}
