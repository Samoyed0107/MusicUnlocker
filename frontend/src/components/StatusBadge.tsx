import { errorLabel, statusLabel } from '../api'
export function StatusBadge({status, code}: {status: string; code?: string}) { const detail=errorLabel(code); return <span className={`status-badge status-${status.toLowerCase()}`} title={detail || code}>{statusLabel(status)}{detail && <small>{detail}</small>}</span> }
