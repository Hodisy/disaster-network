import WebSocket from 'ws'

// Live-only adapter. Never silently substitute /realtime or expose upstream bodies/keys.
export async function liveRequest(config, path, body) {
  const key = config.openaiApiKey || process.env.OPENAI_API_KEY
  if (!key) throw Object.assign(new Error('OPENAI_NOT_CONFIGURED'), { code: 'OPENAI_NOT_CONFIGURED', statusCode: 503 })
  let response
  try {
    response = await fetch(`https://api.openai.com/v1/live/sessions${path}`, {
      method: 'POST', headers: { Authorization: `Bearer ${key}`, 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body), signal: AbortSignal.timeout(25000)
    })
  } catch { throw Object.assign(new Error('OPENAI_CONNECTION_FAILED'), { code: 'OPENAI_CONNECTION_FAILED', statusCode: 502 }) }
  if (!response.ok) {
    const code = [401, 403, 404].includes(response.status) ? 'OPENAI_LIVE_ACCESS_UNAVAILABLE' : response.status === 429 ? 'OPENAI_RATE_LIMITED' : 'OPENAI_CONNECTION_FAILED'
    throw Object.assign(new Error(code), { code, statusCode: 502 })
  }
  if (response.status === 204 || path.endsWith('/hangup')) return null
  try { return await response.json() } catch { throw Object.assign(new Error('OPENAI_PROTOCOL_ERROR'), { code: 'OPENAI_PROTOCOL_ERROR' }) }
}

export function attachLive(config, sessionId, onEvent, onFailure) {
  return new Promise((resolve, reject) => {
    const socket = new WebSocket(`wss://api.openai.com/v1/live/sessions/${encodeURIComponent(sessionId)}/attach`, {
      headers: { Authorization: `Bearer ${config.openaiApiKey || process.env.OPENAI_API_KEY}` },
      handshakeTimeout: 12000, maxPayload: 2 * 1024 * 1024
    })
    const timeout = setTimeout(() => { socket.terminate(); reject(Object.assign(new Error('OPENAI_SIDEBAND_FAILED'), { code: 'OPENAI_SIDEBAND_FAILED' })) }, 13000)
    socket.once('open', () => { clearTimeout(timeout); resolve(socket) })
    socket.on('message', data => {
      try { onEvent(JSON.parse(data.toString())) } catch { onFailure('OPENAI_PROTOCOL_ERROR') }
    })
    socket.on('error', () => { clearTimeout(timeout); onFailure('OPENAI_SIDEBAND_FAILED'); reject(Object.assign(new Error('OPENAI_SIDEBAND_FAILED'), { code: 'OPENAI_SIDEBAND_FAILED' })) })
    socket.on('close', () => { clearTimeout(timeout); onFailure('OPENAI_DISCONNECTED') })
  })
}
