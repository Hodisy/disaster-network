<script setup>
const route = useRoute()
const { state, reset, stopCall } = useDemo()
const active = computed(() => route.path.split('/')[1] || 'hotline')
const account = computed(() => active.value === 'assureur' ? 'insurer' : 'client')

async function switchAccount(value) {
  state.value.accountMenu = false
  stopCall()
  if (value === 'insurer') await navigateTo('/assureur')
  else if (account.value === 'insurer') await navigateTo(state.value.claim ? '/client' : '/hotline')
}
async function restart() { reset(); await navigateTo('/hotline') }
watch(() => route.path, () => { state.value.accountMenu = false; state.value.resetMenu = false })
</script>

<template>
  <DemoShell :active="active" :account="account" :menu-open="state.accountMenu" :reset-open="state.resetMenu" @toggle-account="state.accountMenu = !state.accountMenu; state.resetMenu = false" @close-account="state.accountMenu = false" @switch-account="switchAccount" @toggle-reset="state.resetMenu = !state.resetMenu; state.accountMenu = false" @reset="restart">
    <div v-if="state.notice" class="toast" role="status">{{ state.notice }}<button class="icon-button" aria-label="Fermer la notification" @click="state.notice = ''"><UiIcon name="X" :size="14" /></button></div>
    <NuxtPage />
  </DemoShell>
</template>

<style scoped>
.icon-button { display: inline-flex; align-items: center; justify-content: center; width: 27px; height: 27px; border-radius: 6px; color: #787882; padding: 0; }
.icon-button:hover { background: #f1f1f3; color: var(--ink); }
.icon-button:disabled { opacity: .4; }
.toast { position: fixed; z-index: 50; top: 70px; right: 20px; display: flex; align-items: center; gap: 15px; padding: 8px 12px 8px 16px; background: #fff; border: 1px solid var(--border); border-radius: 10px; box-shadow: 0 7px 25px #00000010; font-size: 12px; }
.file-list .icon-button { height: 25px; width: 25px; margin-right: 3px; }
</style>
