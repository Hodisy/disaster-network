import { DEMO_CLAIM_REFERENCE } from '~/config/demo.js'

// Read the existing synthetic portfolio; interactions change only this preview.
export function useMarketingWorkspacePreview() {
  const workspace = useInsurerWorkspace()
  const selectedId = ref(DEMO_CLAIM_REFERENCE)
  const panel = ref('priority')
  const advanced = ref(false)
  const selectedProvider = ref('')
  const selected = computed(() => workspace.filtered.value.find(claim => claim.id === selectedId.value) || null)
  const preview = computed(() => ({
    stats: workspace.stats.value,
    queue: workspace.queue.value,
    queueTotal: workspace.queue.value.length,
    records: panel.value === 'priority' ? workspace.queue.value : workspace.pageClaims.value,
    total: panel.value === 'priority' ? workspace.queue.value.length : workspace.filtered.value.length,
    panel: panel.value,
    page: workspace.page.value,
    pageCount: workspace.pageCount.value,
    selected: selected.value,
    view: workspace.view.value,
    query: { ...workspace.query },
    options: workspace.options.value,
    advanced: advanced.value,
    resources: workspace.providers.value,
    missions: workspace.missions.value,
    selectedProvider: selectedProvider.value,
    claims: workspace.claimGeojson.value,
    providers: workspace.providerGeojson.value,
    assignments: workspace.assignmentGeojson.value,
    proposed: workspace.proposedGeojson.value
  }))

  function selectPreview(id) { selectedId.value = id }
  function clearPreview() { selectedId.value = null }
  function setPreviewPanel(value) { panel.value = value; selectedId.value = null }
  function setPreviewPage(value) { workspace.page.value = value }
  function setPreviewView(value) {
    if (['situation', 'mobilization'].includes(value)) workspace.view.value = value
  }
  function filterPreview(key, value) { workspace.updateFilter(key, value) }
  function clearPreviewFilters() { workspace.clearFilters() }
  function togglePreviewFilters() { advanced.value = !advanced.value }
  function selectPreviewProvider(id) {
    selectedProvider.value = id
    if (id) workspace.view.value = 'mobilization'
  }

  return { preview, selectPreview, clearPreview, setPreviewPanel, setPreviewPage, setPreviewView, filterPreview, clearPreviewFilters, togglePreviewFilters, selectPreviewProvider }
}
