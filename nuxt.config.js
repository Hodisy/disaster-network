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
      link: [
        { rel: 'icon', type: 'image/svg+xml', sizes: 'any', href: '/brand/disaster-network-mark.svg' },
        { rel: 'manifest', href: '/manifest.webmanifest' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/brand/apple-touch-icon.png' }
      ],
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1, viewport-fit=cover' },
        { name: 'theme-color', content: '#ffffff' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-title', content: 'Disaster Network' },
        { name: 'apple-mobile-web-app-status-bar-style', content: 'default' },
        { name: 'description', content: 'Insurance claim coordination with optional AI voice and verified document collection. Synthetic evaluation scenario.' }
      ]
    }
  }
})
