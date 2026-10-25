import{test,expect}from'@playwright/test';
import sharp from'sharp';
import{crc32}from'node:zlib';
import{mkdir}from'node:fs/promises';
import path from'node:path';

// Small uncompressed ZIP fixture: no personal export and no new dependency.
function zip(files:Record<string,Buffer>){
 const local:Buffer[]=[],central:Buffer[]=[];let offset=0;
 for(const[name,data]of Object.entries(files)){
  const filename=Buffer.from(name),crc=crc32(data),header=Buffer.alloc(30);
  header.writeUInt32LE(0x04034b50);header.writeUInt16LE(20,4);header.writeUInt32LE(crc,14);header.writeUInt32LE(data.length,18);header.writeUInt32LE(data.length,22);header.writeUInt16LE(filename.length,26);
  local.push(header,filename,data);
  const record=Buffer.alloc(46);record.writeUInt32LE(0x02014b50);record.writeUInt16LE(20,4);record.writeUInt16LE(20,6);record.writeUInt32LE(crc,16);record.writeUInt32LE(data.length,20);record.writeUInt32LE(data.length,24);record.writeUInt16LE(filename.length,28);record.writeUInt32LE(offset,42);central.push(record,filename);offset+=header.length+filename.length+data.length;
 }
 const directory=Buffer.concat(central),end=Buffer.alloc(22);end.writeUInt32LE(0x06054b50);end.writeUInt16LE(Object.keys(files).length,8);end.writeUInt16LE(Object.keys(files).length,10);end.writeUInt32LE(directory.length,12);end.writeUInt32LE(offset,16);return Buffer.concat([...local,directory,end]);
}

test('Instagram ZIP live progress, five-post sample, all-post private import, duplicates and mobile layout',async({page})=>{
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('/');await page.getByLabel('Email address').fill('creator@example.test');await page.getByLabel('Password',{exact:true}).fill(process.env.MYSOCIAL_TEST_PASSWORD!);await page.getByRole('button',{name:'Sign in',exact:true}).click();
 await expect(page.getByRole('heading',{name:'The everyday, shared.'})).toBeVisible();
 await page.goto('/#settings');await page.getByRole('link',{name:'Import Instagram ZIP',exact:true}).click();
 const photo=await sharp({create:{width:80,height:60,channels:3,background:'#438ca8'}}).png().toBuffer();
 const archive=zip({'your_instagram_activity/content/posts_1.json':Buffer.from(JSON.stringify([{media:[{uri:'media/posts/a.png',creation_timestamp:1600000000,title:'An imported day ☀️'},{uri:'media/posts/b.png',creation_timestamp:1600000000}]},...Array.from({length:6},(_,i)=>({title:'Hidden sample post '+(i+2),creation_timestamp:1600000010+i,media:[{uri:'media/posts/a.png'}]})),{title:'Missing media example',creation_timestamp:1600000100,media:[{uri:'media/posts/missing.png'}]}])),'media/posts/a.png':photo,'media/posts/b.png':photo,'messages/inbox/ignored.json':Buffer.from('{"message":"excluded"}')});
 await page.getByLabel('Instagram export ZIP').setInputFiles({name:'instagram-synthetic.zip',mimeType:'application/zip',buffer:archive});
 const network=await page.context().newCDPSession(page);await network.send('Network.enable');await network.send('Network.emulateNetworkConditions',{offline:false,latency:40,downloadThroughput:2048,uploadThroughput:4096});
 await page.getByRole('button',{name:'Upload and preview ZIP',exact:true}).click();
 await expect(page.getByRole('progressbar')).toBeVisible();
 await expect(page.getByText('Your ZIP is saved. Processing continues on the server if you leave this page. Return here to see progress and results.',{exact:true})).toBeVisible({timeout:15000});
 await expect(page.getByText(/Elapsed \d+:\d+/)).toBeVisible();
 await expect(page.getByRole('heading',{name:'Review your import'})).toBeVisible();
 await expect(page.getByText('8 posts found · 7 available · 1 unavailable',{exact:true})).toBeVisible();
 await expect(page.getByLabel('Who can see imported posts?')).toHaveValue('private');
 await expect(page.locator('.card--import-item')).toHaveCount(5);
 await expect(page.getByText('Showing 5 of 8 posts. Import includes all 7 available posts.',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'Import 7 posts',exact:true})).toBeEnabled();
 await network.send('Network.emulateNetworkConditions',{offline:false,latency:0,downloadThroughput:-1,uploadThroughput:-1});
 const output=process.env.MYSOCIAL_TEST_OUTPUT||'../docs/screenshots';await mkdir(output,{recursive:true});
 await page.screenshot({path:path.join(output,'instagram-import-desktop.png'),fullPage:true});
 await page.setViewportSize({width:390,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 await page.screenshot({path:path.join(output,'instagram-import-mobile.png'),fullPage:true});
 await page.getByRole('button',{name:'Import 7 posts',exact:true}).click();
 await expect(page.getByText('7 imported · 0 already imported · 0 failed. Successful posts are saved; you can retry failed posts safely.',{exact:true})).toBeVisible();
 await page.getByRole('button',{name:'Upload and preview ZIP',exact:true}).click();
 await expect(page.getByText('8 posts found · 0 available · 1 unavailable',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'Import 0 posts',exact:true})).toBeDisabled();
 await expect(page.getByText('Hidden sample post 7',{exact:true})).toHaveCount(0);
 await page.getByRole('button',{name:'Cancel import',exact:true}).click();await page.goto('/#profile');await expect(page.getByText('An imported day ☀️',{exact:true})).toBeVisible();
 const imported=page.locator('.card--post').filter({hasText:'An imported day ☀️'});
 await expect(imported.getByRole('button',{name:'Next photo',exact:true})).toBeVisible();
 await imported.getByRole('button',{name:'Next photo',exact:true}).click();await expect(imported.getByText('2 / 2',{exact:true})).toBeVisible();
 await page.reload();await expect(page.getByText('An imported day ☀️',{exact:true})).toHaveCount(1);
 expect(errors).toEqual([]);
});

