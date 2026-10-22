const steps = ['À constituer', 'Dossier complet', 'Proposition de garantie', 'Assistance organisée', 'Sinistre clôturé']
const documents = [
  { id: 'photos', title: 'Photos des dommages', description: 'Vues d’ensemble et détails des zones touchées.' },
  { id: 'address', title: 'Justificatif de domicile', description: 'Un document récent à votre nom.' },
  { id: 'inventory', title: 'Inventaire des biens', description: 'La liste des biens endommagés, même provisoire.' }
]
const stepDescriptions = [
  'Ajoutez vos documents pour compléter votre dossier.',
  'Vos trois pièces ont été réunies.',
  'Une proposition fictive est disponible.',
  'Les interventions sont simulées.',
  'Le parcours de démonstration est terminé.'
]
const script = [
  { transcript: 'Bonjour, je suis Claire Martin. L’eau est entrée dans ma maison à Istres. Le rez-de-chaussée est inondé.', response: 'Je suis là pour vous accompagner, Claire. Vous êtes bien identifiée pour cette démonstration. Êtes-vous en sécurité et votre logement est-il toujours habitable ?' },
  { transcript: 'Oui, je suis en sécurité chez ma sœur. Mais je ne peux pas rentrer chez moi. Il y a encore de l’eau et les meubles sont abîmés.', response: 'Merci Claire. Je note un besoin d’hébergement temporaire et d’assèchement. Une expertise permettra aussi d’évaluer les dommages. Rien n’est encore réservé à ce stade.' },
  { transcript: 'D’accord. Je peux prendre des photos et préparer la liste des meubles endommagés. Que dois-je vous transmettre ?', response: 'Votre déclaration est prête. Ouvrez votre dossier, puis ajoutez des photos des dommages, un justificatif de domicile et un inventaire des biens. Vous pourrez suivre la suite depuis « Mon dossier ».' }
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
    openClaim: () => wait({
      id: 'CAT-2026-001', name: 'Claire Martin', address: '12 rue des Oliviers, 13800 Istres',
      event: 'Inondation', date: '22 octobre 2026', stage: 0,
      summary: 'Le rez-de-chaussée est inondé et le logement est temporairement inhabitable. Claire est en sécurité chez sa sœur. Des meubles ont été endommagés.',
      needs: ['Hébergement temporaire', 'Assèchement', 'Expertise']
    }),
    sendMessage: (text, stage) => wait(stage === 0
      ? 'Merci Claire, votre message est bien reçu dans cette démonstration. Vous pouvez ajouter les pièces manquantes indiquées dans votre dossier, à votre rythme.'
      : 'Merci Claire, votre message est bien reçu dans cette démonstration. Vos pièces sont réunies. Le parcours à droite indique l’étape actuelle de votre dossier.', 1100),
    advance: stage => wait(stage + 1, 800)
  }
}

