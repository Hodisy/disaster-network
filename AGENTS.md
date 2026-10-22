# Agent instructions

## Project status

Disaster Network is a Nuxt 4 / Vue frontend demo for insurance-claim coordination. Its original name was CAT//RESPONSE; preserve the current Disaster Network branding and the fictional Allstorm Insurance name. **English is the default interface language; a user-requested EN / FR toggle also provides French.** Keep source phrases, documentation, comments, and new design mockups in English. All visible copy, mock conversations/data labels, errors, placeholders, tooltips, accessibility labels, and page titles must use the translation layer. Preserve proper names, real addresses, user-written messages, and filenames. Use English route/file names and locale-aware `en-GB` / `fr-FR` date/time formatting; retain the `Europe/Paris` time zone for the scenario.

As of October 22, 2026, source code implements `/hotline`, `/client` and the approved map-first `/insurer` workspace. The insurer fixture contains **80 synthetic household records / principal claims covering 170 people**, including **Jules** exactly once as `CAT-2026-001`. Before the client report, 79 records are opened and Jules is reserved; opening the hotline claim merges it into that same portfolio record. The client uses Jules, his partner and child, the shared seven-line EUR 17,500 gross estimate, and separate event/report dates. Read `docs/scenarios/insurer-workspace-brief.md` for the fixture and fictional Allstorm benefits. Historical design images with older names or 200-claim figures are layout references only. Source implementation does not mean build or rendering has been verified.

Read `README.md` for current capabilities and limitations. The approved visual references are `docs/design/hotline-reference-v2.png` and `docs/design/client-reference-v2.png`. Preserve their monochrome SaaS direction, animated particle orb, activity timeline, and details rail.

## Package manager

- Use **Bun**, not npm, pnpm, or Yarn.
- Commands: `bun install`, `bun run dev`, `bun run build`, `bun run preview`.
- Commit and maintain `bun.lock` when dependencies change. The legacy local `package-lock.json` is ignored; do not generate or update npm lockfiles. Do not run dependency installation solely to edit documentation.

## Current authorized scope

- Frontend code, components, composables/hooks, CSS, and mocks only.
- Do not implement business backend code, Nitro API endpoints, server repositories, databases, or Aiven connections in this phase.
- Do not add real AI/weather calls, microphone capture, voice sessions, uploads, authentication, bookings, or payments.
- Keep all identities, guarantees, and outcomes clearly fictional. The report/demo date is October 22, 2026; the flood event date is October 20, 2026. The insurer design uses the flood around Berre Lagoon northwest of Marseille as dated context; its 80-household portfolio, zones, classifications, resources, and costs remain synthetic, never live or official data. Jules's gross EUR 17,500 estimate covers multiple benefit lines and is not an approved or paid amount. Keep the fictional contractual flood route separate from any unverified statutory CatNat recognition.
- Do not publish or deploy the app, except for the explicitly authorized GitHub setup pushes described below. Pushing setup files does not authorize deploying the application.
- Preserve the existing documents and unrelated user changes in this folder.

The user currently wants code generation only and is checking the build and rendering themselves. **Do not launch the app, run builds, tests, browser checks, or external service calls unless explicitly requested.** Do not add unit tests without an explicit request. Report unperformed validation honestly.

## OpenAI Live hotline integration authorization

On 22 October 2026, the user explicitly requested a dedicated chat to implement OpenAI Live on the hotline: the server service, audio streaming, initial harness tools in Nitro, and an `.env.example` entry for the API key. This request supersedes the frontend-only restriction for this specific integration.

- Implement within the existing Nuxt monolith, in JavaScript only. Keep endpoints thin: endpoint -> core/use case -> repository or external service. All business orchestration belongs in cores.
- Use the current official OpenAI documentation to distinguish Live from Realtime and implement the supported protocol; do not silently substitute one API for the other. Surface missing model/account access honestly.
- Audio capture starts only after an explicit user action in the app. Keep the permanent API key and session configuration on the server; never put secrets in public runtime configuration, source files, client bundles or logs.
- Preserve existing database environment entries in `.env.example`; add empty OpenAI placeholders and document setup. Do not read, copy, commit or print actual credentials.
- Initial tools may retrieve the fictional customer, policy and dated incident context, collect a report and prepare document requests. Use existing repositories/data contracts where applicable. Do not add backend mocks or report successful persistence if storage is unavailable. Keep real coverage decisions, bookings and payments outside scope.
- Preserve Jules, the shared 80-household scenario, client/insurer state continuity, approved UI and flat EN/FR translation dictionaries. Keep mock mode frontend-only and explicitly distinguish it from the live voice mode.
- Do not apply database changes, reset data, provision paid resources, deploy, launch the application, run builds/tests/browser checks or make live API test calls unless separately requested. Installing dependencies needed for the implementation uses Bun. Report source changes and unperformed validation accurately.

This exception does not authorize unrelated backend work or changes to another agent's insurer implementation. Coordinate changes to shared composables and configuration files.

## Aiven setup and GitHub push authorization

The user explicitly authorizes preparing, committing, and pushing Aiven Runtime setup changes to the `main` branch of `https://github.com/Hodisy/disaster-network`. This standing authorization includes the Git network operations needed to inspect the remote branch and push those scoped commits; no additional confirmation is required for these setup pushes.

The exception covers `Dockerfile`, `.dockerignore`, Aiven deployment manifests, and directly related setup documentation, including this authorization in `AGENTS.md`. Keep commits limited to those changes. Do not include unrelated frontend work, other agents' changes, credentials, or local environment files. Do not force-push or rewrite remote history.

