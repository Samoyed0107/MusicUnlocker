import { useTranslation } from '../i18n'
import { Icon, type IconName } from './Icon'

const statusKeys: Record<string, Parameters<ReturnType<typeof useTranslation>['t']>[0]> = {
  SUCCESS: 'status.completed', SKIPPED_NOT_ENCRYPTED: 'status.notEncrypted', SKIPPED_OUTPUT_EXISTS: 'status.outputExists',
  BLOCKED_EXTERNAL_KEY: 'status.keyRequired', BLOCKED_EXTERNAL_DATABASE: 'status.databaseRequired', UNSUPPORTED: 'status.unsupported',
  UNKNOWN: 'status.unknown', FAILED: 'status.failed', CANCELLED: 'status.cancelled',
}

const presentation: Record<string, { tone: string; icon: IconName }> = {
  SUCCESS: { tone: 'completed', icon: 'check' },
  BLOCKED_EXTERNAL_KEY: { tone: 'blocked', icon: 'warning' }, BLOCKED_EXTERNAL_DATABASE: { tone: 'blocked', icon: 'warning' },
  FAILED: { tone: 'failed', icon: 'error' }, CANCELLED: { tone: 'neutral', icon: 'skip' },
  SKIPPED_NOT_ENCRYPTED: { tone: 'skipped', icon: 'skip' }, SKIPPED_OUTPUT_EXISTS: { tone: 'skipped', icon: 'skip' },
  UNSUPPORTED: { tone: 'skipped', icon: 'unknown' }, UNKNOWN: { tone: 'skipped', icon: 'unknown' },
}

export function StatusBadge({ status, code }: { status: string; code?: string }) {
  const { t } = useTranslation()
  const state = presentation[status] ?? { tone: 'running', icon: 'pending' as IconName }
  const label = t(statusKeys[status] || 'status.waiting')

  return <span className={`status-badge status--${state.tone}`} title={code} aria-label={label}>
    <span className="status-symbol"><Icon name={state.icon} /></span>
    <span>{label}</span>
  </span>
}
