import{test}from'node:test';import assert from'node:assert/strict';import{crc32 as nativeCRC}from'node:zlib';
import{exportPosts,safeMediaURL,importCaption,zipBlob}from'../src/archive.mjs';

// Independently inspect standard ZIP records and verify each payload checksum.
export async function unzip(blob){const b=Buffer.from(await blob.arrayBuffer()),out={};let p=0;while(b.readUInt32LE(p)===0x04034b50){const crc=b.readUInt32LE(p+14),size=b.readUInt32LE(p+18),nl=b.readUInt16LE(p+26),el=b.readUInt16LE(p+28),name=b.toString('utf8',p+30,p+30+nl),start=p+30+nl+el,data=b.subarray(start,start+size);assert.equal(nativeCRC(data),crc);assert(!out[name]);out[name]=data;p=start+size}assert.equal(b.readUInt32LE(p),0x02014b50);return out}
const image=()=>new Uint8Array([137,80,78,71,13,10,26,10]);
const post=(code,media=1)=>({code,permalink:'https://www.instagram.com/p/'+code+'/',caption:'A synthetic moment ☀️',created:1600000000,kind:'post',media:Array.from({length:media},(_,i)=>({url:'https://scontent.cdninstagram.com/'+code+i,alt:'Image '+i,duration:0}))});

test('host allowlist rejects credentials, lookalikes, local URLs and non-HTTPS',()=>{
 for(const u of ['http://a.cdninstagram.com/a','https://cdninstagram.com.evil.test/a','https://x@cdninstagram.com/a','https://127.0.0.1/a','https://scontent.fbcdn.net:444/a','data:image/png,x','blob:https://www.instagram.com/x'])assert.equal(safeMediaURL(u),'');
 assert.equal(safeMediaURL('https://scontent-a.fbcdn.net/x?oe=123'),'https://scontent-a.fbcdn.net/x?oe=123');
});
test('self-contained ZIPs preserve dates/order, split carousels and keep full captions in reports',async()=>{
 const p=post('ABC',12);p.caption='🌿'.repeat(1000);const archives=[];
 const result=await exportPosts([p],{now:1700000000,fetchFile:async()=>({data:image(),ext:'png'}),emit:async b=>archives.push(b)});
 assert.equal(result.importable,2);assert.equal(archives.length,1);const files=await unzip(archives[0]),entries=JSON.parse(files['content/posts_1.json']);assert.deepEqual(entries.map(x=>x.media.length),[10,2]);
 assert.equal(entries[0].creation_timestamp,1600000000);assert(entries[0].title.endsWith('…'));assert(Buffer.byteLength(entries[0].title)<=2200);assert.equal(JSON.parse(files['capture/report.json']).posts[0].caption,p.caption);assert(entries.flatMap(x=>x.media).every(m=>files[m.uri]));
 assert.equal(importCaption('a'.repeat(2200)).length,2200);
});
test('failed carousel is not silently imported; recovered media and failures remain in the ZIP',async()=>{
 const archives=[];let n=0;
 const result=await exportPosts([post('BAD',2),post('GOOD')],{fetchFile:async()=>{if(++n===2)throw Error('HTTP 403');return{data:image(),ext:'png'}},emit:async b=>archives.push(b)});
 assert.equal(result.failed,1);const files=await unzip(archives[0]);assert.equal(JSON.parse(files['content/posts_1.json']).length,1);assert(Object.keys(files).some(n=>n.startsWith('capture/incomplete/')));const report=JSON.parse(files['capture/report.json']);assert.equal(report.posts[0].importable,false);assert.match(report.posts[0].errors[0],/403/);
});
test('part boundaries retain matching manifests, stories and reels, and report capture-time fallback',async()=>{
 const archives=[],a=post('AAA'),b=post('BBB'),c=post('CCC');a.kind='story';b.kind='reel';c.created=0;
 await exportPosts([a,b,c],{now:1700000000,partLimit:8,fetchFile:async()=>({data:image(),ext:'png'}),emit:async b=>archives.push(b)});
 assert.equal(archives.length,3);for(const archive of archives){const files=await unzip(archive);for(const name of ['content/posts_1.json','content/reels.json','content/stories.json','capture/report.json'])assert(files[name])}
 const last=JSON.parse((await unzip(archives[2]))['capture/report.json']);assert.equal(last.posts[0].dateSource,'capture-time');assert.equal(last.posts[0].created,1700000000);
});
test('cancelled and malicious archive member paths are safe',async()=>{
 assert.throws(()=>zipBlob([{name:'../escape',data:image()}]),/Unsafe/);const controller=new AbortController();controller.abort();let emitted=false;const result=await exportPosts([post('AAA')],{signal:controller.signal,emit:()=>{emitted=true}});assert(!emitted);assert(result.cancelled);
});

test('preview-only media is recovered without pretending it is a complete upload',async()=>{
 const archives=[],p=post('PREVIEW');p.incompleteMedia=true;
 const result=await exportPosts([p],{fetchFile:async()=>({data:image(),ext:'png'}),emit:async b=>archives.push(b)});
 assert.equal(result.importable,0);assert.equal(result.photos,1);assert.equal(result.mediaFiles,1);assert.equal(result.failed,1);
 const files=await unzip(archives[0]);assert.equal(JSON.parse(files['content/posts_1.json']).length,0);
 const report=JSON.parse(files['capture/report.json']);assert(report.posts[0].files[0].startsWith('capture/incomplete/'));assert(files[report.posts[0].files[0]]);
});

test('reads and downloads one upload before reading the next, saving checkpoints along the way',async()=>{
 const events=[],archives=[];
 async function* read(){for(const code of ['ONE','TWO','THREE']){events.push('read '+code);yield post(code,2)}}
 const result=await exportPosts(read(),{checkpointPosts:2,fetchFile:async url=>{events.push('fetch '+url.split('/').at(-1));return{data:image(),ext:'png'}},emit:async b=>{events.push('save');archives.push(await unzip(b))}});
 assert.deepEqual(events,['read ONE','fetch ONE0','fetch ONE1','read TWO','fetch TWO0','fetch TWO1','save','read THREE','fetch THREE0','fetch THREE1','save']);
 assert.equal(result.photos,6);assert.equal(result.parts,2);
 assert.deepEqual(archives.map(f=>JSON.parse(f['content/posts_1.json']).length),[2,1]);
});

test('stopping a streaming capture keeps completed media without reading more uploads',async()=>{
 const controller=new AbortController(),archives=[];let reads=0;
 async function* read(){reads++;yield post('DONE');if(!controller.signal.aborted){reads++;yield post('LATER')}}
 const result=await exportPosts(read(),{signal:controller.signal,fetchFile:async()=>{controller.abort();return{data:image(),ext:'png'}},emit:async b=>archives.push(await unzip(b))});
 assert.equal(reads,1);assert(result.cancelled);assert.equal(result.photos,1);assert.equal(archives.length,1);assert.equal(JSON.parse(archives[0]['content/posts_1.json']).length,1);
});

test('a reader failure flushes completed uploads before reporting the error',async()=>{
 const archives=[];async function* read(){yield post('DONE');throw Error('Page closed')}
 await assert.rejects(exportPosts(read(),{fetchFile:async()=>({data:image(),ext:'png'}),emit:async b=>archives.push(await unzip(b))}),/Page closed/);
 assert.equal(archives.length,1);assert.equal(JSON.parse(archives[0]['content/posts_1.json']).length,1);
});
