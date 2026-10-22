import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { Client } from 'pg'
import { DEMO_SCHEMA } from '../../database/fixtures.js'
import { buildDemoSeed, RESET_DEMO_SQL, DEMO_SCHEMA_MARKER, DEMO_LOCK_NAME } from '../../database/operations.js'

function failure(code) {
  const error = new Error(code)
  error.code = code
  return error
}

export async function connectionOptions(config) {
  let url
  try {
    url = new URL(config.databaseUrl || process.env.DATABASE_URL || '')
    if (!['postgres:', 'postgresql:'].includes(url.protocol) || !url.hostname || !url.username || !url.pathname.slice(1)) throw new Error()
    if (url.username === 'USER' || url.hostname === 'host') throw new Error()
  } catch { throw failure('DATABASE_NOT_CONFIGURED') }

  const local = ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname)
  let ca = config.databaseCaCert || process.env.DATABASE_CA_CERT || ''
  if (!local && !ca) {
    const certificatePath = config.databaseCaCertPath || process.env.DATABASE_CA_CERT_PATH
    if (!certificatePath) throw failure('DATABASE_NOT_CONFIGURED')
    try { ca = await readFile(resolve(certificatePath), 'utf8') }
    catch { throw failure('DATABASE_NOT_CONFIGURED') }
  }
  try {
    // Explicit fields prevent URL sslmode parameters from overriding verified TLS.
    return {
      host: url.hostname.replace(/^\[|\]$/g, ''), port: Number(url.port || 5432),
      database: decodeURIComponent(url.pathname.slice(1)),
      user: decodeURIComponent(url.username), password: decodeURIComponent(url.password),
      ssl: local ? false : { ca, rejectUnauthorized: true },
      connectionTimeoutMillis: 10000, query_timeout: 35000,
      statement_timeout: 30000, lock_timeout: 5000,
      application_name: 'disaster-network-demo-reset'
    }
  } catch { throw failure('DATABASE_NOT_CONFIGURED') }
}

export async function resetDemoDataset(config) {
  const client = new Client(await connectionOptions(config))
  let transactionOpen = false
  try {
    await client.connect()
    await client.query('BEGIN')
    transactionOpen = true
    await client.query('SELECT pg_advisory_xact_lock(hashtext($1))', [DEMO_LOCK_NAME])
    const marker = await client.query("SELECT obj_description(oid, 'pg_namespace') AS marker FROM pg_namespace WHERE nspname = $1", [DEMO_SCHEMA])
    if (!marker.rows.length) throw failure('DEMO_SCHEMA_MISSING')
    if (marker.rows[0].marker !== DEMO_SCHEMA_MARKER) throw failure('DEMO_SCHEMA_MISMATCH')

    await client.query(RESET_DEMO_SQL)
    await client.query(buildDemoSeed())
    const counts = await client.query(`SELECT
      (SELECT count(*)::integer FROM disaster_demo.customers) AS customers,
      (SELECT count(*)::integer FROM disaster_demo.policies) AS policies,
      (SELECT count(*)::integer FROM disaster_demo.claims) AS claims,
      (SELECT count(*)::integer FROM disaster_demo.documents) AS documents`)
    await client.query('COMMIT')
    transactionOpen = false
    return { counts: counts.rows[0] }
  } catch (error) {
    if (transactionOpen) await client.query('ROLLBACK').catch(() => {})
    // Never expose SQL statements, personal data, connection strings or credentials.
    const known = ['DEMO_SCHEMA_MISSING', 'DEMO_SCHEMA_MISMATCH']
    throw failure(known.includes(error.code) ? error.code : 'DATABASE_RESET_FAILED')
  } finally {
    await client.end().catch(() => {})
  }
}
