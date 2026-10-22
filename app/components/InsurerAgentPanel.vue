<script setup>
import { insurerServiceLabels } from '../composables/useInsurerAgent.js'
const { t, formatMoney, formatDate } = useAppI18n()
defineProps({ reference: String, proposal: Object, busy: String, error: String, consent: Boolean, reviewed: Boolean, note: String })
defineEmits(['analyze', 'approve', 'reject', 'download', 'document', 'update:consent', 'update:reviewed', 'update:note'])
const money = amount => formatMoney(amount / 100, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
</script>
<template>
  <section class="agent-review">
    <div class="intro"><span>{{ reference }}</span><h2>{{ t('AI decision proposal') }}</h2><p>{{ t('The agent reads the stored dossier and policy PDF, checks evidence and selects teams using skills, capacity and approximate distance. A human decides what to apply.') }}</p></div>
    <label class="checkbox"><input type="checkbox" :checked="consent" :disabled="!!busy" @change="$emit('update:consent', $event.target.checked)"><span>{{ t('I agree to send this dossier and the files requested by the agent to OpenAI for this analysis.') }}</span></label>
    <UiButton class="dark" :disabled="!!busy || !consent" @click="$emit('analyze')">{{ t(busy === 'analyze' ? 'Agent analysis in progress…' : 'Analyze dossier with AI') }}</UiButton>
    <p v-if="busy === 'analyze'" role="status">{{ t('Reading the dossier, sources and resource constraints. This may take a moment.') }}</p>
    <p v-if="error" class="error" role="alert">{{ t(error) }}</p>
    <template v-if="proposal">
      <div class="proposal-summary"><div class="tags"><span>{{ proposal.data.proposal.group }}</span><span>{{ proposal.data.proposal.priority }}</span><span>{{ t(proposal.decision ? proposal.decision.decision === 'approve' ? 'Human review recorded' : 'Proposal rejected' : 'Awaiting human review') }}</span></div><p>{{ proposal.data.proposal.summary }}</p><p>{{ proposal.data.proposal.rationale }}</p><small>{{ formatDate(proposal.created_at, { hour: '2-digit', minute: '2-digit' }) }} · {{ proposal.data.model }}</small></div>
      <p v-if="proposal.stale" class="warning">{{ t('The dossier or resources changed. Run a new analysis before approving.') }}</p>
      <h3>{{ t('Coverage and reimbursement proposal') }}</h3>
      <article v-for="line in proposal.data.proposal.coverage_lines" :key="line.benefit" class="line"><div><strong>{{ t(insurerServiceLabels[line.benefit]) }}</strong><span>{{ line.amount_cents === null ? t('Unknown') : money(line.amount_cents) }}</span></div><small>{{ t(line.status === 'proposed' ? 'Coverage proposed · Human review required' : line.status === 'not_applicable' ? 'Not applicable · Proposed conclusion' : 'To review') }}</small><p>{{ line.reason }}</p><button @click="$emit('download')">{{ t('Policy clause {clause} · Page {page}', { clause: line.clause_reference, page: line.pdf_page }) }}</button></article>
      <div class="refund-total"><span>{{ t('Evidenced reimbursement proposal') }}</span><strong>{{ proposal.data.proposal.reimbursement_cents === null ? t('Unknown') : money(proposal.data.proposal.reimbursement_cents) }}</strong></div><p v-if="proposal.data.proposal.reimbursement_incomplete" class="warning">{{ t('Partial proposal · Some benefit amounts remain unknown.') }}</p><p>{{ proposal.data.proposal.reimbursement_explanation }}</p><p>{{ t(proposal.data.proposal.deductible_status) }}</p><p class="warning">{{ t('Unassessed losses remain unknown. This proposal does not approve payment.') }}</p>
      <h3>{{ t('Proposed teams and operations') }}</h3>
      <article v-for="operation in proposal.data.proposal.operations" :key="operation.service" class="line"><div><strong>{{ t(insurerServiceLabels[operation.service]) }}</strong><span>{{ t(operation.eligible ? 'Ready for human review' : 'Blocked') }}</span></div><p>{{ operation.name || t('Provider not assigned') }}<template v-if="operation.distance_km !== null && operation.distance_km !== undefined"> · {{ t('About {distance} km in a straight line', { distance: operation.distance_km }) }}</template></p><p>{{ operation.reason }}</p><p v-for="reason in operation.blocked_reasons" :key="reason" class="warning">{{ t(reason) }}</p><small v-if="operation.start">{{ formatDate(operation.start, { hour: '2-digit', minute: '2-digit' }) }} · {{ operation.estimated_cents != null ? money(operation.estimated_cents) : t('Unknown') }}</small></article>
      <p>{{ t('Approval records the reviewed assignment. Dispatch, booking confirmation and payment remain separate operational steps.') }}</p>
      <h3>{{ t('Information still needed') }}</h3><ul><li v-for="item in proposal.data.proposal.missing_information" :key="item">{{ item }}</li><li v-for="item in proposal.data.proposal.guardrails" :key="item">{{ t(item) }}</li></ul>
      <details><summary>{{ t('Sources and agent tools') }}</summary><ul><li v-for="(step, index) in proposal.data.trace" :key="index"><code>{{ step.tool }}</code><template v-if="step.document_id"> · <button @click="$emit('document', step.document_id, step.filename)">{{ step.filename || step.document_id }}</button></template><template v-if="step.result?.sourceUrl"> · <a :href="step.result.sourceUrl" target="_blank" rel="noopener noreferrer">{{ t('View official source') }}</a> · {{ t(step.result.status === 'unavailable' ? 'Unavailable' : 'Source consulted') }}</template><template v-if="step.citations"><p v-if="step.text">{{ step.text }}</p><a v-for="citation in step.citations" :key="citation.url" :href="citation.url" target="_blank" rel="noopener noreferrer">{{ citation.title }}</a></template></li></ul></details>
      <div v-if="!proposal.decision" class="human-review"><label for="review-note">{{ t('Human review note') }}</label><textarea id="review-note" rows="3" maxlength="2000" :value="note" :disabled="!!busy" @input="$emit('update:note', $event.target.value)" /><label class="checkbox"><input type="checkbox" :checked="reviewed" :disabled="!!busy" @change="$emit('update:reviewed', $event.target.checked)"><span>{{ t('I reviewed the evidence, policy clauses, proposed amounts and operational constraints. Only eligible operations will be recorded; blocked operations remain blocked.') }}</span></label><div class="review-actions"><UiButton class="dark" :disabled="!!busy || !reviewed || !note.trim() || proposal.stale" @click="$emit('approve')">{{ t('Approve reviewed proposal') }}</UiButton><UiButton :disabled="!!busy || !reviewed || !note.trim()" @click="$emit('reject')">{{ t('Reject proposal') }}</UiButton></div></div>
      <p v-else class="recorded">{{ t('The human decision is saved in the dossier. No payment has been executed.') }}<br>{{ proposal.decision.note }}</p>
    </template>
  </section>
</template>
<style scoped>
.agent-review { display: flex; flex-direction: column; gap: 14px; padding: 5px; }
.intro > span { color: #9797a2; font-size: 11px; } h2 { font-size: 21px; margin-top: 6px; } h3 { font-size: 14px; margin-top: 10px; }
p, li { font-size: 12px; line-height: 1.7; color: #72727e; } small { font-size: 10px; color: #9898a2; } .intro p { margin-top: 8px; }
.checkbox { display: flex; align-items: flex-start; gap: 9px; font-size: 11px; color: #797985; line-height: 1.7; } .checkbox input { margin-top: 3px; accent-color: #222225; }
.proposal-summary { padding: 16px; border-radius: 10px; background: #f5f5f7; } .proposal-summary p { margin: 9px 0; }
.tags { display: flex; flex-wrap: wrap; gap: 8px; } .tags span { font-size: 10px; background: white; border: 1px solid #e2e2e8; padding: 5px 8px; border-radius: 5px; }
.line { border: 1px solid #e7e7ec; padding: 13px; border-radius: 9px; } .line > div { display: flex; align-items: center; justify-content: space-between; gap: 10px; font-size: 12px; } .line strong { font-weight: 550; } .line p { margin-top: 8px; } .line button { margin-top: 8px; font-size: 10px; text-decoration: underline; color: #686875; }
.refund-total { display: flex; justify-content: space-between; gap: 12px; font-size: 13px; } .refund-total strong { font-size: 18px; }
.warning { color: #a07c47; } .error { background: #fff4f3; color: #a24c48; padding: 12px; border-radius: 8px; }
ul { padding-left: 18px; } details { font-size: 11px; border-top: 1px solid #e6e6eb; padding-top: 13px; } summary { cursor: pointer; } details ul { margin-top: 10px; } details li { font-size: 10px; overflow-wrap: anywhere; } details a { display: inline-block; text-decoration: underline; margin: 3px 6px 3px 0; }
.human-review { display: flex; flex-direction: column; gap: 12px; padding-top: 17px; border-top: 1px solid #e5e5eb; } .human-review > label:first-child { font-size: 12px; }
textarea { width: 100%; resize: vertical; border: 1px solid #dedee5; padding: 10px; border-radius: 8px; font-size: 12px; }
.review-actions { display: flex; gap: 12px; } .recorded { padding: 12px; border-radius: 8px; background: #f0f6f2; color: #487058; }
</style>
