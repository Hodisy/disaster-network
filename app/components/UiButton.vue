<script setup>
defineProps({ to: String, type: { type: String, default: 'button' }, disabled: Boolean })
const link = resolveComponent('NuxtLink')
</script>

<template>
  <component :is="to ? link : 'button'" :to="to" :type="to ? undefined : type" :disabled="to ? undefined : disabled" class="button"><slot /></component>
</template>

<style scoped>
.button { display: inline-flex; align-items: center; justify-content: center; gap: 9px; padding: 11px 17px; min-height: 39px; border: 1px solid var(--border); background: #fff; border-radius: 8px; font-size: 12px; line-height: 18px; font-weight: 500; white-space: nowrap; transition: background .16s, border-color .16s; }
.button:hover { background: #f6f6f7; border-color: #d7d7dc; }
.button.dark { border-color: var(--ink); background: var(--ink); color: white; }
.button.dark:hover { background: #353538; border-color: #353538; }
.button.small { min-height: 31px; padding: 8px 12px; font-size: 11px; line-height: 14px; }
.button:disabled { opacity: .45; cursor: default; }
</style>
