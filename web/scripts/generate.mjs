import {parse,compileReact} from '@codedia/parser';
import {readFile,readdir,mkdir,writeFile} from 'node:fs/promises';
const sources=new URL('../src/screens/',import.meta.url),dest=new URL('../src/generated/',import.meta.url);
await mkdir(dest,{recursive:true});
for(const name of (await readdir(sources)).filter(x=>x.endsWith('.aui'))){await writeFile(new URL(name.replace(/\.aui$/,'.tsx'),dest),`// Generated from screens/${name}; edit the AUI source.\n${compileReact(parse(await readFile(new URL(name,sources),'utf8')))}`)}
