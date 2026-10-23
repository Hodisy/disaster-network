<script setup>
defineProps({ services: { type: Array, required: true }, live: Boolean })
defineEmits(['service', 'resources'])
const { t } = useAppI18n()
</script>

<template>
  <section class="response-plan">
    <header><h2>{{ t('Response priorities') }}</h2><p>{{ t('Entire portfolio · Needs and service capacity') }}</p></header>
    <div class="plan-scroll">
      <div class="plan-intro"><UiIcon name="ShieldCheck" :size="16" /><p>{{ t(live ? 'Review a dossier to request an AI proposal. Assignments require human validation.' : 'Prioritise accommodation, confirm site access and review provider capacity before assigning missions.') }}</p></div>
      <article v-for="service in services" :key="service.skill" class="service-card">
        <div class="service-heading"><UiIcon :name="service.skill === 'Temporary accommodation' ? 'House' : 'Truck'" :size="16" /><h3>{{ t(service.skill) }}</h3><strong>{{ service.households }}</strong></div>
        <p>{{ t('Households with this reported need') }}</p>
        <div class="service-capacity"><span>{{ t('{count} available providers', { count: service.available }) }}</span><span>{{ t('{count} scheduled missions', { count: service.scheduled }) }}</span></div>
        <div v-for="capacity in service.units" :key="capacity.unit" class="capacity-line"><span>{{ t(capacity.unit) }}</span><strong>{{ t('{count} free', { count: capacity.remaining }) }}</strong></div>
        <p v-if="!service.units.length" class="blocked">{{ t('No matching provider capacity is listed.') }}</p>
        <p v-if="service.blocked" class="blocked">{{ t('{count} missions need access or availability review', { count: service.blocked }) }}</p>
        <button @click="$emit('service', service.skill)">{{ t(service.skill === 'Temporary accommodation' ? 'Review housing requests' : 'Review needs and access') }}<UiIcon name="ChevronRight" :size="13" /></button>
      </article>
      <p v-if="!services.length" class="empty-plan">{{ t('No service needs or missions are recorded in this portfolio.') }}</p>
    </div>
    <footer><UiButton class="dark" @click="$emit('resources')">{{ t('Review deployment') }}<UiIcon name="ArrowRight" :size="14" /></UiButton><p>{{ t('Capacity is service-specific. Review access and each proposed assignment before validation.') }}</p></footer>
  </section>
</template>

<style scoped>
.response-plan { display: flex; flex: 1; flex-direction: column; min-height: 0; }
.response-plan > header { flex-shrink: 0; padding: 16px 18px 12px; }
h2 { font-size: 14px; letter-spacing: -.3px; }
.response-plan > header p { margin-top: 5px; font-size: 10px; color: #91919c; }
.plan-scroll { flex: 1; min-height: 0; overflow-y: auto; overscroll-behavior: contain; padding: 0 16px 12px; }
.plan-intro { display: flex; align-items: flex-start; gap: 8px; padding: 10px; margin-bottom: 10px; border-radius: 7px; background: #f5f5f7; color: #777783; }
.plan-intro p { font-size: 10px; line-height: 16px; }
.service-card { padding: 13px 0; border-bottom: 1px solid var(--border); }
.service-heading { display: flex; align-items: center; gap: 8px; }
h3 { flex: 1; font-size: 12px; font-weight: 550; }
.service-heading > strong { font-size: 20px; font-weight: 550; letter-spacing: -.5px; }
.service-card > p { margin-top: 5px; font-size: 9px; line-height: 14px; color: #93939e; }
.service-capacity { display: flex; justify-content: space-between; gap: 10px; margin-top: 10px; font-size: 10px; color: #73737f; }
.capacity-line { display: flex; align-items: baseline; justify-content: space-between; gap: 10px; margin-top: 6px; font-size: 9px; color: #92929d; }
.capacity-line strong { color: #5e826c; font-weight: 500; }
.service-card > .blocked { color: #ae7d39; }
.service-card > button { display: flex; align-items: center; justify-content: space-between; gap: 8px; width: 100%; margin-top: 10px; padding: 0; text-align: left; font-size: 10px; color: #555561; }
.service-card > button:hover { color: #171719; }
footer { flex-shrink: 0; padding: 12px 16px; border-top: 1px solid var(--border); }
footer :deep(.button) { width: 100%; justify-content: space-between; }
footer p, .empty-plan { font-size: 9px; line-height: 15px; color: #94949f; margin-top: 8px; }

@media (max-width: 1000px) {
  .response-plan { flex: none; }
  .plan-scroll { flex: none; overflow: visible; }
  .service-card > button { min-height: 44px; font-size: 12px; }
}

</style>
