<script lang="ts" setup>
import HintMark from './HintMark.vue'
import type { Job } from '../types'

defineProps<{
  jobs: Job[]
  downloadRoot?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  create: []
  edit: [job: Job]
  remove: [job: Job]
  start: [job: Job]
}>()

function platformOf(p: { os: string; architecture: string; variant?: string }) {
  return p.variant ? `${p.os}/${p.architecture}/${p.variant}` : `${p.os}/${p.architecture}`
}
</script>

<template>
  <section class="card">
    <header class="card-head summary-head">
      <h2 class="label-with-hint">
        下载任务
        <HintMark text="一个任务可包含多个镜像，触发后按顺序一次性下载并导出 TAR。下载位置使用设置中的默认根目录。" />
      </h2>
      <button class="btn primary" :disabled="disabled" type="button" @click="emit('create')">新建</button>
    </header>

    <p v-if="!jobs.length" class="empty">还没有任务。创建一个任务，把多个镜像放在一起下载。</p>

    <ul v-else class="job-list">
      <li v-for="job in jobs" :key="job.id" class="job-item">
        <div class="job-item-head">
          <strong class="job-item-name">{{ job.name }}</strong>
          <div class="job-item-actions">
            <button
              class="btn"
              :disabled="disabled || !(job.images || []).length"
              type="button"
              @click="emit('start', job)"
            >
              下载
            </button>
            <button class="btn ghost" :disabled="disabled" type="button" @click="emit('edit', job)">编辑</button>
            <button class="btn ghost danger-text" :disabled="disabled" type="button" @click="emit('remove', job)">
              删除
            </button>
          </div>
        </div>
        <p class="hint job-item-meta">
          {{ (job.images || []).length }} 个镜像
          <template v-if="job.pack"> · 完成后打包 all.tar.zst</template>
          <template v-if="downloadRoot">
            · <span class="mono wrap">{{ downloadRoot }}/{{ job.name }}</span>
          </template>
        </p>
        <ul class="job-images">
          <li v-for="(img, i) in (job.images || [])" :key="i" class="mono">
            {{ img.image }}
            <template v-if="img.targetTag">
              <span class="muted"> → </span>{{ img.targetTag }}
            </template>
            <span class="muted"> {{ platformOf(img.platform) }}</span>
          </li>
        </ul>
      </li>
    </ul>
  </section>
</template>
