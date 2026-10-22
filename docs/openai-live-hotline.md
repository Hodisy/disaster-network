# GPT-Live hotline implementation

Source implementation prepared on 22 October 2026. No app, build, test, browser, microphone, OpenAI request or database connection was run to validate it. Only the `ws` dependency was installed using Bun, with lifecycle scripts disabled. No migration, seed, reset or deployment was applied.

## Protocol and architecture

This integration implements **GPT-Live**, not Realtime. The official [WebRTC guide](https://developers.openai.com/api/docs/guides/voice-webrtc?api=live) specifies JSON `POST /v1/live/sessions` with `session` and `transport: { type: 'webrtc', sdp }`. Its response contains `session.id` and `transport.sdp`. Audio travels over browser WebRTC media tracks; the `oai-events` data channel carries JSON transcript/lifecycle events. Creation starts the session: the browser never sends `session.start`.

The server attaches an authenticated sideband at `wss://api.openai.com/v1/live/sessions/{session_id}/attach`. It alone executes function calls from nested `response.event` / `response.output_item.done` events. Functions are collected by response/delegation ID, then results are submitted as `response.item.create` with `function_call_output`, followed by `response.create`. See [delegation and tools](https://developers.openai.com/api/docs/guides/live-delegation) and [server controls](https://developers.openai.com/api/docs/guides/voice-server-controls?api=live). No hosted search, weather, remote MCP, booking or payment tool is registered.

The transport is isolated in `server/services/openaiLive.js`, using standard server `fetch` for REST and `ws` for the private sideband. No OpenAI SDK version is required. There is **no Realtime fallback**. HTTP 401/403/404 gives an explicit key/model/access error; delegated-model validation errors stop the call. Access to both configured models must be enabled in the same OpenAI project; it has not been checked for this account.

Flow:

```text
hotline.vue / client.vue / insurer.vue
  -> useLiveHotline / useHotlineClaim
  -> POST /api/hotline/{action}
  -> server/core/hotline/sessions.js + harness.js
  -> repositories/hotlineClaims.js OR services/openaiLive.js
```

`[action].post.js` handles HTTP transport, a bounded JSON body, origin/header checks, cookie handling and safe error codes. Cores own admission, verification, tool orchestration, browser confirmation, close and projection access. The repository uses parameterized PostgreSQL statements and the existing schema; no backend mock exists. Components remain presentation-only with props/events and scoped CSS.

## Exact environment settings

Copy `.env.example` to the ignored `.env` **only if `.env` does not already exist**. Otherwise add the new empty/example entries to your existing file without replacing database credentials. Enter the permanent OpenAI key directly in that local file or a private server secret; do not paste it into chat, the hotline form or public runtime configuration. The enabled hotline is public: it has no workspace-key requirement. Mode controls and the audio disclosure remain in Settings.

Required to enable Live:

| Variable | Value / purpose |
| --- | --- |
| `OPENAI_API_KEY` | Your private project API key with GPT-Live and delegated-model access. No default; left empty in the example. Alternatively set private `NUXT_OPENAI_API_KEY`. When both exist, `NUXT_OPENAI_API_KEY` wins. |
| `NUXT_HOTLINE_ENABLED` | Set `true` to opt in. Default `false`. |
| `DATABASE_URL` | Existing PostgreSQL connection URL. URL-encode special characters in the credentials. No real default. Alternatively private `NUXT_DATABASE_URL`, which takes precedence. |
| `DATABASE_CA_CERT_PATH` | For external PostgreSQL, a readable path to its CA certificate. Alternatively `NUXT_DATABASE_CA_CERT_PATH`, or full PEM via `DATABASE_CA_CERT` / `NUXT_DATABASE_CA_CERT`. At least one CA option is required for an external DB. |

Optional settings with defaults:

| Variable | Default / behavior |
| --- | --- |
| `NUXT_HOTLINE_ORIGIN` | `http://127.0.0.1:3000`. Public domains must exactly match the browser origin, including scheme and port, without a trailing slash. For a configured loopback origin only, `localhost`, `127.0.0.1` and `[::1]` are accepted with the same scheme and port. Other ports and HTTP/HTTPS changes still require configuration. Do not derive it from untrusted request headers. |
| `NUXT_OPENAI_LIVE_MODEL` | `gpt-live-1`. Requires actual account access. |
| `NUXT_OPENAI_BACKEND_MODEL` | `gpt-6-luna`, used for Responses delegation and function selection. Requires actual account access. |
| `NUXT_OPENAI_LIVE_VOICE` | `marin`. Set another supported Live voice if desired. Voice is fixed at startup; EN/FR changes update the conversation instructions during the call. |
| `NUXT_DEMO_RESET_ENABLED` | Existing maintenance feature, default `false`. Not needed for Live. |
| `NUXT_DEMO_RESET_TOKEN` | Existing maintenance secret; not needed for Live. It protects database reset independently of public voice admission. |

The `NUXT_*` names map to private `runtimeConfig` fields in `nuxt.config.js`. `OPENAI_API_KEY`, `DATABASE_URL` and the non-NUXT TLS variables have explicit server-only fallbacks. No setting is under `runtimeConfig.public`. Loopback PostgreSQL (`localhost`, `127.0.0.1`, `[::1]`) is the only case that disables TLS. External connections require CA validation with `rejectUnauthorized: true`; connection-string SSL parameters cannot override it. Do not log the resulting connection options.

Nuxt development loads `.env`. A directly executed built Node server needs its environment injected separately; a production `.env` is not assumed to load automatically. No deployment is authorized here.

## PostgreSQL prerequisites

The existing `disaster_demo` schema must carry marker `disaster-network-demo:v1`. Use the existing schema and seed preparation workflow if they have not yet been installed. From this folder, the user may run `bun run db:setup` after configuring the existing database and installing `psql`. That command writes the schema/fixtures; it was **not run here**. No reset is needed to try Live, and no reset should run during an active call.

Required linked records:

- Customer `10000000-0000-4000-8000-000000000001`: Jules, household size 3, verification question and `scrypt:<salt>:<digest>` hash.
- Policy `20000000-0000-4000-8000-000000000001`, related to Jules, contract `123456`, fictional Allstorm coverage, its general/specific terms documents and extracted `document_sections`.
- Reserved claim `30000000-0000-4000-8000-000000000001`, reference `CAT-2026-001`, related to that policy. It can start at `awaiting_report`.
- Existing `messages` and `claim_events` tables to retain a confirmed report and audit event.

The small seed has **three customers**, not the full insurer portfolio. The other 79 frontend households remain synthetic frontend records. A Live save updates Jules's reserved claim; it does not insert an 81st household. All operations take the same bounded advisory transaction lock as maintenance reset. Writes recheck the record version and preserve claim stage, assessment, eligibility, financial approval/payment and site permission. Intake can restrict access; it cannot authorize a physical visit.

## Implemented tools and permissions

| Tool | Actual behavior |
| --- | --- |
| `find_demo_customer` | Checks caller-supplied first name and contract against the reserved linked DB customer/policy, and returns only found/challenge information. No private projection. |
| `verify_demo_customer` | Reads the stored hash, lowercases/trims the supplied answer, derives scrypt and compares equal-length digests with `timingSafeEqual`. At most three attempts in the active voice session. Neither the stored hash nor expected answer is returned. |
| `get_customer` | Verified-only DB projection of first name, household size and stable ID. |
| `get_policy` | Verified-only DB policy/coverage JSON and actual contractual sections with document ID, section ID, clause reference, filename, provenance and nullable page references. Adds deterministic `pdf_page` references in the generated policy PDF; no vector retrieval. |
| `get_claim` | Verified-only read of the exact linked reserved claim; an already reported claim can renew separate 30-minute claim-read and explicit evidence-action capabilities. |
| `get_claim_evidence` | Reads stored dossier file metadata and completed analysis, with stale comparisons flagged. Does not send a file for AI reading. |
| `check_official_catnat` | Consults Géorisques GASPAR for municipality/event-date flood records and stores the dated result; no general web search or coverage decision. |
| `get_incident_context` | Reads the versioned local `docs/scenarios/hotline-incident-context.json` through its repository: fixed documented context for 20 October / report 22 October 2026, Europe/Paris; 80 households / 170 people; CatNat not verified. No weather call or live event search. |
| `prepare_report` | Validates the report, safety, household, habitability, access, accommodation dates/nights/constraints, assistance codes and document categories. Creates an in-memory draft bound to the current DB version. No persistence or booking. |
| `save_confirmed_report` | Can save only the exact draft previously confirmed by a browser click. Model-generated/spoken confirmation is insufficient. The browser confirmation endpoint calls the same core directly, so saving does not depend on another model turn. |

All sensitive tool calls check active server-side verification. IDs are fixed and relationship-scoped; no frontend `verified` boolean or caller-chosen claim/customer ID authorizes access. The public verification word is not embedded as an expected answer in prompts, tool responses or logs. Do not include verification exchanges in report summaries. This is a public fictional demonstration challenge, not real identity authentication.

Preparation and saving are separate. Confirmation updates only the reserved claim's report, habitability, restricted/unknown access, identified needs, requested categories and operational priority. Safety, declared household, current accommodation, requested dates/nights and constraints are retained in `claim_events.data.draft`. It stores the confirmed narrative as a voice-channel client message. Full audio and the full call transcript are not written to PostgreSQL. The customer's policy household is not silently rewritten if the caller declares a different size; a mismatch needs review.

A deterministic draft/operation ID in `claim_events` and the transaction lock make repeated confirmation of the same draft idempotent. Optimistic version checks reject stale drafts from another tab/handler; preparing a fresh draft is required. Repeating or rephrasing a report never creates another claim. No schema migration was added.

Tool traces retain only call ID, registered tool name, outcome/code and elapsed milliseconds, capped at 80 entries in server memory. No arguments, verification answers, API keys, connection URLs, raw SQL errors or audio payloads are logged. Confirmed business reports are intentionally persisted as claim data, separate from operational tool traces.

## Lifecycle and controlled screen continuity

The microphone starts only from the orb/start action. `POST /api/hotline/access` automatically creates or reuses isolated browser admission and returns metadata only. No hotline key is required. The one-hour HttpOnly cookie remains tied to process-local state; expiry/restart issues fresh admission without restoring verified policy or dossier access. Mode selection, disclosure and local/database resets are confined to Settings; the hotline retains a compact hang-up control near the orb. It is never enabled on mount, mode selection, locale change or an error. The browser waits for ICE gathering, exchanges SDP through Nitro, then waits for `session.started`. Input/output transcript fragments preserve their exact spacing and approximate `start_ms` / `end_ms` in a bounded local fragment collection. Display entries group each speaker independently across approximate 1,400 ms timeline gaps and are then ordered by their start time; these are not asserted to be complete turns or exact playback timing. Exact deltas and timestamps are retained up to 10,000 fragments, with an explicit notice when that bound is reached. The full retained text is visible without the previous per-speaker 220 px scroll boxes. Generated speech/transcripts and confirmed caller content remain verbatim; EN/FR changes the interface and future speech, without rewriting prior utterances.

Clicking the orb during a call temporarily mutes playback and sends an application-authored interruption instruction via the server. This is not cancellation of an already authorized database write. The **End call** icon near the orb mutes input, waits for current tool work, sends `session.close`, and retains media/event connections for `session.closed`. Finalization has a timeout and explicitly reports uncertainty when the final event is missing. Microphone tracks and connections are then released. Leaving the hotline or unloading releases local devices immediately and requests server close with keepalive; a final browser acknowledgment is not guaranteed on unload. The server retains a 15-minute maximum call timer and hangup fallback. No automatic reconnect or paid creation retry occurs after an ambiguous failure.

Verification expires with the voice call (maximum 15 minutes). Separately, an explicitly confirmed save, or a verified read of an existing report, grants a **30-minute claim-read** capability and a separate **30-minute evidence-action** capability to the same browser's HTTP-only cookie. `claim` reads PostgreSQL again through the core on client/insurer navigation. It is not a login or an insurer role. Switching the demo account is still a presentation action within this one authorized browser. New browsers, expired capabilities and server restarts require a new verified call. Reports remain in PostgreSQL even when browser/demo state is reset. No grant is stored in local storage or accepted as a public ID. Evidence actions require explicit browser file/reading/source actions; the workspace gate alone is insufficient.

Only Jules's persisted intake fields overlay the existing insurer record. Its synthetic financial estimates (including the gross EUR 17,500 scenario) and resource actions remain clearly labeled frontend simulations. Other households are not read from this three-customer database fixture. Neither local previews nor persisted uploads/AI comparisons advance the Live claim's stored stage. The separate document conversation now supports uploads and optional reading; free-form Live text messaging, booking and coverage decisions remain outside scope. A switch to the scripted mode is explicit, never error recovery. If a saved Live claim is already open, it is retained and does not become a mock write target.

Public voice admission retains exact configured origin checks, a strict HTTP-only cookie, a custom request header, 64 KiB maximum JSON body, 240 API requests / minute / direct source IP, 30 new browser admissions / 10 minutes / source IP, 3 paid session creations / 10 minutes / source IP, one active call per cookie owner, a 15-minute call cap and bounded in-memory maps. IP limits use the direct connection, not arbitrary forwarded headers. All state/limits are process-local: use one persistent Node/Nitro process for this demo. This is not production customer/insurer authentication or distributed rate limiting. Do not expose it as a public paid-session service without production access controls.

## User-operated verification steps

These are instructions for the user, not checks already executed:

1. Ensure Bun and a compatible Node runtime (project Node 24.15) are available. Run `bun install`. Configure the private environment settings above. If the marked schema/fixtures are absent, deliberately run the existing `bun run db:setup` with `psql` and the existing Aiven TLS details. Do not reset existing work.
2. Start the app yourself using `bun run dev`. Open `http://127.0.0.1:3000/hotline` (or the exact configured origin). Open Settings, select **Live AI voice**, and read the AI/audio disclosure. The scripted demo remains a separate choice there.
3. No workspace key is needed. Return to the hotline, click the orb, and grant browser microphone permission. If autoplay is blocked, select **Play assistant audio**. A missing OpenAI key, DB/schema, inaccessible GPT-Live/delegated model, or sideband failure must produce a visible error without a scripted fallback.
4. Say: “I am Jules, contract 123456. After the flood on 20 October, I need accommodation for my partner, our child and me. We are safe at my sister's, but cannot stay there.” Choose EN or FR; future call speech follows the app language.
5. The stored public demo challenge is **“What is your security word?”** Its fixture answer is **`olivier`**, normalized by trim/lowercase. First give a wrong answer: the result must be false and policy/claim tools must remain unavailable. Then give `olivier` within the three-attempt limit. After three failures, start a new call after the applicable rate limit. This is not real identity verification.
6. After success, ask the assistant to read the actual accommodation clause and give its source reference. Expected fixture terms: up to 7 nights, EUR 150 per household/night, EUR 1,050 total, subject to facts/review. These are fictional ceilings, not eligibility or a booking. Ask about the other benefits and deductible if useful; responses must follow the DB sections.
7. Give the requested accommodation start, number of nights and constraints, confirm three people, the unusable home, damage and unknown/restricted site access. Say photos are already available; request photos, proof of address and a provisional inventory in My claim, with invoices later. The assistant should not tell you to re-enter an unsafe property or delay all urgent assistance until the inventory is complete.
8. Ask to prepare the report. Review the on-screen draft, including safety, household, accommodation dates, needs and requested categories. Ask for corrections if needed. Click **Confirm and save report** only when accurate. Before the click, no report write should occur. Repeating the click/request for the same draft must not duplicate the claim or event.
9. Open **My claim**. Check reference `CAT-2026-001`, the saved narrative/needs and requested document categories. The document conversation can save photos/PDFs, optionally read them with explicit agreement and download a policy PDF. See [the evidence workflow](evidence-workflow.md). No simulated progression or fabricated Live chat response is enabled.
10. Switch to the insurer account and select Jules. The same reference must overlay his existing place in the 80-household / 170-person synthetic portfolio; there must not be 81 households. The other 79 records, money estimates and provider actions remain frontend simulations. Financial estimates are not approved/paid amounts.
11. End the call while still on the hotline to observe graceful finalization, or note that navigation itself ends the call and releases the microphone. Try reconnecting after a wrong answer, changing language, losing permission/network, a stale draft in a second tab, and an expired grant as desired. Confirm microphone indicators disappear and errors never fall back to mock replies.

All live account access, audio behavior, OpenAI protocol acceptance, PostgreSQL queries and visual layout remain unverified until those user-operated checks are performed.

## Hotline presentation refinement

The welcome is instructed to identify Allstorm Insurance and the AI assistant, then ask how to help before checking relevant safety needs. Fictional provenance remains in Settings, documentation, data and tool results; it is not a repeated spoken greeting. Both voice and delegated backend instructions follow the caller’s spoken language after the UI-language greeting. These source changes have not been built or checked with a live call.

## Realistic reference data — 22 October 2026

The policy reference is now **123456** (also accepted with spaces or hyphens), product **Allstorm Home Protect**. Its security-word challenge is “What is your security word?”; the public evaluation answer is **olivier**. Camille and Sam use 123457 and 123458. Old contract references and the old verification word are not runtime aliases. The expected answer remains absent from frontend configuration, prompts and tool outputs.

Operational labels, clauses and spoken guidance use natural insurance language. Evaluation provenance remains in Settings, metadata, documentation and the exported situation report. State accuracy is unchanged: requests are not bookings, estimates are not approval, inaccessible sites remain blocked, and scripted previews stay browser-local. Verified Live document uploads use the separate evidence capability.

The user authorized and the agent applied a targeted Aiven reference-data transaction, preserving all claims, customer messages, claim events, assignments and non-policy attachments. The offline seed was regenerated separately. This data update did not test a voice call, restart the application, run builds/tests or deploy anything. Start a fresh call after reloading the UI.

## Subsequent document workflow

The approved [document workflow](evidence-workflow.md) adds protected nested evidence endpoints, generated policy PDF pages, actual photo/PDF storage, opt-in Responses extraction, deterministic comparisons and a targeted Géorisques GASPAR lookup. Private `NUXT_OPENAI_EVIDENCE_MODEL` defaults to `gpt-6-luna`. Existing clause documents are not replaced and no schema migration is needed. Analysis and source results never approve coverage or set CatNat status. No runtime validation or database operation was executed for this source addition.

## Stored policy PDFs and insurer agent

The subsequently authorized [insurer agent](insurer-agent.md) now reads the full stored synthetic portfolio under a separate insurer gate. Additive Aiven preparation stored 80 policy PDFs and preserved existing work. `get_policy` exposes the matching stored PDF metadata; client PDF downloads use its actual bytes. The agent has real dossier/document/resource/Géorisques tools plus optional official-domain web search, but its classifications and operations await human validation. No real booking or payment is introduced, and no paid AI/runtime validation was run for this addition.

## Generic model instructions — 22 October 2026

Hotline voice/delegation and insurer instructions contain workflow and review rules, without seeded customer names, policy references, household sizes, event dates or numerical policy limits. Identity lookup uses caller-supplied values checked against PostgreSQL; household facts and contractual conditions come from verified tools. Dated incident context comes from its versioned context tool. The insurer reads the selected database dossier and PDF; supported benefit codes are supplied in the dossier tool result. Estimates never become reimbursements by instruction. Browser-save commentary uses the actual repository result rather than a fixed claim reference.

The hotline repository remains deliberately scoped to the reserved fixture UUIDs, and its tool names retain their existing protocol identifiers. This change does not claim production authentication or arbitrary-customer support. Database data, provider fulfilment and payments retain their documented evaluation provenance. No paid AI call is implied by a successful build or deployment.

## Public jury access

Public hotline admission was explicitly requested by the user. Existing `NUXT_HOTLINE_ACCESS_TOKEN` values are ignored; enabled voice calling automatically admits each browser separately. Public admission never substitutes for the spoken customer challenge, verified dossier grants or explicit report confirmation. The insurer key and maintenance key are unchanged. These source changes were not built or tested and do not deploy themselves.
