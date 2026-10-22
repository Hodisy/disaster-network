<script setup>
const { t } = useAppI18n()
const props = defineProps({ active: String, account: String, menuOpen: Boolean, locale: String })
const emit = defineEmits(['toggle-account', 'close-account', 'switch-account', 'set-locale'])
const accountTrigger = ref(null)
const accountPopover = ref(null)
watch(() => props.menuOpen, async open => {
  await nextTick()
  if (open) accountPopover.value?.querySelector('.account-option')?.focus()
  else accountTrigger.value?.focus()
})
function closeOverlays() {
  if (props.menuOpen) emit('close-account')
}
</script>

<template>
  <div class="app-shell" @keydown.esc="closeOverlays">
    <a class="skip-link" href="#main-content">{{ t('Skip to content') }}</a>
    <aside class="sidebar">
      <NuxtLink to="/hotline" class="brand" :aria-label="t('Disaster Network, hotline home')">Disaster Network<span class="brand-period">.</span></NuxtLink>
      <div class="sidebar-label">{{ t(account === 'client' ? 'Client workspace' : 'Insurer workspace') }}</div>
      <nav v-if="account === 'client'" class="main-nav" :aria-label="t('Client navigation')">
        <NuxtLink to="/hotline" :class="{ selected: active === 'hotline' }" :aria-current="active === 'hotline' ? 'page' : undefined"><UiIcon name="Mic" />{{ t('Hotline') }}</NuxtLink>
        <NuxtLink to="/client" :class="{ selected: active === 'client' }" :aria-current="active === 'client' ? 'page' : undefined"><UiIcon name="FileText" />{{ t('My claim') }}</NuxtLink>
      </nav>
      <nav v-else class="main-nav" :aria-label="t('Insurer navigation')"><NuxtLink to="/insurer" class="selected" aria-current="page"><UiIcon name="Building2" />{{ t('Operations') }}</NuxtLink></nav>
      <div class="sidebar-bottom">
        <div class="account-anchor">
          <button v-if="menuOpen" class="dismiss-layer" :aria-label="t('Close account switcher')" tabindex="-1" @click="$emit('close-account')" />
          <div v-if="menuOpen" ref="accountPopover" class="account-popover" role="dialog" :aria-label="t('Switch account')">
            <div class="popover-heading">{{ t('Switch account') }}<button class="icon-button" :aria-label="t('Close')" @click="$emit('close-account')"><UiIcon name="X" :size="15" /></button></div>
            <button class="account-option" :class="{ current: account === 'client' }" @click="$emit('switch-account', 'client')"><span class="avatar">J</span><span>Jules<small>{{ t('Client account') }}</small></span><UiIcon v-if="account === 'client'" name="Check" :size="16" /></button>
            <button class="account-option" :class="{ current: account === 'insurer' }" @click="$emit('switch-account', 'insurer')"><span class="avatar insurer-avatar">AS</span><span>Allstorm Insurance<small>{{ t('Insurer account') }}</small></span><UiIcon v-if="account === 'insurer'" name="Check" :size="16" /></button>
          </div>
          <button ref="accountTrigger" class="account-button" :aria-expanded="menuOpen" aria-haspopup="dialog" :aria-label="t('Switch account: {name}, {role}', { name: account === 'client' ? 'Jules' : 'Allstorm Insurance', role: t(account === 'client' ? 'Client account' : 'Insurer account') })" @click="$emit('toggle-account')"><span class="avatar" :class="{ 'insurer-avatar': account === 'insurer' }">{{ account === 'client' ? 'J' : 'AS' }}</span><span>{{ account === 'client' ? 'Jules' : 'Allstorm Insurance' }}<small>{{ t(account === 'client' ? 'Client account' : 'Insurer account') }}</small></span><UiIcon name="ChevronsUpDown" :size="15" /></button>
        </div>
      </div>
    </aside>
    <div class="workspace">
      <header class="topbar"><div class="breadcrumb">{{ t(account === 'client' ? 'Client workspace' : 'Insurer workspace') }}<UiIcon name="ChevronRight" :size="13" /><span>{{ t(active === 'hotline' ? 'Hotline' : active === 'client' ? 'My claim' : active === 'settings' ? 'Settings' : 'Operations') }}</span></div><div class="topbar-controls"><NuxtLink to="/settings" class="settings-link" :class="{ selected: active === 'settings' }" :aria-label="t('Settings')" :title="t('Settings')" :aria-current="active === 'settings' ? 'page' : undefined"><UiIcon name="Settings" :size="16" /></NuxtLink><LanguageToggle :locale="locale" @change="$emit('set-locale', $event)" /></div></header>
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
.dismiss-layer { position: fixed; z-index: 39; inset: 0; cursor: default; }
.account-popover { position: absolute; z-index: 40; left: 0; bottom: calc(100% + 6px); width: 258px; padding: 10px; background: #fff; border: 1px solid var(--border); border-radius: 12px; box-shadow: 0 9px 30px #14141b13, 0 2px 5px #14141b05; animation: pop-in .15s ease-out; }
.popover-heading { display: flex; align-items: center; justify-content: space-between; padding: 4px 7px 10px; font-size: 11px; font-weight: 500; }
.account-option { display: flex; align-items: center; gap: 10px; width: 100%; border-radius: 7px; padding: 9px 7px; text-align: left; font-size: 12px; }
.account-option > span:nth-child(2) { flex: 1; }
.account-option:hover { background: #f5f5f6; }
.account-option.current { background: #fafafa; }
.workspace { display: flex; flex: 1; flex-direction: column; margin-left: 260px; min-width: 0; }
.topbar { display: flex; flex-shrink: 0; align-items: center; justify-content: space-between; gap: 16px; height: 58px; padding: 0 28px; border-bottom: 1px solid var(--border); }
.settings-link { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; color: #93939c; border-radius: 6px; }
.settings-link:hover, .settings-link.selected { color: var(--ink); background: #f1f1f3; }
.settings-link:focus-visible { outline: 2px solid #74747c; outline-offset: 3px; }
.topbar-controls { display: flex; align-items: center; gap: 18px; }
.breadcrumb { display: flex; align-items: center; gap: 13px; font-size: 12px; color: #75757f; }
.breadcrumb > span { color: var(--ink); }
.icon-button { display: inline-flex; align-items: center; justify-content: center; width: 27px; height: 27px; border-radius: 6px; color: #787882; padding: 0; }
.icon-button:hover { background: #f1f1f3; color: var(--ink); }
.icon-button:disabled { opacity: .4; }
main { flex: 1; min-height: 0; outline: none; }
@keyframes pop-in { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
</style>
