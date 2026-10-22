import { PDFDocument, StandardFonts, rgb } from 'pdf-lib'

const fail = () => { throw Object.assign(new Error('POLICY_PDF_UNAVAILABLE'), { code: 'POLICY_PDF_UNAVAILABLE' }) }
const ink = rgb(.10, .10, .12), muted = rgb(.43, .43, .48)
export async function renderPolicyPdf(policy) {
  const pdf = await PDFDocument.create()
  const regular = await pdf.embedFont(StandardFonts.Helvetica)
  const bold = await pdf.embedFont(StandardFonts.HelveticaBold)
  const clean = value => Array.from(String(value ?? '').replace(/\r\n?/g, '\n')).map(character => {
    if (character === '\n') return '\n'
    try { regular.encodeText(character); return character } catch { fail() }
  }).join('')
  const text = (page, value, x, y, size = 11, strong = false, color = ink) => page.drawText(clean(value), { x, y, size, font: strong ? bold : regular, color })
  function paragraph(page, value, y, size = 11) {
    for (const line of clean(value).split('\n')) {
      let current = ''
      for (const word of line.split(/\s+/)) {
        if (regular.widthOfTextAtSize(word, size) > 475) fail()
        const next = current ? `${current} ${word}` : word
        if (regular.widthOfTextAtSize(next, size) > 475) {
          if (y < 100) fail()
          text(page, current, 60, y, size); y -= 18; current = word
        } else current = next
      }
      if (y < 100) fail()
      text(page, current, 60, y, size); y -= 26
    }
    return y
  }
  const count = Math.max(...policy.sections.map(section => section.pdf_page))
  function page(number, label) {
    const page = pdf.addPage([595.28, 841.89])
    text(page, 'ALLSTORM', 60, 783, 18, true)
    text(page, 'INSURANCE', 60, 765, 8, false, muted)
    text(page, label, 60, 713, 11, false, muted)
    page.drawLine({ start: { x: 60, y: 690 }, end: { x: 535, y: 690 }, thickness: .7, color: rgb(.84, .84, .87) })
    text(page, `${policy.contract_number}  |  ${policy.product_name}`, 60, 54, 9, false, muted)
    text(page, `${number} / ${count}`, 503, 54, 9, false, muted)
    return page
  }
  const cover = page(1, 'POLICY DOCUMENTS / 2026 EDITION')
  text(cover, 'Your home insurance', 60, 643, 26, true)
  const schedule = policy.sections[0]
  if (schedule?.pdf_page !== 1) fail()
  let y = paragraph(cover, `Policy number: ${policy.contract_number}  /  Version ${policy.version}`, 598)
  y = paragraph(cover, schedule.text, y - 18)
  paragraph(cover, 'This PDF reproduces the schedule and clauses currently linked to this policy in the claim database. Read the schedule and general terms together. The following pages retain their clause references. A requested benefit is subject to assessment and supporting evidence.', y - 20, 10)
  text(cover, `Source: ${schedule.clause_reference} / ${schedule.id}`, 60, 82, 8, false, muted)
  for (let number = 2; number <= count; number++) {
    const clause = page(number, 'GENERAL TERMS / BENEFITS AND CONDITIONS')
    let y = 651
    for (const section of policy.sections.filter(item => item.pdf_page === number)) {
      const titleEnd = paragraph(clause, `${section.clause_reference} / ${section.title}`, y, 13)
      y = paragraph(clause, section.text, titleEnd - 6, 10) - 22
    }
  }
  pdf.setTitle(`${policy.product_name} - ${policy.contract_number}`)
  pdf.setAuthor('Allstorm Insurance')
  pdf.setSubject('Synthetic evaluation policy; generated from stored policy clauses')
  return Buffer.from(await pdf.save())
}
