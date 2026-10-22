<script setup>
const props = defineProps({ active: String, account: String, menuOpen: Boolean, resetOpen: Boolean })
const emit = defineEmits(['toggle-account', 'close-account', 'switch-account', 'toggle-reset', 'reset'])
const accountTrigger = ref(null)
const accountPopover = ref(null)
watch(() => props.menuOpen, async open => {
  await nextTick()
  if (open) accountPopover.value?.querySelector('.account-option')?.focus()
  else accountTrigger.value?.focus()
})
function closeOverlays() {
  if (props.menuOpen) emit('close-account')
  if (props.resetOpen) emit('toggle-reset')
}
</script>

<template>
  <div class="app-shell" @keydown.esc="closeOverlays">
    <a class="skip-link" href="#main-content">Aller au contenu</a>
    <aside class="sidebar">
      <NuxtLink to="/hotline" class="brand" aria-label="Disaster Network, accueil hotline">Disaster Network<span class="brand-period">.</span></NuxtLink>
      <div class="sidebar-label">{{ account === 'client' ? 'Espace client' : 'Espace assureur' }}</div>
      <nav v-if="account === 'client'" class="main-nav" aria-label="Menu client">
        <NuxtLink to="/hotline" :class="{ selected: active === 'hotline' }" :aria-current="active === 'hotline' ? 'page' : undefined"><UiIcon name="Mic" />Hotline</NuxtLink>
        <NuxtLink to="/client" :class="{ selected: active === 'client' }" :aria-current="active === 'client' ? 'page' : undefined"><UiIcon name="FileText" />Mon dossier</NuxtLink>
      </nav>
      <nav v-else class="main-nav" aria-label="Menu assureur"><NuxtLink to="/assureur" class="selected" aria-current="page"><UiIcon name="Building2" />Vue opérationnelle</NuxtLink></nav>
      <div class="sidebar-bottom">
        <div class="account-anchor">
          <button v-if="menuOpen" class="dismiss-layer" aria-label="Fermer le sélecteur de compte" tabindex="-1" @click="$emit('close-account')" />
          <div v-if="menuOpen" ref="accountPopover" class="account-popover" role="dialog" aria-label="Changer de compte">
            <div class="popover-heading">Changer de compte<button class="icon-button" aria-label="Fermer" @click="$emit('close-account')"><UiIcon name="X" :size="15" /></button></div>
            <button class="account-option" :class="{ current: account === 'client' }" @click="$emit('switch-account', 'client')"><span class="avatar">CM</span><span>Claire Martin<small>Compte client</small></span><UiIcon v-if="account === 'client'" name="Check" :size="16" /></button>
            <button class="account-option" :class="{ current: account === 'insurer' }" @click="$emit('switch-account', 'insurer')"><span class="avatar insurer-avatar">RA</span><span>Rivage Assurance<small>Compte assureur · À venir</small></span><UiIcon v-if="account === 'insurer'" name="Check" :size="16" /></button>
          </div>
          <button ref="accountTrigger" class="account-button" :aria-expanded="menuOpen" aria-haspopup="dialog" :aria-label="`Changer de compte : ${account === 'client' ? 'Claire Martin, compte client' : 'Rivage Assurance, compte assureur'}`" @click="$emit('toggle-account')"><span class="avatar" :class="{ 'insurer-avatar': account === 'insurer' }">{{ account === 'client' ? 'CM' : 'RA' }}</span><span>{{ account === 'client' ? 'Claire Martin' : 'Rivage Assurance' }}<small>{{ account === 'client' ? 'Compte client' : 'Compte assureur' }}</small></span><UiIcon name="ChevronsUpDown" :size="15" /></button>
        </div>
        <div v-if="resetOpen" class="reset-confirm"><p>Recommencer le parcours ? Les pièces locales et les échanges seront effacés.</p><div><UiButton class="small" @click="$emit('toggle-reset')">Annuler</UiButton><UiButton class="dark small" @click="$emit('reset')">Recommencer</UiButton></div></div>
        <button class="reset-button" :aria-expanded="resetOpen" @click="$emit('toggle-reset')"><UiIcon name="RotateCcw" :size="14" />Réinitialiser la démo</button>
      </div>
    </aside>
    <div class="workspace">
      <header class="topbar"><div class="breadcrumb">{{ account === 'client' ? 'Espace client' : 'Espace assureur' }}<UiIcon name="ChevronRight" :size="13" /><span>{{ active === 'hotline' ? 'Hotline' : active === 'client' ? 'Mon dossier' : 'Vue opérationnelle' }}</span></div><span class="demo-tag"><span class="small-dot" />Démonstration</span></header>
      <main id="main-content" tabindex="-1"><slot /></main>
    </div>
  </div>
</template>

