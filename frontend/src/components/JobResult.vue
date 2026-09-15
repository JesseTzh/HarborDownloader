<script lang="ts" setup>
import type { JobCompleteEvent, JobItem } from '../types'

const props = defineProps<{
  result: JobCompleteEvent
}>()

const emit = defineEmits<{
  open: [path: string]
  copy: [path: string]
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

function fileName(p: string) {
  const parts = (p || '').split(/[/\\]/)
  return parts[parts.length - 1] || p
}

function statusLabel(status: string) {
  switch (status) {
    case 'completed':
      return '完成'
    case 'cancelled':
      return '已取消'
    case 'failed':
      return '失败'
    default:
      return status
  }
}

const counts = () => {
  const items = props.result.items || []
  return {
    ok: items.filter((it: JobItem) => it.status === 'completed').length,
    fail: items.filter((it: JobItem) => it.status === 'failed').length,
    cancel: items.filter((it: JobItem) => it.status === 'cancelled').length,
    total: items.length,
  }
}

const title = () => {
  if (props.result.status === 'cancelled') return '任务已取消'
  const c = counts()
  if (c.fail > 0 && c.ok > 0) return '部分镜像下载失败'
  if (c.fail > 0 || c.ok === 0) return '下载失败'
  return '下载完成'
}

const failed = () => props.result.status !== 'completed'
</script>

<template>
  <section class="card result-card" :class="{ failed: failed() }">
    <header class="card-head">
      <h2>{{ title() }}</h2>
      <p class="hint">{{ result.name }} · 成功 {{ counts().ok }} / {{ counts().total }}</p>
    </header>
    <p v-if="result.error && failed()" class="error-block">{{ result.error }}</p>
    <div v-if="result.packagePath" class="pack-result">
      <div>
        <span class="tag ok">打包</span>
        <strong class="mono wrap">{{ fileName(result.packagePath) }}</strong>
        <p class="hint">{{ result.packagePath }}</p>
      </div>
      <div class="job-item-actions">
        <button class="btn ghost" type="button" @click="emit('open', result.packagePath)">打开</button>
        <button class="btn ghost" type="button" @click="emit('copy', result.packagePath)">复制</button>
      </div>
    </div>
    <ul class="job-result-list">
      <li v-for="(item, i) in result.items" :key="i" class="job-result-item">
        <div>
          <span class="tag" :class="item.status === 'completed' ? 'ok' : item.status === 'failed' ? 'warn' : 'muted'">
            {{ statusLabel(item.status) }}
          </span>
          <strong class="mono wrap">{{ item.image }}</strong>
          <p v-if="item.targetTag" class="hint">保存为 <strong class="mono">{{ item.targetTag }}</strong></p>
          <p class="hint">{{ item.platform }}<template v-if="item.size"> · {{ fmtBytes(item.size) }}</template></p>
          <p v-if="item.error" class="hint bad">{{ item.error }}</p>
          <p v-else-if="item.outputPath" class="hint">
            {{ fileName(item.outputPath) }}
          </p>
        </div>
        <div v-if="item.outputPath && item.status === 'completed'" class="job-item-actions">
          <button class="btn ghost" type="button" @click="emit('open', item.outputPath)">打开</button>
          <button class="btn ghost" type="button" @click="emit('copy', item.outputPath)">复制</button>
        </div>
      </li>
    </ul>
    <div class="actions">
      <button class="btn ghost" type="button" @click="emit('done')">完成</button>
    </div>
  </section>
</template>
