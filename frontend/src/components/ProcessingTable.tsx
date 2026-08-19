import { useTranslation } from '../i18n'
import { classifyStatus, QueueItem } from '../api'

const errorKeys: Record<string, any> = {
  QMC_EXTERNAL_KEY_REQUIRED: 'errors.qmcKey',
  QMC_EXTERNAL_KEY_ATTEMPT_FAILED: 'errors.qmcKeyAttemptFailed',
  KGG_DATABASE_REQUIRED: 'errors.kggDatabase', DISCOVERY_ERROR: 'errors.discovery',
  OUTPUT_EXISTS: 'errors.outputExists', DECODED_AUDIO_FORMAT_UNKNOWN: 'errors.unknownAudio',
  PREVIEW_FILE_MISSING: 'errors.previewMissing', PREVIEW_SERVER_UNAVAILABLE: 'errors.previewUnavailable',
}
const statusKeys: Record<string, any> = {
  SUCCESS:'status.completed', SKIPPED_NOT_ENCRYPTED:'status.notEncrypted', SKIPPED_OUTPUT_EXISTS:'status.outputExists',
  BLOCKED_EXTERNAL_KEY:'status.keyRequired', BLOCKED_EXTERNAL_DATABASE:'status.databaseRequired', UNSUPPORTED:'status.unsupported',
  UNKNOWN:'status.unknown', FAILED:'status.failed', CANCELLED:'status.cancelled',
}

export function ProcessingTable({items, filter, onPlay, onRetry, retryingIndex}: {
  items: QueueItem[]; filter: string; onPlay: (item: QueueItem) => void
  onRetry: (item: QueueItem) => void; retryingIndex?: number
}) {
  const {t}=useTranslation()
  const visible=items.filter(item=>filter==='all'||classifyStatus(item.status)===filter)
  return <div className="table-wrap">{visible.length===0?<div className="empty-table">{items.length?t('progress.emptyFilter'):t('progress.empty')}</div>:visible.map(item=>{
    const retryable=(item.status==='BLOCKED_EXTERNAL_KEY' && item.errorCode==='QMC_EXTERNAL_KEY_REQUIRED') || (item.status==='FAILED' && (item.retryableExternalKey || item.errorCode==='QMC_EXTERNAL_KEY_ATTEMPT_FAILED'))
    const retrying=retryingIndex===item.index
    return <div className="table-row" key={item.index}>
      <div className="file-cell"><strong title={item.sourcePath}>{item.sourceDisplayName}</strong><small>{item.family||t('progress.unclassified')}{item.variant?` · ${item.variant}`:''}</small></div>
      <div><span className={`status-badge status-${item.status.toLowerCase()}`} title={item.errorCode}>{t(statusKeys[item.status]||'status.waiting')}</span></div>
      <div className="progress-cell"><span>{formatBytes(item.bytesWritten)} {t('progress.written')}</span>{item.status==='SUCCESS'&&<div className="progress-bar complete"><i/></div>}{item.status!=='SUCCESS'&&item.sourceBytesKnown&&item.sourceBytes>0&&<div className="progress-bar"><i style={{width:`${Math.min(100,item.bytesWritten/item.sourceBytes*100)}%`}}/></div>}</div>
      <div className="detail-cell">{item.errorCode&&<details><summary>{t('progress.details')}</summary><p>{errorKeys[item.errorCode]?t(errorKeys[item.errorCode]):item.errorDetail}</p></details>}{item.outputPath&&item.status==='SUCCESS'&&<><small title={item.outputPath}>{item.outputPath}</small><button className="play-button" onClick={()=>onPlay(item)} aria-label={t('progress.play')}>{t('progress.play')}</button></>}{retryable&&<button className="retry-button" onClick={()=>onRetry(item)} disabled={retryingIndex!==undefined} aria-label={retrying?t('retry.retrying'):item.status==='FAILED'?t('retry.another'):t('retry.provide')}>{retrying?t('retry.retrying'):item.status==='FAILED'?t('retry.another'):t('retry.provide')}</button>}</div>
    </div>
  })}</div>
}

const formatBytes=(bytes:number)=>bytes<1024?`${bytes} B`:bytes<1024*1024?`${(bytes/1024).toFixed(1)} KB`:`${(bytes/1024/1024).toFixed(1)} MB`
