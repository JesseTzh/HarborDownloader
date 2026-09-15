<script lang="ts" setup>
import { computed } from 'vue'

const props = defineProps<{
  disabled?: boolean
  testing?: boolean
  saving?: boolean
  hasPassword?: boolean
  testMessage?: string
  testOk?: boolean | null
  saveMessage?: string
  saveOk?: boolean | null
}>()

const registry = defineModel<string>('registry', { required: true })
const username = defineModel<string>('username', { required: true })
const password = defineModel<string>('password', { required: true })
const insecure = defineModel<boolean>('insecure', { required: true })

const tlsVerify = computed({
  get: () => !insecure.value,
  set: (v: boolean) => {
    insecure.value = !v
  },
})

const emit = defineEmits<{
  test: []
  save: []
}>()

const passwordPlaceholder = computed(() =>
  props.hasPassword && !password.value ? '已保存，重新输入可修改' : '密码或 Robot Secret'
)
</script>

<template>
  <section class="card">
    <header class="card-head">
      <h2>仓库配置</h2>
      <p class="hint">地址、账号和 TLS 设置会保存到本机，下次启动自动填充。密码保存后无法再次查看，日志也不会记录密码。</p>
    </header>
    <div class="field">
      <label for="registry">Harbor 地址</label>
      <input
        id="registry"
        v-model="registry"
        :disabled="disabled"
        autocomplete="off"
        placeholder="harbor.company.local"
        spellcheck="false"
      />
    </div>
    <div class="field">
      <label for="username">Username</label>
      <input
        id="username"
        v-model="username"
        :disabled="disabled"
        autocomplete="username"
        placeholder="robot$project+download"
        spellcheck="false"
      />
    </div>
    <div class="field">
      <label for="password">Password</label>
      <input
        id="password"
        v-model="password"
        :disabled="disabled"
        autocomplete="new-password"
        :placeholder="passwordPlaceholder"
        type="password"
      />
    </div>
    <label class="check">
      <input v-model="tlsVerify" :disabled="disabled" type="checkbox" />
      校验 TLS 证书
    </label>
    <p class="hint">关闭后允许 HTTP 或自签名证书。</p>
    <div class="actions">
      <button class="btn ghost" :disabled="disabled || saving || !registry" type="button" @click="emit('save')">
        {{ saving ? '正在保存…' : '保存配置' }}
      </button>
      <button class="btn" :disabled="disabled || testing || !registry" type="button" @click="emit('test')">
        {{ testing ? '正在测试…' : '测试连接' }}
      </button>
    </div>
    <pre v-if="saveMessage" class="test-msg" :class="{ ok: saveOk === true, bad: saveOk === false }">{{ saveMessage }}</pre>
    <pre v-if="testMessage" class="test-msg" :class="{ ok: testOk === true, bad: testOk === false }">{{ testMessage }}</pre>
  </section>
</template>
