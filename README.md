# Disaster Network

Disaster Network connects insurance claims intake, evidence review and provider coordination in one Nuxt 4 / Vue application, written in JavaScript. It combines an AI voice hotline, a shared client dossier and a map-centred insurer operations workspace. English and French are available through the interface language toggle.

## Current capabilities — October 22, 2026

| Route | Current implementation |
| --- | --- |
| `/hotline` | Voice workspace with GPT-Live streaming, server-side customer and policy verification, dossier tools, report review and explicit save. |
| `/client` | Shared claim timeline, verified-dossier photo/PDF collection, optional AI evidence reading, deterministic policy comparisons, downloadable policy PDF and a Géorisques check. |
| `/insurer` | Portfolio-wide urgency, cost and capacity summaries; map and service needs; provider mobilization and replacement planning; on-demand AI classification and operation proposals with human review. |
| `/overview` | Insurer-focused product page, reference A map illustration, Jules's southern France journey and five further response scenarios. |
| `/settings` | Language, voice mode and access controls, insurer agent access, local workspace reset and a separately protected database reset. |

### Data and verification

The reference portfolio contains **80 synthetic household records representing 170 people**, including Jules once. The additive insurer preparation also provides 80 stored dossiers and their policy PDFs in PostgreSQL; the original maintenance/reset fixture remains smaller. Allstorm is the scenario insurer. The [scenario brief](docs/scenarios/insurer-workspace-brief.md) documents identities, contractual assumptions and data provenance. These are development/evaluation records, not real policyholder or catastrophe-impact data.

The operations page starts with local records. Database loading and AI analysis are explicit actions; the existing server harness reads stored sources, proposes classifications and operations, and records human-reviewed decisions. Capacity calculations and replacement planning execute on the supplied records. External dispatch, hotel bookings and payments are separate integrations. The sections below document configuration and implementation evidence. The production build passed during the generic-prompt release on 22 October 2026. No paid AI/voice call or complete browser journey was run for this release; no runtime AI result is claimed.

The client navigation contains **Hotline** and **My claim**. The account-switching overlay uses the fictional Jules client account and the Allstorm Insurance insurer account. Jules's stable fixture identifiers are exported by `app/config/demo.js` for the frontend and Live repository; they are not authentication credentials. Returning to the client account retains the current claim. Local reset and database reset are separate controls in `/settings`; the shared shell has no demo badge or sidebar reset button. The insurer page and links use the English route `/insurer`.

The visual direction is a monochrome SaaS interface: a particle orb on the hotline, a claim activity stream, and a right-hand details rail. English versions of the approved mockups are saved in `docs/design/hotline-reference-v2.png` and `docs/design/client-reference-v2.png`. Design references retain the earlier fictional insurer name, Rivage; application copy uses Allstorm Insurance.

## Optional GPT-Live hotline

The source now connects the hotline to **GPT-Live**, using the official Live protocol rather than silently replacing it with Realtime. The permanent OpenAI key, session configuration, PostgreSQL verification and tool execution stay on Nitro. Microphone capture requires an explicit orb/start click. A separate, explicitly selected scripted mode remains available in Settings; a Live error never activates it. Settings holds the AI/audio disclosure. At the user’s explicit request, voice calling is public when enabled: no hotline key is required locally or remotely. The metadata-only access endpoint automatically creates or reuses a separate one-hour HttpOnly browser cookie without starting audio or an OpenAI session. Expiry or server restart automatically creates fresh browser admission, but never restores customer verification or dossier grants. No secret is embedded in the frontend.

See [the full Live setup and user-operated walkthrough](docs/openai-live-hotline.md) for the exact required/optional `.env` variables, TLS/database prerequisites, public demo verification challenge, registered tools, browser review/save, controlled 30-minute claim read grant and teardown limitations. Start with the empty entries in [.env.example](.env.example): private `OPENAI_API_KEY`, `NUXT_HOTLINE_ENABLED=true`, the exact `NUXT_HOTLINE_ORIGIN`, and the existing PostgreSQL URL/CA. Defaults are `gpt-live-1`, delegated `gpt-6-luna` and voice `marin`; actual model/account access is unverified.

