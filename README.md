# Disaster Network

A desktop insurance-claim demo built with Nuxt 4 and Vue, entirely in JavaScript. The project was originally called CAT//RESPONSE; the current application branding is **Disaster Network**. **English is the default interface language, with a French option available through the EN / FR toggle in the header and in Settings.** Navigation, mock conversations, statuses, errors, accessibility labels, dates, and page titles follow the selected language. Source phrases, documentation, comments, and design mockups remain in English. Preserve proper names and real addresses.

## Current status — October 22, 2026

The approved client screens and map-first insurer workspace have been implemented in source with English and French interface dictionaries. The insurer uses an **80-household, 80-principal-claim synthetic portfolio representing 170 people**, including Jules exactly once. The user is handling build and rendering verification; no successful build or browser check is claimed. The [insurer scenario brief](docs/scenarios/insurer-workspace-brief.md) defines the current portfolio, fictional Allstorm benefits and Jules's case.

| Route | Current implementation |
| --- | --- |
| `/hotline` | Clean call workspace with the particle orb, chronological speaker captions, report review/save and a compact hang-up control. Mode and access setup live in Settings. Optional GPT-Live operation remains unverified after the latest source changes. |
| `/client` | Shared claim timeline and local previews. Live reports/categories are re-read from PostgreSQL through a verified browser grant; mock messaging/stage actions remain restricted to mock claims. No document upload. |
| `/insurer` | MapLibre regional map, clustered household markers, filters, Priority queue / All claims right-panel tabs, linked claim details, provider mobilization, simulated outage/replanning and a downloadable local situation report. |
| `/overview` | Insurer-focused marketing page, Repline-inspired layout, shared insurer map/list preview, Jules's southern France scenario, generated hotline illustration and five secondary catastrophe use cases. |
| `/settings` | Hotline mode selection and cookie-based voice unlock, confirmed local reset, and a separate key-protected database reset. The database reset was verified locally before the latest UI changes. |

The client navigation contains **Hotline** and **My claim**. The account-switching overlay uses the fictional Jules client account and the Allstorm Insurance insurer account. Jules's stable fixture identifiers are exported by `app/config/demo.js` for the frontend and Live repository; they are not authentication credentials. Returning to the client account retains the current claim. Local reset and database reset are separate controls in `/settings`; the shared shell has no demo badge or sidebar reset button. The insurer page and links use the English route `/insurer`.

The visual direction is a monochrome SaaS interface: a particle orb on the hotline, a claim activity stream, and a right-hand details rail. English versions of the approved mockups are saved in `docs/design/hotline-reference-v2.png` and `docs/design/client-reference-v2.png`. Design references retain the earlier fictional insurer name, Rivage; application copy uses Allstorm Insurance.

## Optional GPT-Live hotline

The source now connects the hotline to **GPT-Live**, using the official Live protocol rather than silently replacing it with Realtime. The permanent OpenAI key, session configuration, PostgreSQL verification and tool execution stay on Nitro. Microphone capture requires an explicit orb/start click. A separate, explicitly selected scripted mode remains available in Settings; a Live error never activates it. Settings also holds the AI/audio disclosure and access-key form. Unlocking creates the existing one-hour HttpOnly cookie; returning to the hotline or reloading checks a metadata-only access endpoint, without starting audio or an OpenAI session. No secret is embedded in the frontend.

See [the full Live setup and user-operated walkthrough](docs/openai-live-hotline.md) for the exact required/optional `.env` variables, TLS/database prerequisites, public demo verification challenge, registered tools, browser review/save, controlled 30-minute claim read grant and teardown limitations. Start with the empty entries in [.env.example](.env.example): private `OPENAI_API_KEY`, `NUXT_HOTLINE_ENABLED=true`, a separate `NUXT_HOTLINE_ACCESS_TOKEN`, the exact `NUXT_HOTLINE_ORIGIN`, and the existing PostgreSQL URL/CA. Defaults are `gpt-live-1`, delegated `gpt-6-luna` and voice `marin`; actual model/account access is unverified.

