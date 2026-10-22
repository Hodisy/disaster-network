<script setup>
defineProps({ story: { type: Object, required: true }, portfolio: { type: Array, required: true } })
const { t, formatDate } = useAppI18n()
</script>

<template>
  <section id="jules" class="journey" aria-labelledby="journey-title">
    <div class="section-heading">
      <span class="section-label">{{ t('How it works') }}</span>
      <h2 id="journey-title">{{ t('One call becomes an actionable claim.') }}</h2>
      <p>{{ t('Start with Jules: three people need accommodation after flooding near Istres. Keep the conversation, evidence and next actions connected.') }}</p>
    </div>
    <div class="story-row">
      <figure class="story-photo">
        <img :src="story.image" :alt="t(story.alt)" width="1536" height="1024" loading="lazy" decoding="async" />
        <figcaption><span>{{ t('Istres · Berre Lagoon area') }}</span><span>{{ t('AI-generated illustration') }}</span></figcaption>
        <p class="dates">{{ t('Flood: {eventDate} · Report: {reportDate}', { eventDate: formatDate(story.eventDate), reportDate: formatDate(story.reportDate) }) }}</p>
      </figure>
      <div class="voice-story">
        <div class="voice-illustration"><img :src="story.hotlineImage" :alt="t(story.hotlineAlt)" width="1536" height="1024" loading="lazy" decoding="async" /></div>
        <h3>{{ t('An AI hotline that carries the context forward.') }}</h3>
        <p>{{ t('Identify the caller, retrieve the contract context and capture the need. Request the supporting documents in My claim.') }}</p>
        <div class="conversation" :aria-label="t('Illustrative conversation')"><blockquote><cite>Jules</cite><p>{{ t('We need somewhere to stay. There are three of us.') }}</p></blockquote><blockquote><cite>{{ t('AI hotline') }}</cite><p>{{ t('I’ll check your contract and list the documents to send in My claim.') }}</p></blockquote></div>
        <NuxtLink to="/hotline" class="journey-link">{{ t('Try Jules’s hotline') }}</NuxtLink>
      </div>
    </div>
    <ol class="steps">
      <li v-for="step in story.steps" :key="step.number"><span class="step-number" aria-hidden="true">{{ step.number }}</span><h3>{{ t(step.title) }}</h3><p>{{ t(step.description) }}</p><p class="step-detail">{{ t(step.detail) }}</p></li>
    </ol>
    <p class="review-note">{{ t('Collecting documents does not approve coverage, book a hotel or close the claim.') }}</p>
    <div class="portfolio-band">
      <div class="portfolio-context"><h3>{{ t('See the wider response') }}</h3><p>{{ t('For the insurer, each household becomes part of a shared operational picture: unmet needs, missing evidence, access constraints and available providers.') }}</p></div>
      <div class="portfolio-summary"><dl><div v-for="item in portfolio" :key="item.label"><dt>{{ item.value }}</dt><dd>{{ t(item.label) }}</dd></div></dl><p>{{ t('Household portfolio') }}</p></div>
    </div>
  </section>
</template>