The harness finds Jules and contract `123456` in PostgreSQL, checks the stored scrypt challenge (public fixture answer `olivier`, maximum three attempts), then permits scoped customer/policy/claim tools. It reads actual text clauses/source references, prepares accommodation/assistance/document requests and saves only after an explicit on-screen confirmation. It updates `CAT-2026-001`, retaining report messages/events without another claim. The legacy portfolio preview remains a frontend simulation. The separately authorized insurer agent now reads all 80 stored dossiers and actual PDF bytes; provider fulfilment and payments remain simulated. In a verified Live dossier, supporting documents can now be saved and optionally read through the separate [evidence workflow](docs/evidence-workflow.md). Reading never proves eligibility or triggers an intervention. Scripted previews remain local.

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

With a loopback `NUXT_HOTLINE_ORIGIN`, voice, document and insurer origin checks also accept `localhost`, `127.0.0.1` and `[::1]` on the same protocol and port. Public origins still require an exact match. Cookies remain separate for each hostname, so changing hostname starts a separate browser call context and requires policy verification again for dossier access. This origin-alias change has not been checked with a build or browser call.

To build and preview the production output, stop the development server first:

```sh
bun run build
bun run preview
```

`bun.lock` is the project lockfile and is committed with the source. The legacy local `package-lock.json` is ignored. Use Bun for subsequent dependency changes; do not generate or update an npm lockfile.

## Aiven deployment preparation

The root `Dockerfile` and `.dockerignore` package the Nuxt frontend and Nitro integrations for Aiven Runtime using Bun for installation/build and Node 24 for production execution on port 8080. The versioned incident-context JSON is included in the Docker build context. See [Aiven setup](docs/aiven-setup.md) for the GitHub and console steps. Optional hotline/database integration source has since been added; it has not been connected or tested. The container has not been built or run, and the application has not been deployed as part of this setup.

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

State is shared in memory while navigating between pages and accounts. Refreshing the page or resetting the demo clears browser mock state. Persisted Live reports survive and can be re-read through an unexpired verified browser grant. In scripted mode, selected documents are kept in the current browser tab using object URLs. In a verified Live dossier, the separate document adapter stores selected files in PostgreSQL and re-fetches them for previews. The demo accepts JPG, PNG, WebP, and PDF files, with a 10 MB per-file limit, a 30 MB total limit, and up to 12 files. In mock mode, one file in each requested category completes document collection; this does not represent real review. Scripted previews never count as persisted evidence. Live uploads and AI comparisons also never advance claim stage or approve coverage.

The claim stages remain distinct: **Collecting documents → Documents complete → Coverage proposal → Assistance arranged → Claim closed**. Later transitions are explicit simulated actions, with no real booking, coverage decision, or payment.

The scripted claim journey remains frontend-only. The separately authorized optional GPT-Live path now has server session control, PostgreSQL verification/reads, and confirmed report persistence. Real weather, production authentication, booking, payment and coverage decisions remain outside scope. The [document workflow](docs/evidence-workflow.md) adds genuine photo/PDF uploads, optional AI extraction and a targeted public CatNat lookup; their runtime is not yet verified. A separately authorized maintenance-only Nitro endpoint now implements database reset; it is disabled until configured. The nine-table PostgreSQL schema and small fictional dataset have now been applied to the existing Aiven service. The maintenance reset endpoint is configured locally. The scripted journey uses frontend mocks; optional Live intake uses the separate adapters described above, whose runtime behavior has not been checked in this task. Nuxt is configured with `ssr: false`.

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

Open `/insurer` directly or use the account overlay above the bottom-left account button. Every visit starts in **Operations overview**, with no selected household and no automatic database loading or AI inference. Operational labels focus on needs, capacities and next actions; data provenance is documented separately. The initial response panel groups needs by service; three portfolio-wide summary cards show urgent households/accommodation/blocked missions, known gross estimates versus approved/paid amounts, and providers with free capacity versus mobilized providers and scheduled/draft missions. Counts and costs are derived from the current records. The initial classification belongs to the seeded scenario; on-demand model results are stored separately with their sources and review status.

