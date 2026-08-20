import { useTranslation } from '../i18n'

export function BatchCommandBar({
  statusText, running, cancelling, startDisabled, onStart, onCancel, notice, showAddHint, showOutputHint,
}: {
  statusText: string
  running: boolean
  cancelling: boolean
  startDisabled: boolean
  onStart: () => void
  onCancel: () => void
  notice: string
  showAddHint: boolean
  showOutputHint: boolean
}) {
  const { t } = useTranslation()

  return <section className="batch-command-surface" aria-label={t('workflow.actions')}>
    <div className="batch-command-main">
      <div className="batch-command-state">
        <p className="section-step">3</p>
        <div>
          <h2>{t('workflow.start')}</h2>
          <span className={`pill batch-status ${running ? 'status--running' : ''}`}>{statusText}</span>
        </div>
      </div>
      <div className="footer-actions">
        <button className="primary" onClick={onStart} disabled={startDisabled}>{t('batch.start')}</button>
        {running && <button className="danger" onClick={onCancel} disabled={cancelling}>{cancelling ? t('batch.cancelling') : t('batch.cancel')}</button>}
      </div>
    </div>
    {showAddHint && <p className="helper">{t('batch.addHint')}</p>}
    {showOutputHint && <p className="helper">{t('output.chooseBeforeStart')}</p>}
    {notice && <div className="notice" role="status">{notice}</div>}
  </section>
}
