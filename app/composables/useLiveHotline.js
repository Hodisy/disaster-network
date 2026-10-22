import { hotlineRequest } from './useHotlineClaim.js'

const errorPhrases = {
  HOTLINE_DISABLED: 'Live voice is disabled. Enable it in the private server configuration.',
  HOTLINE_NOT_CONFIGURED: 'Set the allowed application origin on the server.',
  HOTLINE_ACCESS_DENIED: 'Voice access could not be prepared. Reload the page and check that its address matches the configured application origin.',
  HOTLINE_RATE_LIMITED: 'Too many requests. Wait ten minutes before trying again.',
  HOTLINE_ALREADY_ACTIVE: 'A voice session is already active. End it before starting another.',
  DATABASE_NOT_CONFIGURED: 'PostgreSQL is not configured. Set the database URL and verified TLS settings.',
  DATABASE_UNAVAILABLE: 'PostgreSQL is unavailable. Nothing was saved.',
  DEMO_SCHEMA_MISSING: 'The evaluation database schema is missing or has an unexpected version. Prepare the reference data before starting.',
  DEMO_RECORD_MISSING: 'Jules’s linked customer, policy, claim or policy clauses are missing.',
  OPENAI_NOT_CONFIGURED: 'Set OPENAI_API_KEY on the server before starting Live voice.',
  OPENAI_LIVE_ACCESS_UNAVAILABLE: 'GPT-Live access is unavailable for this key or model. Check project access; Realtime was not substituted.',
  OPENAI_BACKEND_UNAVAILABLE: 'The delegated OpenAI model or tool configuration is unavailable. End the call and check server configuration.',
  OPENAI_RATE_LIMITED: 'OpenAI rejected the request because of a rate or usage limit.',
  OPENAI_SIDEBAND_FAILED: 'The private voice tool connection failed. The call has ended.',
  OPENAI_DISCONNECTED: 'The voice connection was lost. Final session usage is unconfirmed.',
  LIVE_FINALIZATION_UNCONFIRMED: 'The microphone has stopped. Final session usage could not be confirmed.',
  CUSTOMER_VERIFICATION_REQUIRED: 'Customer verification is required or has expired. Start a new call to verify again.',
  CLAIM_ACCESS_EXPIRED: 'Access to the saved report has expired. Start a new verified call to review it.',
  DRAFT_STALE: 'The report changed. Ask the assistant to prepare a new draft and review it again.',
  CLAIM_CLOSED: 'This claim is closed. A handler must review further changes.',
  MICROPHONE_DENIED: 'Microphone permission was denied or no microphone is available.',
  WEBRTC_UNAVAILABLE: 'Live voice requires a browser with WebRTC and microphone access over HTTPS or localhost.',
  AUDIO_PLAYBACK_BLOCKED: 'Select Play assistant audio to hear the assistant.',
  HOTLINE_FAILED: 'The Live request failed. No simulated reply has been substituted.'
}
export const hotlineErrorPhrase = code => errorPhrases[code] || errorPhrases.HOTLINE_FAILED

