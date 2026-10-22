// Editorial fixtures only; visiting the overview never creates or changes a claim.
export function useMarketingContent() {
  // Original concept map artwork; historical legend figures are not portfolio data.
  const workspaceImage = {
    src: '/images/marketing/insurer-map-a-clean.png',
    alt: 'Berre Lagoon map from the original interface concept, with claim markers and a legend.',
    caption: 'Illustrative map · Original concept'
  }

  const story = {
    name: 'Jules',
    claimId: 'CAT-2026-001',
    eventDate: '2026-10-20',
    reportDate: '2026-10-22',
    image: '/images/marketing/jules-provence.webp',
    alt: 'AI-generated illustration of Jules and his family on a rain-soaked residential lane in the Berre Lagoon area of southern France.',
    hotlineImage: '/images/marketing/hotline-flat.webp',
    hotlineAlt: 'Flat illustration of a customer on the phone beside a roofless house, speaking with a robot wearing a hotline headset.',
    steps: [
      {
        number: '01', title: 'Call for help',
        description: 'Identify Jules, check the Allstorm contract and capture the urgent housing need.',
        detail: 'Three people need accommodation. They are safe with family for now, but cannot return home.'
      },
      {
        number: '02', title: 'Build the claim',
        description: 'Request photos and supporting documents in My claim. Keep missing evidence visible.',
        detail: 'Damage photos, proof of address and an inventory stay attached to the same claim and conversation.'
      },
      {
        number: '03', title: 'Coordinate assistance',
        description: 'Prepare accommodation and specialist visits, subject to review and availability.',
        detail: 'Electrical checks, drying, storage and repairs are separate needs. Physical visits wait for permitted access.'
      }
    ],
    review: [
      { label: 'Documents', value: 'Requested' },
      { label: 'Coverage', value: 'Under review' },
      { label: 'Assistance', value: 'Proposed' }
    ]
  }

  const portfolio = [
    { value: '80', label: 'households' },
    { value: '170', label: 'people' },
    { value: '1', label: 'shared view' }
  ]

  const scenarios = [
    {
      id: 'wildfire', title: 'Wildfire', location: 'Los Angeles', image: '/images/marketing/los-angeles-wildfire.webp',
      alt: 'AI-generated illustration of a quiet Los Angeles residential street with smoke haze after a wildfire.',
      description: 'Connect evacuation needs, temporary housing and smoke-damage assessments.',
      need: 'A safe place to stay', response: 'Group accommodation needs and organise property assessments.'
    },
    {
      id: 'windstorm', title: 'Windstorm', location: 'Brittany', image: '/images/marketing/brittany-storm.webp',
      alt: 'AI-generated illustration of a Breton granite house with roof damage and fallen branches after a storm.',
      description: 'Prioritise exposed homes and coordinate roof protection and repairs.',
      need: 'Protect the home', response: 'Match urgent protection work with available roofers.'
    },
    {
      id: 'hailstorm', title: 'Hailstorm', location: 'Southwest France', image: '/images/marketing/hailstorm.webp',
      alt: 'AI-generated illustration of hailstones and a damaged tiled roof in southwest France.',
      description: 'Turn a surge of roof and window reports into an organised assessment queue.',
      need: 'Document the damage', response: 'Collect evidence and group nearby inspection requests.'
    },
    {
      id: 'earthquake', title: 'Earthquake', location: 'Mediterranean region', image: '/images/marketing/earthquake.webp',
      alt: 'AI-generated illustration of an engineer assessing a cordoned residential building after a hypothetical earthquake.',
      description: 'Keep safety, access restrictions and specialist assessment visible in every claim.',
      need: 'Know the next safe step', response: 'Escalate structural reviews and keep unsafe visits blocked.'
    },
    {
      id: 'landslide', title: 'Landslide', location: 'French Alps', image: '/images/marketing/landslide.webp',
      alt: 'AI-generated illustration of a mountain home with its access lane blocked by a landslide.',
      description: 'Coordinate support for households whose homes or access routes need assessment.',
      need: 'Reach practical support', response: 'Track access constraints before planning specialist visits.'
    }
  ]

  return { workspaceImage, story, portfolio, scenarios }
}
