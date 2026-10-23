<script setup>
const { t, formatDate } = useAppI18n()
defineProps({ draft: Object, busy: Boolean })
defineEmits(['confirm'])
const needs = { accommodation: 'Temporary accommodation', essential_items: 'Essential items', electrical_safety: 'Electrical safety', cleaning_drying: 'Cleaning and drying', moving_storage: 'Moving and storage', contents: 'Damaged contents', building: 'Insured building repairs' }
const documents = { photos: 'Damage photos', address: 'Proof of address', inventory: 'Belongings inventory', invoice: 'Item invoices' }
</script>
<template>
  <section class="report-review" :aria-label="t('Review your report')">
    <h2>{{ t('Review your report') }}</h2><p>{{ t('Confirm only if this report matches what you told the assistant. Ask for corrections before saving.') }}</p>
    <blockquote>{{ draft.description }}</blockquote>
    <dl>
      <div><dt>{{ t('Household') }}</dt><dd>{{ t(draft.householdSize === 1 ? '{count} person' : '{count} people', { count: draft.householdSize }) }}</dd></div>
      <div><dt>{{ t('Household safety') }}</dt><dd>{{ t(({ yes: 'Safe', no: 'Not safe', unknown: 'Unknown' })[draft.safe]) }}</dd></div>
      <div><dt>{{ t('Home') }}</dt><dd>{{ t(({ habitable: 'Home habitable', uninhabitable: 'Home uninhabitable', unknown: 'Unknown' })[draft.habitability]) }}</dd></div>
      <div><dt>{{ t('Site access') }}</dt><dd>{{ t(draft.access === 'restricted' ? 'Restricted' : 'Unknown') }}</dd></div>
      <div><dt>{{ t('Current accommodation') }}</dt><dd>{{ draft.currentAccommodation || t('Not provided') }}</dd></div>
      <div><dt>{{ t('Accommodation needed') }}</dt><dd>{{ t(draft.accommodationNeeded ? 'Yes' : 'No') }}</dd></div>
      <div><dt>{{ t('Requested start') }}</dt><dd>{{ draft.accommodationStart ? formatDate(draft.accommodationStart) : t('Not provided') }}</dd></div>
      <div><dt>{{ t('Requested nights') }}</dt><dd>{{ draft.accommodationNights || t('Not provided') }}</dd></div>
      <div><dt>{{ t('Accommodation constraints') }}</dt><dd>{{ draft.accommodationConstraints || t('Not provided') }}</dd></div>
      <div><dt>{{ t('Identified needs') }}</dt><dd>{{ draft.needs.map(code => t(needs[code])).join(', ') || t('Not provided') }}</dd></div>
      <div><dt>{{ t('Requested documents') }}</dt><dd>{{ draft.documents.map(code => t(documents[code])).join(', ') || t('Not provided') }}</dd></div>
    </dl>
    <p v-for="warning in draft.warnings || []" :key="warning" role="status">{{ t(warning) }}</p>
    <p>{{ t('Saving records your report and requests. It does not approve coverage, book accommodation or make a payment.') }}</p>
    <UiButton class="dark" :disabled="busy" @click="$emit('confirm')">{{ t(busy ? 'Saving report…' : 'Confirm and save report') }}</UiButton>
  </section>
</template>
<style scoped>
.report-review { margin-top: 24px; border: 1px solid #e3e3e7; border-radius: 14px; padding: 22px; }
h2 { font-size: 17px; margin-bottom: 8px; }
p { font-size: 12px; line-height: 1.7; color: #75757f; margin-bottom: 14px; }
blockquote { border-left: 2px solid #d5d5dc; padding: 6px 14px; font-size: 14px; line-height: 1.7; white-space: pre-wrap; overflow-wrap: anywhere; }
dl { margin: 16px 0; font-size: 12px; line-height: 1.7; }
dl > div { display: flex; gap: 14px; padding: 5px 0; }
dt { width: 165px; flex-shrink: 0; color: #85858e; }
dd { flex: 1; margin: 0; overflow-wrap: anywhere; }
@media (max-width: 760px) {
  .report-review { padding: 16px; }
  blockquote { margin: 16px 0; padding: 4px 10px; }
  dl > div { flex-direction: column; gap: 2px; padding: 7px 0; }
  dt { width: auto; }
  .report-review :deep(.button) { width: 100%; min-height: 44px; white-space: normal; }
}
</style>
