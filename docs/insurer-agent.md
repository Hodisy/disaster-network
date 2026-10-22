# Insurer decision-support agent

Source implementation on 22 October 2026. This is a real OpenAI Responses tool loop over stored PostgreSQL data, with human-reviewed proposals. Policies, policyholders, approximate locations, provider capacity, assignments and payment outcomes remain an invented evaluation scenario. There is no real dispatch, booking or money transfer.

## What was actually prepared

The user explicitly authorized additive database preparation and storing the policy PDFs. `bun run db:insurer` completed against the existing Aiven service with verified TLS. Its transaction reported **80 claims, 170 people, 10 providers, 80 stored policy PDFs**, with **80 PDFs inserted**, `existingWorkPreserved: true` and `resetPerformed: false`. Before/after fingerprints confirmed every pre-existing claim, message, attachment and assignment row remained unchanged. Existing customer/policy/document IDs were inserted only when absent; edited source documents were preserved. No schema migration, reset or paid AI request was performed.

The preparation imports the same portfolio generator as the UI into the existing nine tables, rather than inventing backend tool responses. It retains Jules once and the existing Camille/Sam records. New portfolio amounts are explicitly estimates; no approved or paid money was created. Existing three providers retain their database capacities and prices, even where the historical UI preview differs. Additional scenario providers are inserted without replacing existing rows. No initial assignment is fabricated from the old browser preview.

`policy_pdf_stored` events link each generated PDF document ID to its policy and the hash of the exact stored clauses/schedule. PDFs are real `application/pdf` bytes in `documents.content`, not embeddings. Sources remain separately preserved. Re-running preparation skips an unchanged source hash; a changed source produces another PDF version without overwriting a prior document. Client and insurer downloads now read the stored PDF. A missing/current-source-mismatched PDF is unavailable, not a silent generated fallback. The small legacy fixture/reset remains three customers: after an explicitly requested maintenance reset, run `bun run db:insurer` again to restore the additive portfolio and PDF versions. Do not reset to refresh a PDF.

## API choice

