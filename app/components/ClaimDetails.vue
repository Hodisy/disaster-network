<script setup>
defineProps({ claim: Object, steps: Array, stepDescriptions: Array })
</script>

<template>
  <aside class="claim-details" aria-label="Informations du dossier">
    <div class="identity-card"><span class="avatar large">CM</span><div><h2>{{ claim.name }}</h2><p>Assurée fictive</p><span class="subtle-tag">Rivage Assurance</span></div></div>
    <section class="detail-section"><h2>Le sinistre<UiIcon name="ArrowUpRight" :size="15" /></h2><dl><div><dt>Dossier</dt><dd>{{ claim.id }}</dd></div><div><dt>Événement</dt><dd>{{ claim.event }}</dd></div><div><dt>Date</dt><dd>{{ claim.date }}</dd></div><div class="address-detail"><dt>Adresse</dt><dd>{{ claim.address }}</dd></div></dl></section>
    <section class="detail-section"><h2>Besoins identifiés</h2><ul class="needs-list"><li v-for="(need, index) in claim.needs" :key="need"><UiIcon :name="['House', 'Waves', 'ShieldCheck'][index]" :size="15" />{{ need }}</li></ul><p class="detail-note">Besoins fictifs, à confirmer dans un parcours réel.</p></section>
    <section class="detail-section"><h2>Le parcours</h2><ol class="stage-list"><li v-for="(step, index) in steps" :key="step" :class="{ done: claim.stage > index, current: claim.stage === index }" :aria-current="claim.stage === index ? 'step' : undefined"><span class="stage-marker"><UiIcon v-if="claim.stage > index" name="Check" :size="11" /><span v-else>{{ index + 1 }}</span></span><span><span class="stage-name">{{ step }}</span><small class="stage-description">{{ stepDescriptions[index] }}</small></span></li></ol></section>
    <p class="contract-note"><UiIcon name="ShieldCheck" :size="14" />Garanties de démonstration, sans valeur contractuelle.</p>
  </aside>
</template>

<style scoped>
.avatar { display: inline-flex; flex-shrink: 0; align-items: center; justify-content: center; width: 33px; height: 33px; border-radius: 50%; background: #19191b; color: #fff; font-size: 11px; font-weight: 500; letter-spacing: -.2px; }
.avatar.insurer-avatar { background: #e8e8eb; color: #57575e; }
.avatar.tiny { width: 38px; height: 38px; background: #f0f0f2; color: #46464d; font-size: 12px; }
.avatar.large { width: 52px; height: 52px; font-size: 18px; }
.subtle-tag { display: inline-flex; align-items: center; gap: 6px; padding: 5px 9px; border-radius: 6px; background: #f3f3f4; color: #62626b; font-size: 10px; font-weight: 400; line-height: 1.2; white-space: nowrap; }
.subtle-tag .small-dot { width: 5px; height: 5px; }
.claim-details { border-left: 1px solid var(--border); padding: 24px 21px 20px; min-width: 0; }
.identity-card { display: flex; align-items: center; gap: 13px; border: 1px solid #e7e7ea; border-radius: 14px; padding: 15px 13px; }
.identity-card > .avatar { flex-shrink: 0; }
.identity-card h2 { font-size: 13px; line-height: 1.5; align-self: end; }
.identity-card p, .identity-card .subtle-tag { color: #85858f; font-size: 10px; line-height: 1.7; }
.identity-card .subtle-tag { background: none; border-radius: 0; padding: 0; }
.detail-section { padding: 20px 0 18px; border-bottom: 1px solid var(--border); }
.detail-section h2 { display: flex; align-items: center; justify-content: space-between; font-size: 13px; line-height: 1.5; letter-spacing: -.2px; }
.detail-section h2 > svg { color: #7f7f88; }
.detail-section dl { margin: 12px 0 0; font-size: 11px; line-height: 1.6; }
.detail-section dl > div { display: flex; gap: 10px; margin-top: 8px; }
.detail-section dt { color: #92929c; }
.detail-section dd { margin: 0; overflow-wrap: anywhere; }
.needs-list { display: flex; flex-direction: column; gap: 14px; margin-top: 17px; }
.needs-list li { display: flex; align-items: center; gap: 13px; font-size: 11px; color: #696973; }
.needs-list svg { color: #29292f; }
.detail-note { color: #a0a0aa; font-size: 9px; line-height: 1.6; margin-top: 15px; }
.stage-list { margin: 15px 0 0 10px; }
.stage-list li { position: relative; display: flex; gap: 17px; padding-bottom: 16px; color: #8b8b95; font-size: 11px; line-height: 1.5; }
.stage-list li::before { content: ''; position: absolute; left: 10px; width: 1px; top: 21px; bottom: 0; background: #e5e5e9; }
.stage-list li:last-child { padding-bottom: 0; }
.stage-list li:last-child::before { display: none; }
.stage-marker { position: relative; display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0; width: 22px; height: 22px; border-radius: 50%; border: 1px solid #e3e3e7; background: #fff; font-size: 10px; }
.stage-list li.current { color: #202025; font-weight: 500; }
.stage-list li.current .stage-marker { color: #fff; background: #171719; border-color: #171719; }
.stage-list li.done .stage-marker { background: #ebf9f0; border-color: #ebf9f0; color: #37a76a; }
.stage-name { display: block; }
.stage-description { display: block; font-size: 10px; line-height: 1.45; color: #a0a0aa; margin-top: 1px; font-weight: 400; }
.contract-note { display: flex; align-items: flex-start; gap: 6px; font-size: 9px; color: #a0a0aa; line-height: 1.65; margin-top: 15px; }
.contract-note svg { margin-top: 1px; }
.message .avatar.tiny, .small-mark { width: 26px; height: 26px; font-size: 8px; }
.claim-details { width: 330px; flex-shrink: 0; }
.detail-section dt { width: 92px; flex-shrink: 0; }
.detail-section dd { flex: 1; min-width: 0; }
</style>
