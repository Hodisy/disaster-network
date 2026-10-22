<script setup>
const { state, steps, stepDescriptions, documentRows, received, stageLabel, addFiles, removeFile, sendMessage, advance } = useDemo()
const draft = ref('')
const preview = ref(null)
const claim = computed(() => state.value.claim)
const nextAction = computed(() => ({ 1: 'Simuler la proposition', 2: 'Simuler l’assistance', 3: 'Simuler la clôture' })[claim.value?.stage])
const nextDescription = computed(() => ({ 0: 'Ajoutez les trois pièces pour compléter votre dossier.', 1: 'Vos pièces sont réunies. Vous pouvez simuler une proposition de garantie.', 2: 'Proposition fictive disponible. L’étape suivante simule l’organisation de l’assistance.', 3: 'L’assistance est organisée dans le scénario. Vous pouvez simuler la clôture.', 4: 'Le parcours de démonstration est terminé. Aucun paiement n’a été effectué.' })[claim.value?.stage])
async function send() { const text = draft.value; if (await sendMessage(text)) draft.value = '' }
useHead({ title: 'Mon dossier — Disaster Network' })
</script>

<template>
  <div v-if="!claim" class="empty-page"><span class="empty-page-icon"><UiIcon name="FileText" :size="26" /></span><h1>Votre dossier commence ici.</h1><p>Racontez d’abord ce qui s’est passé à notre assistant.<br>Votre déclaration et vos pièces se retrouveront ici.</p><UiButton to="/hotline" class="dark">Parler à l’assistant<UiIcon name="ArrowRight" :size="15" /></UiButton><small>Conversation de démonstration · Aucun appel réel</small></div>
  <div v-else class="client-page">
    <header class="claim-page-heading"><div><div class="claim-reference">{{ claim.id }}<span class="subtle-tag"><span class="small-dot" />{{ stageLabel }}</span></div><h1>Inondation à Istres</h1><p>Votre déclaration, vos documents et nos échanges.</p></div><span class="claim-date">22 octobre 2026</span></header>
    <div class="claim-workspace">
      <div class="claim-activity"><div class="activity-heading"><h2>Activité du dossier</h2><span class="subtle-tag">Simulation</span></div><div class="activity-date"><span />Jeudi 22 octobre 2026<span /></div>
        <div class="timeline-event"><span class="timeline-icon"><UiIcon :name="received === 3 ? 'Check' : 'Paperclip'" :size="16" /></span><div class="timeline-event-body"><div class="event-heading"><h3>{{ received === 3 ? 'Pièces réunies' : 'Compléter votre dossier' }}</h3><span>{{ received }} / 3 pièces</span></div><div class="task-card"><div class="task-card-intro"><h2>{{ received === 3 ? 'Merci, tout est là.' : 'Quelques pièces pour avancer.' }}</h2><p>Ajoutez vos documents et consultez-les à tout moment.</p></div><DocumentList :rows="documentRows" :busy="!!state.busy" :locked="claim.stage > 1" @add="addFiles" @preview="preview = $event" @remove="removeFile" /><p class="file-disclaimer">Photos JPG, PNG, WebP ou PDF · 10 Mo par fichier<br>Conservés uniquement dans cet onglet. Aucun envoi externe.</p></div><div class="next-step"><div><span>{{ stageLabel }}</span><p>{{ nextDescription }}</p></div><UiButton v-if="nextAction" class="dark small" :disabled="!!state.busy" @click="advance">{{ state.busy === 'advance' ? 'Simulation…' : nextAction }}<UiIcon name="ArrowRight" :size="14" /></UiButton></div></div></div>
        <div class="timeline-event completed-event"><span class="timeline-icon green"><UiIcon name="Check" :size="14" /></span><div class="timeline-event-body"><div class="event-heading"><h3>Déclaration reçue</h3><span>Scénario fictif</span></div><p class="event-description">{{ claim.summary }}</p></div></div>
        <div class="timeline-event completed-event last-event"><span class="timeline-icon green"><UiIcon name="Check" :size="14" /></span><div class="timeline-event-body"><div class="event-heading"><h3>Claire Martin identifiée</h3></div><p class="event-description">Identité fictive reconnue pendant l’appel simulé.</p></div></div>
        <div v-if="state.error" class="inline-error" role="alert">{{ state.error }}</div>
        <MessageThread :messages="state.messages" v-model:draft="draft" :busy="!!state.busy" @send="send" />
      </div>
      <ClaimDetails :claim="claim" :steps="steps" :step-descriptions="stepDescriptions" />
    </div>
    <UiDialog v-if="preview" :title="preview.name" wide @close="preview = null"><img v-if="preview.type === 'image'" class="file-image" :src="preview.url" :alt="`Aperçu local : ${preview.name}`" /><iframe v-else class="file-pdf" :src="preview.url" :title="`Aperçu PDF : ${preview.name}`" /><p class="preview-note">Aperçu local. Ce fichier reste dans votre navigateur.</p></UiDialog>
  </div>
