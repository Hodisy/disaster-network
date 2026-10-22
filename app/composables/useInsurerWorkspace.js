import { DEMO_CONTRACT_NUMBER, DEMO_PRODUCT_NAME } from '../config/demo.js'
import { claimStages, createReferenceClaim, demoBenefits } from './useClaimScenario.js'

const clusters = [
  { label: 'Entressen', municipality: 'Istres', count: 20, coordinates: [4.939, 43.599] },
  { label: 'Les Heures Claires', municipality: 'Istres', count: 10, coordinates: [4.984, 43.491] },
  { label: 'Other Istres locations', municipality: 'Istres', count: 15, coordinates: [4.977, 43.519] },
  { label: 'Saint-Chamas', municipality: 'Saint-Chamas', count: 15, coordinates: [5.040, 43.553] },
  { label: 'Saint-Mitre-les-Remparts', municipality: 'Saint-Mitre-les-Remparts', count: 10, coordinates: [5.013, 43.456] },
  { label: 'Martigues area', municipality: 'Martigues', count: 10, coordinates: [5.052, 43.416] }
]
const groupLabels = { A: 'Home uninhabitable', B: 'Evacuated or inaccessible', C: 'Damaged but habitable', D: 'Limited or unconfirmed' }
const needLabels = ['Temporary accommodation', 'Electrical safety', 'Cleaning and drying', 'Moving and storage', 'Damage assessment']

const policyholderNames = ['Camille', 'Sam', 'Alice', 'Hugo', 'Nora', 'Louis', 'Lina', 'Gabriel', 'Inès', 'Arthur', 'Sarah', 'Adam', 'Emma', 'Raphaël', 'Chloé', 'Noah', 'Jade', 'Paul', 'Louise', 'Nathan', 'Manon', 'Lucas', 'Élise', 'Tom', 'Léa', 'Victor', 'Anna', 'Sacha', 'Jeanne', 'Robin', 'Clara', 'Baptiste', 'Mila', 'Mathis', 'Agathe', 'Antoine', 'Lucie', 'Maxime', 'Rose', 'Valentin', 'Zoé', 'Alexandre', 'Eva', 'Simon', 'Iris', 'Théo', 'Margaux', 'Nicolas', 'Salomé', 'Thomas', 'Amina', 'Romain', 'Yasmine', 'Julien', 'Sofia', 'Benoît', 'Maya', 'David', 'Nina', 'Pierre', 'Émilie', 'Adrien', 'Anaïs', 'Karim', 'Marion', 'Olivier', 'Céline', 'Mehdi', 'Julie', 'Laurent', 'Amélie', 'Marc', 'Sophie', 'Florian', 'Élodie', 'Guillaume', 'Aurore', 'Sébastien', 'Mathilde']
const situationNotes = {
  A: [
    'Water entered the ground-floor living areas. Kitchen units and several appliances are damaged. The household is staying with relatives while accommodation and access are reviewed.',
    'Floodwater affected the ground floor, including flooring and furniture. The household cannot stay in the property. Electrical inspection and drying are needed once access is confirmed.',
    'The living room and bedrooms on the ground floor were affected. Mattresses and low-level furniture are damaged. A temporary accommodation request is being assessed.',
    'Water remains in part of the property. The household has left the home and is seeking a stable place to stay. Damage assessment and safe access are still pending.'
  ],
  B: [
    'The household left the area after an evacuation instruction. The home remains inaccessible and internal damage has not been established.',
    'Access to the residence is restricted. The household is staying away from the property; assistance needs and the next contact date are being reviewed.',
    'The household cannot reach the residence. No inspection has taken place, and no building or contents loss is assumed at this stage.'
  ],
  C: [
    'Water entered the garage and damaged stored belongings. The living areas remain usable. An inventory and photographs are being collected.',
    'Floodwater affected a ground-floor utility room. The household remains at home. Cleaning and drying requirements are being assessed.',
    'Water damaged flooring and furniture in one room. The remaining living areas are usable. Photos and a repair assessment are requested.',
    'A small amount of water entered the entrance hall. The household remains in the property. Skirting boards and flooring require assessment.'
  ],
  D: [
    'Water was reported near the property, with no internal damage confirmed. A follow-up call is needed to clarify the situation.',
    'The household reported minor water ingress at an outbuilding. The extent of damage and any assistance needs remain to be confirmed.',
    'The initial contact did not establish property damage. The household has been asked to share any available photographs without entering a restricted area.'
  ]
}

