export type NamingConfig = { mode: string; customTemplate?: string }
export type NamingValidation = { valid: boolean; error?: string; example?: string }
export type BatchStartRequest = { inputs: string[]; outputRoot: string; maxWorkers: number; naming?: NamingConfig }
export type SelectedInput = { path: string; displayName: string; kind: string }
export type QueueItem = {
  index: number; sourcePath: string; sourceDisplayName: string; relativePath?: string; outputPath?: string
  family?: string; variant?: string; decodedFormat?: string; sourceBytes: number; sourceBytesKnown: boolean
  bytesWritten: number; status: string; errorCode?: string; errorDetail?: string; namingFallback?: boolean; namingFallbackReason?: string; retryableExternalKey?: boolean; retryableExternalDatabase?: boolean
}
export type AppInfo = { name: string; version: string; maxWorkers: number; supportsDrop: boolean; supportsPause: boolean }
export type BatchSnapshot = { successes: number; skippedNotEncrypted: number; skippedExisting: number; blockedExternalKey: number; blockedExternalDatabase: number; unsupported: number; unknown: number; failed: number; cancelled: number; bytesWritten: number; results: QueueItem[] }
export type BatchCompleteEvent = { batchId: number; cancelled: boolean; snapshot: BatchSnapshot; errorCode?: string; errorDetail?: string }
export type PreviewResponse = { available: boolean; url?: string; format?: string }
export type BatchEvent = { batchId: number; item: QueueItem; bytesWritten?: number; sourceBytes?: number; sourceBytesKnown?: boolean }
export type LaunchPathsEvent = { requestId: string; paths: string[]; source: 'INITIAL_PROCESS'|'SECOND_INSTANCE'; deferred: boolean }
export type Preferences = { maxWorkers: number; theme: 'system'|'light'|'dark'; autoSuggestOutput: boolean; namingMode: string; customNamingTemplate?: string; language?: 'SYSTEM'|'zh-CN'|'en-US' }
export const sortItems = (items: QueueItem[]) => [...items].sort((a,b) => a.index-b.index)
export const formatBytes = (bytes: number) => bytes < 1024 ? `${bytes} B` : bytes < 1024*1024 ? `${(bytes/1024).toFixed(1)} KB` : `${(bytes/1024/1024).toFixed(1)} MB`
export const normalizePath = (path: string) => path.replaceAll('/', '\\').replaceAll('\\\\', '\\').toLowerCase()
export const dedupePaths = (paths: string[]) => { const seen = new Set<string>(); return paths.filter(path => { const key=normalizePath(path); if(!path || seen.has(key)) return false; seen.add(key); return true }) }
export const classifyStatus = (status: string) => status === 'SUCCESS' ? 'completed' : ['SKIPPED_NOT_ENCRYPTED','SKIPPED_OUTPUT_EXISTS'].includes(status) ? 'skipped' : ['BLOCKED_EXTERNAL_KEY','BLOCKED_EXTERNAL_DATABASE'].includes(status) ? 'blocked' : ['FAILED','CANCELLED'].includes(status) ? 'problems' : 'other'

declare global { interface Window { go: any; runtime: any } }
export const backend = () => window.go?.main?.App
