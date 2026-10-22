# Insurer-focused overview: active art direction

Page: `/overview`. Updated: 22 October 2026. English default, French toggle.

## Audience and story

The primary buyer is an insurance operations team handling a catastrophe-related claim surge. The first viewport explains a coordinated response and leads to the insurer workspace. Jules illustrates how an AI hotline, a contract lookup, evidence requests and subsequent assistance relate to the same claim. The principal scenario stays in the villages and residential areas around Istres and Berre Lagoon in southern France. The five other hazards remain secondary illustrative applications.

The portfolio is 80 synthetic households / 170 people. Opened reports, provisional gross estimates and resource availability come from the existing insurer composable. These are not official impact counts, approved indemnities, measured savings or live capacity.

## Repline reference

The requested art-direction reference is [Repline](https://repline.ai/), inspected visually and through visible computed styles on 22 October 2026. Relevant characteristics: centered headings, white background, light regular typography, generous vertical spacing, black pill actions and pale-grey product-preview bands. Repline's brand, sales copy, customer logos and interface were not copied.

The active hero uses a compact two-column layout: 320px text column, 46px/50px regular heading, original reference image beside it and native pill links. The header background spans the viewport; its inner 1440px container aligns with the hero. The sections below retain their existing typography and imagery.

## Reference A map and legend — active implementation

The user requested only the map and bottom legend from reference A, without the surrounding interface or floating claim card. The original source is `docs/design/insurer-option-a-map-first.png`. The built-in ImageGen tool performed a precise cleanup of the obscured map area: floating card, selected black pin and zoom controls removed; original palette, coastline, labels, clusters, provider icons and legend retained. Its result is `public/images/marketing/insurer-map-a-clean.png`. The original and previous delivery image remain preserved.

`MarketingWorkspacePreview.vue` displays a CSS crop of the cleaned 1536×1024 image: x=272–1160, y=216–1024 (888×808). This excludes the sidebar, heading, filters and right queue, keeping the map and legend. Its maximum display width is 620px. Metadata comes from `useMarketingContent.js`; `/overview` does not invoke `useMarketingWorkspacePreview.js` or mount a map. The illustration links to `/insurer`. Native caption and alternative text use the translation dictionaries.

The map is concept artwork, not geographic evidence or official incident data. Historical legend counts are preserved artwork; current application data remains Jules, Allstorm and 80 households / 170 people. No operational component, actual map geometry or portfolio record is changed. The Disaster Network mark and SVG favicon are retained.

## Regional photographic reference

A real [Saint-Chamas photograph on the Métropole Aix-Marseille-Provence website](https://ampmetropole.fr/metropole/le-territoire/les-communes/saint-chamas/) was inspected as an architecture/landscape reference. Its [image source](https://ampmetropole.fr/wp-content/uploads/2025/04/diapo-2-saint-chamas.jpg) shows the real locality; it is not a photograph of the flood or of Jules. The downloaded reference was used only as ImageGen input and is not redistributed in public assets.

The generated production scene is a fictional residential lane in the Istres / Berre Lagoon area. It changes the viewpoint and buildings; no depicted address is claimed to be an actual damaged insured property. Low hills, terracotta roofs, light plaster, shutters and Mediterranean vegetation replace the earlier generic setting. Its visible caption identifies it as an AI-generated illustration.

## Production assets

| File | Purpose |
| --- | --- |
| `public/images/marketing/jules-provence.webp` | Generated regional Jules illustration, based on the real regional reference; cwebp quality 85. |
| `public/images/marketing/hotline-flat.webp` | Generated and edited flat-style service illustration: caller holding a phone, damaged roofless house behind him, robot wearing a hotline headset; lossless WebP, alpha preserved. |

The headset robot illustration is a generic metaphor for the hotline, not additional evidence that Jules's flood claim includes a destroyed roof. Jules's existing household, contract, damage and assistance scenario remains unchanged. The original five supporting catastrophe illustrations are preserved. The earlier Jules image and initial design references are retained as first-pass history.

All image work used the built-in ImageGen tool. No CLI image-generation fallback or API key was used. Original PNG outputs remain in the ImageGen output directory; WebP encoding only prepares browser delivery. Page text and controls remain native Vue and use the flat EN/FR dictionaries; embedded English labels in the original concept image are deliberately preserved. No handwritten TypeScript, dependency install or backend change was introduced.

## Verification limits

The reference website and generated assets were inspected visually. Source was reviewed for the compact hero, reference A map cleanup, CSS map-and-legend cropping, full-width navigation, figure labelling and insurer-first links. The updated app was not launched; no build, tests or local browser/render checks were run. The source implementation is not a claim of verified browser fidelity.

## Final production prompts

### Regional Jules scene — built-in ImageGen generation with a regional photograph as reference

Use case: photorealistic-natural. Create one original editorial landscape photograph,1536x1024, based on Image1 as a REGIONAL ARCHITECTURE AND LANDSCAPE REFERENCE ONLY. Do not reproduce its exact composition or landmarks. Image1 is a real view of Saint-Chamas by the Berre Lagoon in southern France, not an incident image. New scene: a modest residential lane on the flat land near Istres/Entressen, in the Berre Lagoon area of Bouches-du-Rhône. Close-set simple pale stucco and local limestone houses with terracotta roofs, faded grey-blue shutters, small olive trees, low limestone wooded hills far away, NO mountains or Alpine scenery, NO luxury mansion, NO sea cliffs. After heavy autumn rain and a fictional flood, damp walls, a few muddy puddles and a little debris in the lane, no active dangerous flow. Fictional Jules, adult father in his30s with partner and one school-age child, seen from behind from a respectful distance, safe on dry ground with an overnight bag, no faces close-up. Natural overcast light, quiet restrained neutral colours, real regional architecture, dignified and humane. No text, logos or UI. This is a fictional generated illustration, not evidence of a real incident or a real damaged address.

### Hotline service illustration — built-in ImageGen edit

Edit the supplied illustration to follow the user's new composition. Keep its monochrome flat vector-like SaaS editorial art direction, landscape3:2 and transparent background. LEFT: replace the laptop pose with a fictional adult male customer HOLDING A MOBILE PHONE TO HIS EAR and talking. Behind him place a small modest Provençal house with a clearly missing roof, broken roof edges and one exposed roof beam, simple architectural silhouette; no fire, no flood, no injuries, no debris clutter. RIGHT: REMOVE THE FOLDER AND PAPERS COMPLETELY, replace them with a friendly clearly recognizable ROBOT HOTLINE AGENT wearing a headset with over-ear earcups and a small boom microphone. Robot should have a simple rounded-square head and small shoulders, professional and approachable, no sci-fi armour. CENTRE: keep a compact speech waveform and a fine connector between client and robot, showing a conversation. Strong flat silhouettes and crisp uniform charcoal outlines, white and very light neutral-grey solid fills, minimal geometry, generous negative space. Absolutely no glow, halo, gradient, blur, shadow, atmospheric backdrop, texture or3D shading. No laptop. No documents. No text or numbers. Entire background remains fully transparent. This is a generic illustrative hotline service image, not a depiction of Jules's actual flood damage.