function createPortfolio() {
  const reference = { ...createReferenceClaim(), opened: false, chronology: [] }
  const places = clusters.flatMap(cluster => Array.from({ length: cluster.count }, () => cluster))
  places.splice(places.findIndex(place => place.label === reference.cluster), 1)
  // Keep the two additional PostgreSQL fixture households in the same portfolio records.
  const evacuated = places.splice(places.findIndex(place => place.label === 'Les Heures Claires'), 1)[0]
  places.splice(1, 0, evacuated)
  const sizes = [...Array(30).fill(1), ...Array(25).fill(2), ...Array(9).fill(3), ...Array(15).fill(4)]
  sizes[1] = 2
  sizes[30] = 1
  let otherIndex = 0
  return [reference, ...places.map((place, index) => {
    const ordinal = index + 2
    const isEvacuated = place.label === 'Les Heures Claires'
    const order = isEvacuated ? 0 : ++otherIndex
    const group = isEvacuated ? 'B' : index === 0 ? 'C' : order <= 15 ? 'A' : order <= 49 ? 'C' : 'D'
    const accommodationNeeded = group === 'A' ? order <= 9 : group === 'B' ? index % 3 === 0 : false
    const closed = group === 'D' && index % 4 === 0
    const stage = closed ? 4 : group === 'C' ? index % 3 : 0
    const needs = group === 'A' ? ['Temporary accommodation', 'Electrical safety', 'Cleaning and drying', 'Damage assessment']
      : group === 'B' ? ['Temporary accommodation', 'Damage assessment']
        : group === 'C' ? ['Cleaning and drying', 'Damage assessment'] : closed ? [] : ['Damage assessment']
    if ((group === 'A' || group === 'C') && index % 3 === 1) needs.push('Moving and storage')
    if (group === 'C' && index % 4 === 2) needs.push('Electrical safety')
    const moneyLines = demoBenefits.map((benefit, benefitIndex) => {
      let amount = null
      if (closed) amount = 0
      else if (benefit.id === 'accommodation') amount = accommodationNeeded ? [750, 900, 625, 1000, 525][index % 5] : 0
      else if (group === 'A' || group === 'C') {
        const relevant = benefit.id === 'electrical' ? needs.includes('Electrical safety') : benefit.id === 'storage' ? needs.includes('Moving and storage') : true
        const severity = (group === 'A' ? 0.65 : 0.12) + ((index * 7 + benefitIndex * 3) % 13) / (group === 'A' ? 10 : 24)
        amount = relevant ? Math.min(benefit.limit, Math.round(benefit.amount * severity / 25) * 25) : 0
      }
      return { ...benefit, amount, approved: 0, paid: 0, review: 'Not reviewed' }
    })
    const hour = String(11 - index % 10).padStart(2, '0')
    const minute = String(index * 17 % 60).padStart(2, '0')
    const reportedAt = index < 2 ? '2026-10-22T08:15:00+02:00' : `2026-10-22T${hour}:${minute}:00+02:00`
    const name = policyholderNames[index]
    return {
      id: `CAT-2026-${String(ordinal).padStart(3, '0')}`, name, nameParameters: {}, initials: name.slice(0, 1),
      contractNumber: String(Number(DEMO_CONTRACT_NUMBER) + ordinal - 1), productName: DEMO_PRODUCT_NAME,
      householdSize: sizes[index], municipality: place.municipality, cluster: place.label,
      // Approximate scenario locations; no actual resident or damaged property is identified.
      coordinates: [place.coordinates[0] + (((index * 7) % 13) - 6) * 0.00065, place.coordinates[1] + (((index * 11) % 13) - 6) * 0.0005],
      address: `${place.label}, ${place.municipality}`, addressNote: 'Approximate household location.', group, situation: groupLabels[group],
      priority: index === 0 ? 'P3' : accommodationNeeded ? 'P1' : group === 'A' || group === 'B' || (group === 'C' && index % 3 === 0) ? 'P2' : 'P3',
      priorityReason: accommodationNeeded ? 'No lasting accommodation solution is available.' : group === 'B' ? 'Access is restricted and the damage is not established.' : group === 'A' ? 'The home is uninhabitable; the household has a temporary place to stay.' : 'No immediate accommodation need has been reported.',
      accommodationNeeded, accommodationPlaced: false,
      currentAccommodation: accommodationNeeded ? 'Temporary arrangement needs replacement.' : group === 'A' || group === 'B' ? 'Staying with relatives; no immediate placement requested.' : 'At home.',
      access: index === 0 ? 'Unknown' : group === 'B' ? 'Restricted' : group === 'C' || closed ? 'Permitted' : index % 4 === 0 ? 'Restricted' : 'Unknown',
      coverage: closed ? 'Reviewed' : stage >= 2 ? 'Proposal' : 'Not reviewed', catNat: 'Not verified', opened: true,
      waitingHours: Math.floor((Date.parse('2026-10-22T12:00:00+02:00') - Date.parse(reportedAt)) / 3600000),
      event: 'Flooding', eventDate: '2026-10-20', date: '2026-10-20', reportedAt, stage, occupancy: 'Owner-occupier', needs,
      summary: closed ? 'The household reported no property damage or outstanding assistance need. The file was closed after contact.' : situationNotes[group][index % situationNotes[group].length],
      moneyLines,
      chronology: [{ label: 'Report received', at: reportedAt }, ...(closed ? [{ label: 'Claim closed without a payment.', at: '2026-10-22T11:45:00+02:00' }] : [])]
    }
  })]
}

