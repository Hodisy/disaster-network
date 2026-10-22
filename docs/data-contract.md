# Demo data contract, version 1

Disaster Network / fictional Allstorm Insurance. This document defines the original database preparation and subsequent scoped Live/document adapters. A subsequently authorized maintenance-only reset endpoint is available when explicitly configured. No login, users table or sessions table is introduced. All data is synthetic. Field names and canonical values are English; the frontend translates display labels.

## Scope and existing Aiven service

Reuse the PostgreSQL service identified in [Aiven setup](aiven-setup.md). That document describes an existing service, not a verified application connection. No credentials were present during initial preparation. The subsequent user-authorized data load configured an ignored local `.env` from the existing Aiven service and its CA certificate. Adding a connection variable in Aiven does not make the current browser-only gateway persistent.

The database scripts use a dedicated `disaster_demo` schema. They do not use `public`, provision a database, deploy Nuxt, expose API routes or start a server. Installing the nine tables and fixtures requires a deliberate command with your existing connection details.

## Relationships

```text
customers 1 -- 0..1 policies 1 -- N claims
policies -- documents (general and specific terms)
claims 1 -- N documents / messages / claim_events / assignments
documents 1 -- N document_sections
providers 1 -- N assignments
```

`policies.customer_id` is unique: at most one policy per customer, and exactly one for each seeded customer. No separate properties, policy versions, identity checks, conversations, incidents or account tables. `conversation_id` groups messages without authenticating anyone. The current single-policy demo does not model renewals; do not replace policy terms after claims refer to them. A future renewal requirement needs a separate versioning migration.

## Definitive fields for version 1

The executable definition and constraints are in [schema.sql](../database/schema.sql). IDs are UUIDs; timestamps are `timestamptz`, dates are `date`, and JSON fields use `jsonb`.

| Table | Fields |
| --- | --- |
| `customers` | `id`, `first_name`, nullable `last_name`, nullable `email`, nullable `phone`, `household_size`, `verification_question`, `verification_answer_hash`, `created_at` |
| `policies` | `id`, unique `customer_id`, unique `contract_number`, `version`, `product_name`, `valid_from`, `valid_to`, `address`, `city`, `postal_code`, `insee_code`, nullable `latitude` / `longitude`, `property_type`, `occupancy`, `general_terms_document_id`, `specific_terms_document_id`, `coverage_data`, `created_at` |
| `documents` | `id`, nullable `claim_id`, `kind`, `filename`, `mime_type`, `size_bytes`, `content` (`bytea`), `sha256`, `language`, nullable `source_url`, `provenance`, `extraction_status`, `review_status`, `created_at` |
| `document_sections` | `id`, `document_id`, `position`, `title`, nullable `clause_reference`, nullable `page_start` / `page_end`, `text`, `language`, generated `search_index` (`tsvector`) |
| `claims` | `id`, unique `reference`, `policy_id`, `incident_type`, `incident_at`, nullable `reported_at`, `intake_status`, nullable `description` / `summary`, `stage`, `habitability`, `access_status`, nullable `priority`, `priority_reasons`, `needs`, `required_documents`, `assessment`, `cost_estimate`, `updated_at`, nullable `closed_at` |
| `messages` | `id`, nullable `customer_id` / `claim_id`, `conversation_id`, `channel`, `role`, `content`, `language`, `created_at` |
| `claim_events` | `id`, `claim_id`, `type`, `actor`, `data`, `created_at` |
| `providers` | `id`, `name`, `service_types`, nullable `latitude` / `longitude`, `service_area`, `capacity`, `availability`, `pricing` |
| `assignments` | `id`, `claim_id`, `provider_id`, `service_type`, `starts_at`, `ends_at`, `status`, `estimated_cost_cents`, `currency`, `instructions`, `created_at` |

### JSON contracts

- `coverage_data`: `{ fictional, currency, catnat_status, benefits: [{ code, limit_cents }], accommodation_max_nights, accommodation_max_nightly_cents, contents_building_deductible_cents }`. It mirrors demonstration wording, not a real insurer's policy or approved eligibility.
- `needs`: `[{ code, status, quantity }]`; codes match benefit/service codes. Initially `to_confirm`; an identified need is not an assignment.
- `required_documents`: array of category codes, initially `photos`, `address`, `inventory`. Counts and collection status are derived from attached documents; collection is not evidence review.
- `assessment`: `{ status, source_section_ids, missing_information, fictional }`. Initially `not_reviewed`. Future factual assertions must retain their evidence references. Previous assessments belong in `claim_events`.
- `cost_estimate`: `{ currency, gross_cents, approved_cents, paid_cents, lines, fictional }`. Unknown estimates/approval are `null`, not zero. Initial paid amount is zero. Each future line should carry its benefit code, estimated amount and assumptions.
- `availability`: `{ status, starts_at, ends_at, fictional }`.
- `pricing`: `{ currency, unit, estimated_unit_cents }`.
- `claim_events.data`: event-specific facts and evidence references. Never put verification answers, API keys or full document binaries in logs.