This exception does not authorize publishing unrelated application changes, triggering a deployment, creating paid resources, connecting databases, or implementing backend integrations. Build, test, and browser validation still require an explicit user request.

## Architecture

- Use Nuxt / Vue and JavaScript only. Do not write TypeScript manually; optional JSDoc is fine. Framework-generated files are not handwritten source.
- Keep presentation components stateless with respect to business data: props in, events out. Local DOM behavior such as focus handling and canvas animation belongs in the relevant component.
- Pages compose the screens and bind actions through composables. Client state/actions live in `app/composables/useDemo.js`; shared canonical claim data in `useClaimScenario.js`; insurer fixtures, derived totals, filters and simulated actions in `useInsurerWorkspace.js`. Preserve the public fixture identifiers in `app/config/demo.js`.
- Introduce mocks only in composables/hooks. Never fetch data or call services inside presentation components.
- Preserve coherent data/action boundaries so future frontend API adapters can replace the mocks without rewriting presentation components.
- Keep claim state consistent across navigation and account switching. Merge the shared client dossier into the reserved insurer record; never append an 81st household. Current state is in memory only; refresh/reset clears it. File previews remain browser-local. The insurer reset watcher must also clear linked resource simulation state when the client demo revision changes.
- Keep document collection, a guarantee proposal, organized assistance, and claim closure as distinct stages. Do not imply that collecting files proves coverage or that a mock action performs a real intervention.

## Localization

- Translation files are `app/locales/en.json` and `app/locales/fr.json`.
- Keys are complete English phrases. Both files must be flat JSON objects with string values: no nested objects, namespace keys, or dot-path lookup.
- Use `useAppI18n()` for literal `t(phrase, parameters)` lookup and date/time formatting. Add new phrases to both dictionaries, using the source phrase as its own English value.
- Keep canonical English mock content in state and translate it at display time, so switching languages updates existing replies and statuses without clearing the claim. Preserve user input verbatim.
- The language preference is separate from demo state and persisted in browser local storage when available. Resetting the demo must retain the selected language.
- The header toggle is a presentation component with props/events. Keep English routes unchanged.

## CSS and UI conventions

- Each Vue component owns its styles in its own `<style scoped>` block.
- Keep global CSS limited to the reset, base typography, and a minimal shared color palette.
- Use Flexbox for layout. Do not use CSS Grid.
- Use `px` or `%` for dimensions and spacing. Do not use `em`/`rem` or introduce spacing/sizing variables.
- Keep CSS variables minimal and color-focused.
- Focus on desktop. Do not add mobile-specific layouts or responsive breakpoints in this phase.
- The client sidebar has exactly two entries: **Hotline** and **My claim**, translated through the dictionaries when French is selected.
- Keep the account switcher as a small overlay anchored above the bottom sidebar account control. The insurer account currently leads to `/insurer`, and switching back retains the client claim.
- The hotline is centered on a clickable animated particle orb, with simulated transcription and replies below it. Do not turn it into a dense dashboard.
- Preserve the Thinking Orbs vendor code's MIT license and provenance. Use the Vue wrapper; do not introduce React for the orb.

## Handoff

Keep changes scoped to the user's request. Update the README when capabilities, setup, or limitations change. Distinguish implemented source from behavior that has actually been built or verified. The approved insurer frontend is implemented in source: map, right-panel Priority queue / All claims tabs, linked details, filters, mobilization, provider-outage replanning and local situation reports. Continue from user build/render feedback; do not regenerate concepts unless requested. Backend integrations remain outside this insurer task.

## Insurer map and operations

- Use **MapLibre GL JS 6.15.0**, installed with Bun, and the public **OpenFreeMap Positron** style. No API key or Vue map wrapper is required. `useOperationsMap.js` is the rendering adapter, receiving only local GeoJSON from the page. Keep the module worker bundled through Vite and dispose of the map/resize observer on unmount.
- The basemap requires internet access and WebGL. Keep its visible attribution and the fallback/retry state; the claim list remains available if it fails. Public basemap downloads are separate from business API integrations. Do not fetch incident, routing, weather or customer data.
- Map locations, groups and links are synthetic. Never draw or label them as an official flood perimeter, live affected properties or measured road routes. Keep the default extent around Berre Lagoon with Marseille as geographic context.
- Keep the map visible when switching the right panel between Priority queue, All claims and a selected claim detail. Situation / Mobilization switches the workspace view, with provider capacity below the map in mobilization.
- Derive totals from the records: household sizes 30×1, 25×2, 10×3, 15×4; situation groups A15 / B10 / C35 / D20. Keep portfolio totals distinct from filtered counts and opened reports. Unknown estimates stay unknown rather than becoming zero-valued losses.
- Physical missions require permitted access, a matching provider skill and free capacity. Unknown access may be changed only with an explicitly labelled simulation; Restricted access cannot be bypassed. Capacity units remain service-specific. Household accommodation uses household slots, not a booking per person.
- Replanning only moves disrupted missions to eligible capacity. Show the before/proposed provider, assumed time/cost change and unresolved blocked missions before applying the simulated plan. This does not approve coverage or create payments.
- Claim detail shows gross estimates, fictional caps, approved commitments and payments separately. Accommodation placement may resolve the unmet need and lower urgency; it must not make the home habitable or complete the document/coverage process.
- The situation report is created and downloaded locally from the full synthetic portfolio, independently of filters. It is not a real catastrophe loss report, reserve calculation or measured savings claim.
