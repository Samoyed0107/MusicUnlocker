import { NamingConfig, NamingValidation } from '../api'
import { useTranslation } from '../i18n'

export const namingModes = [
  ['KEEP_SOURCE', 'naming.keepSource'], ['TITLE', 'naming.titleMode'], ['ARTIST_TITLE', 'naming.artistTitle'],
  ['TITLE_ARTIST', 'naming.titleArtist'], ['TRACK_TITLE', 'naming.trackTitle'],
  ['ARTIST_ALBUM_TRACK_TITLE', 'naming.artistAlbumTrackTitle'], ['CUSTOM', 'naming.custom'],
] as const
export const namingTokens = ['{title}', '{artist}', '{album}', '{albumartist}', '{track}', '{disc}', '{year}', '{source}']

export function FileNaming({ config, validation, disabled, onChange }: { config: NamingConfig; validation: NamingValidation; disabled: boolean; onChange: (config: NamingConfig) => void }) {
  const {t}=useTranslation()
  return <section className="naming-section" aria-labelledby="file-naming-heading">
    <h2 id="file-naming-heading">{t('naming.title')}</h2>
    <label htmlFor="naming-mode">{t('naming.outputFilename')}</label>
    <select id="naming-mode" value={config.mode} disabled={disabled} onChange={e => onChange({ ...config, mode: e.target.value })}>
      {namingModes.map(([value, key]) => <option value={value} key={value}>{t(key as any)}</option>)}
    </select>
    {config.mode === 'CUSTOM' && <div className="custom-naming">
      <label htmlFor="naming-template">{t('naming.template')}</label>
      <input id="naming-template" aria-describedby="naming-tokens naming-validation" value={config.customTemplate ?? ''} disabled={disabled} placeholder="{artist} - {title}" onChange={e => onChange({ ...config, customTemplate: e.target.value })} />
      <small id="naming-tokens" className="helper">{t('naming.tokens',{tokens:namingTokens.join(' ')})}</small>
      {!validation.valid && <p id="naming-validation" className="validation-error" role="alert">{validation.error}</p>}
    </div>}
    {validation.valid && <p className="naming-example"><span>{t('naming.example')}</span> {validation.example}</p>}
    <p className="helper">{t('naming.extensionHelp')}</p>
  </section>
}
