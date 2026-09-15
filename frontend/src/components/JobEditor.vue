<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import type { ImageReference, Job, JobImage, Platform } from '../types'

const props = defineProps<{
  job: Job
  platforms: Platform[]
  disabled?: boolean
  saving?: boolean
}>()

const emit = defineEmits<{
  save: [job: Job, start: boolean]
  cancel: []
  parse: [image: string, index: number]
}>()

function cloneImages(list: JobImage[] | undefined): JobImage[] {
  const rows = (list || []).map((img) => ({
    image: img.image,
    targetTag: img.targetTag || '',
    platform: { ...img.platform },
  }))
  if (rows.length) return rows
  return [{ image: '', targetTag: '', platform: { os: 'linux', architecture: 'amd64' } }]
}

const name = ref(props.job.name)
const pack = ref(!!props.job.pack)
const images = ref<JobImage[]>(cloneImages(props.job.images))
const parsed = ref<Record<number, ImageReference | null>>({})
const parseError = ref<Record<number, string>>({})

watch(
  () => props.job,
  (job) => {
    name.value = job.name
    pack.value = !!job.pack
    images.value = cloneImages(job.images)
    parsed.value = {}
    parseError.value = {}
  }
)

const canSave = computed(() => {
  const hasImage = images.value.some((img) => img.image.trim())
  return name.value.trim() !== '' && hasImage
})

function keyOf(p: Platform) {
  return p.variant ? `${p.os}/${p.architecture}/${p.variant}` : `${p.os}/${p.architecture}`
}

function onPlatform(i: number, ev: Event) {
  const value = (ev.target as HTMLSelectElement).value
  const parts = value.split('/')
  images.value[i] = {
    ...images.value[i],
    platform: { os: parts[0], architecture: parts[1], variant: parts[2] || '' },
  }
}

function addImage() {
  const last = images.value[images.value.length - 1]
  const platform = last?.platform?.os ? { ...last.platform } : { os: 'linux', architecture: 'amd64' }
  images.value = [...images.value, { image: '', targetTag: '', platform }]
}

function removeImage(i: number) {
  images.value = images.value.filter((_, idx) => idx !== i)
}

function onParse(i: number) {
  parseError.value = { ...parseError.value, [i]: '' }
  emit('parse', images.value[i].image.trim(), i)
}

function setParseResult(index: number, value: ImageReference | null, error: string) {
  parsed.value = { ...parsed.value, [index]: value }
  parseError.value = { ...parseError.value, [index]: error }
}

function draft(): Job {
  return {
    id: props.job.id,
    name: name.value.trim(),
    pack: pack.value,
    images: images.value.map((img) => ({
      image: img.image.trim(),
      targetTag: (img.targetTag || '').trim(),
      platform: { ...img.platform },
    })),
  }
}

function onSave(start = false) {
  emit('save', draft(), start)
}

defineExpose({ setParseResult })
</script>

<template>
  <section class="card">
    <header class="card-head">
      <h2>{{ job.id ? '编辑任务' : '新建任务' }}</h2>
    </header>
    <div class="field">
      <label for="job-name">任务名称</label>
      <input
        id="job-name"
        v-model="name"
        :disabled="disabled"
        placeholder="例如：生产环境镜像"
        spellcheck="false"
      />
    </div>
    <label class="check">
      <input v-model="pack" :disabled="disabled" type="checkbox" />
      是否打包
    </label>
    <p class="hint">勾选后，全部 TAR 下载完成后会合并并使用 zstd -19 压缩为 all.tar.zst。下载目录在「设置」中配置，本任务会自动使用同名子文件夹。</p>

    <div class="field">
      <label>镜像列表</label>
      <p class="hint">每个镜像可单独选择平台，并可填写「保存为」：导出的 TAR 会以该名称作为 docker load 后的镜像 tag，文件名也按该名称生成。留空则保持原镜像名。</p>
    </div>

    <div v-for="(img, i) in images" :key="i" class="job-image-block">
      <div class="job-image-row">
        <input
          v-model="img.image"
          :disabled="disabled"
          :id="'job-image-' + i"
          placeholder="project/app-gateway:1.0.0 或 harbor.company.local/project/app-gateway:1.0.0"
          spellcheck="false"
          @keydown.enter.prevent="onParse(i)"
        />
        <select :disabled="disabled" :value="keyOf(img.platform)" @change="onPlatform(i, $event)">
          <option v-for="p in platforms" :key="keyOf(p)" :value="keyOf(p)">{{ keyOf(p) }}</option>
        </select>
        <button class="btn ghost" :disabled="disabled || !img.image.trim()" type="button" @click="onParse(i)">解析</button>
        <button class="btn ghost" :disabled="disabled || images.length <= 1" type="button" @click="removeImage(i)">删除</button>
      </div>
      <p v-if="parseError[i]" class="hint bad">{{ parseError[i] }}</p>
      <p v-else-if="parsed[i]" class="hint">
        Registry <strong>{{ parsed[i]!.registry }}</strong>
        · Repository <strong>{{ parsed[i]!.repository }}</strong>
        · Tag <strong>{{ parsed[i]!.tag || parsed[i]!.digest }}</strong>
      </p>
      <div class="job-image-tag">
        <label :for="'job-target-' + i">保存为</label>
        <input
          v-model="img.targetTag"
          :disabled="disabled"
          :id="'job-target-' + i"
          placeholder="可选，例如 app-gateway:1.0.0 或 company/app:prod"
          spellcheck="false"
        />
      </div>
    </div>

    <div class="actions">
      <button class="btn ghost" :disabled="disabled" type="button" @click="addImage">添加镜像</button>
    </div>
    <div class="actions">
      <button class="btn primary" :disabled="disabled || saving || !canSave" type="button" @click="onSave(false)">
        {{ saving ? '保存中…' : '保存' }}
      </button>
      <button class="btn" :disabled="disabled || saving || !canSave" type="button" @click="onSave(true)">
        保存并下载
      </button>
      <button class="btn ghost" :disabled="saving" type="button" @click="emit('cancel')">取消</button>
    </div>
  </section>
</template>
