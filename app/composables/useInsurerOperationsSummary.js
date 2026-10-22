// Read-only portfolio projections. No classification, booking or dispatch is performed.
export function useInsurerOperationsSummary({ claims, providers, missions, stats }) {
  const summary = computed(() => {
    const active = claims.value.filter(claim => claim.stage < 4)
    return {
      ...stats.value,
      urgent: active.filter(claim => claim.priority === 'P1').length,
      availableProviders: providers.value.filter(provider => provider.available && provider.remaining > 0).length,
      mobilizedProviders: providers.value.filter(provider => provider.used > 0).length,
      providerCount: providers.value.length,
      scheduled: missions.value.filter(mission => mission.status === 'Scheduled').length,
      proposed: missions.value.filter(mission => mission.status === 'Proposed').length,
      disrupted: missions.value.filter(mission => mission.status === 'Disrupted').length
    }
  })
  const services = computed(() => {
    const labels = [...new Set([...claims.value.flatMap(claim => claim.needs), ...providers.value.map(provider => provider.skill)])].filter(Boolean)
    return labels.map(skill => {
      const teams = providers.value.filter(provider => provider.skill === skill)
      const work = missions.value.filter(mission => mission.skill === skill)
      const units = [...new Set(teams.map(provider => provider.unit))].map(unit => ({
        unit,
        remaining: teams.filter(provider => provider.unit === unit && provider.available).reduce((sum, provider) => sum + provider.remaining, 0)
      }))
      return {
        skill,
        households: claims.value.filter(claim => claim.stage < 4 && (skill === 'Temporary accommodation' ? claim.unmetAccommodation : claim.needs.includes(skill))).length,
        available: teams.filter(provider => provider.available && provider.remaining > 0).length,
        scheduled: work.filter(mission => mission.status === 'Scheduled').length,
        blocked: work.filter(mission => mission.status === 'Blocked' || mission.status === 'Disrupted').length,
        units
      }
    }).filter(service => service.households || service.scheduled || service.blocked)
      .sort((a, b) => (a.skill === 'Temporary accommodation' ? -1 : b.skill === 'Temporary accommodation' ? 1 : b.households - a.households))
  })
  return { summary, services }
}
