// Public fixture identifiers only. This is not authentication or verification.
// Future frontend API adapters can use these stable IDs without exposing secrets.
export const DEMO_CUSTOMER_ID = '10000000-0000-4000-8000-000000000001'
export const DEMO_POLICY_ID = '20000000-0000-4000-8000-000000000001'
export const DEMO_CLAIM_ID = '30000000-0000-4000-8000-000000000001'
export const DEMO_CLAIM_REFERENCE = 'CAT-2026-001'

// Stable public policy details; the verification answer never belongs in client configuration.
export const DEMO_CONTRACT_NUMBER = '123456'
export const DEMO_PRODUCT_NAME = 'Allstorm Home Protect'
export function normalizeContractNumber(value) {
  if (typeof value !== 'string') return ''
  const number = value.replace(/[\s-]/g, '')
  return /^\d{6}$/.test(number) ? number : ''
}
