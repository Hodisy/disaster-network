<script setup>
const { t, formatTime, formatDate } = useAppI18n()
defineProps({ providers: Array, missions: Array, selectedProvider: String, embedded: Boolean })
defineEmits(['provider', 'claim'])
</script>

<template>
  <section class="resources" :class="{ embedded }">
    <header><h2>{{ t('Provider deployment and assignments') }}</h2><span>{{ t('Capacity units are service-specific.') }}</span></header>
    <div class="provider-table" role="table" :aria-label="t('Provider capacity')">
      <div class="table-header" role="row"><span role="columnheader">{{ t('Provider') }}</span><span role="columnheader">{{ t('Service') }}</span><span role="columnheader">{{ t('Capacity') }}</span><span role="columnheader">{{ t('Availability') }}</span></div>
      <div class="provider-scroll"><button v-for="provider in providers" :key="provider.id" class="provider-row" :class="{ selected: provider.id === selectedProvider, unavailable: !provider.available }" role="row" @click="$emit('provider', provider.id)"><span role="cell"><UiIcon :name="provider.skill === 'Temporary accommodation' ? 'House' : 'Truck'" :size="15" /><strong>{{ provider.name }}</strong></span><span role="cell">{{ t(provider.skill) }}</span><span role="cell">{{ provider.used }} / {{ provider.capacity }}<small>{{ t(provider.unit) }}</small></span><span role="cell"><i :class="provider.status.toLowerCase()" />{{ t(provider.status) }}<UiIcon name="ChevronRight" :size="12" /></span></button></div>
    </div>
    <div v-if="selectedProvider" class="provider-missions"><div class="mission-heading"><strong>{{ providers.find(provider => provider.id === selectedProvider)?.name }}</strong><button @click="$emit('provider', '')">{{ t('Show all providers') }}</button></div><p v-if="!missions.some(mission => mission.providerId === selectedProvider)">{{ t('No missions assigned to this provider.') }}</p><button v-for="mission in missions.filter(mission => mission.providerId === selectedProvider)" :key="mission.id" class="assignment" @click="$emit('claim', mission.claimId)"><span>{{ t(mission.claim.name, mission.claim.nameParameters) }}<small>{{ mission.claimId }} · {{ mission.claim.municipality }}</small></span><span>{{ t(mission.status) }}<small v-if="mission.start">{{ formatDate(mission.start, { month: 'short', year: undefined }) }} · {{ formatTime(mission.start) }}</small></span><UiIcon name="ChevronRight" :size="13" /></button></div>
    <p class="resource-note">{{ t('Assignment lines are illustrative connections, not routes or measured travel times.') }}</p>
  </section>
</template>

<style scoped>
.resources { display: flex; flex-direction: column; background: #fff; border-top: 1px solid var(--border); }
.resources.embedded > * { flex-shrink: 0; }
.resources.embedded { flex: 1; min-height: 0; overflow-y: auto; overscroll-behavior: contain; border-top: 0; }
.resources.embedded > header { flex-direction: column; align-items: flex-start; gap: 5px; padding: 16px 18px 12px; }
.embedded .provider-table { margin: 0 12px; }
.embedded .provider-scroll { max-height: none; overflow: visible; }
.embedded .table-header, .embedded .provider-row { gap: 6px; padding: 10px 5px; }
.embedded .provider-row > span:first-child { flex-direction: column; align-items: flex-start; gap: 5px; }
.embedded .provider-row strong { font-size: 10px; }
.embedded .provider-row > span:last-child { flex-wrap: wrap; font-size: 9px; }
.embedded .provider-missions { padding: 12px 16px; }
.resources > header { display: flex; justify-content: space-between; align-items: center; gap: 15px; padding: 18px 21px; }.resources h2 { font-size: 13px; }.resources > header > span { color: #a0a0ab; font-size: 9px; }
.provider-table { margin: 0 18px; }.table-header, .provider-row { display: flex; align-items: center; gap: 10px; text-align: left; width: 100%; }.table-header { background: #f6f6f8; padding: 10px; border-radius: 6px; font-size: 9px; color: #92929e; }.provider-scroll { max-height: 205px; overflow-y: auto; }
.table-header > span:nth-child(1), .provider-row > span:nth-child(1) { width: 28%; }.table-header > span:nth-child(2), .provider-row > span:nth-child(2) { width: 26%; }.table-header > span:nth-child(3), .provider-row > span:nth-child(3) { width: 26%; }.table-header > span:nth-child(4), .provider-row > span:nth-child(4) { flex: 1; }
.provider-row { border-bottom: 1px solid #eeeeef; padding: 12px 10px; font-size: 10px; color: #898995; }.provider-row.selected { background: #f2f2f5; }.provider-row:hover { background: #f8f8fa; }.provider-row.unavailable { background: #fff3f4; }
.provider-row > span { display: flex; align-items: center; gap: 8px; }.provider-row strong { font-weight: 500; font-size: 10px; color: #62626e; }.provider-row > span:nth-child(3) { flex-direction: column; align-items: flex-start; gap: 4px; }.provider-row small { color: #a2a2ac; font-size: 8px; }.provider-row > span:last-child { font-size: 9px; }
.provider-row i { width: 5px; height: 5px; border-radius: 50%; flex-shrink: 0; }.available { background: #33b575; }.active { background: #747784; }.unavailable { background: #df454e; }
.provider-missions { padding: 14px 21px; border-top: 1px solid var(--border); }.mission-heading { display: flex; justify-content: space-between; gap: 12px; font-size: 11px; margin-bottom: 10px; }.mission-heading strong { font-weight: 500; }.mission-heading button { color: #92929e; text-decoration: underline; font-size: 9px; }.provider-missions > p { font-size: 10px; color: #a1a1ac; }
.assignment { display: flex; align-items: center; gap: 15px; width: 100%; border-top: 1px solid #f0f0f3; padding: 10px 0; text-align: left; font-size: 10px; color: #797987; }.assignment > span:first-child { flex: 1; }.assignment small { display: block; font-size: 9px; color: #a0a0ab; margin-top: 4px; }
.resource-note { font-size: 9px; color: #aaaab4; padding: 12px 21px; }

@media (max-width: 1000px) {
  .resources.embedded { flex: none; overflow: visible; }
  .provider-row > span { min-width: 0; overflow-wrap: anywhere; }
  .embedded .provider-row > span { font-size: 10px; }
  .embedded .provider-row > span:last-child { flex-direction: column; align-items: flex-start; gap: 5px; }
  .assignment { min-height: 44px; }
}

</style>