The harness finds Jules and contract `123456` in PostgreSQL, checks the stored scrypt challenge (public fixture answer `olivier`, maximum three attempts), then permits scoped customer/policy/claim tools. It reads actual text clauses/source references, prepares accommodation/assistance/document requests and saves only after an explicit on-screen confirmation. It updates `CAT-2026-001`, retaining report messages/events without another claim. The other 79 insurer households and all financial/resource scenarios remain frontend simulations. Documents are requested in My claim; previews remain local and never prove eligibility or trigger an intervention.

`ws` was added with Bun and the lockfile updated. No app, build, tests, browser, microphone, OpenAI call, DB connection, migration, reset or deployment was run for this integration. Model access, transport behavior, SQL execution and rendering still require the user's checks.

The hotline welcome now presents Allstorm Insurance and identifies the assistant as AI without repeatedly announcing the fictional scenario. Initial speech follows the UI language; subsequent replies are instructed to follow the caller's spoken language, including when the interface stays English. Greeting context is seeded at startup and appended once after `session.started`, following [OpenAI's session guide](https://developers.openai.com/api/docs/guides/live-conversations). This prompt behavior still needs a real-call check.

Captions retain exact incoming fragments and timing, group each speaker independently across approximate 1,400 ms gaps, and order the displayed groups on the session timeline. They are display groups, not provider-confirmed turns or a verified verbatim recording. The previous two clipped speaker blocks have been replaced; a bounded 10,000-fragment history displays an explicit notice if older fragments are removed. No text is invented to fill missing audio or transcript events. No build, browser check or paid call was run for this refinement.

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

## Aiven deployment preparation

The root `Dockerfile` and `.dockerignore` prepare the current mock frontend for Aiven Runtime using Bun for installation/build and Node 24 for production execution on port 8080. See [Aiven setup](docs/aiven-setup.md) for the GitHub and console steps. Optional hotline/database integration source has since been added; it has not been connected or tested. The container has not been built or run, and the application has not been deployed as part of this setup.

## Code organization

| Location | Responsibility |
| --- | --- |
| `app/app.vue` | Shared shell, route-aware account selection, overlay visibility, and reset wiring. |
| `app/pages/` | Screen composition, local UI state, and action bindings. |
| `app/components/` | Presentation components with props/events and their own scoped CSS. |
| `app/composables/useLiveHotline.js`, `useHotlineClaim.js` | Browser WebRTC lifecycle, private API requests, transcripts and the verified PostgreSQL claim projection. |
| `server/core/hotline/`, `server/repositories/hotlineClaims.js`, `server/services/openaiLive.js` | Live session/tool orchestration, scoped DB access and the private OpenAI transport. |
| `app/composables/useDemo.js` | Shared client state, mock asynchronous actions, document handling, and claim transitions. |
| `app/composables/useClaimScenario.js` | Canonical Jules record, benefit limits and claim stages shared by the client and insurer mocks. |
| `app/composables/useInsurerWorkspace.js` | The 80-record portfolio, provider/mission fixtures, filters, totals, GeoJSON and simulated operational actions. |
| `app/composables/useOperationsMap.js` | MapLibre rendering, sources, clusters, selection, controls, loading/retry and teardown. No business data fetching. |
| `app/composables/useAppI18n.js` | Reactive language selection, literal phrase lookup, interpolation, and locale-aware date/time/currency formatting. |
| `app/locales/en.json`, `app/locales/fr.json` | Flat translation dictionaries keyed by complete English phrases. |
| `app/components/LanguageToggle.vue` | Presentation-only EN / FR control; the app handles language selection. |
| `app/assets/css/main.css` | Global reset, base typography, and five color variables. |
| `app/vendor/thinking-orbs/` | Thinking Orbs 0.3.2 JavaScript rendering engine, used through a Vue canvas wrapper. Its MIT license and origin are preserved. |
| `docs/design/` | Approved client mockups, insurer concepts and the three approved right-panel layout states. |

Presentation components contain no data fetching or service calls. The pages bind the composables; mock behavior is introduced only at the composable boundary. MapLibre renders page-supplied local GeoJSON through a dedicated adapter; only the public basemap assets are downloaded.

## Frontend conventions

- JavaScript only; no handwritten TypeScript. JSDoc is optional.
- English source phrases, documentation, and comments. All interface copy uses flat translation dictionaries, with English as the default and French available from the header. Use `en-GB` or `fr-FR` date/time formatting according to the selected language, retaining the `Europe/Paris` time zone. Preserve proper names and real addresses.
- Each Vue component owns its CSS in a `<style scoped>` block.
- Use Flexbox for layout, not CSS Grid.
- Use pixels or percentages for dimensions and spacing, not `em` or `rem`.
- Keep CSS variables minimal and limited to shared colors; do not add spacing or sizing tokens.
- Desktop is the current target. Mobile layouts and responsive breakpoints are not part of this phase.

## Interface languages

Settings includes a dedicated Language section using the same persisted preference as the header. The production build passed locally on 22 October 2026 for the English-default deployment update; this does not validate paid voice calls or database operations.

The header EN / FR toggle changes the current screen without navigation, restarting the call, or clearing the claim. The language is shared across pages and accounts. English is the default on first use. The preference is stored separately under `disaster-network-language` in browser local storage, so it survives reloads and demo resets; if storage is unavailable, selection still works for the current session. Mock claim data and file previews remain in memory only; explicitly confirmed Live reports persist in PostgreSQL.

Translations are in `app/locales/en.json` and `app/locales/fr.json`. Both files are **flat JSON objects with string values**, using complete English phrases as literal keys, including punctuation. There are no nested objects, namespace prefixes, or dot-path lookups. For example, the French dictionary contains:

```json
{
  "My claim": "Mon dossier",
  "Open my claim": "Ouvrir mon dossier",
  "Add: {document}": "Ajouter : {document}"
}
```

Use `const { t } = useAppI18n()` and `t('My claim')` at the presentation boundary. Named placeholders use `t('Add: {document}', { document: t(row.title) })`. Add each source phrase to both dictionaries; the English value repeats its key. Missing translations fall back to English, then to the source phrase. No translation dependency or service is used.

Mock state retains canonical English phrases; scripted replies/transcripts, statuses and errors are translated when displayed, including content already received before switching languages. Live speech/transcripts and confirmed caller content stay verbatim; the toggle changes the interface and future call language. User-written messages, drafts, filenames, proper names, addresses, and identifiers are preserved verbatim. Claim dates and message timestamps are stored in ISO format and rendered using the selected locale. Document language and the HTML `lang` attribute update with the toggle. Language selection does not create translated routes or alter the underlying scenario.

## Mock behavior and limitations

The approved scenario is a fictional Allstorm household reporting on **22 October 2026** about the dated flood event of **20 October 2026**. The reference client is **Jules**, an owner-occupier living with his partner and one child, included within the 80-household portfolio. His gross synthetic estimate is EUR 17,500 across accommodation, essential items, electrical safety, cleaning/drying, moving/storage, contents and building repairs. It is not an approved indemnity. The client source now uses Jules and separates the event date (`date`, 20 October) from the report date (`reportedAt`, 22 October). The shared record now includes all seven benefit lines; their estimates, fictional caps, approved commitments and payments are visible in insurer claim detail. Initial approved and paid amounts are both EUR 0. Guarantees, assistance and outcomes remain demonstration content without contractual value.

State is shared in memory while navigating between pages and accounts. Refreshing the page or resetting the demo clears browser mock state. Persisted Live reports survive and can be re-read through an unexpired verified browser grant. Selected documents are kept in the current browser tab using object URLs; they are not uploaded or stored on a server. The demo accepts JPG, PNG, WebP, and PDF files, with a 10 MB per-file limit, a 30 MB total limit, and up to 12 files. In mock mode, one file in each requested category completes document collection; this does not represent real review. Local previews never advance a persisted Live claim or count as uploaded evidence.

The claim stages remain distinct: **Collecting documents → Documents complete → Coverage proposal → Assistance arranged → Claim closed**. Later transitions are explicit simulated actions, with no real booking, coverage decision, or payment.

The scripted claim journey remains frontend-only. The separately authorized optional GPT-Live path now has server session control, PostgreSQL verification/reads, and confirmed report persistence. No real weather, production authentication, upload, booking, payment or coverage decision is implemented. A separately authorized maintenance-only Nitro endpoint now implements database reset; it is disabled until configured. The nine-table PostgreSQL schema and small fictional dataset have now been applied to the existing Aiven service. The maintenance reset endpoint is configured locally. The scripted journey uses frontend mocks; optional Live intake uses the separate adapters described above, whose runtime behavior has not been checked in this task. Nuxt is configured with `ssr: false`.

## Validation and next work

The current workflow is code generation only. The user is handling the build and visual review. No successful build or browser validation is claimed, and no unit tests have been added. Do not run the app, builds, tests, or service calls unless the user explicitly requests them.

Next work is to incorporate the user’s build/render feedback on the implemented insurer workspace. Its map focuses on Berre Lagoon northwest of Marseille. The October 20, 2026 flood is dated context only; claims, priority classifications, provider availability, costs and replanning outcomes are explicitly synthetic. Operational urgency remains separate from claim progress and coverage review.

Future integrations can replace the Promise-based methods in `createMockGateway()` inside `useDemo.js`. Local document selection and preview handling are isolated in `addFiles()` and `removeFile()`. Keep the presentation component props/events stable when those integrations are introduced.

## Insurer workflow and design brief

The proposed workflow is inspired by a mutual insurer's needs. MAIF's public customer journey describes emergency assistance, accommodation, damage assessment, and settlement according to the situation and contract. This demo does not reproduce MAIF's internal procedures or imply any affiliation. [MAIF claim journey](https://content.maif.fr/public/fiche-sinistre/parcours-catastrophe-naturelle-habitation.html)

| Operational decision | Business impact to consider | Proposed interface support |
| --- | --- | --- |
| Which households need help first? | Assistance delays and care for displaced or vulnerable policyholders. | Explainable priority, unmet needs, contact status, and waiting time. |
| Where should available crews be assigned? | Limited provider capacity and geographic coordination. | One map linking claim clusters, needs, and available providers. |
| Which losses need early assessment or damage mitigation? | Further damage, assessment backlog, and uncertainty about claim costs. | Habitability, reported damage, access information, documents, and assessment status. |
| What happens when a provider becomes unavailable? | Missed appointments, extra assistance costs, and delayed support. | Affected claims, a proposed replacement, delay/cost assumptions, and explicit simulated approval. |
| What financial exposure should be anticipated? | Reserving, funding needs, and potential reinsurance recoveries. | Separate gross estimates, uncertainty, unpaid obligations, payments, and any future treaty-dependent net view. |

These are proposed product decisions, not measured benefits or an insurer's official policy. CCR describes both the financial assessment of catastrophe events and the role of reinsurance; the demo does not calculate actual reserves or recoveries. [CCR catastrophe reinsurance](https://www.ccr.fr/decouvrir-ccr/activites/catastrophes-naturelles/)

The suggested classification is **P1 Urgent / P2 Priority / P3 Standard**, based on declared habitability, vulnerability, unmet assistance, and waiting time, with reasons visible per claim. It is a demo rule, not a MAIF classification. Keep priority separate from estimated claim value, coverage review, file completeness, and operational assignment. A complete document set is not an accepted guarantee, and an assigned crew is not claim closure.

The implemented map frames **Berre Lagoon, Istres, Saint-Chamas, and Martigues**, with Marseille as southeastern context. The City of Istres reports a major event on October 20, 2026, particularly west of the lagoon. All displayed portfolios, provider locations, affected-area overlays, availability, cost assumptions, and travel/delay estimates remain fictional. [City of Istres event statement](https://www.istres.fr/index.php?id=229&tx_news_pi1%5Bnews%5D=13451&tx_news_pi1%5Bcontroller%5D=News&tx_news_pi1%5Baction%5D=detail)

Three English ImageGen concept proposals document the design exploration:

- [A — Map-first crisis desk](docs/design/insurer-option-a-map-first.png): a large regional map and an urgent-needs queue.
- [B — Claim investigation desk](docs/design/insurer-option-b-claim-desk.png): a linked claim list, map, and open claim detail panel.
- [C — Mobilization and replanning](docs/design/insurer-option-c-mobilization.png): crew assignments and a provider-outage replacement preview.

The selected direction is A for the main situation view, with C informing mobilization. The right panel switches between Priority queue, All claims and a selected claim detail while the map stays visible. Figures in the images are historical layout assumptions, including older 200-claim references; source totals are derived from the approved 80-record fixture. Use the [scenario brief](docs/scenarios/insurer-workspace-brief.md) for the 80 households, 170 people, situation groups, benefit limits and Jules's case. Keep map geometry and locations accurate in implementation; ImageGen maps are layout concepts, not usable geographic data.

Contributor instructions are in [AGENTS.md](AGENTS.md). Preserve the existing reference and setup documents in this folder.


## Insurer workspace: implemented source

Open `/insurer` directly or use the account overlay above the bottom-left account button. The default Situation view preserves the large map and the right-panel **Priority queue / All claims** tabs. Select a marker for a compact preview; open it or a list row to show dossier details in the same panel. The All claims list uses eight records per page. Search and municipality, priority, need, stage, situation-group and access filters drive both markers and records. Map counts identify the filtered records; portfolio summaries and reports cover all 80 households.

The fixture contains 30 one-person, 25 two-person, 10 three-person and 15 four-person households, totalling 170 people. Situation groups contain A15 / B10 / C35 / D20. Jules’s reserved `CAT-2026-001` is included once: before the hotline report the portfolio has 79 opened reports; opening the client claim updates that record to 80. Its documents and claim stage stay shared across account navigation. Physical damage remains unknown for inaccessible group B records; incomplete estimates are labelled partial rather than treated as a complete zero-valued loss.

Claim detail separates household needs, habitability, access, document progress, coverage review, statutory CatNat status, seven budget lines, fictional benefit caps, approved commitments and paid amounts. Jules starts at P1 because his family needs lasting accommodation. Assigning one household accommodation slot resolves that unmet need and moves him to P2; it does not repair the home, validate coverage or close the claim. Unknown access can be changed by an explicitly labelled simulation. Restricted properties cannot be scheduled.

The Mobilization view adds provider capacity and assignment details below the map. Six drying teams have two concurrent sites each; other providers use visits per day, household storage slots or household accommodation slots. Selecting a provider shows its missions. **Plan assistance** creates draft physical missions for the selected record’s outstanding needs; scheduling uses matching skills, free capacity and permitted access. Accommodation placement is a separate household action.

**Simulate provider outage** makes Delta Drying unavailable and disrupts its two seeded assignments. The replacement preview considers eligible free capacity, reserves slots across the proposed replacements and shows assumed start-delay and mission-cost changes. Applying the plan only moves affected missions; blocked properties remain blocked, and approved/paid totals do not change. Replacement timing and cost deltas are fixture inputs, not routing calculations, quotes or measured savings. Reset resource simulation restores provider and mission fixtures; the global demo reset also clears insurer overrides and local client state. Both are distinct from the separately configured database-reset operation.

The Situation report dialog summarizes the full synthetic portfolio, unmet accommodation needs, blocked missions, partial known estimates and zero initial approved/paid amounts. Download creates a text file in the browser without an export service.

### Map dependency and public basemap

The only added map library is **MapLibre GL JS 6.15.0**, installed using Bun. The map uses [OpenFreeMap Positron](https://openfreemap.org/quick_start/) without an API key, bundled module worker, clustered GeoJSON points and local assignment lines. Existing Lucide icons are reused; no Vue map wrapper, routing library or charting dependency was introduced. See [MapLibre documentation](https://maplibre.org/maplibre-gl-js/docs/).

The basemap requires **internet access and WebGL** to load its public style, tiles, fonts and sprites; attribution remains visible. Claim/provider data stays in browser memory. Map errors show a retry state while the dossier list remains usable. Markers are invented approximate locations, and assignment lines are straight illustrative links: there is no live flood layer, official hazard boundary, geolocation, customer-data request or route calculation. Keep the service’s attribution when changing the styling.

The insurer work is frontend-only. It does not connect to the separately prepared database, add endpoints or start a real voice session. No application launch, build, test, browser review or live business API call was performed for this implementation.

## Event research

[Southern France flooding on 20 October 2026](docs/research/20-october-2026-flood-evidence.md) separates sourced event facts, attributed impact counts, insurance procedures and a proposed synthetic Allstorm scenario. The research snapshot is dated 22 October 2026; it does not add live integrations or change the application.

## Demo database preparation

The initial database preparation provided a simplified contract and a small dataset. A later explicit authorization adds the optional Live claim intake/verification endpoints and controlled PostgreSQL projection described in [the Live guide](docs/openai-live-hotline.md). A subsequently requested settings page adds a narrowly scoped maintenance reset endpoint, described below. See [the data contract](docs/data-contract.md) for all fields, JSON shapes and reset behaviour. There is no login or sessions table. Each customer has at most one policy, enforced by a unique foreign key; each seeded customer has one.

The offline-generated dataset contains three fictional customers (Jules, Camille and Sam), three policies, four plain-text contractual fixtures, eleven search sections, three reserved/reported claims, four messages, three events, three providers and no assignments. It is a small development fixture, not the complete 80-household portfolio. Jules has a reserved `CAT-2026-001` record awaiting the hotline report; the Live core updates it instead of creating a duplicate. No real insurance PDF has been imported.

With Bun, a local `psql` client, the existing Aiven connection details and its CA certificate:

```sh
# Entirely offline: regenerate the committed JSON and SQL fixtures.
bun run db:export

# Reads the ignored .env; creates disaster_demo and inserts missing fixtures.
bun run db:setup

# Inserts missing fixture IDs only; preserves existing demo progress.
bun run db:seed

# Deletes all records/files in disaster_demo, then restores the small fixture.
bun run db:reset --confirm-demo-reset
```

Copy `.env.example` to `.env` and set `DATABASE_URL` and `DATABASE_CA_CERT_PATH` before database commands. No credentials belong in the frontend. Setup is a manual CLI operation. Reset is available through the CLI or the explicitly enabled, maintenance-key-protected settings endpoint. Reset is scoped to the marked `disaster_demo` schema and runs transactionally; it does not drop the database or touch `public`. The UI reset button still clears browser memory only. The deployed Node image does not include the CLI's `psql` dependency. The settings endpoint uses the `pg` Node driver instead.

Offline fixture export and `db:setup` were executed. A verified-TLS read-only check confirmed nine Aiven tables, three customers, three policies, three claim records, four documents and eleven sections. After explicit user authorization, Nuxt was launched locally and the Settings reset succeeded against Aiven: the browser displayed three restored customers, policies and claim records. Shared database modules are explicitly bundled by Nitro to retain valid imports. No production build or unit tests were run.

## Discreet settings page

`/settings` is accessible through the small settings icon in the header, without adding another client sidebar entry. Its translated EN/FR panel asks for confirmation and a maintenance key before requesting a database reset. No login or sessions table is introduced. A hidden link is not an authorization boundary: the server requires an explicit feature flag and a private key.

The flow is page -> `useDemoSettings()` -> `POST /api/demo/reset` -> `server/core/resetDemoDatabase.js` -> `server/repositories/demoDatabase.js`. The endpoint performs a real reset when configured; it never pretends that a browser-only reset changed Aiven. On confirmed success, the browser mock state is also reset, which clears its local previews and resets any linked insurer mock state. On an error, browser claim state is preserved. The language preference survives reset.

Required private server settings:

- `NUXT_DEMO_RESET_ENABLED=true` to explicitly enable this maintenance operation; default is false.
- `NUXT_DEMO_RESET_TOKEN`: a random key of 32-256 characters, entered manually in the settings confirmation. It is not exposed through public runtime config or stored in browser storage. Existing API credentials are not this key.
- `DATABASE_URL` or `NUXT_DATABASE_URL`: the existing PostgreSQL connection details.
- `DATABASE_CA_CERT_PATH` / `NUXT_DATABASE_CA_CERT_PATH`, or full PEM through `DATABASE_CA_CERT` / `NUXT_DATABASE_CA_CERT`, for verified TLS to Aiven. Certificate files are not bundled automatically; use a runtime secret if no file is mounted.

Run `bun run db:setup` manually once before using reset. The endpoint deliberately cannot create or repair a schema. It validates the ownership/version marker and uses the same fictional seed and reset SQL as the CLI, with one database client and one transaction. There is no `DROP`, `TRUNCATE`, `CASCADE`, public-schema reset or user-supplied table target. Database reset removes all demo records and files, including records created after seed. Raw database errors and credentials are not sent to the browser.

The page and endpoint are implemented, with the local ignored `.env` configured for the existing Aiven PostgreSQL service, verified TLS and enabled maintenance reset. Setup, read-only database checks and the subsequently authorized local Settings reset succeeded. An earlier launch rejection was resolved by explicit user authorization. The `pg` dependency was installed with Bun with lifecycle scripts disabled and `bun.lock` updated. The optional fictional-customer Live intake APIs are implemented separately; no real customer integration is included.

### Aiven data load verified - 22 October 2026

The existing `pg-3351099e` service contains the `disaster_demo` schema with its version marker and the small fixture. Jules has contract `123456` and reserved claim `CAT-2026-001`; the reference-data refresh preserves its current intake status. Database setup did not modify `public`. Connection details and a random maintenance key are stored only in the ignored `.env` (restricted file permissions); the CA certificate is in ignored `.data/aiven-ca.crt`. No credentials are committed or exposed in public Nuxt configuration.

The local server loaded the environment and the authorized Settings reset succeeded. Hosted Runtime environment changes or deployment have not been performed by this database-load task. The maintenance key is `NUXT_DEMO_RESET_TOKEN` in the local `.env`; it is not printed in documentation or chat. Git pushes transfer code and safe fixture/migration sources, not the running database or local secrets.

## Project overview landing page

`/overview` is a standalone desktop marketing page for **insurers**, with the existing EN/FR toggle. Its Repline-inspired presentation uses centered regular-weight typography, white space, black pill links and a pale-grey product band. It leads with the insurer interface and the synthetic **80 households / 170 people**, then follows Jules's fictional flood claim near Istres: accommodation for three people, an AI hotline, evidence collection in My claim, coverage review and proposed assistance. Five secondary applications remain: Los Angeles wildfire, Brittany windstorm, southwest France hailstorm, Mediterranean earthquake and French Alps landslide.

The five additional scenarios are editorial illustrations, not separate implemented workflows, historic incident reports or extensions of the fictional Allstorm flood guarantees. Visible labels distinguish generated imagery and synthetic portfolio figures. No live loss counts, insurer endorsement, measured savings or actual bookings are claimed.

The hero, product preview, journey, use cases and closing section are presentation components (`Marketing*.vue`). The preview reuses the existing `InsurerMap` and `InsurerClaimList` components; it does not display an invented AI interface or an AI-generated geographic map. `useMarketingWorkspacePreview.js` reads the existing synthetic insurer portfolio and manages local preview selection, tabs and pagination. It preserves the public OpenFreeMap attribution and the existing WebGL/network requirements and error fallback. No claim, coverage or assignment is changed by preview interactions. PostgreSQL and the voice service are not queried by the landing page.

The hero now embeds the approved **A — Map-first crisis desk** composition: branded sidebar, operations heading, full-width map and right-hand priority queue. Jules is selected initially. The existing toolbar provides search and filters; Situation / Mobilization, provider inspection, map selection, popup dismissal and list pagination are frontend preview interactions. The popup opens the full insurer workspace rather than promising to open a particular claim detail. No assistance, coverage, outage or dispatch action is triggered from the preview.

The hero keeps the existing OpenFreeMap Positron geometry and attribution, with a local `reference-a` appearance that recolours water and parks toward the reference image's light-blue/green palette. It does not use the ImageGen map as geographic data and adds no official affected-area boundary. The normal insurer map retains its default appearance. The existing `public/brand/disaster-network-mark.svg` is reused next to the landing wordmark and in the footer, and declared as the SVG favicon in the shared Nuxt head; the preview sidebar uses the existing full wordmark asset. No icon-generation dependency is needed. These source updates have not been launched, built or reviewed in a browser.

Editorial fixtures live in `useMarketingContent.js`. The updated Jules photograph uses a real Saint-Chamas image as a regional architecture reference; the output remains a labelled fictional illustration, not a photograph of the incident. A small flat-style service illustration shows a customer holding a phone, a damaged roofless house behind him and a robot wearing a hotline headset. It is a generic service metaphor and does not change Jules's flood damage. The five earlier supporting photographs are preserved; delivery assets are WebP, with original generated PNGs retained separately. See [the active art direction and exact final prompts](docs/design/marketing-insurer-direction.md) and [initial imagery](docs/design/marketing-image-prompts.md).

Native links open `/hotline` or `/insurer`; section links navigate within the landing page. The existing `/` redirect, sidebar entries, account switching and client/insurer state are retained. Text, dialogue and controls use the EN/FR dictionaries; no page UI is baked into an image. No dependency, backend integration or deployment was added. Web research and requested image generation were performed, but no app launch, build, test or local browser/render check was run for this update.

## Realistic scenario data

The operational screens use natural claims language and plausible synthetic records. Jules's policy is **123456**, product **Allstorm Home Protect**, 2026 edition. Camille and Sam use **123457** and **123458**. The security-word question is “What is your security word?” and its public evaluation answer is **olivier**. The expected answer is checked on the server; never embed it in frontend configuration or a voice prompt. These details identify a fictional evaluation scenario, not real policyholders or insurance cover.

The 80-household / 170-person portfolio uses invented first names, varied reports, distinct contact times and severity-dependent estimates. Main screens and voice replies avoid repeated demonstration jargon. Evaluation provenance remains in Settings, source/data metadata, a discreet workspace footer and exported reports. Proposals, site-access constraints, document previews, approvals and payments retain their actual meaning. No new booking, payment, file-upload or evidence-review capability is implied. Historical design images retain their older copy.

`bun run db:export` regenerates the local JSON/SQL fixtures without contacting PostgreSQL. Existing installed data requires a targeted reference-data refresh: ordinary `db:seed` only inserts missing rows and does not rename an existing policy. The authorized refresh preserves Jules's report and all customer-written content; it does not reset the dataset.

### Reference-data refresh applied — 22 October 2026

The user explicitly authorized the corresponding Aiven update. A transaction refreshed three customer challenges, three policy schedules, four controlled contractual text documents, eleven source sections, three provider names and two unchanged seeded assistant messages. The security word is now **olivier** and Jules's policy is **123456**. Before/after database fingerprints confirmed that claims, claim events, assignments, customer messages and non-policy attachments were unchanged. No reset or schema migration occurred. Only unchanged source-fixture policy documents were eligible; an edited source would have aborted the transaction.

The offline JSON/SQL export was regenerated. The frontend, EN/FR dictionaries and agent instructions were reviewed in source; no build, tests, browser check, app restart or OpenAI call was run for this pass. Reload the UI and begin a new call to discard an older in-memory conversation or portfolio.
