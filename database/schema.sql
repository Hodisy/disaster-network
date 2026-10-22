-- Initial contract, version 1. Applied only by scripts/database.js.
-- All data in this schema is disposable, fictional demonstration data.
CREATE SCHEMA disaster_demo;
COMMENT ON SCHEMA disaster_demo IS 'disaster-network-demo:v1';

CREATE TABLE disaster_demo.customers (
  id uuid PRIMARY KEY,
  first_name text NOT NULL,
  last_name text,
  email text,
  phone text,
  household_size smallint NOT NULL DEFAULT 1 CHECK (household_size > 0),
  verification_question text NOT NULL,
  verification_answer_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE disaster_demo.documents (
  id uuid PRIMARY KEY,
  claim_id uuid,
  kind text NOT NULL CHECK (kind IN ('general_terms', 'specific_terms', 'endorsement', 'photos', 'address', 'inventory', 'invoice', 'other')),
  filename text NOT NULL,
  mime_type text NOT NULL,
  size_bytes integer NOT NULL CHECK (size_bytes BETWEEN 1 AND 10485760),
  content bytea NOT NULL,
  sha256 text NOT NULL CHECK (sha256 ~ '^[a-f0-9]{64}$'),
  language text NOT NULL DEFAULT 'en' CHECK (language IN ('en', 'fr')),
  source_url text,
  provenance text NOT NULL DEFAULT 'synthetic' CHECK (provenance IN ('synthetic', 'public_document', 'customer_provided')),
  extraction_status text NOT NULL DEFAULT 'pending' CHECK (extraction_status IN ('pending', 'complete', 'failed')),
  review_status text NOT NULL DEFAULT 'received' CHECK (review_status IN ('received', 'needs_review', 'accepted_as_evidence', 'unusable')),
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (octet_length(content) = size_bytes)
);

CREATE TABLE disaster_demo.policies (
  id uuid PRIMARY KEY,
  customer_id uuid NOT NULL UNIQUE REFERENCES disaster_demo.customers(id),
  contract_number text NOT NULL UNIQUE,
  version integer NOT NULL DEFAULT 1 CHECK (version > 0),
  product_name text NOT NULL,
  valid_from date NOT NULL,
  valid_to date NOT NULL CHECK (valid_to >= valid_from),
  address text NOT NULL,
  city text NOT NULL,
  postal_code text NOT NULL,
  insee_code text NOT NULL,
  latitude double precision CHECK (latitude BETWEEN -90 AND 90),
  longitude double precision CHECK (longitude BETWEEN -180 AND 180),
  property_type text NOT NULL CHECK (property_type IN ('house', 'apartment')),
  occupancy text NOT NULL CHECK (occupancy IN ('owner_occupier', 'tenant', 'landlord')),
  general_terms_document_id uuid NOT NULL REFERENCES disaster_demo.documents(id),
  specific_terms_document_id uuid NOT NULL REFERENCES disaster_demo.documents(id),
  coverage_data jsonb NOT NULL CHECK (jsonb_typeof(coverage_data) = 'object'),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE disaster_demo.claims (
  id uuid PRIMARY KEY,
  reference text NOT NULL UNIQUE,
  policy_id uuid NOT NULL REFERENCES disaster_demo.policies(id),
  incident_type text NOT NULL,
  incident_at timestamptz NOT NULL,
  reported_at timestamptz,
  intake_status text NOT NULL DEFAULT 'awaiting_report' CHECK (intake_status IN ('awaiting_report', 'reported')),
  description text,
  summary text,
  stage text NOT NULL DEFAULT 'collecting_documents' CHECK (stage IN ('collecting_documents', 'documents_complete', 'coverage_proposal', 'assistance_arranged', 'claim_closed')),
  habitability text NOT NULL DEFAULT 'unknown' CHECK (habitability IN ('habitable', 'uninhabitable', 'unknown')),
  access_status text NOT NULL DEFAULT 'unknown' CHECK (access_status IN ('unknown', 'restricted', 'permitted_in_demo')),
  priority text CHECK (priority IN ('P1', 'P2', 'P3')),
  priority_reasons jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(priority_reasons) = 'array'),
  needs jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(needs) = 'array'),
  required_documents jsonb NOT NULL DEFAULT '["photos", "address", "inventory"]' CHECK (jsonb_typeof(required_documents) = 'array'),
  assessment jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(assessment) = 'object'),
  cost_estimate jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(cost_estimate) = 'object'),
  updated_at timestamptz NOT NULL DEFAULT now(),
  closed_at timestamptz,
  CHECK ((intake_status = 'reported') = (reported_at IS NOT NULL)),
  CHECK (reported_at IS NULL OR reported_at >= incident_at),
  CHECK ((stage = 'claim_closed') = (closed_at IS NOT NULL)),
  CHECK (closed_at IS NULL OR (reported_at IS NOT NULL AND closed_at >= reported_at))
);

