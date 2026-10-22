<script setup>
defineProps({ summary: { type: Object, required: true } })
defineEmits(['priority', 'resources'])
const { t, formatMoney } = useAppI18n()
</script>

<template>
  <section class="operations-summary" :aria-label="t('Entire portfolio overview')">
    <article class="summary-card urgent-card">
      <div class="card-heading"><span>{{ t('Urgent needs') }}</span><button @click="$emit('priority')">{{ t('Review priorities') }}<UiIcon name="ArrowUpRight" :size="12" /></button></div>
      <div class="headline"><strong>{{ summary.urgent }}</strong><span>{{ t('P1 households') }}</span><small>{{ t('{count} households · {people} people', { count: summary.households, people: summary.people }) }}</small></div>
      <p><span>{{ t('{count} need accommodation', { count: summary.accommodation }) }}</span><span>{{ t('{count} blocked missions', { count: summary.blocked }) }}</span></p>
    </article>
    <article class="summary-card">
      <div class="card-heading"><span>{{ t('Known gross estimates') }}</span><small>{{ t('Partial · Not approved') }}</small></div>
      <div class="headline"><strong class="money">{{ formatMoney(summary.gross) }}</strong><small>{{ t('{count} incomplete estimates', { count: summary.incomplete }) }}</small></div>
      <p><span>{{ t('Approved commitment') }} <b>{{ formatMoney(summary.approved) }}</b></span><span>{{ t('Paid') }} <b>{{ formatMoney(summary.paid) }}</b></span></p>
    </article>
    <article class="summary-card">
      <div class="card-heading"><span>{{ t('Response resources') }}</span><button @click="$emit('resources')">{{ t('Review deployment') }}<UiIcon name="ArrowUpRight" :size="12" /></button></div>
      <div class="headline"><strong>{{ summary.availableProviders }}</strong><span>{{ t('providers with free capacity') }}</span><small>{{ t('{count} mobilized providers', { count: summary.mobilizedProviders }) }}</small></div>
      <p><span>{{ t('{count} scheduled missions', { count: summary.scheduled }) }}</span><span>{{ t('{count} draft missions', { count: summary.proposed }) }}</span><span v-if="summary.disrupted" class="disrupted">{{ t('{count} disrupted missions', { count: summary.disrupted }) }}</span></p>
    </article>
  </section>
</template>

<style scoped>
.operations-summary { display: flex; flex-shrink: 0; border-bottom: 1px solid var(--border); }
.summary-card { flex: 1; min-width: 0; padding: 9px 20px; }
.summary-card + .summary-card { border-left: 1px solid var(--border); }
.card-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; font-size: 11px; color: #60606c; }
.card-heading > span { font-weight: 550; }
.card-heading button { display: inline-flex; align-items: center; gap: 4px; padding: 0; font-size: 10px; color: #777782; }
.card-heading button:hover { color: #171719; }
.card-heading small { font-size: 9px; color: #92929d; }
.headline { display: flex; align-items: baseline; flex-wrap: wrap; gap: 7px; margin-top: 5px; }
.headline strong { font-size: 25px; line-height: 28px; font-weight: 550; letter-spacing: -.8px; }
.urgent-card .headline > strong { color: #d3444b; }
.headline .money { font-size: 25px; }
.headline > span { font-size: 11px; color: #6f6f7c; }
.headline > small { font-size: 9px; line-height: 14px; color: #92929d; }
.summary-card p { display: flex; flex-wrap: wrap; gap: 5px 12px; margin-top: 4px; font-size: 9px; line-height: 15px; color: #858590; }
.summary-card b { font-weight: 500; color: #51515d; }
.summary-card .disrupted { color: #c55b62; }
</style>
