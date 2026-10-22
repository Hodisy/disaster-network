<script setup>
import { Mic, FileText, ChevronsUpDown, ChevronRight, ChevronDown, ArrowUp, ArrowUpRight, RotateCcw, X, Check, Plus, Paperclip, Phone, ShieldCheck, House, Waves, CircleCheck, Circle, ArrowLeft, ArrowRight, Building2, Clock3, Image, Trash2, Square, MoreHorizontal, PanelLeftClose } from 'lucide-vue-next'
defineProps({ name: { type: String, required: true }, size: { type: Number, default: 18 } })
const icons = { Mic, FileText, ChevronsUpDown, ChevronRight, ChevronDown, ArrowUp, ArrowUpRight, RotateCcw, X, Check, Plus, Paperclip, Phone, ShieldCheck, House, Waves, CircleCheck, Circle, ArrowLeft, ArrowRight, Building2, Clock3, Image, Trash2, Square, MoreHorizontal, PanelLeftClose }
</script>

<template>
  <component :is="icons[name] || icons.Circle" class="ui-icon" :size="size" :stroke-width="1.7" aria-hidden="true" />
</template>

<style scoped>
.ui-icon { flex-shrink: 0; vertical-align: middle; }
</style>
