<script setup>
const { t } = useAppI18n()
</script>

<template>
  <section class="hero" aria-labelledby="hero-title">
    <div class="hero-copy">
      <h1 id="hero-title">{{ t('Turn catastrophe claims into a coordinated response.') }}</h1>
      <p>{{ t('For insurance teams handling a surge of claims. Connect an AI hotline, policy context and field assistance in one shared workspace.') }}</p>
      <div class="hero-actions">
        <NuxtLink to="/hotline" class="primary-link">{{ t('Open hotline') }}</NuxtLink>
        <NuxtLink to="/insurer" class="text-link">{{ t('Open insurer workspace') }}</NuxtLink>
      </div>
    </div>
    <slot />
  </section>
</template>

<style scoped>
.hero { display: flex; align-items: center; justify-content: space-between; gap: 40px; width: 100%; max-width: 1264px; margin: 0 auto; padding: 40px 32px 48px; }
.hero-copy { flex: 0 0 380px; text-align: left; }
h1 { margin: 0; font-size: 46px; line-height: 50px; letter-spacing: -1.38px; font-weight: 400; color: #0a0a0a; }
.hero-copy > p { margin: 22px 0 0; font-size: 16px; line-height: 25px; font-weight: 400; color: #737373; }
.hero-actions { display: flex; flex-direction: column; align-items: flex-start; gap: 8px; margin-top: 26px; }
.primary-link { display: inline-flex; align-items: center; justify-content: center; min-height: 48px; padding: 12px 23px; background: #171719; color: #fff; border-radius: 28px; font-size: 14px; line-height: 24px; font-weight: 500; }
.primary-link:hover { background: #363638; }
.text-link { padding: 10px 0; font-size: 14px; line-height: 24px; font-weight: 500; }
.text-link:hover { color: #737373; }
@media (max-width: 1000px) {
  .hero { gap: 28px; padding: 36px 24px 40px; }
  .hero-copy { flex-basis: 340px; }
  h1 { font-size: 40px; line-height: 44px; }
}
@media (max-width: 760px) {
  .hero { flex-direction: column; align-items: stretch; gap: 28px; padding: 32px 18px 24px; }
  .hero-copy { flex: none; }
  h1 { font-size: 35px; line-height: 39px; letter-spacing: -1px; }
  .hero-copy > p { margin-top: 18px; font-size: 15px; line-height: 24px; }
  .hero-actions { align-items: stretch; gap: 6px; margin-top: 22px; }
  .primary-link, .text-link { justify-content: center; text-align: center; min-height: 44px; white-space: normal; }
}
</style>
