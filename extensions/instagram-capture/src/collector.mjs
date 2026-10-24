// This entire function is serialized into the chosen tab by chrome.scripting.
// Keep helpers inside it: no page globals, private API, cookies or eval needed.
export async function collectPage(options = {}) {
  const allowedPage = u => u.protocol === 'https:' && ['www.instagram.com','instagram.com'].includes(u.hostname) && !/^\/(direct|accounts|challenge|emails|privacy|oauth)(\/|$)/.test(u.pathname);
  const source = new URL(location.href);
  if (!allowedPage(source)) throw Error('Open an Instagram profile, feed or post tab. Messages and account/security pages are excluded.');
  const mediaURL = value => {
    try { const u=new URL(value);return u.protocol==='https:'&&!u.username&&!u.password&&(!u.port||u.port==='443')&&['cdninstagram.com','fbcdn.net'].some(h=>u.hostname===h||u.hostname.endsWith('.'+h)) ? u.href : ''; } catch{return ''}
  };
  const postURL = value => {
    try { const u=new URL(value,source);return allowedPage(u)&&/^\/(p|reel|tv)\/[A-Za-z0-9_-]{3,64}\/?$/.test(u.pathname) ? u.origin+u.pathname.replace(/\/$/,'')+'/' : ''; } catch{return ''}
  };
  let doc=document,url=source.href.split(/[?#]/)[0];
  if(options.url){
    const target=postURL(options.url);if(!target)throw Error('Invalid Instagram post link');
    const controller=new AbortController(),timer=setTimeout(()=>controller.abort(),20000);
    try {
      const response=await fetch(target,{credentials:'same-origin',signal:controller.signal,redirect:'error'});
      if(!response.ok)throw Error('Post page returned HTTP '+response.status);
      const reader=response.body.getReader();let size=0;const chunks=[];
      while(true){const{value,done}=await reader.read();if(done)break;size+=value.length;if(size>12*1024*1024){await reader.cancel();throw Error('Post page exceeds the capture limit')}chunks.push(value)}
      doc=new DOMParser().parseFromString(await new Blob(chunks).text(),'text/html');url=target;
    } finally {clearTimeout(timer)}
  }
  const posts=new Map(),warnings=[],links=new Set();
  const cleanText=x=>typeof x==='string'?x.slice(0,16000):'';
  const timestamp=x=>{const n=Number(x);return Number.isFinite(n)&&n>0&&n<=Date.now()/1000?Math.floor(n):0};
  const best=xs=>{const x=(Array.isArray(xs)?xs:[]).filter(x=>mediaURL(x?.url||x?.src)).sort((a,b)=>(b.width||b.config_width||0)*(b.height||b.config_height||0)-(a.width||a.config_width||0)*(a.height||a.config_height||0))[0];return x?.url||x?.src||''};
  const mediaOf=n=>{
    const video=best(n.video_versions)||mediaURL(n.video_url)||(/VideoObject/.test(String(n['@type']))?mediaURL(n.contentUrl):'');
    const image=best(n.image_versions2?.candidates)||mediaURL(n.display_url)||mediaURL(n.display_src)||best(n.display_resources)||mediaURL(n.image)||mediaURL(n.contentUrl);
    const u=mediaURL(video||image);if(!u)return null;
    return {url:u,type:video?'video':'image',alt:cleanText(n.accessibility_caption||n.caption?.text),duration:Number(n.video_duration)||0,id:String(n.pk||n.id||new URL(u).pathname).slice(0,300)};
  };
  const merge=p=>{
    if(!p?.permalink||!p.media?.length)return;
    const old=posts.get(p.permalink);
    if(!old){posts.set(p.permalink,p);return}
    if(p.media.length>old.media.length)old.media=p.media;
    if(!old.caption&&p.caption)old.caption=p.caption;
    if(!old.created&&p.created)old.created=p.created;
    if(!old.owner&&p.owner)old.owner=p.owner;
    if(!old.comments&&p.comments)old.comments=p.comments;
  };
  const visit=(n)=>{
    if(!n||typeof n!=='object'||Array.isArray(n))return;
    const story=source.pathname.startsWith('/stories/')&&n.taken_at&&(n.pk||n.id)&&n.image_versions2;
    const code=n.shortcode||n.code||(story?'story-'+String(n.pk||n.id).replace(/[^A-Za-z0-9_-]/g,''):'');
    const ownMedia=mediaOf(n);
    if(typeof code!=='string'||!/^[A-Za-z0-9_-]{3,64}$/.test(code)||(!ownMedia&&!n.carousel_media&&!n.edge_sidecar_to_children))return;
    const children=n.carousel_media||n.edge_sidecar_to_children?.edges?.map(x=>x.node)||[n];
    const media=children.slice(0,100).map(mediaOf).filter(Boolean);
    const reel=n.product_type==='clips';
    merge({permalink:story?source.origin+'/stories/'+source.pathname.split('/')[2]+'/'+String(n.pk||n.id).split('_')[0].replace(/[^0-9]/g,'')+'/':'https://www.instagram.com/'+(reel?'reel':'p')+'/'+code+'/',code,kind:story?'story':reel&&media.length===1&&media[0].type==='video'?'reel':'post',caption:cleanText(n.caption?.text||n.edge_media_to_caption?.edges?.[0]?.node?.text||''),created:timestamp(n.taken_at||n.taken_at_timestamp),owner:cleanText(n.owner?.username||n.user?.username),media,comments:'',origin:'page-data'});
  };
  let examined=0;
  const walk=root=>{
    const stack=[root];while(stack.length&&examined<150000&&posts.size<1000){const x=stack.pop();examined++;if(!x||typeof x!=='object')continue;visit(x);for(const key of Object.keys(x)){const v=x[key];if(v&&typeof v==='object')stack.push(v)}}
  };
  for(const script of Array.from(doc.querySelectorAll('script[type="application/json"],script[type="application/ld+json"]')).slice(0,200)){
    if(script.textContent.length>8*1024*1024){warnings.push('A large page-data block was skipped');continue}
    try {
      const value=JSON.parse(script.textContent);walk(value);
      // JSON-LD is useful when Instagram's hydrated data is absent.
      for(const n of [value,...(value?.['@graph']||[])]){
        const permalink=postURL(n?.url||n?.mainEntityOfPage?.['@id']||url),type=String(n?.['@type']||'');
        const m=mediaOf(n);
        if(permalink&&m&&/ImageObject|VideoObject|SocialMediaPosting/.test(type))merge({permalink,code:permalink.split('/').at(-2),kind:'post',caption:cleanText(n.caption||n.description),created:timestamp(Date.parse(n.datePublished||n.uploadDate||'')/1000),owner:cleanText(n.author?.alternateName||n.author?.name),media:[m],comments:'',origin:'structured-data'});
      }
    }catch{/* Non-JSON data is ignored; never evaluate scripts. */}
  }
  for(const article of doc.querySelectorAll('article')){
    const anchor=Array.from(article.querySelectorAll('a[href]')).find(a=>postURL(a.getAttribute('href')));
    const permalink=postURL(anchor?.getAttribute('href')||url);if(!permalink)continue;
    const time=article.querySelector('time[datetime]'),media=[],seen=new Set();
    for(const el of article.querySelectorAll('video,img')){
      if(el.closest('header')||el.tagName==='IMG'&&(/profile picture|avatar/i.test(el.alt||'')||el.width&&el.width<100))continue;
      const src=mediaURL(el.currentSrc||el.src||el.querySelector?.('source')?.src);if(!src||seen.has(new URL(src).pathname))continue;
      seen.add(new URL(src).pathname);media.push({url:src,type:el.tagName==='VIDEO'?'video':'image',alt:cleanText(el.alt),duration:el.tagName==='VIDEO'&&Number.isFinite(el.duration)?el.duration:0,id:new URL(src).pathname});
    }
    const caption=cleanText(article.querySelector('h1')?.textContent||article.querySelector('ul li span[dir="auto"]')?.textContent||'');
    merge({permalink,code:permalink.split('/').at(-2),kind:'post',caption,created:timestamp(Date.parse(time?.dateTime||'')/1000),owner:'',media,comments:cleanText(article.querySelector('ul')?.innerText||article.querySelector('ul')?.textContent||''),origin:'visible-page'});
  }
  const canonical=postURL(doc.querySelector('link[rel="canonical"]')?.href||url);
  if(canonical&&!posts.has(canonical)){
    const video=mediaURL(doc.querySelector('meta[property="og:video"],meta[property="og:video:secure_url"]')?.content);
    const image=mediaURL(doc.querySelector('meta[property="og:image"]')?.content);
    if(video||image){const time=doc.querySelector('time[datetime]')?.getAttribute('datetime')||doc.querySelector('meta[property="article:published_time"]')?.content;
      merge({permalink:canonical,code:canonical.split('/').at(-2),kind:'post',caption:cleanText(doc.querySelector('meta[property="og:description"],meta[name="description"]')?.content),created:timestamp(Date.parse(time||'')/1000),owner:'',media:[{url:video||image,type:video?'video':'image',alt:'',duration:0,id:new URL(video||image).pathname}],comments:'',origin:'page-preview'});
      warnings.push('A page preview was captured; its caption or carousel may be incomplete');
    }
  }
  for(const a of doc.querySelectorAll('a[href]')){const p=postURL(a.getAttribute('href'));if(p&&links.size<1000)links.add(p)}
  for(const p of posts.values())links.add(p.permalink);
  const profile={url:source.origin+source.pathname,title:cleanText(doc.querySelector('meta[property="og:title"]')?.content||doc.title),description:cleanText(doc.querySelector('meta[property="og:description"]')?.content),visibleHeader:cleanText(doc.querySelector('main header')?.textContent)};
  if(options.scroll&&!options.url)window.scrollTo({top:Math.max(document.body.scrollHeight,document.documentElement.scrollHeight),behavior:'instant'});
  return {posts:Array.from(posts.values()),links:Array.from(links),warnings:[...new Set(warnings)],profile,height:document.documentElement.scrollHeight};
}
