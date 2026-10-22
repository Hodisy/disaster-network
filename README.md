# Disaster Network

A desktop insurance-claim demo built with Nuxt 4 and Vue, entirely in JavaScript. The project was originally called CAT//RESPONSE; the current application branding is **Disaster Network**. The interface remains in French.

## Current status — October 22, 2026

The two approved client mockups have been implemented in source code. The insurer dashboard is deferred.

| Route | Current implementation |
| --- | --- |
| `/hotline` | Animated particle orb, three scripted exchanges, simulated transcription and replies, interruption/resume, and claim creation. No microphone access or real voice session. |
| `/client` | The same claim, an activity timeline, requested documents, local photo/PDF previews, simulated messaging, and explicit claim-stage transitions. An empty state directs the user to the hotline before a claim exists. |
| `/assureur` | A “coming later” placeholder only. No operational dashboard or synthetic portfolio has been built. |

The client navigation contains **Hotline** and **Mon dossier**. An account-switching overlay at the bottom of the sidebar switches between the fictional Claire Martin client account and the Rivage Assurance insurer account. Returning to the client account retains the current claim. A reset control restarts the demo.

The visual direction is a monochrome SaaS interface: a particle orb on the hotline, a claim activity stream, and a right-hand details rail. The approved mockups are saved in `docs/design/hotline-reference-v2.png` and `docs/design/client-reference-v2.png`.

## Local development with Bun

**Use Bun, not npm, for dependency management and project commands.** Install Bun before using the commands below. Nuxt currently expects a compatible Node runtime; Node 24.15 or later is recommended, and a local Node 24.15 runtime is declared in the development dependencies.

```sh
bun install
bun run dev
```

Open the address printed in the terminal, normally [http://127.0.0.1:3000/hotline](http://127.0.0.1:3000/hotline). `/` redirects to `/hotline`.

To build and preview the production output, stop the development server first:

```sh
bun run build
bun run preview
```

`bun.lock` is the project lockfile and is committed with the source. The legacy local `package-lock.json` is ignored. Use Bun for subsequent dependency changes; do not generate or update an npm lockfile.

## Code organization

| Location | Responsibility |
| --- | --- |
| `app/app.vue` | Shared shell, route-aware account selection, overlay visibility, and reset wiring. |
| `app/pages/` | Screen composition, local UI state, and action bindings. |
| `app/components/` | Presentation components with props/events and their own scoped CSS. |
| `app/composables/useDemo.js` | Shared frontend state, scenario data, mock asynchronous actions, document handling, and claim transitions. |
| `app/assets/css/main.css` | Global reset, base typography, and five color variables. |
| `app/vendor/thinking-orbs/` | Thinking Orbs 0.3.2 JavaScript rendering engine, used through a Vue canvas wrapper. Its MIT license and origin are preserved. |
| `docs/design/` | The two approved interface mockups. |

Presentation components contain no data fetching or service calls. The pages use the composable; mock behavior is introduced only at the composable boundary.

## Frontend conventions

- JavaScript only; no handwritten TypeScript. JSDoc is optional.
- Each Vue component owns its CSS in a `<style scoped>` block.
- Use Flexbox for layout, not CSS Grid.
- Use pixels or percentages for dimensions and spacing, not `em` or `rem`.
- Keep CSS variables minimal and limited to shared colors; do not add spacing or sizing tokens.
- Desktop is the current target. Mobile layouts and responsive breakpoints are not part of this phase.

## Mock behavior and limitations

The scenario is a fictional flood in Istres on October 22, 2026. Claire Martin, Rivage Assurance, guarantees, assistance, and any outcomes are demonstration content, not live information or contractual commitments.

State is shared in memory while navigating between pages and accounts. Refreshing the page or resetting the demo clears the session. Selected documents are kept in the current browser tab using object URLs; they are not uploaded or stored on a server. The demo accepts JPG, PNG, WebP, and PDF files, with a 10 MB per-file limit, a 30 MB total limit, and up to 12 files. One file in each requested category completes the document collection; this does not represent real document review.

The claim stages remain distinct: **À constituer → Dossier complet → Proposition de garantie → Assistance organisée → Sinistre clôturé**. Later transitions are explicit simulated actions, with no real booking, coverage decision, or payment.

No business backend, Nitro API endpoints, database, Aiven connection, real AI/weather call, real authentication, or microphone capture is implemented. Nuxt is configured with `ssr: false`.

## Validation and next work

The current workflow is code generation only. The user is handling the build and visual review. No successful build or browser validation is claimed, and no unit tests have been added. Do not run the app, builds, tests, or service calls unless the user explicitly requests them.

Next steps are to incorporate the user's build/render feedback and, when requested, implement the insurer space. Future integrations can replace the Promise-based methods in `createMockGateway()` inside `useDemo.js`. Local document selection and preview handling are isolated in `addFiles()` and `removeFile()`. Keep the presentation component props/events stable when those integrations are introduced.

Contributor instructions are in [AGENTS.md](AGENTS.md). Preserve the existing reference and setup documents in this folder.
