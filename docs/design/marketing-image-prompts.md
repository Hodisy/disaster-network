# Marketing imagery and design direction

This document records the initial image-led version. The active insurer-focused positioning, Repline-inspired layout, existing insurer-component preview and updated Jules/robot assets are documented in [the current direction](marketing-insurer-direction.md). The original five supporting photographs remain in use; the original hero/layout specifications below are historical references.

Product: Disaster Network. Fictional insurer: Allstorm Insurance.

Generation mode: built-in ImageGen; nine fresh generations, opaque backgrounds, no reference photographs. All production photographs are AI-generated fictional illustrations. None depicts an actual customer or proves an actual event. The six delivery assets are encoded as WebP with cwebp quality 85, at their original 1536×1024 dimensions; original generated PNGs remain in the ImageGen output directory.

Page: `/overview`. The existing `/` redirect and client/insurer workspaces remain unchanged. The page has a quiet navigation header, an image-led hero, Jules's shared claim journey, the synthetic portfolio context, five additional illustrative use cases and a closing demo link.

Visual specification: true white #ffffff; near-black #171719; cool grey body copy and rules; Inter/system sans serif; headline 80px/84px; section headings 48px/54px; white/black pill links; flat editorial photo tiles with 10–12px corners. No image overlays. Desktop only. Native Vue text and controls; images are separate assets.

The approved Jules fixture is one household of three within the 80-household / 170-person synthetic portfolio. The flood event date is 20 October 2026; the report date is 22 October 2026, Europe/Paris. Accommodation, evidence collection, coverage review, site access and interventions stay separate. The five additional scenarios are hypothetical product applications; they do not extend the fictional Allstorm flood policy or constitute separate working demos.

## Assets

| Asset | Role |
| --- | --- |
| `public/images/marketing/jules-flood.webp` | Production illustration |
| `public/images/marketing/los-angeles-wildfire.webp` | Production illustration |
| `public/images/marketing/brittany-storm.webp` | Production illustration |
| `public/images/marketing/hailstorm.webp` | Production illustration |
| `public/images/marketing/earthquake.webp` | Production illustration |
| `public/images/marketing/landslide.webp` | Production illustration |
| `docs/design/marketing-hero.png` | Design reference |
| `docs/design/marketing-journey.png` | Design reference |
| `docs/design/marketing-scenarios.png` | Design reference |

## Implementation decisions

- Hero concept controls the page's overall visual system. Journey and scenario concepts supply the downstream section composition.
- The journey concept invented a mountain photograph; implementation uses the generated Jules illustration, keeping the Berre Lagoon setting coherent.
- The scenario concept added a second introductory hero; this is omitted because the page already has its requested hero.
- The closing band uses solid near-black rather than the scenario concept's unrequested landscape background.
- EN/FR toggle, localised dates, explicit illustration labels, document/coverage caveats, expanded practical needs and the insurer portfolio explanation are intentional native-copy additions.
- The three portfolio values are fixture context, not official impact figures or measured product outcomes.
- Components use scoped CSS, Flexbox, px/% dimensions and props/events. Editorial fixtures live in useMarketingContent.js. No services, backend endpoints or external data calls were added.

## Validation status

All six production illustrations and all three design concepts were inspected as local images. Composition, white background, typography hierarchy, monochrome controls, photo treatment and section order were reviewed against the concepts in source. Browser/render fidelity is unverified: no app launch, build, tests or browser checks were performed, as required by the project's code-only workflow. No browser screenshot exists and visual fidelity is not claimed.

## Exact generation prompts

### jules-flood

Use case: photorealistic-natural. Generate one high-resolution landscape editorial photograph, 3:2 composition. Cohesive restrained documentary art direction: cool neutral colours, soft overcast daylight, realistic materials, natural detail, modest film grain, no orange-and-teal filter. This is an AI-created fictional scene for an insurance coordination demo, not a photograph of a real event. No words, logos, captions, montage, UI, borders or watermarks. No injured people, sensationalism or graphic destruction. Human-scale, dignified, quietly hopeful. Scene: an ordinary southern French residential street near Berre Lagoon after a flood. Foreground left: fictional Jules, a father in his thirties, seen from behind beside his partner and their school-age child, carrying a small overnight bag. They stand safely on dry pavement, looking toward their low pale-stone and stucco house with blue-grey shutters. Background shows shallow receding muddy water and damp ground-floor walls, no emergency danger. Olive tree, Mediterranean roof tiles, modest lived-in architecture. Eye-level wide photograph. People small in frame; the house and the feeling of returning after evacuation are the focus. Reserve upper sky visually quiet. No faces close up.

