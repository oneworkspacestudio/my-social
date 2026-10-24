export const PART_LIMIT=400*1024*1024;
export const MEDIA_LIMIT=100*1024*1024;
export function safeMediaURL(value){
 try{const u=new URL(value);return u.protocol==='https:'&&!u.username&&!u.password&&(!u.port||u.port==='443')&&['cdninstagram.com','fbcdn.net'].some(h=>u.hostname===h||u.hostname.endsWith('.'+h))?u.href:''}catch{return ''}
}
export function mediaType(b){
 if(b[0]===255&&b[1]===216&&b[2]===255)return 'jpg';
 if(b[0]===137&&String.fromCharCode(...b.slice(1,4))==='PNG')return 'png';
 if(String.fromCharCode(...b.slice(0,3))==='GIF')return 'gif';
 if(String.fromCharCode(...b.slice(0,4))==='RIFF'&&String.fromCharCode(...b.slice(8,12))==='WEBP')return 'webp';
 if(String.fromCharCode(...b.slice(4,8))==='ftyp')return 'mp4';
 if(b[0]===26&&b[1]===69&&b[2]===223&&b[3]===163)return 'webm';
 return '';
}
const table=Uint32Array.from({length:256},(_,n)=>{for(let i=0;i<8;i++)n=n&1?0xedb88320^(n>>>1):n>>>1;return n>>>0});
export function crc32(b){let c=0xffffffff;for(const n of b)c=table[(c^n)&255]^(c>>>8);return(c^0xffffffff)>>>0}
export function zipBlob(files){
 const body=[],directory=[];let offset=0;
 for(const{name,data}of files){
  if(!/^[A-Za-z0-9_./-]+$/.test(name)||name.startsWith('/')||name.split('/').includes('..'))throw Error('Unsafe archive member');
  const filename=new TextEncoder().encode(name),crc=crc32(data),h=new Uint8Array(30),v=new DataView(h.buffer);
  v.setUint32(0,0x04034b50,true);v.setUint16(4,20,true);v.setUint16(6,0x800,true);v.setUint32(14,crc,true);v.setUint32(18,data.length,true);v.setUint32(22,data.length,true);v.setUint16(26,filename.length,true);body.push(h,filename,data);
  const d=new Uint8Array(46),w=new DataView(d.buffer);w.setUint32(0,0x02014b50,true);w.setUint16(4,20,true);w.setUint16(6,20,true);w.setUint16(8,0x800,true);w.setUint32(16,crc,true);w.setUint32(20,data.length,true);w.setUint32(24,data.length,true);w.setUint16(28,filename.length,true);w.setUint32(42,offset,true);directory.push(d,filename);offset+=h.length+filename.length+data.length;
 }
 const size=directory.reduce((n,x)=>n+x.length,0),end=new Uint8Array(22),v=new DataView(end.buffer);v.setUint32(0,0x06054b50,true);v.setUint16(8,files.length,true);v.setUint16(10,files.length,true);v.setUint32(12,size,true);v.setUint32(16,offset,true);
 if(files.length>20000||offset+size+22>512*1024*1024)throw Error('Archive exceeds My Social import limits');
 return new Blob([...body,...directory,end],{type:'application/zip'});
}
const jsonFile=(name,value)=>({name,data:new TextEncoder().encode(JSON.stringify(value,null,2))});
export function importCaption(value){value=value||'';if(new TextEncoder().encode(value).length<=2200)return value;let out='',bytes=0;for(const c of value){const size=new TextEncoder().encode(c).length;if(bytes+size>2197)return out+'…';out+=c;bytes+=size}return out}

export async function fetchMedia(url,signal){
 const safe=safeMediaURL(url);if(!safe)throw Error('Media URL is outside Instagram media hosts');
 const controller=new AbortController(),timer=setTimeout(()=>controller.abort(),45000),abort=()=>controller.abort();signal?.addEventListener('abort',abort,{once:true});
 if(signal?.aborted)controller.abort();
 try{
  const response=await fetch(safe,{credentials:'omit',redirect:'error',referrerPolicy:'no-referrer',signal:controller.signal});
  if(!response.ok)throw Error('Media returned HTTP '+response.status);
  if(Number(response.headers.get('Content-Length'))>MEDIA_LIMIT)throw Error('Media exceeds 100 MiB');
  const reader=response.body.getReader(),chunks=[];let size=0;
  while(true){const{value,done}=await reader.read();if(done)break;size+=value.length;if(size>MEDIA_LIMIT){await reader.cancel();throw Error('Media exceeds 100 MiB')}chunks.push(value)}
  const data=new Uint8Array(await new Blob(chunks).arrayBuffer()),ext=mediaType(data);if(!ext)throw Error('Unsupported media response');return {data,ext};
 } finally{clearTimeout(timer);signal?.removeEventListener('abort',abort)}
}

