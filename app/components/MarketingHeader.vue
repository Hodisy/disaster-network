<script setup>
defineProps({ locale: { type: String, required: true } })
defineEmits(['set-locale'])
const { t } = useAppI18n()
</script>

<template>
  <header class="marketing-header">
    <a href="#top" class="wordmark"><img src="/brand/disaster-network-mark.svg" alt="" width="28" height="28" /><span>Disaster Network.</span></a>
    <nav :aria-label="t('Project navigation')">
      <a href="#jules">{{ t('How it works') }}</a>
      <a href="#operations">{{ t('Insurer workspace') }}</a>
      <a href="#scenarios">{{ t('Other scenarios') }}</a>
    </nav>
    <div class="header-actions">
      <LanguageToggle :locale="locale" @change="$emit('set-locale', $event)" />
      <NuxtLink to="/insurer" class="demo-link">{{ t('Explore the workspace') }}</NuxtLink>
    </div>
  </header>
</template>

<style scoped>
.marketing-header { position: sticky; top: 0; z-index: 25; display: flex; align-items: center; justify-content: space-between; gap: 24px; max-width: 1264px; min-height: 64px; margin: 0 auto; padding: 0 32px; background: #ffffffed; backdrop-filter: blur(12px); }
.wordmark { display: inline-flex; align-items: center; gap: 9px; font-size: 20px; line-height: 28px; font-weight: 700; letter-spacing: -.8px; white-space: nowrap; }
.wordmark img { display: block; width: 28px; height: 28px; }
.wordmark span { color: #171719; }
nav { display: flex; align-items: center; gap: 28px; }
nav a { font-size: 13px; line-height: 22px; font-weight: 450; color: #262626; white-space: nowrap; }
nav a:hover { color: #737373; }
.header-actions { display: flex; align-items: center; gap: 22px; }
.demo-link { display: inline-flex; align-items: center; justify-content: center; min-height: 34px; padding: 7px 17px; background: #ededed; border-radius: 24px; font-size: 13px; line-height: 20px; font-weight: 500; white-space: nowrap; }
.demo-link:hover { background: #dedede; }
</style>