function createProviders() {
  const drying = Array.from({ length: 6 }, (_, index) => ({
    id: `dry-${index + 1}`, name: ['Delta Drying', 'Berre Drying', 'Provence Drying', 'Riviera Drying', 'West Lagoon Drying', 'Istres Drying'][index],
    skill: 'Cleaning and drying', capacity: 2, unit: 'Concurrent sites', available: true,
    coordinates: [[4.95, 43.58], [5.035, 43.553], [5.008, 43.455], [5.049, 43.432], [4.975, 43.545], [5.057, 43.414]][index],
    delayHours: index < 3 ? 0 : 2 + index - 3, costDelta: index < 3 ? 0 : 90 + (index - 3) * 25
  }))
  return [...drying,
    { id: 'electrical-1', name: 'Lagoon Electrical', skill: 'Electrical safety', capacity: 4, unit: 'Visits per day', available: true, coordinates: [4.970, 43.530] },
    { id: 'assessment-1', name: 'Provence Assessment', skill: 'Damage assessment', capacity: 6, unit: 'Visits per day', available: true, coordinates: [5.036, 43.468] },
    { id: 'storage-1', name: 'Berre Storage', skill: 'Moving and storage', capacity: 3, unit: 'Household storage slots', available: true, coordinates: [5.014, 43.572] },
    { id: 'accommodation-1', name: 'Lagoon Family Accommodation', skill: 'Temporary accommodation', capacity: 8, unit: 'Household accommodation slots', available: true, coordinates: [5.046, 43.535] }
  ]
}

function createMissions(portfolio) {
  const accessible = portfolio.filter(claim => claim.access === 'Permitted' && claim.needs.includes('Cleaning and drying') && claim.stage < 4).slice(0, 6)
  return [...accessible.map((claim, index) => ({
    id: `mission-${index + 1}`, claimId: claim.id, providerId: `dry-${1 + Math.floor(index / 2)}`, skill: 'Cleaning and drying',
    status: 'Scheduled', start: '2026-10-23T09:00:00+02:00', end: '2026-10-23T17:00:00+02:00', estimate: 540
  })),
  { id: 'jules-drying', claimId: 'CAT-2026-001', providerId: null, skill: 'Cleaning and drying', status: 'Blocked', start: null, end: null, estimate: 1800 },
  { id: 'jules-electrical', claimId: 'CAT-2026-001', providerId: null, skill: 'Electrical safety', status: 'Blocked', start: null, end: null, estimate: 350 },
  { id: 'jules-assessment', claimId: 'CAT-2026-001', providerId: null, skill: 'Damage assessment', status: 'Blocked', start: null, end: null, estimate: null }]
}

const collection = features => ({ type: 'FeatureCollection', features })

