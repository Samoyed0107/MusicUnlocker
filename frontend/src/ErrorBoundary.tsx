import { Component, type ErrorInfo, type ReactNode } from 'react'
import { useTranslation } from './i18n'

type Props = { children: ReactNode }
type State = { failed: boolean }

function ErrorFallback() {
  const { t } = useTranslation()
  return <main className="card"><h1>{t('errors.interfaceLoad')}</h1><p>{t('errors.restartApp')}</p></main>
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { failed: false }

  static getDerivedStateFromError(): State { return { failed: true } }

  componentDidCatch(error: Error, info: ErrorInfo) { console.error('MusicUnlocker frontend initialization failed', error, info) }

  render() {
    if (this.state.failed) return <ErrorFallback />
    return this.props.children
  }
}