Money is integer cents in EUR. Status codes and IDs are stable values, not translated labels. Human priority, document collection, coverage analysis, access and mission completion remain independent. Access must be `permitted_in_demo` before a future core approves physical work. Capacity, conflicting bookings, expected costs and matching service types require future core logic; the initial SQL schema does not implement an allocation engine.

## Jules and the small fixture

The public IDs are defined once in [app/config/demo.js](../app/config/demo.js). The frontend composable exposes `customerId` and includes these identifiers on its mock claim for a future API adapter. The initial preparation did not include claim-handling APIs. Subsequent Live and document additions now use the existing schema; see the additions below.

| Item | Initial data |
| --- | --- |
| Customers | Jules, Camille and Sam, all fictional; no surname is invented for Jules |
| Policies | Three, one per customer; Jules's contract is `123456` |
| Documents | Four synthetic UTF-8 text documents: shared general terms and three specific terms |
| Search sections | Eleven, with English lexical search indexes and exact source document references |
| Claims | Three reserved records; Jules's `CAT-2026-001` is `awaiting_report`, and the other two have reported scenarios |
| Messages / events | Four sample messages on the other two claims; three setup/report events |
| Providers / assignments | Three fictional providers; no existing assignments or real bookings |

The small fixture is not the complete 80-household portfolio defined in the [insurer brief](scenarios/insurer-workspace-brief.md). Jules is included once. On hotline completion the Live core updates his reserved claim instead of inserting a duplicate; do not count an `awaiting_report` row as an opened claim.

Event date: 20 October 2026. Report/demo date: 22 October 2026, Europe/Paris. Town-scale coordinates are illustrative, not exact households. Jules retains the address already used by the frontend as a scenario setting; it is not evidence of an actual damaged residence or a real resident. Contacts use the reserved `.invalid` domain and no real phone numbers. Product limits follow the fictional Allstorm brief. The sample general terms include exclusions and the fictional EUR 300 contents/building deductible. No source has been fetched, no real insurance PDF imported and no PDF pages fabricated. Text fixtures have nullable page numbers; real PDFs can later use `application/pdf`, extracted text, and actual page references.

The scripted frontend simulates document collection and stage transitions; the separately authorized Live document adapter stores actual selected files. This work does not implement real identity verification. The fixture verification word is `olivier`; its stored format is `scrypt:<salt>:<digest>`, with lowercase/trimmed answer semantics for the Live verifier. This is public demo material, not a real identity factor. Before a successful check, the Live harness only receives verification results, never the private customer projection or answer hash. Keep the temporary verification result in the active call context; no login or database sessions are needed for this demo.

## Commands and reset

Prerequisites: Bun, the `psql` PostgreSQL client, the existing Aiven database credentials, and its CA certificate. No new package is required.

1. Copy `.env.example` to the ignored `.env` and set `DATABASE_URL` and `DATABASE_CA_CERT_PATH`. Encode special characters in the connection URL. Keep credentials and certificates out of Git.
2. Run `bun run db:setup` to create the demo schema and load fixtures in one transaction. If the matching schema already exists, setup only inserts missing fixture IDs. It does not overwrite existing claim work or migrate schema changes.
3. `bun run db:seed` inserts missing fixture IDs only. Conflicts do not overwrite existing rows; this is not a reset.
4. `bun run db:reset --confirm-demo-reset` deletes records, binaries and messages **only inside the marked `disaster_demo` schema**, then reloads the small fixture in the same transaction. It does not drop tables, reset the Aiven service or modify `public`. This discards any demo progress and custom records in that schema.
5. `bun run db:export` regenerates the committed JSON and SQL fixtures entirely offline, without opening a database connection.

The scripts verify the schema ownership/version marker before seed/reset, serialize their own operations with an advisory transaction lock, use bounded lock/statement timeouts, and roll back uncommitted changes on an error. External connections use `verify-full` TLS with the configured CA. Stop active Live calls before reset. Live repository operations and reset share the same advisory transaction lock; Live drafts also reject a changed record version. The optional settings reset endpoint requires a private server feature flag, maintenance key and explicit confirmation; never expose it without those checks. The existing UI **Reset demo** button still clears browser memory only.

