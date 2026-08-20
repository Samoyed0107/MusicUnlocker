import type { LanguagePreference } from '../i18n'
import { normalizeLanguage, useTranslation } from '../i18n'

export function AppHeader({ appName, preference, onPreferenceChange }: {
  appName: string
  preference: LanguagePreference
  onPreferenceChange: (value: LanguagePreference) => void
}) {
  const { t } = useTranslation()

  return <header className="app-header">
    <div className="app-identity">
      <p className="eyebrow">{t('app.eyebrow')}</p>
      <h1>{appName}</h1>
      <p className="muted">{t('app.subtitle')}</p>
    </div>
    <div className="header-tools">
      <label>{t('language.label')}
        <select value={preference} onChange={event => onPreferenceChange(normalizeLanguage(event.target.value))}>
          <option value="SYSTEM">{t('language.system')}</option>
          <option value="zh-CN">{t('language.zhCN')}</option>
          <option value="en-US">{t('language.enUS')}</option>
        </select>
      </label>
    </div>
  </header>
}
