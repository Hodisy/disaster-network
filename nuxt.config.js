export default defineNuxtConfig({
  compatibilityDate: '2026-10-22',
  ssr: false,
  devtools: { enabled: false },
  css: ['~/assets/css/main.css'],
  routeRules: { '/': { redirect: '/hotline' } },
  app: {
    head: {
      htmlAttrs: { lang: 'fr' },
      title: 'Disaster Network — Votre sinistre, accompagné',
      meta: [{ name: 'description', content: 'Démonstration fictive de déclaration et de suivi de sinistre. Aucun appel ou envoi réel.' }]
    }
  }
})
