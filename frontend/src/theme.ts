export type ThemePreference = 'SYSTEM' | 'LIGHT' | 'DARK'

export function normalizeTheme(value: unknown): ThemePreference {
  if (value === 'LIGHT' || value === 'light') return 'LIGHT'
  if (value === 'DARK' || value === 'dark') return 'DARK'
  return 'SYSTEM'
}

export function applyTheme(theme: ThemePreference, setNativeTheme?: (theme: ThemePreference) => void) {
  document.documentElement.dataset.theme = theme
  document.documentElement.style.colorScheme = theme === 'SYSTEM' ? 'light dark' : theme.toLowerCase()
  setNativeTheme?.(theme)
}
