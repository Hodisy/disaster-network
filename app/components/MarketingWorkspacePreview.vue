<script setup>
defineProps({ preview: { type: Object, required: true } })
defineEmits(['select', 'clear', 'panel', 'page', 'open', 'view', 'filter', 'clear-filters', 'advanced', 'provider'])
const { t, formatMoney, formatDate } = useAppI18n()
</script>

<template>
  <section id="operations" class="product-band" :aria-label="t('Insurer workspace preview')">
    <div class="product-frame">
      <aside class="preview-sidebar">
        <NuxtLink to="/insurer" class="preview-brand"><img src="/brand/disaster-network-logo.svg" alt="Disaster Network" width="156" height="25" /></NuxtLink>
        <p class="sidebar-label">{{ t('Insurer workspace') }}</p>
        <nav :aria-label="t('Insurer navigation')"><NuxtLink to="/insurer" class="operations-link"><UiIcon name="Building2" :size="17" />{{ t('Operations') }}</NuxtLink></nav>
        <NuxtLink to="/insurer" class="insurer-account"><span class="account-avatar">AS</span><span>Allstorm Insurance<small>{{ t('Insurer account') }}</small></span><UiIcon name="ChevronRight" :size="13" /></NuxtLink>
      </aside>
      <div class="preview-workspace">
        <div class="product-bar"><span>{{ t('Insurer workspace') }}<UiIcon name="ChevronRight" :size="12" />{{ t('Operations') }}</span><span>{{ t('Berre Lagoon · Portfolio overview') }}</span></div>
        <div class="product-heading"><div><h2>{{ t('Berre Lagoon flood response') }}</h2><p>{{ t('Event: {date}', { date: formatDate('2026-10-20', { month: 'short' }) }) }} · {{ t('Scenario snapshot: {date}', { date: formatDate('2026-10-22', { month: 'short' }) }) }}</p></div><NuxtLink to="/insurer" class="open-link">{{ t('Open insurer workspace') }}<UiIcon name="ArrowUpRight" :size="15" /></NuxtLink></div>
        <div class="portfolio-totals"><span><strong>{{ preview.stats.households }}</strong>{{ t('Portfolio households') }}</span><span><strong>{{ preview.stats.people }}</strong>{{ t('People in households') }}</span><span class="estimate"><strong>{{ formatMoney(preview.stats.gross) }}</strong>{{ t('Known gross estimates') }}<small>{{ t('Partial · Not approved') }}</small></span></div>
        <InsurerToolbar :view="preview.view" :query="preview.query" :options="preview.options" :advanced="preview.advanced" @view="$emit('view', $event)" @filter="(key, value) => $emit('filter', key, value)" @advanced="$emit('advanced')" @clear="$emit('clear-filters')" />
        <div class="product-body" :class="{ mobilization: preview.view === 'mobilization' }">
          <div class="map-column">
            <InsurerMap :claims="preview.claims" :providers="preview.providers" :assignments="preview.assignments" :proposed="preview.proposed" :selected="preview.selected" :show-popup="!!preview.selected" :stats="preview.stats" appearance="reference-a" open-label="Open insurer workspace" @select="$emit('select', $event)" @dismiss-popup="$emit('clear')" @open="$emit('open')" @provider="$emit('provider', $event)" />
            <div v-if="preview.view === 'mobilization'" class="resource-preview"><InsurerResources :providers="preview.resources" :missions="preview.missions" :selected-provider="preview.selectedProvider" @provider="$emit('provider', $event)" @claim="$emit('select', $event)" /></div>
          </div>
          <aside class="priority-panel" :aria-label="t('Claim workspace')">
            <div class="panel-tabs" role="tablist" :aria-label="t('Right panel view')"><button id="preview-priority-tab" role="tab" :aria-selected="preview.panel === 'priority'" aria-controls="preview-claim-panel" :class="{ active: preview.panel === 'priority' }" @click="$emit('panel', 'priority')">{{ t('Priority queue') }}</button><button id="preview-claims-tab" role="tab" :aria-selected="preview.panel === 'claims'" aria-controls="preview-claim-panel" :class="{ active: preview.panel === 'claims' }" @click="$emit('panel', 'claims')">{{ t('All claims') }}</button></div>
            <div id="preview-claim-panel" class="panel-content" role="tabpanel" :aria-labelledby="preview.panel === 'priority' ? 'preview-priority-tab' : 'preview-claims-tab'"><InsurerClaimList :mode="preview.panel" :claims="preview.records" :total="preview.total" :selected-id="preview.selected?.id" :page="preview.page" :page-count="preview.pageCount" @select="$emit('select', $event)" @page="$emit('page', $event)" @clear="$emit('clear-filters')" /></div>
          </aside>
        </div>
      </div>
    </div>
    <p class="preview-caption">{{ t('Select a household to locate it. Open the workspace to review claims and assistance.') }}</p>
  </section>
