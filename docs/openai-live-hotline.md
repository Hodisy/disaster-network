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

Copy `.env.example` to the ignored `.env` **only if `.env` does not already exist**. Otherwise add the new empty/example entries to your existing file without replacing database credentials. Enter the permanent OpenAI key directly in that local file or a private server secret; do not paste it into chat, the hotline form or public runtime configuration. The hotline form takes a separate demo access key.

Required to enable Live:

| Variable | Value / purpose |
| --- | --- |
| `OPENAI_API_KEY` | Your private project API key with GPT-Live and delegated-model access. No default; left empty in the example. Alternatively set private `NUXT_OPENAI_API_KEY`. When both exist, `NUXT_OPENAI_API_KEY` wins. |
| `NUXT_HOTLINE_ENABLED` | Set `true` to opt in. Default `false`. |
| `NUXT_HOTLINE_ACCESS_TOKEN` | A separate random 32–256 character secret for the demonstration gate. No default. Enter this value in the hotline UI; never use the OpenAI, DB or maintenance key here. |
| `DATABASE_URL` | Existing PostgreSQL connection URL. URL-encode special characters in the credentials. No real default. Alternatively private `NUXT_DATABASE_URL`, which takes precedence. |
| `DATABASE_CA_CERT_PATH` | For external PostgreSQL, a readable path to its CA certificate. Alternatively `NUXT_DATABASE_CA_CERT_PATH`, or full PEM via `DATABASE_CA_CERT` / `NUXT_DATABASE_CA_CERT`. At least one CA option is required for an external DB. |

Optional settings with defaults:

| Variable | Default / behavior |
| --- | --- |
| `NUXT_HOTLINE_ORIGIN` | `http://127.0.0.1:3000`. Must exactly equal the browser origin, including scheme and port, without a trailing slash. Change it if using `localhost`, another port, or HTTPS. Do not derive it from untrusted request headers. |
| `NUXT_OPENAI_LIVE_MODEL` | `gpt-live-1`. Requires actual account access. |
| `NUXT_OPENAI_BACKEND_MODEL` | `gpt-6-luna`, used for Responses delegation and function selection. Requires actual account access. |
| `NUXT_OPENAI_LIVE_VOICE` | `marin`. Set another supported Live voice if desired. Voice is fixed at startup; EN/FR changes update the conversation instructions during the call. |
| `NUXT_DEMO_RESET_ENABLED` | Existing maintenance feature, default `false`. Not needed for Live. |
| `NUXT_DEMO_RESET_TOKEN` | Existing maintenance secret; not needed for Live. Never enter it as the hotline access key. |

The `NUXT_*` names map to private `runtimeConfig` fields in `nuxt.config.js`. `OPENAI_API_KEY`, `DATABASE_URL` and the non-NUXT TLS variables have explicit server-only fallbacks. No setting is under `runtimeConfig.public`. Loopback PostgreSQL (`localhost`, `127.0.0.1`, `[::1]`) is the only case that disables TLS. External connections require CA validation with `rejectUnauthorized: true`; connection-string SSL parameters cannot override it. Do not log the resulting connection options.

Nuxt development loads `.env`. A directly executed built Node server needs its environment injected separately; a production `.env` is not assumed to load automatically. No deployment is authorized here.

## PostgreSQL prerequisites

The existing `disaster_demo` schema must carry marker `disaster-network-demo:v1`. Use the existing schema and seed preparation workflow if they have not yet been installed. From this folder, the user may run `bun run db:setup` after configuring the existing database and installing `psql`. That command writes the schema/fixtures; it was **not run here**. No reset is needed to try Live, and no reset should run during an active call.

Required linked records:

- Customer `10000000-0000-4000-8000-000000000001`: Jules, household size 3, verification question and `scrypt:<salt>:<digest>` hash.
- Policy `20000000-0000-4000-8000-000000000001`, related to Jules, contract `ALL-DEMO-001`, fictional Allstorm coverage, its general/specific terms documents and extracted `document_sections`.
- Reserved claim `30000000-0000-4000-8000-000000000001`, reference `CAT-2026-001`, related to that policy. It can start at `awaiting_report`.
- Existing `messages` and `claim_events` tables to retain a confirmed report and audit event.

The small seed has **three customers**, not the full insurer portfolio. The other 79 frontend households remain synthetic frontend records. A Live save updates Jules's reserved claim; it does not insert an 81st household. All operations take the same bounded advisory transaction lock as maintenance reset. Writes recheck the record version and preserve claim stage, assessment, eligibility, financial approval/payment and site permission. Intake can restrict access; it cannot authorize a physical visit.

