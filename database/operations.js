import { createDemoDataset, DEMO_SCHEMA, TABLE_ORDER } from './fixtures.js'

export const DEMO_SCHEMA_MARKER = 'disaster-network-demo:v1'
export const DEMO_LOCK_NAME = 'disaster-network-demo-setup'

const quote = value => `'${String(value).replaceAll("'", "''")}'`
const valueSql = value => value === null ? 'NULL' : typeof value === 'number' ? String(value) : typeof value === 'boolean' ? String(value) : typeof value === 'object' ? `${quote(JSON.stringify(value))}::jsonb` : quote(value)

export function buildDemoSeed(dataset = createDemoDataset()) {
  return TABLE_ORDER.flatMap(table => dataset[table].map(row => {
    const columns = Object.keys(row).map(key => key === 'content_hex' ? 'content' : key)
    const values = Object.entries(row).map(([key, value]) => key === 'content_hex' ? `decode(${quote(value)}, 'hex')` : valueSql(value))
    return `INSERT INTO ${DEMO_SCHEMA}.${table} (${columns.join(', ')}) VALUES (${values.join(', ')}) ON CONFLICT (id) DO NOTHING;`
  })).join('\n')
}

// Fixed schema/table names only. Never accept a reset target from the browser.
export const RESET_DEMO_SQL = `DELETE FROM ${DEMO_SCHEMA}.assignments;
DELETE FROM ${DEMO_SCHEMA}.providers;
DELETE FROM ${DEMO_SCHEMA}.messages;
DELETE FROM ${DEMO_SCHEMA}.claim_events;
DELETE FROM ${DEMO_SCHEMA}.document_sections;
UPDATE ${DEMO_SCHEMA}.documents SET claim_id = NULL WHERE claim_id IS NOT NULL;
DELETE FROM ${DEMO_SCHEMA}.claims;
DELETE FROM ${DEMO_SCHEMA}.policies;
DELETE FROM ${DEMO_SCHEMA}.documents;
DELETE FROM ${DEMO_SCHEMA}.customers;`