// Each ZIP contains its own complete manifest/media. Download one part before
// releasing its buffers; all parts stay below the existing 512 MiB importer cap.
/** @param {Array<any>} posts @param {any} options */
export async function exportPosts(posts,{fetchFile=fetchMedia,emit,signal,onProgress=(_message)=>{},profile={},now=Math.floor(Date.now()/1000),partLimit=PART_LIMIT}={}){
 let files=[],entries=[],report=[],bytes=0,part=0,importable=0,failed=0,mediaFiles=0,photos=0,videos=0;
 const flush=async()=>{
  if(!files.length&&!report.length)return;
  const archive=zipBlob([...files,jsonFile('content/posts_1.json',entries.filter(x=>x.kind!=='reel'&&x.kind!=='story').map(x=>x.entry)),jsonFile('content/reels.json',{ig_reels_media:entries.filter(x=>x.kind==='reel').map(x=>x.entry)}),jsonFile('content/stories.json',{ig_stories:entries.filter(x=>x.kind==='story').map(x=>x.entry)}),jsonFile('capture/report.json',{format:'my-social-instagram-capture-v1',capturedAt:now,profile,posts:report,limitations:['Only accessible tab/page content was captured. Missing items are reported.','Dates marked capture-time are not original publication dates.','Carousels larger than 10 items are split into multiple My Social posts.','Long captions are shortened for import; full captions remain in this report.','My Social video import requires FFmpeg/FFprobe and videos up to 90 seconds.']} )]);
  await emit(archive,++part,{posts:entries.length,bytes:archive.size});files=[];entries=[];report=[];bytes=0;
 };
 for(let index=0;index<posts.length;index++){
  if(signal?.aborted)break;
  const post=posts[index],downloaded=[],errors=[];
  onProgress('Downloading post '+(index+1)+' of '+posts.length);
  for(let i=0;i<post.media.length;i++){
   if(signal?.aborted){errors.push('Cancelled');break}
   try{const f=await fetchFile(post.media[i].url,signal);if(f.data.length>MEDIA_LIMIT||!mediaType(f.data))throw Error('Unsupported or oversized file');if(downloaded.reduce((n,x)=>n+x.data.length,0)+f.data.length>350*1024*1024){errors.push('Post exceeds the 350 MiB capture budget; remaining media was not fetched');break}downloaded.push({...f,index:i})}
   catch(e){errors.push('Media '+(i+1)+': '+(e.name==='AbortError'?'Interrupted or timed out':e.message))}
  }
  const id=String(post.code||('post-'+index)).replace(/[^A-Za-z0-9_-]/g,'').slice(0,70)||'post-'+index;
  if(post.incompleteMedia)errors.push('Only a preview or incomplete player/carousel was accessible');
  const complete=!errors.length&&downloaded.length===post.media.length&&downloaded.length>0;
  if(!complete)failed++;
  let group=[],groupBytes=0;
  const commit=async()=>{
   if(!group.length)return;
   if(bytes+groupBytes>partLimit||entries.length>=100||report.length>=100)await flush();
   const media=[];
   for(const f of group){const name=(complete?'media/posts/':'capture/incomplete/')+id+'-'+index+'-'+f.index+'.'+f.ext;files.push({name,data:f.data});bytes+=f.data.length;mediaFiles++;if(['mp4','webm'].includes(f.ext))videos++;else photos++;media.push({uri:name,creation_timestamp:post.created||now,title:importCaption(post.caption)})}
   if(complete){entries.push({kind:post.kind,entry:{title:importCaption(post.caption),creation_timestamp:post.created||now,media}});importable++}
   report.push({permalink:post.permalink,caption:post.caption||'',owner:post.owner||'',created:post.created||now,dateSource:post.created?'original':'capture-time',origin:post.origin,importable:complete,expectedMedia:post.media.length,capturedMedia:group.length,files:media.map(m=>m.uri),errors,altText:group.map(f=>post.media[f.index].alt||''),videoDurations:group.map(f=>post.media[f.index].duration||0),visibleComments:post.comments||''});
   group=[];groupBytes=0;
  };
  for(const f of downloaded){if(group.length>=10||groupBytes+f.data.length>Math.min(partLimit,350*1024*1024))await commit();group.push(f);groupBytes+=f.data.length}
  if(!downloaded.length){if(report.length>=100)await flush();report.push({permalink:post.permalink,importable:false,expectedMedia:post.media.length,capturedMedia:0,errors})}
  await commit();if(signal?.aborted)break;
 }
 await flush();return {parts:part,importable,failed,mediaFiles,photos,videos,cancelled:!!signal?.aborted};
}