Classification uses [Responses](https://developers.openai.com/api/docs/guides/function-calling) and [structured output](https://developers.openai.com/api/docs/guides/structured-outputs), rather than a separate classification endpoint. The private existing OpenAI key is reused; no Dev API vendor or separate classification credential is needed. `gpt-6-luna` is the default insurer model. Model/account availability and inference behaviour have not been tested in this implementation pass.

There is no agent framework, vector database, queue, embedding index or additional infrastructure. The Nuxt monolith retains endpoint -> core -> repository/service boundaries in JavaScript. `server/core/insurer/agent.js` orchestrates the bounded tool loop; `planning.js` enforces operational/financial constraints; `server/repositories/insurerAgent.js` reads and persists actual data. Presentation components have props/events only.

## Configuration and local access

Keep existing OpenAI and Aiven settings unchanged. Private runtime settings:

```dotenv
NUXT_INSURER_AGENT_ENABLED=true
NUXT_INSURER_ACCESS_TOKEN=<separate random 32-256 character private insurer key>
NUXT_OPENAI_INSURER_MODEL=gpt-6-luna
NUXT_INSURER_WEB_SEARCH_ENABLED=true
```

Defaults in `.env.example` are disabled/empty. For this local checkout, missing settings were appended to the ignored `.env`: insurer enabled, a newly generated separate insurer gate, the model default and optional official-domain web search enabled. Existing settings and credentials were preserved. The operation confirmed an existing OpenAI key was present and printed no secret. Restart the development server yourself to load the new settings. In Settings, copy the **value of `NUXT_INSURER_ACCESS_TOKEN` from your local `.env`** into Private insurer access key, then unlock. Do not put the OpenAI API key in this form or send it through chat.

The insurer gate is distinct from the client hotline gate and maintenance key. A signed HttpOnly, SameSite Strict cookie grants the insurer role for 30 minutes under `/api/insurer`. An account-switch UI is not authorization. Exact configured origin and custom request header are checked. Access is restricted to this evaluation workspace; it is not production insurer SSO. HTTP and analysis rate limits are process-local.

## Harness tools and access

| Tool | Actual access |
| --- | --- |
| `get_dossier` | Selected stored claim, household, policy clauses, evidence/previous extraction, messages, source checks and human review notes; no verification hashes or credentials |
| `read_policy_pdf` | Current stored PDF bytes and verified ID/SHA, passed to the model as a real file input |
| `read_supporting_document` | Actual attached PDF/photo; reuses current extraction or genuinely invokes the existing Responses extractor and stores derived facts before comparing |
| `get_available_teams` | PostgreSQL provider skills, service areas, availability windows, pricing, existing assignments, free capacity and straight-line distance from the policy location |
| `list_portfolio` | Stored portfolio priorities, housing/access states and household totals |
| `check_official_catnat` | Actual Géorisques GASPAR lookup by stored municipality and event date, with distinct sourced/absent/incomplete/unavailable outcomes |
| `search_official_context` | Optional OpenAI `web_search` over official domains, using only the public municipality/date query, with visible source links |

All attached dossier files are addressable. Analysis is bounded to 24 tool calls, a 150-second overall deadline and 32 MiB of unique file bytes. Repeated file reads reuse the existing context. Each file is limited to 10 MiB. There is no arbitrary SQL, secret access, other insurer data, real booking, payment or model-side decision-approval tool. Failures and budget limits return unavailability, not fixture proposals. Files and report/source text are untrusted data, never system instructions.

Géorisques API v1 requires no extra token. [OpenAI web search](https://developers.openai.com/api/docs/guides/tools-web-search) uses the existing OpenAI key, can incur tool charges and requires account/model support. It is separately switchable; the direct Géorisques harness tool remains available when web search is disabled. The search request contains no customer identity, private address or supporting-file contents. Search is limited to Géorisques, Légifrance, the Bouches-du-Rhône prefecture and Météo-France domains; returned HTTPS citations are shown in the source trace. Public context never establishes contractual coverage or official recognition by itself.

## Proposals and human review

Choose Use AI-assisted review to load the database workspace, then select a dossier on the map/list and choose Review with AI. Explicitly agree to sending this dossier and requested files to OpenAI, then analyze. The agent must read the dossier and current policy PDF, and consult team candidates for every proposed service. It returns an A-D classification, P1-P3 priority, reasons, one coverage line per seven benefits, missing information, operations/provider choices and a reimbursement explanation. The exact validated proposal, model, language, tool trace, document IDs/hashes and data fingerprint are persisted as `insurer_ai_proposal` in `claim_events`.

The server validates source IDs, clause/benefit correspondence, factual habitability constraints, monetary ceilings, source-backed amounts, duplicate services and resource eligibility. A declared urgent accommodation need prevents an unjustified low priority. Photos describe visible observations only. The model cannot declare physical access permitted. Team candidates respect skill, service area, shared free capacity and recorded assignments. Straight-line distances from approximate coordinates are not measured routes or arrival-time guarantees. Accommodation units are household slots; a window allocation is not a confirmed multi-night hotel stay.

Coverage can be proposed conditionally while an amount stays unknown. The initial monetary corroboration supports accommodation invoices through extracted fields and deterministic comparison with the declared stay, household, policy validity, nights and limits. Duplicate content/invoice facts do not inflate the sum. Unassessed contents/building or other unsupported amounts remain unknown, not a zero-valued loss or the gross scenario estimate. The single combined contents/building deductible stays outstanding until a loss is assessed. A proposed amount is not an approved financial commitment. Broader invoice categories and assessed-loss workflows can extend this bounded core later.

The reviewer can inspect the PDF/read supporting files from the source trace, write a note and explicitly confirm review, then approve or reject the exact proposal ID. On approval the repository re-reads the complete case/policy/evidence/resource snapshot under the same advisory transaction lock as maintenance. Any relevant intervening change rejects the old proposal. Eligibility/capacity is rechecked and only eligible operations become `approved_in_demo` assignments. Blocked work remains blocked. Classification, reasons, coverage/reimbursement proposal and the review event are stored; no paid money, approved financial commitment, claim stage, customer narrative or access permission is changed. Duplicate matching approval is idempotent; conflicting second decisions are rejected. Approved simulated assignments consume capacity and update map lines after refresh. Subsequent access restrictions block their physical display/dispatch state.

The UI separates Database workspace from the explicit legacy Portfolio preview. In database mode all map/list/priority/resource data comes from PostgreSQL, without an error fallback to browser fixtures. Legacy outage/placement/access-preview buttons are disabled/hidden there. Pending eligible AI operations appear as proposal map links; reviewed classification and simulated assignments update from the database. Client My claim can reflect the reviewed coverage-proposal status on its next authorized refresh. The exported situation report includes human-reviewed AI classifications, summaries and reimbursement proposals separately from gross estimates and paid totals.

## Validation status

Executed: additive Aiven preparation and its in-transaction preservation checks; local private configuration preparation; source/import/dictionary/diff inspection. The policy sample PDF had already been visually inspected in the document task. No application launch, build, test suite, browser check, paid OpenAI analysis, live web/CatNat test, human approval exercise, GitHub push or deployment was performed in this pass. The complete runtime interaction remains for the user's validation.

## Operations overview and explicit activation — 22 October 2026

`/insurer` now opens the local Operations overview with no claim selected. It does not restore database mode or start inference on mount. The read-only overview derives urgencies, gross estimates/approved commitments/payments, free provider capacity and scheduled/draft/blocked missions from the active dataset. The response plan groups reported needs by service; capacity units remain separate and household need counts do not imply unassigned missions. Provider details are in the right panel; the map stays visible within the desktop viewport.

To exercise the existing real analysis, explicitly load the database workspace, unlock access in Settings if required, select a dossier, then request Review with AI and consent. The existing proposal, evidence trace and approve/reject process is preserved. No new portfolio-wide inference or blanket approval is added. Prepared preview classifications are not claimed as AI outputs, and this interface update does not modify the separately prepared 80-dossier database. Source changes only: no build, tests, browser checks, database calls or paid inference was run for this UI pass.

### Product copy

Operational screens use insurer vocabulary: needs, costs, capacity, proposed operations and review. Fixture provenance remains in this technical documentation and Settings rather than repeated throughout the operational journey. This presentation change does not convert local records into live customer data or add a dispatch, booking or payment integration. The model and tool trace remain available for each generated proposal.

## Generic model instructions — 22 October 2026

Hotline voice/delegation and insurer instructions contain workflow and review rules, without seeded customer names, policy references, household sizes, event dates or numerical policy limits. Identity lookup uses caller-supplied values checked against PostgreSQL; household facts and contractual conditions come from verified tools. Dated incident context comes from its versioned context tool. The insurer reads the selected database dossier and PDF; supported benefit codes are supplied in the dossier tool result. Estimates never become reimbursements by instruction. Browser-save commentary uses the actual repository result rather than a fixed claim reference.

The hotline repository remains deliberately scoped to the reserved fixture UUIDs, and its tool names retain their existing protocol identifiers. This change does not claim production authentication or arbitrary-customer support. Database data, provider fulfilment and payments retain their documented evaluation provenance. No paid AI call is implied by a successful build or deployment.
