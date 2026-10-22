const fail = code => { throw Object.assign(new Error(code), { code }) }
export async function insurerResponse(config, body, signal) {
  const key = config.openaiApiKey || process.env.OPENAI_API_KEY
  if (!key) fail('OPENAI_NOT_CONFIGURED')
  let response
  try {
    response = await fetch('https://api.openai.com/v1/responses', { method: 'POST', signal,
      headers: { Authorization: `Bearer ${key}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ model: config.openaiInsurerModel || 'gpt-6-luna', store: false, reasoning: { effort: 'low' }, ...body }) })
  } catch { fail('INSURER_AI_UNAVAILABLE') }
  if (!response.ok) fail(response.status === 429 ? 'OPENAI_RATE_LIMITED' : 'INSURER_AI_UNAVAILABLE')
  let result
  try { result = await response.json() } catch { fail('INSURER_AI_UNAVAILABLE') }
  if (result.status !== 'completed') fail('INSURER_AI_UNAVAILABLE')
  return result
}
export const responseText = result => (result.output || []).filter(i => i.type === 'message').flatMap(i => i.content || []).filter(i => i.type === 'output_text').map(i => i.text).join('')
export async function searchOfficialContext(config, municipality, eventDate, signal) {
  const result = await insurerResponse(config, {
    instructions: 'Research only the supplied public municipality and date. Cite official sources, distinguish historic context from matching event evidence. Do not infer insurance eligibility. Never invent a source.',
    input: `Official public flood context and published notices for ${municipality}, France, event date ${eventDate}. If none is established say so.`,
    tools: [{ type: 'web_search', search_context_size: 'low', filters: { allowed_domains: ['georisques.gouv.fr', 'legifrance.gouv.fr', 'bouches-du-rhone.gouv.fr', 'meteofrance.fr'] } }],
    max_output_tokens: 1800
  }, signal)
  const citations = (result.output || []).flatMap(i => i.content || []).flatMap(i => i.annotations || []).filter(a => a.type === 'url_citation')
    .filter(a => { try { const url = new URL(a.url); return url.protocol === 'https:' && /(^|\.)(georisques\.gouv\.fr|legifrance\.gouv\.fr|bouches-du-rhone\.gouv\.fr|meteofrance\.fr)$/.test(url.hostname) } catch { return false } })
    .map(a => ({ url: a.url, title: a.title }))
  return { text: responseText(result), citations, consultedAt: new Date().toISOString(), coverageConfirmed: false }
}
