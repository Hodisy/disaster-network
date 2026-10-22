<script setup>
const { t } = useAppI18n()
defineProps({ mode: String, authorized: Boolean, checking: Boolean, error: String })
defineEmits(['update:mode'])
</script>

<template>
  <section class="settings-panel" :aria-label="t('Hotline setup')">
    <h2>{{ t('Hotline setup') }}</h2>
    <p>{{ t('Choose the conversation mode before presenting the application.') }}</p>
    <div class="mode-controls">
      <button :aria-pressed="mode === 'live'" @click="$emit('update:mode', 'live')">{{ t('Live AI voice') }}</button>
      <button :aria-pressed="mode === 'mock'" @click="$emit('update:mode', 'mock')">{{ t('Scripted conversation · Microphone off') }}</button>
    </div>
    <p>{{ t('Your voice is sent to OpenAI during the call. The assistant is AI; Allstorm and all insurance outcomes are fictional.') }}</p>
    <p>{{ t('In the verified Live dossier, photos and PDFs are saved in PostgreSQL. AI reading sends a selected file to OpenAI only after your agreement. Scripted previews remain local.') }}</p>
    <p v-if="authorized" class="access-ready" role="status"><UiIcon name="CircleCheck" :size="16" />{{ t('Voice calling is ready. No access key is required.') }}</p>
    <p v-else-if="checking" role="status">{{ t('Preparing voice calling…') }}</p>
    <p>{{ t('Voice calls are available without a workspace key. Policy verification is still required before accessing customer information.') }}</p>
    <p v-if="error" class="error-message" role="alert">{{ t(error) }}</p>
  </section>
</template>

<style scoped>
.settings-panel { border: 1px solid var(--border); border-radius: 14px; padding: 24px; background: #fff; }
h2 { font-size: 15px; line-height: 1.5; font-weight: 600; }
p { margin-top: 10px; font-size: 12px; line-height: 1.8; color: #777780; }
.mode-controls { display: flex; gap: 8px; margin-top: 18px; }
.mode-controls button { padding: 9px 14px; border: 1px solid var(--border); border-radius: 8px; font-size: 12px; }
.mode-controls button[aria-pressed="true"] { background: var(--ink); color: #fff; }
.access-ready { display: flex; align-items: center; gap: 8px; color: #326743; }
.error-message { color: #8e4141; }
</style>