<style scoped>
.app-shell { display: flex; min-height: 100%; }
.skip-link { position: fixed; top: -70px; left: 16px; z-index: 100; padding: 12px; border-radius: 8px; background: var(--ink); color: white; }
.skip-link:focus { top: 12px; }
.sidebar { position: fixed; inset: 0 auto 0 0; z-index: 20; width: 260px; background: var(--surface); border-right: 1px solid var(--border); display: flex; flex-direction: column; padding: 25px 14px 14px; }
.brand { display: block; width: fit-content; padding: 0 11px; font-size: 22px; font-weight: 750; line-height: 1.3; letter-spacing: -.9px; }
.brand span { letter-spacing: -2px; }
.brand .brand-period { letter-spacing: 0; }
.sidebar-label { padding: 0 12px; margin: 44px 0 15px; font-size: 10px; line-height: 1.5; font-weight: 500; letter-spacing: .8px; text-transform: uppercase; color: #74747c; }
.main-nav { display: flex; flex-direction: column; gap: 5px; }
.main-nav a { display: flex; align-items: center; gap: 16px; padding: 13px 17px; min-height: 46px; border-radius: 9px; font-size: 14px; line-height: 1.4; transition: background .16s; }
.main-nav a:hover { background: #ededec; }
.main-nav a.selected { background: #e9e9e8; color: #101012; }
.main-nav svg { width: 20px; height: 20px; }
.sidebar-bottom { margin-top: auto; position: relative; padding-top: 36px; }
.account-anchor { position: relative; }
.account-button { width: 100%; display: flex; align-items: center; gap: 11px; text-align: left; padding: 11px 10px; border-radius: 9px; }
.account-button:hover { background: #ededec; }
.account-button > span:nth-child(2) { flex: 1; min-width: 0; font-size: 13px; line-height: 1.5; }
.account-button small, .account-option small { display: block; color: var(--muted); font-size: 11px; line-height: 1.6; font-weight: 400; }
.avatar { display: inline-flex; flex-shrink: 0; align-items: center; justify-content: center; width: 33px; height: 33px; border-radius: 50%; background: #19191b; color: #fff; font-size: 11px; font-weight: 500; letter-spacing: -.2px; }
.avatar.insurer-avatar { background: #e8e8eb; color: #57575e; }
.reset-button { display: flex; align-items: center; gap: 12px; width: calc(100% - 20px); margin: 9px 10px 0; padding: 17px 2px 6px; border-top: 1px solid #e3e3e4; color: #5b5b63; font-size: 11px; text-align: left; }
.reset-button:hover { color: var(--ink); }
.dismiss-layer { position: fixed; z-index: 39; inset: 0; cursor: default; }
.account-popover { position: absolute; z-index: 40; left: 0; bottom: calc(100% + 6px); width: 258px; padding: 10px; background: #fff; border: 1px solid var(--border); border-radius: 12px; box-shadow: 0 9px 30px #14141b13, 0 2px 5px #14141b05; animation: pop-in .15s ease-out; }
.popover-heading { display: flex; align-items: center; justify-content: space-between; padding: 4px 7px 10px; font-size: 11px; font-weight: 500; }
.account-option { display: flex; align-items: center; gap: 10px; width: 100%; border-radius: 7px; padding: 9px 7px; text-align: left; font-size: 12px; }
.account-option > span:nth-child(2) { flex: 1; }
.account-option:hover { background: #f5f5f6; }
.account-option.current { background: #fafafa; }
.reset-confirm { padding: 13px; border: 1px solid var(--border); border-radius: 10px; margin: 10px 0; background: white; box-shadow: 0 2px 10px #00000005; }
.reset-confirm p { font-size: 12px; line-height: 1.6; color: #696973; }
.reset-confirm > div { display: flex; gap: 6px; margin-top: 12px; }
.workspace { display: flex; flex: 1; flex-direction: column; margin-left: 260px; min-width: 0; }
.topbar { display: flex; flex-shrink: 0; align-items: center; justify-content: space-between; gap: 16px; height: 58px; padding: 0 28px; border-bottom: 1px solid var(--border); }
.breadcrumb { display: flex; align-items: center; gap: 13px; font-size: 12px; color: #75757f; }
.breadcrumb > span { color: var(--ink); }
.demo-tag { display: inline-flex; align-items: center; gap: 8px; color: #777780; font-size: 11px; white-space: nowrap; }
.small-dot { display: inline-block; width: 7px; height: 7px; flex-shrink: 0; background: #b1b1b7; border-radius: 50%; }
.icon-button { display: inline-flex; align-items: center; justify-content: center; width: 27px; height: 27px; border-radius: 6px; color: #787882; padding: 0; }
.icon-button:hover { background: #f1f1f3; color: var(--ink); }
.icon-button:disabled { opacity: .4; }
main { flex: 1; min-height: 0; outline: none; }
@keyframes pop-in { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
</style>
