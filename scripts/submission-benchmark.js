// Offline submission evidence. Executes the shipped frontend operations code;
// no database, microphone, OpenAI request, booking or deployment is involved.
import { ref, reactive, computed, watch } from 'vue'
import { createHash } from 'node:crypto'
import { readFile, writeFile, mkdir } from 'node:fs/promises'
import { useDemo } from '../app/composables/useDemo.js'
import { useInsurerWorkspace } from '../app/composables/useInsurerWorkspace.js'
import { createReferenceClaim } from '../app/composables/useClaimScenario.js'

const states = new Map()
const locale = ref('en')
// Minimal Nuxt lifecycle/state bridge for a standalone execution of the real
// composables. Identity translations affect labels only, never allocations.
Object.assign(globalThis, {
  ref, reactive, computed, watch,
  useState(key, initialize) {
    if (!states.has(key)) states.set(key, ref(initialize()))
    return states.get(key)
  },
  useAppI18n: () => ({ t: phrase => phrase, locale }),
  useDemo
})

const demo = useDemo()
demo.state.value.claim = createReferenceClaim()
const workspace = useInsurerWorkspace()
const original = structuredClone(JSON.parse(JSON.stringify({ claims: workspace.claims.value, providers: workspace.providers.value, missions: workspace.missions.value })))
await workspace.simulateOutage()
const affected = workspace.missions.value.filter(mission => mission.status === 'Disrupted')
const providers = JSON.parse(JSON.stringify(workspace.providers.value))
const disruptedIds = new Set(affected.map(mission => mission.id))
const untouched = JSON.parse(JSON.stringify(workspace.missions.value.filter(mission => !disruptedIds.has(mission.id)).map(({ claim, provider, ...mission }) => mission)))

// Ablation: same outage and skill directory, but capacity/access data is withheld.
// Assign each affected mission to the first available matching provider.
const naive = affected.map(mission => ({ missionId: mission.id, claimId: mission.claimId,
  providerId: providers.find(provider => provider.available && provider.skill === mission.skill)?.id || null }))
const proposed = workspace.replan.value.changes.map(change => ({ missionId: change.mission.id,
  claimId: change.mission.claimId, providerId: change.replacement?.id || null,
  assumedDelayHours: change.delayHours ?? null, assumedCostDeltaEUR: change.costDelta ?? null }))

function score(plan) {
  const allocated = new Map()
  let invalidAssignments = 0
  for (const assignment of plan) {
    const mission = affected.find(item => item.id === assignment.missionId)
    const provider = providers.find(item => item.id === assignment.providerId)
    const prior = allocated.get(assignment.providerId) || 0
    const valid = provider && provider.available && provider.skill === mission.skill
      && mission.claim.access === 'Permitted' && prior < provider.remaining
    if (!valid) invalidAssignments += 1
    if (provider) allocated.set(provider.id, prior + 1)
  }
  return { affectedMissions: affected.length, feasibleAssignments: plan.length - invalidAssignments,
    invalidAssignments, unassignedMissions: affected.length - plan.filter(item => item.providerId).length }
}

const blockedBefore = workspace.missions.value.filter(mission => mission.status === 'Blocked').map(mission => mission.id)
await workspace.applyPlan()
const blockedAfter = workspace.missions.value.filter(mission => mission.status === 'Blocked').map(mission => mission.id)
const afterUntouched = JSON.parse(JSON.stringify(workspace.missions.value.filter(mission => !disruptedIds.has(mission.id)).map(({ claim, provider, ...mission }) => mission)))
const applied = proposed.every(proposal => workspace.missions.value.some(mission => mission.id === proposal.missionId && mission.providerId === proposal.providerId && mission.status === 'Scheduled'))
const sourceFiles = ['app/composables/useInsurerWorkspace.js', 'app/composables/useClaimScenario.js', 'app/config/demo.js']
const sourceSha256 = {}
for (const file of sourceFiles) sourceSha256[file] = createHash('sha256').update(await readFile(file)).digest('hex')
const result = {
  generatedAt: new Date().toISOString(), scope: 'Offline deterministic frontend allocation comparison; not an LLM, production or live-service benchmark.',
  dataset: { households: original.claims.length, people: original.claims.reduce((sum, claim) => sum + claim.householdSize, 0),
    providers: original.providers.length, initialMissions: original.missions.length,
    sha256: createHash('sha256').update(JSON.stringify(original)).digest('hex') },
  sourceSha256,
  scenario: 'Delta Drying becomes unavailable; two scheduled drying missions require replacement.',
  baseline: { description: 'First available matching provider; ignores remaining capacity and property access.', assignments: naive, ...score(naive) },
  dataAware: { description: 'Shipped useInsurerWorkspace replan and applyPlan; checks availability, matching skill, remaining capacity, access and per-plan reservations.', assignments: proposed, ...score(proposed) },
  integrity: { planApplied: applied, blockedMissionsPreserved: JSON.stringify(blockedBefore) === JSON.stringify(blockedAfter),
    untouchedMissionsPreserved: JSON.stringify(untouched) === JSON.stringify(afterUntouched), blockedMissionCount: blockedAfter.length },
  limitations: ['One hand-authored outage on a synthetic portfolio, not statistical generalization.',
    'Both planners receive the same affected missions; the baseline deliberately lacks capacity/access context.',
    'Reported delays and mission prices are fixture assumptions, not measured response times or savings.',
    'No paid voice call, policy-verification or PostgreSQL persistence was exercised.',
    'This demonstrates the value of resource data and deterministic rules, not superior AI reasoning.']
}
if (!applied || !result.integrity.blockedMissionsPreserved || !result.integrity.untouchedMissionsPreserved) throw new Error('Evidence integrity failed; no success report written.')
await mkdir('docs/submission/evidence', { recursive: true })
await writeFile('docs/submission/evidence/benchmark-results.json', JSON.stringify(result, null, 2) + '\n')
await writeFile('docs/submission/evidence/benchmark-results.csv', 'metric,naive_no_capacity_context,data_aware_application\n' +
  ['affectedMissions', 'feasibleAssignments', 'invalidAssignments', 'unassignedMissions'].map(key => `${key},${result.baseline[key]},${result.dataAware[key]}`).join('\n') + '\n')
console.log(JSON.stringify(result, null, 2))