The subsequent authorized data load executed `db:setup` successfully. A read-only check confirmed verified TLS, nine tables and the expected fixture counts. No reset, app launch, build, test or browser validation was executed; automatic approval review rejected launching the app for the Settings check. The CLI requires `psql`; the deployed Node image does not include this maintenance tool and does not run these commands automatically.

## Optional settings reset

The later request for a discreet settings page adds `/settings` and `POST /api/demo/reset`, using a core and repository inside the Nuxt monolith. See [README configuration](../README.md#discreet-settings-page). The route uses `pg`, so it does not depend on `psql` in the deployed container. It resets only an existing, correctly marked schema; setup remains a separate CLI operation. CLI and endpoint share `database/operations.js`. Aiven connectivity and the initial load have since been verified, but no actual reset has been performed.

## Live intake addition — 22 October 2026

The optional Live harness uses the existing schema without migrations. Before verification, lookup returns only the challenge/result. Verification remains in active server memory and is revoked on voice termination. A separate 30-minute same-browser read grant after confirmed save or verified reading of an existing report supports My claim and Jules's insurer projection. It is not production identity or insurer-role authentication.

On explicit browser confirmation, the repository transaction updates Jules's existing claim: report, summary, habitability, restricted/unknown access, identified needs, required categories and priority. `claim_events.data` for `hotline_report_saved` contains `operation_id`, `fictional`, stable customer/policy IDs, `confirmation: explicit_browser_review` and `draft`. The draft retains safety, declared household, current accommodation, requested start/nights/constraints, needs and document codes. It does not store verification answers. A voice-channel client message stores the confirmed report narrative; complete call transcripts/audio are not persisted. Repeated operation IDs are idempotent under the maintenance advisory lock; stale drafts are rejected.

This intake never changes contractual terms, assessment, coverage approval, site permission, money, assignments or claim stage. Scripted previews remain browser-only; the subsequent verified Live evidence adapter stores attachments separately. The 80-household insurer portfolio still uses frontend synthetic fixtures for the other 79 households and financial/resource assumptions; the small PostgreSQL seed has only three customers. See the [Live guide](openai-live-hotline.md) for all tools, environment names, authorization/retention limits and steps the user can run. No service connection or live behavior has been verified.

## Realistic reference presentation — 22 October 2026

Policy labels are **Allstorm Home Protect**, with six-digit references: Jules **123456**, Camille **123457**, Sam **123458**. Existing UUIDs and claim references are unchanged. The verification question is “What is your security word?” with the public evaluation word **olivier**, still stored only as the scrypt hash and checked on the server. Contacts remain non-deliverable reserved-domain examples. Policy documents remain UTF-8 text, with HP-prefixed clause references and an explicit evaluation-copy notice; they are not PDFs.

A separately authorized Aiven reference refresh updated these labels, challenge hashes, controlled clauses, provider names and unchanged seeded assistant replies. It did not reset the database or change existing claims, user-written messages, attachments, events or assignments. Updated JSON and SQL fixtures match the new reference presentation. Full audio, app and browser validation remains unperformed in this pass.

## Supporting-document addition — 22 October 2026

See [the document workflow](evidence-workflow.md). The existing `documents` table stores actual selected JPG/PNG/WebP/PDF binaries and `customer_provided` provenance for the verified reserved dossier. The existing `claim_events` table records received documents, bounded AI extraction results/comparisons, analysis leases/failures and dated official-source checks. No schema migration or initial data update was run. Source policy documents remain text; runtime PDFs render current clauses with explicitly computed export-page references. Original source-page fields remain unchanged.

AI extraction uses a private Responses model only after explicit browser agreement. `extraction_status = complete` means reading finished, not evidence accepted. Uploaded documents remain `needs_review` after reading. Consistent numerical checks do not approve eligibility, authenticity, payment or CatNat recognition. Claim stages, assessments, money and assignments are preserved. The official lookup uses Géorisques GASPAR, without general web search or embeddings. Runtime behaviour remains unverified in this implementation pass.

## Insurer agent addition — 22 October 2026

The user-authorized `bun run db:insurer` extended the existing schema additively to 80 claims / 170 people, 10 providers and 80 stored synthetic policy PDFs. Existing claim/message/document/assignment rows were fingerprint-checked unchanged. No reset or migration occurred. `policy_pdf_stored` events bind policy ID, PDF ID/SHA and current-source hash. `insurer_ai_proposal` events retain the bounded model proposal, source/data hash and tool trace. `insurer_human_decision` retains the exact proposal ID, approve/reject choice, note and evaluation-assignment IDs. Human approval can update classification/assessment and create `approved_in_demo` assignments under the maintenance lock; financial commitments, paid money, stage and site permission remain independent. See [the insurer agent guide](insurer-agent.md) for scopes, configuration and unperformed runtime validation.
