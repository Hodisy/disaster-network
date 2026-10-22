<script setup>
defineProps({ scenarios: { type: Array, required: true } })
const { t } = useAppI18n()
</script>

<template>
  <section id="scenarios" class="scenarios" aria-labelledby="scenarios-title">
    <span class="section-label">{{ t('Other scenarios') }}</span>
    <h2 id="scenarios-title">{{ t('Different disasters. The same need to coordinate.') }}</h2>
    <p class="section-intro">{{ t('Five response scenarios beyond Jules’s flood claim.') }}</p>
    <div class="scenario-list">
      <article v-for="scenario in scenarios" :key="scenario.id" class="scenario" :aria-labelledby="`scenario-${scenario.id}`">
        <img :src="scenario.image" :alt="t(scenario.alt)" width="1536" height="1024" loading="lazy" decoding="async" />
        <div class="scenario-heading"><h3 :id="`scenario-${scenario.id}`">{{ t(scenario.title) }}</h3><span>{{ t(scenario.location) }}</span></div>
        <p class="description">{{ t(scenario.description) }}</p>
        <dl><div><dt>{{ t('Household need') }}</dt><dd>{{ t(scenario.need) }}</dd></div><div><dt>{{ t('Coordinated response') }}</dt><dd>{{ t(scenario.response) }}</dd></div></dl>
      </article>
    </div>
    <div class="scenario-note"><span>{{ t('Response scenarios') }}</span><p>{{ t('Apply the coordination approach to other hazards, with contracts, local rules and specialist review adapted to each situation.') }}</p></div>
  </section>
</template>

<style scoped>
.scenarios { max-width: 1264px; margin: 0 auto; padding: 112px 32px 96px; scroll-margin-top: 80px; }
.section-label { display: flex; align-items: center; justify-content: center; width: fit-content; margin: 0 auto; padding: 3px 11px; background: #f5f5f5; border-radius: 20px; color: #777; font-size: 12px; line-height: 20px; }
h2 { max-width: 740px; margin: 24px auto 0; text-align: center; font-size: 40px; line-height: 48px; letter-spacing: -1.2px; font-weight: 400; color: #0a0a0a; }
.section-intro { margin-top: 24px; text-align: center; font-size: 17px; line-height: 28px; color: #737373; }
.scenario-list { display: flex; flex-wrap: wrap; gap: 48px 28px; margin-top: 56px; }
.scenario { width: calc((100% - 56px) / 3); min-width: 0; }
.scenario:nth-child(n+4) { width: calc((100% - 28px) / 2); }
.scenario > img { display: block; width: 100%; height: 240px; object-fit: cover; border-radius: 10px; background: #f3f3f4; }
.scenario:nth-child(n+4) > img { height: 306px; }
.scenario-heading { display: flex; flex-direction: column; gap: 4px; margin-top: 19px; }
.scenario-heading h3 { font-size: 18px; line-height: 25px; letter-spacing: -.3px; font-weight: 500; }
.scenario-heading > span { font-size: 12px; line-height: 20px; color: #8a8a93; }
.description { margin-top: 11px; max-width: 470px; color: #737373; font-size: 14px; line-height: 24px; }
dl { margin: 18px 0 0; border-top: 1px solid #e7e7ea; }
dl > div { display: flex; gap: 16px; padding: 12px 0; border-bottom: 1px solid #eeeeee; }
dt { flex: 0 0 100px; color: #999; font-size: 10px; line-height: 18px; }
dd { margin: 0; color: #777; font-size: 11px; line-height: 18px; }
.scenario-note { display: flex; flex-direction: column; align-items: center; gap: 12px; padding-top: 30px; margin-top: 35px; }
.scenario-note > span { color: #888; font-size: 11px; line-height: 20px; }
.scenario-note > p { max-width: 840px; color: #999; font-size: 11px; line-height: 20px; text-align: center; }
</style>
