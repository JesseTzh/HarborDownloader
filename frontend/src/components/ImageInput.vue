<script lang="ts" setup>
import type { ImageReference } from '../types'

defineProps<{
  disabled?: boolean
  parsed?: ImageReference | null
  parseError?: string
}>()

const image = defineModel<string>('image', { required: true })

const emit = defineEmits<{
  parse: []
}>()
</script>

<template>
  <div class="field">
    <label for="image">Image</label>
    <div class="row">
      <input
        id="image"
        v-model="image"
        :disabled="disabled"
        placeholder="project/app-gateway:1.0.0 或 harbor.company.local/project/app-gateway:1.0.0"
        spellcheck="false"
        @keydown.enter.prevent="emit('parse')"
      />
      <button class="btn ghost" :disabled="disabled || !image" type="button" @click="emit('parse')">解析镜像</button>
    </div>
    <p v-if="parseError" class="hint bad">{{ parseError }}</p>
    <p v-else-if="parsed" class="hint">
      Registry <strong>{{ parsed.registry }}</strong>
      · Repository <strong>{{ parsed.repository }}</strong>
      · Tag <strong>{{ parsed.tag || parsed.digest }}</strong>
    </p>
  </div>
</template>
