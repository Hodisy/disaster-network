import { claimStages, createReferenceClaim } from './useClaimScenario.js'

import { DEMO_CUSTOMER_ID, DEMO_CONTRACT_NUMBER } from '../config/demo.js'

const steps = claimStages
const documents = [
  { id: 'photos', title: 'Damage photos', description: 'Overview and close-up photos of affected areas.' },
  { id: 'address', title: 'Proof of address', description: 'A recent document in your name.' },
  { id: 'inventory', title: 'Belongings inventory', description: 'A list of damaged belongings, even a provisional one.' }
]
const stepDescriptions = [
  'Add your documents to complete your claim file.',
  'All three requested documents have been collected.',
  'A coverage proposal is ready for review.',
  'Review access and provider availability before arranging visits.',
  'The claim file is closed.'
]
const script = [
  { transcript: `Hello, I am Jules. My contract number is ${DEMO_CONTRACT_NUMBER}. Water entered the ground floor of my home in Istres, and I need help with accommodation.`, response: 'I can help with that, Jules. Are you and everyone in your household safe, and can you still stay in your home?' },
  { transcript: 'My partner, our child and I are safe at my sister’s home. There are three of us. We cannot return home, and we cannot stay here for much longer. I do not know whether access is allowed.', response: 'Thank you, Jules. I have noted temporary accommodation, electrical safety, cleaning and drying, storage, and a damage assessment. Physical visits depend on confirmed access. Nothing has been booked at this stage.' },
  { transcript: 'I already have photos taken before we left. I can prepare a provisional list of damaged belongings. What should I send you?', response: 'Your report is ready. In My claim, add the photos you already have, proof of address and a provisional inventory. Urgent accommodation assistance does not need to wait for a complete inventory. Do not return to the property while access is uncertain.' }
]

const initialState = () => ({
  accountMenu: false, resetMenu: false, callRevision: 0,
  call: { phase: 'idle', turn: 0, transcript: '', response: '', complete: false },
  claim: null, files: [], messages: [], busy: null, error: '', notice: '', revision: 0
})

// Mock boundary: replace these Promise-based actions with a future frontend API adapter.
// Presentation components never import this adapter or the scenario data.
function createMockGateway() {
  const wait = (data, delay = 500) => new Promise(resolve => setTimeout(() => resolve({ data }), delay))
  return {
    transcribe: turn => wait(script[turn].transcript, 1400),
    respond: turn => wait(script[turn].response, 1100),
    openClaim: () => wait(createReferenceClaim()),
    sendMessage: (text, stage) => wait(stage === 0
      ? 'Your message has been added to the claim, Jules. Add the supporting documents already available to you; an initial inventory can remain provisional.'
      : 'Your message has been added to the claim, Jules. The requested categories have been collected and still require review. The timeline shows the next step.', 1100),
    advance: stage => wait(stage + 1, 800)
  }
}

