<script lang="ts" setup>
import type { Platform } from '../types'

defineProps<{
  disabled?: boolean
  platforms: Platform[]
}>()

const model = defineModel<Platform>({ required: true })

function keyOf(p: Platform) {
  return p.variant ? `${p.os}/${p.architecture}/${p.variant}` : `${p.os}/${p.architecture}`
}

function onChange(ev: Event) {
  const value = (ev.target as HTMLSelectElement).value
  const found = (ev.target as HTMLSelectElement)
  void found
  const parts = value.split('/')
  model.value = { os: parts[0], architecture: parts[1], variant: parts[2] || '' }
}
</script>

<template>
  <div class="field">
    <label for="platform">Platform</label>
    <select id="platform" :disabled="disabled" :value="keyOf(model)" @change="onChange">
      <option v-for="p in platforms" :key="keyOf(p)" :value="keyOf(p)">
        {{ keyOf(p) }}
      </option>
    </select>
  </div>
</template>
