import React,{useEffect,useRef,useState}from'react';
import{Fields}from'./components/ui';
import{InstagramImport}from'./generated/InstagramImport';
type ImportItem={key:string;kind:string;caption:string;created:number;mediaCount:number;archived:boolean;status:string;reason:string};
type Preview={digest:string;items:ImportItem[]};
export function InstagramImporter({request}:{request:(path:string,options?:any)=>Promise<any>}){
 const[fields,setFields]=useState<any>({visibility:'private'}),[review,setReview]=useState<Preview|null>(null),[busy,setBusy]=useState(''),[error,setError]=useState(''),[notice,setNotice]=useState(''),[progress,setProgress]=useState<number|null>(null),[elapsed,setElapsed]=useState(0);
 const running=useRef(false),alive=useRef(true),controller=useRef<AbortController|null>(null),started=useRef(0);
 useEffect(()=>()=>{alive.current=false;controller.current?.abort()},[]);
 useEffect(()=>{if(!busy)return;const timer=setInterval(()=>setElapsed(Math.floor((Date.now()-started.current)/1000)),1000);return()=>clearInterval(timer)},[!!busy]);
 const setField=(name:string,value:any)=>{if(running.current)return;setFields((f:any)=>({...f,[name]:value}));if(name==='archive'){setReview(null);setNotice('');setError('')}};
 const action=async(action:string)=>{
  if(action==='stop'){controller.current?.abort();return}if(running.current)return;
  setError('');setNotice('');
  const file:File|undefined=fields.archive?.[0];if(!file){setError('Choose your Instagram export ZIP first.');return}
  if(file.size>512*1024*1024){setError('ZIP must be at most 512 MiB. Export a smaller date range from Instagram.');return}
  const keys=review?.items.filter(i=>i.status==='ready').map(i=>i.key)||[];
  if(action==='apply'&&(!review||!keys.length))return;
  running.current=true;started.current=Date.now();setElapsed(0);setProgress(0);setBusy('Uploading ZIP · 0%');controller.current=new AbortController();const signal=controller.current.signal;
  const show=(text:string,value:number|null)=>{if(alive.current&&!signal.aborted){setBusy(text);setProgress(value)}};
  const transfer=async(mode:string,batch:string[]=[],completed=0)=>{
   const body=new FormData();body.append('options',JSON.stringify({stream:true,...(mode==='apply'?{digest:review!.digest,keys:batch,visibility:fields.visibility}:{})}));body.append('archive',file);
   return request('/api/imports/instagram/'+mode,{method:'POST',body,signal,
    onUploadProgress:({loaded,total}:any)=>{const percent=total?Math.min(100,Math.round(loaded/total*100)):null;show(`Uploading ZIP · ${(loaded/1048576).toFixed(1)}${total?' of '+(total/1048576).toFixed(1):''} MiB${percent===null?'':' · '+percent+'%'}${completed?' · '+completed+' posts processed':''}`,percent)},
    onUploaded:()=>show('Upload complete · waiting for the ZIP check…',null),
    onImportProgress:(p:any)=>{const done=p.phase==='importing'?completed+p.done:p.done,total=p.phase==='importing'?keys.length:p.total,percent=total?Math.round(done/total*100):null;const label=p.phase==='importing'?p.label.replace(/post \d+ of \d+/,`post ${completed+p.done+1} of ${keys.length}`):p.label;show(`${label}${total?' · '+done+' of '+total+' processed ('+percent+'%)':''}`,percent)}
   });
  };
  try{
   if(action==='preview'){const result=await transfer('preview');if(!alive.current)return;setReview(result);if(!result.items.length)setNotice('This ZIP has no posts to import. Check the categories and date range of your export.');return}
   const totals={imported:0,duplicates:0,failed:0};
   for(let offset=0;offset<keys.length&&!signal.aborted;offset+=100){
    const result=await transfer('apply',keys.slice(offset,offset+100),offset);if(!alive.current)return;
    totals.imported+=result.imported;totals.duplicates+=result.duplicates;totals.failed+=result.failed;
    const results=new Map<string,any>(result.results.map((r:any)=>[r.key,r]));
    setReview(previous=>previous?{...previous,items:previous.items.map(item=>{const r=results.get(item.key);return r?{...item,status:r.status==='failed'?'ready':r.status,reason:r.message}:item})}:null);
   }
   if(signal.aborted)throw new DOMException('Stopped','AbortError');
   const message=`${totals.imported} imported · ${totals.duplicates} already imported · ${totals.failed} failed. Successful posts are saved; you can retry failed posts safely.`;
   if(totals.failed)setError(message);else setNotice(message);
  }catch(e:any){if(alive.current){if(e.name==='AbortError')setNotice('Stopped. Completed posts are kept. Preview this ZIP again to refresh the counts, then retry safely.');else setError(e.message+' Completed posts are kept; preview and retry the same ZIP safely.')}}
  finally{running.current=false;if(alive.current){setBusy('');setProgress(null)}}
 };
 const all=review?.items||[],ready=all.filter(i=>i.status==='ready').length;
 const items=all.slice(0,5).map(item=>({...item,caption:item.caption||'(No caption)',date:new Date(item.created*1000).toLocaleString(undefined,{year:'numeric',month:'short',day:'numeric'}),archiveText:item.archived?'Stays archived':'',mediaText:`${item.mediaCount} media ${item.mediaCount===1?'file':'files'}`,tone:item.status==='unavailable'?'warning':item.status==='ready'?'muted':'success'}));
 const data={review:!!review,items,error,notice,busy,progress,elapsedText:`Elapsed ${Math.floor(elapsed/60)}:${String(elapsed%60).padStart(2,'0')} · Keep this page open.`,filename:fields.archive?.[0]?.name||'',summary:`${all.length} posts found · ${ready} available · ${all.filter(i=>i.status==='unavailable').length} unavailable`,sampleText:`Showing ${items.length} of ${all.length} posts. Import includes all ${ready} available posts.`,previewDisabled:!fields.archive?.length,applyDisabled:!ready,applyLabel:`Import ${ready} ${ready===1?'post':'posts'}`,audiences:[{value:'private',label:'Only me'},{value:'followers',label:'Followers'},{value:'close-friends',label:'Close Friends'},{value:'public',label:'Community · all signed-in members'}]};
 return React.createElement(Fields.Provider,{value:{fields,setField,disabled:!!busy}},React.createElement(InstagramImport,{data,onAction:(name:string)=>void action(name)}));
}
