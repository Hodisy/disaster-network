<script setup>
const { t } = useAppI18n()
defineProps({ phase: String, complete: Boolean, turn: Number, transcript: String, response: String, orbState: String, label: String, live: Boolean, busy: Boolean, microphoneState: String, entries: { type: Array, default: () => [] }, transcriptTruncated: Boolean })
defineEmits(['speak', 'end'])
</script>

<template>
  <section class="voice-session" :class="{ live }" :aria-label="t(live ? 'Insurance support call' : 'Scripted conversation')">
    <div class="orb-stage">
      <button class="orb-button" :class="{ 'is-listening': phase === 'listening' }" :disabled="complete || busy" :aria-label="t(live ? (['listening', 'responding'].includes(phase) ? 'Interrupt the assistant' : 'Start call') : complete ? 'Conversation complete' : ['listening', 'responding'].includes(phase) ? 'Interrupt the call' : turn ? 'Continue the call' : 'Start the call')" @click="$emit('speak')">
        <ThinkingOrb :state="orbState" :size="240" />
        <span class="orb-hover-icon"><UiIcon :name="['listening', 'responding'].includes(phase) ? 'Square' : complete ? 'Check' : 'Mic'" :size="20" /></span>
      </button>
      <div class="voice-state" aria-live="polite"><span v-if="phase === 'listening'" class="status-dot listening-dot" />{{ t(label) }}</div>
      <div class="call-controls"><span v-if="live && microphoneState !== 'off'" class="microphone-state">{{ t(microphoneState === 'on' ? 'Microphone is on' : 'Microphone is muted') }}</span><button v-if="live && ['connecting', 'listening', 'responding', 'closing'].includes(phase)" class="end-call" :disabled="phase === 'closing'" :aria-label="t('End call')" :title="t('End call')" @click="$emit('end')"><UiIcon name="Phone" :size="16" /></button></div>
    </div>
    <div class="conversation">
      <div class="conversation-heading"><span>{{ t('The conversation') }}</span></div>
      <template v-if="live && entries.length">
        <div v-for="entry in entries" :key="entry.id" class="utterance" :class="{ 'adviser-utterance': entry.speaker === 'assistant' }">
          <span v-if="entry.speaker === 'user'" class="avatar tiny">J</span><span v-else class="adviser-mark"><span>//</span></span>
          <div><div class="speaker-label">{{ entry.speaker === 'user' ? t('You') : 'Allstorm' }}</div><p>{{ entry.text }}</p></div>
        </div>
        <p v-if="transcriptTruncated" class="transcript-note">{{ t('Earlier transcript text is no longer displayed. Recent speech remains visible.') }}</p>
      </template>
      <template v-else>
      <div class="utterance">
        <span class="avatar tiny">J</span>
        <div><div class="speaker-label">{{ t('You') }}</div><p :class="{ 'empty-text': !transcript }">{{ transcript ? (live ? transcript : t(transcript)) : t('Your transcript will appear here. Click the orb to get started.') }}</p></div>
      </div>
      <div class="utterance adviser-utterance">
        <span class="adviser-mark"><span>//</span></span>
        <div><div class="speaker-label">Allstorm<span>{{ t('Your assistant') }}</span></div><div v-if="phase === 'responding' && !live" class="typing-dots" :aria-label="t('Preparing the reply')"><i /><i /><i /></div><p v-else :class="{ 'empty-text': !response }">{{ response ? (live ? response : t(response)) : t('I am here to help, at your pace.') }}</p></div>
      </div>
      </template>
    </div>
  </section>
</template>

