import { spawnSync } from 'node:child_process'
import { readFileSync, writeFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createDemoDataset, DEMO_SCHEMA } from '../database/fixtures.js'
import { buildDemoSeed, RESET_DEMO_SQL, DEMO_SCHEMA_MARKER, DEMO_LOCK_NAME } from '../database/operations.js'

// Bun loads .env. psql is used as a local client; no application backend is added.
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const action = process.argv[2]
const marker = DEMO_SCHEMA_MARKER
const allowed = new Set(['export', 'setup', 'seed', 'reset'])
if (!allowed.has(action)) {
  console.error('Usage: bun scripts/database.js export|setup|seed|reset [--confirm-demo-reset]')
  process.exit(1)
}

const dataset = createDemoDataset()
const seed = buildDemoSeed(dataset)
if (action === 'export') {
  // This action is offline: no credentials are read and no connection is opened.
  writeFileSync(resolve(root, 'database/demo-data.json'), `${JSON.stringify(dataset, null, 2)}\n`)
  writeFileSync(resolve(root, 'database/seed.sql'), `-- Generated fictional dataset. Use bun run db:seed; do not run against another schema.\n${seed}\n`)
  console.log('Prepared database/demo-data.json and database/seed.sql. No database was contacted.')
  process.exit(0)
}

if (action === 'reset' && !process.argv.includes('--confirm-demo-reset')) {
  console.error('Reset deletes all records, files and messages inside disaster_demo only. Run bun run db:reset --confirm-demo-reset to restore the fictional dataset.')
  process.exit(1)
}

let connection
try {
  connection = new URL(process.env.DATABASE_URL || '')
  if (!['postgres:', 'postgresql:'].includes(connection.protocol) || !connection.hostname || !connection.pathname.slice(1)) throw new Error()
  if (['USER', 'HOST', 'PASSWORD'].includes(connection.hostname) || connection.username === 'USER') throw new Error()
} catch {
  console.error('Set a real PostgreSQL DATABASE_URL in the ignored .env file. No credentials are printed.')
  process.exit(1)
}

const local = ['localhost', '127.0.0.1', '[::1]'].includes(connection.hostname)
const environment = { ...process.env }
// Do not inherit unrelated libpq options, connection files or TLS overrides.
for (const key of Object.keys(environment)) if (key.startsWith('PG')) delete environment[key]
Object.assign(environment, {
  PGHOST: connection.hostname, PGPORT: connection.port || '5432',
  PGDATABASE: decodeURIComponent(connection.pathname.slice(1)),
  PGUSER: decodeURIComponent(connection.username), PGPASSWORD: decodeURIComponent(connection.password),
  PGCONNECT_TIMEOUT: '10', PGSSLMODE: local ? 'disable' : 'verify-full', PGAPPNAME: 'disaster-network-demo-setup'
})
if (!local) {
  const certPath = process.env.DATABASE_CA_CERT_PATH
  if (!certPath) {
    console.error('Set DATABASE_CA_CERT_PATH to the existing Aiven service CA certificate. Remote connections require verified TLS.')
    process.exit(1)
  }
  try { readFileSync(resolve(root, certPath)) } catch {
    console.error('The configured CA certificate file cannot be read. No connection was attempted.')
    process.exit(1)
  }
  environment.PGSSLROOTCERT = resolve(root, certPath)
}

const guard = `DO $guard$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = '${DEMO_SCHEMA}') THEN
    RAISE EXCEPTION 'Demo schema is missing. Run db:setup first.';
  END IF;
  IF obj_description('${DEMO_SCHEMA}'::regnamespace, 'pg_namespace') IS DISTINCT FROM '${marker}' THEN
    RAISE EXCEPTION 'Schema ownership/version marker does not match. No demo data was changed.';
  END IF;
END $guard$;`

const schema = readFileSync(resolve(root, 'database/schema.sql'), 'utf8')
const setup = `SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = '${DEMO_SCHEMA}') AS demo_schema_exists \\gset
\\if :demo_schema_exists
${guard}
\\else
${schema}
\\endif`
const input = `BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
SELECT pg_advisory_xact_lock(hashtext('${DEMO_LOCK_NAME}'));
${action === 'setup' ? setup : guard}
${action === 'reset' ? RESET_DEMO_SQL : ''}
${seed}
COMMIT;
`
const result = spawnSync('psql', ['-X', '--no-password', '--set=ON_ERROR_STOP=1', '--quiet'], {
  input, encoding: 'utf8', env: environment, maxBuffer: 1024 * 1024, timeout: 60000
})
if (result.error || result.status !== 0) {
  // Do not echo psql diagnostics: a failed SQL statement can contain document data.
  console.error(result.error?.code === 'ENOENT'
    ? 'The psql client is required. Install PostgreSQL client tools, then rerun this command.'
    : 'Database preparation did not complete. Uncommitted changes were rolled back. Check connectivity, certificate, permissions and schema version locally; raw errors are not printed to avoid exposing data.')
  process.exit(1)
}
console.log(`Demo ${action} completed in disaster_demo. Public schema and other databases were not modified.`)
