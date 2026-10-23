<script setup>
const props = defineProps({ claims: Object, providers: Object, assignments: Object, proposed: Object, selected: Object, showPopup: Boolean, fit: Boolean, stats: Object, appearance: { type: String, default: 'workspace' }, openLabel: { type: String, default: 'Open claim' } })
const emit = defineEmits(['select', 'open', 'provider', 'dismiss-popup'])
const { t } = useAppI18n()
const canvas = ref(null)
const { loading, error, popupPosition, resetView, zoom, retry } = useOperationsMap(canvas, props, emit)
</script>

<template>
  <section class="map-region" :class="{ fit }" :aria-label="t('Reported households around Berre Lagoon')">
    <div class="map-surface">
      <div ref="canvas" class="map-canvas" />
      <div class="map-caption"><span />{{ t('Approximate locations · Household privacy') }}<small>{{ t('{count} shown on map', { count: claims.features.length }) }}</small></div>
      <div class="map-controls"><button :aria-label="t('Zoom in')" @click="zoom(1)"><UiIcon name="Plus" /></button><button :aria-label="t('Zoom out')" @click="zoom(-1)"><UiIcon name="Minus" /></button><button :aria-label="t('Reset map view')" @click="resetView"><UiIcon name="LocateFixed" :size="17" /></button></div>
      <div v-if="loading" class="map-message" role="status"><span class="loading-dot" />{{ t('Loading the basemap…') }}</div>
      <div v-if="error" class="map-error" role="status"><UiIcon name="MapPin" :size="22" /><p>{{ t(error) }}</p><UiButton class="small" @click="retry">{{ t('Retry map') }}</UiButton></div>
      <div v-if="selected && showPopup && popupPosition && !error" class="claim-popup" :style="{ left: `${popupPosition.left}px`, top: `${popupPosition.top}px` }">
        <div class="popup-name"><strong>{{ t(selected.name, selected.nameParameters) }}</strong><button :aria-label="t('Close map preview')" @click="$emit('dismiss-popup')"><UiIcon name="X" :size="13" /></button></div>
        <span>{{ selected.municipality }}</span><span>{{ selected.id }}</span>
        <span class="priority-label" :class="selected.priority"><i />{{ t(selected.priority === 'P1' ? 'P1 Urgent' : selected.priority === 'P2' ? 'P2 Priority' : 'P3 Standard') }}</span>
        <p><UiIcon name="House" :size="14" />{{ t(selected.situation) }}</p>
        <UiButton class="dark small" @click="$emit('open', selected.id)">{{ t(openLabel) }}<UiIcon name="ChevronRight" :size="14" /></UiButton>
      </div>
    </div>
    <footer class="map-legend">
      <div class="legend-group"><span class="legend-title">{{ t('Claim priority · Entire portfolio') }}</span><div class="priority-items"><span v-for="item in stats.priorities" :key="item.priority" class="legend-item"><i :class="item.priority" /><span>{{ t(item.priority === 'P1' ? 'P1 Urgent' : item.priority === 'P2' ? 'P2 Priority' : 'P3 Standard') }}<small>{{ t('{count} households', { count: item.count }) }}</small></span></span></div></div>
      <div class="legend-group provider-legend"><span class="legend-title">{{ t('Providers') }}</span><div class="legend-item"><UiIcon name="Truck" :size="21" /><span>{{ t('Drying teams') }}<small>{{ t('{count} available', { count: stats.dryingAvailable }) }}</small></span></div></div>
      <div class="accommodation-total"><UiIcon name="House" :size="22" /><span>{{ t('{count} households need accommodation', { count: stats.accommodation }) }}<small>{{ t('Households, not individual bookings') }}</small></span></div>
    </footer>
  </section>
</template>

