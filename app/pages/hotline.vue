<script setup>
const { state, speak, stopCall, openClaim } = useDemo()
const call = computed(() => state.value.call)
const orbState = computed(() => ({ listening: 'listening', responding: 'solving', ready: call.value.complete ? 'searching' : 'weaving', paused: 'searching' })[call.value.phase] || 'searching')
const label = computed(() => {
  if (call.value.complete) return 'Votre déclaration est prête'
  return ({ listening: 'Je vous écoute…', responding: 'Un instant, je prépare votre réponse…', paused: 'Appel interrompu · Cliquez pour reprendre', ready: 'Cliquez pour continuer' })[call.value.phase] || 'Cliquez pour commencer'
})
async function createClaim() { await openClaim(); if (state.value.claim) await navigateTo('/client') }
onBeforeUnmount(stopCall)
useHead({ title: 'Hotline — Disaster Network' })
</script>

<template>
  <div class="hotline-page">
    <div class="hotline-heading"><h1>On vous écoute.</h1><p>Racontez ce qui s’est passé. On s’occupe de la suite.</p></div>
    <VoiceSession :phase="call.phase" :complete="call.complete" :turn="call.turn" :transcript="call.transcript" :response="call.response" :orb-state="orbState" :label="label" @speak="speak" />
    <div v-if="state.error" class="inline-error" role="alert">{{ state.error }}</div>
    <div v-if="call.complete || state.claim" class="claim-ready"><span><UiIcon name="CircleCheck" :size="17" />{{ state.claim ? 'Votre dossier est ouvert.' : 'Tout est prêt pour ouvrir votre dossier.' }}</span><UiButton v-if="state.claim" to="/client" class="dark">Voir mon dossier<UiIcon name="ArrowRight" :size="15" /></UiButton><UiButton v-else class="dark" :disabled="!!state.busy" @click="createClaim">{{ state.busy === 'open' ? 'Ouverture…' : 'Ouvrir mon dossier' }}<UiIcon name="ArrowRight" :size="15" /></UiButton></div>
    <footer class="quiet-footer"><UiIcon name="ShieldCheck" :size="14" /><span>Rivage Assurance est fictive. Scénario d’inondation à Istres, le 22 octobre 2026.</span></footer>
  </div>
</template>

<style scoped>
.inline-error { margin: 16px 0; padding: 12px 15px; color: #a13d3d; background: #fff5f4; border: 1px solid #f1d8d5; border-radius: 8px; font-size: 12px; line-height: 1.6; }
.hotline-page { max-width: 1160px; margin: 0 auto; padding: 44px 52px 24px; }
.hotline-heading { text-align: center; }
.hotline-heading h1 { font-size: 36px; font-weight: 680; letter-spacing: -1.45px; line-height: 1.3; }
.hotline-heading p { margin-top: 7px; font-size: 16px; line-height: 1.6; color: #73737e; letter-spacing: -.2px; }
.claim-ready { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-top: 15px; padding: 17px 0; border-top: 1px solid var(--border); }
.claim-ready > span { display: flex; align-items: center; gap: 8px; font-size: 12px; color: #686872; }
.claim-ready > span svg { color: #39ac6b; }
.quiet-footer { display: flex; align-items: center; gap: 7px; margin-top: 27px; color: #a0a0aa; font-size: 10px; line-height: 1.6; }
</style>