</template>

<style scoped>
.small-dot { display: inline-block; width: 7px; height: 7px; flex-shrink: 0; background: #b1b1b7; border-radius: 50%; }
.subtle-tag { display: inline-flex; align-items: center; gap: 6px; padding: 5px 9px; border-radius: 6px; background: #f3f3f4; color: #62626b; font-size: 10px; font-weight: 400; line-height: 1.2; white-space: nowrap; }
.subtle-tag .small-dot { width: 5px; height: 5px; }
.inline-error { margin: 16px 0; padding: 12px 15px; color: #a13d3d; background: #fff5f4; border: 1px solid #f1d8d5; border-radius: 8px; font-size: 12px; line-height: 1.6; }
.claim-page-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 24px; padding: 29px 30px 25px; border-bottom: 1px solid var(--border); }
.claim-reference { display: flex; align-items: center; gap: 13px; font-size: 12px; line-height: 1.5; color: #72727d; }
.claim-page-heading h1 { font-size: 27px; line-height: 1.3; letter-spacing: -.85px; margin-top: 10px; }
.claim-page-heading p { color: #81818b; font-size: 13px; line-height: 1.6; margin-top: 4px; }
.claim-date { font-size: 12px; line-height: 1.5; color: #777780; padding-top: 15px; white-space: nowrap; }
.claim-workspace { display: flex; align-items: stretch; }
.claim-activity { min-width: 0; padding: 24px 30px 35px; }
.activity-heading { display: flex; align-items: center; gap: 10px; }
.activity-heading h2 { font-size: 17px; letter-spacing: -.4px; }
.activity-date { display: flex; align-items: center; gap: 14px; margin: 23px 0; font-size: 10px; color: #9898a2; line-height: 1.4; white-space: nowrap; }
.activity-date > span { flex: 1; border-top: 1px solid var(--border); }
.timeline-event { display: flex; align-items: flex-start; gap: 20px; position: relative; padding: 0 0 18px; }
.timeline-event::before { content: ''; position: absolute; left: 14px; top: 29px; bottom: -1px; width: 1px; background: #e7e7ea; }
.timeline-event.last-event::before { display: none; }
.timeline-icon { position: relative; display: flex; align-items: center; justify-content: center; flex-shrink: 0; width: 29px; height: 29px; border-radius: 50%; background: #171719; color: white; z-index: 1; box-shadow: 0 0 0 4px #fff; }
.timeline-icon.green { background: var(--green); }
.timeline-event-body { flex: 1; min-width: 0; }
.event-heading { display: flex; align-items: center; justify-content: space-between; gap: 9px; min-height: 29px; }
.event-heading h3 { font-size: 13px; font-weight: 550; line-height: 1.5; }
.event-heading > span { font-size: 10px; line-height: 1.4; color: #9999a3; white-space: nowrap; }
.task-card { border: 1px solid #e6e6e9; border-radius: 15px; margin-top: 10px; overflow: hidden; }
.task-card-intro { padding: 15px 18px 14px; }
.task-card-intro h2 { font-size: 13px; line-height: 1.5; letter-spacing: -.2px; }
.task-card-intro p { font-size: 11px; line-height: 1.6; color: #9898a2; margin-top: 3px; }
.file-disclaimer { border-top: 1px solid #ededf0; padding: 11px 18px; color: #9999a3; font-size: 10px; line-height: 1.6; }
.next-step { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 10px 14px; margin: 12px 0 0; }
.next-step > div > span { font-size: 10px; font-weight: 500; color: #65656f; }
.next-step p { margin-top: 3px; color: #9b9ba4; font-size: 10px; line-height: 1.55; }
.next-step .button { border-radius: 20px; }
.completed-event { padding-bottom: 17px; }
.event-description { font-size: 11px; color: #9696a0; line-height: 1.65; margin-top: 1px; overflow-wrap: anywhere; }
.file-image { display: block; max-width: 100%; max-height: 600px; margin: 0 auto; object-fit: contain; border-radius: 6px; }
.file-pdf { width: 100%; height: 600px; border: 0; background: #f4f4f6; }
.preview-note { font-size: 11px; text-align: center; margin-top: 14px; color: #8d8d98; }
.empty-page { min-height: 700px; display: flex; align-items: center; justify-content: center; flex-direction: column; gap: 18px; padding: 45px 24px; text-align: center; }
.empty-page-icon { display: flex; align-items: center; justify-content: center; width: 64px; height: 64px; border: 1px solid var(--border); border-radius: 18px; color: #858590; margin-bottom: 6px; }
.empty-page h1 { font-size: 27px; letter-spacing: -.8px; line-height: 1.3; }
.empty-page p { color: #888893; font-size: 13px; line-height: 1.8; }
.empty-page small { color: #aaaab2; font-size: 10px; margin-top: 4px; }
.claim-activity { flex: 1; }
</style>
