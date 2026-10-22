<script setup>
const { t } = useAppI18n()
defineProps({ rows: Array, busy: Boolean, locked: Boolean, localOnly: Boolean })
defineEmits(['add', 'preview', 'remove'])
function select(event, kind, emit) { emit('add', kind, event.target.files); event.target.value = '' }
</script>

<template>
  <div class="documents-list">
    <div v-for="row in rows" :key="row.id" class="document-row">
      <div class="document-main"><span class="document-icon" :class="{ received: row.files.length }"><UiIcon :name="row.files.length ? 'Check' : row.id === 'photos' ? 'Image' : 'FileText'" :size="17" /></span><div><h3>{{ t(row.title) }}</h3><p>{{ t(row.description) }}</p></div><label v-if="!locked" class="button small file-label" :class="{ disabled: busy }"><UiIcon name="Plus" :size="14" />{{ t(row.files.length ? 'Add' : 'Attach') }}<input type="file" :accept="row.id === 'photos' ? 'image/jpeg,image/png,image/webp' : 'image/jpeg,image/png,image/webp,application/pdf'" multiple :disabled="busy" :aria-label="t('Add: {document}', { document: t(row.title) })" @change="select($event, row.id, $emit)" /></label><span v-else class="received-label">{{ t(localOnly ? 'Not uploaded' : 'Received') }}</span></div>
      <ul v-if="row.files.length" class="file-list"><li v-for="file in row.files" :key="file.id"><button class="file-preview-button" :disabled="busy" @click="$emit('preview', file)"><UiIcon :name="file.type === 'application/pdf' ? 'FileText' : 'Image'" :size="14" /><span>{{ file.name }}</span><UiIcon name="ArrowUpRight" :size="13" /></button><button v-if="!locked && !file.persisted" class="icon-button" :disabled="busy" :aria-label="t('Remove {name}', { name: file.name })" @click="$emit('remove', file.id)"><UiIcon name="X" :size="13" /></button></li></ul>
    </div>
  </div>
</template>

<style scoped>
button:focus-visible, a:focus-visible, textarea:focus-visible, input:focus-visible, .file-label:focus-within { outline: 2px solid #74747c; outline-offset: 4px; }
.icon-button { display: inline-flex; align-items: center; justify-content: center; width: 27px; height: 27px; border-radius: 6px; color: #787882; padding: 0; }
.icon-button:hover { background: #f1f1f3; color: var(--ink); }
.icon-button:disabled { opacity: .4; }
.document-row { border-top: 1px solid #ededf0; }
.document-main { display: flex; align-items: center; gap: 13px; padding: 13px 18px; }
.document-icon { display: inline-flex; flex-shrink: 0; color: #27272c; }
.document-icon.received { color: #35af6b; }
.document-main > div { flex: 1; min-width: 0; }
.document-main h3 { font-size: 12px; font-weight: 500; line-height: 1.5; }
.document-main p { color: #92929c; font-size: 10px; line-height: 1.5; margin-top: 2px; }
.file-label { position: relative; overflow: hidden; border-color: transparent; background: #f4f4f5; padding: 8px 13px !important; min-width: 72px; cursor: pointer; }
.file-label input { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; font-size: 0; }
.file-label.disabled input { cursor: default; }
.file-label svg { display: none; }
.received-label { color: #37a467; font-size: 10px; }
.file-list { display: flex; flex-direction: column; gap: 4px; padding: 0 15px 11px 48px; }
.file-list li { display: flex; align-items: center; gap: 5px; border-radius: 6px; background: #f7f7f9; min-width: 0; }
.file-preview-button { display: flex; align-items: center; gap: 7px; flex: 1; min-width: 0; padding: 7px 8px; text-align: left; font-size: 10px; color: #64646e; }
.file-preview-button span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
.file-list .icon-button { height: 25px; width: 25px; margin-right: 3px; }
.file-label { display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0; border: 1px solid transparent; border-radius: 8px; font-size: 11px; line-height: 14px; font-weight: 500; }
.file-label.disabled { opacity: .45; cursor: default; }
</style>
