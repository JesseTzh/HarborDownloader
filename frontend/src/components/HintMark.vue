<script lang="ts" setup>
defineProps<{
  text: string
}>()
</script>

<template>
  <span class="hint-mark" tabindex="0" :aria-label="text" @click.prevent.stop>
    !
    <span class="hint-pop" role="tooltip">{{ text }}</span>
  </span>
</template>