## Implemented tools and permissions

| Tool | Actual behavior |
| --- | --- |
| `find_demo_customer` | Checks caller-supplied first name and contract against the reserved linked DB customer/policy, and returns only found/challenge information. No private projection. |
| `verify_demo_customer` | Reads the stored hash, lowercases/trims the supplied answer, derives scrypt and compares equal-length digests with `timingSafeEqual`. At most three attempts in the active voice session. Neither the stored hash nor expected answer is returned. |
| `get_customer` | Verified-only DB projection of first name, household size and stable ID. |
| `get_policy` | Verified-only DB policy/coverage JSON and actual contractual sections with document ID, section ID, clause reference, filename, provenance and nullable page references. No PDF/RAG claim. |
| `get_claim` | Verified-only read of the exact linked reserved claim; an already reported claim can renew the separate read-only frontend grant. |
| `get_incident_context` | Reads the versioned local `docs/scenarios/hotline-incident-context.json` through its repository: fixed documented context for 20 October / report 22 October 2026, Europe/Paris; 80 households / 170 people; CatNat not verified. No weather call or live event search. |
| `prepare_report` | Validates the report, safety, household, habitability, access, accommodation dates/nights/constraints, assistance codes and document categories. Creates an in-memory draft bound to the current DB version. No persistence or booking. |
| `save_confirmed_report` | Can save only the exact draft previously confirmed by a browser click. Model-generated/spoken confirmation is insufficient. The browser confirmation endpoint calls the same core directly, so saving does not depend on another model turn. |

All sensitive tool calls check active server-side verification. IDs are fixed and relationship-scoped; no frontend `verified` boolean or caller-chosen claim/customer ID authorizes access. The public verification word is not embedded as an expected answer in prompts, tool responses or logs. Do not include verification exchanges in report summaries. This is a public fictional demonstration challenge, not real identity authentication.

Preparation and saving are separate. Confirmation updates only the reserved claim's report, habitability, restricted/unknown access, identified needs, requested categories and operational priority. Safety, declared household, current accommodation, requested dates/nights and constraints are retained in `claim_events.data.draft`. It stores the confirmed narrative as a voice-channel client message. Full audio and the full call transcript are not written to PostgreSQL. The customer's policy household is not silently rewritten if the caller declares a different size; a mismatch needs review.

A deterministic draft/operation ID in `claim_events` and the transaction lock make repeated confirmation of the same draft idempotent. Optimistic version checks reject stale drafts from another tab/handler; preparing a fresh draft is required. Repeating or rephrasing a report never creates another claim. No schema migration was added.

Tool traces retain only call ID, registered tool name, outcome/code and elapsed milliseconds, capped at 80 entries in server memory. No arguments, verification answers, API keys, connection URLs, raw SQL errors or audio payloads are logged. Confirmed business reports are intentionally persisted as claim data, separate from operational tool traces.

## Lifecycle and controlled screen continuity

The microphone starts only from the orb/start action, after the explicit demo gate. It is never enabled on mount, mode selection, locale change or an error. The browser waits for ICE gathering, exchanges SDP through Nitro, then waits for `session.started`. Input/output transcript fragments preserve their exact spacing and approximate `start_ms` / `end_ms` in a bounded local fragment collection. Display text is grouped by speaker, not asserted to be complete turns or exact playback timing. Generated speech/transcripts and confirmed caller content remain verbatim; EN/FR changes the interface and future speech, without rewriting prior utterances.

Clicking the orb during a call temporarily mutes playback and sends an application-authored interruption instruction via the server. This is not cancellation of an already authorized database write. **End call** mutes input, waits for current tool work, sends `session.close`, and retains media/event connections for `session.closed`. Finalization has a timeout and explicitly reports uncertainty when the final event is missing. Microphone tracks and connections are then released. Leaving the hotline or unloading releases local devices immediately and requests server close with keepalive; a final browser acknowledgment is not guaranteed on unload. The server retains a 15-minute maximum call timer and hangup fallback. No automatic reconnect or paid creation retry occurs after an ambiguous failure.

Verification expires with the voice call (maximum 15 minutes). Separately, an explicitly confirmed save, or a verified read of an existing report, grants a **30-minute, read-only** capability to the same browser's HTTP-only cookie. `claim` reads PostgreSQL again through the core on client/insurer navigation. It is not a login or an insurer role. Switching the demo account is still a presentation action within this one authorized browser. New browsers, expired capabilities and server restarts require a new verified call. Reports remain in PostgreSQL even when browser/demo state is reset. No grant is stored in local storage or accepted as a public ID.

