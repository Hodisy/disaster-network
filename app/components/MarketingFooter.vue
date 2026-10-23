<script setup>
const { t } = useAppI18n()
</script>

<template>
  <footer>
    <div class="closing-band"><div class="closing-inner">
      <h2>{{ t('Make the next step clearer.') }}</h2>
      <p>{{ t('Bring your team from the first call to a coordinated catastrophe response.') }}</p>
      <div class="closing-actions">
        <NuxtLink to="/hotline" class="primary-link">{{ t('Open hotline') }}</NuxtLink>
        <NuxtLink to="/insurer" class="secondary-link">{{ t('Open insurer workspace') }}</NuxtLink>
      </div>
    </div></div>
    <div class="footer-bottom"><div><a href="#top" class="footer-brand"><img src="/brand/disaster-network-mark.svg" alt="" width="26" height="26" /><span>Disaster Network.</span></a><p>{{ t('Claims intake, evidence review and coordinated assistance.') }}</p></div><a href="#top" class="back-link">{{ t('Back to top') }}</a></div>
  </footer>
</template>

<style scoped>
footer { background: #fff; color: #0a0a0a; }
.closing-band { background: #f5f5f4; }
.closing-inner { max-width: 1264px; margin: 0 auto; padding: 88px 32px 86px; text-align: center; }
h2 { max-width: 800px; margin: 0 auto; font-size: 40px; line-height: 48px; font-weight: 400; letter-spacing: -1.2px; }
.closing-inner > p { max-width: 620px; margin: 24px auto 0; color: #737373; font-size: 17px; line-height: 28px; }
.closing-actions { display: flex; align-items: center; justify-content: center; gap: 28px; margin-top: 34px; }
.primary-link { display: inline-flex; align-items: center; justify-content: center; min-height: 48px; padding: 12px 23px; border-radius: 28px; background: #171719; color: #fff; font-size: 14px; line-height: 24px; font-weight: 500; }
.primary-link:hover { background: #363638; }
.secondary-link { display: inline-flex; align-items: center; justify-content: center; min-height: 48px; padding: 11px 23px; border: 1px solid #d8d8d8; border-radius: 28px; background: #fff; color: #171719; font-size: 14px; line-height: 24px; font-weight: 500; }
.secondary-link:hover { background: #ededed; border-color: #b5b5bd; }
.footer-bottom { display: flex; max-width: 1264px; margin: 0 auto; justify-content: space-between; align-items: flex-start; gap: 64px; padding: 37px 32px 30px; }
.footer-brand { display: inline-flex; align-items: center; gap: 9px; font-size: 18px; line-height: 26px; letter-spacing: -.7px; font-weight: 600; }
.footer-brand img { display: block; width: 26px; height: 26px; }
.footer-bottom p { margin-top: 12px; color: #929292; font-size: 11px; line-height: 19px; }
.back-link { display: inline-flex; align-items: center; padding-top: 3px; font-size: 12px; line-height: 20px; color: #737373; white-space: nowrap; }
.back-link:hover { color: #171719; }
@media (max-width: 760px) {
  .closing-inner { padding: 48px 18px; }
  h2 { font-size: 30px; line-height: 36px; letter-spacing: -.8px; }
  .closing-inner > p { margin-top: 18px; font-size: 15px; line-height: 24px; }
  .closing-actions { flex-direction: column; align-items: stretch; gap: 12px; margin-top: 24px; }
  .primary-link, .secondary-link { white-space: normal; text-align: center; }
  .footer-bottom { flex-direction: column; gap: 20px; padding: 26px 18px; }
  .back-link { min-height: 44px; padding-top: 0; }
}
</style>
