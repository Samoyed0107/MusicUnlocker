import type { NamingConfig, NamingValidation, SelectedInput } from '../api'
import { useTranslation } from '../i18n'
import { DropZone } from './DropZone'
import { FileNaming } from './FileNaming'
import { InputQueue } from './InputQueue'

export function SetupWorkspace({
  disabled, inputs, output, explicitOutput, workers, naming, namingValidation,
  onAddPaths, onAddFiles, onAddFolder, onRemoveInput, onClearInputs,
  onBrowseOutput, onResetSuggestion, onWorkersChange, onNamingChange,
}: {
  disabled: boolean
  inputs: SelectedInput[]
  output: string
  explicitOutput: boolean
  workers: number
  naming: NamingConfig
  namingValidation: NamingValidation
  onAddPaths: (paths: string[]) => void
  onAddFiles: () => void
  onAddFolder: () => void
  onRemoveInput: (path: string) => void
  onClearInputs: () => void
  onBrowseOutput: () => void
  onResetSuggestion: () => void
  onWorkersChange: (workers: number) => void
  onNamingChange: (config: NamingConfig) => void
}) {
  const { t } = useTranslation()

  return <section className="setup-workspace" aria-label={t('workflow.setup')}>
    <section className="card setup-primary" aria-labelledby="choose-music-heading">
      <div className="section-title">
        <div>
          <h2 id="choose-music-heading">{t('workflow.chooseMusic')}</h2>
        </div>
      </div>
      <DropZone disabled={disabled} onDrop={onAddPaths}>
        <div className="actions">
          <button onClick={onAddFiles} disabled={disabled}>{t('inputs.addFiles')}</button>
          <button onClick={onAddFolder} disabled={disabled}>{t('inputs.addFolder')}</button>
        </div>
      </DropZone>
      <InputQueue items={inputs} disabled={disabled} onRemove={onRemoveInput} onClear={onClearInputs}/>
    </section>

    <section className="card setup-secondary" aria-labelledby="destination-heading">
      <div className="section-title">
        <div>
          <h2 id="destination-heading">{t('workflow.destination')}</h2>
        </div>
      </div>
      <label>{t('output.folder')}
        <div className="inline">
          <input value={output} readOnly placeholder={t('output.choose')}/>
          <button onClick={onBrowseOutput} disabled={disabled}>{t('output.browse')}</button>
          {output && <button className="text-button" onClick={onResetSuggestion} disabled={disabled || !explicitOutput}>{t('output.reset')}</button>}
        </div>
      </label>
      <FileNaming config={naming} validation={namingValidation} disabled={disabled} onChange={onNamingChange}/>
      <details className="advanced-options">
        <summary>{t('workflow.advanced')}</summary>
        <label>{t('jobs.label')}
          <select value={workers} onChange={event => onWorkersChange(Number(event.target.value))} disabled={disabled}>
            {[1, 2, 4, 8].map(count => <option key={count}>{count}</option>)}
          </select>
        </label>
      </details>
    </section>
  </section>
}