### los-angeles-wildfire

Use case: photorealistic-natural. Generate one high-resolution landscape editorial photograph, 3:2 composition. Cohesive restrained documentary art direction: cool neutral colours, soft overcast daylight, realistic materials, natural detail, modest film grain, no orange-and-teal filter. This is an AI-created fictional scene for an insurance coordination demo, not a photograph of a real event. No words, logos, captions, montage, UI, borders or watermarks. No injured people, sensationalism or graphic destruction. Human-scale, dignified, quietly hopeful. Scene: an evacuated Los Angeles hillside neighbourhood in the aftermath of a fictional wildfire. Wide residential street with characteristic California architecture, dry vegetation, a distant smoky hazy ridge, one carefully parked unbranded fire engine far away, a few ash-covered garden plants. One house has visible but restrained smoke and roof damage. No giant flames or people in peril. Frame focuses on the street and houses, not skyline landmarks.

### brittany-storm

Use case: photorealistic-natural. Generate one high-resolution landscape editorial photograph, 3:2 composition. Cohesive restrained documentary art direction: cool neutral colours, soft overcast daylight, realistic materials, natural detail, modest film grain, no orange-and-teal filter. This is an AI-created fictional scene for an insurance coordination demo, not a photograph of a real event. No words, logos, captions, montage, UI, borders or watermarks. No injured people, sensationalism or graphic destruction. Human-scale, dignified, quietly hopeful. Scene: a fictional windstorm aftermath in coastal Brittany, France. A modest granite house with slate roof, a small section of displaced roof tiles, fallen branches across a garden, wet road, windswept trees, distant grey Atlantic coast. Calm assessment after the storm, nobody in danger. Beautiful believable regional architecture and wide scene.

### hailstorm

Use case: photorealistic-natural. Generate one high-resolution landscape editorial photograph, 3:2 composition. Cohesive restrained documentary art direction: cool neutral colours, soft overcast daylight, realistic materials, natural detail, modest film grain, no orange-and-teal filter. This is an AI-created fictional scene for an insurance coordination demo, not a photograph of a real event. No words, logos, captions, montage, UI, borders or watermarks. No injured people, sensationalism or graphic destruction. Human-scale, dignified, quietly hopeful. Scene: aftermath of a fictional hailstorm in southwest France. A simple red-tiled suburban house, broken skylight and a few damaged tiles, white hailstones collected beside a garden path, battered green vines, wet surfaces under a clearing grey sky. Wide architectural composition, grounded and ordinary, no cars as main subject.

### earthquake

Use case: photorealistic-natural. Generate one high-resolution landscape editorial photograph, 3:2 composition. Cohesive restrained documentary art direction: cool neutral colours, soft overcast daylight, realistic materials, natural detail, modest film grain, no orange-and-teal filter. This is an AI-created fictional scene for an insurance coordination demo, not a photograph of a real event. No words, logos, captions, montage, UI, borders or watermarks. No injured people, sensationalism or graphic destruction. Human-scale, dignified, quietly hopeful. Scene: a hypothetical earthquake aftermath in a Mediterranean town. A pale plaster residential building with visible diagonal masonry cracks, safely cordoned entrance with plain unlettered tape, a structural engineer in a high-visibility vest seen from behind at a distance inspecting the exterior. No collapsed buildings, no injuries. Wide human-scale street, respectful and calm, stone and muted terracotta palette.

### landslide

Use case: photorealistic-natural. Generate one high-resolution landscape editorial photograph, 3:2 composition. Cohesive restrained documentary art direction: cool neutral colours, soft overcast daylight, realistic materials, natural detail, modest film grain, no orange-and-teal filter. This is an AI-created fictional scene for an insurance coordination demo, not a photograph of a real event. No words, logos, captions, montage, UI, borders or watermarks. No injured people, sensationalism or graphic destruction. Human-scale, dignified, quietly hopeful. Scene: a hypothetical landslide in a French Alpine village. A small timber-and-stone home remains standing, with a muddy debris deposit and rocks cutting off its access lane; nearby green mountain slopes under low clouds. No people in danger, no dramatic avalanche, no destroyed village. A realistic wide photograph illustrating access restrictions and the need for assessment.

### concept-hero

