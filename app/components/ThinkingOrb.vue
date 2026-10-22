<script setup>
import { MODE_FRAMES, resolvePreset, paintFrame } from '~/vendor/thinking-orbs/engine.js'

const props = defineProps({ state: { type: String, default: 'breathing' }, size: { type: Number, default: 240 } })
const canvas = ref(null)
let frameId = 0
let motionQuery
let visible = true
let observer

function restart() {
  cancelAnimationFrame(frameId)
  const element = canvas.value
  if (!element) return
  const context = element.getContext('2d')
  if (!context) return
  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  element.width = props.size * dpr
  element.height = props.size * dpr
  const preset = resolvePreset(props.state, 64)
  const draw = time => {
    context.setTransform(dpr, 0, 0, dpr, 0, 0)
    context.clearRect(0, 0, props.size, props.size)
    paintFrame(context, MODE_FRAMES[preset.mode](props.size, time * preset.speed * 0.5, preset.opts), false)
  }
  if (motionQuery?.matches) { draw(0.6); return }
  const loop = time => {
    draw(time / 1000)
    if (visible && !document.hidden) frameId = requestAnimationFrame(loop)
  }
  if (visible && !document.hidden) frameId = requestAnimationFrame(loop)
}

onMounted(() => {
  motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  motionQuery.addEventListener('change', restart)
  document.addEventListener('visibilitychange', restart)
  observer = new IntersectionObserver(entries => { visible = entries[0].isIntersecting; restart() })
  observer.observe(canvas.value)
  restart()
})
watch(() => [props.state, props.size], restart)
onBeforeUnmount(() => {
  cancelAnimationFrame(frameId)
  observer?.disconnect()
  motionQuery?.removeEventListener('change', restart)
  document.removeEventListener('visibilitychange', restart)
})
</script>

<template>
  <canvas ref="canvas" class="thinking-orb" :style="{ width: `${size}px`, height: `${size}px` }" aria-hidden="true" />
</template>

<style scoped>
.thinking-orb { display: block; flex-shrink: 0; }
</style>
