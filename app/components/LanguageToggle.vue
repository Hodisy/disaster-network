<script setup>
defineProps({ locale: { type: String, required: true } })
defineEmits(['change'])
const { t } = useAppI18n()
</script>

<template>
  <div class="language-toggle" role="group" :aria-label="t('Language')">
    <button type="button" :aria-pressed="locale === 'en'" :aria-label="t('Switch to English')" :title="t('Switch to English')" @click="$emit('change', 'en')">EN</button>
    <button type="button" :aria-pressed="locale === 'fr'" :aria-label="t('Switch to French')" :title="t('Switch to French')" @click="$emit('change', 'fr')">FR</button>
  </div>
</template>

<style scoped>
.language-toggle { display: inline-flex; align-items: center; gap: 2px; padding: 3px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); }
.language-toggle button { min-width: 31px; padding: 5px 7px; border-radius: 5px; font-size: 10px; line-height: 14px; font-weight: 500; color: var(--muted); }
.language-toggle button:hover { color: var(--ink); background: #ededec; }
.language-toggle button[aria-pressed="true"] { background: #fff; color: var(--ink); box-shadow: 0 1px 3px #00000012; }
.language-toggle button:focus-visible { outline: 2px solid #74747c; outline-offset: 2px; }
@media (max-width: 760px) {
  .language-toggle button { min-width: 36px; min-height: 36px; }
}
</style>
