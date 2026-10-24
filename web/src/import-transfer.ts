// XHR exposes upload bytes and incremental response text without buffering the
// entire import on the server. Authentication headers still come from request().
export function uploadRequest(url:string,options:any,headers:Record<string,string>):Promise<Response>{
 return new Promise((resolve,reject)=>{
  const xhr=new XMLHttpRequest();let settled=false,offset=0,pending='',result:any,problem:any;
  const abort=()=>xhr.abort();
  const finish=(error?:any,response?:Response)=>{if(settled)return;settled=true;options.signal?.removeEventListener('abort',abort);error?reject(error):resolve(response!)};
  const streaming=()=>xhr.getResponseHeader('Content-Type')?.includes('application/x-ndjson');
  const consume=(last=false)=>{
   if(!streaming())return;
   pending+=xhr.responseText.slice(offset);offset=xhr.responseText.length;
   const lines=pending.split('\n');pending=lines.pop()||'';if(last&&pending.trim()){lines.push(pending);pending=''}
   for(const line of lines){if(!line.trim())continue;const packet=JSON.parse(line);if(packet.type==='progress')options.onImportProgress?.(packet);else if(packet.type==='result')result=packet.result;else if(packet.type==='error')problem=Object.assign(new Error(packet.error||'Import failed'),{status:packet.status})}
  };
  xhr.open(options.method||'POST',url);xhr.withCredentials=true;
  for(const[name,value]of Object.entries(headers))xhr.setRequestHeader(name,value);
  xhr.upload.onprogress=e=>options.onUploadProgress?.({loaded:e.loaded,total:e.lengthComputable?e.total:0});
  xhr.upload.onload=()=>options.onUploaded?.();
  xhr.onprogress=()=>{try{consume()}catch{finish(Error('Import progress could not be read. Preview and retry this ZIP safely.'));xhr.abort()}};
  xhr.onload=()=>{try{
   consume(true);if(problem){finish(problem);return}
   if(streaming()&&result===undefined){finish(Error('Import connection ended before completion. Preview and retry this ZIP safely.'));return}
   finish(undefined,new Response(streaming()?JSON.stringify(result):xhr.responseText,{status:xhr.status,headers:{'Content-Type':streaming()?'application/json':xhr.getResponseHeader('Content-Type')||'text/plain','X-Workspace-Account':xhr.getResponseHeader('X-Workspace-Account')||''}}));
  }catch{finish(Error('Import response could not be read. Preview and retry this ZIP safely.'))}};
  xhr.onerror=()=>finish(Error('Connection lost. Preview and retry this ZIP; completed posts will be skipped.'));
  xhr.onabort=()=>finish(new DOMException('Import stopped','AbortError'));
  options.signal?.addEventListener('abort',abort,{once:true});
  if(options.signal?.aborted){finish(new DOMException('Import stopped','AbortError'));return}
  xhr.send(options.body);
 });
}
