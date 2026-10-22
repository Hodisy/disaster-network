<script setup>
const { t } = useAppI18n()
const { pending, confirming, maintenanceKey, error, result, requestReset, cancelReset, resetDatabase } = useDemoSettings()
useHead(() => ({ title: t('Settings — Disaster Network') }))
</script>

<template>
  <div class="settings-page">
    <header><span>{{ t('Demo controls') }}</span><h1>{{ t('Settings') }}</h1><p>{{ t('A small maintenance area for preparing the next demonstration.') }}</p></header>
    <DemoSettingsPanel :pending="pending" :confirming="confirming" v-model:maintenance-key="maintenanceKey" :error="error" :result="result" @request-reset="requestReset" @cancel-reset="cancelReset" @confirm-reset="resetDatabase" />
    <NuxtLink to="/hotline" class="back-link"><UiIcon name="ArrowLeft" :size="14" />{{ t('Back to the hotline') }}</NuxtLink>
  </div>
</template>

<style scoped>
.settings-page { width: 100%; max-width: 830px; margin: 0 auto; padding: 46px 40px; }
header { margin-bottom: 27px; }
header > span { font-size: 10px; color: #9696a0; letter-spacing: 1px; text-transform: uppercase; }
h1 { margin-top: 9px; font-size: 28px; font-weight: 650; letter-spacing: -1px; line-height: 1.4; }
header p { margin-top: 8px; font-size: 13px; color: #80808a; line-height: 1.8; }
.back-link { display: inline-flex; align-items: center; gap: 7px; margin-top: 22px; font-size: 11px; color: #85858e; padding: 5px 0; }
.back-link:hover { color: var(--ink); }
.back-link:focus-visible { outline: 2px solid #74747c; outline-offset: 4px; }
</style>
