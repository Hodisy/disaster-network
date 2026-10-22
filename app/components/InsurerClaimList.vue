<script setup>
const { t, formatMoney } = useAppI18n()
defineProps({ claims: Array, total: Number, mode: String, selectedId: String, page: Number, pageCount: Number })
defineEmits(['select', 'page', 'clear'])
</script>

<template>
  <section class="claim-list">
    <div class="list-heading"><h2>{{ t(mode === 'priority' ? 'Needs attention' : 'All claims') }}<span>{{ total }}</span></h2><p>{{ t(mode === 'priority' ? 'Unmet urgent needs, highest waiting time first.' : 'Synthetic households, one principal claim each.') }}</p></div>
    <div v-if="!claims.length" class="empty-list"><UiIcon name="Search" :size="24" /><p>{{ t('No claims match these filters.') }}</p><button @click="$emit('clear')">{{ t('Clear filters') }}</button></div>
    <div v-else class="list-rows">
      <button v-for="claim in claims" :key="claim.id" class="claim-row" :class="{ selected: selectedId === claim.id, urgent: claim.priority === 'P1' }" @click="$emit('select', claim.id)">
        <i class="priority-dot" :class="claim.priority" />
        <span class="claim-row-content"><strong>{{ t(claim.name, claim.nameParameters) }}</strong><small>{{ claim.id }} · {{ claim.municipality }}</small><span v-if="mode === 'priority'" class="need-line"><UiIcon :name="claim.unmetAccommodation ? 'House' : 'Clock3'" :size="12" />{{ t(claim.unmetAccommodation ? 'Accommodation needed' : 'Urgent review needed') }}</span><small v-else class="stage-label">{{ t(claim.opened ? claim.stageLabel : 'Awaiting report') }}</small></span>
        <span v-if="mode === 'priority'" class="waiting"><small>{{ t('Waiting') }}</small><strong>{{ t('{count}h', { count: claim.waitingHours }) }}</strong></span><span v-else class="estimate">{{ claim.estimateIncomplete ? t('Partial') : formatMoney(claim.grossEstimate) }}</span>
        <UiIcon name="ChevronRight" :size="14" />
      </button>
    </div>
    <footer v-if="mode === 'claims' && total" class="pagination"><span>{{ t('{from}–{to} of {total}', { from: (page - 1) * 8 + 1, to: Math.min(page * 8, total), total }) }}</span><div><button :disabled="page <= 1" :aria-label="t('Previous page')" @click="$emit('page', page - 1)"><UiIcon name="ChevronLeft" :size="15" /></button><button :disabled="page >= pageCount" :aria-label="t('Next page')" @click="$emit('page', page + 1)"><UiIcon name="ChevronRight" :size="15" /></button></div></footer>
    <p v-if="mode === 'priority' && claims.length" class="queue-note">{{ t('Priority is based on needs, not claim value.') }}</p>
  </section>
</template>

<style scoped>
.claim-list { display: flex; flex: 1; flex-direction: column; min-height: 0; }
.list-heading { padding: 19px 20px 12px; }
.list-heading h2 { display: flex; align-items: center; gap: 8px; font-size: 14px; letter-spacing: -0.3px; }
.list-heading h2 > span { padding: 3px 7px; border-radius: 6px; background: #f0f0f3; font-size: 10px; }
.list-heading p { margin-top: 7px; font-size: 10px; line-height: 16px; color: #9a9aa5; }
.list-rows { flex: 1; overflow-y: auto; padding: 0 15px; }
.claim-row { display: flex; align-items: center; gap: 10px; width: 100%; padding: 15px 5px; text-align: left; border-bottom: 1px solid #ededf0; }
.claim-row.selected { background: #f1f1f4; border-radius: 8px; padding-inline: 9px; }
.claim-row.urgent.selected { background: #fdf2f3; }
.claim-row:hover { background: #f6f6f8; border-radius: 8px; }
.priority-dot { flex-shrink: 0; align-self: flex-start; margin-top: 4px; width: 10px; height: 10px; border-radius: 50%; }
.P1 { background: #ed4548; }.P2 { background: #efaa24; }.P3 { background: #9397a4; }
.claim-row-content { display: flex; flex-direction: column; flex: 1; gap: 5px; min-width: 0; }
.claim-row-content strong { font-size: 12px; font-weight: 550; }
.claim-row-content small { font-size: 9px; color: #91919c; line-height: 14px; }
.need-line { display: flex; align-items: center; gap: 5px; font-size: 9px; color: #81818c; line-height: 14px; }
.waiting { display: flex; flex-direction: column; gap: 4px; text-align: right; color: #db4149; }
.waiting small { color: #9999a3; font-size: 8px; }
.waiting strong { font-size: 13px; font-weight: 500; }
.estimate { font-size: 10px; color: #7c7c88; white-space: nowrap; }
.claim-row > svg { color: #878792; }
.pagination { display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 15px 20px; border-top: 1px solid var(--border); font-size: 10px; color: #8e8e99; }
.pagination > div { display: flex; gap: 7px; }
.pagination button { display: flex; align-items: center; justify-content: center; width: 28px; height: 28px; border: 1px solid var(--border); border-radius: 6px; }
.pagination button:disabled { opacity: 0.3; }
.queue-note { padding: 15px 20px; color: #a0a0aa; font-size: 9px; border-top: 1px solid var(--border); }
.empty-list { display: flex; align-items: center; flex-direction: column; justify-content: center; flex: 1; gap: 15px; color: #a0a0aa; padding: 30px; text-align: center; }
.empty-list p { font-size: 12px; line-height: 20px; }.empty-list button { font-size: 11px; text-decoration: underline; }
</style>
