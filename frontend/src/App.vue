<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import RegistryForm from './components/RegistryForm.vue'
import DownloadSettings from './components/DownloadSettings.vue'
import ConfigBackup from './components/ConfigBackup.vue'
import CacheManager from './components/CacheManager.vue'
import JobList from './components/JobList.vue'
import JobEditor from './components/JobEditor.vue'
import JobProgress from './components/JobProgress.vue'
import JobResult from './components/JobResult.vue'
import LogViewer from './components/LogViewer.vue'
import type {
  CachedLayer,
  ImageReference,
  Job,
  JobCompleteEvent,
  JobStatusEvent,
  LayerCacheInventory,
  Platform,
  ProgressEvent,
  UiState,
} from './types'
import {
  CancelDownload,
  ClearLayerCache,
  DeleteCachedLayer,
  DeleteJob,
  ExportConfig,
  GetDefaultOutputDir,
  GetLayerCacheInventory,
  GetVersion,
  ImportConfig,
  ListJobs,
  ListPlatforms,
  LoadConfig,
  OpenLayerCacheDir,
  OpenOutputDir,
  ParseImage,
  PruneLayerCache,
  SaveConfig,
  SaveJob,
  SelectOutputDir,
  StartJob,
  TestRegistry,
} from '../wailsjs/go/app/App'
import { EventsOff, EventsOn } from '../wailsjs/runtime/runtime'

type Tab = 'tasks' | 'settings' | 'cache'
type TaskView = 'list' | 'editor'