test('One Import click includes 101 available posts beyond the five-row sample',async({page})=>{
 test.setTimeout(120000);await page.goto('/');await page.getByLabel('Email address').fill('creator@example.test');await page.getByLabel('Password',{exact:true}).fill(process.env.MYSOCIAL_TEST_PASSWORD!);await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.getByRole('heading',{name:'The everyday, shared.'})).toBeVisible();await page.goto('/#instagram');
 const photo=await sharp({create:{width:32,height:24,channels:3,background:'#24658a'}}).png().toBuffer();const archive=zip({'content/posts_1.json':Buffer.from(JSON.stringify(Array.from({length:101},(_,i)=>({title:'Bulk import '+(i+1),creation_timestamp:1590000000+i,media:[{uri:'media/posts/a.png'}]})))),'media/posts/a.png':photo});
 let imports=0,uploads=0;page.on('request',r=>{if(/\/jobs\/[^/]+\/import$/.test(r.url()))imports++;if(r.url().includes('/chunks?offset='))uploads++});
 await page.getByLabel('Instagram export ZIP').setInputFiles({name:'bulk-synthetic.zip',mimeType:'application/zip',buffer:archive});await page.getByRole('button',{name:'Upload and preview ZIP',exact:true}).click();await expect(page.getByText('101 posts found · 101 available · 0 unavailable',{exact:true})).toBeVisible();await expect(page.locator('.card--import-item')).toHaveCount(5);const uploaded=uploads;await page.getByRole('button',{name:'Import 101 posts',exact:true}).click();await expect(page.getByText('101 imported · 0 already imported · 0 failed. Successful posts are saved; you can retry failed posts safely.',{exact:true})).toBeVisible({timeout:90000});expect(imports).toBe(1);expect(uploads).toBe(uploaded);
 await page.getByRole('button',{name:'Upload and preview ZIP',exact:true}).click();await expect(page.getByText('101 posts found · 0 available · 0 unavailable',{exact:true})).toBeVisible();await expect(page.locator('.card--import-item')).toHaveCount(5);await page.getByRole('button',{name:'Cancel import',exact:true}).click();
});

test('Closing the tab resumes saved upload chunks and server import completes with no page open',async({page,context})=>{
 await page.goto('/');await page.getByLabel('Email address').fill('creator@example.test');await page.getByLabel('Password',{exact:true}).fill(process.env.MYSOCIAL_TEST_PASSWORD!);await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.getByRole('heading',{name:'The everyday, shared.'})).toBeVisible();await page.goto('/#instagram');
 const photo=await sharp({create:{width:40,height:30,channels:3,background:'#518e46'}}).png().toBuffer();const archive=zip({'content/posts_1.json':Buffer.from(JSON.stringify([{title:'Continued while the tab was closed',creation_timestamp:1580000000,media:[{uri:'media/posts/a.png'}]}])),'media/posts/a.png':photo,'padding.bin':Buffer.alloc(6*1024*1024)});
 const input={name:'resume-closed-tab.zip',mimeType:'application/zip',buffer:archive};await page.getByLabel('Instagram export ZIP').setInputFiles(input);const network=await context.newCDPSession(page);await network.send('Network.enable');await network.send('Network.emulateNetworkConditions',{offline:false,latency:20,downloadThroughput:-1,uploadThroughput:1048576});
 const first=page.waitForResponse(r=>r.url().includes('/chunks?offset=0')&&r.status()===200);await page.getByRole('button',{name:'Upload and preview ZIP',exact:true}).click();await first;await page.close();
 const jobs=await(await context.request.get('/api/imports/instagram/jobs')).json();const job=jobs.jobs.find((j:any)=>j.filename===input.name);expect(job.received).toBeGreaterThanOrEqual(2*1048576);expect(job.received).toBeLessThan(archive.length);expect(job.state).toBe('uploading');
 const resumed=await context.newPage();await resumed.goto('/#instagram');await expect(resumed.getByText(/Upload paused at/)).toBeVisible();await resumed.getByLabel('Instagram export ZIP').setInputFiles(input);const offsets:number[]=[];resumed.on('request',r=>{if(r.url().includes('/chunks?offset='))offsets.push(Number(new URL(r.url()).searchParams.get('offset')))});await resumed.getByRole('button',{name:'Resume ZIP upload',exact:true}).click();await expect(resumed.getByText('1 posts found · 1 available · 0 unavailable',{exact:true})).toBeVisible();expect(offsets.length).toBeGreaterThan(0);expect(offsets.every(n=>n>=job.received)).toBe(true);
 const queued=resumed.waitForResponse(r=>r.url().endsWith('/'+job.id+'/import')&&r.status()===202);await resumed.getByRole('button',{name:'Import 1 post',exact:true}).click();await queued;await resumed.close();
 await expect.poll(async()=>{const result=await(await context.request.get('/api/imports/instagram/jobs/'+job.id)).json();return result.state},{timeout:30000}).toBe('completed');
 const returned=await context.newPage();await returned.goto('/#instagram');await expect(returned.getByText('1 imported · 0 already imported · 0 failed. Successful posts are saved; you can retry failed posts safely.',{exact:true})).toBeVisible();await expect(returned.locator('.card--import-item')).toHaveCount(1);await returned.close();
});
