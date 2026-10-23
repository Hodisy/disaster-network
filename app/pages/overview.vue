<script setup>
definePageMeta({ marketing: true, alias: ['/'] })
const { locale, setLocale, t } = useAppI18n()
const { workspaceImage, story, portfolio, scenarios } = useMarketingContent()

useHead(() => ({
  title: t('Disaster Network — Catastrophe coordination for insurers'),
  meta: [{ name: 'description', content: t('An AI claims hotline, a shared claim file and provider coordination for insurance teams.') }]
}))
</script>

<template>
  <div id="top" class="marketing-page">
    <a href="#overview-content" class="skip-link">{{ t('Skip to content') }}</a>
    <MarketingHeader :locale="locale" @set-locale="setLocale" />
    <main id="overview-content" tabindex="-1">
      <MarketingHero>
        <MarketingWorkspacePreview :image="workspaceImage" />
      </MarketingHero>
      <MarketingJourney :story="story" :portfolio="portfolio" />
      <MarketingScenarios :scenarios="scenarios" />
    </main>
    <MarketingFooter />
  </div>
</template>

<style scoped>
.marketing-page { width: 100%; min-width: 0; min-height: 100%; background: #fff; color: #171719; }
.skip-link { position: fixed; z-index: 100; top: -80px; left: 20px; padding: 12px 18px; background: #171719; color: #fff; border-radius: 8px; font-size: 13px; line-height: 20px; }
.skip-link:focus { top: 12px; }
</style>
