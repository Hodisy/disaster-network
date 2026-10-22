<script setup>
const { t, formatDate, formatMoney } = useAppI18n()
const { documentRows } = useDemo()
const { data: agentData, busy: agentBusy, error: agentError, restore: restoreAgent, refresh: refreshAgent, loadProposal, analyze: analyzeAgent, review: reviewAgent, downloadPolicy } = useInsurerAgent()
// Every visit starts in the local scenario. Database loading and AI analysis are explicit actions.
agentData.value.live = false
const { state, view, panel, detailOpen, selectedId, selected, query, page, pageCount, pageClaims, claims, filtered, queue, providers, missions, selectedMissions, stats, options, replan,
  claimGeojson, providerGeojson, assignmentGeojson, proposedGeojson, selectClaim, updateFilter, clearFilters, simulateOutage, applyPlan, cancelOutage, permitAccess, placeAccommodation, scheduleSelected, planAssistance } = useInsurerWorkspace()
const { summary, services } = useInsurerOperationsSummary({ claims, providers, missions, stats })
const connectingWorkspace = ref(false)
const workspaceNotice = ref('')
const advanced = ref(false)
const popupOpen = ref(false)
const reportOpen = ref(false)
const replanOpen = ref(false)
const selectedProvider = ref('')
const agentOpen = ref(false), agentReference = ref(''), agentConsent = ref(false), humanReviewed = ref(false), humanNote = ref('')
const agentProposal = computed(() => agentData.value.proposals[agentReference.value] || null)
watch(() => agentProposal.value?.id, () => { humanReviewed.value = false; humanNote.value = '' })
async function openAgent() { if (!selectedId.value) return; agentReference.value = selectedId.value; agentOpen.value = true; agentConsent.value = false; await loadProposal(agentReference.value) }
async function applyAgentDecision(decision) { await reviewAgent(agentReference.value, agentProposal.value.id, decision, humanNote.value, humanReviewed.value); view.value = 'mobilization' }
const selectedAudit = computed(() => state.value.audit.filter(entry => entry.claimId === selectedId.value))
function openClaim(id) { selectClaim(id, true); panel.value = 'claims'; replanOpen.value = false; popupOpen.value = false }
function selectOnMap(id) { selectClaim(id); popupOpen.value = true }
function showPanel(value) {
  panel.value = value; detailOpen.value = false; replanOpen.value = false; popupOpen.value = false
  if (value === 'resources') view.value = 'mobilization'
  if (value === 'plan' || value === 'resources') selectClaim(null)
}
function showResources() { view.value = 'mobilization'; showPanel('resources') }
function changeView(value) { view.value = value; showPanel(value === 'mobilization' ? 'resources' : 'plan') }
function reviewService(skill) { updateFilter('need', skill); showPanel('claims') }
async function connectAgentWorkspace() {
  if (agentBusy.value || connectingWorkspace.value) return
  connectingWorkspace.value = true; workspaceNotice.value = ''
  try {
    await restoreAgent()
    if (!agentData.value.authorized) workspaceNotice.value = 'Unlock insurer access in Settings to continue.'
  } finally { connectingWorkspace.value = false }
}
watch(() => agentData.value.live, () => {
  selectClaim(null); detailOpen.value = false; popupOpen.value = false
  panel.value = 'plan'; replanOpen.value = false; selectedProvider.value = ''; clearFilters()
})
function selectProvider(id) { selectedProvider.value = id; showResources() }
async function outage() { await simulateOutage(); replanOpen.value = !!replan.value }
async function approveReplan() { await applyPlan(); if (state.value.planApplied) replanOpen.value = false }
async function resetResources() { await cancelOutage(); replanOpen.value = false; selectedProvider.value = '' }
watch(() => state.value.epoch, () => { replanOpen.value = false; popupOpen.value = false; panel.value = 'plan'; selectedProvider.value = '' })
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
  t('The contractual flood cover is reviewed separately from statutory CatNat recognition, which remains unverified.'),
  t('Gross estimates are not reserves, approved amounts, or measured savings.'), '',
  ...state.value.audit.map(entry => t(entry.label)),
  ...(agentData.value.live ? [t('AI classification and human decisions'),
    ...agentData.value.claims.filter(c => c.assessment.status === 'human_reviewed_proposal').map(c => `${c.reference} · ${c.assessment.group} / ${c.priority} · ${c.assessment.summary} · ${t('Evidenced reimbursement proposal')}: ${c.assessment.reimbursement_proposal_cents == null ? t('Unknown') : formatMoney(c.assessment.reimbursement_proposal_cents / 100)} · ${t('No payment executed')}`)] : [])
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
    <header class="operations-heading">
      <div><h1>{{ t('Berre Lagoon flood response') }}</h1><p>{{ t('Event: {date}', { date: formatDate('2026-10-20', { month: 'short' }) }) }}<span>·</span>{{ t('Scenario snapshot: {date}', { date: formatDate('2026-10-22', { month: 'short' }) }) }}</p></div>
      <div class="heading-actions"><span class="demo-data">{{ t('Portfolio snapshot') }}</span><UiButton v-if="!agentData.live && view === 'mobilization' && !state.outage" class="small" :disabled="state.pending" @click="outage"><UiIcon name="TriangleAlert" :size="14" />{{ t('Preview provider outage') }}</UiButton><UiButton v-else-if="!agentData.live && state.outage" class="small" :disabled="state.pending" @click="resetResources">{{ t('Restore resource plan') }}</UiButton><UiButton class="dark" @click="reportOpen = true"><UiIcon name="FileText" :size="15" />{{ t('Situation report') }}</UiButton></div>
    </header>
    <InsurerOperationsSummary :summary="summary" @priority="showPanel('priority')" @resources="showResources" />
    <div class="agent-controls">
      <span>{{ t(agentData.live ? 'AI-assisted review · Proposals require approval' : 'Operations overview · Needs, costs and deployment') }}</span>
      <button v-if="!agentData.live" :disabled="!!agentBusy || connectingWorkspace" @click="connectAgentWorkspace">{{ t(connectingWorkspace ? 'Loading dossiers…' : 'Use AI-assisted review') }}</button>
      <template v-else><button @click="agentData.live = false">{{ t('Return to operations overview') }}</button><button :disabled="!!agentBusy" @click="refreshAgent">{{ t('Refresh') }}</button></template>
      <NuxtLink v-if="!agentData.authorized" to="/settings">{{ t('Unlock insurer agent in Settings') }}</NuxtLink>
      <UiButton class="dark small" :disabled="!agentData.live || !selectedId || !!agentBusy" :title="t('Select a dossier to request an AI proposal')" @click="openAgent">{{ t('Review with AI') }}</UiButton>
    </div>
    <p v-if="workspaceNotice && !agentData.live" class="workspace-notice" role="status">{{ t(workspaceNotice) }}</p>
    <p v-if="agentError && !agentOpen" class="operation-error" role="alert">{{ t(agentError) }}</p>
    <InsurerToolbar :view="view" :query="query" :options="options" :advanced="advanced" @view="changeView" @filter="updateFilter" @advanced="advanced = !advanced" @clear="clearFilters" />
    <div v-if="state.error" class="operation-error" role="alert">{{ t(state.error) }}</div>
    <div v-if="!agentData.live && state.outage" class="operation-notice" :class="{ applied: state.planApplied }" role="status"><UiIcon :name="state.planApplied ? 'CircleCheck' : 'TriangleAlert'" :size="15" /><span>{{ t(state.planApplied ? 'Replacement plan updated. Review provider availability before dispatch.' : 'A drying provider is unavailable. Review the affected missions before applying a replacement.') }}</span><button v-if="!state.planApplied" @click="replanOpen = true; view = 'mobilization'">{{ t('Review replanning') }}<UiIcon name="ChevronRight" :size="13" /></button></div>
    <div class="operations-body" :class="{ mobilization: view === 'mobilization' }">
      <div class="map-column">
        <InsurerMap fit :claims="claimGeojson" :providers="providerGeojson" :assignments="assignmentGeojson" :proposed="proposedGeojson" :selected="selected" :show-popup="popupOpen && !detailOpen" :stats="stats" @select="selectOnMap" @open="openClaim" @provider="selectProvider" @dismiss-popup="popupOpen = false" />
      </div>
      <aside class="operations-panel" :aria-label="t('Claim workspace')">
        <div class="panel-tabs" role="tablist" :aria-label="t('Right panel view')">
          <button v-for="tab in [{ id: 'plan', label: 'Response plan' }, { id: 'resources', label: 'Providers' }, { id: 'priority', label: 'Priority queue' }, { id: 'claims', label: 'All claims' }]" :id="`${tab.id}-tab`" :key="tab.id" role="tab" :aria-selected="panel === tab.id" aria-controls="claim-panel" :class="{ active: panel === tab.id }" @click="showPanel(tab.id)">{{ t(tab.label) }}</button>
        </div>
        <div id="claim-panel" class="panel-content" role="tabpanel" :aria-labelledby="`${panel}-tab`">
          <InsurerReplanPreview v-if="!agentData.live && replanOpen && replan" :plan="replan" :pending="state.pending" :blocked="stats.blocked" @apply="approveReplan" @close="replanOpen = false" @claim="openClaim" />
          <InsurerClaimDetail v-else-if="detailOpen && selected" :claim="selected" :missions="selectedMissions" :audit="selectedAudit" :pending="state.pending || agentData.live" :agent-mode="agentData.live" :client-documents="documentRows" @back="showPanel('claims')" @plan="planAssistance" @permit-access="permitAccess" @accommodate="placeAccommodation" @schedule="scheduleSelected" @client="navigateTo('/client')" />
          <InsurerResponsePlan v-else-if="panel === 'plan'" :services="services" :live="agentData.live" @service="reviewService" @resources="showResources" />
          <InsurerResources v-else-if="panel === 'resources'" embedded :providers="providers" :missions="missions" :selected-provider="selectedProvider" @provider="selectProvider" @claim="openClaim" />
          <InsurerClaimList v-else :mode="panel" :claims="panel === 'priority' ? queue : pageClaims" :total="panel === 'priority' ? queue.length : filtered.length" :selected-id="selectedId" :page="page" :page-count="pageCount" @select="openClaim" @page="page = $event" @clear="clearFilters" />
        </div>
      </aside>
    </div>
    <UiDialog v-if="agentOpen" :title="t('AI decision proposal')" wide @close="agentOpen = false"><InsurerAgentPanel :reference="agentReference" :proposal="agentProposal" :busy="agentBusy" :error="agentError" v-model:consent="agentConsent" v-model:reviewed="humanReviewed" v-model:note="humanNote" @analyze="analyzeAgent(agentReference, agentConsent)" @approve="applyAgentDecision('approve')" @reject="applyAgentDecision('reject')" @download="downloadPolicy(agentReference)" @document="(id, name) => downloadPolicy(agentReference, id, name)" /></UiDialog>
    <UiDialog v-if="reportOpen" :title="t('Disaster Network — Situation report')" wide @close="reportOpen = false"><div class="report-body"><span class="report-label">{{ t('Portfolio situation report') }}</span><p>{{ t('This snapshot uses the entire portfolio, independently of current map filters.') }}</p><pre>{{ reportText }}</pre><UiButton class="dark" @click="downloadReport"><UiIcon name="Download" :size="15" />{{ t('Download report') }}</UiButton></div></UiDialog>
  </div>
