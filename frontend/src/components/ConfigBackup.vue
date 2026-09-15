<script lang="ts" setup>
defineProps<{
  disabled?: boolean
  exporting?: boolean
  importing?: boolean
  message?: string
  ok?: boolean | null
}>()

const emit = defineEmits<{
  export: []
  import: []
}>()
</script>

<template>
  <section class="card">
    <header class="card-head">
      <h2>导入导出</h2>
      <p class="hint">导出当前已保存的仓库设置和全部任务。导入会覆盖本机已保存的配置。导出文件使用程序内置密钥加密，密码不会以明文出现。</p>
    </header>
    <div class="actions">
      <button class="btn ghost" :disabled="disabled || exporting || importing" type="button" @click="emit('export')">
        {{ exporting ? '正在导出…' : '导出配置' }}
      </button>
      <button class="btn ghost" :disabled="disabled || exporting || importing" type="button" @click="emit('import')">
        {{ importing ? '正在导入…' : '导入配置' }}
      </button>
    </div>
    <pre v-if="message" class="test-msg" :class="{ ok: ok === true, bad: ok === false }">{{ message }}</pre>
  </section>
</template>
