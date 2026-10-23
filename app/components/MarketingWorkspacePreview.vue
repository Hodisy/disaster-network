<script setup>
defineProps({ image: { type: Object, required: true } })
const { t } = useAppI18n()
</script>

<template>
  <figure id="operations" class="workspace-preview">
    <NuxtLink to="/insurer" class="workspace-link" :aria-label="t('Explore insurer workspace')">
      <div class="workspace-art" role="img" :aria-label="t(image.alt)">
        <div class="map-crop" aria-hidden="true">
          <img :src="image.src" alt="" width="1536" height="1024" fetchpriority="high" decoding="async" />
        </div>
      </div>
    </NuxtLink>
    <figcaption>{{ t(image.caption) }}<NuxtLink to="/insurer">{{ t('Explore the workspace') }}</NuxtLink></figcaption>
  </figure>
</template>

<style scoped>
.workspace-preview { flex: 1; max-width: 620px; min-width: 0; margin: 0; scroll-margin-top: 88px; }
.workspace-link { display: block; padding: 8px; background: #f3f3f2; border: 1px solid #e5e5e3; border-radius: 14px; }
.workspace-link:focus-visible { outline: 2px solid #171719; outline-offset: 4px; }
.workspace-art { overflow: hidden; background: #fff; border: 1px solid #e5e5e3; border-radius: 7px; }
/* Retain only the original map and legend: x=272–1160, y=216–1024. */
.map-crop { position: relative; width: 100%; aspect-ratio: 888 / 808; overflow: hidden; }
.map-crop img { position: absolute; left: -30.630631%; top: -26.732673%; display: block; width: 172.972973%; height: auto; max-width: none; }
figcaption { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 12px 2px 0; font-size: 11px; line-height: 18px; color: #737373; }
figcaption a { display: inline-flex; align-items: center; color: #363638; }
figcaption a:hover { color: #737373; }
@media (max-width: 760px) {
  .workspace-preview { flex: none; align-self: center; width: 100%; scroll-margin-top: 80px; }
  .workspace-link { padding: 5px; border-radius: 10px; }
  figcaption { flex-wrap: wrap; gap: 4px 12px; padding-top: 9px; font-size: 10px; line-height: 16px; }
}
</style>
