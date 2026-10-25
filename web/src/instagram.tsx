import React,{useEffect,useRef,useState}from'react';
import{Fields}from'./components/ui';
import{InstagramImport}from'./generated/InstagramImport';
type Item={key:string;kind:string;caption:string;created:number;mediaCount:number;archived:boolean;status:string;reason:string};
type Chunk={offset:number;size:number;sha256:string};
type Job={revision:number;retained:boolean;id:string;state:string;filename:string;size:number;received:number;fingerprint:string;chunks:Chunk[];chunkSize:number;items:Item[];postCount:number;ready:number;unavailable:number;label:string;done:number;total:number;error:string;imported:number;duplicates:number;failed:number;created:number;updated:number;visibility:string};
const processing=(j:Job|null)=>!!j&&['checking','queued_import','importing'].includes(j.state);
const hash=async(blob:Blob)=>Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',await blob.arrayBuffer())),n=>n.toString(16).padStart(2,'0')).join('');
export function InstagramImporter({request}:{request:(path:string,options?:any)=>Promise<any>}){
 const[fields,setFields]=useState<any>({visibility:'private'}),[job,setJob]=useState<Job|null>(null),[jobs,setJobs]=useState<Job[]>([]),[transfer,setTransfer]=useState(''),[uploadProgress,setUploadProgress]=useState<number|null>(null),[error,setError]=useState(''),[notice,setNotice]=useState(''),[clock,setClock]=useState(Date.now());
 const alive=useRef(true),running=useRef(false),controller=useRef<AbortController|null>(null),selected=useRef('');
 useEffect(()=>{alive.current=true;let timer:ReturnType<typeof setTimeout>;const poll=async()=>{try{const result=await request('/api/imports/instagram/jobs');if(alive.current){setJobs(result.jobs);setJob(previous=>{const next=result.jobs.find((j:Job)=>j.id===previous?.id);return next&&(!previous||next.revision>=previous.revision)?next:previous||result.jobs[0]||null})}}catch(e:any){if(alive.current)setError(e.message)}finally{if(alive.current)timer=setTimeout(poll,1000)}};void poll();const tick=setInterval(()=>setClock(Date.now()),1000);return()=>{alive.current=false;clearTimeout(timer);clearInterval(tick);controller.current?.abort()}},[request]);
 useEffect(()=>{if(job)setFields((previous:any)=>({...previous,visibility:job.visibility||'private'}))},[job?.id]);
 const setField=(name:string,value:any)=>{if(running.current)return;setFields((previous:any)=>({...previous,[name]:value}));setError('');setNotice('')};
 const show=(text:string,percent:number|null)=>{if(alive.current){setTransfer(text);setUploadProgress(percent)}};
 const upload=async()=>{
  const file:File|undefined=fields.archive?.[0];if(!file)throw Error('Choose your Instagram export ZIP first.');if(file.size>512*1048576)throw Error('ZIP must be at most 512 MiB.');
  const signal=controller.current!.signal;const fingerprint=await hash(new Blob([file.name+'|'+file.size,file.slice(0,65536),file.slice(Math.max(0,file.size-65536))]));
  let current:Job=await request('/api/imports/instagram/jobs',{method:'POST',signal,body:JSON.stringify({filename:file.name,size:file.size,fingerprint})});if(alive.current)setJob(previous=>previous&&previous.id===current.id&&previous.revision>current.revision?previous:current);
  const verify=async(j:Job)=>{for(const chunk of j.chunks||[]){if(signal.aborted)throw new DOMException('Paused','AbortError');show('Checking saved upload chunks…',null);if(await hash(file.slice(chunk.offset,chunk.offset+chunk.size))!==chunk.sha256)throw Error('This ZIP differs from the saved upload. Cancel that upload or choose the original ZIP.')}};
  await verify(current);
  while(current.received<file.size){
   const offset=current.received,part=file.slice(offset,Math.min(file.size,offset+current.chunkSize)),digest=await hash(part);let completed=false;
   for(let attempt=0;attempt<4&&!completed;attempt++){
    if(signal.aborted)throw new DOMException('Paused','AbortError');
    try{current=await request('/api/imports/instagram/jobs/'+current.id+'/chunks?offset='+offset,{method:'PUT',body:part,signal,headers:{'Content-Type':'application/octet-stream','X-Chunk-SHA256':digest},onUploadProgress:({loaded}:any)=>{const done=Math.min(file.size,offset+loaded),percent=Math.round(done/file.size*100);show(`Uploading ZIP · ${(done/1048576).toFixed(1)} of ${(file.size/1048576).toFixed(1)} MiB · ${percent}%`,percent)}});completed=true;if(alive.current)setJob(previous=>previous&&previous.id===current.id&&previous.revision>current.revision?previous:current)}catch(e:any){if(signal.aborted)throw e;if(attempt===3||e.status===400||e.status===403||e.status===401)throw e;show('Reconnecting to your saved upload…',null);await new Promise(resolve=>setTimeout(resolve,1000*(attempt+1)));current=await request('/api/imports/instagram/jobs/'+current.id,{signal});await verify(current);if(current.received>offset)completed=true}
   }
  }
  current=await request('/api/imports/instagram/jobs/'+current.id+'/complete',{method:'POST',signal,body:'{}'});if(alive.current){setJob(previous=>previous&&previous.id===current.id&&previous.revision>current.revision?previous:current);setNotice('ZIP saved on the server. You can leave this page while it is checked.')}
 };
 const action=async(name:string)=>{
  if(name==='pause'){controller.current?.abort();return}
  if(name==='openJob'){setJob(jobs.find(j=>j.id===selected.current)||null);return}
  if(running.current)return;running.current=true;setError('');setNotice('');controller.current=new AbortController();
  try{
   if(name==='preview'){show('Preparing resumable upload…',null);await upload();return}
   if(!job)return;
   const endpoint=name==='apply'?'import':name==='retry'?'retry':'cancel';const result=await request('/api/imports/instagram/jobs/'+job.id+'/'+endpoint,{method:'POST',body:JSON.stringify({visibility:fields.visibility}),signal:controller.current.signal});if(alive.current)setJob(previous=>previous&&previous.id===result.id&&previous.revision>result.revision?previous:result);
  }catch(e:any){if(alive.current){if(e.name==='AbortError')setNotice('Upload paused. The server keeps completed chunks; choose the same ZIP and resume.');else setError(e.message+(name==='preview'?' Completed upload chunks are kept. Choose the same ZIP to resume.':''))}}
  finally{running.current=false;if(alive.current){setTransfer('');setUploadProgress(null)}}
 };
 const busy=transfer||(processing(job)?job!.label:'');const percent=transfer?uploadProgress:job&&job.total?Math.min(100,Math.round(job.done/job.total*100)):null;
 const seconds=job?Math.max(0,Math.floor(clock/1000-job.created)):0;
 const summary=job?`${job.postCount} posts found · ${job.ready} available · ${job.unavailable} unavailable`:'';
 const items=(job?.items||[]).map(item=>({...item,caption:item.caption||'(No caption)',date:new Date(item.created*1000).toLocaleString(undefined,{year:'numeric',month:'short',day:'numeric'}),archiveText:item.archived?'Stays archived':'',mediaText:`${item.mediaCount} media ${item.mediaCount===1?'file':'files'}`,tone:item.status==='unavailable'?'warning':item.status==='ready'?'muted':'success'}));
 const complete=job?.state==='completed',failed=job?.failed||0;
 const data={items,review:!!job&&['ready','completed'].includes(job.state),error:error||job?.error,notice:notice||(complete?`${job!.imported} imported · ${job!.duplicates} already imported · ${failed} failed. Successful posts are saved; you can retry failed posts safely.`:job?.state==='cancelled'?'Cancelled. Completed posts are kept.':''),busy,progress:percent,phaseCounts:processing(job)&&job!.total?`${job!.done} of ${job!.total} processed (${percent}%)`:'',elapsedText:`Elapsed ${Math.floor(seconds/60)}:${String(seconds%60).padStart(2,'0')}`,serverMessage:transfer?'Keep this page open until the ZIP upload finishes. Interrupted uploads can be resumed.':processing(job)?'Your ZIP is saved. Processing continues on the server if you leave this page. Return here to see progress and results.':job?.state==='uploading'?`Upload paused at ${(job.received/1048576).toFixed(1)} of ${(job.size/1048576).toFixed(1)} MiB. Choose the same ZIP and resume.`:'',showPause:!!transfer,showCancel:!!job&&!transfer&&job.retained,showRetry:job?.state==='failed'&&job.retained&&job.received===job.size,filename:fields.archive?.[0]?.name||job?.filename||'',summary,sampleText:`Showing ${items.length} of ${job?.postCount||0} posts. Import includes all ${job?.ready||0} available posts.`,previewDisabled:!fields.archive?.length,previewLabel:job?.state==='uploading'?'Resume ZIP upload':'Upload and preview ZIP',applyDisabled:job?.state!=='ready'&&!(complete&&failed>0)||!job?.ready&&!(complete&&failed>0),applyLabel:complete&&failed>0?`Retry ${failed} failed ${failed===1?'post':'posts'}`:`Import ${job?.ready||0} ${job?.ready===1?'post':'posts'}`,jobs:jobs.map(j=>({id:j.id,label:`${j.filename} · ${j.state} · ${j.done} of ${j.total}`})),hasJobs:jobs.length>1,audiences:[{value:'private',label:'Only me'},{value:'followers',label:'Followers'},{value:'close-friends',label:'Close Friends'},{value:'public',label:'Community · all signed-in members'}]};
 return React.createElement(Fields.Provider,{value:{fields,setField,select:(id:string)=>{selected.current=id},disabled:!!transfer}},React.createElement(InstagramImport,{data,onAction:(name:string)=>void action(name)}));
}
