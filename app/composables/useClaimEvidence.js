const errorPhrases = {
  CLAIM_ACCESS_EXPIRED: 'Verify your policy on the hotline again to access the documents.',
  HOTLINE_ACCESS_DENIED: 'Unlock the Live workspace in Settings, then verify your policy on the hotline.',
  DOCUMENT_INVALID_SIZE: 'Choose a non-empty file up to 10 MB.',
  DOCUMENT_FORMAT_UNSUPPORTED: 'Choose a JPG, PNG, WebP photo or a PDF.',
  PHOTO_IMAGE_REQUIRED: 'For damage photos, choose a JPG, PNG or WebP image.',
  DOCUMENT_PDF_PAGE_LIMIT: 'Choose a PDF with no more than 12 pages.',
  DOCUMENT_PDF_UNREADABLE: 'This PDF cannot be read. Choose an unlocked, valid PDF.',
  DOCUMENT_LIMIT_REACHED: 'This dossier can hold up to 12 files and 30 MB of supporting documents.',
  DOCUMENT_ANALYSIS_BUSY: 'This document is already being read. Refresh the documents shortly.',
  EVIDENCE_CONSENT_REQUIRED: 'Allow AI reading before requesting an analysis.',
  EVIDENCE_ANALYSIS_UNAVAILABLE: 'The file is saved, but AI reading is unavailable. You can retry.',
  OPENAI_NOT_CONFIGURED: 'The file is saved. Configure the server OpenAI key to read it.',
  OPENAI_RATE_LIMITED: 'The AI service is busy. The file remains saved; try again later.',
  POLICY_PDF_UNAVAILABLE: 'The policy PDF could not be created from the stored clauses.',
  DATABASE_UNAVAILABLE: 'The database is unavailable. The operation could not be confirmed.',
  CLAIM_CLOSED: 'This dossier is closed. New documents require a handler review.'
}

export function useClaimEvidence() {
  const { state } = useDemo()
  const { locale } = useAppI18n()
  const evidence = useState('claim-evidence', () => ({ documents: [], officialSource: null }))
  const busy = ref(''), error = ref(''), consent = ref(false)
  const urls = new Set()
  const live = computed(() => state.value.claim?.source === 'postgresql')
  const files = computed(() => evidence.value.documents.map(doc => ({ ...doc, persisted: true, name: doc.filename,
    size: doc.size_bytes, type: doc.mime_type === 'application/pdf' ? 'application/pdf' : 'image' })))
  const describe = problem => errorPhrases[problem?.code] || 'The document operation could not be confirmed. Refresh and try again.'
  async function request(action, body = {}, { form, binary = false } = {}) {
    const response = await fetch(`/api/hotline/evidence/${action}`, { method: 'POST', credentials: 'same-origin',
      headers: form ? { 'X-Hotline-Request': '1' } : { 'Content-Type': 'application/json', 'X-Hotline-Request': '1' },
      body: form || JSON.stringify({ ...body, locale: locale.value }) })
    if (binary && response.ok && !response.headers.get('content-type')?.includes('application/json')) return response.blob()
    const result = await response.json()
    if (!response.ok || !result.ok) throw Object.assign(new Error(), { code: result.code })
    return result
  }
  async function refresh({ quiet = false } = {}) {
    if (!live.value) return
    const revision = state.value.revision
    try {
      const result = await request('list')
      if (state.value.revision === revision && live.value) evidence.value = { documents: result.documents, officialSource: result.officialSource }
    } catch (problem) { if (!quiet) error.value = describe(problem) }
  }
  async function upload(kind, selected) {
    if (!live.value || busy.value) return
    const revision = state.value.revision
    error.value = ''; busy.value = 'upload'
    try {
      for (const file of Array.from(selected || [])) {
        if (state.value.revision !== revision || !live.value) break
        if (!file.size || file.size > 10 * 1024 * 1024) throw Object.assign(new Error(), { code: 'DOCUMENT_INVALID_SIZE' })
        const form = new FormData()
        form.append('file', file); form.append('kind', kind); form.append('locale', locale.value)
        const result = await request('upload', {}, { form })
        if (state.value.revision !== revision || !live.value) break
        await refresh()
        if (consent.value) {
          busy.value = 'analyze'
          await request('analyze', { documentId: result.document.id, consent: true })
          busy.value = 'upload'
          await refresh()
        }
      }
    } catch (problem) { if (state.value.revision === revision) error.value = describe(problem) }
    finally { if (state.value.revision === revision) { busy.value = ''; await refresh({ quiet: true }) } }
  }
  async function analyze(documentId) {
    if (busy.value || !live.value) return
    if (!consent.value) { error.value = errorPhrases.EVIDENCE_CONSENT_REQUIRED; return }
    busy.value = 'analyze'; error.value = ''
    try { await request('analyze', { documentId, consent: true }) }
    catch (problem) { error.value = describe(problem) }
    finally { busy.value = ''; await refresh({ quiet: true }) }
  }
  async function checkSource() {
    if (busy.value || !live.value) return
    busy.value = 'source'; error.value = ''
    try { const result = await request('source'); evidence.value.officialSource = result.officialSource }
    catch (problem) { error.value = describe(problem) }
    finally { busy.value = '' }
  }
  async function previewSaved(file) {
    if (busy.value) return null
    busy.value = 'download'; error.value = ''
    try {
      const blob = await request('download', { documentId: file.id }, { binary: true })
      const url = URL.createObjectURL(blob); urls.add(url)
      return { ...file, name: file.name || file.filename, type: file.mime_type === 'application/pdf' ? 'application/pdf' : 'image', persisted: true, url }
    } catch (problem) { error.value = describe(problem); return null }
    finally { busy.value = '' }
  }
  async function downloadContract() {
    if (busy.value || !live.value) return
    busy.value = 'contract'; error.value = ''
    try {
      const blob = await request('contract', {}, { binary: true })
      const url = URL.createObjectURL(blob); urls.add(url)
      const anchor = document.createElement('a')
      anchor.href = url; anchor.download = `allstorm-home-protect-${state.value.claim.contractNumber}.pdf`; anchor.click()
    } catch (problem) { error.value = describe(problem) }
    finally { busy.value = '' }
  }
  function clear() {
    for (const url of urls) URL.revokeObjectURL(url)
    urls.clear(); evidence.value = { documents: [], officialSource: null }; error.value = ''; busy.value = ''; consent.value = false
  }
  watch(() => state.value.revision, clear)
  onUnmounted(() => { for (const url of urls) URL.revokeObjectURL(url); urls.clear() })
  return { evidence, files, busy, error, consent, refresh, upload, analyze, checkSource, previewSaved, downloadContract }
}
