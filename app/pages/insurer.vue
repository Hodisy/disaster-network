<script setup>
const { t, formatDate, formatMoney } = useAppI18n()
const { documentRows, state: clientClaimState } = useDemo()
const { refreshLiveClaim, claimError } = useHotlineClaim()
onMounted(() => { void refreshLiveClaim({ quiet: true }) })
const { state, view, panel, detailOpen, selectedId, selected, query, page, pageCount, pageClaims, filtered, queue, providers, missions, selectedMissions, stats, options, replan,
  claimGeojson, providerGeojson, assignmentGeojson, proposedGeojson, selectClaim, updateFilter, clearFilters, simulateOutage, applyPlan, cancelOutage, permitAccess, placeAccommodation, scheduleSelected, planAssistance } = useInsurerWorkspace()
const advanced = ref(false)
const popupOpen = ref(true)
const reportOpen = ref(false)
const replanOpen = ref(false)
const selectedProvider = ref('')
const selectedAudit = computed(() => state.value.audit.filter(entry => entry.claimId === selectedId.value))
function openClaim(id) { selectClaim(id, true); panel.value = 'claims'; replanOpen.value = false; popupOpen.value = false }
function selectOnMap(id) { selectClaim(id); popupOpen.value = true }
function showPanel(value) { panel.value = value; detailOpen.value = false; replanOpen.value = false }
function selectProvider(id) { selectedProvider.value = id; view.value = 'mobilization' }
async function outage() { await simulateOutage(); replanOpen.value = !!replan.value }
async function approveReplan() { await applyPlan(); if (state.value.planApplied) replanOpen.value = false }
async function resetResources() { await cancelOutage(); replanOpen.value = false; selectedProvider.value = '' }
watch(() => state.value.epoch, () => { replanOpen.value = false; popupOpen.value = true; selectedProvider.value = '' })
const reportText = computed(() => [
  t('Disaster Network — Situation report'),
  `${t('Scenario snapshot')}: ${formatDate('2026-10-22')}`,
  `${t('Flood event')}: ${formatDate('2026-10-20')}`,
  t('All figures below are synthetic, not real-event totals.'), '',
  t('{count} households', { count: stats.value.households }), t('{count} people', { count: stats.value.people }),
  t('{count} opened reports', { count: stats.value.opened }),
  ...stats.value.groups.map(group => `${group.id} — ${t(group.label)}: ${group.count}`), '',
  t('{count} households need accommodation', { count: stats.value.accommodation }),
  t('{count} blocked missions stay blocked', { count: stats.value.blocked }),
  `${t('Known gross estimates')}: ${formatMoney(stats.value.gross)}`,
  t('{count} records have incomplete estimates.', { count: stats.value.incomplete }),
  `${t('Approved commitment')}: ${formatMoney(stats.value.approved)}`, `${t('Paid')}: ${formatMoney(stats.value.paid)}`, '',
  t('Priority is based on needs, not claim value.'),
  t('Confirm accommodation needs; review access before scheduling physical work; use qualified available providers.'),
  t('Demo coverage is contractual and fictional. Statutory CatNat recognition is not verified.'),
  t('Gross estimates are not reserves, approved amounts, or measured savings.'), '',
  ...state.value.audit.map(entry => t(entry.label))
].join('\n'))
function downloadReport() {
  const url = URL.createObjectURL(new Blob([reportText.value], { type: 'text/plain;charset=utf-8' }))
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = 'disaster-network-situation-2026-10-22.txt'
  anchor.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
useHead(() => ({ title: t('Insurer workspace — Disaster Network') }))
</script>

<template>
  <div class="insurer-page">
    <p v-if="clientClaimState.claim?.source === 'postgresql'" class="live-report-source">{{ t('Jules: saved Live report from PostgreSQL. The other 79 households, financial estimates and resource actions remain frontend simulations.') }}</p>
    <p v-if="clientClaimState.claim?.source === 'postgresql' && claimError" class="live-report-source" role="alert">{{ t(hotlineErrorPhrase(claimError)) }}</p>
    <header class="operations-heading">
      <div><h1>{{ t('Berre Lagoon flood response') }}</h1><p>{{ t('Event: {date}', { date: formatDate('2026-10-20', { month: 'short' }) }) }}<span>·</span>{{ t('Scenario snapshot: {date}', { date: formatDate('2026-10-22', { month: 'short' }) }) }}</p></div>
      <div class="heading-actions"><span class="demo-data">{{ t('Synthetic data / Illustrative locations') }}</span><UiButton v-if="view === 'mobilization' && !state.outage" class="small" :disabled="state.pending" @click="outage"><UiIcon name="TriangleAlert" :size="14" />{{ t('Simulate provider outage') }}</UiButton><UiButton v-else-if="state.outage" class="small" :disabled="state.pending" @click="resetResources">{{ t('Reset resource simulation') }}</UiButton><UiButton class="dark" @click="reportOpen = true"><UiIcon name="FileText" :size="15" />{{ t('Situation report') }}</UiButton></div>
    </header>
    <div class="portfolio-summary"><span><strong>{{ stats.households }}</strong>{{ t('Portfolio households') }}</span><span><strong>{{ stats.people }}</strong>{{ t('Synthetic people') }}</span><span><strong>{{ stats.opened }}</strong>{{ t('Opened reports') }}</span><span class="summary-amount"><strong>{{ formatMoney(stats.gross) }}</strong>{{ t('Known gross estimates') }}<small>{{ t('Partial · Not approved') }}</small></span></div>
    <InsurerToolbar :view="view" :query="query" :options="options" :advanced="advanced" @view="view = $event" @filter="updateFilter" @advanced="advanced = !advanced" @clear="clearFilters" />
    <div v-if="state.error" class="operation-error" role="alert">{{ t(state.error) }}</div>
    <div v-if="state.outage" class="operation-notice" :class="{ applied: state.planApplied }" role="status"><UiIcon :name="state.planApplied ? 'CircleCheck' : 'TriangleAlert'" :size="15" /><span>{{ t(state.planApplied ? 'The replacement plan has been applied in the demo. No real dispatch or payment occurred.' : 'A drying provider is unavailable. Review the affected missions before applying a replacement.') }}</span><button v-if="!state.planApplied" @click="replanOpen = true; view = 'mobilization'">{{ t('Review replanning') }}<UiIcon name="ChevronRight" :size="13" /></button></div>
    <div class="operations-body" :class="{ mobilization: view === 'mobilization' }">
      <div class="map-column">
        <InsurerMap :claims="claimGeojson" :providers="providerGeojson" :assignments="assignmentGeojson" :proposed="proposedGeojson" :selected="selected" :show-popup="popupOpen && !detailOpen" :stats="stats" @select="selectOnMap" @open="openClaim" @provider="selectProvider" @dismiss-popup="popupOpen = false" />
        <InsurerResources v-if="view === 'mobilization'" :providers="providers" :missions="missions" :selected-provider="selectedProvider" @provider="selectProvider" @claim="openClaim" />
      </div>
      <aside class="operations-panel" :aria-label="t('Claim workspace')">
        <div class="panel-tabs" role="tablist" :aria-label="t('Right panel view')"><button id="priority-tab" role="tab" :aria-selected="panel === 'priority'" aria-controls="claim-panel" :class="{ active: panel === 'priority' }" @click="showPanel('priority')">{{ t('Priority queue') }}</button><button id="claims-tab" role="tab" :aria-selected="panel === 'claims'" aria-controls="claim-panel" :class="{ active: panel === 'claims' }" @click="showPanel('claims')">{{ t('All claims') }}</button></div>
        <div id="claim-panel" class="panel-content" role="tabpanel" :aria-labelledby="panel === 'priority' ? 'priority-tab' : 'claims-tab'">
          <InsurerReplanPreview v-if="replanOpen && replan" :plan="replan" :pending="state.pending" :blocked="stats.blocked" @apply="approveReplan" @close="replanOpen = false" @claim="openClaim" />
          <InsurerClaimDetail v-else-if="detailOpen && selected" :claim="selected" :missions="selectedMissions" :audit="selectedAudit" :pending="state.pending" :client-documents="documentRows" @back="showPanel('claims')" @plan="planAssistance" @permit-access="permitAccess" @accommodate="placeAccommodation" @schedule="scheduleSelected" @client="navigateTo('/client')" />
          <InsurerClaimList v-else :mode="panel" :claims="panel === 'priority' ? queue : pageClaims" :total="panel === 'priority' ? queue.length : filtered.length" :selected-id="selectedId" :page="page" :page-count="pageCount" @select="openClaim" @page="page = $event" @clear="clearFilters" />
        </div>
      </aside>
    </div>
    <UiDialog v-if="reportOpen" :title="t('Disaster Network — Situation report')" wide @close="reportOpen = false"><div class="report-body"><span class="report-label">{{ t('Frontend-generated demo report') }}</span><p>{{ t('This snapshot uses the entire portfolio, independently of current map filters.') }}</p><pre>{{ reportText }}</pre><UiButton class="dark" @click="downloadReport"><UiIcon name="Download" :size="15" />{{ t('Download report') }}</UiButton></div></UiDialog>
  </div>
</template>

<style scoped>
.live-report-source { padding: 10px 24px; color: #777781; font-size: 11px; line-height: 1.6; border-bottom: 1px solid #e8e8ed; }
.insurer-page { display: flex; flex-direction: column; min-width: 1050px; background: #fff; }
.operations-heading { display: flex; align-items: center; justify-content: space-between; gap: 20px; padding: 27px 24px 23px; }.operations-heading h1 { font-size: 27px; letter-spacing: -0.9px; line-height: 35px; }.operations-heading p { display: flex; align-items: center; gap: 10px; margin-top: 7px; color: #9696a2; font-size: 11px; }.operations-heading p > span { color: #bcbcc4; }
.heading-actions { display: flex; align-items: center; gap: 10px; }.demo-data { border: 1px solid var(--border); border-radius: 20px; color: #9b9ba6; font-size: 9px; padding: 9px 12px; white-space: nowrap; }
.portfolio-summary { display: flex; align-items: center; gap: 26px; padding: 0 24px 18px; border-bottom: 1px solid var(--border); }.portfolio-summary > span { display: flex; align-items: baseline; gap: 8px; font-size: 10px; color: #9999a4; }.portfolio-summary strong { font-size: 17px; font-weight: 550; letter-spacing: -0.5px; color: #44444f; }.portfolio-summary small { font-size: 8px; color: #b0b0b9; }.summary-amount { margin-left: auto; }
.operations-body { display: flex; align-items: stretch; height: 740px; }.operations-body.mobilization { height: 1060px; }.map-column { display: flex; flex-direction: column; flex: 1; min-width: 0; }.operations-panel { display: flex; flex-direction: column; width: 360px; flex-shrink: 0; border-left: 1px solid var(--border); background: #fff; min-height: 0; }
.panel-tabs { display: flex; align-items: stretch; padding: 0 18px; border-bottom: 1px solid #eeeeef; flex-shrink: 0; }.panel-tabs button { flex: 1; padding: 16px 5px; border-bottom: 2px solid transparent; font-size: 11px; color: #9a9aa5; }.panel-tabs button.active { border-color: #29292f; color: var(--ink); font-weight: 550; }.panel-content { display: flex; flex-direction: column; flex: 1; min-height: 0; }
.operation-notice { display: flex; align-items: center; gap: 10px; padding: 11px 24px; border-bottom: 1px solid #f4dbdf; background: #fff5f5; color: #b8787f; font-size: 10px; }.operation-notice > span { flex: 1; }.operation-notice button { display: flex; align-items: center; gap: 6px; font-size: 10px; }.operation-notice.applied { color: #5f9b76; background: #f1fbf5; border-color: #d4ecdd; }
.operation-error { padding: 12px 24px; color: #bb6871; background: #fff4f4; font-size: 11px; border-bottom: 1px solid #f1dddd; }
.report-body { display: flex; flex-direction: column; gap: 14px; padding: 5px; }.report-label { width: fit-content; color: #9c9ca7; background: #f3f3f5; border-radius: 5px; padding: 6px 9px; font-size: 10px; }.report-body > p { color: #9c9ca7; font-size: 11px; line-height: 19px; }.report-body pre { white-space: pre-wrap; overflow-wrap: anywhere; color: #747481; font: 12px/23px -apple-system, BlinkMacSystemFont, sans-serif; background: #f7f7f9; padding: 18px; margin: 0; border-radius: 8px; }.report-body :deep(.button) { align-self: flex-start; }
</style>
