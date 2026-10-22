<script setup>
const { t } = useAppI18n()
defineProps({ confirming: Boolean })
defineEmits(['request', 'cancel', 'confirm'])
</script>

<template>
  <section class="settings-panel" :aria-label="t('Local workspace')">
    <h2>{{ t('Local workspace') }}</h2>
    <p>{{ t('Clear local progress, messages, file previews and resource plans. Saved database records and the selected language are retained.') }}</p>
    <div v-if="confirming" class="confirmation"><p>{{ t('Clear this workspace? Local files and messages will be removed.') }}</p><div class="actions"><UiButton @click="$emit('cancel')">{{ t('Cancel') }}</UiButton><UiButton class="dark" @click="$emit('confirm')">{{ t('Restart') }}</UiButton></div></div>
    <UiButton v-else class="reset-button" @click="$emit('request')"><UiIcon name="RotateCcw" :size="14" />{{ t('Clear local workspace') }}</UiButton>
  </section>
</template>

<style scoped>
.settings-panel { border: 1px solid var(--border); border-radius: 14px; padding: 24px; background: #fff; }
h2 { font-size: 15px; line-height: 1.5; font-weight: 600; }
p { margin-top: 8px; font-size: 12px; line-height: 1.8; color: #777780; }
.reset-button { margin-top: 18px; }
.confirmation { margin-top: 18px; padding-top: 12px; border-top: 1px solid var(--border); }
.actions { display: flex; gap: 8px; margin-top: 14px; }
</style>
