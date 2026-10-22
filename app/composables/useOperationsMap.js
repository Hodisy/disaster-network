// Rendering adapter only. Business data is supplied by the page as local GeoJSON.
// OpenFreeMap supplies the public basemap; it is not a live incident-data service.
export function useOperationsMap(container, props, emit) {
  const { t, locale } = useAppI18n()
  const loading = ref(true)
  const error = ref('')
  const popupPosition = ref(null)
  let map
  let observer
  let removed = false
  let generation = 0
  let ready = false
  const bounds = [[4.865, 43.29], [5.40, 43.645]]
  const selectedFilter = () => ['==', ['get', 'id'], props.selected?.id || '']

  function positionPopup() {
    if (!map || !props.selected || !props.showPopup) { popupPosition.value = null; return }
    const position = map.project(props.selected.coordinates)
    const width = container.value?.clientWidth || 800
    const height = container.value?.clientHeight || 600
    popupPosition.value = { left: Math.max(12, Math.min(width - 210, position.x + 15)), top: Math.max(14, Math.min(height - 240, position.y - 20)) }
  }
  function setSources() {
    if (!ready) return
    map.getSource('claims')?.setData(props.claims)
    map.getSource('providers')?.setData(props.providers)
    map.getSource('assignments')?.setData(props.assignments)
    map.getSource('proposed')?.setData(props.proposed)
    map.setFilter('selected-ring', selectedFilter())
    positionPopup()
  }
  function resetView() {
    if (!map) return
    map.fitBounds(bounds, { padding: 30, duration: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 550 })
  }
  function zoom(delta) { if (map) map.easeTo({ zoom: map.getZoom() + delta, duration: 200 }) }
  function localizeControls() {
    if (!map) return
    map.getCanvas().setAttribute('aria-label', t('Synthetic household claims around Berre Lagoon'))
    const attribution = container.value?.querySelector('.maplibregl-ctrl-attrib-button')
    if (attribution) {
      attribution.title = t('Map attribution')
      attribution.setAttribute('aria-label', t('Map attribution'))
    }
  }
  async function initialize() {
    const current = ++generation
    loading.value = true
    error.value = ''
    ready = false
    try {
      const [maplibre, worker] = await Promise.all([
        import('maplibre-gl'), import('maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url')
      ])
      if (removed || current !== generation || !container.value) return
      map?.remove()
      maplibre.setWorkerUrl(worker.default)
      map = new maplibre.Map({
        container: container.value, style: 'https://tiles.openfreemap.org/styles/positron',
        center: [5.04, 43.505], zoom: 10, attributionControl: false,
        dragRotate: false, pitchWithRotate: false, maxPitch: 0,
        canvasContextAttributes: { antialias: true },
        locale: { 'Map.Title': t('Synthetic household claims around Berre Lagoon'), 'AttributionControl.ToggleAttribution': t('Map attribution') }
      })
      map.addControl(new maplibre.AttributionControl({ compact: true }), 'bottom-right')
      localizeControls()
      map.on('load', () => {
        if (removed || current !== generation) return
        const colors = ['match', ['get', 'priority'], 'P1', '#ed4548', 'P2', '#efaa24', '#9397a4']
        map.addSource('claims', { type: 'geojson', data: props.claims, cluster: true, clusterRadius: 38, clusterMaxZoom: 13,
          clusterProperties: { urgent: ['+', ['case', ['==', ['get', 'priority'], 'P1'], 1, 0]], priority: ['+', ['case', ['==', ['get', 'priority'], 'P2'], 1, 0]] } })
        map.addLayer({ id: 'cluster-halo', type: 'circle', source: 'claims', filter: ['has', 'point_count'], paint: {
          'circle-radius': ['step', ['get', 'point_count'], 29, 10, 35, 20, 42], 'circle-opacity': 0.16,
          'circle-color': ['case', ['>', ['get', 'urgent'], 0], '#ed4548', ['>', ['get', 'priority'], 0], '#efaa24', '#9397a4']
        } })
        map.addLayer({ id: 'clusters', type: 'circle', source: 'claims', filter: ['has', 'point_count'], paint: {
          'circle-radius': ['step', ['get', 'point_count'], 16, 10, 20, 20, 24], 'circle-stroke-width': 2, 'circle-stroke-color': '#ffffff',
          'circle-color': ['case', ['>', ['get', 'urgent'], 0], '#ed4548', ['>', ['get', 'priority'], 0], '#efaa24', '#9397a4']
        } })
        map.addLayer({ id: 'cluster-labels', type: 'symbol', source: 'claims', filter: ['has', 'point_count'], layout: {
          'text-field': ['get', 'point_count_abbreviated'], 'text-size': 12, 'text-font': ['Noto Sans Regular'], 'text-allow-overlap': true
        }, paint: { 'text-color': '#ffffff' } })
        map.addLayer({ id: 'claim-points', type: 'circle', source: 'claims', filter: ['!', ['has', 'point_count']], paint: {
          'circle-radius': ['interpolate', ['linear'], ['zoom'], 9, 5, 15, 8], 'circle-color': colors, 'circle-stroke-color': '#ffffff', 'circle-stroke-width': 1.8
        } })
        // Separate unclustered source keeps a selected household visible at any zoom.
        map.addSource('selected', { type: 'geojson', data: props.claims })
        map.addLayer({ id: 'selected-ring', type: 'circle', source: 'selected', filter: selectedFilter(), paint: {
          'circle-radius': 11, 'circle-color': '#ffffff', 'circle-opacity': 0.2, 'circle-stroke-color': '#19191b', 'circle-stroke-width': 3
        } })
        map.addSource('assignments', { type: 'geojson', data: props.assignments })
        map.addLayer({ id: 'assignment-lines', type: 'line', source: 'assignments', paint: { 'line-color': ['case', ['get', 'disrupted'], '#dc4d50', '#656971'], 'line-width': 2, 'line-dasharray': [3, 3] } })
        map.addSource('proposed', { type: 'geojson', data: props.proposed })
        map.addLayer({ id: 'proposed-lines', type: 'line', source: 'proposed', paint: { 'line-color': '#28a967', 'line-width': 3, 'line-dasharray': [2, 2] } })
        map.addSource('providers', { type: 'geojson', data: props.providers })
        map.addLayer({ id: 'provider-points', type: 'circle', source: 'providers', paint: { 'circle-radius': 10, 'circle-color': ['match', ['get', 'status'], 'Unavailable', '#dd4249', 'Available', '#29ae69', '#22252b'], 'circle-stroke-width': 2, 'circle-stroke-color': '#ffffff' } })
        map.addLayer({ id: 'provider-labels', type: 'symbol', source: 'providers', layout: { 'text-field': '●', 'text-font': ['Noto Sans Regular'], 'text-size': 9, 'text-allow-overlap': true }, paint: { 'text-color': '#ffffff' } })
        map.on('click', 'clusters', async event => {
          const feature = event.features?.[0]
          if (!feature) return
          try {
            const nextZoom = await map.getSource('claims').getClusterExpansionZoom(feature.properties.cluster_id)
            map.easeTo({ center: feature.geometry.coordinates, zoom: nextZoom, duration: 350 })
          } catch { /* A source update can invalidate an old cluster while filters change. */ }
        })
        map.on('click', 'claim-points', event => { const id = event.features?.[0]?.properties.id; if (id) emit('select', id) })
        map.on('click', 'selected-ring', () => { if (props.selected) emit('select', props.selected.id) })
        map.on('click', 'provider-points', event => { const id = event.features?.[0]?.properties.id; if (id) emit('provider', id) })
        for (const layer of ['clusters', 'claim-points', 'selected-ring', 'provider-points']) {
          map.on('mouseenter', layer, () => { map.getCanvas().style.cursor = 'pointer' })
          map.on('mouseleave', layer, () => { map.getCanvas().style.cursor = '' })
        }
        ready = true
        loading.value = false
        error.value = ''
        resetView()
        positionPopup()
      })
      map.on('move', positionPopup)
      map.on('error', () => {
        if (removed || current !== generation) return
        if (!ready) { loading.value = false; error.value = 'The basemap could not be loaded. You can still use the claim list.' }
      })
    } catch {
      if (!removed && current === generation) { loading.value = false; error.value = 'The map is unavailable in this browser. You can still use the claim list.' }
    }
  }
  onMounted(() => {
    initialize()
    observer = new ResizeObserver(() => { map?.resize(); positionPopup() })
    observer.observe(container.value)
  })
  watch(() => [props.claims, props.providers, props.assignments, props.proposed], () => { setSources(); if (ready) map.getSource('selected')?.setData(props.claims) })
  watch(() => [props.selected?.id, props.showPopup], () => {
    if (ready) {
      map.setFilter('selected-ring', selectedFilter())
      if (props.selected) map.easeTo({ center: props.selected.coordinates, duration: 400 })
      positionPopup()
    }
  })
  watch(locale, () => { positionPopup(); localizeControls() })
  onBeforeUnmount(() => { removed = true; generation += 1; observer?.disconnect(); map?.remove() })
  return { loading, error, popupPosition, resetView, zoom, retry: initialize }
}
