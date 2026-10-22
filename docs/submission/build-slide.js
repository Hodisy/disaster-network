// Local editable pitch slide. Diagram nodes represent application data flow.
export default async function buildPitch({ presentation }) {
  const slide = presentation.slides.getItem(0)
  const ink = '#17201B', muted = '#6E7771', green = '#447D5C', pale = '#E6F1E8', line = '#CED9D1'
  slide.background.fill = '#FEFEFC'
  function text(name, value, left, top, width, height, fontSize, bold = false, color = ink, alignment = 'left') {
    const shape = slide.shapes.add({ geometry: 'textbox', name, position: { left, top, width, height }, fill: 'none', line: { fill: 'none', width: 0 } })
    shape.text = value
    shape.text.style = { typeface: 'Arial', fontSize, bold, color, alignment, insets: 0 }
    return shape
  }
  function node(name, label, left, top, size, accented = false) {
    const shape = slide.shapes.add({ geometry: 'ellipse', name, position: { left, top, width: size, height: size }, fill: accented ? pale : '#FEFEFC', line: { fill: accented ? '#A9C8B3' : line, width: 1.5 } })
    shape.text = label
    shape.text.style = { typeface: 'Arial', fontSize: accented ? 27 : 21, bold: true, color: accented ? green : ink, alignment: 'center', verticalAlignment: 'middle', insets: 0 }
    return shape
  }
  function connect(from, to, fromSide = 'right', toSide = 'left') {
    slide.shapes.connect(from, to, { kind: 'straight', fromSide, toSide, line: { fill: line, width: 1.5 }, tail: { type: 'arrow', width: 'sm', length: 'sm' } })
  }
  text('brand', 'Disaster Network', 56, 40, 600, 44, 31, true)
  text('category', 'AI-native crisis operations for insurers', 56, 112, 600, 30, 20, false, green)
  text('headline', 'Keep claims moving\nwhen crisis hits.', 56, 173, 590, 170, 57, true)
  text('summary', 'Automate the first call.\nCoordinate the response at scale.', 56, 369, 570, 85, 26, false, muted)
  text('value', 'Faster assistance. Optimized costs.', 56, 474, 580, 35, 22)

  const database = text('database', 'Aiven PostgreSQL\nIdentity · Policy · Claim', 814, 140, 310, 60, 19, false, muted, 'center')
  const hotline = node('hotline-node', '01', 722, 238, 64)
  const verified = node('verified-node', 'AI', 930, 230, 80, true)
  const operations = node('operations-node', '03', 1140, 238, 64)
  connect(hotline, verified)
  connect(verified, operations)
  connect(database, verified, 'bottom', 'top')
  text('hotline-label', 'Hotline', 679, 320, 150, 30, 22, true, ink, 'center')
  text('hotline-subtitle', 'Voice intake', 674, 355, 160, 26, 17, false, muted, 'center')
  text('verified-label', 'Verified claim', 875, 320, 190, 30, 22, true, ink, 'center')
  text('verified-subtitle', 'Checks before action', 865, 355, 210, 26, 17, false, muted, 'center')
  text('operations-label', 'Operations', 1084, 320, 176, 30, 22, true, ink, 'center')
  text('operations-subtitle', 'Capacity + access', 1084, 355, 176, 26, 17, false, muted, 'center')
  text('proof-number', '2/2', 722, 415, 120, 62, 49, true, green)
  text('proof-explanation', 'valid reassignments\nvs 0/2 without capacity data', 858, 422, 385, 58, 19)
  text('proof-scope', 'Rules benchmark · One synthetic outage · 3 blocked missions preserved', 722, 487, 515, 38, 14, false, muted)

  text('for-label', 'FOR', 56, 569, 330, 26, 17, true, green)
  text('for-copy', 'Claims and assistance teams\nat insurers and TPAs', 56, 607, 345, 64, 21)
  text('why-label', 'WHY NOW', 451, 569, 330, 26, 17, true, green)
  text('why-copy', 'Claim surges outpace teams.\nVoice agents can now act.', 451, 607, 345, 64, 21)
  text('users-label', 'FIRST USERS', 863, 569, 360, 26, 17, true, green)
  text('users-copy', 'Direct outreach to claims leaders.\nOne regional flood-response pilot.', 863, 607, 365, 64, 21)
  slide.speakerNotes.text = `Target buyer: claims/assistance leaders at a French home insurer or claims administrator (TPA). Proposed first-users strategy, not existing customer traction: direct outreach and a narrowly scoped paid pilot with historical flood cases.
Working data flow in source: browser voice via WebRTC; Nitro executes bounded identity/policy/claim tools; PostgreSQL on Aiven holds three fictional customers, one policy each, policy text sections and Jules's reserved claim. The 80-household insurer portfolio is a separate frontend synthetic fixture, not 80 DB-backed customers. File previews remain local; no live weather, real bookings, payments or coverage approvals.
Evidence: bun run scripts/submission-benchmark.js. The real useInsurerWorkspace replan/applyPlan code handles one Delta Drying outage. Baseline selects the first available matching provider without remaining capacity. Both receive the same two affected missions. Baseline: zero feasible replacements, two capacity violations; application: two feasible replacements, no capacity violation. Three blocked missions and all unrelated missions are preserved. Dataset 80 households/170 people, 10 providers. This is a deterministic data/rules comparison, not an AI reasoning or production benchmark.
Financial hypothesis: manual 30 min × €50/hour = €25/intake. With automation, 2 min oversight + 30% escalated × 20 min = 8 min mean human time; €6.67 + assumed €2 AI/infrastructure = €8.67. Excludes software subscription, integration, maintenance, indemnities and actual field intervention costs. No measured financial result.
French pitch: Disaster Network relie hotline automatisée et opérations pour absorber les pics de sinistres, accélérer l’assistance et optimiser les coûts. Les règles vérifient les contraintes avant les actions; les dossiers ambigus restent à valider par un humain.`
  return { slideCount: presentation.slides.items.length }
}
