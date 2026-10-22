<script setup>
const { t, formatTime } = useAppI18n()
defineProps({ messages: Array, draft: String, busy: Boolean })
defineEmits(['update:draft', 'send'])
</script>

<template>
  <section class="message-thread" :aria-label="t('Simulated messaging')">
    <div class="section-heading"><h2>{{ t('Stay in touch') }}</h2><span class="subtle-tag">{{ t('Simulated messages') }}</span></div>
    <div class="messages" role="log" aria-live="polite" :aria-label="t('Message thread')"><div v-for="message in messages" :key="message.id" class="message" :class="message.role"><span :class="message.role === 'client' ? 'avatar tiny' : 'adviser-mark small-mark'">{{ message.role === 'client' ? 'J' : '//' }}</span><div><div class="speaker-label">{{ message.role === 'client' ? t('You') : 'Allstorm' }}<time :datetime="message.createdAt">{{ message.createdAt ? formatTime(message.createdAt) : message.time }}</time></div><p>{{ message.role === 'client' ? message.text : t(message.text) }}</p></div></div><div v-if="busy" class="message adviser"><span class="adviser-mark small-mark">//</span><div class="typing-dots" :aria-label="t('Preparing the simulated reply')"><i /><i /><i /></div></div></div>
    <form class="message-composer" @submit.prevent="$emit('send')"><textarea :value="draft" rows="2" maxlength="2000" :aria-label="t('Your message')" :placeholder="t('Write a message…')" :disabled="busy" @input="$emit('update:draft', $event.target.value)" @keydown.enter.exact.prevent="$emit('send')" /><div><span>{{ t('Press Enter to send') }}</span><button class="send-button" :disabled="busy || !draft.trim()" :aria-label="t('Send message')"><UiIcon name="ArrowUp" :size="16" /></button></div></form>
  </section>
</template>

<style scoped>
.avatar { display: inline-flex; flex-shrink: 0; align-items: center; justify-content: center; width: 33px; height: 33px; border-radius: 50%; background: #19191b; color: #fff; font-size: 11px; font-weight: 500; letter-spacing: -.2px; }
.avatar.insurer-avatar { background: #e8e8eb; color: #57575e; }
.avatar.tiny { width: 38px; height: 38px; background: #f0f0f2; color: #46464d; font-size: 12px; }
.avatar.large { width: 52px; height: 52px; font-size: 18px; }
.subtle-tag { display: inline-flex; align-items: center; gap: 6px; padding: 5px 9px; border-radius: 6px; background: #f3f3f4; color: #62626b; font-size: 10px; font-weight: 400; line-height: 1.2; white-space: nowrap; }
.subtle-tag .small-dot { width: 5px; height: 5px; }
.speaker-label { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; font-size: 12px; line-height: 1.5; font-weight: 500; }
.speaker-label > span, .speaker-label time { color: #92929c; font-weight: 400; font-size: 11px; }
.speaker-label > span::before, .speaker-label time::before { content: '·'; margin-right: 8px; color: #a0a0a7; }
.adviser-mark { flex-shrink: 0; display: inline-flex; align-items: center; justify-content: center; width: 38px; height: 38px; background: #171719; color: white; border-radius: 10px; font-size: 20px; font-style: italic; font-weight: 650; letter-spacing: -3px; padding-right: 3px; }
.typing-dots { display: flex; align-items: center; gap: 4px; height: 28px; }
.typing-dots i { width: 4px; height: 4px; border-radius: 50%; background: #b2b2bb; animation: typing 1.1s infinite; }
.typing-dots i:nth-child(2) { animation-delay: .15s; }
.typing-dots i:nth-child(3) { animation-delay: .3s; }
.identity-card > .avatar { flex-shrink: 0; }
.identity-card p, .identity-card .subtle-tag { color: #85858f; font-size: 10px; line-height: 1.7; }
.identity-card .subtle-tag { background: none; border-radius: 0; padding: 0; }
.message-thread { margin-top: 6px; padding-top: 18px; border-top: 1px solid var(--border); }
.section-heading { display: flex; justify-content: space-between; align-items: center; gap: 10px; }
.section-heading h2 { font-size: 13px; letter-spacing: -.2px; }
.messages { display: flex; flex-direction: column; gap: 16px; padding: 20px 0 15px; max-height: 320px; overflow-y: auto; overscroll-behavior: contain; }
.message { display: flex; gap: 12px; min-width: 0; align-items: flex-start; }
.message > div { min-width: 0; flex: 1; background: #f5f5f7; border-radius: 9px; padding: 9px 12px 10px; }
.message.client > div { background: #eeeeef; }
.message .speaker-label { font-size: 10px; }
.message .speaker-label time { font-size: 9px; }
.message p { font-size: 11px; line-height: 1.65; color: #757580; overflow-wrap: anywhere; margin-top: 3px; }
.message .avatar.tiny, .small-mark { width: 26px; height: 26px; font-size: 8px; }
.small-mark { border-radius: 7px; font-size: 14px; letter-spacing: -2px; padding-right: 2px; }
.message-composer { position: relative; border: 1px solid #e1e1e6; border-radius: 14px; padding: 11px 12px 8px; }
.message-composer:focus-within { border-color: #b4b4be; }
.message-composer textarea { display: block; width: 100%; min-height: 37px; max-height: 180px; background: transparent; color: var(--ink); resize: vertical; border: 0; padding: 0; font-size: 11px; line-height: 1.6; outline: none; }
.message-composer textarea::placeholder { color: #a0a0ab; }
.message-composer > div { display: flex; align-items: center; justify-content: space-between; gap: 15px; }
.message-composer > div > span { font-size: 9px; color: #b0b0b9; }
.send-button { display: flex; align-items: center; justify-content: center; width: 28px; height: 28px; border-radius: 50%; background: var(--ink); color: #fff; }
.send-button:hover { background: #424247; }
.send-button:disabled { background: #dfdfe4; }
@keyframes typing { 0%, 70%, 100% { opacity: .4; transform: translateY(0); } 35% { opacity: 1; transform: translateY(-3px); } }
@media (prefers-reduced-motion: reduce) { .typing-dots i { animation: none; } }
</style>
