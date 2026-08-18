export type BatchStartRequest = { inputs: string[]; outputRoot: string; maxWorkers: number }
export type QueueItem = {
  index: number; sourcePath: string; sourceDisplayName: string; relativePath?: string; outputPath?: string
  family?: string; variant?: string; decodedFormat?: string; sourceBytes: number; sourceBytesKnown: boolean
  bytesWritten: number; status: string; errorCode?: string; errorDetail?: string
}
export type AppInfo = { name: string; version: string; maxWorkers: number; supportsDrop: boolean; supportsPause: boolean }
export type BatchSnapshot = { successes: number; skippedNotEncrypted: number; skippedExisting: number; blockedExternalKey: number; blockedExternalDatabase: number; unsupported: number; unknown: number; failed: number; cancelled: number; bytesWritten: number; results: QueueItem[] }
export type BatchCompleteEvent = { batchId: number; cancelled: boolean; snapshot: BatchSnapshot; errorCode?: string; errorDetail?: string }
export type BatchEvent = { batchId: number; item: QueueItem; bytesWritten?: number; sourceBytes?: number; sourceBytesKnown?: boolean }
export const statusLabel = (status: string) => ({SUCCESS:'Completed', SKIPPED_NOT_ENCRYPTED:'Not encrypted', SKIPPED_OUTPUT_EXISTS:'Already exists', BLOCKED_EXTERNAL_KEY:'Key required', BLOCKED_EXTERNAL_DATABASE:'Database required', UNSUPPORTED:'Unsupported', UNKNOWN:'Unknown format', FAILED:'Failed', CANCELLED:'Cancelled'} as Record<string,string>)[status] ?? 'Waiting'
export const sortItems = (items: QueueItem[]) => [...items].sort((a,b) => a.index-b.index)
export const formatBytes = (bytes: number) => bytes < 1024 ? `${bytes} B` : bytes < 1024*1024 ? `${(bytes/1024).toFixed(1)} KB` : `${(bytes/1024/1024).toFixed(1)} MB`

declare global { interface Window { go: any; runtime: any } }
export const backend = () => window.go?.main?.App
