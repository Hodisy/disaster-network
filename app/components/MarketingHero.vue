<script setup>
const { t } = useAppI18n()
</script>

<template>
  <section class="hero" aria-labelledby="hero-title">
    <div class="hero-copy">
      <h1 id="hero-title">{{ t('Turn catastrophe claims into a coordinated response.') }}</h1>
      <p>{{ t('For insurance teams handling a surge of claims. Connect an AI hotline, policy context and field assistance in one shared workspace.') }}</p>
      <div class="hero-actions">
        <NuxtLink to="/insurer" class="primary-link">{{ t('Explore insurer workspace') }}<UiIcon name="ArrowRight" :size="18" /></NuxtLink>
        <a href="#jules" class="text-link">{{ t('See Jules’s journey') }}</a>
      </div>
    </div>
    <slot />
  </section>
</template>

<style scoped>
.hero { width: 100%; }
.hero-copy { max-width: 1000px; margin: 0 auto; padding: 64px 32px 56px; text-align: center; }
h1 { max-width: 940px; margin: 0 auto; font-size: 64px; line-height: 64px; letter-spacing: -1.92px; font-weight: 400; color: #0a0a0a; }
.hero-copy > p { max-width: 690px; margin: 28px auto 0; font-size: 18px; line-height: 28px; font-weight: 400; color: #737373; }
.hero-actions { display: flex; align-items: center; justify-content: center; gap: 28px; margin-top: 30px; }
.primary-link { display: inline-flex; align-items: center; justify-content: center; gap: 16px; min-height: 48px; padding: 12px 23px; background: #171719; color: #fff; border-radius: 28px; font-size: 14px; line-height: 24px; font-weight: 500; }
.primary-link:hover { background: #363638; }
.text-link { padding: 10px 0; font-size: 14px; line-height: 24px; font-weight: 500; }
.text-link:hover { color: #737373; }
</style>
