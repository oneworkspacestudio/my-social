"""Inventory resolved dependency licences without embedding local paths or secrets."""
import argparse,json,subprocess
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--go',required=True);p.add_argument('--out',required=True);a=p.parse_args()
root=Path(__file__).resolve().parents[1];out=Path(a.out);out.mkdir(parents=True,exist_ok=True)
text=subprocess.check_output([a.go,'list','-m','-json','all'],cwd=root).decode();modules=[];decoder=json.JSONDecoder()
while text.strip():
 text=text.lstrip();obj,n=decoder.raw_decode(text);text=text[n:];modules.append(obj)
records=[];chunks=['Resolved build dependency notices. FFmpeg is operator supplied and not redistributed.\n']
for m in modules:
 if m.get('Main'):continue
 d=Path(m.get('Dir',''));files=[f for f in d.iterdir() if f.is_file() and f.name.lower().split('.')[0] in ('license','copying','notice','copyright')] if d.is_dir() else []
 records.append({'ecosystem':'go','name':m['Path'],'version':m.get('Version',''),'licenceFiles':[f.name for f in files]})
 for f in files:chunks.append('\n===== '+m['Path']+' '+m.get('Version','')+' / '+f.name+' =====\n'+f.read_text(encoding='utf-8',errors='replace'))
for package in (root/'web/node_modules').rglob('package.json'):
 try:meta=json.loads(package.read_text(encoding='utf-8'))
 except (ValueError,OSError):continue
 if not meta.get('name') or not meta.get('version'):continue
 files=[f for f in package.parent.iterdir() if f.is_file() and f.name.lower().split('.')[0] in ('license','licence','copying','notice','copyright')]
 records.append({'ecosystem':'npm','name':meta['name'],'version':meta['version'],'license':meta.get('license','see upstream'),'licenceFiles':[f.name for f in files]})
 for f in files:chunks.append('\n===== '+meta['name']+' '+meta['version']+' / '+f.name+' =====\n'+f.read_text(encoding='utf-8',errors='replace'))
(out/'THIRD-PARTY-LICENSES.txt').write_text('\n'.join(chunks),encoding='utf-8');(out/'dependencies.json').write_text(json.dumps(records,indent=2),encoding='utf-8')
print('Inventoried',len(records),'dependency records')