<style scoped>
.journey { max-width: 1264px; margin: 0 auto; padding: 112px 32px 0; scroll-margin-top: 80px; }
.section-heading { text-align: center; }
.section-label { display: inline-flex; align-items: center; padding: 3px 11px; background: #f5f5f5; border-radius: 20px; color: #777; font-size: 12px; line-height: 20px; }
h2 { max-width: 850px; margin: 24px auto 0; font-size: 40px; line-height: 48px; font-weight: 400; letter-spacing: -1.2px; color: #0a0a0a; }
.section-heading > p { max-width: 760px; margin: 24px auto 0; color: #737373; font-size: 17px; line-height: 28px; }
.story-row { display: flex; align-items: flex-start; gap: 32px; margin-top: 64px; }
.story-photo { flex: 0 0 58%; min-width: 0; margin: 0; }
.story-photo > img { display: block; width: 100%; height: 530px; object-fit: cover; object-position: 63% center; border-radius: 14px; background: #f5f5f5; }
figcaption { display: flex; flex-direction: column; gap: 4px; margin-top: 14px; color: #71717b; font-size: 12px; line-height: 19px; }
figcaption > span:last-child { color: #93939c; font-size: 10px; line-height: 17px; }
.dates { margin-top: 9px; color: #9999a0; font-size: 10px; line-height: 17px; }
.voice-story { flex: 1; min-width: 0; }
.voice-illustration { display: flex; align-items: center; justify-content: center; height: 210px; padding: 12px 20px; border-radius: 16px; background: #f7f7f7; }
.voice-illustration > img { display: block; width: 100%; height: 100%; object-fit: contain; }
.voice-story > h3 { margin-top: 27px; font-size: 29px; line-height: 35px; letter-spacing: -.9px; font-weight: 400; }
.voice-story > p { margin-top: 18px; color: #737373; font-size: 14px; line-height: 24px; }
.conversation { margin-top: 23px; }
blockquote { display: flex; align-items: baseline; gap: 14px; margin: 13px 0 0; }
cite { flex: 0 0 62px; color: #36363c; font-size: 10px; line-height: 18px; font-weight: 500; font-style: normal; }
blockquote p { color: #898991; font-size: 12px; line-height: 20px; }
.journey-link { display: inline-flex; align-items: center; margin-top: 24px; padding-bottom: 4px; border-bottom: 1px solid #b5b5bd; font-size: 13px; line-height: 22px; }
.journey-link:hover { border-color: #171719; }
.steps { display: flex; align-items: stretch; gap: 24px; margin-top: 56px; }
.steps li { display: flex; flex: 1; min-width: 0; flex-direction: column; align-items: flex-start; padding: 30px 30px 28px; background: #fafafa; border-radius: 20px; }
.step-number { display: inline-flex; align-items: center; justify-content: center; min-width: 30px; padding: 2px 8px; border-radius: 20px; font-size: 12px; line-height: 20px; background: #171719; color: #fff; }
.steps h3 { margin-top: 17px; font-size: 18px; line-height: 25px; font-weight: 500; letter-spacing: -.3px; }
.steps p { margin-top: 15px; color: #737373; font-size: 14px; line-height: 24px; }
.steps .step-detail { margin-top: auto; padding-top: 22px; color: #94949b; font-size: 11px; line-height: 19px; }
.review-note { max-width: 760px; margin: 23px auto 0; text-align: center; color: #94949b; font-size: 11px; line-height: 19px; }
.portfolio-band { display: flex; align-items: center; justify-content: space-between; gap: 64px; margin-top: 62px; padding: 37px 0; border-top: 1px solid #e7e7e7; border-bottom: 1px solid #e7e7e7; }
.portfolio-context { flex: 1; max-width: 540px; }
.portfolio-context h3 { font-size: 28px; line-height: 36px; font-weight: 400; letter-spacing: -.8px; }
.portfolio-context p { margin-top: 12px; color: #808080; font-size: 13px; line-height: 22px; }
.portfolio-summary { flex: 0 0 43%; }
.portfolio-summary dl { display: flex; margin: 0; }
.portfolio-summary dl > div { display: flex; flex: 1; align-items: center; flex-direction: column; gap: 7px; padding: 0 16px; border-left: 1px solid #e4e4e4; }
.portfolio-summary dl > div:first-child { border: 0; }
.portfolio-summary dt { font-size: 34px; line-height: 40px; font-weight: 400; letter-spacing: -1px; }
.portfolio-summary dd { margin: 0; font-size: 12px; line-height: 19px; color: #777; text-align: center; }
.portfolio-summary > p { margin-top: 20px; text-align: center; font-size: 10px; line-height: 17px; color: #999; }
</style>
