import { DEMO_CUSTOMER_ID, DEMO_POLICY_ID, DEMO_CLAIM_ID, DEMO_CLAIM_REFERENCE, DEMO_CONTRACT_NUMBER, DEMO_PRODUCT_NAME } from '../config/demo.js'

// Canonical, fictional scenario shared by the client and insurer composables.
export const claimStages = ['Collecting documents', 'Documents complete', 'Coverage proposal', 'Assistance arranged', 'Claim closed']

export const demoBenefits = [
  { id: 'accommodation', label: 'Emergency accommodation', limit: 1050, amount: 1000 },
  { id: 'essentials', label: 'Essential items', limit: 250, amount: 250 },
  { id: 'electrical', label: 'Electrical inspection and safety', limit: 500, amount: 350 },
  { id: 'drying', label: 'Cleaning and drying', limit: 3000, amount: 1800 },
  { id: 'contents', label: 'Damaged contents', limit: 30000, amount: 4500 },
  { id: 'building', label: 'Insured building repairs', limit: 200000, amount: 9000 },
  { id: 'storage', label: 'Moving and temporary storage', limit: 1000, amount: 600 }
]

export function createReferenceClaim() {
  return {
    id: DEMO_CLAIM_REFERENCE, databaseId: DEMO_CLAIM_ID, customerId: DEMO_CUSTOMER_ID, policyId: DEMO_POLICY_ID,
    contractNumber: DEMO_CONTRACT_NUMBER, productName: DEMO_PRODUCT_NAME,
    name: 'Jules', nameParameters: {}, initials: 'J', householdSize: 3,
    address: '12 rue des Oliviers, 13800 Istres', addressNote: 'Declared main residence.',
    event: 'Flooding', eventDate: '2026-10-20', date: '2026-10-20', reportedAt: '2026-10-22', stage: 0,
    municipality: 'Istres', cluster: 'Other Istres locations', coordinates: [4.9784, 43.5148],
    group: 'A', priority: 'P1', access: 'Unknown', coverage: 'Not reviewed', catNat: 'Not verified',
    opened: true, waitingHours: 2, occupancy: 'Owner-occupier',
    situation: 'Home uninhabitable', accommodationNeeded: true, accommodationPlaced: false,
    currentAccommodation: 'Staying with relatives; a lasting solution is needed.',
    priorityReason: 'The household cannot return home and has no lasting accommodation solution.',
    summary: 'The ground floor is flooded and furniture and appliances are damaged. Jules, his partner and their child are safe at his sister’s home, but that arrangement cannot continue. Site access is not yet confirmed.',
    needs: ['Temporary accommodation', 'Electrical safety', 'Cleaning and drying', 'Moving and storage', 'Damage assessment'],
    moneyLines: demoBenefits.map(benefit => ({ ...benefit, review: 'Not reviewed', approved: 0, paid: 0 })),
    chronology: [
      { label: 'Flood event', at: '2026-10-20T06:00:00+02:00' },
      { label: 'Report received', at: '2026-10-22T09:42:00+02:00' },
      { label: 'Household contacted', at: '2026-10-22T10:20:00+02:00' }
    ]
  }
}
