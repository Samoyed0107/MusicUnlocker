import { useEffect, useRef } from 'react'
import type { AppInfo } from '../api'
import { useTranslation } from '../i18n'
import { Icon } from './Icon'

export function AboutDialog({ open, appInfo, onClose }: { open: boolean; appInfo: AppInfo; onClose: () => void }) {
  const { t } = useTranslation()
  const dialogRef = useRef<HTMLDialogElement>(null)

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    if (open && !dialog.open) dialog.showModal()
    if (!open && dialog.open) dialog.close()
  }, [open])

  return <dialog ref={dialogRef} className="about-dialog" aria-labelledby="about-title" onClose={onClose} onCancel={onClose} onMouseDown={event => { if (event.target === event.currentTarget) onClose() }}>
    <div className="dialog-heading">
      <div>
        <p className="eyebrow">{t('about.title')}</p>
        <h2 id="about-title">{appInfo.name}</h2>
      </div>
      <button className="icon-button" type="button" onClick={onClose} aria-label={t('about.close')}><Icon name="close" /></button>
    </div>
    <dl className="about-details">
      <div><dt>{t('about.version')}</dt><dd>{appInfo.version}</dd></div>
      <div><dt>{t('about.supportedFormats')}</dt><dd>{t('about.formatsValue')}</dd></div>
      <div><dt>{t('about.processing')}</dt><dd>{t('about.offline')}</dd></div>
      <div><dt>{t('about.build')}</dt><dd>{appInfo.buildIdentifier}</dd></div>
      <div><dt>{t('about.built')}</dt><dd>{appInfo.buildTimestamp}</dd></div>
    </dl>
  </dialog>
}
