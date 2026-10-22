# Agent instructions

## Project status

Disaster Network is a Nuxt 4 / Vue frontend demo for insurance-claim coordination. Its original name was CAT//RESPONSE; preserve the current Disaster Network branding. The UI is in French; project documentation is in English.

As of October 22, 2026, source code exists for the two approved client screens: `/hotline` and `/client`. `/assureur` is deliberately a placeholder. The insurer dashboard, its portfolio, filters, resource planning, and provider-reallocation demo are deferred until requested.

Read `README.md` for current capabilities and limitations. The approved visual references are `docs/design/hotline-reference-v2.png` and `docs/design/client-reference-v2.png`. Preserve their monochrome SaaS direction, animated particle orb, activity timeline, and details rail.

## Package manager

- Use **Bun**, not npm, pnpm, or Yarn.
- Commands: `bun install`, `bun run dev`, `bun run build`, `bun run preview`.
- Commit and maintain `bun.lock` when dependencies change. The legacy local `package-lock.json` is ignored; do not generate or update npm lockfiles. Do not run dependency installation solely to edit documentation.

## Current authorized scope

- Frontend code, components, composables/hooks, CSS, and mocks only.
- Do not implement business backend code, Nitro API endpoints, server repositories, databases, or Aiven connections in this phase.
- Do not add real AI/weather calls, microphone capture, voice sessions, uploads, authentication, bookings, or payments.
- Keep all identities, guarantees, and outcomes clearly fictional. The dated Istres flood scenario is not live information.
- Do not publish or deploy the app.
- Preserve the existing documents and unrelated user changes in this folder.

The user currently wants code generation only and is checking the build and rendering themselves. **Do not launch the app, run builds, tests, browser checks, or external service calls unless explicitly requested.** Do not add unit tests without an explicit request. Report unperformed validation honestly.

## Architecture

- Use Nuxt / Vue and JavaScript only. Do not write TypeScript manually; optional JSDoc is fine. Framework-generated files are not handwritten source.
- Keep presentation components stateless with respect to business data: props in, events out. Local DOM behavior such as focus handling and canvas animation belongs in the relevant component.
- Pages compose the screens and bind actions through composables. Shared state and mock actions currently live in `app/composables/useDemo.js`.
- Introduce mocks only in composables/hooks. Never fetch data or call services inside presentation components.
- Preserve coherent data/action boundaries so future frontend API adapters can replace the mocks without rewriting presentation components.
- Keep claim state consistent across navigation and account switching. Current state is in memory only; refresh/reset clears it. File previews remain browser-local.
- Keep document collection, a guarantee proposal, organized assistance, and claim closure as distinct stages. Do not imply that collecting files proves coverage or that a mock action performs a real intervention.

## CSS and UI conventions

- Each Vue component owns its styles in its own `<style scoped>` block.
- Keep global CSS limited to the reset, base typography, and a minimal shared color palette.
- Use Flexbox for layout. Do not use CSS Grid.
- Use `px` or `%` for dimensions and spacing. Do not use `em`/`rem` or introduce spacing/sizing variables.
- Keep CSS variables minimal and color-focused.
- Focus on desktop. Do not add mobile-specific layouts or responsive breakpoints in this phase.
- The client sidebar has exactly two entries: **Hotline** and **Mon dossier**.
- Keep the account switcher as a small overlay anchored above the bottom sidebar account control. The insurer account leads to its placeholder, and switching back retains the client claim.
- The hotline is centered on a clickable animated particle orb, with simulated transcription and replies below it. Do not turn it into a dense dashboard.
- Preserve the Thinking Orbs vendor code's MIT license and provenance. Use the Vue wrapper; do not introduce React for the orb.

## Handoff

Keep changes scoped to the user's request. Update the README when capabilities, setup, or limitations change. Distinguish implemented source from behavior that has actually been built or verified. Continue to defer backend integrations and the insurer dashboard until explicitly requested.
