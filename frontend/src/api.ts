export type BatchStartRequest = { inputs: string[]; outputRoot: string; maxWorkers: number }
export type SelectedInput = { path: string; displayName: string; kind: string }
export type QueueItem = {
  index: number; sourcePath: string; sourceDisplayName: string; relativePath?: string; outputPath?: string
  family?: string; variant?: string; decodedFormat?: string; sourceBytes: number; sourceBytesKnown: boolean
  bytesWritten: number; status: string; errorCode?: string; errorDetail?: string
}
export type AppInfo = { name: string; version: string; maxWorkers: number; supportsDrop: boolean; supportsPause: boolean }
export type BatchSnapshot = { successes: number; skippedNotEncrypted: number; skippedExisting: number; blockedExternalKey: number; blockedExternalDatabase: number; unsupported: number; unknown: number; failed: number; cancelled: number; bytesWritten: number; results: QueueItem[] }
export type BatchCompleteEvent = { batchId: number; cancelled: boolean; snapshot: BatchSnapshot; errorCode?: string; errorDetail?: string }
export type PreviewResponse = { available: boolean; url?: string; format?: string }
export type BatchEvent = { batchId: number; item: QueueItem; bytesWritten?: number; sourceBytes?: number; sourceBytesKnown?: boolean }
export type LaunchPathsEvent = { requestId: string; paths: string[]; source: 'INITIAL_PROCESS'|'SECOND_INSTANCE'; deferred: boolean }
export type Preferences = { maxWorkers: number; theme: 'system'|'light'|'dark'; autoSuggestOutput: boolean }
export const statusLabel = (status: string) => ({SUCCESS:'Completed', SKIPPED_NOT_ENCRYPTED:'Not encrypted', SKIPPED_OUTPUT_EXISTS:'Already exists', BLOCKED_EXTERNAL_KEY:'Key required', BLOCKED_EXTERNAL_DATABASE:'Database required', UNSUPPORTED:'Unsupported', UNKNOWN:'Unknown format', FAILED:'Failed', CANCELLED:'Cancelled'} as Record<string,string>)[status] ?? 'Waiting'
export const errorLabel = (code?: string) => ({QMC_EXTERNAL_KEY_REQUIRED:'This file requires an external QQ Music media key.', KGG_DATABASE_REQUIRED:'This Kugou format requires the Kugou local database.', DISCOVERY_ERROR:'The selected path could not be accessed.', OUTPUT_EXISTS:'A decoded file already exists.', DECODED_AUDIO_FORMAT_UNKNOWN:'The resulting audio format could not be identified.', PREVIEW_FILE_MISSING:'Output file is no longer available.', PREVIEW_SERVER_UNAVAILABLE:'Preview is unavailable in this application session.'} as Record<string,string>)[code ?? ''] ?? ''
export const sortItems = (items: QueueItem[]) => [...items].sort((a,b) => a.index-b.index)
export const formatBytes = (bytes: number) => bytes < 1024 ? `${bytes} B` : bytes < 1024*1024 ? `${(bytes/1024).toFixed(1)} KB` : `${(bytes/1024/1024).toFixed(1)} MB`
export const normalizePath = (path: string) => path.replaceAll('/', '\\').replaceAll('\\\\', '\\').toLowerCase()
export const dedupePaths = (paths: string[]) => { const seen = new Set<string>(); return paths.filter(path => { const key=normalizePath(path); if(!path || seen.has(key)) return false; seen.add(key); return true }) }
export const classifyStatus = (status: string) => status === 'SUCCESS' ? 'completed' : ['SKIPPED_NOT_ENCRYPTED','SKIPPED_OUTPUT_EXISTS'].includes(status) ? 'skipped' : ['BLOCKED_EXTERNAL_KEY','BLOCKED_EXTERNAL_DATABASE'].includes(status) ? 'blocked' : ['FAILED','CANCELLED'].includes(status) ? 'problems' : 'other'

declare global { interface Window { go: any; runtime: any } }
export const backend = () => window.go?.main?.App
