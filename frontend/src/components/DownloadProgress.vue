<script lang="ts" setup>
import type { ProgressEvent } from '../types'

const props = defineProps<{
  image: string
  progress: ProgressEvent | null
  cancelling?: boolean
}>()

const emit = defineEmits<{
  cancel: []
}>()

function fmtBytes(n: number) {
  if (!n || n < 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let v = n
  let i = 0
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(i === 0 ? 0 : 2)} ${units[i]}`
}

function fmtSpeed(n: number) {
  return `${fmtBytes(n)}/s`
}

function fmtEta(sec: number) {
  if (!sec || sec < 0) return '—'
  if (sec < 60) return `${Math.ceil(sec)} sec`
  const m = Math.floor(sec / 60)
  const s = Math.ceil(sec % 60)
  return `${m} min ${s} sec`
}

const pct = () => Math.max(0, Math.min(100, props.progress?.percentage ?? 0))
</script>

<template>
  <section class="card progress-card">
    <header class="card-head">
      <h2>{{
        progress?.phase === 'packaging' ? 'Packaging' : progress?.phase === 'exporting' ? 'Exporting' : 'Downloading'
      }}</h2>
    </header>
    <p class="image-line">{{ image }}</p>
    <div class="bar" :aria-valuenow="pct()" aria-valuemin="0" aria-valuemax="100" role="progressbar">
      <div class="bar-fill" :style="{ width: pct() + '%' }" />
    </div>
    <p class="pct">{{ pct().toFixed(1) }}%</p>
    <dl class="stats">
      <div>
        <dt>已下载</dt>
        <dd>{{ fmtBytes(progress?.downloadedBytes ?? 0) }} / {{ fmtBytes(progress?.total ?? 0) }}</dd>
      </div>
      <div>
        <dt>Speed</dt>
        <dd>{{ fmtSpeed(progress?.speedBytes ?? 0) }}</dd>
      </div>
      <div>
        <dt>ETA</dt>
        <dd>{{ fmtEta(progress?.etaSeconds ?? 0) }}</dd>
      </div>
      <div v-if="progress?.phase !== 'packaging'">
        <dt>Layer</dt>
        <dd>{{ progress?.currentLayer ?? 0 }} / {{ progress?.totalLayers ?? 0 }}</dd>
      </div>
      <div v-if="progress?.phase !== 'packaging' && (progress?.cachedLayers ?? 0) > 0">
        <dt>缓存命中</dt>
        <dd>{{ progress?.cachedLayers ?? 0 }} / {{ progress?.totalLayers ?? 0 }}</dd>
      </div>
    </dl>
    <p v-if="progress?.message" class="hint">{{ progress.message }}</p>
    <div class="actions">
      <button class="btn danger" :disabled="cancelling" type="button" @click="emit('cancel')">
        {{ cancelling ? '正在取消…' : '取消' }}
      </button>
    </div>
  </section>
</template>