</template>

<style scoped>
.product-band { padding: 28px 24px 28px; background: #f5f7f8; scroll-margin-top: 64px; }
.product-frame { display: flex; align-items: stretch; max-width: 1440px; margin: 0 auto; border: 1px solid #e3e6e9; border-radius: 13px; overflow: hidden; background: #fff; box-shadow: 0 14px 40px #1726350a; }
.preview-sidebar { display: flex; flex-direction: column; width: 192px; flex-shrink: 0; padding: 24px 14px 18px; background: #f7f7f6; border-right: 1px solid #eaeaec; }
.preview-brand { display: block; padding: 0 3px; }
.preview-brand img { display: block; width: 156px; height: 25px; }
.sidebar-label { margin: 40px 0 13px; padding: 0 9px; color: #83838d; font-size: 9px; line-height: 16px; text-transform: uppercase; letter-spacing: .6px; }
.operations-link { display: flex; align-items: center; gap: 12px; padding: 12px 10px; border-radius: 8px; background: #e9e9e8; font-size: 12px; line-height: 20px; }
.operations-link:hover { background: #dfdfdf; }
.insurer-account { display: flex; align-items: center; gap: 8px; margin-top: auto; padding-top: 32px; }
.account-avatar { display: flex; align-items: center; justify-content: center; flex: 0 0 29px; height: 29px; border-radius: 50%; background: #171719; color: #fff; font-size: 9px; }
.insurer-account > span:nth-child(2) { flex: 1; font-size: 10px; line-height: 17px; }
.insurer-account small { display: block; color: #93939d; font-size: 9px; line-height: 15px; }
.insurer-account > svg { color: #80808a; }
.preview-workspace { flex: 1; min-width: 0; }
.product-bar { display: flex; justify-content: space-between; align-items: center; gap: 24px; padding: 15px 22px; border-bottom: 1px solid #eeeeee; font-size: 11px; line-height: 18px; color: #85858d; }
.product-bar > span:first-child { display: flex; align-items: center; gap: 12px; }
.product-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 24px; padding: 25px 24px 18px; }
.product-heading h2 { font-size: 27px; line-height: 35px; letter-spacing: -.9px; font-weight: 600; }
.product-heading p { margin-top: 7px; color: #9696a2; font-size: 11px; line-height: 18px; }
.open-link { display: inline-flex; align-items: center; gap: 8px; margin-top: 6px; font-size: 11px; line-height: 18px; white-space: nowrap; }
.open-link:hover { text-decoration: underline; text-underline-offset: 4px; }
.portfolio-totals { display: flex; align-items: center; gap: 24px; padding: 0 24px 18px; border-bottom: 1px solid #eeeeee; }
.portfolio-totals > span { display: flex; align-items: baseline; gap: 8px; font-size: 10px; line-height: 18px; color: #9999a4; }
.portfolio-totals strong { font-size: 17px; line-height: 25px; font-weight: 550; letter-spacing: -.5px; color: #44444f; }
.portfolio-totals .estimate { margin-left: auto; }
.portfolio-totals small { font-size: 8px; }
.product-body { display: flex; align-items: stretch; height: 740px; }
.product-body.mobilization { height: 1140px; }
.map-column { display: flex; flex-direction: column; flex: 1; min-width: 0; min-height: 0; }
.resource-preview { flex-shrink: 0; max-height: 390px; overflow-y: auto; }
.priority-panel { display: flex; flex-direction: column; width: 320px; flex-shrink: 0; min-height: 0; border-left: 1px solid #eaeaec; background: #fff; }
.panel-tabs { display: flex; align-items: stretch; padding: 0 18px; border-bottom: 1px solid #eeeeef; flex-shrink: 0; }
.panel-tabs button { flex: 1; padding: 16px 5px; border-bottom: 2px solid transparent; font-size: 11px; line-height: 16px; color: #9a9aa5; }
.panel-tabs button.active { border-color: #29292f; color: #171719; font-weight: 550; }
.panel-content { display: flex; flex-direction: column; flex: 1; min-height: 0; }
.preview-caption { max-width: 1200px; margin: 18px auto 0; color: #909098; font-size: 11px; line-height: 19px; text-align: center; }
</style>
