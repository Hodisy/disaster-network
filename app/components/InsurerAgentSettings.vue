<script setup>
const { t } = useAppI18n()
defineProps({ authorized: Boolean, accessKey: String, busy: String, error: String })
defineEmits(['update:accessKey', 'unlock'])
</script>
<template>
  <section class="agent-settings"><h2>{{ t('Insurer AI access') }}</h2><p>{{ t('The insurer agent reads stored dossiers, contracts, evidence and resource capacity. Access is separate from the client hotline.') }}</p><p v-if="authorized" class="ready">{{ t('Insurer access is unlocked for this browser.') }}</p><form v-else @submit.prevent="$emit('unlock')"><label for="insurer-key">{{ t('Private insurer access key') }}</label><div><input id="insurer-key" type="password" autocomplete="off" maxlength="256" :value="accessKey" :disabled="!!busy" @input="$emit('update:accessKey', $event.target.value)"><UiButton type="submit" class="dark" :disabled="!!busy || !accessKey.trim()">{{ t(busy === 'unlock' ? 'Checking access…' : 'Unlock insurer agent') }}</UiButton></div></form><p>{{ t('The key stays on the server. Browser access lasts 30 minutes. AI proposals require a separate human review before applying.') }}</p><p v-if="error" class="error" role="alert">{{ t(error) }}</p></section>
</template>
<style scoped>
.agent-settings { border: 1px solid var(--border); border-radius: 14px; padding: 24px; }
h2 { font-size: 15px; } p { margin-top: 10px; color: #777780; font-size: 12px; line-height: 1.8; }
form { margin-top: 17px; } label { display: block; font-size: 12px; margin-bottom: 8px; } form > div { display: flex; gap: 10px; }
input { flex: 1; min-width: 0; border: 1px solid #dedee4; padding: 10px 12px; border-radius: 8px; font-size: 12px; }
.ready { color: #326743; } .error { color: #9b4141; }
</style>
