<script setup>
const { t } = useAppI18n()
</script>

<template>
  <footer>
    <div class="closing-band"><div class="closing-inner">
      <h2>{{ t('Make the next step clearer.') }}</h2>
      <p>{{ t('Bring your team from the first call to a coordinated catastrophe response.') }}</p>
      <div class="closing-actions"><NuxtLink to="/insurer" class="primary-link">{{ t('Explore insurer workspace') }}<UiIcon name="ArrowRight" :size="18" /></NuxtLink><NuxtLink to="/hotline" class="secondary-link">{{ t('Try Jules’s hotline') }}<UiIcon name="ArrowUpRight" :size="16" /></NuxtLink></div>
    </div></div>
    <div class="footer-bottom"><div><a href="#top" class="footer-brand"><img src="/brand/disaster-network-mark.svg" alt="" width="26" height="26" /><span>Disaster Network.</span></a><p>{{ t('Illustrative insurance scenarios. AI-generated images.') }}</p><p class="disclaimer">{{ t('Allstorm Insurance is fictional. No real claim decision, emergency service, booking or payment is provided by this workspace.') }}</p></div><a href="#top" class="back-link">{{ t('Back to top') }}<UiIcon name="ArrowUp" :size="16" /></a></div>
  </footer>
</template>

<style scoped>
footer { background: #fff; color: #0a0a0a; }
.closing-band { background: #f5f5f4; }
.closing-inner { max-width: 1264px; margin: 0 auto; padding: 88px 32px 86px; text-align: center; }
h2 { max-width: 800px; margin: 0 auto; font-size: 40px; line-height: 48px; font-weight: 400; letter-spacing: -1.2px; }
.closing-inner > p { max-width: 620px; margin: 24px auto 0; color: #737373; font-size: 17px; line-height: 28px; }
.closing-actions { display: flex; align-items: center; justify-content: center; gap: 28px; margin-top: 34px; }
.primary-link { display: inline-flex; align-items: center; justify-content: center; gap: 16px; min-height: 48px; padding: 12px 23px; border-radius: 28px; background: #171719; color: #fff; font-size: 14px; line-height: 24px; font-weight: 500; }
.primary-link:hover { background: #363638; }
.secondary-link { display: inline-flex; align-items: center; gap: 10px; font-size: 14px; line-height: 24px; font-weight: 450; }
.secondary-link:hover { color: #737373; }
.footer-bottom { display: flex; max-width: 1264px; margin: 0 auto; justify-content: space-between; align-items: flex-start; gap: 64px; padding: 37px 32px 30px; }
.footer-brand { display: inline-flex; align-items: center; gap: 9px; font-size: 18px; line-height: 26px; letter-spacing: -.7px; font-weight: 600; }
.footer-brand img { display: block; width: 26px; height: 26px; }
.footer-bottom p { margin-top: 12px; color: #929292; font-size: 11px; line-height: 19px; }
.footer-bottom .disclaimer { margin-top: 4px; max-width: 800px; color: #a0a0a0; font-size: 10px; line-height: 18px; }
.back-link { display: inline-flex; align-items: center; gap: 12px; padding-top: 3px; font-size: 12px; line-height: 20px; color: #737373; white-space: nowrap; }
.back-link:hover { color: #171719; }
</style>
