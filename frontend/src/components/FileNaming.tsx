import { NamingConfig, NamingValidation } from '../api'

export const namingModes = [
  ['KEEP_SOURCE', 'Keep source name'], ['TITLE', 'Title'], ['ARTIST_TITLE', 'Artist - Title'],
  ['TITLE_ARTIST', 'Title - Artist'], ['TRACK_TITLE', 'Track - Title'],
  ['ARTIST_ALBUM_TRACK_TITLE', 'Artist - Album - Track - Title'], ['CUSTOM', 'Custom template'],
] as const
export const namingTokens = ['{title}', '{artist}', '{album}', '{albumartist}', '{track}', '{disc}', '{year}', '{source}']

export function FileNaming({ config, validation, disabled, onChange }: { config: NamingConfig; validation: NamingValidation; disabled: boolean; onChange: (config: NamingConfig) => void }) {
  return <section className="naming-section" aria-labelledby="file-naming-heading">
    <h2 id="file-naming-heading">File naming</h2>
    <label htmlFor="naming-mode">Output filename</label>
    <select id="naming-mode" value={config.mode} disabled={disabled} onChange={e => onChange({ ...config, mode: e.target.value })}>
      {namingModes.map(([value, label]) => <option value={value} key={value}>{label}</option>)}
    </select>
    {config.mode === 'CUSTOM' && <div className="custom-naming">
      <label htmlFor="naming-template">Custom template</label>
      <input id="naming-template" aria-describedby="naming-tokens naming-validation" value={config.customTemplate ?? ''} disabled={disabled} placeholder="{artist} - {title}" onChange={e => onChange({ ...config, customTemplate: e.target.value })} />
      <small id="naming-tokens" className="helper">Tokens: {namingTokens.join(' ')}</small>
      {!validation.valid && <p id="naming-validation" className="validation-error" role="alert">{validation.error}</p>}
    </div>}
    {validation.valid && <p className="naming-example"><span>Example</span> {validation.example}</p>}
    <p className="helper">The decoded audio extension is added automatically.</p>
  </section>
}
