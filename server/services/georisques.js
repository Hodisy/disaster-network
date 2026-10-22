export const CATNAT_SOURCE_URL = 'https://www.georisques.gouv.fr/donnees/bases-de-donnees/procedures-administratives-relatives-aux-risques'
const API = 'https://www.georisques.gouv.fr/api/v1/gaspar/catnat'
const fail = () => { throw Object.assign(new Error('OFFICIAL_SOURCE_UNAVAILABLE'), { code: 'OFFICIAL_SOURCE_UNAVAILABLE' }) }
const date = value => typeof value === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(value) && Number.isFinite(Date.parse(value)) && new Date(value).toISOString().slice(0, 10) === value

// Public API v1 remains token-free according to the official API documentation.
// Only the municipality code leaves the server, never the address, identity or file.
export async function lookupCatNat(inseeCode, incidentDate) {
  if (!/^\d{5}$/.test(inseeCode) || !date(incidentDate)) fail()
  const checkedAt = new Date().toISOString()
  const records = []
  const signal = AbortSignal.timeout(15000)
  let complete = false
  const queries = []
  for (let page = 1; page <= 5; page++) {
    const url = `${API}?${new URLSearchParams({ code_insee: inseeCode, page: String(page), page_size: '100' })}`
    queries.push(url)
    let body
    try {
      const response = await fetch(url, { headers: { Accept: 'application/json' }, redirect: 'error', signal })
      if (!response.ok || Number(response.headers.get('content-length') || 0) > 1024 * 1024) fail()
      const raw = await response.text()
      if (raw.length > 1024 * 1024) fail()
      body = JSON.parse(raw)
    } catch { fail() }
    if (!Array.isArray(body.data) || body.data.length > 100 || !Number.isInteger(body.total_pages) || body.total_pages < 0 || body.page !== page || !Number.isInteger(body.results) || body.results < 0 || body.data.length > body.results || (body.response_code !== undefined && body.response_code !== 200) || (body.results > 0 && !body.data.length)) fail()
    for (const row of body.data) {
      if (row.code_insee !== inseeCode || !date(row.date_debut_evt) || !date(row.date_fin_evt) || row.date_fin_evt < row.date_debut_evt || typeof row.libelle_risque_jo !== 'string' || typeof row.code_national_catnat !== 'string' || !date(row.date_publication_jo)) fail()
      // Match the event period and flood risk, not an unrelated historic order.
      if (row.date_debut_evt <= incidentDate && row.date_fin_evt >= incidentDate && /inondation/i.test(row.libelle_risque_jo) && row.date_publication_jo <= checkedAt.slice(0, 10)) {
        records.push({ reference: row.code_national_catnat, risk: row.libelle_risque_jo.slice(0, 250), start: row.date_debut_evt,
          end: row.date_fin_evt, published: row.date_publication_jo })
      }
    }
    if (body.total_pages <= page) { complete = true; break }
  }
  return { provider: 'Géorisques · GASPAR', sourceUrl: CATNAT_SOURCE_URL, queries, inseeCode, incidentDate, checkedAt,
    status: records.length ? 'matching_record_found' : complete ? 'no_matching_record' : 'incomplete', complete,
    records, coverageConfirmed: false, recognitionDecision: 'requires_official_order_review',
    limitation: 'The database may lag official publication. Absence is not a rejection. A matching record is not an insurance coverage decision.' }
}
