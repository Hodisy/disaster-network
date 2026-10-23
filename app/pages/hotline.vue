<script setup>
const { t } = useAppI18n()
const { state, speak, stopCall, openClaim } = useDemo()
const { source, refreshLiveClaim, claimError } = useHotlineClaim()
const { live, active, captionEntries, start, end, interruptSpeech, confirm, playAudio } = useLiveHotline()
const call = computed(() => source.value === 'live' ? { ...live, complete: false, turn: 0 } : state.value.call)
const orbState = computed(() => ({ listening: 'listening', responding: 'solving', connecting: 'weaving', closing: 'weaving', ready: call.value.complete ? 'searching' : 'weaving', paused: 'searching' })[call.value.phase] || 'searching')
const label = computed(() => {
  if (source.value === 'live') return ({ connecting: 'Connecting your call…', listening: 'Listening…', responding: 'Allstorm is speaking…', closing: 'Ending the call…', ended: 'Call ended', error: 'Call unavailable' })[live.phase] || 'Click to start your call'
  if (call.value.complete) return 'Your report is ready'
  return ({ listening: 'Listening…', responding: 'One moment, preparing your reply…', paused: 'Call interrupted · Click to resume', ready: 'Click to continue' })[call.value.phase] || 'Click to get started'
})
async function useOrb() {
  if (source.value === 'mock') return speak()
  if (active.value) return interruptSpeech()
  await start()
}
async function createClaim() { await openClaim(); if (state.value.claim) await navigateTo('/client') }
onMounted(() => { if (source.value === 'live') void refreshLiveClaim({ quiet: true }) })
onBeforeUnmount(stopCall)
useHead(() => ({ title: t('Hotline — Disaster Network') }))
</script>

<template>
  <div class="hotline-page">
    <div class="hotline-heading"><h1>{{ t('We are listening.') }}</h1><p>{{ t('Tell us what happened. We will help with the next steps.') }}</p></div>
    <VoiceSession :phase="call.phase" :complete="call.complete" :turn="call.turn" :transcript="call.transcript" :response="call.response" :entries="captionEntries" :transcript-truncated="live.transcriptTruncated" :orb-state="orbState" :label="label" :live="source === 'live'" :microphone-state="live.microphoneState" :busy="source === 'live' && ['connecting', 'closing'].includes(live.phase)" @speak="useOrb" @end="end()" />
    <UiButton v-if="source === 'live' && live.audioBlocked" class="audio-recovery small" @click="playAudio">{{ t('Play assistant audio') }}</UiButton>
    <div v-if="source === 'live' && live.error" class="inline-error" role="alert">{{ t(live.error) }}</div>
    <div v-if="source === 'live' && claimError" class="inline-error" role="alert">{{ t(hotlineErrorPhrase(claimError)) }}</div>
    <LiveReportReview v-if="source === 'live' && live.draft" :draft="live.draft" :busy="live.confirming" @confirm="confirm" />
    <p v-if="source === 'live' && live.saved" class="live-saved">{{ t('Report saved in PostgreSQL. Requested documents are available in My claim.') }}</p>
    <div v-if="source === 'mock' && state.error" class="inline-error" role="alert">{{ t(state.error) }}</div>
    <div v-if="(source === 'mock' && call.complete) || state.claim" class="claim-ready"><span><UiIcon name="CircleCheck" :size="17" />{{ t(state.claim ? 'Your claim is open.' : 'Everything is ready to open your claim.') }}</span><UiButton v-if="state.claim" to="/client" class="dark">{{ t('View my claim') }}<UiIcon name="ArrowRight" :size="15" /></UiButton><UiButton v-else class="dark" :disabled="!!state.busy" @click="createClaim">{{ t(state.busy === 'open' ? 'Opening…' : 'Open my claim') }}<UiIcon name="ArrowRight" :size="15" /></UiButton></div>
  </div>
</template>

<style scoped>
.audio-recovery { display: flex; margin: 18px auto; }
.live-saved { margin: 20px 0; color: #46755a; font-size: 12px; line-height: 1.6; }
.inline-error { margin: 16px 0; padding: 12px 15px; color: #a13d3d; background: #fff5f4; border: 1px solid #f1d8d5; border-radius: 8px; font-size: 12px; line-height: 1.6; }
.hotline-page { max-width: 1160px; margin: 0 auto; padding: 44px 52px 24px; }
.hotline-heading { text-align: center; }
.hotline-heading h1 { font-size: 36px; font-weight: 680; letter-spacing: -1.45px; line-height: 1.3; }
.hotline-heading p { margin-top: 7px; font-size: 16px; line-height: 1.6; color: #73737e; letter-spacing: -.2px; }
.claim-ready { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-top: 15px; padding: 17px 0; border-top: 1px solid var(--border); }
.claim-ready > span { display: flex; align-items: center; gap: 8px; font-size: 12px; color: #686872; }
.claim-ready > span svg { color: #39ac6b; }
@media (max-width: 760px) {
  .hotline-page { width: 100%; padding: 20px 18px 24px; }
  .hotline-heading h1 { font-size: 28px; line-height: 35px; letter-spacing: -.8px; }
  .hotline-heading p { font-size: 14px; line-height: 22px; }
  .claim-ready { flex-direction: column; align-items: stretch; gap: 12px; }
  .claim-ready > span { line-height: 19px; }
  .audio-recovery { min-height: 44px; white-space: normal; }
}
</style>
