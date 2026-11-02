"""Package the five owned payloads for independent Workspace app updates."""
import hashlib
import json
from pathlib import Path
import sys

root = Path(__file__).resolve().parents[1]
release = Path(sys.argv[1]).resolve()
package = release / 'package'
contract = json.loads((root / 'compatibility.json').read_text(encoding='utf-8-sig'))
version = contract['version']
tag = 'v' + version
repository = 'oneworkspacestudio/my-social'
archive = f'my-social-{version}-windows-amd64.zip'
app = {
    'id': 'social', 'name': 'My Social',
    'description': 'Capture, compose and share your moments, Stories and Reels',
    'version': version, 'kind': 'process', 'depends': ['accounts', 'drive'],
    'package': 'social', 'executable': 'MySocial.exe',
    'args': ['--data', '${DATA}/social', '--listen', '127.0.0.1:${PORT}',
             '--origin', '${ORIGIN}/social', '--no-browser'],
    'port': 9013, 'route': '/social', 'stripPrefix': True, 'health': '/healthz',
    'callback': '/social/auth/callback', 'oidc': True,
    'source': f'https://github.com/{repository}/releases/tag/{tag}',
}
owned = {
    'social': ('MySocial.exe', 'MySocial.exe'),
    'social-LICENSE': ('LICENSE', 'licenses/MySocial-LICENSE'),
    'social-NOTICE': ('NOTICE', 'licenses/MySocial-NOTICE'),
    'social-THIRD-PARTY-LICENSES.txt': ('THIRD-PARTY-LICENSES.txt', 'licenses/MySocial-THIRD-PARTY-LICENSES.txt'),
    'social-contract': ('compatibility.json', 'docs/MySocial/compatibility.json'),
}
packages = {}
for key, (member, target) in owned.items():
    source = package / member
    entry = {'file': target, 'sha256': hashlib.sha256(source.read_bytes()).hexdigest(),
             'size': source.stat().st_size, 'asset': 'MySocial.exe' if key == 'social' else archive}
    if key != 'social':
        entry['member'] = member
    packages[key] = entry
manifest = {'schema': 2, 'integration': 3, 'repository': repository, 'tag': tag,
            'apps': [app], 'packages': packages, 'minimumManager': contract['minimumWorkspace'],
            'requiredCapabilities': ['scoped-social-storage-v2']}
(release / 'workspace-update.json').write_text(json.dumps(manifest, indent=2) + '\n', encoding='utf-8')
print(f'Prepared {tag}: five owned packages, direct executable and preserved minimum manager')
