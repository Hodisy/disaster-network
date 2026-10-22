<script setup>
const { t } = useAppI18n()
defineProps({ title: String, wide: Boolean })
const emit = defineEmits(['close'])
const dialog = ref(null)
onMounted(() => dialog.value?.showModal())
function backdrop(event) { if (event.target === dialog.value) emit('close') }
</script>

<template>
  <Teleport to="body"><dialog ref="dialog" class="dialog" :class="{ wide }" aria-labelledby="dialog-title" @cancel.prevent="$emit('close')" @click="backdrop"><header><h2 id="dialog-title">{{ title }}</h2><button class="icon-button" :aria-label="t('Close preview')" @click="$emit('close')"><UiIcon name="X" /></button></header><div class="dialog-content"><slot /></div></dialog></Teleport>
</template>

<style scoped>
.dialog { border: 1px solid var(--border); border-radius: 16px; padding: 0; color: var(--ink); width: 620px; max-width: 94%; max-height: 90%; box-shadow: 0 20px 80px #00000020; }
.dialog.wide { width: 900px; }
.dialog::backdrop { background: #22222877; backdrop-filter: blur(3px); }
.dialog header { display: flex; justify-content: space-between; align-items: center; gap: 18px; padding: 15px 18px; border-bottom: 1px solid var(--border); }
.dialog h2 { font-size: 13px; overflow-wrap: anywhere; }
.dialog-content { padding: 18px; }
.icon-button { display: inline-flex; align-items: center; justify-content: center; width: 27px; height: 27px; padding: 0; border-radius: 6px; color: var(--muted); }
.icon-button:hover { background: #f1f1f3; color: var(--ink); }
</style>
