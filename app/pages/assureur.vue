<script setup>
useHead({ title: 'Espace assureur — Disaster Network' })
</script>

<template>
  <div class="empty-page"><span class="empty-page-icon"><UiIcon name="Building2" :size="26" /></span><h1>Le côté assureur, bientôt.</h1><p>La priorité est au parcours client.<br>La vue opérationnelle sera construite dans un second temps.</p><UiButton to="/hotline" class="dark"><UiIcon name="ArrowLeft" :size="15" />Revenir au compte client</UiButton><small>Rivage Assurance · Compte fictif de démonstration</small></div>
</template>

<style scoped>
.empty-page { min-height: 700px; display: flex; align-items: center; justify-content: center; flex-direction: column; gap: 18px; padding: 45px 24px; text-align: center; }
.empty-page-icon { display: flex; align-items: center; justify-content: center; width: 64px; height: 64px; border: 1px solid var(--border); border-radius: 18px; color: #858590; margin-bottom: 6px; }
.empty-page h1 { font-size: 27px; letter-spacing: -.8px; line-height: 1.3; }
.empty-page p { color: #888893; font-size: 13px; line-height: 1.8; }
.empty-page small { color: #aaaab2; font-size: 10px; margin-top: 4px; }
</style>
