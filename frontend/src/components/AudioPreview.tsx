import { useEffect, useRef, useState } from 'react'
import { PreviewResponse, QueueItem } from '../api'
import './preview.css'

export function formatDuration(value: number) {
  if (!Number.isFinite(value) || value < 0) return '0:00'
  const seconds = Math.floor(value)
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`
}

export function AudioPreview({ item, batchId, api, onClose, onError }: { item?: QueueItem; batchId?: number; api: any; onClose: () => void; onError: (message: string) => void }) {
  const audio = useRef<HTMLAudioElement>(null)
  const request = useRef(0)
  const [url, setUrl] = useState('')
  const [playing, setPlaying] = useState(false)
  const [current, setCurrent] = useState(0)
  const [duration, setDuration] = useState(0)
  const [failed, setFailed] = useState(false)

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
    }).catch((error: any) => { if (request.current === generation) onError(String(error)) })
    return () => { element.pause(); element.removeAttribute('src'); element.load() }
  }, [item?.index, batchId, api, onError])

  if (!item) return null
  const toggle = () => { const element = audio.current; if (!element) return; if (element.paused) void element.play().then(() => setPlaying(true)); else { element.pause(); setPlaying(false) } }
  const close = () => { audio.current?.pause(); onClose() }
  return <div className="audio-preview" aria-label="Audio preview">
    <audio ref={audio} preload="metadata" onLoadedMetadata={e => setDuration(e.currentTarget.duration)} onDurationChange={e => setDuration(e.currentTarget.duration)} onTimeUpdate={e => setCurrent(e.currentTarget.currentTime)} onEnded={() => setPlaying(false)} onError={() => { setPlaying(false); setFailed(true) }} />
    <div className="preview-heading"><strong>{item.sourceDisplayName}</strong><button className="text-button" onClick={close} aria-label="Close preview">Close</button></div>
    {failed ? <p className="preview-error" role="status">Preview unavailable for this audio format.</p> : <div className="preview-controls"><button onClick={toggle} disabled={!url} aria-label={playing ? 'Pause preview' : 'Play preview'}>{playing ? 'Pause' : 'Play'}</button><input type="range" min="0" max={Number.isFinite(duration) ? duration : 0} step="0.1" value={Math.min(current, duration || 0)} onChange={e => { const next=Number(e.target.value); if (audio.current) audio.current.currentTime=next; setCurrent(next) }} aria-label="Seek preview" disabled={!url || !duration}/><span>{formatDuration(current)} / {formatDuration(duration)}</span></div>}
  </div>
}
