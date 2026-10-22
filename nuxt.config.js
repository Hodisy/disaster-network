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
    openaiLiveVoice: 'marin'
  },
  css: ['~/assets/css/main.css'],
  routeRules: { '/': { redirect: '/hotline' } },
  app: {
    head: {
      htmlAttrs: { lang: 'en' },
      title: 'Disaster Network — Support for your insurance claim',
      meta: [{ name: 'description', content: 'Fictional insurance demo. Live AI voice is optional; files stay in your browser.' }]
    }
  }
})
