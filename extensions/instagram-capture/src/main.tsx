import React,{useEffect,useRef,useState}from'react';
import{createRoot}from'react-dom/client';
import{Export}from'./generated/Export';
import{Fields}from'./components/ui';
import{collectPage}from'./collector.mjs';
import{exportPosts}from'./archive.mjs';
import './styles.css';
declare const chrome:any;
const pause=(ms:number)=>new Promise(r=>setTimeout(r,ms));
const usernameOf=(url:string)=>{try{const p=new URL(url).pathname.split('/').filter(Boolean);return p.length<=2&&(!p[1]||p[1]==='reels')&&/^[A-Za-z0-9_.]{1,30}$/.test(p[0])&&!['explore','reels','direct','accounts','stories'].includes(p[0])?p[0]:''}catch{return''}};
async function saveZIP(blob:Blob,part:number,stamp:string,username:string){
 const url=URL.createObjectURL(blob),filename=`MySocial/instagram-${username}-${stamp}-part-${String(part).padStart(3,'0')}.zip`;
 try{
  const id=await chrome.downloads.download({url,filename,saveAs:false});
  for(let n=0;n<900;n++){const[d]=await chrome.downloads.search({id});if(d?.state==='complete')return{id,filename};if(d?.state==='interrupted')throw Error('Download interrupted: '+(d.error||'retry the download'));await pause(1000)}
  throw Error('Download is still pending in Chrome. Check Downloads before closing this tab.');
 }finally{URL.revokeObjectURL(url)}
}
function App(){
 const[fields,setFields]=useState<any>({username:'',scroll:true,scrolls:'100',reels:true,enrich:true}),[source,setSource]=useState(''),[posts,setPosts]=useState<any[]>([]),[selected,setSelected]=useState<string[]>([]),[busy,setBusy]=useState(''),[error,setError]=useState(''),[notice,setNotice]=useState(''),[warnings,setWarnings]=useState<string[]>([]),[downloads,setDownloads]=useState<any[]>([]),[discovered,setDiscovered]=useState(0);
 const controller=useRef<AbortController|null>(null),running=useRef(false),choice=useRef(''),profile=useRef<any>({}),capturedAt=useRef(0);
 const tabID=Number(new URL(location.href).searchParams.get('tab'));
 const tab=async(id=tabID)=>{
  if(!Number.isInteger(id)||id<1)throw Error('Open your Instagram profile and click the extension toolbar icon.');
  const t=await chrome.tabs.get(id),u=new URL(t.url||'');
  if(u.protocol!=='https:'||!['www.instagram.com','instagram.com'].includes(u.hostname)||/^\/(direct|accounts|challenge|emails|privacy|oauth)(\/|$)/.test(u.pathname))throw Error('Keep the selected tab on your Instagram profile.');
  return t;
 };
 useEffect(()=>{tab().then(t=>{const u=new URL(t.url);setSource(u.origin+u.pathname);setFields((f:any)=>({...f,username:usernameOf(t.url)}))}).catch(e=>setError(e.message));return()=>controller.current?.abort()},[]);
 const call=async(args:any,id=tabID)=>{await tab(id);const result=await chrome.scripting.executeScript({target:{tabId:id},world:'ISOLATED',func:collectPage,args:[args]});if(!result?.[0]?.result)throw Error('Instagram did not return accessible upload content.');return result[0].result};
 const action=async(name:string)=>{
  if(name==='stop'){controller.current?.abort();setBusy('Stopping after the current page or file…');return}
  if(name==='reveal'){if(downloads.length)await chrome.downloads.show(downloads.at(-1).id);return}
  if(running.current)return;
  setError('');setNotice('');
  if(name==='all'){setSelected(posts.map(p=>p.permalink));return}if(name==='none'){setSelected([]);return}
  if(name==='toggle'){const key=choice.current;setSelected(xs=>xs.includes(key)?xs.filter(x=>x!==key):[...xs,key]);return}
  const username=String(fields.username||'').replace(/^@/,'').trim().toLowerCase();
  if(!/^[a-z0-9_.]{1,30}$/.test(username)){setError('Enter your Instagram username or open your profile and click the extension again.');return}
  running.current=true;controller.current=new AbortController();const signal=controller.current.signal;let workerID:number|undefined;
  const renderPage=async(url:string)=>{
   const parsed=new URL(url),segments=parsed.pathname.split('/').filter(Boolean);if(segments.length===3&&['p','reel','tv'].includes(segments[1]))url=parsed.origin+'/'+segments.slice(-2).join('/')+'/';
   if(workerID===undefined){const w=await chrome.tabs.create({url:'about:blank',active:false});workerID=w.id}
   await chrome.tabs.update(workerID,{url});
   for(let n=0;n<120&&!signal.aborted;n++){
    await pause(250);const t=await chrome.tabs.get(workerID);
    if((t.url||'').split(/[?#]/)[0].replace(/\/$/,'')!==url.replace(/\/$/,''))continue;
    // Usable media can be ready while analytics/images keep tabs.status loading.
    try{const result=await chrome.scripting.executeScript({target:{tabId:workerID},world:'ISOLATED',func:()=>({url:location.origin+location.pathname,ready:document.readyState})});
     const state=result?.[0]?.result;if(state?.url.replace(/\/$/,'')===url.replace(/\/$/,'')&&state.ready!=='loading')return workerID!;
    }catch{/* Navigation has not committed its accessible document yet. */}
   }
   throw Error(signal.aborted?'Scan stopped':'Instagram page did not become readable.');
  };
  try{
   if(name==='capture'){
    const current=await tab();if(usernameOf(current.url).toLowerCase()!==username)throw Error('Open https://www.instagram.com/'+username+'/ in the source tab, then scan your uploads. Feed, saved and tagged posts are excluded.');
    capturedAt.current=Math.floor(Date.now()/1000);setDownloads([]);setPosts([]);setSelected([]);setWarnings([]);setDiscovered(0);
    const found=new Map<string,any>(),links=new Map<string,string>(),notes=new Set<string>();
    const add=(result:any)=>{
     for(const p of result.posts){
      if(p.owner&&p.owner.toLowerCase().replace(/^@/,'')!==username)continue;
      const old=found.get(p.code),rank:any={'grid-preview':0,'page-preview':1,'visible-page':2,'structured-data':3,'page-data':4};
      if(!old&&found.size>=1000){notes.add('Reached the 1,000-upload limit. Download these files before another scan.');break}
      if(!old||p.media.length>old.media.length||(rank[p.origin]||0)>(rank[old.origin]||0))found.set(p.code,{...p,caption:p.caption||old?.caption||'',created:p.created||old?.created||0});
      else{if(!old.caption)old.caption=p.caption;if(!old.created)old.created=p.created}
     }
     for(const link of result.links){const code=link.split('/').filter(Boolean).at(-1);if(links.size<1000)links.set(code,link)}
     for(const note of result.warnings)notes.add(note);
     setDiscovered(found.size);setPosts([...found.values()]);setSelected([...found.values()].map(p=>p.permalink));setWarnings([...notes]);
    };
    const scanGrid=async(id:number,label:string)=>{
     let previous='',stale=0;const steps=fields.scroll?Math.min(100,Math.max(1,Number(fields.scrolls)||100)):1;
     for(let i=0;i<steps&&!signal.aborted;i++){
      setBusy(`Finding your ${label} · scroll ${i+1} of ${steps} · ${links.size} uploads found`);
      const result=await call({username,scroll:!!fields.scroll,wait:i===0},id);if(label==='posts')profile.current=result.profile;add(result);
      const mark=links.size+':'+result.height;if(mark===previous)stale++;else stale=0;previous=mark;
      if(stale>=4||links.size>=1000)break;if(i===steps-1&&fields.scroll)notes.add(label+' scan reached the selected scroll limit; more uploads may remain.');
      if(fields.scroll)await pause(1500);
     }
    };
    const postsID=new URL(current.url).pathname.replace(/\/$/,'')==='/'+username?tabID:await renderPage('https://www.instagram.com/'+username+'/');
    await scanGrid(postsID,'posts');
    if(fields.reels&&!signal.aborted){
     setBusy('Opening your Reels grid…');
     try{await scanGrid(await renderPage('https://www.instagram.com/'+username+'/reels/'),'Reels')}
     catch(e:any){notes.add('Reels grid could not be read; continuing with the uploads already found. '+e.message)}
    }
    if(fields.enrich){
     const discoveredLinks=[...links];
     for(let i=0;i<discoveredLinks.length&&!signal.aborted;i++){
      const[code,url]=discoveredLinks[i];setBusy(`Reading your upload ${i+1} of ${discoveredLinks.length} · photos, videos and carousel slides`);
      try{
       const bits=new URL(url).pathname.split('/').filter(Boolean),trustedProfileCode=(bits.length===3&&bits[0].toLowerCase()===username)||!!found.get(code)?.ownerVerified;
       add(await call({username,url,trustedProfileCode}));const p=found.get(code);
       if(!p||p.incompleteMedia||!['page-data','structured-data'].includes(p.origin)){
        const id=await renderPage(url);add(await call({username,wait:true,carousel:true,trustedProfileCode},id));
       }
      }catch(e:any){notes.add('An upload could not be read: '+e.message);if(/HTTP (403|429)/.test(e.message)){notes.add('Instagram limited page reads. Download the available files and try again later.');break}}
      await pause(750);
     }
    }
    setWarnings([...notes]);setNotice((signal.aborted?'Scan stopped. ':'')+found.size+' uploads found. Photos and videos are downloaded when you click “Download my media ZIPs”.');
    if(!found.size){notes.add('No media was found. Check that your profile is loaded and signed in, then scan again.');setWarnings([...notes])}
   }else if(name==='export'){
    setDownloads([]);const stamp=new Date().toISOString().replace(/[:.]/g,'-');
    const result=await exportPosts(posts.filter(p=>selected.includes(p.permalink)),{signal,profile:profile.current,now:capturedAt.current,onProgress:setBusy,emit:async(blob:Blob,part:number,info:any)=>{setBusy('Saving media ZIP '+part+'…');const saved=await saveZIP(blob,part,stamp,username);setDownloads(xs=>[...xs,{...saved,label:saved.filename+' · '+info.posts+' complete uploads'}])}});
    setNotice(`${result.photos} photos and ${result.videos} videos saved in ${result.parts} ZIP files · ${result.failed} uploads incomplete. Unzip into media/posts/ to see your files. My Social can also import these same ZIPs.${result.cancelled?' Download stopped; completed ZIPs are kept.':''}`);
    if(!result.mediaFiles)setError('No photos or videos could be downloaded. The ZIP contains a failure report, not a media backup.');
   }
  }catch(e:any){setError(e.message+' Any completed ZIP downloads remain in Chrome Downloads.');}
  finally{if(workerID!==undefined)try{await chrome.tabs.remove(workerID)}catch{}running.current=false;setBusy('')}
 };
 const photos=posts.reduce((n,p)=>n+(p.incompleteMedia?0:p.media.filter((m:any)=>m.type==='image').length),0),videos=posts.reduce((n,p)=>n+(p.incompleteMedia?0:p.media.filter((m:any)=>m.type==='video').length),0);
 const data={source,busy,error,notice,noSource:!source,hasPosts:!!posts.length,noSelection:!selected.length,summary:`${discovered} uploads discovered · ${photos} photos · ${videos} videos · ${selected.length} uploads selected`,scrollOptions:['10','25','50','100'],posts:posts.map(p=>({...p,hasPreview:!!(p.media.find((m:any)=>m.type==='image')?.url||p.media[0]?.poster),preview:p.media.find((m:any)=>m.type==='image')?.url||p.media[0]?.poster,caption:p.caption||'(No caption available)',date:p.created?new Date(p.created*1000).toLocaleString():'Date unavailable',details:p.media.length+' media files',warning:p.incompleteMedia?'Only a preview or part of this upload was accessible. Recovered files are kept separately and marked incomplete.':p.origin==='page-preview'?'Preview only: the original carousel or video may be incomplete.':!p.created?'Original publication date unavailable.':'',selected:selected.includes(p.permalink),variant:selected.includes(p.permalink)?'selected':'secondary',label:selected.includes(p.permalink)?'Selected · remove':'Select upload'})),hasWarnings:!!warnings.length,warnings:warnings.map(text=>({text})),hasDownloads:!!downloads.length,downloads};
 return React.createElement(Fields.Provider,{value:{fields,setField:(key:string,value:any)=>{setFields((f:any)=>({...f,[key]:value}));if(key==='username'){setPosts([]);setSelected([]);setDownloads([]);setDiscovered(0)}},disabled:!!busy,select:(key:string)=>{choice.current=key}}},React.createElement(Export,{data,onAction:(name:string)=>void action(name)}));
}
createRoot(document.getElementById('root')!).render(React.createElement(App));