The right panel switches between **Response plan / Providers / Priority queue / All claims**. Select a marker only when a household preview is useful; opening it or a list row shows dossier details. The All claims list uses eight records per page. Search and municipality, priority, need, stage, situation-group and access filters drive markers and records. The plan and summary cover the whole current portfolio, independently of map filters. Service need counts are household counts, not unassigned-mission counts; free capacity remains separated by service and unit.

**Use AI-assisted review** explicitly restores insurer access and reads database dossiers (unlock in Settings if needed). This is separate from requesting inference. Select a database dossier, choose **Review with AI**, consent to the analysis and review the resulting proposal before approving or rejecting through the existing harness. The additive insurer dataset contains 80 stored dossiers; the actual loaded record count is always derived from the database response. Returning to Operations overview restores local fixtures without merging invented classifications into database results. The existing AI services, tools, proposal validation and persistence are unchanged by this UI update.

The fixture contains 30 one-person, 25 two-person, 10 three-person and 15 four-person households, totalling 170 people. Situation groups contain A15 / B10 / C35 / D20. Jules’s reserved `CAT-2026-001` is included once: before the hotline report the portfolio has 79 opened reports; opening the client claim updates that record to 80. Its documents and claim stage stay shared across account navigation. Physical damage remains unknown for inaccessible group B records; incomplete estimates are labelled partial rather than treated as a complete zero-valued loss.

Claim detail separates household needs, habitability, access, document progress, coverage review, statutory CatNat status, seven budget lines, fictional benefit caps, approved commitments and paid amounts. Jules starts at P1 because his family needs lasting accommodation. Assigning one household accommodation slot resolves that unmet need and moves him to P2; it does not repair the home, validate coverage or close the claim. Unknown access can be changed by an explicitly labelled simulation. Restricted properties cannot be scheduled.

The Mobilization view opens provider capacity and assignment details in the right panel, alongside the map. Six drying teams have two concurrent sites each; other providers use visits per day, household storage slots or household accommodation slots. Selecting a provider shows its missions. **Plan assistance** creates draft physical missions for the selected record’s outstanding needs; scheduling uses matching skills, free capacity and permitted access. Accommodation placement is a separate household action.

**Simulate provider outage** makes Delta Drying unavailable and disrupts its two seeded assignments. The replacement preview considers eligible free capacity, reserves slots across the proposed replacements and shows assumed start-delay and mission-cost changes. Applying the plan only moves affected missions; blocked properties remain blocked, and approved/paid totals do not change. Replacement timing and cost deltas are fixture inputs, not routing calculations, quotes or measured savings. Reset resource simulation restores provider and mission fixtures; the global demo reset also clears insurer overrides and local client state. Both are distinct from the separately configured database-reset operation.

The Situation report dialog summarizes the full synthetic portfolio, unmet accommodation needs, blocked missions, partial known estimates and zero initial approved/paid amounts. Download creates a text file in the browser without an export service.

### Map dependency and public basemap