function fallbackVersion() {
  const d = new Date()
  const yy = String(d.getFullYear()).slice(-2)
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${yy}${mm}${dd}-nightly.0`
}

const version = ref(fallbackVersion())
const tab = ref<Tab>('tasks')
const taskView = ref<TaskView>('list')
const ui = ref<UiState>('IDLE')
const registry = ref('')
const username = ref('')
const password = ref('')
const hasPassword = ref(false)
const insecure = ref(false)
const outputDir = ref('')
const platforms = ref<Platform[]>([
  { os: 'linux', architecture: 'amd64' },
  { os: 'linux', architecture: 'arm64' },
])
const testing = ref(false)
const testMessage = ref('')
const testOk = ref<boolean | null>(null)
const progress = ref<ProgressEvent | null>(null)
const errorText = ref('')
const cancelling = ref(false)
const logs = ref<string[]>([])
const inventory = ref<LayerCacheInventory | null>(null)
const clearingCache = ref(false)
const pruningCache = ref(false)
const savingRegistry = ref(false)
const saveMessage = ref('')
const saveOk = ref<boolean | null>(null)
const savedSnapshot = ref('')
const exportingConfig = ref(false)
const importingConfig = ref(false)
const configIOMessage = ref('')
const configIOOk = ref<boolean | null>(null)
const jobs = ref<Job[]>([])
const editing = ref<Job | null>(null)
const editorRef = ref<{
  setParseResult: (i: number, v: ImageReference | null, e: string) => void
} | null>(null)
const savingJob = ref(false)
const runId = ref('')
const jobRun = ref<JobStatusEvent | null>(null)
const jobResult = ref<JobCompleteEvent | null>(null)

function registrySnapshot() {
  return JSON.stringify({
    registry: registry.value.trim(),
    username: username.value,
    password: password.value,
    insecure: insecure.value,
    outputDir: outputDir.value.trim(),
  })
}

function markSaved() {
  savedSnapshot.value = registrySnapshot()
}

function applyPasswordState(saved: boolean, typed = '') {
  hasPassword.value = saved || !!typed
  password.value = ''
}

const registryDirty = computed(() => registrySnapshot() !== savedSnapshot.value)

const defaultPlatform = computed(
  () => platforms.value[0] || { os: 'linux', architecture: 'amd64' }
)

function blankJob(): Job {
  return {
    id: '',
    name: '',
    pack: false,
    images: [{ image: '', platform: { ...defaultPlatform.value } }],
  }
}

async function refreshCache() {
  try {
    inventory.value = await GetLayerCacheInventory()
  } catch {
    /* ignore */
  }
}

async function refreshJobs() {
  try {
    jobs.value = (await ListJobs()) || []
  } catch {
    jobs.value = []
  }
}

async function onTab(next: Tab) {
  if (tab.value === 'settings' && next !== 'settings' && registryDirty.value) {
    await persistRegistry(false)
  }
  tab.value = next
  if (next === 'cache') refreshCache()
  if (next === 'tasks') refreshJobs()
}

const busy = computed(() =>
  ['CHECKING', 'DOWNLOADING', 'EXPORTING', 'PACKAGING'].includes(ui.value)
)

function log(line: string) {
  const ts = new Date().toLocaleTimeString()
  logs.value = [...logs.value.slice(-80), `${ts}  ${line}`]
}

function mapStatus(status: string): UiState {
  switch (status) {
    case 'checking':
      return 'CHECKING'
    case 'downloading':
      return 'DOWNLOADING'
    case 'exporting':
      return 'EXPORTING'
    case 'packaging':
      return 'PACKAGING'
    case 'pending':
      return 'DOWNLOADING'
    case 'completed':
      return 'SUCCESS'
    case 'failed':
      return 'FAILED'
    case 'cancelled':
      return 'CANCELLED'
    default:
      return ui.value
  }
}

onMounted(async () => {
  EventsOn('download-progress', (ev: ProgressEvent) => {
    progress.value = ev
    if (ev.phase === 'checking') ui.value = 'CHECKING'
    if (ev.phase === 'downloading') ui.value = 'DOWNLOADING'
    if (ev.phase === 'exporting') ui.value = 'EXPORTING'
    if (ev.phase === 'packaging') ui.value = 'PACKAGING'
  })
  EventsOn('download-status', (task: { status: string; error?: string }) => {
    if (!jobRun.value) ui.value = mapStatus(task.status)
    if (task.error) errorText.value = task.error
    log(`状态 ${task.status}`)
  })
  EventsOn('download-error', (ev: { taskId: string; message: string }) => {
    errorText.value = ev.message
    log(ev.message.replace(/\n/g, ' '))
  })
  EventsOn('download-completed', (ev: { outputPath: string }) => {
    log('镜像完成 ' + ev.outputPath)
    refreshCache()
  })
  EventsOn('job-status', (ev: JobStatusEvent) => {
    jobRun.value = ev
    if (ev.status === 'pending' || ev.status === 'downloading') ui.value = 'DOWNLOADING'
    else if (ev.status === 'packaging') ui.value = 'PACKAGING'
    else ui.value = mapStatus(ev.status)
    log(`任务 ${ev.name} ${ev.current}/${ev.total} ${ev.status}`)
  })
  EventsOn('job-completed', (ev: JobCompleteEvent) => {
    jobResult.value = ev
    jobRun.value = {
      runId: ev.runId,
      jobId: ev.jobId,
      name: ev.name,
      status: ev.status,
      current: ev.items?.length || 0,
      total: ev.items?.length || 0,
      items: ev.items,
      error: ev.error,
    }
    ui.value = mapStatus(ev.status)
    log('任务结束 ' + ev.name + ' ' + ev.status)
    refreshCache()
  })

  try {
    version.value = await GetVersion()
  } catch {
    /* ignore */
  }
  try {
    const cfg = await LoadConfig()
    registry.value = cfg.registry || ''
    username.value = cfg.username || ''
    applyPasswordState(!!cfg.hasPassword || !!cfg.password, cfg.password || '')
    insecure.value = !!cfg.insecure
    outputDir.value = cfg.outputDir || (await GetDefaultOutputDir())
    markSaved()
    if (!registry.value.trim()) tab.value = 'settings'
  } catch {
    try {
      outputDir.value = await GetDefaultOutputDir()
    } catch {
      /* ignore */
    }
  }
  await refreshCache()
  await refreshJobs()
  try {
    const list = await ListPlatforms()
    if (list?.length) {
      platforms.value = list
    }
  } catch {
    /* keep defaults */
  }
})

onUnmounted(() => {
  EventsOff('download-progress')
  EventsOff('download-status')
  EventsOff('download-error')
  EventsOff('download-completed')
  EventsOff('job-status')
  EventsOff('job-completed')
})

async function persistRegistry(showMessage: boolean) {
  savingRegistry.value = true
  if (showMessage) {
    saveOk.value = null
    saveMessage.value = ''
  }
  try {
    const res = await SaveConfig({
      registry: registry.value.trim(),
      username: username.value,
      password: password.value,
      insecure: insecure.value,
      outputDir: outputDir.value.trim(),
    } as any)
    if (res.ok) {
      applyPasswordState(hasPassword.value, password.value)
      markSaved()
      if (showMessage) {
        saveOk.value = true
        saveMessage.value = res.message || '已保存'
      }
      log('仓库配置已保存')
    } else {
      if (showMessage) {
        saveOk.value = false
        saveMessage.value = res.error || res.message || '保存失败'
      }
      log(res.error || res.message || '保存仓库配置失败')
    }
  } catch (e) {
    if (showMessage) {
      saveOk.value = false
      saveMessage.value = String(e)
    }
    log(String(e))
  } finally {
    savingRegistry.value = false
  }
}

async function onSaveRegistry() {
  await persistRegistry(true)
}

async function applyLoadedConfig() {
  try {
    const cfg = await LoadConfig()
    registry.value = cfg.registry || ''
    username.value = cfg.username || ''
    applyPasswordState(!!cfg.hasPassword || !!cfg.password, cfg.password || '')
    insecure.value = !!cfg.insecure
    outputDir.value = cfg.outputDir || (await GetDefaultOutputDir())
    markSaved()
  } catch {
    /* ignore */
  }
  await refreshJobs()
}

async function onExportConfig() {
  if (registryDirty.value && registry.value.trim()) {
    await persistRegistry(false)
  }
  exportingConfig.value = true
  configIOOk.value = null
  configIOMessage.value = ''
  try {
    const res = await ExportConfig()
    if (!res.ok) {
      configIOOk.value = false
      configIOMessage.value = res.error || res.message || '导出失败'
      log(configIOMessage.value)
      return
    }
    if (!res.message) return
    configIOOk.value = true
    configIOMessage.value = res.message
    log(res.message)
  } catch (e) {
    configIOOk.value = false
    configIOMessage.value = String(e)
    log(String(e))
  } finally {
    exportingConfig.value = false
  }
}

async function onImportConfig() {
  const ok = window.confirm('导入将覆盖当前已保存的仓库设置和全部任务。是否继续？')
  if (!ok) return
  importingConfig.value = true
  configIOOk.value = null
  configIOMessage.value = ''
  try {
    const res = await ImportConfig()
    if (!res.ok) {
      configIOOk.value = false
      configIOMessage.value = res.error || res.message || '导入失败'
      log(configIOMessage.value)
      return
    }
    if (!res.message) return
    await applyLoadedConfig()
    configIOOk.value = true
    configIOMessage.value = res.message
    log(res.message)
  } catch (e) {
    configIOOk.value = false
    configIOMessage.value = String(e)
    log(String(e))
  } finally {
    importingConfig.value = false
  }
}

async function onTest() {
  testing.value = true
  testOk.value = null
  testMessage.value = ''
  try {
    const res = await TestRegistry({
      registry: registry.value.trim(),
      username: username.value,
      password: password.value,
      insecure: insecure.value,
    } as any)
    testOk.value = res.ok
    testMessage.value = res.ok ? res.message : res.error || res.message
    log(res.ok ? '连接测试成功' : '连接测试失败')
    if (res.ok) {
      applyPasswordState(hasPassword.value, password.value)
      markSaved()
    }
  } catch (e) {
    testOk.value = false
    testMessage.value = String(e)
    log('连接测试失败')
  } finally {
    testing.value = false
  }
}

async function onSelectDir() {
  try {
    const dir = await SelectOutputDir()
    if (dir) {
      outputDir.value = dir
    }
  } catch (e) {
    log(String(e))
  }
}

function onCreateJob() {
  editing.value = blankJob()
  taskView.value = 'editor'
}

function onEditJob(job: Job) {
  editing.value = {
    ...job,
    images: job.images.map((img) => ({
      image: img.image,
      targetTag: img.targetTag || '',
      platform: { ...img.platform },
    })),
  }
  taskView.value = 'editor'
}

function onCancelEdit() {
  editing.value = null
  taskView.value = 'list'
}

async function onSaveJob(job: Job, start = false) {
  savingJob.value = true
  try {
    const saved = await SaveJob(job as any)
    log('已保存任务 ' + saved.name)
    await refreshJobs()
    editing.value = null
    taskView.value = 'list'
    if (start) await onStartJob(saved)
  } catch (e) {
    log(String(e))
    errorText.value = String(e)
  } finally {
    savingJob.value = false
  }
}

async function onDeleteJob(job: Job) {
  if (!window.confirm(`删除任务「${job.name}」？`)) return
  try {
    const res = await DeleteJob(job.id)
    if (res.ok) {
      log('已删除任务 ' + job.name)
      await refreshJobs()
    } else {
      log(res.error || res.message || '删除失败')
    }
  } catch (e) {
    log(String(e))
  }
}

async function onStartJob(job: Job) {
  if (!registry.value.trim()) {
    tab.value = 'settings'
    log('请先配置仓库')
    return
  }
  errorText.value = ''
  progress.value = null
  jobResult.value = null
  cancelling.value = false
  ui.value = 'DOWNLOADING'
  taskView.value = 'list'
  editing.value = null
  jobRun.value = {
    runId: '',
    jobId: job.id,
    name: job.name,
    status: 'downloading',
    current: 0,
    total: (job.images || []).length,
    items: (job.images || []).map((img) => ({
      image: img.image,
      targetTag: img.targetTag,
      platform: img.platform.variant
        ? `${img.platform.os}/${img.platform.architecture}/${img.platform.variant}`
        : `${img.platform.os}/${img.platform.architecture}`,
      status: 'pending',
    })),
  }
  try {
    const id = await StartJob(job.id)
    runId.value = id
    log('任务已开始 ' + job.name)
  } catch (e) {
    ui.value = 'FAILED'
    errorText.value = String(e)
    jobResult.value = {
      runId: '',
      jobId: job.id,
      name: job.name,
      status: 'failed',
      items: (job.images || []).map((img) => ({
        image: img.image,
        targetTag: img.targetTag,
        platform: `${img.platform.os}/${img.platform.architecture}`,
        status: 'failed',
        error: String(e),
      })),
      error: String(e),
    }
  }
}

async function onParseJobImage(image: string, index: number) {
  try {
    const parsed = await ParseImage(image, registry.value.trim(), insecure.value)
    editorRef.value?.setParseResult(index, parsed, '')
  } catch (e) {
    editorRef.value?.setParseResult(index, null, String(e))
  }
}

async function onCancel() {
  if (!runId.value) return
  cancelling.value = true
  try {
    await CancelDownload(runId.value)
  } catch (e) {
    errorText.value = String(e)
    cancelling.value = false
  }
}

async function onPruneCache() {
  const unused = inventory.value?.unreferencedFiles ?? 0
  const ok = window.confirm(
    unused > 0
      ? `删除 ${unused} 个未关联任何镜像的 layer？已被镜像引用的层会保留。`
      : '没有未引用层。'
  )
  if (!ok || unused <= 0) return
  pruningCache.value = true
  try {
    const res = await PruneLayerCache()
    if (res.ok) {
      log(res.message || '未引用层已清理')
    } else {
      log(res.error || res.message || '清理未引用层失败')
    }
  } catch (e) {
    log(String(e))
  } finally {
    pruningCache.value = false
    await refreshCache()
  }
}

async function onClearCache() {
  const files = inventory.value?.files ?? 0
  const ok = window.confirm(
    files > 0
      ? `清空全部 layer 缓存和镜像索引（${files} 个文件）？下次下载相同层需要重新从 Harbor 拉取。`
      : '缓存已经是空的。'
  )
  if (!ok || files <= 0) return
  clearingCache.value = true
  try {
    const res = await ClearLayerCache()
    if (res.ok) {
      log(res.message || '缓存已清空')
    } else {
      log(res.error || res.message || '清空缓存失败')
    }
  } catch (e) {
    log(String(e))
  } finally {
    clearingCache.value = false
    await refreshCache()
  }
}

async function onDeleteLayer(digest: string) {
  const layer = (inventory.value?.layers ?? []).find((l: CachedLayer) => l.digest === digest)
  const names = (layer?.images ?? []).map((img) => (img.platform ? `${img.image} (${img.platform})` : img.image))
  const msg =
    names.length > 0
      ? `删除层 ${digest.slice(0, 19)}… ？\n它正被以下镜像引用：\n${names.join('\n')}\n删除后这些镜像下次需重新下载该层。`
      : `删除未引用层 ${digest.slice(0, 19)}… ？`
  if (!window.confirm(msg)) return
  try {
    const res = await DeleteCachedLayer(digest)
    if (res.ok) {
      log(res.message || '已删除该层')
    } else {
      log(res.error || res.message || '删除失败')
    }
  } catch (e) {
    log(String(e))
  } finally {
    await refreshCache()
  }
}

async function onOpenCache() {
  try {
    const res = await OpenLayerCacheDir()
    if (!res.ok) {
      log(res.error || res.message || '无法打开缓存目录')
    }
  } catch (e) {
    log(String(e))
  }
}

async function onOpenPath(path: string) {
  if (!path) return
  await OpenOutputDir(path)
}

async function onCopyPath(path: string) {
  if (!path) return
  try {
    await navigator.clipboard.writeText(path)
    log('已复制路径')
  } catch {
    log('复制失败')
  }
}

function onDone() {
  ui.value = 'IDLE'
  progress.value = null
  errorText.value = ''
  runId.value = ''
  jobRun.value = null
  jobResult.value = null
  cancelling.value = false
  taskView.value = 'list'
}

const showForm = computed(() => ui.value === 'IDLE')
const showProgress = computed(() => busy.value)
const showResult = computed(() => ui.value === 'SUCCESS' || ui.value === 'FAILED' || ui.value === 'CANCELLED')
</script>

<template>
  <div class="app">
    <header class="top">
      <div>
        <h1>Harbor Downloader</h1>
        <p>从 Harbor 私有仓库下载镜像并导出 Docker TAR，无需 Docker Engine。</p>
        <nav class="tabs" aria-label="主导航">
          <button class="tab" :class="{ on: tab === 'tasks' }" type="button" @click="onTab('tasks')">任务</button>
          <button class="tab" :class="{ on: tab === 'settings' }" type="button" @click="onTab('settings')">设置</button>
          <button class="tab" :class="{ on: tab === 'cache' }" type="button" @click="onTab('cache')">缓存</button>
        </nav>
      </div>
      <span class="ver">Version {{ version }}</span>
    </header>

    <template v-if="tab === 'tasks'">
      <JobList
        v-if="showForm && taskView === 'list'"
        :jobs="jobs"
        :download-root="outputDir"
        :disabled="busy"
        @create="onCreateJob"
        @edit="onEditJob"
        @remove="onDeleteJob"
        @start="onStartJob"
      />

      <JobEditor
        v-if="showForm && taskView === 'editor' && editing"
        ref="editorRef"
        :job="editing"
        :platforms="platforms"
        :disabled="busy"
        :saving="savingJob"
        @save="onSaveJob"
        @cancel="onCancelEdit"
        @parse="onParseJobImage"
      />

      <JobProgress
        v-if="showProgress && jobRun"
        :name="jobRun.name"
        :status="jobRun.status"
        :current="jobRun.current"
        :total="jobRun.total"
        :items="jobRun.items"
        :progress="progress"
        :cancelling="cancelling"
        @cancel="onCancel"
      />

      <JobProgress
        v-else-if="showProgress"
        name="下载任务"
        :current="progress?.imageIndex || 0"
        :total="progress?.imageTotal || 0"
        :items="[]"
        :progress="progress"
        :cancelling="cancelling"
        @cancel="onCancel"
      />

      <JobResult
        v-if="showResult && jobResult"
        :result="jobResult"
        @open="onOpenPath"
        @copy="onCopyPath"
        @done="onDone"
      />
    </template>

    <template v-else-if="tab === 'settings'">
      <RegistryForm
        v-model:registry="registry"
        v-model:username="username"
        v-model:password="password"
        v-model:insecure="insecure"
        :disabled="busy"
        :testing="testing"
        :saving="savingRegistry"
        :has-password="hasPassword"
        :test-message="testMessage"
        :test-ok="testOk"
        :save-message="saveMessage"
        :save-ok="saveOk"
        @test="onTest"
        @save="onSaveRegistry"
      />
      <DownloadSettings
        v-model:output-dir="outputDir"
        :disabled="busy"
        :saving="savingRegistry"
        :save-message="saveMessage"
        :save-ok="saveOk"
        @select="onSelectDir"
        @save="onSaveRegistry"
      />
      <ConfigBackup
        :disabled="busy"
        :exporting="exportingConfig"
        :importing="importingConfig"
        :message="configIOMessage"
        :ok="configIOOk"
        @export="onExportConfig"
        @import="onImportConfig"
      />
    </template>

    <CacheManager
      v-else-if="tab === 'cache'"
      :inventory="inventory"
      :disabled="busy"
      :clearing="clearingCache"
      :pruning="pruningCache"
      @refresh="refreshCache"
      @prune="onPruneCache"
      @clear="onClearCache"
      @open="onOpenCache"
      @delete-layer="onDeleteLayer"
    />

    <LogViewer :lines="logs" />
  </div>
</template>
