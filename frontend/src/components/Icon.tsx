import type { ReactNode } from 'react'

export type IconName = 'check' | 'warning' | 'error' | 'skip' | 'unknown' | 'pending' | 'info' | 'close'

const paths: Record<IconName, ReactNode> = {
  check: <path d="m5 12 4 4L19 6" />,
  warning: <><path d="M12 9v4"/><path d="M12 17h.01"/><path d="M10.3 3.8 2.4 18a2 2 0 0 0 1.7 3h15.8a2 2 0 0 0 1.7-3L13.7 3.8a2 2 0 0 0-3.4 0Z"/></>,
  error: <><circle cx="12" cy="12" r="9"/><path d="m9 9 6 6m0-6-6 6"/></>,
  skip: <><circle cx="12" cy="12" r="9"/><path d="M8 12h8"/></>,
  unknown: <><circle cx="12" cy="12" r="9"/><path d="M9.8 9a2.4 2.4 0 1 1 3.5 2.1c-.8.4-1.3 1-1.3 1.9"/><path d="M12 17h.01"/></>,
  pending: <><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></>,
  info: <><circle cx="12" cy="12" r="9"/><path d="M12 11v6"/><path d="M12 7h.01"/></>,
  close: <path d="m7 7 10 10M17 7 7 17"/>,
}

export function Icon({ name, size = 16 }: { name: IconName; size?: number }) {
  return <svg className="icon" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" focusable="false">
    {paths[name]}
  </svg>
}
