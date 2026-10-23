<script setup>
const { t, formatTime, formatDate, formatMoney } = useAppI18n()
defineProps({ plan: Object, pending: Boolean, blocked: Number })
defineEmits(['apply', 'close', 'claim'])
</script>

<template>
  <section class="replan-preview">
    <header><h2>{{ t('Replacement plan') }}</h2><button :aria-label="t('Close replacement plan')" @click="$emit('close')"><UiIcon name="X" :size="15" /></button></header>
    <div class="outage-note"><UiIcon name="Truck" :size="19" /><span>{{ t('{name} is unavailable.', { name: plan.provider.name }) }}<small>{{ t('{count} assignments affected', { count: plan.changes.length }) }}</small></span></div>
    <h3>{{ t('Affected missions') }}</h3>
    <div v-for="change in plan.changes" :key="change.mission.id" class="change-card">
      <button class="affected-name" @click="$emit('claim', change.mission.claimId)"><span>{{ t(change.mission.claim.name, change.mission.claim.nameParameters) }}<small>{{ change.mission.claimId }} · {{ change.mission.claim.municipality }}</small></span><UiIcon name="ChevronRight" :size="13" /></button>
      <div class="comparison"><div><span>{{ t('Before') }}</span><strong>{{ plan.provider.name }}</strong><small>{{ formatDate(change.mission.start, { month: 'short', year: undefined }) }} · {{ formatTime(change.mission.start) }}</small></div><UiIcon name="ArrowRight" :size="14" /><div :class="{ replacement: change.replacement }"><span>{{ t('Proposed') }}</span><strong>{{ change.replacement?.name || t('No eligible replacement') }}</strong><small v-if="change.start">{{ formatTime(change.start) }} · {{ t('Access reviewed') }}</small></div></div>
    </div>
    <h3>{{ t('Planning impact') }}</h3><div class="impact-row"><UiIcon name="Clock3" :size="16" /><span>{{ t('Maximum start delay') }}</span><strong>+{{ t('{count}h', { count: plan.maxDelay }) }}</strong></div><div class="impact-row"><UiIcon name="FileText" :size="16" /><span>{{ t('Estimated mission cost change') }}</span><strong>+{{ formatMoney(plan.costDelta) }}</strong></div><div class="impact-row"><UiIcon name="ShieldCheck" :size="16" /><span>{{ t('{count} blocked missions stay blocked', { count: blocked }) }}</span></div>
    <p class="assumptions">{{ t('Proposed timing and costs remain estimates. Review provider availability before applying; coverage and payment require separate approval.') }}</p>
    <UiButton class="dark" :disabled="pending || !plan.canApply" @click="$emit('apply')">{{ t(pending ? 'Updating…' : 'Apply proposed plan') }}</UiButton><UiButton :disabled="pending" @click="$emit('close')">{{ t('Cancel') }}</UiButton>
  </section>
</template>

<style scoped>
.replan-preview { display: flex; flex: 1; flex-direction: column; gap: 10px; padding: 19px 20px 22px; overflow-y: auto; }.replan-preview > header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 9px; }.replan-preview h2 { font-size: 15px; letter-spacing: -0.3px; }.replan-preview > header button { padding: 0; color: #9696a1; }.replan-preview h3 { font-size: 11px; margin-top: 13px; margin-bottom: 3px; }
.outage-note { display: flex; align-items: center; gap: 11px; padding: 12px; border-radius: 7px; background: #fff0f1; color: #c75b65; font-size: 11px; line-height: 17px; }.outage-note small { display: block; color: #d48c93; font-size: 10px; }
.change-card { border: 1px solid #ebebef; border-radius: 9px; padding: 12px; }.affected-name { display: flex; align-items: center; gap: 12px; justify-content: space-between; width: 100%; padding: 0; font-size: 11px; font-weight: 500; text-align: left; }.affected-name small { display: block; margin-top: 5px; color: #a3a3ad; font-size: 9px; font-weight: 400; }
.comparison { display: flex; gap: 8px; align-items: center; margin-top: 14px; }.comparison > div { display: flex; flex-direction: column; gap: 6px; flex: 1; background: #f5f5f7; padding: 9px; border-radius: 6px; min-width: 0; }.comparison span { font-size: 8px; color: #a0a0ab; }.comparison strong { font-size: 9px; line-height: 14px; font-weight: 500; }.comparison small { font-size: 8px; line-height: 13px; color: #9b9ba6; }.comparison > div.replacement { background: #eef9f2; }.comparison > svg { color: #a4a4ad; }
.impact-row { display: flex; align-items: center; gap: 9px; padding: 10px 0; border-bottom: 1px solid #eeeef0; font-size: 10px; color: #8c8c98; }.impact-row > span { flex: 1; }.impact-row strong { color: #42424c; font-size: 11px; font-weight: 500; }
.assumptions { padding: 12px; margin: 8px 0; border-radius: 7px; background: #f5f6fa; color: #9297aa; font-size: 9px; line-height: 16px; }

@media (max-width: 1000px) {
  .replan-preview { flex: none; overflow: visible; }
  .affected-name { min-height: 44px; }
  .replan-preview :deep(.button) { min-height: 44px; }
}

</style>
