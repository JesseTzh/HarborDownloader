<script lang="ts" setup>
defineProps<{
  lines: string[]
}>()
</script>

<template>
  <section v-if="lines.length" class="card log-card">
    <header class="card-head">
      <h2>日志</h2>
    </header>
    <pre class="log">{{ lines.join('\n') }}</pre>
  </section>
</template>
