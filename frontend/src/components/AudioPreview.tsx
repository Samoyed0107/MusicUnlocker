import { useEffect, useRef, useState } from 'react'
import { PreviewResponse, QueueItem } from '../api'
import './preview.css'
import { useTranslation } from '../i18n'

export function formatDuration(value: number) {
  if (!Number.isFinite(value) || value < 0) return '0:00'
  const seconds = Math.floor(value)
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`
}

export function AudioPreview({ item, batchId, api, onClose, onError }: { item?: QueueItem; batchId?: number; api: any; onClose: () => void; onError: (message: string) => void }) {
  const {t}=useTranslation()
  const audio = useRef<HTMLAudioElement>(null)
  const request = useRef(0)
  const onErrorRef = useRef(onError)
  const [url, setUrl] = useState('')
  const [playing, setPlaying] = useState(false)
  const [current, setCurrent] = useState(0)
  const [duration, setDuration] = useState(0)
  const [failed, setFailed] = useState(false)

  useEffect(() => { onErrorRef.current = onError }, [onError])

  useEffect(() => {
    const element = audio.current
    request.current++
    setUrl(''); setPlaying(false); setCurrent(0); setDuration(0); setFailed(false)
    if (!element) return
    element.pause(); element.removeAttribute('src'); element.load()
    if (!item || batchId === undefined) return
    const generation = request.current
    api.GetPreviewURL(batchId, item.index).then((response: PreviewResponse) => {
      if (request.current !== generation || !response?.available || !response.url) return
      setUrl(response.url)
      element.src = response.url
      void element.play().then(() => { if (request.current === generation) setPlaying(true) }).catch(() => {})
    }).catch((error: any) => { if (request.current === generation) onErrorRef.current(String(error)) })
    return () => { element.pause(); element.removeAttribute('src'); element.load() }
  }, [item?.index, batchId, api])

  if (!item) return null
  const toggle = () => { const element = audio.current; if (!element) return; if (element.paused) void element.play().then(() => setPlaying(true)); else { element.pause(); setPlaying(false) } }
  const close = () => { audio.current?.pause(); onClose() }
  return <div className="audio-preview" aria-label={t('preview.label')}>
    <audio ref={audio} preload="metadata" onLoadedMetadata={e => setDuration(e.currentTarget.duration)} onDurationChange={e => setDuration(e.currentTarget.duration)} onTimeUpdate={e => setCurrent(e.currentTarget.currentTime)} onEnded={() => setPlaying(false)} onError={() => { setPlaying(false); setFailed(true) }} />
    <div className="preview-heading"><span className="preview-identity"><strong title={item.sourcePath}>{item.sourceDisplayName}</strong>{item.outputPath&&<small title={item.outputPath} aria-label={item.outputPath}>{item.outputPath}</small>}</span><button className="text-button" onClick={close} aria-label={t('preview.close')}>{t('preview.close')}</button></div>
    {failed ? <p className="preview-error" role="status">{t('preview.unavailable')}</p> : <div className="preview-controls"><button onClick={toggle} disabled={!url} aria-label={playing ? t('preview.pause') : t('preview.play')}>{playing ? t('preview.pause') : t('preview.play')}</button><input type="range" min="0" max={Number.isFinite(duration) ? duration : 0} step="0.1" value={Math.min(current, duration || 0)} onChange={e => { const next=Number(e.target.value); if (audio.current) audio.current.currentTime=next; setCurrent(next) }} aria-label={t('preview.seek')} disabled={!url || !duration}/><span>{formatDuration(current)} / {formatDuration(duration)}</span></div>}
  </div>
}