</template>

<style scoped>
.agent-controls { display: flex; align-items: center; gap: 14px; padding: 4px 24px; min-height: 31px; border-bottom: 1px solid #e7e7ed; font-size: 11px; }
.workspace-notice { padding: 7px 24px; font-size: 10px; line-height: 16px; color: #8c713f; background: #fff9ec; border-bottom: 1px solid #f1e6d0; }
.agent-controls > span { flex: 1; color: #858590; }
.agent-controls > button, .agent-controls > a { font-size: 10px; text-decoration: underline; }
.live-report-source { padding: 5px 24px; color: #777781; font-size: 9px; line-height: 14px; border-bottom: 1px solid #e8e8ed; }
.insurer-page { display: flex; flex: 1; flex-direction: column; min-width: 1050px; min-height: 0; overflow: hidden; background: #fff; }
.insurer-page > :not(.operations-body) { flex-shrink: 0; }
.operations-heading { display: flex; align-items: center; justify-content: space-between; gap: 20px; padding: 10px 24px 8px; }.operations-heading h1 { font-size: 24px; letter-spacing: -0.7px; line-height: 30px; }.operations-heading p { display: flex; align-items: center; gap: 10px; margin-top: 3px; color: #9696a2; font-size: 11px; }.operations-heading p > span { color: #bcbcc4; }
.heading-actions { display: flex; align-items: center; gap: 10px; }.demo-data { border: 1px solid var(--border); border-radius: 20px; color: #9b9ba6; font-size: 9px; padding: 9px 12px; white-space: nowrap; }
.operations-body { display: flex; flex: 1; align-items: stretch; min-height: 0; overflow: hidden; }
.map-column { display: flex; flex-direction: column; flex: 1; min-width: 0; min-height: 0; }
.operations-panel { display: flex; flex-direction: column; width: 380px; flex-shrink: 0; border-left: 1px solid var(--border); background: #fff; min-height: 0; overflow: hidden; }
.panel-tabs { display: flex; align-items: stretch; padding: 0 12px; border-bottom: 1px solid #eeeeef; flex-shrink: 0; }.panel-tabs button { flex: 1; padding: 13px 4px; border-bottom: 2px solid transparent; font-size: 10px; color: #9a9aa5; }.panel-tabs button.active { border-color: #29292f; color: var(--ink); font-weight: 550; }.panel-content { display: flex; flex-direction: column; flex: 1; min-height: 0; }
.operation-notice { display: flex; align-items: center; gap: 10px; padding: 11px 24px; border-bottom: 1px solid #f4dbdf; background: #fff5f5; color: #b8787f; font-size: 10px; }.operation-notice > span { flex: 1; }.operation-notice button { display: flex; align-items: center; gap: 6px; font-size: 10px; }.operation-notice.applied { color: #5f9b76; background: #f1fbf5; border-color: #d4ecdd; }
.operation-error { padding: 12px 24px; color: #bb6871; background: #fff4f4; font-size: 11px; border-bottom: 1px solid #f1dddd; }
.report-body { display: flex; flex-direction: column; gap: 14px; padding: 5px; }.report-label { width: fit-content; color: #9c9ca7; background: #f3f3f5; border-radius: 5px; padding: 6px 9px; font-size: 10px; }.report-body > p { color: #9c9ca7; font-size: 11px; line-height: 19px; }.report-body pre { white-space: pre-wrap; overflow-wrap: anywhere; color: #747481; font: 12px/23px -apple-system, BlinkMacSystemFont, sans-serif; background: #f7f7f9; padding: 18px; margin: 0; border-radius: 8px; }.report-body :deep(.button) { align-self: flex-start; }
</style>