ALTER TABLE disaster_demo.documents ADD CONSTRAINT documents_claim_fk
  FOREIGN KEY (claim_id) REFERENCES disaster_demo.claims(id);

CREATE TABLE disaster_demo.document_sections (
  id uuid PRIMARY KEY,
  document_id uuid NOT NULL REFERENCES disaster_demo.documents(id),
  position integer NOT NULL CHECK (position >= 0),
  title text NOT NULL,
  clause_reference text,
  page_start integer CHECK (page_start > 0),
  page_end integer CHECK (page_end >= page_start),
  text text NOT NULL,
  language text NOT NULL DEFAULT 'en' CHECK (language IN ('en', 'fr')),
  search_index tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector(CASE WHEN language = 'fr' THEN 'french'::regconfig ELSE 'english'::regconfig END, title), 'A') ||
    setweight(to_tsvector(CASE WHEN language = 'fr' THEN 'french'::regconfig ELSE 'english'::regconfig END, text), 'B')
  ) STORED,
  UNIQUE (document_id, position),
  CHECK ((page_start IS NULL) = (page_end IS NULL))
);
CREATE INDEX document_sections_search_idx ON disaster_demo.document_sections USING gin(search_index);

CREATE TABLE disaster_demo.messages (
  id uuid PRIMARY KEY,
  customer_id uuid REFERENCES disaster_demo.customers(id),
  claim_id uuid REFERENCES disaster_demo.claims(id),
  conversation_id uuid NOT NULL,
  channel text NOT NULL CHECK (channel IN ('voice', 'chat')),
  role text NOT NULL CHECK (role IN ('client', 'assistant', 'handler')),
  content text NOT NULL,
  language text NOT NULL DEFAULT 'en' CHECK (language IN ('en', 'fr')),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE disaster_demo.claim_events (
  id uuid PRIMARY KEY,
  claim_id uuid NOT NULL REFERENCES disaster_demo.claims(id),
  type text NOT NULL,
  actor text NOT NULL CHECK (actor IN ('client', 'agent', 'handler', 'demo_setup')),
  data jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(data) = 'object'),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE disaster_demo.providers (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  service_types jsonb NOT NULL CHECK (jsonb_typeof(service_types) = 'array'),
  latitude double precision CHECK (latitude BETWEEN -90 AND 90),
  longitude double precision CHECK (longitude BETWEEN -180 AND 180),
  service_area jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(service_area) = 'array'),
  capacity integer NOT NULL CHECK (capacity >= 0),
  availability jsonb NOT NULL CHECK (jsonb_typeof(availability) = 'object'),
  pricing jsonb NOT NULL CHECK (jsonb_typeof(pricing) = 'object')
);

CREATE TABLE disaster_demo.assignments (
  id uuid PRIMARY KEY,
  claim_id uuid NOT NULL REFERENCES disaster_demo.claims(id),
  provider_id uuid NOT NULL REFERENCES disaster_demo.providers(id),
  service_type text NOT NULL,
  starts_at timestamptz NOT NULL,
  ends_at timestamptz NOT NULL CHECK (ends_at > starts_at),
  status text NOT NULL DEFAULT 'proposed' CHECK (status IN ('proposed', 'approved_in_demo', 'in_progress_in_demo', 'completed_in_demo', 'cancelled')),
  estimated_cost_cents integer NOT NULL CHECK (estimated_cost_cents >= 0),
  currency text NOT NULL DEFAULT 'EUR' CHECK (currency = 'EUR'),
  instructions text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX claims_policy_idx ON disaster_demo.claims(policy_id);
CREATE INDEX documents_claim_idx ON disaster_demo.documents(claim_id);
CREATE INDEX messages_claim_date_idx ON disaster_demo.messages(claim_id, created_at);
CREATE INDEX messages_conversation_idx ON disaster_demo.messages(conversation_id, created_at);
CREATE INDEX claim_events_date_idx ON disaster_demo.claim_events(claim_id, created_at);
CREATE INDEX assignments_claim_idx ON disaster_demo.assignments(claim_id);
CREATE INDEX assignments_provider_time_idx ON disaster_demo.assignments(provider_id, starts_at, ends_at);
