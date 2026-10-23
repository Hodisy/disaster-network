<script setup>
const { t, locale, setLocale } = useAppI18n()
const { pending, confirming, maintenanceKey, error, result, requestReset, cancelReset, resetDatabase, localConfirming, resetLocal } = useDemoSettings()
const { source } = useHotlineClaim()
const { access, restoreAccess } = useHotlineAccess()
const { data: insurerData, accessKey: insurerKey, busy: insurerBusy, error: insurerError, restore: restoreInsurer, unlock: unlockInsurer } = useInsurerAgent()
onMounted(() => { void restoreAccess(); void restoreInsurer() })
useHead(() => ({ title: t('Settings — Disaster Network') }))
</script>

<template>
  <div class="settings-page">
    <header><span>{{ t('Workspace controls') }}</span><h1>{{ t('Settings') }}</h1><p>{{ t('Configure voice access and manage the evaluation workspace.') }}</p></header>
    <section class="language-settings" :aria-label="t('Language')">
      <h2>{{ t('Language') }}</h2>
      <LanguageToggle :locale="locale" @change="setLocale" />
    </section>
    <HotlineSettingsPanel v-model:mode="source" :authorized="access.authorized" :checking="access.checking" :error="access.error ? hotlineErrorPhrase(access.error) : ''" />
    <InsurerAgentSettings :authorized="insurerData.authorized" v-model:access-key="insurerKey" :busy="insurerBusy" :error="insurerError" @unlock="unlockInsurer" />
    <LocalDemoResetPanel :confirming="localConfirming" @request="localConfirming = true" @cancel="localConfirming = false" @confirm="resetLocal" />
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
.settings-page > section + section { margin-top: 18px; }
.language-settings { display: flex; align-items: center; justify-content: space-between; padding: 20px; border: 1px solid var(--border); border-radius: 10px; background: #fff; }
.language-settings h2 { font-size: 14px; font-weight: 600; }
@media (max-width: 760px) {
  .settings-page { padding: 24px 16px; }
  header { margin-bottom: 20px; }
}
@media (max-width: 760px) {
  .settings-page :deep(input:not([type='checkbox'])), .settings-page :deep(textarea) { font-size: 16px; min-width: 0; }
}
</style>