export function useDemo() {
  const state = useState('cat-response-demo', initialState)
  // Migrate only canonical demo identity fields; user messages and local files stay intact.
  if (state.value.claim?.id === 'CAT-2026-001' && state.value.claim.source !== 'postgresql' && !state.value.claim.moneyLines) {
    state.value.claim = { ...createReferenceClaim(), stage: state.value.claim.stage }
  }
  const gateway = createMockGateway()
  const version = () => state.value.revision
  const documentRows = computed(() => {
    const live = state.value.claim?.source === 'postgresql'
    const rows = live ? [...documents, { id: 'invoice', title: 'Item invoices', description: 'Later invoices can support individual item values.' }].filter(doc => state.value.claim.requiredDocuments.includes(doc.id)) : documents
    return rows.map(doc => ({ ...doc, files: state.value.files.filter(file => file.kind === doc.id) }))
  })
  const received = computed(() => documentRows.value.filter(doc => doc.files.length > 0).length)
  const stageLabel = computed(() => steps[state.value.claim?.stage ?? 0])
  // Keep canonical English phrases and ISO dates so existing content can change language.
  const addMessage = (role, text) => state.value.messages.push({ id: crypto.randomUUID(), role, text, createdAt: new Date().toISOString() })

  async function speak() {
    const call = state.value.call
    if (['listening', 'responding'].includes(call.phase)) return stopCall()
    if (call.complete) return
    const revision = ++state.value.callRevision
    const epoch = version()
    const isCurrent = () => state.value.callRevision === revision && version() === epoch
    call.phase = 'listening'
    call.transcript = ''
    call.response = ''
    state.value.error = ''
    try {
      const transcript = await gateway.transcribe(call.turn)
      if (!isCurrent()) return
      call.transcript = transcript.data
      call.phase = 'responding'
      const response = await gateway.respond(call.turn)
      if (!isCurrent()) return
      call.response = response.data
      call.turn += 1
      call.complete = call.turn === script.length
      call.phase = 'ready'
    } catch {
      if (isCurrent()) { call.phase = 'idle'; state.value.error = 'The conversation was interrupted. You can try again.' }
    }
  }

  function stopCall() {
    state.value.callRevision += 1
    if (['listening', 'responding'].includes(state.value.call.phase)) state.value.call.phase = 'paused'
  }

  async function openClaim() {
    if (state.value.claim || state.value.busy || !state.value.call.complete) return
    state.value.busy = 'open'
    state.value.error = ''
    const revision = version()
    try {
      const result = await gateway.openClaim()
      if (revision !== version()) return
      if (state.value.claim?.source === 'postgresql') return
      state.value.claim = result.data
      addMessage('adviser', 'Hello Jules, your report has been received. Add the three requested documents to complete your claim file. You can contact me here whenever you need help.')
    } catch { if (revision === version()) state.value.error = 'The claim could not be opened. Please try again.' }
    finally { if (revision === version()) state.value.busy = null }
  }

  async function addFiles(kind, inputFiles) {
    state.value.error = ''
    if (!state.value.claim || state.value.busy || state.value.claim.stage > 1) return
    if (!documentRows.value.some(doc => doc.id === kind)) return
    const revision = version()
    state.value.busy = 'files'
    const selected = Array.from(inputFiles || [])
    try {
      if (state.value.files.length + selected.length > 12) throw new Error('You can keep up to 12 files in this tab.')
      const total = [...state.value.files, ...selected].reduce((sum, file) => sum + file.size, 0)
      if (total > 30 * 1024 * 1024) throw new Error('You can keep up to 30 MB of files in this browser.')
      const accepted = []
      for (const file of selected) {
        if (file.size > 10 * 1024 * 1024) throw new Error('Each file must be smaller than 10 MB.')
        const bytes = new Uint8Array(await file.slice(0, 12).arrayBuffer())
        const isPdf = bytes[0] === 37 && bytes[1] === 80 && bytes[2] === 68 && bytes[3] === 70 && bytes[4] === 45
        const isPng = [137, 80, 78, 71, 13, 10, 26, 10].every((byte, index) => bytes[index] === byte)
        const isJpg = bytes[0] === 255 && bytes[1] === 216 && bytes[2] === 255
        const isWebp = String.fromCharCode(...bytes.slice(0, 4)) === 'RIFF' && String.fromCharCode(...bytes.slice(8, 12)) === 'WEBP'
        if (!isPdf && !isPng && !isJpg && !isWebp) throw new Error('Unrecognised format. Choose a JPG, PNG, WebP image or a PDF.')
        if (revision !== version()) return
        accepted.push({ file, type: isPdf ? 'application/pdf' : 'image' })
      }
      if (revision !== version()) return
      // Validate the entire selection before adding any browser-local previews.
      for (const { file, type } of accepted) state.value.files.push({ id: crypto.randomUUID(), kind, name: file.name, size: file.size, type, url: URL.createObjectURL(file) })
      updateCompleteness()
    } catch (error) { if (revision === version()) { updateCompleteness(); state.value.error = error.message || 'The file could not be read.' } }
    finally { if (revision === version()) state.value.busy = null }
  }

  function updateCompleteness() {
    if (state.value.claim?.source === 'postgresql') return
    if (state.value.claim && state.value.claim.stage <= 1) state.value.claim.stage = received.value === documents.length ? 1 : 0
  }

  function removeFile(id) {
    if (!state.value.claim || state.value.claim.stage > 1 || state.value.busy) return
    const file = state.value.files.find(file => file.id === id)
    if (file) URL.revokeObjectURL(file.url)
    state.value.files = state.value.files.filter(file => file.id !== id)
    updateCompleteness()
  }

  async function sendMessage(text) {
    const value = text.trim().slice(0, 2000)
    if (!value || !state.value.claim || state.value.claim.source === 'postgresql' || state.value.busy) return false
    state.value.busy = 'message'
    state.value.error = ''
    addMessage('client', value)
    const revision = version()
    try {
      const result = await gateway.sendMessage(value, state.value.claim.stage)
      if (revision !== version() || state.value.claim?.source === 'postgresql') return false
      addMessage('adviser', result.data)
      return true
    } catch { if (revision === version()) state.value.error = 'The reply could not be displayed.'; return false }
    finally { if (revision === version()) state.value.busy = null }
  }

  async function advance() {
    const claim = state.value.claim
    if (!claim || claim.source === 'postgresql' || claim.stage < 1 || claim.stage > 3 || state.value.busy) return
    state.value.busy = 'advance'
    state.value.error = ''
    const revision = version()
    try {
      const result = await gateway.advance(claim.stage)
      if (revision !== version() || state.value.claim?.source === 'postgresql') return
      claim.stage = result.data
      if (claim.stage === 2) claim.coverage = 'Proposal'
      const messages = {
        2: 'A coverage proposal is ready for review. The reported damage, applicable limits and supporting evidence still need to be assessed.',
        3: 'The assistance plan is prepared. Physical visits remain subject to confirmed access and provider availability. Accommodation remains a proposal until a booking is confirmed.',
        4: 'The claim file is closed. No payment has been recorded.'
      }
      addMessage('adviser', messages[claim.stage])
    } catch { if (revision === version()) state.value.error = 'This step could not be completed. Please try again.' }
    finally { if (revision === version()) state.value.busy = null }
  }

  function reset() {
    const nextRevision = version() + 1
    state.value.files.forEach(file => URL.revokeObjectURL(file.url))
    state.value = { ...initialState(), revision: nextRevision, notice: 'Local workspace cleared.' }
  }

  return { state, customerId: DEMO_CUSTOMER_ID, steps, stepDescriptions, scenarioDate: '2026-10-20', eventDate: '2026-10-20', documentRows, received, stageLabel, speak, stopCall, openClaim, addFiles, removeFile, sendMessage, advance, reset }
}
