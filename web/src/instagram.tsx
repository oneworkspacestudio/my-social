import React,{useEffect,useRef,useState}from'react';
import{Fields}from'./components/ui';
import{InstagramImport}from'./generated/InstagramImport';

type ImportItem={key:string;kind:string;caption:string;created:number;mediaCount:number;archived:boolean;status:string;reason:string};
type Preview={digest:string;items:ImportItem[]};
export function InstagramImporter({request}:{request:(path:string,options?:any)=>Promise<any>}){
 const[fields,setFields]=useState<any>({visibility:'private'}),[review,setReview]=useState<Preview|null>(null),[selected,setSelected]=useState<string[]>([]),[busy,setBusy]=useState(''),[error,setError]=useState(''),[notice,setNotice]=useState('');
 const running=useRef(false),alive=useRef(true),controller=useRef<AbortController|null>(null);
 useEffect(()=>()=>{alive.current=false;controller.current?.abort()},[]);
 const setField=(name:string,value:any)=>{if(running.current)return;setFields((f:any)=>({...f,[name]:value}));if(name==='archive'){setReview(null);setSelected([]);setNotice('');setError('')}};
 const action=async(action:string,key?:string)=>{
  if(running.current)return;
  setError('');setNotice('');
  if(action==='selectAll'){setSelected(review?.items.filter(i=>i.status==='ready').slice(0,100).map(i=>i.key)||[]);return}
  if(action==='selectNone'){setSelected([]);return}
  if(action==='toggle'&&key){if(selected.includes(key))setSelected(xs=>xs.filter(x=>x!==key));else if(selected.length<100)setSelected(xs=>[...xs,key]);else setError('Import these 100 posts first, then select the next batch.');return}
  const file:File|undefined=fields.archive?.[0];
  if(!file){setError('Choose your Instagram export ZIP first.');return}
  if(file.size>512*1024*1024){setError('ZIP must be at most 512 MiB. Export a smaller date range from Instagram.');return}
  if(action==='apply'&&(!review||!selected.length))return;
  running.current=true;setBusy(action==='preview'?'Uploading and checking your ZIP…':'Importing selected posts… Keep this page open. If interrupted, preview and retry the same ZIP safely.');
  controller.current=new AbortController();
  try{
   const body=new FormData();body.append('options',JSON.stringify(action==='apply'?{digest:review!.digest,keys:selected,visibility:fields.visibility}:{}));body.append('archive',file);
   const result=await request('/api/imports/instagram/'+action,{method:'POST',body,signal:controller.current.signal});
   if(!alive.current)return;
   if(action==='preview'){setReview(result);setSelected([]);if(!result.items.length)setNotice('This ZIP has no posts to import. Check the categories and date range of your export.');return}
   const results=new Map<string,any>(result.results.map((r:any)=>[r.key,r]));
   setReview(previous=>previous?{...previous,items:previous.items.map(item=>{const r=results.get(item.key);return r?{...item,status:r.status==='failed'?'ready':r.status,reason:r.message}:item})}:null);
   setSelected(result.results.filter((r:any)=>r.status==='failed').map((r:any)=>r.key));
   const message=`${result.imported} imported · ${result.duplicates} already imported · ${result.failed} failed. Successful posts are saved; you can retry failed posts safely.`;
   if(result.failed)setError(message);else setNotice(message);
  }catch(e:any){if(alive.current&&e.name!=='AbortError')setError(e.message+' If an import was interrupted, preview and retry the same ZIP; completed posts will be skipped.')}
  finally{running.current=false;if(alive.current)setBusy('')}
 };
 const items=(review?.items||[]).map(item=>({...item,caption:item.caption||'(No caption)',date:new Date(item.created*1000).toLocaleString(undefined,{year:'numeric',month:'short',day:'numeric'}),archiveText:item.archived?'Stays archived':'',mediaText:`${item.mediaCount} media ${item.mediaCount===1?'file':'files'}`,tone:item.status==='unavailable'?'warning':item.status==='ready'?'muted':'success',disabled:item.status!=='ready',selected:selected.includes(item.key),variant:selected.includes(item.key)?'active':'secondary',selectLabel:item.status==='ready'?(selected.includes(item.key)?'Selected · remove':'Select post'):item.status==='unavailable'?'Unavailable':'Already imported'}));
 const data={review:!!review,items,error,notice,busy,filename:fields.archive?.[0]?.name||'',summary:`${items.length} posts found · ${items.filter(i=>i.status==='ready').length} available · ${items.filter(i=>i.status==='unavailable').length} unavailable`,selectionText:`${selected.length} of 100 selected for this batch`,previewDisabled:!fields.archive?.length,applyDisabled:!selected.length,audiences:[{value:'private',label:'Only me'},{value:'followers',label:'Followers'},{value:'close-friends',label:'Close Friends'},{value:'public',label:'Community · all signed-in members'}]};
 let selectedKey='';
 return React.createElement(Fields.Provider,{value:{fields,setField,disabled:!!busy,select:(key:string)=>{selectedKey=key}}},React.createElement(InstagramImport,{data,onAction:(name:string)=>void action(name,selectedKey)}));
}