<style src="maplibre-gl/dist/maplibre-gl.css"></style>
<style scoped>
.map-region { display: flex; flex-direction: column; flex: 1; min-width: 0; background: #f1f3f3; }
.map-region.fit { min-height: 0; }
.fit .map-surface { min-height: 0; }
.fit .map-legend { flex-shrink: 0; gap: 12px; padding: 10px 14px; min-height: 62px; }
.fit .priority-items { gap: 12px; }
.fit .provider-legend { padding-left: 12px; }
.fit .accommodation-total { max-width: 180px; }
.map-surface { position: relative; flex: 1; min-height: 560px; overflow: hidden; }
.map-canvas { position: absolute; inset: 0; width: 100%; height: 100%; }
.map-caption { position: absolute; left: 12px; top: 12px; display: flex; align-items: center; gap: 6px; padding: 7px 10px; background: #fffffff0; border: 1px solid #e5e5e8; border-radius: 6px; color: #8c8c96; font-size: 9px; pointer-events: none; }
.map-caption small { border-left: 1px solid #e4e4e8; padding-left: 6px; font-size: 9px; }.map-caption > span { width: 5px; height: 5px; border-radius: 50%; background: #a6a6af; }
.map-controls { position: absolute; top: 14px; right: 14px; display: flex; flex-direction: column; border: 1px solid #e4e4e7; border-radius: 8px; background: #fff; box-shadow: 0 3px 15px #0000000b; overflow: hidden; }
.map-controls button { display: flex; align-items: center; justify-content: center; width: 36px; height: 36px; }
.map-controls button + button { border-top: 1px solid #eeeeef; }
.map-controls button:hover { background: #f5f5f7; }
.map-message { position: absolute; left: 50%; top: 45%; transform: translate(-50%, -50%); display: flex; align-items: center; gap: 8px; background: #ffffffed; padding: 13px 17px; border: 1px solid var(--border); border-radius: 8px; font-size: 12px; color: #80808a; }
.loading-dot { width: 6px; height: 6px; border-radius: 50%; background: #8f8f9b; }
.map-error { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 15px; padding: 55px; color: #878792; background: #f1f3f1ed; text-align: center; }
.map-error p { font-size: 13px; line-height: 22px; max-width: 330px; }
.claim-popup { position: absolute; z-index: 2; display: flex; flex-direction: column; gap: 5px; width: 198px; padding: 13px; border-radius: 10px; background: #fff; border: 1px solid #e6e6e9; box-shadow: 0 6px 24px #00000020; }
.popup-name { display: flex; align-items: center; justify-content: space-between; }
.popup-name strong { font-size: 12px; font-weight: 600; }
.popup-name button { display: inline-flex; color: #95959f; padding: 0; }
.claim-popup > span { font-size: 10px; color: #9898a3; }
.priority-label { display: inline-flex; align-items: center; gap: 5px; font-size: 10px; }
.priority-label i { width: 7px; height: 7px; border-radius: 50%; background: currentColor; }
.claim-popup .P1 { color: #df474b; }.claim-popup .P2 { color: #bb8b24; }.claim-popup .P3 { color: #9296a0; }
.claim-popup p { display: flex; align-items: center; gap: 6px; padding-top: 8px; margin-top: 4px; border-top: 1px solid var(--border); color: #8c8c97; font-size: 10px; }
.claim-popup :deep(.button) { width: 100%; margin-top: 6px; justify-content: space-between; }
.map-legend { display: flex; align-items: center; flex-wrap: wrap; gap: 16px; padding: 15px 20px; background: #fff; border-top: 1px solid var(--border); min-height: 78px; }
.legend-group { display: flex; flex-direction: column; gap: 9px; }
.legend-title { font-size: 9px; font-weight: 600; }
.priority-items { display: flex; gap: 20px; }
.legend-item { display: flex; align-items: center; gap: 8px; font-size: 10px; }
.legend-item > i { width: 11px; height: 11px; border-radius: 50%; flex-shrink: 0; }
i.P1 { background: #ed4548; }i.P2 { background: #efaa24; }i.P3 { background: #9397a4; }
.legend-item small, .accommodation-total small { display: block; margin-top: 3px; font-size: 9px; color: #93939e; font-weight: 400; }
.provider-legend { padding-left: 18px; border-left: 1px solid var(--border); }
.provider-legend svg { color: #30a76c; }
.accommodation-total { display: flex; align-items: center; gap: 10px; font-size: 10px; font-weight: 500; margin-left: auto; max-width: 200px; }
.map-canvas :deep(.maplibregl-ctrl-attrib) { font: 9px -apple-system, BlinkMacSystemFont, sans-serif; }

@media (max-width: 1000px) {
  .map-region.fit { height: 100%; }
  .fit .map-surface { min-height: 250px; }
  .fit .map-legend { gap: 12px; padding: 10px 16px; }
  .fit .provider-legend { border-left: 0; padding-left: 0; }
  .fit .accommodation-total { margin-left: 0; max-width: none; }
}
@media (max-width: 760px) {
  .map-caption { left: 8px; top: 8px; max-width: calc(100% - 64px); flex-wrap: wrap; padding: 6px 8px; }
  .map-caption small { width: 100%; border-left: 0; padding-left: 11px; }
  .map-controls { top: 8px; right: 8px; }
  .map-controls button { width: 44px; height: 44px; }
  .fit .map-legend { gap: 8px 14px; }
  .fit .priority-items { gap: 10px; }
  .legend-group { gap: 6px; }
  .legend-item { gap: 5px; font-size: 9px; }
  .map-error { padding: 24px; }
  .map-message { width: max-content; max-width: 85%; font-size: 11px; }
}

</style>
