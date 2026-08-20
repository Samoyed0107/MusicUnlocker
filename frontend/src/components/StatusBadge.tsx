import { useTranslation } from '../i18n'

const statusKeys: Record<string, Parameters<ReturnType<typeof useTranslation>['t']>[0]> = {
  SUCCESS: 'status.completed', SKIPPED_NOT_ENCRYPTED: 'status.notEncrypted', SKIPPED_OUTPUT_EXISTS: 'status.outputExists',
  BLOCKED_EXTERNAL_KEY: 'status.keyRequired', BLOCKED_EXTERNAL_DATABASE: 'status.databaseRequired', UNSUPPORTED: 'status.unsupported',
  UNKNOWN: 'status.unknown', FAILED: 'status.failed', CANCELLED: 'status.cancelled',
}

const presentation: Record<string, { tone: string; symbol: string }> = {
  SUCCESS: { tone: 'completed', symbol: '✓' },
  BLOCKED_EXTERNAL_KEY: { tone: 'blocked', symbol: '!' }, BLOCKED_EXTERNAL_DATABASE: { tone: 'blocked', symbol: '!' },
  FAILED: { tone: 'failed', symbol: '×' }, CANCELLED: { tone: 'neutral', symbol: '–' },
  SKIPPED_NOT_ENCRYPTED: { tone: 'skipped', symbol: '–' }, SKIPPED_OUTPUT_EXISTS: { tone: 'skipped', symbol: '–' },
  UNSUPPORTED: { tone: 'skipped', symbol: '?' }, UNKNOWN: { tone: 'skipped', symbol: '?' },
}

export function StatusBadge({ status, code }: { status: string; code?: string }) {
  const { t } = useTranslation()
  const state = presentation[status] ?? { tone: 'running', symbol: '•' }
  const label = t(statusKeys[status] || 'status.waiting')

  return <span className={`status-badge status--${state.tone}`} title={code} aria-label={label}>
    <span className="status-symbol" aria-hidden="true">{state.symbol}</span>
    <span>{label}</span>
  </span>
}
