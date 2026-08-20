import type { LanguagePreference } from '../i18n'
import { normalizeLanguage, useTranslation } from '../i18n'
import { normalizeTheme, type ThemePreference } from '../theme'

export function AppHeader({ appName, preference, onPreferenceChange, theme, onThemeChange, onAbout }: {
  appName: string
  preference: LanguagePreference
  onPreferenceChange: (value: LanguagePreference) => void
  theme: ThemePreference
  onThemeChange: (value: ThemePreference) => void
  onAbout: () => void
}) {
  const { t } = useTranslation()

  return <header className="app-header">
    <div className="app-identity">
      <p className="eyebrow">{t('app.eyebrow')}</p>
      <h1>{appName}</h1>
      <p className="muted">{t('app.subtitle')}</p>
    </div>
    <div className="header-tools">
      <label><span>{t('theme.label')}</span>
        <select value={theme} onChange={event => onThemeChange(normalizeTheme(event.target.value))}>
          <option value="SYSTEM">{t('theme.system')}</option>
          <option value="LIGHT">{t('theme.light')}</option>
          <option value="DARK">{t('theme.dark')}</option>
        </select>
      </label>
      <label><span>{t('language.label')}</span>
        <select value={preference} onChange={event => onPreferenceChange(normalizeLanguage(event.target.value))}>
          <option value="SYSTEM">{t('language.system')}</option>
          <option value="zh-CN">{t('language.zhCN')}</option>
          <option value="en-US">{t('language.enUS')}</option>
        </select>
      </label>
      <button className="text-button about-button" type="button" onClick={onAbout}>{t('about.title')}</button>
    </div>
  </header>
}