Only Jules's persisted intake fields overlay the existing insurer record. Its synthetic financial estimates (including the gross EUR 17,500 scenario) and resource actions remain clearly labeled frontend simulations. Other households are not read from this three-customer database fixture. Local file previews cannot advance the Live claim's stored stage; Live text messaging, uploads, booking and coverage decisions are not implemented. A switch to the scripted mode is explicit, never error recovery. If a saved Live claim is already open, it is retained and does not become a mock write target.

The demo gate uses a private 32–256 character token, exact configured origin checks, a strict HTTP-only cookie, a custom request header, 64 KiB maximum JSON body, 240 API requests / minute / direct source IP, 8 authorization attempts / 10 minutes / source IP, 3 paid session creations / 10 minutes / source IP, one active call per cookie owner, a 15-minute call cap and bounded in-memory maps. IP limits use the direct connection, not arbitrary forwarded headers. All state/limits are process-local: use one persistent Node/Nitro process for this demo. This is not production customer/insurer authentication or distributed rate limiting. Do not expose it as a public paid-session service without production access controls.

## User-operated verification steps

These are instructions for the user, not checks already executed:

1. Ensure Bun and a compatible Node runtime (project Node 24.15) are available. Run `bun install`. Configure the private environment settings above. If the marked schema/fixtures are absent, deliberately run the existing `bun run db:setup` with `psql` and the existing Aiven TLS details. Do not reset existing work.
2. Start the app yourself using `bun run dev`. Open `http://127.0.0.1:3000/hotline` (or the exact configured origin). Select **Live AI voice**. The scripted demo remains a separate choice.
3. Enter `NUXT_HOTLINE_ACCESS_TOKEN` in **Demo access key**, click the orb, and grant browser microphone permission. If autoplay is blocked, select **Play assistant audio**. A missing OpenAI key, DB/schema, inaccessible GPT-Live/delegated model, or sideband failure must produce a visible error without a scripted fallback.
4. Say: “I am Jules, contract ALL-DEMO-001. After the flood on 20 October, I need accommodation for my partner, our child and me. We are safe at my sister's, but cannot stay there.” Choose EN or FR; future call speech follows the app language.
5. The stored public demo challenge is **“What is the demonstration verification word?”** Its fixture answer is **`demo`**, normalized by trim/lowercase. First give a wrong answer: the result must be false and policy/claim tools must remain unavailable. Then give `demo` within the three-attempt limit. After three failures, start a new call after the applicable rate limit. This is not real identity verification.
6. After success, ask the assistant to read the actual accommodation clause and give its source reference. Expected fixture terms: up to 7 nights, EUR 150 per household/night, EUR 1,050 total, subject to facts/review. These are fictional ceilings, not eligibility or a booking. Ask about the other benefits and deductible if useful; responses must follow the DB sections.
7. Give the requested accommodation start, number of nights and constraints, confirm three people, the unusable home, damage and unknown/restricted site access. Say photos are already available; request photos, proof of address and a provisional inventory in My claim, with invoices later. The assistant should not tell you to re-enter an unsafe property or delay all urgent assistance until the inventory is complete.
8. Ask to prepare the report. Review the on-screen draft, including safety, household, accommodation dates, needs and requested categories. Ask for corrections if needed. Click **Confirm and save report** only when accurate. Before the click, no report write should occur. Repeating the click/request for the same draft must not duplicate the claim or event.
9. Open **My claim**. Check reference `CAT-2026-001`, the saved narrative/needs and requested document categories. A visible label distinguishes PostgreSQL intake from local file previews. No attachment is uploaded and no simulated progression or fake Live chat response is enabled for this record.
10. Switch to the insurer account and select Jules. The same reference must overlay his existing place in the 80-household / 170-person synthetic portfolio; there must not be 81 households. The other 79 records, money estimates and provider actions remain frontend simulations. Financial estimates are not approved/paid amounts.
11. End the call while still on the hotline to observe graceful finalization, or note that navigation itself ends the call and releases the microphone. Try reconnecting after a wrong answer, changing language, losing permission/network, a stale draft in a second tab, and an expired grant as desired. Confirm microphone indicators disappear and errors never fall back to mock replies.

All live account access, audio behavior, OpenAI protocol acceptance, PostgreSQL queries and visual layout remain unverified until those user-operated checks are performed.
