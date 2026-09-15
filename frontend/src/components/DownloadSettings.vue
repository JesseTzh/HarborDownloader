<script lang="ts" setup>
import OutputSelector from './OutputSelector.vue'

defineProps<{
  disabled?: boolean
  saving?: boolean
  saveMessage?: string
  saveOk?: boolean | null
}>()

const outputDir = defineModel<string>('outputDir', { required: true })

const emit = defineEmits<{
  select: []
  save: []
}>()
</script>

<template>
  <section class="card">
    <header class="card-head">
      <h2>下载目录</h2>
      <p class="hint">所有任务共用这个根目录。每个任务会在其下创建同名子文件夹；再次下载会先删除该文件夹里已有的 TAR 包。</p>
    </header>
    <OutputSelector
      v-model:output-dir="outputDir"
      :disabled="disabled"
      label="默认下载根目录"
      placeholder="选择默认下载根目录"
      @select="emit('select')"
    />
    <div class="actions">
      <button class="btn ghost" :disabled="disabled || saving || !outputDir.trim()" type="button" @click="emit('save')">
        {{ saving ? '正在保存…' : '保存下载目录' }}
      </button>
    </div>
    <pre v-if="saveMessage" class="test-msg" :class="{ ok: saveOk === true, bad: saveOk === false }">{{ saveMessage }}</pre>
  </section>
</template>