The only added map library is **MapLibre GL JS 6.15.0**, installed using Bun. The map uses [OpenFreeMap Positron](https://openfreemap.org/quick_start/) without an API key, bundled module worker, clustered GeoJSON points and local assignment lines. Existing Lucide icons are reused; no Vue map wrapper, routing library or charting dependency was introduced. See [MapLibre documentation](https://maplibre.org/maplibre-gl-js/docs/).

The basemap requires **internet access and WebGL** to load its public style, tiles, fonts and sprites; attribution remains visible. Claim/provider data stays in browser memory. Map errors show a retry state while the dossier list remains usable. Markers are invented approximate locations, and assignment lines are straight illustrative links: there is no live flood layer, official hazard boundary, geolocation, customer-data request or route calculation. Keep the service’s attribution when changing the styling.

This operations-overview update changes frontend components, UI selection defaults and read-only projections only. The insurer shell fills the available desktop height; the map shrinks with it and long panels scroll internally. The former fixed 740px/1060px content heights are removed. The existing database/AI harness is retained behind explicit actions. No application launch, build, test, browser review, database request, AI call or deployment was performed for this update.

## Event research

[Southern France flooding on 20 October 2026](docs/research/20-october-2026-flood-evidence.md) separates sourced event facts, attributed impact counts, insurance procedures and a proposed synthetic Allstorm scenario. The research snapshot is dated 22 October 2026; it does not add live integrations or change the application.

## Demo database preparation

The initial database preparation provided a simplified contract and a small dataset. A later explicit authorization adds the optional Live claim intake/verification endpoints and controlled PostgreSQL projection described in [the Live guide](docs/openai-live-hotline.md). A subsequently requested settings page adds a narrowly scoped maintenance reset endpoint, described below. See [the data contract](docs/data-contract.md) for all fields, JSON shapes and reset behaviour. There is no login or sessions table. Each customer has at most one policy, enforced by a unique foreign key; each seeded customer has one.

The offline-generated dataset contains three fictional customers (Jules, Camille and Sam), three policies, four plain-text contractual fixtures, eleven search sections, three reserved/reported claims, four messages, three events, three providers and no assignments. It is a small development fixture, not the complete 80-household portfolio. Jules has a reserved `CAT-2026-001` record awaiting the hotline report; the Live core updates it instead of creating a duplicate. No real insurer policy has been imported. Subsequently generated synthetic policy PDFs are stored separately by the additive insurer preparation described below.

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

The header and hero primary CTA open `/hotline`; a secondary text link below the hero button opens `/insurer`. The closing section offers both routes as buttons, with the hotline first. Landing CTA arrow icons have been removed. Labels use the EN/FR dictionaries. These navigation changes were authored without launching the app, building, running tests or checking a browser.

`/overview` is a standalone desktop marketing page for **insurers**, with the existing EN/FR toggle. Its compact hero pairs regular-weight copy with the map and legend extracted for display from **A — Map-first crisis desk**. The navigation background spans the full page width, with its contents aligned to the hero. The original Disaster Network mark remains in the header, footer and shared SVG favicon.

The user requested **only the map and its legend** from reference A, retaining its colours and markers. The built-in ImageGen tool removed the floating claim card, selected pin and zoom controls from the original composition; the cleaned full canvas is saved as `public/images/marketing/insurer-map-a-clean.png`. `MarketingWorkspacePreview.vue` displays only original coordinates x=272–1160 / y=216–1024, excluding the sidebar, header, filters and right queue. The map illustration is capped at 620px wide and links to `/insurer`. The landing does not mount MapLibre or load the insurer portfolio. The earlier interactive preview composable remains unused.

The caption identifies this as an illustrative original concept. Historical counts embedded in the preserved legend are design artwork, not current portfolio totals. The current journey remains Jules's fictional flood claim near Istres, with **80 households / 170 people**, accommodation for three people, an AI hotline, evidence collection in My claim, coverage review and proposed assistance. The original A image and previous full-interface asset are preserved.

Editorial content and image metadata live in `useMarketingContent.js`. The regional Jules photograph remains a labelled AI-generated illustration based on a Saint-Chamas architectural reference. The flat service illustration depicts a caller and headset robot. Five secondary illustrations cover Los Angeles wildfire, Brittany windstorm, southwest France hailstorm, Mediterranean earthquake and French Alps landslide; they are not additional implemented workflows or historic incident reports. Existing assets and insurer operations are unchanged. See [the active art direction](docs/design/marketing-insurer-direction.md) and [initial imagery](docs/design/marketing-image-prompts.md).

Page text, captions and links use the EN/FR dictionaries; the original concept's embedded English text remains part of the image. No dependency, backend integration or deployment was added. No app launch, build, test or local browser/render check was run for this simplification.

## Realistic scenario data

The operational screens use natural claims language and plausible synthetic records. Jules's policy is **123456**, product **Allstorm Home Protect**, 2026 edition. Camille and Sam use **123457** and **123458**. The security-word question is “What is your security word?” and its public evaluation answer is **olivier**. The expected answer is checked on the server; never embed it in frontend configuration or a voice prompt. These details identify a fictional evaluation scenario, not real policyholders or insurance cover.

The 80-household / 170-person portfolio uses invented first names, varied reports, distinct contact times and severity-dependent estimates. Main screens and voice replies avoid repeated demonstration jargon. Evaluation provenance remains in Settings, source/data metadata, a discreet workspace footer and exported reports. Proposals, site-access constraints, document previews, approvals and payments retain their actual meaning. The realism pass itself added no operational capability; the separately authorized document workflow below adds collection and optional reading, while booking, payment and coverage approval remain outside scope. Historical design images retain their older copy.

`bun run db:export` regenerates the local JSON/SQL fixtures without contacting PostgreSQL. Existing installed data requires a targeted reference-data refresh: ordinary `db:seed` only inserts missing rows and does not rename an existing policy. The authorized refresh preserves Jules's report and all customer-written content; it does not reset the dataset.

### Reference-data refresh applied — 22 October 2026

The user explicitly authorized the corresponding Aiven update. A transaction refreshed three customer challenges, three policy schedules, four controlled contractual text documents, eleven source sections, three provider names and two unchanged seeded assistant messages. The security word is now **olivier** and Jules's policy is **123456**. Before/after database fingerprints confirmed that claims, claim events, assignments, customer messages and non-policy attachments were unchanged. No reset or schema migration occurred. Only unchanged source-fixture policy documents were eligible; an edited source would have aborted the transaction.

The offline JSON/SQL export was regenerated. The frontend, EN/FR dictionaries and agent instructions were reviewed in source; no build, tests, browser check, app restart or OpenAI call was run for this pass. Reload the UI and begin a new call to discard an older in-memory conversation or portfolio.

## Hackathon submission materials

The [submission handoff](docs/submission/hand-off.txt) indexes the English/French data-layer copy, timed three-minute recording scripts, one-slide PowerPoint and evidence PDF. The actual live-journey MP4 remains to be recorded. The user subsequently requested the entire slide as one ImageGen image: `docs/submission/output/pitch-slide-speed-scale.png` and the local image-only `pitch-slide-speed-scale.pptx`. It uses a large left headline, a simplified numbered four-step insurance illustration, three bottom sections and the user-supplied #0af06d green reference. Visible slide text is rasterized by request; the generation prompt and provenance are saved in `docs/submission/assets`. The final full-slide image is also available in the updated [public Google Slides presentation](https://docs.google.com/presentation/d/1MzBdjHUjWCAXodPqNkfdBS4Kdi3h7xOoAJRHxqz7JMU/edit); the Share dialog confirmed Anyone with the link / Viewer. A one-page local PDF is `docs/submission/output/disaster-network-pitch.pdf`. Submission copy was reviewed against the current voice, evidence and insurer agent source, including OpenAI classification, PostgreSQL attachment storage, Géorisques and optional official-domain web search. That source review did not validate the new paths in the deployed runtime.

The upload PDF now combines a two-page crisis-operations case with a transparent workload/cost model and the recorded capacity comparison. For the 80-household scenario, planning ranges are 40-60 manual staff hours versus 8-16 assisted hours, about EUR 1.5-4k versus EUR 0.4-1.2k in initial handling cost. Per-claim cost assumptions are about EUR 20-45 manually and EUR 5-15 with assistance. The earlier single-point estimates are superseded; no exact savings percentage is asserted. These are scenario calculations, not observed production performance; assumptions and formulas are in `docs/submission/evidence/operations-model.json`. The PDF also summarizes the current source guardrails and the insurer macro view.

The explicitly requested evidence comparison was executed with `bun run scripts/submission-benchmark.js`. It runs the shipped frontend operations composable offline: one provider outage, two affected missions, a baseline without free-capacity context versus the application's constraint-aware replan/applyPlan. Results: 0/2 versus 2/2 feasible replacements, two versus zero capacity violations, with three blocked missions and unrelated assignments preserved. Raw JSON/CSV and source/dataset hashes are in `docs/submission/evidence`. This is one deterministic synthetic scenario, not a model-accuracy or production benchmark. The slide's cost reduction remains a financial hypothesis. No application build, microphone session, paid AI call, database write or deployment was performed for submission preparation.

## Document workflow — source implemented, runtime unverified

The user approved three connected steps: a downloadable policy PDF; genuine supporting-document collection, including damage photos, with optional AI reading; and one official public event check. See [the evidence workflow](docs/evidence-workflow.md) for architecture, API, limits and exact state semantics.

My claim now includes a document conversation for JPG/PNG/WebP photos and PDF supporting documents. A separate 30-minute dossier capability is issued after Live verification and access to a reported claim. Uploads persist bytes and provenance in the existing PostgreSQL tables; AI reading starts only after explicit agreement to send the selected file to OpenAI. Files can be saved without reading. Extracted observations and invoice fields remain subject to review. Deterministic checks compare dates, nights and amounts with the stored policy and declared need, citing the clause and page of the generated PDF. No stage, coverage, booking, payment or access permission changes.

The PDF is generated from current stored clauses. An offline fixture sample is available at [Jules's policy PDF](output/pdf/allstorm-home-protect-123456.pdf); regenerate it with `bun scripts/export-policy-pdf.js`. Stored policy source documents are still text, without embeddings or semantic retrieval. The rendered PDF page references are explicitly separate from original source-page fields.

The official source is Géorisques GASPAR, filtered by municipality and event date. The UI retains consultation time/source and distinguishes matching records, no match, incomplete results and unavailability. A CatNat database result is not coverage approval; absence is not rejection and publication can precede database updates. There is no general web/news search or MCP connector.

No new infrastructure or schema migration is required. `pdf-lib` was installed with Bun and the lockfile updated. Private `NUXT_OPENAI_EVIDENCE_MODEL` defaults to the image-capable `gpt-6-luna`, using the existing server key. This pass consulted official documentation and authored/visually inspected the offline PDF. It did not start the app, build, run tests, inspect the browser, contact Aiven, execute a paid OpenAI call, validate a live CatNat query, push or deploy. User-operated runtime validation remains outstanding.

## Real insurer agent and stored policy PDFs

The user authorized an OpenAI insurer agent and actual policy PDF storage. See [the insurer agent guide](docs/insurer-agent.md) for the full harness, human validation, configuration and limitations. The source uses Responses tool calling plus structured output, with the existing private OpenAI key. It reads actual stored dossiers, clauses/PDFs, photos/invoices, messages, source checks and provider capacities/locations. It proposes classification, conditional cover, nearby eligible teams and evidenced reimbursement, then waits for an exact human approval/rejection. Source/version checks, clause/benefit matching, access, duplicate services, capacity and monetary limits are enforced outside the model. Approval records classification/proposal and `approved_in_demo` assignments; no real payment, booking or dispatch occurs.

`bun run db:insurer` was executed successfully on Aiven: **80 claims / 170 people, 10 providers and 80 actual policy PDFs**. Its additive transaction inserted missing scenario rows and PDF versions without replacing existing claims, messages, attachments, assignments or edited policy sources. Preservation checks passed and no reset or migration occurred. Downloads read stored PDF bytes linked to their source hash; no vector index is needed. The original small reset fixture remains three households: after an intentional maintenance reset, run this preparation again.

Private local insurer settings were appended to the ignored `.env` without replacing existing values, and an existing OpenAI key was detected without printing it. Restart the server yourself, then enter the local `NUXT_INSURER_ACCESS_TOKEN` value in Settings → Private insurer access key. This gate is separate from the hotline and maintenance keys. `NUXT_INSURER_AGENT_ENABLED=true` enables the agent; `NUXT_OPENAI_INSURER_MODEL` defaults to `gpt-6-luna`. Optional `NUXT_INSURER_WEB_SEARCH_ENABLED=true` enables official-domain OpenAI search using the same API key. Direct Géorisques GASPAR remains a separate token-free harness tool.

In the database workspace, map/list/resource data is fetched from PostgreSQL. Select a dossier and Review with AI, agree to analysis, then inspect the proposal, source files and missing information. A human note and explicit review checkbox are required before applying. Eligible operations update stored capacity and map links; blocked operations stay blocked. Reviewed classification/summary/reimbursement proposals appear in the situation report. Unsupported amounts remain unknown; financial commitments, paid totals and claim stages do not advance automatically. There is no fixture fallback for an AI/service failure.

This pass executed database preparation and private local configuration only, plus source/dictionary inspection. No app launch, build, test suite, browser check, paid AI call, live CatNat/web test, GitHub push or deployment was performed. Model/account access and the complete runtime interaction remain unverified.

## Generic model instructions — 22 October 2026

Hotline voice/delegation and insurer instructions contain workflow and review rules, without seeded customer names, policy references, household sizes, event dates or numerical policy limits. Identity lookup uses caller-supplied values checked against PostgreSQL; household facts and contractual conditions come from verified tools. Dated incident context comes from its versioned context tool. The insurer reads the selected database dossier and PDF; supported benefit codes are supplied in the dossier tool result. Estimates never become reimbursements by instruction. Browser-save commentary uses the actual repository result rather than a fixed claim reference.

The hotline repository remains deliberately scoped to the reserved fixture UUIDs, and its tool names retain their existing protocol identifiers. This change does not claim production authentication or arbitrary-customer support. Database data, provider fulfilment and payments retain their documented evaluation provenance. No paid AI call is implied by a successful build or deployment.

### Public hotline admission

The user explicitly removed the hotline workspace-token gate for jury access. `NUXT_HOTLINE_ACCESS_TOKEN` is no longer read and can be removed from server configuration. Keep `NUXT_HOTLINE_ENABLED=true`, the exact `NUXT_HOTLINE_ORIGIN`, private OpenAI settings and database configuration. Existing origin/header checks, isolated cookie owners, request and call limits, microphone consent, customer verification, evidence grants and browser report confirmation remain. Automatic admission is not authentication and grants no customer data. Insurer and database-reset gates remain separate. Source changes require a deployment before the hosted application changes. No build, browser check or paid API call was run for this change.

## Mobile client access — 23 October 2026

The hotline and shared client shell now adapt below 760px: compact account header, two bottom navigation entries, touch-sized call controls, a smaller particle orb, wrapped transcripts and a stacked report review. My claim and voice settings have basic narrow-screen layouts; the insurer workspace remains desktop-focused. The existing Thinking Orbs implementation and licence are unchanged.

`public/manifest.webmanifest` opens `/hotline` in standalone mode, with PNG icons rendered from the existing brand mark and Apple Home Screen metadata. On iPhone, open the deployed HTTPS `/hotline` in Safari, choose Share → Add to Home Screen, enable Open as Web App when shown, and add it. No signing certificate, App Store package, service worker or offline cache is involved. Installation does not grant microphone permission or start a call; tap the orb and allow the microphone. Network connectivity is required, and a backgrounded/locked iPhone is not a supported continuous-call environment.

The hosted origin remains the existing `https://getdisasternetwork.com`; no tunnel or runtime configuration change is needed. Local responsive review can use `bun run dev`. The greeting follows the UI language; voice and backend instructions follow the caller’s substantive spoken language, without switching on names, numbers, filler sounds or accent alone. This is prompt behaviour, not a guarantee of language-detection accuracy. Sources: [Apple Home Screen web apps](https://support.apple.com/en-ca/guide/iphone/iphea86e5236/27/ios/27), [OpenAI Live prompting](https://developers.openai.com/api/docs/guides/live-prompting).

Source and assets prepared on `codex/mobile-hotline` from the current `main`. Pre-existing tracked recording/local-documentation changes were preserved in a recoverable Git stash; the untracked recording preset was removed at the user's request. The production build passed on 23 October 2026 for the requested release. No microphone/paid AI test was run; physical-iPhone installation and audio remain user checks. Home Screen installation and audio playback still need checking on a physical iPhone after deployment.
