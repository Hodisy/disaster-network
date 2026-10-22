<script setup>
const { t } = useAppI18n()
defineProps({ pending: Boolean, confirming: Boolean, maintenanceKey: { type: String, default: '' }, error: String, result: Object })
defineEmits(['request-reset', 'cancel-reset', 'confirm-reset', 'update:maintenanceKey'])
</script>

<template>
  <section class="settings-panel" :aria-label="t('Reference data')" :aria-busy="pending">
    <div class="panel-heading"><span class="panel-icon"><UiIcon name="RotateCcw" :size="19" /></span><div><h2>{{ t('Reference data') }}</h2><p>{{ t('Restore the initial evaluation records, including Jules and his policy.') }}</p></div></div>
    <p class="scope-note">{{ t('This clears claim progress, messages and files in the evaluation database. Other database schemas are not affected.') }}</p>
    <div v-if="result" class="success-message" role="status"><UiIcon name="CircleCheck" :size="17" /><p>{{ t('Database restored: {customers} customers, {policies} policies and {claims} claim records.', result) }}</p></div>
    <p v-if="error" class="error-message" role="alert">{{ t(error) }}</p>
    <form v-if="confirming" class="confirmation" @submit.prevent="$emit('confirm-reset')">
      <h3>{{ t('Restore the initial records?') }}</h3>
      <p>{{ t('All current evaluation records will be replaced by the initial data. Your current progress cannot be recovered through this action.') }}</p>
      <label for="maintenance-key">{{ t('Maintenance key') }}</label>
      <input id="maintenance-key" type="password" :value="maintenanceKey" :disabled="pending" maxlength="256" autocomplete="off" spellcheck="false" required :placeholder="t('Enter the server maintenance key')" @input="$emit('update:maintenanceKey', $event.target.value)" />
      <small>{{ t('The key is used only for this request and is not saved in your browser storage.') }}</small>
      <div class="actions"><UiButton :disabled="pending" @click="$emit('cancel-reset')">{{ t('Cancel') }}</UiButton><UiButton type="submit" class="dark" :disabled="pending || !maintenanceKey.trim()"><UiIcon name="RotateCcw" :size="15" />{{ t(pending ? 'Restoring…' : 'Confirm database reset') }}</UiButton></div>
    </form>
    <div v-else class="panel-footer"><span>{{ t('Database reset is separate from clearing this browser workspace.') }}</span><UiButton class="small" @click="$emit('request-reset')"><UiIcon name="RotateCcw" :size="14" />{{ t('Restore reference data') }}</UiButton></div>
  </section>
</template>

<style scoped>
.settings-panel { border: 1px solid var(--border); border-radius: 14px; padding: 24px; background: #fff; }
.panel-heading { display: flex; gap: 13px; align-items: flex-start; }
.panel-icon { display: flex; align-items: center; justify-content: center; width: 38px; height: 38px; border: 1px solid var(--border); border-radius: 10px; flex-shrink: 0; }
.panel-heading h2 { font-size: 15px; line-height: 1.5; font-weight: 600; }
.panel-heading p { margin-top: 4px; font-size: 12px; line-height: 1.7; color: #777780; }
.scope-note { font-size: 12px; color: #777780; line-height: 1.8; margin-top: 19px; max-width: 570px; }
.panel-footer { display: flex; gap: 18px; align-items: center; justify-content: space-between; border-top: 1px solid var(--border); padding-top: 18px; margin-top: 22px; }
.panel-footer > span { font-size: 10px; color: #9696a0; line-height: 1.6; }
.confirmation { display: flex; flex-direction: column; margin-top: 22px; padding: 20px; border: 1px solid var(--border); background: #fafafa; border-radius: 10px; }
.confirmation h3 { font-size: 13px; line-height: 1.6; }
.confirmation p { margin-top: 5px; font-size: 12px; line-height: 1.8; color: #777780; }
.confirmation label { font-size: 11px; font-weight: 500; margin-top: 18px; margin-bottom: 7px; }
.confirmation input { width: 100%; border: 1px solid #dcdce1; border-radius: 7px; padding: 11px 12px; font-size: 12px; background: #fff; }
.confirmation input:focus-visible { outline: 2px solid #74747c; outline-offset: 3px; }
.confirmation small { margin-top: 8px; color: #9696a0; font-size: 10px; line-height: 1.6; }
.actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 20px; }
.error-message { margin-top: 18px; padding: 12px; border: 1px solid #eed8d8; border-radius: 8px; background: #fff7f7; color: #8e4141; font-size: 12px; line-height: 1.7; }
.success-message { display: flex; align-items: flex-start; gap: 9px; margin-top: 18px; padding: 12px; border: 1px solid #dce9df; border-radius: 8px; font-size: 12px; line-height: 1.7; color: #326743; background: #f7faf7; }
.success-message svg { margin-top: 2px; }
</style>
