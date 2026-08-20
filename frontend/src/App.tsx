import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { backend, AppInfo, BatchCompleteEvent, BatchEvent, dedupePaths, LaunchPathsEvent, NamingConfig, NamingValidation, normalizePath, Preferences, QueueItem, SelectedInput, sortItems } from './api'
import { AppHeader } from './components/AppHeader'
import { BatchCommandBar } from './components/BatchCommandBar'
import { ResultsWorkspace } from './components/ResultsWorkspace'
import { SetupWorkspace } from './components/SetupWorkspace'
import { useFileDrop } from './hooks/useFileDrop'
import { useLaunchPaths } from './hooks/useLaunchPaths'
import './style.css'
import { normalizeLanguage, useTranslation } from './i18n'

const eventNames=['musicunlocker:batch:discovered','musicunlocker:batch:started','musicunlocker:batch:progress','musicunlocker:batch:finished','musicunlocker:batch:complete']

export default function App() {
  const {t, preference, setPreference}=useTranslation()
  const api=backend(); const [appInfo,setAppInfo]=useState<Pick<AppInfo,'name'>>({name:'MusicUnlocker'}); const [inputs,setInputs]=useState<SelectedInput[]>([]); const [output,setOutput]=useState(''); const [explicitOutput,setExplicitOutput]=useState(false)
  const [workers,setWorkers]=useState(1); const [autoSuggest,setAutoSuggest]=useState(true); const [naming,setNaming]=useState<NamingConfig>({mode:'KEEP_SOURCE'}); const [namingValidation,setNamingValidation]=useState<NamingValidation>({valid:true,example:'source-file.mp3'}); const [prefsLoaded,setPrefsLoaded]=useState(false); const [items,setItems]=useState<QueueItem[]>([])
  const [running,setRunning]=useState(false); const [cancelling,setCancelling]=useState(false); const [retryingIndex,setRetryingIndex]=useState<number>(); const [complete,setComplete]=useState<BatchCompleteEvent|null>(null); const [selectedPreview,setSelectedPreview]=useState<QueueItem>(); const [notice,setNotice]=useState(''); const [filter,setFilter]=useState('all'); const activeId=useRef<number|null>(null)
  const locked=running||cancelling
  useEffect(()=>{api.GetAppInfo?.().then(setAppInfo).catch(()=>{})},[api])
  const addPaths=useCallback(async(paths:string[])=>{const unique=dedupePaths(paths);if(!unique.length)return;try{const described:SelectedInput[]=await api.DescribeInputs(unique);setInputs(prev=>{const existing=new Set(prev.map(x=>normalizePath(x.path)));return [...prev,...described.filter(x=>!existing.has(normalizePath(x.path)))]})}catch(e:any){setNotice(String(e))}},[api])
  const addFiles=async()=>{try{await addPaths(await api.SelectFiles())}catch(e:any){setNotice(String(e))}}
  const addFolder=async()=>{try{const path=await api.SelectFolder();if(path)await addPaths([path])}catch(e:any){setNotice(String(e))}}
  const ignoredDrop=useCallback(()=>setNotice(t('inputs.locked')),[t])
  useFileDrop(locked,addPaths,ignoredDrop)
  const onLaunchPaths=useCallback((event:LaunchPathsEvent)=>{const vars={count:event.paths.length};if(event.deferred){setNotice(t('batch.itemsReceived',vars));return} addPaths(event.paths).then(()=>setNotice(t('batch.itemsAdded',vars)))},[addPaths,t])
  useLaunchPaths(onLaunchPaths)
  useEffect(()=>{const update=(event:BatchEvent)=>{if(activeId.current===null||event.batchId!==activeId.current)return;setItems(prev=>sortItems([...prev.filter(x=>x.index!==event.item.index),event.item]))};const finish=(event:BatchCompleteEvent)=>{if(activeId.current===null||event.batchId!==activeId.current)return;setItems(sortItems(event.snapshot.results));setComplete(event);setRunning(false);setCancelling(false);setRetryingIndex(undefined)};eventNames.slice(0,4).forEach(name=>window.runtime?.EventsOn?.(name,update));window.runtime?.EventsOn?.(eventNames[4],finish);return()=>{window.runtime?.EventsOff?.(...eventNames)}},[])
  useEffect(()=>{api.GetPreferences?.().then((p:Preferences)=>{if(p?.maxWorkers)setWorkers(p.maxWorkers);setAutoSuggest(p?.autoSuggestOutput??true);if(p?.theme)document.documentElement.dataset.theme=p.theme;setNaming({mode:p?.namingMode||'KEEP_SOURCE',customTemplate:p?.customNamingTemplate});setPreference(normalizeLanguage(p?.language))}).catch(()=>{}).finally(()=>setPrefsLoaded(true))},[api,setPreference])
  useEffect(()=>{if(!api.ValidateNamingConfig)return;api.ValidateNamingConfig(naming).then((v:NamingValidation)=>setNamingValidation(v)).catch(()=>setNamingValidation({valid:false,error:t('naming.unavailable')}))},[api,naming,t])
  useEffect(()=>{if(prefsLoaded&&!locked)api.SavePreferences?.({maxWorkers:workers,theme:(document.documentElement.dataset.theme as Preferences['theme'])||'system',autoSuggestOutput:autoSuggest,namingMode:naming.mode,customNamingTemplate:naming.customTemplate,language:preference}).catch(()=>{})},[workers,autoSuggest,naming,locked,api,prefsLoaded,preference])
  useEffect(()=>{if(!explicitOutput&&autoSuggest&&!locked){api.SuggestOutputFolder?.(inputs.map(x=>x.path)).then((suggested:string)=>{if(suggested)setOutput(suggested)}).catch(()=>{})}},[inputs,explicitOutput,autoSuggest,locked,api])
  const browseOutput=async()=>{try{const path=await api.SelectOutputFolder();if(path){setOutput(path);setExplicitOutput(true)}}catch(e:any){setNotice(String(e))}}
  const resetSuggestion=()=>{setExplicitOutput(false);setNotice(t('output.resetDone'))}
  const start=async()=>{setNotice('');setSelectedPreview(undefined);if(!inputs.length||!output){setNotice(t('batch.chooseInputs'));return}if(!namingValidation.valid){setNotice(namingValidation.error||t('batch.fixNaming'));return}try{const response=await api.StartBatch({inputs:inputs.map(x=>x.path),outputRoot:output,maxWorkers:workers,naming:{...naming}});activeId.current=response.batchId;setItems([]);setComplete(null);setRunning(true)}catch(e:any){setNotice(String(e))}}
  const cancel=async()=>{setCancelling(true);try{await api.CancelBatch()}catch(e:any){setNotice(String(e));setCancelling(false)}}
  const retry=async(item: QueueItem)=>{if(retryingIndex!==undefined||!complete)return;setNotice('');setRetryingIndex(item.index);try{if(item.retryableExternalDatabase||item.status==='BLOCKED_EXTERNAL_DATABASE')await api.RetryKGGWithDatabase(complete.batchId,item.index);else await api.RetryQMCWithExternalKey(complete.batchId,item.index)}catch(e:any){const message=String(e);setNotice(message.includes('KGG_DATABASE_INVALID')?t('errors.kggDatabaseInvalid'):message.includes('KGG_DATABASE_SELECTION_FAILED')?'':message.includes('QMC_EXTERNAL_KEY_FILE_INVALID')||message.includes('EKEY_FILE_INVALID')?t('errors.qmcKeyFileInvalid'):message);setRetryingIndex(undefined)}}
  const openOutput=async()=>{try{await api.OpenOutputFolder(output)}catch(e:any){setNotice(String(e))}}
  const removeInput=(path:string)=>{setInputs(prev=>prev.filter(x=>x.path!==path));setItems([]);setComplete(null)}; const clearInputs=()=>{setInputs([]);setItems([]);setComplete(null)}
  const ordered=useMemo(()=>sortItems(items),[items]); const statusText=running?(cancelling?t('batch.cancelling'):t('batch.running')):complete?t('batch.complete'):t('batch.ready')
  const previewError=useCallback((message: string)=>setNotice(message.includes('PREVIEW_FILE_MISSING')?t('errors.previewMissing'):message.includes('PREVIEW_SERVER_UNAVAILABLE')?t('errors.previewUnavailable'):message),[t])

  return <main>
    <AppHeader appName={appInfo.name} preference={preference} onPreferenceChange={setPreference}/>
    <SetupWorkspace disabled={locked} inputs={inputs} output={output} explicitOutput={explicitOutput} workers={workers} naming={naming} namingValidation={namingValidation} onAddPaths={addPaths} onAddFiles={addFiles} onAddFolder={addFolder} onRemoveInput={removeInput} onClearInputs={clearInputs} onBrowseOutput={browseOutput} onResetSuggestion={resetSuggestion} onWorkersChange={setWorkers} onNamingChange={setNaming}/>
    <BatchCommandBar statusText={statusText} running={running} cancelling={cancelling} startDisabled={locked||!inputs.length||!output||!namingValidation.valid} onStart={start} onCancel={cancel} notice={notice} showAddHint={!inputs.length} showOutputHint={!!inputs.length&&!output}/>
    <ResultsWorkspace items={ordered} filter={filter} onFilterChange={setFilter} onPlay={setSelectedPreview} onRetry={retry} retryingIndex={retryingIndex} selectedPreview={selectedPreview} complete={complete} api={api} onClosePreview={()=>setSelectedPreview(undefined)} onPreviewError={previewError} onOpenOutput={openOutput}/>
  </main>
}