<style scoped>
.voice-session.live .utterance p { white-space: pre-wrap; }
.call-controls { display: flex; align-items: center; gap: 12px; margin-top: 8px; }
.microphone-state, .transcript-note { color: #92929c; font-size: 11px; line-height: 1.6; }
.end-call { display: inline-flex; align-items: center; justify-content: center; width: 32px; height: 32px; border: 1px solid #eddcdd; color: #984549; border-radius: 50%; }
.end-call svg { transform: rotate(135deg); }
.end-call:hover { background: #fff6f6; }
.end-call:disabled { opacity: .5; }
.utterance > div { flex: 1; }
.avatar { display: inline-flex; flex-shrink: 0; align-items: center; justify-content: center; width: 33px; height: 33px; border-radius: 50%; background: #19191b; color: #fff; font-size: 11px; font-weight: 500; letter-spacing: -.2px; }
.avatar.insurer-avatar { background: #e8e8eb; color: #57575e; }
.avatar.tiny { width: 38px; height: 38px; background: #f0f0f2; color: #46464d; font-size: 12px; }
.avatar.large { width: 52px; height: 52px; font-size: 18px; }
.orb-stage { display: flex; flex-direction: column; align-items: center; padding-top: 10px; }
.orb-button { position: relative; display: flex; align-items: center; justify-content: center; width: 360px; height: 360px; max-width: 100%; padding: 0; border-radius: 50%; }
.orb-button canvas { display: block; width: 330px !important; height: 330px !important; transition: transform .4s ease; }
.orb-button:hover canvas { transform: scale(1.025); }
.orb-button:disabled { cursor: default; }
.orb-button:disabled:hover canvas { transform: none; }
.orb-hover-icon { position: absolute; display: flex; align-items: center; justify-content: center; inset: 0; width: 43px; height: 43px; margin: auto; border: 1px solid #e3e3e6; border-radius: 50%; background: #ffffffed; opacity: 0; transform: translateY(4px); transition: opacity .18s, transform .18s; }
.orb-button:hover .orb-hover-icon, .orb-button:focus-visible .orb-hover-icon { opacity: 1; transform: translateY(0); }
.voice-state { display: flex; align-items: center; justify-content: center; gap: 9px; min-height: 24px; font-size: 14px; line-height: 1.5; }
.status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: var(--green); }
.listening-dot { animation: soft-pulse 1.7s ease-in-out infinite; }
.conversation { margin-top: 40px; }
.conversation-heading { display: flex; align-items: center; justify-content: space-between; gap: 15px; padding-bottom: 17px; border-bottom: 1px solid var(--border); font-size: 17px; font-weight: 600; letter-spacing: -.25px; }
.utterance { display: flex; gap: 22px; align-items: flex-start; padding: 24px 0 22px; }
.utterance + .utterance { border-top: 1px solid var(--border); }
.utterance > div { min-width: 0; padding-top: 2px; }
.speaker-label { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; font-size: 12px; line-height: 1.5; font-weight: 500; }
.speaker-label > span, .speaker-label time { color: #92929c; font-weight: 400; font-size: 11px; }
.speaker-label > span::before, .speaker-label time::before { content: '·'; margin-right: 8px; color: #a0a0a7; }
.utterance p { margin-top: 6px; font-size: 15px; line-height: 1.65; font-weight: 400; letter-spacing: -.15px; overflow-wrap: anywhere; }
.utterance p.empty-text { color: #93939d; }
.adviser-mark { flex-shrink: 0; display: inline-flex; align-items: center; justify-content: center; width: 38px; height: 38px; background: #171719; color: white; border-radius: 10px; font-size: 20px; font-style: italic; font-weight: 650; letter-spacing: -3px; padding-right: 3px; }
.typing-dots { display: flex; align-items: center; gap: 4px; height: 28px; }
.typing-dots i { width: 4px; height: 4px; border-radius: 50%; background: #b2b2bb; animation: typing 1.1s infinite; }
.typing-dots i:nth-child(2) { animation-delay: .15s; }
.typing-dots i:nth-child(3) { animation-delay: .3s; }
.identity-card > .avatar { flex-shrink: 0; }
.message .speaker-label { font-size: 10px; }
.message .speaker-label time { font-size: 9px; }
.message .avatar.tiny, .small-mark { width: 26px; height: 26px; font-size: 8px; }
@keyframes typing { 0%, 70%, 100% { opacity: .4; transform: translateY(0); } 35% { opacity: 1; transform: translateY(-3px); } }
@keyframes soft-pulse { 50% { opacity: .45; } }
@media (prefers-reduced-motion: reduce) { .listening-dot, .typing-dots i { animation: none; } .orb-button canvas, .orb-hover-icon { transition: none; } }
</style>