Use case: ui-mockup. Design a polished desktop marketing section for Disaster Network, an insurance claim coordination demo using fictional Allstorm Insurance. True white background #ffffff, near-black #171719 type, neutral grey dividers, Inter-like sans serif. Editorial, humane, spacious monochrome SaaS with cool muted landscape photography. No colourful gradients, no glassmorphism, no decorative icons, no hero eyebrow. Text and controls will be implemented in native Nuxt/Vue HTML; this is a concept reference only. Desktop width 1440px, readable English copy. Keep all text exactly as supplied, do not add claims or metrics. Native-like black pill CTA with a restrained arrow. Do not add a sidebar. One standalone complete first viewport, 1440 by 1000. Quiet header: Disaster Network wordmark left; 'Jules’s story' and 'Other scenarios' central; 'Explore the demo' right. Generous top margin. Big left-aligned 76px headline over two lines: 'When disaster hits, help should connect.' Under it a short grey paragraph: 'One call. One shared claim. A coordinated response. Follow Jules and his family from an urgent housing request to the insurer’s response.' Button: 'Meet Jules' and secondary text link 'Open insurer workspace'. Below, one expansive landscape photograph with subtle 16px corners, filling almost the whole page width: a fictional southern French family seen from behind on dry pavement outside their Mediterranean home after a flood, receding shallow water, cool quiet light. In the white margin beneath the photo an understated caption: 'Jules · 3 people · Berre Lagoon'. Plenty of breathing room, no floating stats or cards.

### concept-journey

Use case: ui-mockup. Design a polished desktop marketing section for Disaster Network, an insurance claim coordination demo using fictional Allstorm Insurance. True white background #ffffff, near-black #171719 type, neutral grey dividers, Inter-like sans serif. Editorial, humane, spacious monochrome SaaS with cool muted landscape photography. No colourful gradients, no glassmorphism, no decorative icons, no hero eyebrow. Text and controls will be implemented in native Nuxt/Vue HTML; this is a concept reference only. Desktop width 1440px, readable English copy. Keep all text exactly as supplied, do not add claims or metrics. Native-like black pill CTA with a restrained arrow. Do not add a sidebar. One standalone full middle story section continuing the same site. Wide top headline at left 'One family. A clear next step.' Grey supporting copy at right 'Jules needs a safe place for three people. His claim brings the call, evidence and proposed assistance together.' Below use a two-column 60/40 layout: left large open white numbered vertical journey, three generous rows separated by thin grey rules. Row 01 'Call for help' with 'Identify Jules, check the Allstorm contract and capture the urgent housing need.' Row 02 'Build the claim' with 'Request photos and supporting documents in My claim. Keep missing evidence visible.' Row 03 'Coordinate assistance' with 'Prepare accommodation and specialist visits, subject to review and availability.' On right one native UI preview frame, very subtle border: 'Jules’s claim', 'CAT-2026-001', 'Accommodation for 3', small rows 'Documents / Requested', 'Coverage / Under review', 'Assistance / Proposed'. Under it one simple pale grey band with three values '80 households', '170 people', '1 shared view' and small text 'Synthetic demonstration portfolio'. Bottom text link 'Follow the client journey →'. No complex dashboard, no fake metric, no layered cards.

### concept-scenarios

Use case: ui-mockup. Design a polished desktop marketing section for Disaster Network, an insurance claim coordination demo using fictional Allstorm Insurance. True white background #ffffff, near-black #171719 type, neutral grey dividers, Inter-like sans serif. Editorial, humane, spacious monochrome SaaS with cool muted landscape photography. No colourful gradients, no glassmorphism, no decorative icons, no hero eyebrow. Text and controls will be implemented in native Nuxt/Vue HTML; this is a concept reference only. Desktop width 1440px, readable English copy. Keep all text exactly as supplied, do not add claims or metrics. Native-like black pill CTA with a restrained arrow. Do not add a sidebar. One standalone lower section with heading 'Different disasters. The same need to coordinate.' and supporting sentence 'Five illustrative scenarios beyond Jules’s flood claim.' Below five editorial photographic tiles arranged three across then two across, but open flat presentation, no bordered cards. Consistent landscape 3:2 photographs, small 12px corners. Tile text beneath photo, no overlaid text. Titles and locations: 'Wildfire / Los Angeles', 'Windstorm / Brittany', 'Hailstorm / Southwest France', 'Earthquake / Mediterranean region', 'Landslide / French Alps'. Images depict calm realistic aftermath and residential property, no casualties. Each has a single grey descriptive line. At bottom a spacious near-black full-width closing band with white headline 'Make the next step clearer.' and white button 'Explore Jules’s claim'. Small footer 'Fictional insurance scenarios. AI-generated illustrations.' Avoid testimonials, pricing, partner logos and unsupported performance claims.

