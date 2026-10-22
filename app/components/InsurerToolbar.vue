<script setup>
const { t } = useAppI18n()
defineProps({ view: String, query: Object, options: Object, advanced: Boolean })
defineEmits(['view', 'filter', 'advanced', 'clear'])
</script>

<template>
  <div class="operations-toolbar">
    <div class="primary-row">
      <div class="view-tabs" role="tablist" :aria-label="t('Workspace view')">
        <button v-for="tab in ['situation', 'mobilization']" :key="tab" role="tab" :aria-selected="view === tab" :class="{ active: view === tab }" @click="$emit('view', tab)">{{ t(tab === 'situation' ? 'Situation' : 'Mobilization') }}</button>
      </div>
      <label class="search-field"><UiIcon name="Search" :size="16" /><input :value="query.search" :aria-label="t('Search claims')" :placeholder="t('Search name, claim or location…')" @input="$emit('filter', 'search', $event.target.value)" /></label>
      <select :value="query.municipality" :aria-label="t('Municipality')" @change="$emit('filter', 'municipality', $event.target.value)"><option value="">{{ t('Municipality') }}</option><option v-for="value in options.municipalities" :key="value">{{ value }}</option></select>
      <select :value="query.priority" :aria-label="t('Priority')" @change="$emit('filter', 'priority', $event.target.value)"><option value="">{{ t('Priority') }}</option><option v-for="value in options.priorities" :key="value">{{ value }}</option></select>
      <select :value="query.need" :aria-label="t('Need')" @change="$emit('filter', 'need', $event.target.value)"><option value="">{{ t('Need') }}</option><option v-for="value in options.needs" :key="value" :value="value">{{ t(value) }}</option></select>
      <select :value="query.stage" :aria-label="t('Stage')" @change="$emit('filter', 'stage', $event.target.value)"><option value="">{{ t('Stage') }}</option><option v-for="(value, index) in options.stages" :key="value" :value="String(index)">{{ t(value) }}</option></select>
      <button class="more-button" :class="{ active: advanced || query.group || query.access }" :aria-expanded="advanced" :aria-label="t('More filters')" @click="$emit('advanced')"><UiIcon name="SlidersHorizontal" :size="16" /></button>
    </div>
    <div v-if="advanced" class="advanced-row">
      <span>{{ t('Additional filters') }}</span>
      <select :value="query.group" :aria-label="t('Reported situation')" @change="$emit('filter', 'group', $event.target.value)"><option value="">{{ t('All situations') }}</option><option v-for="group in options.groups" :key="group.id" :value="group.id">{{ group.id }} · {{ t(group.label) }}</option></select>
      <select :value="query.access" :aria-label="t('Site access')" @change="$emit('filter', 'access', $event.target.value)"><option value="">{{ t('All access states') }}</option><option v-for="value in options.access" :key="value" :value="value">{{ t(value) }}</option></select>
      <button class="clear-button" @click="$emit('clear')">{{ t('Clear filters') }}</button>
      <span class="filter-note">{{ t('Situation, priority and coverage are separate.') }}</span>
    </div>
  </div>
</template>

<style scoped>
.operations-toolbar { display: flex; flex-direction: column; border-bottom: 1px solid var(--border); background: #fff; }
.primary-row { display: flex; align-items: center; gap: 9px; padding: 0 23px; min-height: 48px; }
.view-tabs { display: flex; align-self: stretch; gap: 22px; margin-right: 16px; flex-shrink: 0; }
.view-tabs button { padding: 0 2px; border-bottom: 2px solid transparent; font-size: 12px; color: #8a8a95; }
.view-tabs button.active { color: var(--ink); border-color: var(--ink); font-weight: 600; }
.search-field { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 120px; border: 1px solid var(--border); border-radius: 7px; padding: 9px 11px; color: #93939e; }
.search-field input { width: 100%; min-width: 0; border: 0; padding: 0; font-size: 11px; background: none; color: var(--ink); }
.search-field input::placeholder { color: #9999a4; }
select { min-width: 84px; max-width: 140px; height: 34px; padding: 0 9px; border: 1px solid var(--border); background: #fff; border-radius: 7px; color: #555560; font-size: 10px; cursor: pointer; }
.more-button { display: inline-flex; align-items: center; justify-content: center; width: 33px; height: 34px; border: 1px solid var(--border); border-radius: 7px; color: #777783; flex-shrink: 0; }
.more-button.active { background: #f0f0f3; color: var(--ink); }
.advanced-row { display: flex; align-items: center; gap: 12px; padding: 0 23px 13px; }
.advanced-row > span { color: #8f8f9a; font-size: 10px; }
.advanced-row select { max-width: 230px; }
.clear-button { font-size: 10px; color: #666671; text-decoration: underline; }
.filter-note { margin-left: auto; }
</style>
