# Insurer workspace brief for 80 households

Updated: 22 October 2026. Product: Disaster Network. Fictional insurer: Allstorm Insurance.

This is the reference scenario for the insurer workspace and the next client-copy update. The portfolio contains **80 households, 80 principal home claims and 170 synthetic people**, including Jules's household and claim. The historical event occurred on **20 October 2026**; the demonstration takes place on **22 October 2026**, in Europe/Paris. All coverage terms, households, losses, classifications, providers and costs below are fictional demonstration assumptions.

Public event evidence and actual insurance procedure are documented separately in the [event report](../research/20-october-2026-flood-evidence.md). The real statement about approximately 200 households concerned by evacuation is not our portfolio size or evidence of Allstorm's market share.

# Portfolio units and composition

One household means the occupants of one insured main residence. One principal home claim represents that household's report for this event. A claim may have several assistance needs, documents and missions. An occupant is not an additional claim. Exclude separate motor, commercial and condominium claims from this first portfolio.

The estimate of about 170 people is plausible against Istres's 2023 average of 2.15 occupants per main residence. Applying that average to 80 residences gives 172; the following synthetic fixture uses an exact total of 170 instead. Neither number estimates the actual event population. [Insee, household size](https://www.insee.fr/fr/statistiques/2011101?geo=COM-13047)

| Household size | Households | People |
| --- | --- | --- |
| 1 person | 30 | 30 |
| 2 people | 25 | 50 |
| 3 people | 10 | 30 |
| 4 people | 15 | 60 |
| Total | 80 | 170 |

Jules belongs to one of the ten three-person households. The 80 claims include his claim; do not add him as an 81st file when the hotline opens it. Reserve his fixture identifier and merge or upsert the client state into that record. Before the call, it can be marked as a demo record awaiting report; distinguish portfolio households from opened claims if fewer than 80 claims have actually been opened in the current demonstration state.

## Mutually exclusive situation groups

| Group | Households | Reported situation | Initial handling |
| --- | --- | --- | --- |
| A | 15 | Home reported uninhabitable, physical damage reported | Assess accommodation, protection and damage; Jules is included |
| B | 10 | Evacuated or inaccessible, physical damage unknown | Check assistance terms; do not invent building damage |
| C | 35 | Physical damage reported, home still habitable | Assess cleaning, drying, belongings and repairs |
| D | 20 | Limited consequences or facts still unconfirmed | Contact, qualify and collect proportionate evidence |
| Total | 80 | Each household appears in one group | Needs can overlap within a group |

These groups are not priority levels, coverage decisions or claim stages. Accommodation need must distinguish households already staying with relatives from those needing a placement now. Do not label every person in A or B as requiring an immediate hotel room.

## Synthetic geographic distribution

| Cluster | Households |
| --- | --- |
| Entressen | 20 |
| Les Heures Claires | 10 |
| Other Istres locations | 15 |
| Saint-Chamas | 15 |
| Saint-Mitre-les-Remparts | 10 |
| Martigues area | 10 |
| Total | 80 |

For a simple coherent seed, place group B's ten files at Les Heures Claires; place Jules in Other Istres. The other situation groups can be distributed across the remaining clusters. These counts and markers are invented, not official damage counts or flood boundaries. Use accurate basemap geography. Do not attach invented losses to real residents or claim access to a specific real property has been authorised.

# Fictional Allstorm coverage

Product label: **Allstorm Home Protect**. Demonstration version: **1**, effective **1 January to 31 December 2026**. The named insured is an owner-occupier of a declared main residence. All 80 home files use this demo product; their observed facts and review statuses vary.

For this fictional product, direct flood damage and the assistance benefits below can be considered without waiting for a CatNat recognition order. This is an invented contractual route, not a statement of French statutory entitlement or any real insurer's policy. The separate CatNat status remains **Not verified** until matching public recognition evidence is established.

| Benefit | Fictional limit | Required facts or evidence | Treatment |
| --- | --- | --- | --- |
| Emergency accommodation | Up to 7 nights and EUR 150 per household per night, maximum EUR 1,050 | Declared main residence unusable or officially inaccessible, household size, dates, current accommodation and demonstrated need | Assistance proposal; no payment or booking is real |
| Essential items | Up to EUR 250 per household for the event | Identified need and itemised expenditure | Assistance budget, separate from damaged belongings |
| Electrical inspection and immediate safety work | Up to EUR 500 | Access permitted and intervention need recorded | Qualified professional; no remote assertion of electrical safety |
| Cleaning and drying | Up to EUR 3,000 | Relevant damage, site access and work assessment | Separate tasks with equipment, duration and skill requirements |
| Moving and temporary storage | Up to EUR 1,000 | Recoverable belongings requiring protection | Avoid counting the same item as both a total loss and a stored intact item |
| Damaged contents | Up to EUR 30,000 for the event | Inventory, ownership/value evidence and assessment | Repair or assessed compensation; no automatic new-for-old promise |
| Insured building repairs | Up to EUR 200,000 for the event | Owner-occupier status, insured property, cause and assessed scope | Separate from belongings and assistance |

