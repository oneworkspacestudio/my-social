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

test('Instagram ZIP preview, selection, private import, duplicates and mobile layout',async({page})=>{
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('/');await page.getByLabel('Email address').fill('creator@example.test');await page.getByLabel('Password',{exact:true}).fill(process.env.MYSOCIAL_TEST_PASSWORD!);await page.getByRole('button',{name:'Sign in',exact:true}).click();
 await expect(page.getByRole('heading',{name:'The everyday, shared.'})).toBeVisible();
 await page.goto('/#settings');await page.getByRole('link',{name:'Import Instagram ZIP',exact:true}).click();
 const photo=await sharp({create:{width:80,height:60,channels:3,background:'#438ca8'}}).png().toBuffer();
 const archive=zip({'your_instagram_activity/content/posts_1.json':Buffer.from(JSON.stringify([{media:[{uri:'media/posts/a.png',creation_timestamp:1600000000,title:'An imported day ☀️'},{uri:'media/posts/b.png',creation_timestamp:1600000000}]},{title:'Missing media example',creation_timestamp:1600000001,media:[{uri:'media/posts/missing.png'}]}])),'media/posts/a.png':photo,'media/posts/b.png':photo,'messages/inbox/ignored.json':Buffer.from('{"message":"excluded"}')});
 await page.getByLabel('Instagram export ZIP').setInputFiles({name:'instagram-synthetic.zip',mimeType:'application/zip',buffer:archive});
 await page.getByRole('button',{name:'Preview ZIP',exact:true}).click();
 await expect(page.getByRole('heading',{name:'Review your import'})).toBeVisible();
 await expect(page.getByText('2 posts found · 1 available · 1 unavailable',{exact:true})).toBeVisible();
 await expect(page.getByLabel('Who can see imported posts?')).toHaveValue('private');
 await expect(page.getByRole('button',{name:'Import selected posts',exact:true})).toBeDisabled();
 await page.getByRole('button',{name:'Select post',exact:true}).click();
 await expect(page.getByText('1 of 100 selected for this batch',{exact:true})).toBeVisible();
 const output=process.env.MYSOCIAL_TEST_OUTPUT||'../docs/screenshots';await mkdir(output,{recursive:true});
 await page.screenshot({path:path.join(output,'instagram-import-desktop.png'),fullPage:true});
 await page.setViewportSize({width:390,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 await page.screenshot({path:path.join(output,'instagram-import-mobile.png'),fullPage:true});
 await page.getByRole('button',{name:'Import selected posts',exact:true}).click();
 await expect(page.getByText('1 imported · 0 already imported · 0 failed. Successful posts are saved; you can retry failed posts safely.',{exact:true})).toBeVisible();
 await page.getByRole('button',{name:'Preview ZIP',exact:true}).click();
 await expect(page.getByText('2 posts found · 0 available · 1 unavailable',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'Already imported',exact:true})).toBeVisible();
 await page.goto('/#profile');await expect(page.getByText('An imported day ☀️',{exact:true})).toBeVisible();
 const imported=page.locator('.card--post').filter({hasText:'An imported day ☀️'});
 await expect(imported.getByRole('button',{name:'Next photo',exact:true})).toBeVisible();
 await imported.getByRole('button',{name:'Next photo',exact:true}).click();await expect(imported.getByText('2 / 2',{exact:true})).toBeVisible();
 await page.reload();await expect(page.getByText('An imported day ☀️',{exact:true})).toHaveCount(1);
 expect(errors).toEqual([]);
});
