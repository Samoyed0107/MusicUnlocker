import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { backend, BatchCompleteEvent, BatchEvent, dedupePaths, Preferences, QueueItem, SelectedInput, sortItems } from './api'
import { DropZone } from './components/DropZone'
import { InputQueue } from './components/InputQueue'
import { ProcessingTable } from './components/ProcessingTable'
import { AudioPreview } from './components/AudioPreview'
import { Summary } from './components/Summary'
import { useFileDrop } from './hooks/useFileDrop'
import './style.css'

const eventNames=['musicunlocker:batch:discovered','musicunlocker:batch:started','musicunlocker:batch:progress','musicunlocker:batch:finished','musicunlocker:batch:complete']

export default function App() {
  const api=backend(); const [inputs,setInputs]=useState<SelectedInput[]>([]); const [output,setOutput]=useState(''); const [explicitOutput,setExplicitOutput]=useState(false)
  const [workers,setWorkers]=useState(1); const [autoSuggest,setAutoSuggest]=useState(true); const [prefsLoaded,setPrefsLoaded]=useState(false); const [items,setItems]=useState<QueueItem[]>([])
  const [running,setRunning]=useState(false); const [cancelling,setCancelling]=useState(false); const [complete,setComplete]=useState<BatchCompleteEvent|null>(null); const [selectedPreview,setSelectedPreview]=useState<QueueItem>(); const [notice,setNotice]=useState(''); const [filter,setFilter]=useState('all'); const activeId=useRef<number|null>(null)
  const locked=running||cancelling
  const addPaths=useCallback(async(paths:string[])=>{const unique=dedupePaths(paths);if(!unique.length)return;try{const described:SelectedInput[]=await api.DescribeInputs(unique);setInputs(prev=>{const existing=new Set(prev.map(x=>x.path.toLowerCase()));const next=[...prev,...described.filter(x=>!existing.has(x.path.toLowerCase()))];if(next.length!==prev.length){setItems([]);setComplete(null)}return next})}catch(e:any){setNotice(`Could not add the selected paths: ${String(e)}`)}},[api])
  const addFiles=async()=>{try{await addPaths(await api.SelectFiles())}catch(e:any){setNotice(String(e))}}
  const addFolder=async()=>{try{const path=await api.SelectFolder();if(path)await addPaths([path])}catch(e:any){setNotice(String(e))}}
  const ignoredDrop=useCallback(()=>setNotice('Wait for the current batch to finish before adding more inputs.'),[])
  useFileDrop(locked,addPaths,ignoredDrop)
  useEffect(()=>{const update=(event:BatchEvent)=>{if(activeId.current===null||event.batchId!==activeId.current)return;setItems(prev=>sortItems([...prev.filter(x=>x.index!==event.item.index),event.item]))};const finish=(event:BatchCompleteEvent)=>{if(activeId.current===null||event.batchId!==activeId.current)return;setItems(sortItems(event.snapshot.results));setComplete(event);setRunning(false);setCancelling(false)};eventNames.slice(0,4).forEach(name=>window.runtime?.EventsOn?.(name,update));window.runtime?.EventsOn?.(eventNames[4],finish);return()=>{window.runtime?.EventsOff?.(...eventNames)}},[])
  useEffect(()=>{api.GetPreferences?.().then((p:Preferences)=>{if(p?.maxWorkers)setWorkers(p.maxWorkers);setAutoSuggest(p?.autoSuggestOutput??true);if(p?.theme)document.documentElement.dataset.theme=p.theme}).catch(()=>{}).finally(()=>setPrefsLoaded(true))},[api])
  useEffect(()=>{if(prefsLoaded&&!locked)api.SavePreferences?.({maxWorkers:workers,theme:(document.documentElement.dataset.theme as Preferences['theme'])||'system',autoSuggestOutput:autoSuggest}).catch(()=>{})},[workers,autoSuggest,locked,api,prefsLoaded])
  useEffect(()=>{if(!explicitOutput&&autoSuggest&&!locked){api.SuggestOutputFolder?.(inputs.map(x=>x.path)).then((suggested:string)=>{if(suggested)setOutput(suggested)}).catch(()=>{})}},[inputs,explicitOutput,autoSuggest,locked,api])
  const browseOutput=async()=>{try{const path=await api.SelectOutputFolder();if(path){setOutput(path);setExplicitOutput(true)}}catch(e:any){setNotice(String(e))}}
  const resetSuggestion=()=>{setExplicitOutput(false);setNotice('Output folder reset to the current suggestion.')}
  const start=async()=>{setNotice('');setSelectedPreview(undefined);if(!inputs.length||!output){setNotice('Choose at least one input and an output folder before starting.');return}try{const response=await api.StartBatch({inputs:inputs.map(x=>x.path),outputRoot:output,maxWorkers:workers});activeId.current=response.batchId;setItems([]);setComplete(null);setRunning(true)}catch(e:any){setNotice(String(e))}}
  const cancel=async()=>{setCancelling(true);try{await api.CancelBatch()}catch(e:any){setNotice(String(e));setCancelling(false)}}
  const openOutput=async()=>{try{await api.OpenOutputFolder(output)}catch(e:any){setNotice(String(e))}}
  const removeInput=(path:string)=>{setInputs(prev=>prev.filter(x=>x.path!==path));setItems([]);setComplete(null)}; const clearInputs=()=>{setInputs([]);setItems([]);setComplete(null)}
  const ordered=useMemo(()=>sortItems(items),[items]); const statusText=running?(cancelling?'Cancelling':'Running'):complete?'Complete':'Ready'
  const previewError=useCallback((message: string)=>setNotice(message.includes('PREVIEW_FILE_MISSING')?'Output file is no longer available.':message.includes('PREVIEW_SERVER_UNAVAILABLE')?'Preview is unavailable in this application session.':message),[])
  return <main><header><div><p className="eyebrow">LOCAL DESKTOP UTILITY</p><h1>MusicUnlocker</h1><p className="muted">Unlock supported music files while keeping sources untouched.</p></div><span className="pill">{statusText}</span></header>
    <section className="card controls"><DropZone disabled={locked} onDrop={addPaths}><div className="actions"><button onClick={addFiles} disabled={locked}>Add files</button><button onClick={addFolder} disabled={locked}>Add folder</button></div></DropZone><InputQueue items={inputs} disabled={locked} onRemove={removeInput} onClear={clearInputs}/><div className="form-grid"><label>Output folder<div className="inline"><input value={output} readOnly placeholder="Choose an output folder"/><button onClick={browseOutput} disabled={locked}>Browse</button>{output&&<button className="text-button" onClick={resetSuggestion} disabled={locked||!explicitOutput}>Reset suggestion</button>}</div></label><label>Concurrent jobs<select value={workers} onChange={e=>setWorkers(Number(e.target.value))} disabled={locked}>{[1,2,4,8].map(n=><option key={n}>{n}</option>)}</select></label></div><div className="footer-actions"><button className="primary" onClick={start} disabled={locked||!inputs.length||!output}>Start unlocking</button><button className="danger" onClick={cancel} disabled={!running||cancelling}>{cancelling?'Cancelling…':'Cancel'}</button></div>{!inputs.length&&<p className="helper">Add files or folders, or drop them above.</p>}{inputs.length&&!output&&<p className="helper">Choose an output folder to enable Start unlocking.</p>}{notice&&<div className="notice" role="status">{notice}</div>}</section>
    <section className="card"><div className="section-title"><h2>Progress / results</h2><div className="filters">{[['all','All'],['completed','Completed'],['skipped','Skipped'],['blocked','Blocked'],['problems','Problems']].map(([value,label])=><button key={value} className={filter===value?'selected':''} onClick={()=>setFilter(value)}>{label}</button>)}</div></div><div className="table-head"><span>Name / format</span><span>Status</span><span>Progress</span><span>Output / detail</span></div><ProcessingTable items={ordered} filter={filter} onPlay={setSelectedPreview}/><AudioPreview item={selectedPreview} batchId={complete?.batchId} api={api} onClose={()=>setSelectedPreview(undefined)} onError={previewError}/></section>
    <section className="card summary"><div className="section-title"><h2>Summary</h2>{complete&&<button className="text-button" onClick={openOutput}>Open output folder</button>}</div><Summary summary={complete?.snapshot} onOpen={openOutput}/></section>
  </main>
}
