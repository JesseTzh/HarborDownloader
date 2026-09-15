export interface Platform {
  os: string
  architecture: string
  variant?: string
}

export interface RegistryConfig {
  registry: string
  username: string
  password: string
  insecure: boolean
}

export interface FileConfig {
  registry: string
  username?: string
  password?: string
  hasPassword?: boolean
  insecure?: boolean
  outputDir: string
  jobs?: Job[]
}

export interface JobImage {
  image: string
  targetTag?: string
  platform: Platform
}

export interface Job {
  id: string
  name: string
  pack?: boolean
  outputDir?: string
  images: JobImage[]
}

export interface JobItem {
  image: string
  targetTag?: string
  platform: string
  status: string
  outputPath?: string
  digest?: string
  size?: number
  error?: string
}

export interface JobStatusEvent {
  runId: string
  jobId: string
  name: string
  status: string
  current: number
  total: number
  items: JobItem[]
  packagePath?: string
  error?: string
}

export interface JobCompleteEvent {
  runId: string
  jobId: string
  name: string
  status: string
  items: JobItem[]
  packagePath?: string
  error?: string
}

export interface ImageReference {
  registry: string
  repository: string
  tag: string
  digest?: string
  raw: string
}

export interface DownloadRequest {
  image: string
  targetTag?: string
  outputDir: string
  platform: Platform
  registry: string
  username: string
  password: string
  insecure: boolean
}

export interface DownloadTask {
  id: string
  image: string
  outputPath: string
  platform: Platform
  status: string
  startedAt: string
  finishedAt?: string
  error?: string
  digest?: string
  size?: number
}

export interface ProgressEvent {
  taskId: string
  phase: string
  current: number
  total: number
  percentage: number
  speedBytes: number
  downloadedBytes: number
  etaSeconds: number
  currentLayer: number
  totalLayers: number
  cachedLayers: number
  missingLayers: number
  message: string
  runId?: string
  jobName?: string
  image?: string
  imageIndex?: number
  imageTotal?: number
}

export interface TestRegistryResult {
  ok: boolean
  registryReachable: boolean
  authSuccess: boolean
  harborReachable: boolean
  message: string
  error?: string
}

export interface DownloadCompleteEvent {
  taskId: string
  image: string
  platform: string
  digest: string
  outputPath: string
  size: number
}

export interface AppErrorEvent {
  taskId: string
  message: string
}

export interface Result {
  ok: boolean
  message: string
  error?: string
}

export interface LayerCacheInfo {
  path: string
  files: number
  bytes: number
  images: number
  unreferencedFiles: number
  unreferencedBytes: number
}

export interface CachedLayerImage {
  image: string
  platform: string
  digest?: string
}

export interface CachedLayer {
  digest: string
  size: number
  partial: boolean
  referenced: boolean
  images: CachedLayerImage[]
}

export interface CachedImage {
  image: string
  digest?: string
  platform: string
  updatedAt?: string
  layerCount: number
  cachedCount: number
}

export interface LayerCacheInventory {
  path: string
  files: number
  bytes: number
  images: number
  unreferencedFiles: number
  unreferencedBytes: number
  layers: CachedLayer[]
  imageRecords: CachedImage[]
}

export type UiState =
  | 'IDLE'
  | 'CHECKING'
  | 'DOWNLOADING'
  | 'EXPORTING'
  | 'PACKAGING'
  | 'SUCCESS'
  | 'FAILED'
  | 'CANCELLED'
