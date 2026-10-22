<script setup>
const { t, formatDate, formatMoney } = useAppI18n()
defineProps({ documents: Array, officialSource: Object, busy: String, error: String, consent: Boolean, closed: Boolean })
const emit = defineEmits(['upload', 'analyze', 'source', 'contract', 'preview', 'update:consent', 'refresh'])
const kind = ref('photos')
const categories = [
  ['photos', 'Damage photos'], ['invoice', 'Accommodation invoice'], ['address', 'Proof of address'],
  ['inventory', 'Belongings inventory'], ['other', 'Other supporting document']
]
const sourceLabels = {
  matching_record_found: 'A matching flood record was found. The official order still requires review.',
  no_matching_record: 'No matching flood record was found for this municipality and event date. This is not a rejection.',
  incomplete: 'The source returned an incomplete result. Recognition remains unverified.',
  unavailable: 'The official source is unavailable. Recognition remains unverified.'
}
function select(event) {
  emit('upload', kind.value, Array.from(event.target.files || []))
  event.target.value = ''
}
</script>

<template>
  <section class="evidence-chat" :aria-label="t('Document conversation')">
    <div class="heading"><h2>{{ t('Documents and checks') }}</h2><button :disabled="!!busy" @click="$emit('refresh')">{{ t('Refresh') }}</button></div>
    <div class="assistant-note"><span class="assistant-mark">//</span><p>{{ t('Add the photos you already have or your accommodation invoice. I can read them and compare the details with your policy. Only take photos when it is safe.') }}</p></div>
    <div class="contract-link"><UiIcon name="FileText" :size="15" /><span>{{ t('Your policy schedule and conditions') }}</span><button :disabled="!!busy" @click="$emit('contract')">{{ t(busy === 'contract' ? 'Preparing PDF…' : 'Download policy PDF') }}</button></div>
    <div v-if="documents.length" class="document-messages" role="log" :aria-label="t('Saved supporting documents')">
      <article v-for="document in documents" :key="document.id" class="document-message">
        <div class="file-heading"><UiIcon :name="document.mime_type === 'application/pdf' ? 'FileText' : 'Image'" :size="17" /><button class="filename" :disabled="!!busy" @click="$emit('preview', document)">{{ document.filename }}</button><span>{{ t('Saved') }}</span></div>
        <div v-if="document.analysis" class="reading">
          <span class="reading-label">{{ t('AI reading · Requires review') }} · {{ formatDate(document.analysis.checkedAt, { hour: '2-digit', minute: '2-digit' }) }}</span><p>{{ document.analysis.facts.summary }}</p>
          <ul v-if="document.analysis.facts.visible_observations.length"><li v-for="observation in document.analysis.facts.visible_observations" :key="observation">{{ observation }}</li></ul>
          <p v-for="uncertainty in document.analysis.facts.uncertainties" :key="uncertainty" class="uncertainty">{{ uncertainty }}</p>
          <div v-if="document.analysis.facts.invoice_total_cents !== null" class="extracted-amount"><span>{{ t('Extracted invoice total') }}</span><strong>{{ document.analysis.facts.currency === 'EUR' ? formatMoney(document.analysis.facts.invoice_total_cents / 100, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : `${document.analysis.facts.invoice_total_cents / 100} ${document.analysis.facts.currency || '—'}` }}</strong></div>
          <p v-if="document.analysis.stale" class="uncertainty">{{ t('The dossier or policy has changed since this reading. Run the comparison again.') }}</p>
          <div v-for="check in document.analysis.comparison.checks" :key="check.label" class="check-row"><span>{{ t(check.label) }}</span><span :class="document.analysis.stale ? 'needs_review' : check.status">{{ t(!document.analysis.stale && check.status === 'consistent' ? 'Consistent' : 'To review') }}</span></div>
          <button v-if="document.analysis.comparison.citation && !document.analysis.stale" class="citation" :disabled="!!busy" @click="$emit('contract')">{{ t('Policy clause {clause} · Page {page}', { clause: document.analysis.comparison.citation.clause, page: document.analysis.comparison.citation.page }) }}<UiIcon name="ArrowUpRight" :size="12" /></button>
          <button v-if="document.analysis.stale" class="citation" :disabled="!!busy || !consent || closed" @click="$emit('analyze', document.id)">{{ t('Read document') }}</button>
        </div>
        <div v-else class="not-read"><span>{{ t(document.extraction_status === 'failed' ? 'Saved · Reading failed' : 'Saved · Not yet read') }}</span><button :disabled="!!busy || !consent || closed" @click="$emit('analyze', document.id)">{{ t('Read document') }}</button></div>
      </article>
    </div>
    <div v-if="!closed" class="attachment-composer">
      <label class="category-label">{{ t('Supporting document type') }}<select v-model="kind" :disabled="!!busy"><option v-for="[value, label] in categories" :key="value" :value="value">{{ t(label) }}</option></select></label>
      <label class="consent"><input type="checkbox" :checked="consent" :disabled="!!busy" @change="$emit('update:consent', $event.target.checked)"><span>{{ t('Allow AI reading: selected files will be sent to OpenAI for analysis. You can also save them without AI reading.') }}</span></label>
      <div class="attach-row"><span>{{ t('JPG, PNG, WebP or PDF · 10 MB per file · Up to 12 PDF pages') }}</span><label class="attach-button" :class="{ disabled: busy }"><UiIcon name="Paperclip" :size="14" />{{ t(busy === 'upload' ? 'Uploading…' : busy === 'analyze' ? 'Reading…' : consent ? 'Attach and read' : 'Attach file') }}<input type="file" :accept="kind === 'photos' ? 'image/jpeg,image/png,image/webp' : 'image/jpeg,image/png,image/webp,application/pdf'" multiple :disabled="!!busy" :aria-label="t('Attach supporting documents')" @change="select"></label></div>
      <p class="storage-note">{{ t('Files are saved in your dossier. AI reading does not confirm authenticity, coverage or reimbursement.') }}</p>
    </div>
    <div class="official-source"><div class="source-heading"><div><h3>{{ t('Official event check') }}</h3><span>Géorisques · GASPAR</span></div><button :disabled="!!busy || closed" @click="$emit('source')">{{ t(busy === 'source' ? 'Checking…' : 'Check official source') }}</button></div>
      <template v-if="officialSource"><p>{{ t(sourceLabels[officialSource.status] || sourceLabels.incomplete) }}</p><p class="source-meta">{{ t('Municipality {code} · Event {date}', { code: officialSource.inseeCode, date: formatDate(officialSource.incidentDate) }) }}<br>{{ t('Consulted on {date}', { date: formatDate(officialSource.checkedAt, { hour: '2-digit', minute: '2-digit' }) }) }}</p><ul v-if="officialSource.records.length"><li v-for="record in officialSource.records" :key="record.reference">{{ record.reference }} · {{ record.risk }} · {{ t('Published on {date}', { date: formatDate(record.published) }) }}</li></ul><a :href="officialSource.sourceUrl" target="_blank" rel="noopener noreferrer">{{ t('View official source') }}<UiIcon name="ArrowUpRight" :size="12" /></a><small>{{ t('Publication can precede the database update. This check does not establish insurance coverage.') }}</small></template>
      <p v-else>{{ t('Consult the public flood-recognition records for the municipality and event date. Recognition has not been checked.') }}</p>
    </div>
    <p v-if="error" class="evidence-error" role="alert">{{ t(error) }}</p>
  </section>
</template>

<style scoped>
.evidence-chat { margin-top: 12px; padding-top: 20px; border-top: 1px solid #e7e7eb; }
.heading, .source-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.heading h2 { font-size: 14px; }
button, .official-source a { font-size: 11px; color: #494952; }
button:disabled { opacity: .45; cursor: default; }
button:focus-visible, select:focus-visible, input:focus-visible, .attach-button:focus-within { outline: 2px solid #74747c; outline-offset: 3px; }
.assistant-note { display: flex; align-items: flex-start; gap: 10px; margin: 18px 0; }
.assistant-mark { display: flex; align-items: center; justify-content: center; flex-shrink: 0; width: 27px; height: 27px; border-radius: 7px; background: #19191b; color: white; font-size: 14px; }
.assistant-note p { background: #f5f5f7; border-radius: 10px; padding: 11px 13px; color: #72727c; font-size: 11px; line-height: 1.7; }
.contract-link { display: flex; gap: 9px; align-items: center; padding: 12px; border: 1px solid #e8e8ec; border-radius: 9px; font-size: 11px; }
.contract-link > span { flex: 1; }
.document-messages { display: flex; flex-direction: column; gap: 12px; margin-top: 16px; }
.document-message { border: 1px solid #e5e5ea; border-radius: 11px; overflow: hidden; }
.file-heading { display: flex; gap: 9px; align-items: center; padding: 12px; background: #fafafa; }
.filename { flex: 1; min-width: 0; text-align: left; overflow-wrap: anywhere; }
.file-heading > span { font-size: 9px; color: #587963; }
.reading { padding: 12px; }
.reading-label { font-size: 9px; color: #8a8a94; }
.reading p, .reading li { font-size: 11px; line-height: 1.7; color: #676771; }
.reading p { margin-top: 6px; }
.reading ul { padding-left: 16px; margin: 8px 0; }
.reading .uncertainty { color: #8a7461; font-size: 10px; }
.extracted-amount { display: flex; align-items: center; justify-content: space-between; margin: 12px 0; font-size: 11px; }
.check-row { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; padding: 8px 0; border-top: 1px solid #efeff2; font-size: 10px; color: #777781; line-height: 1.6; }
.check-row > span:last-child { flex-shrink: 0; color: #967950; }
.check-row .consistent { color: #4f8260 !important; }
.citation { display: inline-flex; align-items: center; gap: 5px; margin-top: 8px; font-size: 10px; text-decoration: underline; }
.not-read { display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 11px 12px; font-size: 10px; color: #93939c; }
.attachment-composer { border: 1px solid #dfdfe5; border-radius: 12px; padding: 14px; margin-top: 15px; }
.category-label { display: flex; align-items: center; gap: 12px; justify-content: space-between; color: #777781; font-size: 10px; }
.category-label select { max-width: 60%; border: 1px solid #e3e3e8; border-radius: 6px; padding: 6px 8px; background: white; font-size: 10px; color: #575760; }
.consent { display: flex; align-items: flex-start; gap: 8px; font-size: 10px; color: #7b7b85; line-height: 1.6; margin-top: 14px; }
.consent input { margin-top: 2px; accent-color: #202023; }
.attach-row { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: 14px; }
.attach-row > span, .storage-note { font-size: 9px; line-height: 1.6; color: #9696a0; }
.attach-button { position: relative; overflow: hidden; display: inline-flex; align-items: center; gap: 6px; flex-shrink: 0; padding: 9px 12px; border-radius: 7px; background: #202023; color: white; font-size: 11px; cursor: pointer; }
.attach-button input { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
.attach-button.disabled { opacity: .45; }
.storage-note { margin-top: 10px; }
.official-source { margin-top: 18px; border: 1px solid #e6e6eb; border-radius: 11px; padding: 13px; }
.source-heading h3 { font-size: 12px; }
.source-heading span { display: block; margin-top: 4px; font-size: 10px; color: #92929c; }
.official-source p { color: #7c7c87; font-size: 11px; line-height: 1.7; margin-top: 12px; }
.official-source .source-meta { font-size: 10px; color: #9999a2; }
.official-source ul { padding-left: 16px; margin-top: 10px; }
.official-source li { font-size: 10px; line-height: 1.7; color: #73737c; }
.official-source a { display: inline-flex; align-items: center; gap: 5px; margin-top: 10px; text-decoration: underline; }
.official-source small { display: block; margin-top: 9px; font-size: 9px; line-height: 1.7; color: #9696a0; }
.evidence-error { background: #fff5f4; color: #a13d3d; padding: 12px; font-size: 11px; line-height: 1.7; border-radius: 8px; margin-top: 12px; }
</style>
