<script setup>
defineProps({ phase: String, complete: Boolean, turn: Number, transcript: String, response: String, orbState: String, label: String })
defineEmits(['speak'])
</script>

<template>
  <section class="voice-session" aria-label="Conversation de démonstration">
    <div class="orb-stage">
      <button class="orb-button" :class="{ 'is-listening': phase === 'listening' }" :disabled="complete" :aria-label="complete ? 'Conversation terminée' : ['listening', 'responding'].includes(phase) ? 'Interrompre l’appel simulé' : turn ? 'Continuer l’appel simulé' : 'Démarrer l’appel simulé'" @click="$emit('speak')">
        <ThinkingOrb :state="orbState" :size="240" />
        <span class="orb-hover-icon"><UiIcon :name="['listening', 'responding'].includes(phase) ? 'Square' : complete ? 'Check' : 'Mic'" :size="20" /></span>
      </button>
      <div class="voice-state" aria-live="polite"><span v-if="phase === 'listening'" class="status-dot listening-dot" />{{ label }}</div>
      <p class="voice-disclaimer">Appel simulé<span>·</span>Aucun micro activé</p>
    </div>
    <div class="conversation">
      <div class="conversation-heading"><span>La conversation</span><span v-if="turn" class="conversation-count">{{ turn }} / 3 échanges</span><span v-else class="conversation-count">En toute simplicité</span></div>
      <div class="utterance">
        <span class="avatar tiny">CM</span>
        <div><div class="speaker-label">Vous<span>Transcription simulée</span></div><p :class="{ 'empty-text': !transcript }">{{ transcript || 'Votre transcription apparaîtra ici. Cliquez sur l’orbe pour commencer.' }}</p></div>
      </div>
      <div class="utterance adviser-utterance">
        <span class="adviser-mark"><span>//</span></span>
        <div><div class="speaker-label">Rivage<span>Votre assistant</span></div><div v-if="phase === 'responding'" class="typing-dots" aria-label="Préparation de la réponse simulée"><i /><i /><i /></div><p v-else :class="{ 'empty-text': !response }">{{ response || 'Je suis là pour vous accompagner, à votre rythme.' }}</p></div>
      </div>
    </div>
  </section>
</template>

<style scoped>
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
.voice-disclaimer { margin-top: 7px; font-size: 11px; line-height: 1.6; color: #92929c; }
.voice-disclaimer > span { margin: 0 7px; }
.conversation { margin-top: 40px; }
.conversation-heading { display: flex; align-items: center; justify-content: space-between; gap: 15px; padding-bottom: 17px; border-bottom: 1px solid var(--border); font-size: 17px; font-weight: 600; letter-spacing: -.25px; }
.conversation-count { color: #86868f; font-size: 12px; font-weight: 400; letter-spacing: 0; }
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
