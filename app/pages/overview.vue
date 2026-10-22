<script setup>
definePageMeta({ marketing: true })
const { locale, setLocale, t } = useAppI18n()
const { story, portfolio, scenarios } = useMarketingContent()
const {
  preview, selectPreview, clearPreview, setPreviewPanel, setPreviewPage,
  setPreviewView, filterPreview, clearPreviewFilters, togglePreviewFilters, selectPreviewProvider
} = useMarketingWorkspacePreview()

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
        <MarketingWorkspacePreview :preview="preview" @select="selectPreview" @clear="clearPreview" @panel="setPreviewPanel" @page="setPreviewPage" @view="setPreviewView" @filter="filterPreview" @clear-filters="clearPreviewFilters" @advanced="togglePreviewFilters" @provider="selectPreviewProvider" @open="navigateTo('/insurer')" />
      </MarketingHero>
      <MarketingJourney :story="story" :portfolio="portfolio" />
      <MarketingScenarios :scenarios="scenarios" />
    </main>
    <MarketingFooter />
  </div>
</template>

<style scoped>
.marketing-page { min-width: 1120px; min-height: 100%; background: #fff; color: #171719; }
.skip-link { position: fixed; z-index: 100; top: -80px; left: 20px; padding: 12px 18px; background: #171719; color: #fff; border-radius: 8px; font-size: 13px; line-height: 20px; }
.skip-link:focus { top: 12px; }
</style>
