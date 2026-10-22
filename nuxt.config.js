export default defineNuxtConfig({
  compatibilityDate: '2026-10-22',
  ssr: false,
  devtools: { enabled: false },
  nitro: {
    // Bundle shared seed modules so their paths remain valid after server compilation.
    externals: { inline: [/[/\\]database[/\\]/, /[/\\]app[/\\]config[/\\]demo\.js$/] }
  },
  runtimeConfig: {
    databaseUrl: '',
    databaseCaCertPath: '',
    databaseCaCert: '',
    demoResetEnabled: false,
    demoResetToken: '',
    hotlineEnabled: false,
    hotlineAccessToken: '',
    hotlineOrigin: 'http://127.0.0.1:3000',
    openaiApiKey: '',
    openaiLiveModel: 'gpt-live-1',
    openaiBackendModel: 'gpt-6-luna',
    openaiEvidenceModel: 'gpt-6-luna',
    insurerAgentEnabled: false,
    insurerAccessToken: '',
    openaiInsurerModel: 'gpt-6-luna',
    insurerWebSearchEnabled: false,
    openaiLiveVoice: 'marin'
  },
  css: ['~/assets/css/main.css'],
  routeRules: { '/': { redirect: '/hotline' } },
  app: {
    head: {
      htmlAttrs: { lang: 'en' },
      title: 'Disaster Network — Support for your insurance claim',
      link: [{ rel: 'icon', type: 'image/svg+xml', sizes: 'any', href: '/brand/disaster-network-mark.svg' }],
      meta: [{ name: 'description', content: 'Insurance claim coordination with optional AI voice and verified document collection. Synthetic evaluation scenario.' }]
    }
  }
})
