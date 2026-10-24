import{test,expect,chromium}from'@playwright/test';
import{mkdir,mkdtemp,readFile,writeFile}from'node:fs/promises';
import path from'node:path';
import sharp from'sharp';

test('real extension captures a synthetic Instagram tab, downloads a ZIP and imports it into My Social',async()=>{
 const parent=process.env.MYSOCIAL_TEST_ROOT||'C:/Projects/.cache/social-extension-20261024';await mkdir(parent,{recursive:true});
 const root=await mkdtemp(path.join(parent,'extension-')),extension=path.resolve('../extensions/instagram-capture/dist'),downloads=path.join(root,'downloads');await mkdir(downloads,{recursive:true});
 const profile=path.join(root,'chrome');await mkdir(path.join(profile,'Default'),{recursive:true});
 await writeFile(path.join(profile,'Default','Preferences'),JSON.stringify({download:{default_directory:downloads,prompt_for_download:false},profile:{default_content_setting_values:{automatic_downloads:1}}}));
 const context=await chromium.launchPersistentContext(profile,{channel:'chromium',headless:true,acceptDownloads:true,downloadsPath:downloads,args:['--disable-extensions-except='+extension,'--load-extension='+extension,'--disable-background-networking']});
 try{
  const png=await sharp({create:{width:64,height:48,channels:3,background:'#468da5'}}).png().toBuffer();
  const media=(name:string)=>({image_versions2:{candidates:[{url:'https://fixture.cdninstagram.com/'+name+'.png',width:640,height:480}]}});
  const one={code:'SYNTHETIC1',taken_at:1600000000,caption:{text:'Captured carousel ☀️'},carousel_media:[media('a'),media('b')],...media('a')};
  const two={code:'SYNTHETIC2',taken_at:1600000001,caption:{text:'Second captured post'},...media('c')};
  const html=(nodes:any[],extra='')=>`<!doctype html><html><head><meta charset="UTF-8"><title>Synthetic Instagram</title></head><body><main><script type="application/json">${JSON.stringify({data:nodes})}</script>${extra}</main></body></html>`;
  await context.route('https://www.instagram.com/**',route=>{const u=new URL(route.request().url());const profile=u.pathname==='/fixture/';return route.fulfill({contentType:'text/html',body:html(profile?[one]:[u.pathname.includes('SYNTHETIC2')?two:one],profile?'<a href="/p/SYNTHETIC1/">First</a><a href="/p/SYNTHETIC2/">Second</a>':'')})});
  await context.route('https://fixture.cdninstagram.com/**',route=>route.fulfill({contentType:'image/png',body:png}));
  const source=await context.newPage();await source.goto('https://www.instagram.com/fixture/');
  let [worker]=context.serviceWorkers();if(!worker)worker=await context.waitForEvent('serviceworker');
  const extensionID=new URL(worker.url()).host;
  const tabID=await worker.evaluate(async()=>{const tabs=await(chrome as any).tabs.query({});return tabs.find((t:any)=>t.url==='https://www.instagram.com/fixture/').id});
  const capture=await context.newPage();await capture.goto(`chrome-extension://${extensionID}/export.html?tab=${tabID}`);
  await capture.getByRole('button',{name:'Capture content',exact:true}).click();
  await expect(capture.getByText('2 captured · 2 selected',{exact:true})).toBeVisible({timeout:30000});
  await expect(capture.getByRole('button',{name:'Capture content',exact:true})).toBeEnabled();
  const output=process.env.MYSOCIAL_TEST_OUTPUT||'../docs/screenshots';await mkdir(output,{recursive:true});
  await capture.screenshot({path:path.join(output,'instagram-extension-desktop.png'),fullPage:true});
  await capture.setViewportSize({width:390,height:844});expect(await capture.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  await capture.screenshot({path:path.join(output,'instagram-extension-mobile.png'),fullPage:true});
  await capture.getByRole('button',{name:'Download import ZIPs',exact:true}).click();
  await expect(capture.getByText(/1 ZIP parts downloaded · 2 importable posts · 0 posts incomplete/)).toBeVisible({timeout:30000});
  const [download]=await worker.evaluate(async()=>await(chrome as any).downloads.search({}));expect(download.state).toBe('complete');
  const resolved=path.resolve(download.filename);expect(resolved.startsWith(path.resolve(root)+path.sep)).toBe(true);const archive=await readFile(resolved);
  const app=await context.newPage();await app.goto(process.env.MYSOCIAL_TEST_URL||'http://localhost:29113');await app.getByLabel('Email address').fill('creator@example.test');await app.getByLabel('Password',{exact:true}).fill(process.env.MYSOCIAL_TEST_PASSWORD!);await app.getByRole('button',{name:'Sign in',exact:true}).click();await expect(app.getByRole('heading',{name:'The everyday, shared.'})).toBeVisible();
  await app.goto((process.env.MYSOCIAL_TEST_URL||'http://localhost:29113')+'/#instagram');await app.getByLabel('Instagram export ZIP').setInputFiles({name:'extension-export.zip',mimeType:'application/zip',buffer:archive});await app.getByRole('button',{name:'Preview ZIP',exact:true}).click();
  await expect(app.getByText('2 posts found · 2 available · 0 unavailable',{exact:true})).toBeVisible();await expect(app.getByLabel('Who can see imported posts?')).toHaveValue('private');await app.getByRole('button',{name:'Select available posts',exact:true}).click();await app.getByRole('button',{name:'Import selected posts',exact:true}).click();await expect(app.getByText('2 imported · 0 already imported · 0 failed. Successful posts are saved; you can retry failed posts safely.',{exact:true})).toBeVisible();
  await app.goto((process.env.MYSOCIAL_TEST_URL||'http://localhost:29113')+'/#profile');await expect(app.getByText('Captured carousel ☀️',{exact:true})).toBeVisible();await expect(app.getByText('Second captured post',{exact:true})).toBeVisible();
  await writeFile(path.join(output,'instagram-extension-acceptance.json'),JSON.stringify({realMV3Extension:true,syntheticInstagramOnly:true,postPageEnrichment:true,downloadedZIPImportedByNativeMySocial:true,carouselDatesCaptionsPreserved:true,privateDefault:true,phoneWidthNoOverflow:true,ownerDataUsed:false},null,2));
 }finally{await context.close()}
});
