import type { BatchCompleteEvent, QueueItem } from '../api'
import { useTranslation } from '../i18n'
import { AudioPreview } from './AudioPreview'
import { ProcessingTable } from './ProcessingTable'
import { Summary } from './Summary'

const filters = [['all', 'progress.all'], ['completed', 'progress.completed'], ['skipped', 'progress.skipped'], ['blocked', 'progress.blocked'], ['problems', 'progress.problems']] as const

export function ResultsWorkspace({
  items, filter, onFilterChange, onPlay, onRetry, retryingIndex, selectedPreview,
  complete, api, onClosePreview, onPreviewError, onOpenOutput,
}: {
  items: QueueItem[]
  filter: string
  onFilterChange: (filter: string) => void
  onPlay: (item: QueueItem) => void
  onRetry: (item: QueueItem) => void
  retryingIndex?: number
  selectedPreview?: QueueItem
  complete: BatchCompleteEvent | null
  api: any
  onClosePreview: () => void
  onPreviewError: (message: string) => void
  onOpenOutput: () => void
}) {
  const { t } = useTranslation()

  return <section className="card results-workspace" aria-labelledby="results-heading">
    <div className="section-title results-heading">
      <div>
        <h2 id="results-heading">{t('workflow.results')}</h2>
      </div>
      <div className="filters">
        {filters.map(([value, key]) => <button key={value} className={filter === value ? 'selected' : ''} aria-pressed={filter === value} onClick={() => onFilterChange(value)}>{t(key)}</button>)}
      </div>
    </div>
    <div className="table-head"><span>{t('progress.nameFormat')}</span><span>{t('progress.status')}</span><span>{t('progress.progress')}</span><span>{t('progress.outputDetail')}</span></div>
    <ProcessingTable items={items} filter={filter} onPlay={onPlay} onRetry={onRetry} retryingIndex={retryingIndex}/>
    <div className="results-summary">
      <div className="section-title">
        <div>
          <h2>{t('summary.title')}</h2>
        </div>
        {complete && <button className="text-button" onClick={onOpenOutput}>{t('output.open')}</button>}
      </div>
      <Summary summary={complete?.snapshot}/>
    </div>
    <AudioPreview item={selectedPreview} batchId={complete?.batchId} api={api} onClose={onClosePreview} onError={onPreviewError}/>
  </section>
}
