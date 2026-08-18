import { useEffect } from 'react'
import { backend, LaunchPathsEvent } from '../api'

export function useLaunchPaths(onEvent: (event: LaunchPathsEvent) => void) {
  const api = backend()
  useEffect(() => {
    let active = true
    api.GetPendingLaunchPaths?.().then((event: LaunchPathsEvent | null) => {
      if (active && event?.paths?.length) onEvent(event)
    }).catch(() => {})
    const runtime = window.runtime
    runtime?.EventsOn?.('musicunlocker:launch:paths', onEvent)
    return () => {
      active = false
      runtime?.EventsOff?.('musicunlocker:launch:paths')
    }
  }, [api, onEvent])
}
