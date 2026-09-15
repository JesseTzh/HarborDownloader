<script lang="ts" setup>
withDefaults(
  defineProps<{
    disabled?: boolean
    label?: string
    placeholder?: string
  }>(),
  {
    label: 'Output',
    placeholder: '选择 TAR 输出目录',
  }
)

const outputDir = defineModel<string>('outputDir', { required: true })

const emit = defineEmits<{
  select: []
}>()
</script>

<template>
  <div class="field">
    <label for="output">{{ label }}</label>
    <div class="row">
      <input id="output" v-model="outputDir" :disabled="disabled" :placeholder="placeholder" spellcheck="false" />
      <button class="btn ghost" :disabled="disabled" type="button" @click="emit('select')">选择</button>
    </div>
  </div>
</template>