export function useLiveHotline() {
  const { locale } = useAppI18n()
  const { state: demoState } = useDemo()
  const { source, refreshLiveClaim } = useHotlineClaim()
  const { access, restoreAccess } = useHotlineAccess()
  const live = reactive({ phase: 'idle', transcript: '', response: '', error: '',
    verified: false, draft: null, traces: [], saved: false, audioBlocked: false, microphoneState: 'off', finalized: false, confirming: false, fragments: [], transcriptTruncated: false })
  const captionEntries = computed(() => {
    // Each speaker grows independently during overlap. Timeline gaps are a display
    // heuristic, not provider-confirmed turns. Preserve every retained delta verbatim.
    const entries = []
    for (const speaker of ['user', 'assistant']) {
      const fragments = live.fragments.map((fragment, index) => ({ ...fragment, index }))
        .filter(fragment => fragment.speaker === speaker)
        .sort((a, b) => (a.startMs ?? a.index) - (b.startMs ?? b.index) || a.index - b.index)
      let entry
      for (const fragment of fragments) {
        const gap = fragment.startMs != null && entry?.endMs != null ? fragment.startMs - entry.endMs : null
        if (!entry || gap > 1400) {
          entry = { id: `${speaker}-${fragment.index}`, speaker, text: '', startMs: fragment.startMs, endMs: fragment.endMs, index: fragment.index }
          entries.push(entry)
        }
        entry.text += fragment.delta
        if (fragment.endMs != null) entry.endMs = Math.max(entry.endMs ?? 0, fragment.endMs)
      }
    }
    return entries.sort((a, b) => (a.startMs ?? a.index) - (b.startMs ?? b.index) || a.index - b.index)
  })
  const active = computed(() => ['connecting', 'listening', 'responding', 'closing'].includes(live.phase))
  let peer, channel, microphone, audio, pollTimer, closeTimer, startupTimer, disconnectTimer, transcriptTimer
  let generation = 0
  let claimLoaded = false
  let aborter
  let greetingInstruction = '', greeted = false
  const seen = new Set()
  function release() {
    clearTimeout(pollTimer); clearTimeout(closeTimer); clearTimeout(startupTimer); clearTimeout(disconnectTimer); clearTimeout(transcriptTimer)
    microphone?.getTracks().forEach(track => track.stop())
    channel?.close(); peer?.close()
    if (audio) { audio.pause(); audio.srcObject = null }
    microphone = null; channel = null; peer = null; audio = null
    live.verified = false; live.draft = null; live.microphoneState = 'off'
  }
  function fail(error) {
    const code = error?.code || error?.message || 'HOTLINE_FAILED'
    live.error = hotlineErrorPhrase(code)
    if (code === 'HOTLINE_ACCESS_DENIED') access.value.authorized = false
    live.phase = 'error'
    ++generation; aborter?.abort()
    release()
    void hotlineRequest('close', {}, { keepalive: true }).catch(() => {})
  }
  async function poll(epoch) {
    if (epoch !== generation || !active.value) return
    try {
      const snapshot = await hotlineRequest('status')
      if (epoch !== generation || !active.value) return
      Object.assign(live, { verified: snapshot.verified, draft: snapshot.draft, traces: snapshot.traces || [] })
      if ((snapshot.saved && !live.saved) || snapshot.claimReadable && !claimLoaded) { live.saved = snapshot.saved; await refreshLiveClaim(); claimLoaded = true }
      if (snapshot.error) { fail({ code: snapshot.error }); return }
      if (snapshot.active === false) {
        live.finalized = snapshot.finalized === true
        if (!live.finalized) live.error = hotlineErrorPhrase('LIVE_FINALIZATION_UNCONFIRMED')
        live.phase = 'ended'; release(); return
      }
      pollTimer = setTimeout(() => { void poll(epoch) }, 1000)
    } catch (error) { if (epoch === generation) fail(error) }
  }
  function observe(event, epoch) {
    if (epoch !== generation || !active.value) return
    if (event.event_id) {
      if (seen.has(event.event_id)) return
      seen.add(event.event_id)
      if (seen.size > 2000) seen.delete(seen.values().next().value)
    }
    if (event.type === 'session.started') {
      clearTimeout(startupTimer)
      if (live.phase !== 'closing') live.phase = 'listening'
      if (!greeted && greetingInstruction && live.phase !== 'closing' && channel?.readyState === 'open') {
        greeted = true
        channel.send(JSON.stringify({ type: 'session.commentary.append', event_id: `greeting_${crypto.randomUUID()}`, delegation_id: null, content: greetingInstruction }))
      }
    }
    if (['session.input_transcript.delta', 'session.output_transcript.delta'].includes(event.type) && typeof event.delta === 'string') {
      // Exact fragments, including whitespace and repetition. UI groups them by speaker only.
      live.fragments.push({ speaker: event.type === 'session.input_transcript.delta' ? 'user' : 'assistant', delta: event.delta,
        startMs: Number.isFinite(event.start_ms) ? event.start_ms : null, endMs: Number.isFinite(event.end_ms) ? event.end_ms : null })
      if (live.fragments.length > 10000) { live.fragments.shift(); live.transcriptTruncated = true }
      const input = event.type === 'session.input_transcript.delta'
      const key = input ? 'transcript' : 'response'
      live[key] = live.fragments.filter(fragment => fragment.speaker === (input ? 'user' : 'assistant')).map(fragment => fragment.delta).join('')
      if (live.phase !== 'closing') live.phase = input ? 'listening' : 'responding'
      clearTimeout(transcriptTimer)
      transcriptTimer = setTimeout(() => { if (epoch === generation && live.phase === 'responding') live.phase = 'listening' }, 1800)
    }
    if (event.type === 'session.closed') {
      live.finalized = true; live.phase = 'ended'; release()
    }
    if (event.type === 'error') fail({ code: event.error?.code === 'responses_validation_failed' ? 'OPENAI_BACKEND_UNAVAILABLE' : 'OPENAI_PROTOCOL_ERROR' })
  }
  async function start() {
    if (active.value || source.value !== 'live') return
    const epoch = ++generation
    Object.assign(live, { phase: 'connecting', transcript: '', response: '', error: '', verified: false, draft: null, saved: false, finalized: false, traces: [], audioBlocked: false, fragments: [], transcriptTruncated: false })
    seen.clear(); claimLoaded = false
    greetingInstruction = ''; greeted = false
    aborter = new AbortController()
    try {
      if (!navigator.mediaDevices?.getUserMedia || !window.RTCPeerConnection) throw { code: 'WEBRTC_UNAVAILABLE' }
      await restoreAccess()
      if (epoch !== generation) return
      if (!access.value.authorized) throw { code: access.value.error || 'HOTLINE_ACCESS_DENIED' }
      // Called only by the explicit orb/start action. Never on mount or mode/language change.
      const stream = await navigator.mediaDevices.getUserMedia({ audio: { echoCancellation: true, noiseSuppression: true } })
      if (epoch !== generation) { stream.getTracks().forEach(track => track.stop()); return }
      microphone = stream; live.microphoneState = 'on'
      const connection = new RTCPeerConnection()
      peer = connection
      audio = new Audio(); audio.autoplay = true
      connection.addEventListener('track', event => {
        if (epoch !== generation || !audio) return
        audio.srcObject = event.streams[0] || new MediaStream([event.track])
        audio.play().catch(() => { if (epoch === generation) live.audioBlocked = true })
      })
      for (const track of microphone.getAudioTracks()) {
        connection.addTrack(track, microphone)
        track.addEventListener('ended', () => { if (epoch === generation && active.value) fail({ code: 'MICROPHONE_DENIED' }) })
      }
      channel = connection.createDataChannel('oai-events')
      channel.addEventListener('message', ({ data }) => {
        try { observe(JSON.parse(data), epoch) } catch { if (epoch === generation) fail({ code: 'OPENAI_PROTOCOL_ERROR' }) }
      })
      channel.addEventListener('close', () => { if (epoch === generation && active.value && !live.finalized && live.phase !== 'closing') fail({ code: 'OPENAI_DISCONNECTED' }) })
      connection.addEventListener('connectionstatechange', () => {
        if (epoch !== generation || !active.value) return
        if (connection.connectionState === 'failed') fail({ code: 'OPENAI_DISCONNECTED' })
        if (connection.connectionState === 'disconnected') disconnectTimer = setTimeout(() => { if (epoch === generation && connection.connectionState === 'disconnected') fail({ code: 'OPENAI_DISCONNECTED' }) }, 5000)
        else clearTimeout(disconnectTimer)
      })
      await connection.setLocalDescription(await connection.createOffer())
      if (connection.iceGatheringState !== 'complete') await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => { connection.removeEventListener('icegatheringstatechange', changed); reject({ code: 'OPENAI_CONNECTION_FAILED' }) }, 10000)
        function changed() { if (connection.iceGatheringState === 'complete') { clearTimeout(timeout); connection.removeEventListener('icegatheringstatechange', changed); resolve() } }
        connection.addEventListener('icegatheringstatechange', changed); changed()
      })
      if (epoch !== generation) return
      const result = await hotlineRequest('session', { sdp: connection.localDescription.sdp, locale: locale.value })
      greetingInstruction = result.greeting || ''
      if (epoch !== generation) { void hotlineRequest('close', { sessionId: result.session.id }, { keepalive: true }).catch(() => {}); return }
      // Creation starts Live. Never send session.start or Realtime response commands.
      startupTimer = setTimeout(() => { if (epoch === generation && live.phase === 'connecting') fail({ code: 'OPENAI_CONNECTION_FAILED' }) }, 20000)
      await connection.setRemoteDescription({ type: 'answer', sdp: result.transport.sdp })
      void poll(epoch)
    } catch (error) {
      if (epoch === generation) fail(['NotAllowedError', 'NotFoundError', 'NotReadableError'].includes(error?.name) ? { code: 'MICROPHONE_DENIED' } : error)
    }
  }
  async function end({ immediate = false } = {}) {
    if (!active.value) return
    if (live.phase === 'connecting') immediate = true
    live.phase = 'closing'; live.microphoneState = microphone ? 'muted' : 'off'
    microphone?.getTracks().forEach(track => { track.enabled = false })
    if (immediate) {
      ++generation; aborter?.abort(); release(); live.phase = 'ended'
      void hotlineRequest('close', {}, { keepalive: true }).catch(() => {})
      return
    }
    closeTimer = setTimeout(() => { if (live.phase === 'closing') fail({ code: 'LIVE_FINALIZATION_UNCONFIRMED' }) }, 20000)
    try { await hotlineRequest('close') } catch (error) { fail(error) }
  }
  async function interruptSpeech() {
    if (!active.value || live.phase === 'connecting' || live.phase === 'closing') return
    // Drop current playback while a trusted server instruction tells Live to yield.
    const epoch = generation
    const output = audio
    if (output) output.muted = true
    try {
      await hotlineRequest('interrupt')
      if (epoch !== generation || !active.value || live.phase === 'closing') return
      live.phase = 'listening'
      setTimeout(() => { if (epoch === generation && output) output.muted = false }, 1200)
    } catch (error) { if (epoch === generation && active.value) fail(error) }
  }
  async function confirm() {
    if (!live.draft || live.confirming) return
    live.confirming = true; live.error = ''
    try { await hotlineRequest('confirm', { draftId: live.draft.id }); live.saved = true; live.draft = null; await refreshLiveClaim() }
    catch (error) { live.error = hotlineErrorPhrase(error.code) }
    finally { live.confirming = false }
  }
  async function playAudio() {
    try { await audio?.play(); live.audioBlocked = false } catch { live.audioBlocked = true }
  }
  watch(() => demoState.value.revision, () => { void end({ immediate: true }) })
  watch(locale, value => { if (active.value && live.phase !== 'connecting' && live.phase !== 'closing') void hotlineRequest('language', { locale: value }).catch(error => { if (active.value) fail(error) }) })
  const beforeUnload = () => { void end({ immediate: true }) }
  onMounted(() => {
    window.addEventListener('pagehide', beforeUnload)
    void restoreAccess()
  })
  onBeforeUnmount(() => { window.removeEventListener('pagehide', beforeUnload); void end({ immediate: true }) })
  return { live, active, access, captionEntries, start, end, interruptSpeech, confirm, playAudio }
}
