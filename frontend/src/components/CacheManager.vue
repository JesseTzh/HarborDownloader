<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { CachedLayer, LayerCacheInventory } from '../types'

const props = defineProps<{
  inventory: LayerCacheInventory | null
  disabled?: boolean
  pruning?: boolean
  clearing?: boolean
}>()

const emit = defineEmits<{
  refresh: []
  prune: []
  clear: []
  open: []
  deleteLayer: [digest: string]
}>()

const filter = ref<'all' | 'unref' | 'ref'>('all')
const menu = ref<{ x: number; y: number; layer: CachedLayer } | null>(null)

const unused = computed(() => props.inventory?.unreferencedFiles ?? 0)
const files = computed(() => props.inventory?.files ?? 0)

const filteredLayers = computed(() => {
  const layers = props.inventory?.layers ?? []
  if (filter.value === 'unref') return layers.filter((l) => !l.referenced)
  if (filter.value === 'ref') return layers.filter((l) => l.referenced)
  return layers
})

function fmtBytes(n: number) {
  if (!n || n < 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let v = n
  let i = 0
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(i === 0 ? 0 : 2)} ${units[i]}`
}

function shortDigest(d: string) {
  if (!d) return '—'
  const hex = d.includes(':') ? d.slice(d.indexOf(':') + 1) : d
  if (hex.length <= 16) return d
  const prefix = d.includes(':') ? d.slice(0, d.indexOf(':') + 1) : ''
  return prefix + hex.slice(0, 12) + '…' + hex.slice(-6)
}

function imageLabel(image: string, platform: string) {
  return platform ? `${image} (${platform})` : image
}

function openMenu(ev: MouseEvent, layer: CachedLayer) {
  if (props.disabled) return
  menu.value = { x: ev.clientX, y: ev.clientY, layer }
}

function closeMenu() {
  menu.value = null
}

function onDeleteLayer() {
  const layer = menu.value?.layer
  closeMenu()
  if (!layer) return
  emit('deleteLayer', layer.digest)
}

function onWindowClick() {
  closeMenu()
}

onMounted(() => {
  window.addEventListener('click', onWindowClick)
  window.addEventListener('resize', closeMenu)
})
onUnmounted(() => {
  window.removeEventListener('click', onWindowClick)
  window.removeEventListener('resize', closeMenu)
})
</script>

<template>
  <section class="card cache-manager">
    <header class="card-head cache-head">
      <div>
        <h2>缓存管理</h2>
        <p class="hint cache-summary">
          <template v-if="files > 0">
            {{ files }} 个层 · {{ fmtBytes(inventory?.bytes ?? 0) }} · {{ inventory?.images ?? 0 }} 个镜像
            <template v-if="unused > 0"> · {{ unused }} 个未引用（{{ fmtBytes(inventory?.unreferencedBytes ?? 0) }}）</template>
            <template v-else> · 无未引用层</template>
          </template>
          <template v-else>缓存为空。下载成功后会在这里列出 layer 及其关联镜像。</template>
        </p>
        <p v-if="inventory?.path" class="hint cache-path" :title="inventory.path">{{ inventory.path }}</p>
      </div>
    </header>

    <div class="row cache-actions">
      <button class="btn ghost" :disabled="disabled" type="button" @click="emit('refresh')">刷新</button>
      <button class="btn ghost" :disabled="disabled" type="button" @click="emit('open')">打开目录</button>
      <button class="btn ghost" :disabled="disabled || pruning || unused <= 0" type="button" @click="emit('prune')">
        {{ pruning ? '正在清理…' : '清理未引用层' }}
      </button>
      <button class="btn danger" :disabled="disabled || clearing || files <= 0" type="button" @click="emit('clear')">
        {{ clearing ? '正在清空…' : '清空全部' }}
      </button>
    </div>
    <p v-if="disabled" class="hint">下载进行中，缓存只读。</p>
  </section>

  <section v-if="(inventory?.imageRecords?.length ?? 0) > 0" class="card">
    <header class="card-head">
      <h2>已记录镜像</h2>
    </header>
    <div class="table-wrap">
      <table class="cache-table">
        <thead>
          <tr>
            <th>镜像</th>
            <th>平台</th>
            <th>层</th>
            <th>更新时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="img in inventory?.imageRecords" :key="img.image + img.platform">
            <td>
              <div class="mono wrap">{{ img.image }}</div>
              <div v-if="img.digest" class="hint digest">{{ shortDigest(img.digest) }}</div>
            </td>
            <td>{{ img.platform || '—' }}</td>
            <td>{{ img.cachedCount }} / {{ img.layerCount }}</td>
            <td class="muted">{{ img.updatedAt || '—' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>

  <section class="card">
    <header class="card-head cache-layers-head">
      <h2>Layer 文件</h2>
      <div class="filters">
        <button class="chip" :class="{ on: filter === 'all' }" type="button" @click="filter = 'all'">全部</button>
        <button class="chip" :class="{ on: filter === 'unref' }" type="button" @click="filter = 'unref'">未引用</button>
        <button class="chip" :class="{ on: filter === 'ref' }" type="button" @click="filter = 'ref'">有引用</button>
      </div>
    </header>
    <p class="hint">右键某一层可删除。删除后，关联镜像下次下载需重新拉取该层。</p>
    <div v-if="filteredLayers.length === 0" class="empty">没有可显示的层。</div>
    <div v-else class="table-wrap">
      <table class="cache-table">
        <thead>
          <tr>
            <th>Digest</th>
            <th>大小</th>
            <th>状态</th>
            <th>关联镜像</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="layer in filteredLayers"
            :key="layer.digest"
            class="layer-row"
            @contextmenu.prevent.stop="openMenu($event, layer)"
          >
            <td>
              <div class="mono" :title="layer.digest">{{ shortDigest(layer.digest) }}</div>
            </td>
            <td>{{ fmtBytes(layer.size) }}</td>
            <td>
              <span v-if="layer.partial" class="tag warn">不完整</span>
              <span v-else-if="layer.referenced" class="tag ok">已引用</span>
              <span v-else class="tag muted">未引用</span>
            </td>
            <td>
              <ul v-if="layer.images?.length" class="image-list">
                <li v-for="img in layer.images" :key="img.image + img.platform">
                  {{ imageLabel(img.image, img.platform) }}
                </li>
              </ul>
              <span v-else class="muted">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>

  <div
    v-if="menu"
    class="ctx-menu"
    :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
    @click.stop
  >
    <button type="button" class="ctx-item danger" :disabled="disabled" @click="onDeleteLayer">删除此层</button>
  </div>
</template>
