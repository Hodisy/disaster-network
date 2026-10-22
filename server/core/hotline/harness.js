import { createHash, randomUUID } from 'node:crypto'
import * as claims from '../../repositories/hotlineClaims.js'
import { readIncidentContext } from '../../repositories/incidentContext.js'
import { checkOfficialSource, readClaimEvidence } from './evidence.js'
import { DEMO_CUSTOMER_ID, DEMO_POLICY_ID, DEMO_CLAIM_ID } from '../../../app/config/demo.js'

export const NEEDS = ['accommodation', 'essential_items', 'electrical_safety', 'cleaning_drying', 'moving_storage', 'contents', 'building']
const string = { type: 'string' }
const enumeration = values => ({ type: 'string', enum: values })
const tool = (name, description, properties = {}) => ({ type: 'function', name, description, strict: true,
  parameters: { type: 'object', properties, required: Object.keys(properties), additionalProperties: false } })
export const hotlineTools = [
  tool('find_demo_customer', 'Find only the reserved demo customer by first name and contract reference; return only the verification challenge before verification.', { first_name: string, contract_number: string }),
  tool('verify_demo_customer', 'Verify the answer to the database challenge in this active voice session. At most three attempts.', { answer: string }),
  tool('get_customer', 'Read the verified demo customer household from PostgreSQL.'),
  tool('get_policy', 'Read actual fictional policy terms, limits and document_sections source references from PostgreSQL. Never infer eligibility.'),
  tool('get_claim', 'Read the verified reserved claim. No duplicate claim creation.'),
  tool('get_incident_context', 'Read the fixed dated scenario context. No weather service or live research.'),
  tool('get_claim_evidence', 'Read genuinely stored supporting-document metadata and existing AI extraction results. Photos do not prove cause, date, authenticity or coverage. Never claim an unanalysed file was read.'),
  tool('check_official_catnat', 'Consult the public Georisques GASPAR API for the insured municipality and event date. Return sourced dated matches, no match, incomplete or unavailable. Never infer coverage or official recognition from absence. No general web search.'),
  tool('prepare_report', 'Prepare a report and assistance/document requests for browser review. Does NOT persist or book anything. Collect user facts, do not invent missing details. Unknown is allowed.', {
    description: string, safe: enumeration(['yes', 'no', 'unknown']), household_size: { type: 'integer', minimum: 1, maximum: 20 },
    habitability: enumeration(['habitable', 'uninhabitable', 'unknown']), access: enumeration(['unknown', 'restricted']),
    current_accommodation: string, accommodation_needed: { type: 'boolean' },
    accommodation_start: string, accommodation_nights: { type: 'integer', minimum: 0, maximum: 60 },
    accommodation_constraints: string,
    needs: { type: 'array', items: enumeration(NEEDS) },
    documents: { type: 'array', items: enumeration(['photos', 'address', 'inventory', 'invoice']) }
  }),
  tool('save_confirmed_report', 'Persist the prepared report only after the user explicitly confirms its exact draft in the browser. Never accepts model or spoken confirmation as browser authorization.', { draft_id: string })
]
const fail = code => { throw Object.assign(new Error(code), { code }) }
export function assertVerified(session) {
  if (session.closed || !session.verifiedUntil || session.verifiedUntil < Date.now()) fail('CUSTOMER_VERIFICATION_REQUIRED')
}
function argsFor(name, raw) {
  if (typeof raw !== 'string' || raw.length > 12000) fail('INVALID_TOOL_ARGUMENTS')
  let args
  try { args = JSON.parse(raw) } catch { fail('INVALID_TOOL_ARGUMENTS') }
  const definition = hotlineTools.find(item => item.name === name)
  if (!definition || !args || Array.isArray(args) || typeof args !== 'object') fail('INVALID_TOOL_ARGUMENTS')
  const properties = definition.parameters.properties
  if (Object.keys(args).length !== Object.keys(properties).length || Object.keys(args).some(key => !Object.hasOwn(properties, key))) fail('INVALID_TOOL_ARGUMENTS')
  for (const [key, spec] of Object.entries(properties)) {
    const value = args[key]
    if (spec.type === 'integer' ? !Number.isInteger(value) || value < spec.minimum || value > spec.maximum : spec.type === 'array'
      ? !Array.isArray(value) || value.length > 10 || new Set(value).size !== value.length || value.some(item => !spec.items.enum.includes(item))
      : typeof value !== spec.type || (spec.type === 'string' && value.length > 4000)) fail('INVALID_TOOL_ARGUMENTS')
    if (spec.enum && !spec.enum.includes(value)) fail('INVALID_TOOL_ARGUMENTS')
  }
  return args
}
export function publicDraft(draft) {
  if (!draft) return null
  const { baseVersion, ...visible } = draft
  return visible
}
export async function saveDraft(session, config) {
  assertVerified(session)
  const draft = session.draft
  if (!draft || session.approvedDraft !== draft.id) fail('BROWSER_CONFIRMATION_REQUIRED')
  const result = await claims.commitReport(config, draft, draft.id, session.conversationId, session.locale)
  session.saved = true
  session.claimGrantUntil = Date.now() + 30 * 60 * 1000
  session.evidenceGrantUntil = session.claimGrantUntil
  session.claim = result.claim
  return { saved: true, claim_reference: result.claim.reference, replayed: result.replayed, fictional: true }
}
export async function runTool(session, config, name, raw) {
  const args = argsFor(name, raw)
  if (session.closed || session.closing) fail('SESSION_CLOSED')
  if (name === 'find_demo_customer') {
    const challenge = await claims.verificationQuestion(config, args.first_name, args.contract_number)
    if (!challenge) return { found: false }
    session.challengeIssued = true
    return { found: true, ...challenge }
  }
  if (name === 'verify_demo_customer') {
    if (!session.challengeIssued) fail('CHALLENGE_REQUIRED')
    if (session.verificationAttempts >= 3) fail('VERIFICATION_LOCKED')
    session.verificationAttempts += 1
    const result = await claims.verifyCustomer(config, args.answer)
    if (result.verified) session.verifiedUntil = Math.min(session.expiresAt, Date.now() + 20 * 60 * 1000)
    else { session.verifiedUntil = 0; session.claimGrantUntil = 0; session.evidenceGrantUntil = 0; session.approvedDraft = null }
    return result
  }
  if (name === 'get_incident_context') return readIncidentContext()
  assertVerified(session)
  if (name === 'get_customer') return claims.readCustomer(config)
  if (name === 'get_policy') return claims.readPolicy(config)
  if (name === 'get_claim_evidence') return readClaimEvidence(config)
  if (name === 'check_official_catnat') return checkOfficialSource(config)
  if (name === 'get_claim') {
    const claim = await claims.readClaim(config)
    if (claim.intake_status === 'reported') {
      session.claimGrantUntil = Date.now() + 30 * 60 * 1000
      session.evidenceGrantUntil = session.claimGrantUntil
    }
    return claim
  }
  if (name === 'prepare_report') {
    const dateValid = value => value === '' || /^\d{4}-\d{2}-\d{2}$/.test(value) && Number.isFinite(Date.parse(value)) && new Date(`${value}T12:00:00Z`).toISOString().slice(0, 10) === value
    if (!args.description.trim() || !dateValid(args.accommodation_start) || args.accommodation_needed !== args.needs.includes('accommodation')) fail('INVALID_TOOL_ARGUMENTS')
    const current = await claims.readClaim(config)
    const policy = await claims.readPolicy(config)
    const warnings = []
    if (args.household_size !== current.household_size) warnings.push('The declared household differs from the policy. A handler must review it.')
    if (args.accommodation_nights > policy.coverage_data.accommodation_max_nights) warnings.push('The requested stay exceeds the policy limit and requires separate review.')
    if (args.safe !== 'yes') warnings.push('Safety is not confirmed. Do not enter an unsafe property.')
    if (current.closed_at) fail('CLAIM_CLOSED')
    // The draft hash is independent of tool call IDs, preventing duplicate writes on retries.
    const content = { ...args, baseVersion: current.version_token }
    const id = createHash('sha256').update(JSON.stringify(content)).digest('hex')
    if (session.draft?.id !== id) session.saved = false
    session.draft = { id, description: args.description, safe: args.safe, householdSize: args.household_size,
      habitability: args.habitability, access: args.access, currentAccommodation: args.current_accommodation,
      accommodationNeeded: args.accommodation_needed, accommodationStart: args.accommodation_start,
      accommodationNights: args.accommodation_nights, accommodationConstraints: args.accommodation_constraints,
      needs: args.needs, documents: args.documents, warnings, baseVersion: content.baseVersion }
    if (session.approvedDraft !== id) session.approvedDraft = null
    return { status: 'awaiting_browser_confirmation', draft_id: id, fictional: true, saved: false,
      warnings }
  }
  if (name === 'save_confirmed_report') {
    if (args.draft_id !== session.draft?.id) fail('DRAFT_STALE')
    return saveDraft(session, config)
  }
  fail('INVALID_TOOL_ARGUMENTS')
}
export function newHarnessContext(locale) {
  return { conversationId: randomUUID(), locale, verificationAttempts: 0, verifiedUntil: 0, challengeIssued: false,
    draft: null, approvedDraft: null, saved: false, claimGrantUntil: 0, evidenceGrantUntil: 0, claim: null,
    customerId: DEMO_CUSTOMER_ID, policyId: DEMO_POLICY_ID, claimId: DEMO_CLAIM_ID }
}