Accommodation beyond seven nights goes to a distinct review; no automatic extension or additional limit is assumed. Personal injury, vehicles, commercial interruption, undeclared property and unrelated or pre-existing damage are outside this demonstration product. Uncertain causation or ground instability requires review rather than automatic acceptance or rejection.

For this fictional contractual flood route only, apply **one EUR 300 deductible to the combined eligible contents and building amount per household for this event**, not per document or task. The five assistance benefits above have no demo deductible. Do not reuse this fictional deductible for a statutory CatNat calculation, and do not stack the two routes to pay the same expense twice. Deductible calculations occur after the benefit limits and eligible amounts have been assessed.

Public insurer guidance supports the plausibility of flood assistance without prior CatNat recognition and of accommodation/cleaning benefits, but none of the above numerical terms reproduces an actual policy. [MAIF flood guidance](https://www.maif.fr/habitation/guide-assurance-habitation/inondation), [MFA assistance examples](https://mfa.fr/assurance-habitation/assistance-mfa-habitation/)

# Jules reference claim

## Identity and situation

- Display name: **Jules**. Do not invent a surname or reuse a real resident's identity.
- Household: Jules, his partner and one child; three people, counted once within the portfolio.
- Occupancy: owner-occupier of an insured main residence in Istres, in the Other Istres synthetic cluster.
- Event: flood on 20 October 2026. Report and demonstration: 22 October 2026.
- Situation: ground-floor water damage and damaged furniture/appliances. The household is safe with Jules's sister, but that temporary arrangement cannot continue; accommodation is needed.
- Access: unknown at report intake. Physical interventions remain blocked until access can be treated as permitted by an explicitly labelled scenario update; this is not a public authority's real authorisation.
- Claim link: reuse the existing client identifier `CAT-2026-001` when the UI is migrated; do not create a second client claim under another identifier.

The frontend implementation now uses Jules in client scripts, account labels, summaries and both translation dictionaries; the canonical record is shared with the insurer portfolio. Preserve user-written messages, selected filenames and existing real street names; do not present a demo street address as evidence of an actual damaged residence.

## Estimated loss and assistance budget

All amounts are synthetic estimates, not market quotes or a settlement commitment.

| Line | Estimate | Benefit | Initial state |
| --- | --- | --- | --- |
| Initial accommodation | EUR 1,000 | Emergency accommodation | Proposed; actual nights/rate must reconcile within the EUR 1,050 cap |
| Essential items | EUR 250 | Essential items | Needs/receipts to confirm |
| Electrical inspection and safety | EUR 350 | Electrical safety | Pending access and assessment |
| Cleaning and drying | EUR 1,800 | Cleaning and drying | Pending access and assessment |
| Furniture and appliances | EUR 4,500 | Contents | Provisional inventory; value and repairability to assess |
| Floors, walls and insured fixtures | EUR 9,000 | Building repairs | Pending assessment |
| Moving and storage | EUR 600 | Moving and storage | Proposed for recoverable belongings |
| Gross estimate | EUR 17,500 | All seven lines | Not an approved or paid amount |

The gross estimate separates into EUR 4,000 assistance/protection and EUR 13,500 contents/building. Only if every line were approved at these estimates would the fictional EUR 300 deductible reduce the potential insurer total to **EUR 17,200**. This is an illustrative calculation, not Jules's initial approved amount. Initial paid amount is **EUR 0**; no real intervention has occurred.

For a reproducible first accommodation line, the fixture may use five nights at EUR 150 plus two at EUR 125, totalling seven nights and EUR 1,000 for the household. These rates are scenario inputs, not hotel availability or a quote.

## Documents and state transitions

The existing requested categories remain damage photos, proof of address and an inventory. A later invoice can substantiate one item's value; an estimate can refine the building-repair line. Completed upload categories do not prove eligibility, authenticity or a EUR 17,500 loss.

Track these independently:

- Claim progress: Collecting documents, Documents complete, Coverage proposal, Assistance arranged, Claim closed.
- Assistance need: accommodation, electrical safety, cleaning/drying, storage and assessment.
- Coverage review: not reviewed, proposal, requires review, or simulated approval for specific lines.
- Site access: unknown, restricted or permitted in the synthetic scenario.
- Missions: proposed, approved in demo, scheduled, blocked, cancelled or completed in demo.
- Money: gross estimate, eligible proposal, deductible, approved commitment and paid amount.

Jules starts in group A with an unmet accommodation need. Priority is **P1 Urgent** under the demo rule because his household cannot return and has no durable accommodation solution, not because the claim is expensive. An eventual accommodation placement changes the unmet need; it does not make the home habitable or close the claim.

# Insurer workspace requirements

The workspace should answer: who needs help now, where can crews safely work, which capacity is missing, and what provisional financial exposure is emerging?

1. A map-centred overview links the 80 synthetic household records, situation groups, priorities, access status and providers. Distinguish 80 portfolio households from the number of currently opened reports.
2. Claim detail opens Jules's existing client dossier, including household size, chronology, multiple needs, evidence, coverage lines and the gross/approved/paid distinction.
3. Mobilisation proposes qualified resources by skill, location, availability and access. Accommodation counts households and room requirements; it does not allocate one hotel booking per person automatically.
4. Replanning responds to an unavailable drying crew and displays only affected missions, replacement options and outstanding constraints. Restricted-access properties remain blocked.
5. Filters and badges keep priority, document completeness, coverage, access and assignment separate. P1/P2/P3 are fictional operational categories, not a real insurer's classification.

For a small provider fixture, six drying teams plus electrical inspection, assessment, cleaning and accommodation resources are sufficient. Their capacity units must be explicit: concurrent sites, visits per day or available household accommodation slots. Never add different capacity units into one total or invent completion-time savings before a comparable calculation exists.

Portfolio totals are computed from the 80 records. Do not multiply Jules's EUR 17,500 estimate by 80, assume every household needs accommodation or copy financial totals from the historical design images. Needs may overlap, but money lines must not be counted twice. All headline exposure totals remain synthetic and carry an estimate status.

# Implementation scope and handoff

This document remains the scenario specification. A subsequent frontend implementation now provides the 80-record fixture, shared Jules record, map-first insurer screen, filters, dossier detail, mobilization and simulated replanning. The implementation is source-only; no build, tests, browser review or real integration has been executed for it.

Continue to use this scenario for frontend work under AGENTS.md: Nuxt/Vue JavaScript, composable-injected frontend mocks, stateless presentation components, scoped styles and the existing translation layer. Preserve the approved visual direction. The older 200-claim and Claire references in images are historical layout references, not current fixture specifications.

Before delivery, reconcile the records and copy: 80 households; 170 people; situation groups 15/10/35/20; Jules included once; event on 20 October; report on 22 October; seven budget lines totalling EUR 17,500; no approval implied by document collection; no live or official synthetic map layers. No build, tests, browser checks, deployment or external calls are authorised by this documentation handoff.

## Source implementation locations

`app/composables/useClaimScenario.js` owns Jules’s canonical seven-line budget and claim stages. `useInsurerWorkspace.js` generates the fixed 80-household / 170-person portfolio, merges the existing client record, derives all totals, and contains provider/mission mocks and operational actions. `useOperationsMap.js` receives local GeoJSON and renders it with MapLibre GL JS / OpenFreeMap Positron; public basemap downloads do not supply incident data. `app/pages/insurer.vue` binds the presentational components and generates the local report. All new interface phrases use the existing flat EN/FR dictionaries, with English default.

The map keeps its large footprint while the right panel switches Priority queue / All claims / claim detail. Mobilization adds provider capacities and assignments. Provider-outage replanning shows proposed replacements and their assumed timing/cost differences before a simulated apply. Unknown or restricted access remains distinct from a coverage decision; blocked missions cannot be reassigned into inaccessible properties. See the README for current operational controls and limitations.

## Presentation realism

Operational screens use invented first names, varied household situations, coherent report/contact times and severity-dependent estimates. Do not repeat “demo”, “fictional” or “synthetic” in every conversation, status or benefit label. Retain provenance in Settings, source/data metadata, the exported report and one discreet workspace notice. Jules has **Allstorm Home Protect**, policy **123456**, with the existing claim **CAT-2026-001**. No surname is added. The three-record database fixture aligns Camille (123457, one person, Entressen) and Sam (123458, two people, Les Heures Claires) with their portfolio records; the other 77 portfolio records remain frontend-only. Household totals, groups, access constraints, contractual limits and money semantics remain unchanged. Historical design images retain their original text.
