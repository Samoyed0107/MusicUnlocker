import type { ReactNode } from 'react'

export function DropZone({disabled, onDrop, children}: {disabled: boolean; onDrop: (paths: string[])=>void; children: ReactNode}) {
  void onDrop
  return <div className={`drop-zone ${disabled?'drop-disabled':''}`} style={{'--wails-drop-target':'drop'} as any} aria-disabled={disabled}>
    <div className="drop-copy"><strong>Drop music files or folders here</strong><span>or use the buttons below</span></div>
    {children}
  </div>
}
