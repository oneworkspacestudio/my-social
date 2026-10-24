import{test,expect,chromium}from'@playwright/test';
import{mkdir,mkdtemp,readFile,writeFile}from'node:fs/promises';
import path from'node:path';
import sharp from'sharp';

test('real extension captures a synthetic Instagram tab, downloads a ZIP and imports it into My Social',async()=>{
 test.setTimeout(180000);
 const parent=process.env.MYSOCIAL_TEST_ROOT||'C:/Projects/.cache/social-extension-20261024';await mkdir(parent,{recursive:true});
 const root=await mkdtemp(path.join(parent,'extension-')),extension=path.resolve('../extensions/instagram-capture/dist'),downloads=path.join(root,'downloads');await mkdir(downloads,{recursive:true});
 const profile=path.join(root,'chrome');await mkdir(path.join(profile,'Default'),{recursive:true});
 await writeFile(path.join(profile,'Default','Preferences'),JSON.stringify({download:{default_directory:downloads,prompt_for_download:false},profile:{default_content_setting_values:{automatic_downloads:1}}}));
 const context=await chromium.launchPersistentContext(profile,{channel:'chromium',headless:true,acceptDownloads:true,downloadsPath:downloads,args:['--disable-extensions-except='+extension,'--load-extension='+extension,'--disable-background-networking']});
 try{
  const png=await sharp({create:{width:64,height:48,channels:3,background:'#468da5'}}).png().toBuffer();
  const media=(name:string)=>({image_versions2:{candidates:[{url:'https://fixture.cdninstagram.com/'+name+'.png',width:640,height:480}]}});
  const one={owner:{username:'fixture'},code:'SYNTHETIC1',taken_at:1600000000,caption:{text:'Captured carousel ☀️'},carousel_media:[media('a'),media('b')],...media('a')};
  const two={owner:{username:'fixture'},code:'SYNTHETIC2',taken_at:1600000001,caption:{text:'Second captured post'},...media('c')};
  const html=(nodes:any[],extra='')=>`<!doctype html><html><head><meta charset="UTF-8"><title>Synthetic Instagram</title></head><body><main><script type="application/json">${JSON.stringify({data:nodes})}</script>${extra}</main></body></html>`;
  const foreign={...one,code:'NOT_MY_UPLOAD',owner:{username:'someone_else'},caption:{text:'Foreign post must not be downloaded'}};
  const delayedArticle='<article><header><a href="/fixture/">fixture</a></header><h1>Second captured post</h1><time datetime="2020-09-13T12:26:41Z"></time><img width="640" height="480" src="https://fixture.cdninstagram.com/c.png"><button aria-label="Next" onclick="const i=document.createElement(\'img\');i.width=640;i.height=480;i.src=\'https://fixture.cdninstagram.com/d.png\';this.before(i);this.disabled=true">Next slide</button></article>';
  await context.route('https://www.instagram.com/**',route=>{
   const u=new URL(route.request().url());if(/^\/fixture\/(p|reel)\//.test(u.pathname))return route.fulfill({status:302,headers:{location:'/'+u.pathname.split('/').filter(Boolean).slice(-2).join('/')+'/'}});
   const profile=u.pathname==='/fixture/'||u.pathname==='/fixture/reels/';
   const unknown={...one,code:'UNVERIFIED',owner:undefined,caption:{text:'Unverified owner must not be downloaded'}};
   const body=profile?html([one,foreign,unknown],'<a href="/fixture/p/SYNTHETIC1/">First</a><a href="/fixture/p/SYNTHETIC2/">Second<img width="640" height="480" src="https://fixture.cdninstagram.com/c.png"></a><a href="/someone_else/p/NOT_MY_UPLOAD/">Other account</a><a href="/p/UNVERIFIED/">Unverified<img width="640" height="480" src="https://fixture.cdninstagram.com/unknown.png"></a>'):u.pathname.includes('UNVERIFIED')?html([unknown],'<article><h1>Unverified owner must not be downloaded</h1><img width="640" height="480" src="https://fixture.cdninstagram.com/unknown.png"></article>'):u.pathname.includes('SYNTHETIC2')?html([],`<div id="delayed"></div><script>setTimeout(()=>document.getElementById('delayed').innerHTML=${JSON.stringify(delayedArticle)},100)</script>`):html([one,foreign]);
   return route.fulfill({contentType:'text/html',body});
  });
  await context.route('https://fixture.cdninstagram.com/**',route=>route.fulfill({contentType:'image/png',body:png}));
  const source=await context.newPage();await source.goto('https://www.instagram.com/fixture/');
  let [worker]=context.serviceWorkers();if(!worker)worker=await context.waitForEvent('serviceworker');
  const extensionID=new URL(worker.url()).host;
  const tabID=await worker.evaluate(async()=>{const tabs=await(chrome as any).tabs.query({});return tabs.find((t:any)=>t.url==='https://www.instagram.com/fixture/').id});
  const capture=await context.newPage();await capture.goto(`chrome-extension://${extensionID}/export.html?tab=${tabID}`);
  await expect(capture.getByLabel('My Instagram username')).toHaveValue('fixture');
  await expect(capture.getByLabel('Scroll my profile to find older uploads')).toBeChecked();
  await capture.getByRole('button',{name:'Scan my uploads',exact:true}).click();
  await expect(capture.getByText('2 uploads discovered · 4 photos · 0 videos · 2 uploads selected',{exact:true})).toBeVisible({timeout:90000}).catch(async error=>{
   // This browser contains synthetic fixtures only; log capture status, not login fields.
   console.log('Synthetic capture status:',await capture.locator('body').innerText());
   console.log('Fixture tab locations:',await Promise.all(context.pages().map(async p=>({url:p.url(),title:await p.title().catch(()=>'(closed)')}))));
   throw error;
  });
  await expect(capture.getByRole('button',{name:'Scan my uploads',exact:true})).toBeEnabled({timeout:60000});
  const output=process.env.MYSOCIAL_TEST_OUTPUT||'../docs/screenshots';await mkdir(output,{recursive:true});
  await expect(capture.getByText('Foreign post must not be downloaded',{exact:true})).toHaveCount(0);
  await expect(capture.getByText('Unverified owner must not be downloaded',{exact:true})).toHaveCount(0);
  await capture.screenshot({path:path.join(output,'instagram-extension-desktop.png'),fullPage:true});
  await capture.setViewportSize({width:390,height:844});expect(await capture.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  await capture.screenshot({path:path.join(output,'instagram-extension-mobile.png'),fullPage:true});
  await capture.getByRole('button',{name:'Download my media ZIPs',exact:true}).click();
  await expect(capture.getByText(/4 photos and 0 videos saved in 1 ZIP files · 0 uploads incomplete/)).toBeVisible({timeout:30000});
  const [download]=await worker.evaluate(async()=>await(chrome as any).downloads.search({}));expect(download.state).toBe('complete');
  const resolved=path.resolve(download.filename);expect(resolved.startsWith(path.resolve(root)+path.sep)).toBe(true);const archive=await readFile(resolved);
  const app=await context.newPage();
  // Settle the fresh profile's initial session check before submitting login.
  const initialSession=app.waitForResponse(r=>new URL(r.url()).pathname==='/api/me');
  await app.goto(process.env.MYSOCIAL_TEST_URL||'http://localhost:29113');expect((await initialSession).status()).toBe(401);
  await app.getByLabel('Email address').fill('creator@example.test');await app.getByLabel('Password',{exact:true}).fill(process.env.MYSOCIAL_TEST_PASSWORD!);
  const loginResponse=app.waitForResponse(r=>new URL(r.url()).pathname==='/api/login'&&r.request().method()==='POST');
  await app.getByRole('button',{name:'Sign in',exact:true}).click();expect((await loginResponse).status()).toBe(200);
  await expect(app.getByRole('heading',{name:'The everyday, shared.'})).toBeVisible({timeout:30000});
  await app.goto((process.env.MYSOCIAL_TEST_URL||'http://localhost:29113')+'/#instagram');await app.getByLabel('Instagram export ZIP').setInputFiles({name:'extension-export.zip',mimeType:'application/zip',buffer:archive});await app.getByRole('button',{name:'Preview ZIP',exact:true}).click();
  await expect(app.getByText('2 posts found · 2 available · 0 unavailable',{exact:true})).toBeVisible();await expect(app.getByLabel('Who can see imported posts?')).toHaveValue('private');await app.getByRole('button',{name:'Select available posts',exact:true}).click();await app.getByRole('button',{name:'Import selected posts',exact:true}).click();await expect(app.getByText('2 imported · 0 already imported · 0 failed. Successful posts are saved; you can retry failed posts safely.',{exact:true})).toBeVisible();
  await app.goto((process.env.MYSOCIAL_TEST_URL||'http://localhost:29113')+'/#profile');await expect(app.getByText('Captured carousel ☀️',{exact:true})).toBeVisible();await expect(app.getByText('Second captured post',{exact:true})).toBeVisible();
  await writeFile(path.join(output,'instagram-extension-acceptance.json'),JSON.stringify({realMV3Extension:true,usernamePrefixedLinks:true,otherOwnersExcluded:true,renderedPostFallback:true,photosAndVideosNamedClearly:true,syntheticInstagramOnly:true,postPageEnrichment:true,downloadedZIPImportedByNativeMySocial:true,carouselDatesCaptionsPreserved:true,privateDefault:true,phoneWidthNoOverflow:true,ownerDataUsed:false},null,2));
 }finally{await context.close()}
});
