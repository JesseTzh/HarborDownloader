<script lang="ts" setup>
import type { JobItem, ProgressEvent } from '../types'
import DownloadProgress from './DownloadProgress.vue'

defineProps<{
  name: string
  status?: string
  current: number
  total: number
  items: JobItem[]
  progress: ProgressEvent | null
  cancelling?: boolean
}>()

const emit = defineEmits<{
  cancel: []
}>()

function statusLabel(status: string) {
  switch (status) {
    case 'pending':
      return '等待'
    case 'checking':
      return '检查中'
    case 'downloading':
      return '下载中'
    case 'exporting':
      return '导出中'
    case 'packaging':
      return '打包中'
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

function tagClass(status: string) {
  if (status === 'completed') return 'ok'
  if (status === 'failed' || status === 'cancelled') return 'warn'
  return 'muted'
}

const currentImage = (items: JobItem[], current: number, progress: ProgressEvent | null) =>
  progress?.image || items[Math.max(0, current - 1)]?.image || ''
</script>

<template>
  <section class="card">
    <header class="card-head">
      <h2>{{ name }}</h2>
      <p class="hint">
        <template v-if="status === 'packaging'">正在打包 all.tar.zst</template>
        <template v-else>正在下载 {{ current }} / {{ total }}</template>
      </p>
    </header>
    <ol class="job-run-items">
      <li v-for="(item, i) in items" :key="i">
        <span class="tag" :class="tagClass(item.status)">{{ statusLabel(item.status) }}</span>
        <span class="mono wrap">{{ item.image }}</span>
        <span v-if="item.targetTag" class="muted">→</span>
        <span v-if="item.targetTag" class="mono wrap">{{ item.targetTag }}</span>
        <span class="muted">{{ item.platform }}</span>
      </li>
    </ol>
  </section>
  <DownloadProgress
    :image="currentImage(items, current, progress)"
    :progress="progress"
    :cancelling="cancelling"
    @cancel="emit('cancel')"
  />
</template>
