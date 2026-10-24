import React,{useEffect,useRef,useState}from'react';
import{createRoot}from'react-dom/client';
import{Export}from'./generated/Export';
import{Fields}from'./components/ui';
import{collectPage}from'./collector.mjs';
import{exportPosts}from'./archive.mjs';
import './styles.css';
declare const chrome:any;
const pause=(ms:number)=>new Promise(r=>setTimeout(r,ms));

async function saveZIP(blob:Blob,part:number,stamp:string){
 const url=URL.createObjectURL(blob),filename=`MySocial/instagram-capture-${stamp}-part-${String(part).padStart(3,'0')}.zip`;
 try{
  const id=await chrome.downloads.download({url,filename,saveAs:false});
  // The extension page stays alive until Chrome has consumed the Blob URL.
  for(let n=0;n<900;n++){const [d]=await chrome.downloads.search({id});if(d?.state==='complete')return filename;if(d?.state==='interrupted')throw Error('Download interrupted: '+(d.error||'retry the export'));await pause(1000)}
  throw Error('Download is still pending in Chrome. Check Downloads before closing this tab.');
 }finally{URL.revokeObjectURL(url)}
}

function App(){
 const[fields,setFields]=useState<any>({scroll:false,scrolls:'25',enrich:true}),[source,setSource]=useState(''),[posts,setPosts]=useState<any[]>([]),[selected,setSelected]=useState<string[]>([]),[busy,setBusy]=useState(''),[error,setError]=useState(''),[notice,setNotice]=useState(''),[warnings,setWarnings]=useState<string[]>([]),[downloads,setDownloads]=useState<any[]>([]);
 const controller=useRef<AbortController|null>(null),running=useRef(false),choice=useRef(''),profile=useRef<any>({}),capturedAt=useRef(Math.floor(Date.now()/1000));
 const tabID=Number(new URL(location.href).searchParams.get('tab'));
 const tab=async()=>{
  if(!Number.isInteger(tabID)||tabID<1)throw Error('Open an Instagram tab and click the extension toolbar icon.');
  const t=await chrome.tabs.get(tabID),u=new URL(t.url||'');
  if(u.protocol!=='https:'||!['www.instagram.com','instagram.com'].includes(u.hostname)||/^\/(direct|accounts|challenge|emails|privacy|oauth)(\/|$)/.test(u.pathname))throw Error('The selected tab must show an Instagram profile, feed, post or Reel.');
  return t;
 };
 useEffect(()=>{tab().then(t=>{const u=new URL(t.url);setSource(u.origin+u.pathname)}).catch(e=>setError(e.message));return()=>controller.current?.abort()},[]);
 const call=async(args:any)=>{await tab();const result=await chrome.scripting.executeScript({target:{tabId:tabID},world:'ISOLATED',func:collectPage,args:[args]});if(!result?.[0]?.result)throw Error('Instagram did not return accessible post content.');return result[0].result};
 const action=async(name:string)=>{
  if(name==='stop'){controller.current?.abort();setBusy('Stopping after the current request…');return}
  if(running.current)return;
  setError('');setNotice('');
  if(name==='all'){setSelected(posts.map(p=>p.permalink));return}if(name==='none'){setSelected([]);return}
  if(name==='toggle'){const key=choice.current;setSelected(xs=>xs.includes(key)?xs.filter(x=>x!==key):[...xs,key]);return}
  running.current=true;controller.current=new AbortController();const signal=controller.current.signal;
  try{
   if(name==='capture'){
    capturedAt.current=Math.floor(Date.now()/1000);setDownloads([]);setPosts([]);setSelected([]);setWarnings([]);const found=new Map<string,any>(),links=new Set<string>(),notes=new Set<string>();
    const add=(result:any)=>{
     for(const p of result.posts){const old=found.get(p.code);if(!old&&found.size>=1000){notes.add('Capture stopped at the 1,000-post limit. Export these posts first.');break}if(!old||p.media.length>old.media.length)found.set(p.code,{...p,caption:p.caption||old?.caption||'',created:p.created||old?.created||0,comments:p.comments||old?.comments||''});else {if(!old.caption)old.caption=p.caption;if(!old.created)old.created=p.created}}
     for(const link of result.links)if(links.size<1000)links.add(link);for(const note of result.warnings)notes.add(note);
     setPosts([...found.values()]);setSelected([...found.values()].map(p=>p.permalink));setWarnings([...notes]);
    };
    let previous='',stale=0;const steps=fields.scroll?Math.min(100,Math.max(1,Number(fields.scrolls)||25)):1;
    for(let i=0;i<steps&&!signal.aborted;i++){
     setBusy(`Collecting loaded posts · pass ${i+1} of ${steps}`);const result=await call({scroll:!!fields.scroll});profile.current=result.profile;add(result);
     const mark=found.size+':'+links.size+':'+result.height;if(mark===previous)stale++;else stale=0;previous=mark;
     if(stale>=3||found.size>=1000)break;
     if(fields.scroll)await pause(1500);
    }
    if(fields.enrich){
     const discovered=[...links].slice(0,1000);
     for(let i=0;i<discovered.length&&!signal.aborted;i++){
      setBusy(`Reading discovered post ${i+1} of ${discovered.length} · ${found.size} captured`);
      try{add(await call({url:discovered[i]}))}catch(e:any){notes.add('A post page could not be read: '+e.message);if(/HTTP (403|429)/.test(e.message)){notes.add('Instagram limited further page reads. Try again later.');break}}
      await pause(750);
     }
    }
    setWarnings([...notes]);setNotice((signal.aborted?'Capture stopped. ':'')+found.size+' posts captured. Review the selection before downloading.');
    if(!found.size)notes.add('No accessible post media found. Open a post, dismiss any login prompt yourself, or use Instagram’s official JSON export.');setWarnings([...notes]);
   }else if(name==='export'){
    setDownloads([]);const stamp=new Date().toISOString().replace(/[:.]/g,'-');
    const result=await exportPosts(posts.filter(p=>selected.includes(p.permalink)),{signal,profile:profile.current,now:capturedAt.current,onProgress:setBusy,emit:async(blob:Blob,part:number,info:any)=>{setBusy('Saving ZIP part '+part+'…');const filename=await saveZIP(blob,part,stamp);setDownloads(xs=>[...xs,{label:filename+' · '+info.posts+' importable posts'}])}});
    setNotice(`${result.parts} ZIP parts downloaded · ${result.importable} importable posts · ${result.failed} posts incomplete. See capture/report.json inside each ZIP for details.${result.cancelled?' Export stopped; completed parts remain downloaded.':''}`);
   }
  }catch(e:any){setError(e.message+' Completed ZIP downloads remain available.');}
  finally{running.current=false;setBusy('')}
 };
 const data={source,busy,error,notice,noSource:!source,hasPosts:!!posts.length,noSelection:!selected.length,summary:`${posts.length} captured · ${selected.length} selected`,scrollOptions:['10','25','50','100'],posts:posts.map(p=>({...p,caption:p.caption||'(No caption available)',date:p.created?new Date(p.created*1000).toLocaleString():'Capture date will be used',details:p.media.length+' media items · '+p.origin,warning:!p.created?'Original publication date unavailable.':p.origin==='page-preview'?'Only the page preview was available; carousel/caption may be incomplete.':p.media.some((m:any)=>m.duration>90)?'Includes video longer than My Social’s 90-second import limit.':'',selected:selected.includes(p.permalink),variant:selected.includes(p.permalink)?'selected':'secondary',label:selected.includes(p.permalink)?'Selected · remove':'Select post'})),hasWarnings:!!warnings.length,warnings:warnings.map(text=>({text})),hasDownloads:!!downloads.length,downloads};
 return React.createElement(Fields.Provider,{value:{fields,setField:(key:string,value:any)=>setFields((f:any)=>({...f,[key]:value})),disabled:!!busy,select:(key:string)=>{choice.current=key}}},React.createElement(Export,{data,onAction:(name:string)=>void action(name)}));
}
createRoot(document.getElementById('root')!).render(React.createElement(App));
