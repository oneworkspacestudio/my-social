export const aspects:Record<string,[number,number]>={'1:1':[1080,1080],'4:5':[1080,1350],'9:16':[1080,1920],'16:9':[1440,810]};
export const defaults={aspect:'4:5',cropZoom:1,straighten:0,rotation:0,flip:false,brightness:0,contrast:0,saturation:0,warmth:0,blur:0,filter:'Original',fontSize:56,font:'Segoe UI',color:'#ffffff',brushWidth:8,textBackground:true,overlayText:''};
export type Edits=typeof defaults;
export type Item={id:string;file:File;url:string;video:boolean;edits:Edits;scene:any;media?:string;start:number;end:number;speed:number};
export function newID(){return Array.from(crypto.getRandomValues(new Uint8Array(24)),v=>v.toString(16).padStart(2,'0')).join('')}
export function newItem(file:File):Item{return{id:newID(),file,url:URL.createObjectURL(file),video:file.type.startsWith('video/'),edits:{...defaults},scene:null,start:0,end:10,speed:1}}
export function dimensions(aspect:string){return aspects[aspect]||aspects['4:5']}
export function cropTransform(iw:number,ih:number,width:number,height:number,zoom:number,rotation:number){const rotated=Math.abs(rotation%180)===90;const scale=Math.max(width/(rotated?ih:iw),height/(rotated?iw:ih))*Math.max(1,Math.min(3,zoom));return{left:width/2,top:height/2,scaleX:scale,scaleY:scale,angle:rotation,originX:'center' as const,originY:'center' as const}}
export function safeURL(s:string){try{const u=new URL(s);return u.protocol==='https:'&&!u.username&&!u.password}catch{return false}}
