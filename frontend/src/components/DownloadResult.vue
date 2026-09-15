<script lang="ts" setup>
import type { DownloadCompleteEvent } from '../types'

const props = defineProps<{
  result: DownloadCompleteEvent
  error?: string
  failed?: boolean
}>()

const emit = defineEmits<{
  open: []
  copy: []
  done: []
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
  return `${v.toFixed(2)} ${units[i]}`
}

const fileName = () => {
  const p = props.result.outputPath || ''
  const parts = p.split(/[/\\]/)
  return parts[parts.length - 1] || p
}
</script>

<template>
  <section class="card result-card" :class="{ failed: failed }">
    <header class="card-head">
      <h2 v-if="failed">下载失败</h2>
      <h2 v-else>✓ Download completed</h2>
    </header>
    <p v-if="error" class="error-block">{{ error }}</p>
    <dl v-else class="kv">
      <div>
        <dt>Image</dt>
        <dd>{{ result.image }}</dd>
      </div>
      <div>
        <dt>Platform</dt>
        <dd>{{ result.platform }}</dd>
      </div>
      <div>
        <dt>Digest</dt>
        <dd class="mono">{{ result.digest || '—' }}</dd>
      </div>
      <div>
        <dt>Output</dt>
        <dd>
          {{ result.outputPath.replace(fileName(), '') }}<br />
          <strong>{{ fileName() }}</strong>
        </dd>
      </div>
      <div>
        <dt>Size</dt>
        <dd>{{ fmtBytes(result.size) }}</dd>
      </div>
    </dl>
    <div class="actions">
      <button v-if="!failed" class="btn" type="button" @click="emit('open')">打开文件夹</button>
      <button v-if="!failed" class="btn ghost" type="button" @click="emit('copy')">复制路径</button>
      <button class="btn ghost" type="button" @click="emit('done')">完成</button>
    </div>
  </section>
</template>
