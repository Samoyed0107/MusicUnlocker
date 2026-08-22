import type { ReactNode } from 'react'
import { useTranslation } from '../i18n'
import { Icon } from './Icon'

export function DropZone({disabled, onDrop, children}: {disabled: boolean; onDrop: (paths: string[])=>void; children: ReactNode}) {
  void onDrop
  const {t}=useTranslation(); return <div className={`drop-zone ${disabled?'drop-disabled':''}`} style={{'--wails-drop-target':'drop'} as any} aria-disabled={disabled}>
    <span className="drop-icon" aria-hidden="true"><Icon name="addFiles" size={24}/></span>
    <div className="drop-copy"><strong>{t('inputs.drop')}</strong><span>{t('inputs.dropHint')}</span></div>
    {children}
  </div>
}
