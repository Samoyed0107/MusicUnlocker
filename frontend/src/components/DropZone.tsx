import type { ReactNode } from 'react'
import { useTranslation } from '../i18n'

export function DropZone({disabled, onDrop, children}: {disabled: boolean; onDrop: (paths: string[])=>void; children: ReactNode}) {
  void onDrop
  const {t}=useTranslation(); return <div className={`drop-zone ${disabled?'drop-disabled':''}`} style={{'--wails-drop-target':'drop'} as any} aria-disabled={disabled}>
    <div className="drop-copy"><strong>{t('inputs.drop')}</strong><span>{t('inputs.dropHint')}</span></div>
    {children}
  </div>
}