export function useDemo() {
  const state = useState('cat-response-demo', initialState)
  const gateway = createMockGateway()
  const version = () => state.value.revision
  const documentRows = computed(() => documents.map(doc => ({ ...doc, files: state.value.files.filter(file => file.kind === doc.id) })))
  const received = computed(() => documentRows.value.filter(doc => doc.files.length > 0).length)
  const stageLabel = computed(() => steps[state.value.claim?.stage ?? 0])
  const addMessage = (role, text) => state.value.messages.push({ id: crypto.randomUUID(), role, text, time: new Intl.DateTimeFormat('fr-FR', { hour: '2-digit', minute: '2-digit', timeZone: 'Europe/Paris' }).format(new Date()) })

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
      if (isCurrent()) { call.phase = 'idle'; state.value.error = 'La simulation a été interrompue. Vous pouvez réessayer.' }
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
      state.value.claim = result.data
      addMessage('adviser', 'Bonjour Claire, votre déclaration est reçue. Ajoutez les trois pièces demandées pour compléter votre dossier. Je reste à votre disposition ici.')
    } catch { if (revision === version()) state.value.error = 'Le dossier n’a pas pu être ouvert. Réessayez.' }
    finally { if (revision === version()) state.value.busy = null }
  }

  async function addFiles(kind, inputFiles) {
    state.value.error = ''
    if (!state.value.claim || state.value.busy || state.value.claim.stage > 1) return
    if (!documents.some(doc => doc.id === kind)) return
    const revision = version()
    state.value.busy = 'files'
    const selected = Array.from(inputFiles || [])
    try {
      if (state.value.files.length + selected.length > 12) throw new Error('Vous pouvez conserver jusqu’à 12 fichiers dans cet onglet.')
      const total = [...state.value.files, ...selected].reduce((sum, file) => sum + file.size, 0)
      if (total > 30 * 1024 * 1024) throw new Error('La limite de cette démonstration est de 30 Mo au total.')
      const accepted = []
      for (const file of selected) {
        if (file.size > 10 * 1024 * 1024) throw new Error('Chaque fichier doit peser moins de 10 Mo.')
        const bytes = new Uint8Array(await file.slice(0, 12).arrayBuffer())
        const isPdf = bytes[0] === 37 && bytes[1] === 80 && bytes[2] === 68 && bytes[3] === 70 && bytes[4] === 45
        const isPng = [137, 80, 78, 71, 13, 10, 26, 10].every((byte, index) => bytes[index] === byte)
        const isJpg = bytes[0] === 255 && bytes[1] === 216 && bytes[2] === 255
        const isWebp = String.fromCharCode(...bytes.slice(0, 4)) === 'RIFF' && String.fromCharCode(...bytes.slice(8, 12)) === 'WEBP'
        if (!isPdf && !isPng && !isJpg && !isWebp) throw new Error('Format non reconnu. Choisissez une photo JPG, PNG, WebP ou un PDF.')
        if (revision !== version()) return
        accepted.push({ file, type: isPdf ? 'application/pdf' : 'image' })
      }
      if (revision !== version()) return
      // Validate the entire selection before adding any browser-local previews.
      for (const { file, type } of accepted) state.value.files.push({ id: crypto.randomUUID(), kind, name: file.name, size: file.size, type, url: URL.createObjectURL(file) })
      updateCompleteness()
    } catch (error) { if (revision === version()) { updateCompleteness(); state.value.error = error.message || 'Le fichier n’a pas pu être lu.' } }
    finally { if (revision === version()) state.value.busy = null }
  }

  function updateCompleteness() {
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
    if (!value || !state.value.claim || state.value.busy) return false
    state.value.busy = 'message'
    state.value.error = ''
    addMessage('client', value)
    const revision = version()
    try {
      const result = await gateway.sendMessage(value, state.value.claim.stage)
      if (revision !== version()) return false
      addMessage('adviser', result.data)
      return true
    } catch { if (revision === version()) state.value.error = 'La réponse simulée n’a pas pu être affichée.'; return false }
    finally { if (revision === version()) state.value.busy = null }
  }

  async function advance() {
    const claim = state.value.claim
    if (!claim || claim.stage < 1 || claim.stage > 3 || state.value.busy) return
    state.value.busy = 'advance'
    state.value.error = ''
    const revision = version()
    try {
      const result = await gateway.advance(claim.stage)
      if (revision !== version()) return
      claim.stage = result.data
      const messages = {
        2: 'Une proposition de garantie de démonstration est disponible pour les dommages déclarés. Elle est fictive, sans valeur contractuelle, et ne constitue pas une acceptation réelle de prise en charge.',
        3: 'Assistance organisée dans le scénario : hébergement et assèchement prévus le 23 octobre 2026. Aucune réservation ni intervention réelle n’a été effectuée.',
        4: 'La clôture du sinistre a été simulée. Aucun paiement ni règlement réel n’a eu lieu.'
      }
      addMessage('adviser', messages[claim.stage])
    } catch { if (revision === version()) state.value.error = 'Cette étape n’a pas pu être simulée. Réessayez.' }
    finally { if (revision === version()) state.value.busy = null }
  }

  function reset() {
    const nextRevision = version() + 1
    state.value.files.forEach(file => URL.revokeObjectURL(file.url))
    state.value = { ...initialState(), revision: nextRevision, notice: 'La démonstration a été réinitialisée.' }
  }

  return { state, steps, stepDescriptions, documentRows, received, stageLabel, speak, stopCall, openClaim, addFiles, removeFile, sendMessage, advance, reset }
}