export function useInsurerWorkspace() {
  const { state: client } = useDemo()
  const { t, locale } = useAppI18n()
  const state = useState('insurer-workspace', () => {
    const portfolio = createPortfolio()
    return { portfolio, providers: createProviders(), missions: createMissions(portfolio), overrides: {}, outage: null, planApplied: false, audit: [], pending: false, error: '', epoch: client.value.revision }
  })
  const view = ref('situation')
  const panel = ref('priority')
  const detailOpen = ref(false)
  const selectedId = ref('CAT-2026-001')
  const page = ref(1)
  const query = reactive({ search: '', municipality: '', priority: '', need: '', stage: '', group: '', access: '' })

  watch(() => client.value.revision, revision => {
    if (state.value.epoch === revision) return
    const portfolio = createPortfolio()
    state.value = { portfolio, providers: createProviders(), missions: createMissions(portfolio), overrides: {}, outage: null, planApplied: false, audit: [], pending: false, error: '', epoch: revision }
    selectedId.value = 'CAT-2026-001'
    detailOpen.value = false
  }, { immediate: true })

  const claims = computed(() => state.value.portfolio.map(record => {
    const shared = record.id === 'CAT-2026-001' && client.value.claim ? client.value.claim : null
    const claim = { ...record, ...shared, ...(state.value.overrides[record.id] || {}) }
    const money = claim.moneyLines
    return {
      ...claim, stageLabel: claimStages[claim.stage],
      priority: claim.accommodationPlaced && claim.priority === 'P1' ? 'P2' : claim.priority,
      priorityReason: claim.accommodationPlaced ? 'Accommodation is arranged; damage and site access still require review.' : claim.priorityReason,
      grossEstimate: money.reduce((sum, line) => sum + (line.amount ?? 0), 0),
      estimateIncomplete: money.some(line => line.amount === null),
      approvedAmount: money.reduce((sum, line) => sum + line.approved, 0), paidAmount: money.reduce((sum, line) => sum + line.paid, 0),
      unmetAccommodation: claim.accommodationNeeded && !claim.accommodationPlaced && claim.stage < 4,
      displayName: t(claim.name, claim.nameParameters)
    }
  }))
  const selected = computed(() => claims.value.find(claim => claim.id === selectedId.value) || null)
  const filtered = computed(() => {
    const search = query.search.trim().toLocaleLowerCase()
    return claims.value.filter(claim => (!search || [claim.id, claim.displayName, claim.municipality, claim.cluster, t(claim.cluster)].join(' ').toLocaleLowerCase().includes(search))
      && (!query.municipality || claim.municipality === query.municipality)
      && (!query.priority || claim.priority === query.priority)
      && (!query.need || claim.needs.includes(query.need))
      && (!query.stage || String(claim.stage) === query.stage)
      && (!query.group || claim.group === query.group)
      && (!query.access || claim.access === query.access))
      .sort((a, b) => a.priority.localeCompare(b.priority) || b.waitingHours - a.waitingHours || a.id.localeCompare(b.id))
  })
  const queue = computed(() => filtered.value.filter(claim => claim.stage < 4 && (claim.unmetAccommodation || claim.priority === 'P1')))
  const pageCount = computed(() => Math.max(1, Math.ceil(filtered.value.length / 8)))
  const pageClaims = computed(() => filtered.value.slice((page.value - 1) * 8, page.value * 8))
  watch([() => ({ ...query }), locale], () => { page.value = 1 }, { deep: true })
  watch(pageCount, count => { page.value = Math.min(page.value, count) })
  watch(filtered, records => { if (selectedId.value && !records.some(claim => claim.id === selectedId.value)) { selectedId.value = null; detailOpen.value = false } })

  const missions = computed(() => state.value.missions.map(mission => {
    const claim = claims.value.find(item => item.id === mission.claimId)
    const provider = state.value.providers.find(item => item.id === mission.providerId)
    const status = claim.access !== 'Permitted' && mission.skill !== 'Temporary accommodation' ? 'Blocked' : mission.status === 'Blocked' ? 'Proposed' : mission.status
    return { ...mission, status, claim, provider }
  }))
  const providers = computed(() => state.value.providers.map(provider => {
    const assigned = missions.value.filter(mission => mission.providerId === provider.id && mission.status === 'Scheduled').length
    const placements = provider.skill === 'Temporary accommodation' ? claims.value.filter(claim => claim.accommodationPlaced).length : 0
    const used = assigned + placements
    return { ...provider, used, remaining: Math.max(0, provider.capacity - used), status: !provider.available ? 'Unavailable' : used ? 'Active' : 'Available' }
  }))
  const selectedMissions = computed(() => missions.value.filter(mission => mission.claimId === selectedId.value))
  const stats = computed(() => ({
    households: claims.value.length, people: claims.value.reduce((sum, claim) => sum + claim.householdSize, 0),
    opened: claims.value.filter(claim => claim.opened).length,
    priorities: ['P1', 'P2', 'P3'].map(priority => ({ priority, count: claims.value.filter(claim => claim.priority === priority).length })),
    groups: Object.entries(groupLabels).map(([id, label]) => ({ id, label, count: claims.value.filter(claim => claim.group === id).length })),
    accommodation: claims.value.filter(claim => claim.unmetAccommodation).length,
    gross: claims.value.reduce((sum, claim) => sum + claim.grossEstimate, 0),
    incomplete: claims.value.filter(claim => claim.estimateIncomplete).length,
    approved: claims.value.reduce((sum, claim) => sum + claim.approvedAmount, 0), paid: claims.value.reduce((sum, claim) => sum + claim.paidAmount, 0),
    dryingAvailable: providers.value.filter(provider => provider.skill === 'Cleaning and drying' && provider.available && provider.remaining > 0).length,
    blocked: missions.value.filter(mission => mission.status === 'Blocked').length
  }))
  const options = computed(() => ({
    municipalities: [...new Set(claims.value.map(claim => claim.municipality))],
    priorities: ['P1', 'P2', 'P3'], needs: needLabels, stages: claimStages,
    groups: Object.entries(groupLabels).map(([id, label]) => ({ id, label })), access: ['Unknown', 'Restricted', 'Permitted']
  }))
  const claimGeojson = computed(() => collection(filtered.value.map(claim => ({ type: 'Feature', geometry: { type: 'Point', coordinates: claim.coordinates }, properties: { id: claim.id, priority: claim.priority, group: claim.group, opened: claim.opened, selected: claim.id === selectedId.value } }))))
  const providerGeojson = computed(() => collection(providers.value.map(provider => ({ type: 'Feature', geometry: { type: 'Point', coordinates: provider.coordinates }, properties: { id: provider.id, name: provider.name, status: provider.status } }))))
  const assignmentGeojson = computed(() => collection(view.value === 'mobilization' ? missions.value.filter(mission => mission.provider && ['Scheduled', 'Disrupted'].includes(mission.status)).map(mission => ({ type: 'Feature', geometry: { type: 'LineString', coordinates: [mission.provider.coordinates, mission.claim.coordinates] }, properties: { disrupted: mission.status === 'Disrupted' } })) : []))

  const replan = computed(() => {
    if (!state.value.outage || state.value.planApplied) return null
    const affected = missions.value.filter(mission => mission.providerId === state.value.outage && mission.status === 'Disrupted')
    const reservations = new Map()
    const changes = affected.map(mission => {
      const replacement = providers.value.filter(provider => provider.id !== state.value.outage && provider.available && provider.skill === mission.skill
        && provider.remaining - (reservations.get(provider.id) || 0) > 0 && mission.claim.access === 'Permitted')
        .sort((a, b) => a.delayHours - b.delayHours || a.costDelta - b.costDelta)[0]
      if (!replacement) return { mission, replacement: null }
      reservations.set(replacement.id, (reservations.get(replacement.id) || 0) + 1)
      return { mission, replacement, delayHours: replacement.delayHours, costDelta: replacement.costDelta,
        start: new Date(new Date(mission.start).getTime() + replacement.delayHours * 3600000).toISOString(),
        end: new Date(new Date(mission.end).getTime() + replacement.delayHours * 3600000).toISOString() }
    })
    return { provider: providers.value.find(provider => provider.id === state.value.outage), changes,
      canApply: changes.length > 0 && changes.every(change => change.replacement),
      maxDelay: Math.max(0, ...changes.map(change => change.delayHours || 0)), costDelta: changes.reduce((sum, change) => sum + (change.costDelta || 0), 0) }
  })
  const proposedGeojson = computed(() => collection(replan.value ? replan.value.changes.filter(change => change.replacement).map(change => ({ type: 'Feature', geometry: { type: 'LineString', coordinates: [change.replacement.coordinates, change.mission.claim.coordinates] }, properties: {} })) : []))

  function selectClaim(id, open = false) { selectedId.value = id; detailOpen.value = open; state.value.error = '' }
  function updateFilter(key, value) { if (Object.hasOwn(query, key)) query[key] = value }
  function clearFilters() { Object.keys(query).forEach(key => { query[key] = '' }) }
  function log(label, claimId) { state.value.audit.push({ label, claimId, at: new Date().toISOString() }) }
  async function simulate(action) {
    if (state.value.pending) return
    const epoch = client.value.revision
    state.value.pending = true
    state.value.error = ''
    try {
      await new Promise(resolve => setTimeout(resolve, 550))
      if (epoch !== client.value.revision) return
      action()
    } catch { state.value.error = 'The update could not be completed. Please try again.' }
    finally { if (epoch === client.value.revision) state.value.pending = false }
  }
  function simulateOutage() {
    view.value = 'mobilization'
    detailOpen.value = false
    return simulate(() => {
      if (state.value.outage) return
      const provider = state.value.providers.find(item => item.id === 'dry-1')
      provider.available = false
      state.value.outage = provider.id
      state.value.missions.forEach(mission => { if (mission.providerId === provider.id && mission.status === 'Scheduled') mission.status = 'Disrupted' })
      log('Provider availability changed')
    })
  }
  function applyPlan() {
    if (!replan.value?.canApply) return
    return simulate(() => {
      const plan = replan.value
      if (!plan?.canApply) return
      plan.changes.forEach(change => {
        const mission = state.value.missions.find(item => item.id === change.mission.id)
        Object.assign(mission, { providerId: change.replacement.id, start: change.start, end: change.end, status: 'Scheduled', estimate: mission.estimate + change.costDelta })
        log('Mission reassigned', mission.claimId)
      })
      state.value.planApplied = true
    })
  }
  function cancelOutage() {
    return simulate(() => {
      state.value.providers = createProviders()
      state.value.missions = createMissions(state.value.portfolio)
      state.value.outage = null
      state.value.planApplied = false
      log('Resource plan restored')
    })
  }
  function permitAccess() {
    if (!selected.value || selected.value.access !== 'Unknown') return
    const id = selected.value.id
    return simulate(() => {
      state.value.overrides[id] = { ...state.value.overrides[id], access: 'Permitted' }
      log('Access permission previewed', id)
    })
  }
  function scheduleSelected() {
    if (!selected.value || selected.value.access !== 'Permitted' || selected.value.stage === 4) return
    const id = selected.value.id
    return simulate(() => {
      const proposed = missions.value.filter(mission => mission.claimId === id && mission.status === 'Proposed')
      let unscheduled = 0
      for (const mission of proposed) {
        const provider = providers.value.find(item => item.skill === mission.skill && item.available && item.remaining > 0)
        if (!provider) { unscheduled += 1; continue }
        const source = state.value.missions.find(item => item.id === mission.id)
        Object.assign(source, { providerId: provider.id, status: 'Scheduled', start: '2026-10-23T09:00:00+02:00', end: mission.skill === 'Cleaning and drying' ? '2026-10-23T17:00:00+02:00' : '2026-10-23T10:00:00+02:00' })
        log('Accessible mission scheduled', id)
      }
      if (unscheduled) state.value.error = 'Some proposed missions have no qualified capacity available. They remain unassigned.'
    })
  }
  function placeAccommodation() {
    if (!selected.value?.unmetAccommodation) return
    const id = selected.value.id
    return simulate(() => {
      const provider = providers.value.find(item => item.skill === 'Temporary accommodation' && item.available && item.remaining > 0)
      if (!provider) { state.value.error = 'No household accommodation slot is currently available.'; return }
      state.value.overrides[id] = { ...state.value.overrides[id], accommodationPlaced: true }
      log('Household accommodation placement previewed', id)
    })
  }
  function planAssistance() {
    view.value = 'mobilization'
    detailOpen.value = true
    const claim = selected.value
    if (!claim || claim.stage === 4) return
    const benefitIds = { 'Electrical safety': 'electrical', 'Cleaning and drying': 'drying', 'Moving and storage': 'storage' }
    let added = false
    for (const skill of claim.needs.filter(need => need !== 'Temporary accommodation')) {
      if (state.value.missions.some(mission => mission.claimId === claim.id && mission.skill === skill)) continue
      state.value.missions.push({ id: `${claim.id}-${skill}`, claimId: claim.id, providerId: null, skill,
        status: claim.access === 'Permitted' ? 'Proposed' : 'Blocked', start: null, end: null,
        estimate: claim.moneyLines.find(line => line.id === benefitIds[skill])?.amount ?? null })
      added = true
    }
    if (added) log('Assistance needs converted to draft missions', claim.id)
  }
  return { state, view, panel, detailOpen, selectedId, selected, query, page, pageCount, pageClaims, claims, filtered, queue, providers, missions, selectedMissions, stats, options, replan,
    claimGeojson, providerGeojson, assignmentGeojson, proposedGeojson, selectClaim, updateFilter, clearFilters, simulateOutage, applyPlan, cancelOutage, permitAccess, placeAccommodation, scheduleSelected, planAssistance }
}
