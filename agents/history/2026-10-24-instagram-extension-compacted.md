# Instagram capture extension — published; Chrome profile selection pending

User explicitly authorizes implementation inside My Social and push/release/install.
Canonical I:/Projects/my-social; workspace mirror is synchronized after publication.
Source a5fc1bcc44bab6a0746e49603c4ff014a784fe68 on main and codex/instagram-capture.
Commits48322e0/231ed0b implement/package the extension; a5fc1bc improves test login
synchronization only. Production binary is the exact231ed0b build, unchanged by
that test-only delta; build-info explicitly records both commits.

Public release v0.1.0-preview.6, extension0.1.0:
https://github.com/oneworkspacestudio/my-social/releases/tag/v0.1.0-preview.6
CI40107462751 passed. Release assets/digests, anonymous public extension/native ZIP
retrieval, ZIP CRC/member hashes and tag commit verified. Evidence tooling lives
under I:/Projects/tooling/social-extension-20261024, release6/public6/native6.
No private owner information was used as a fixture or included in public evidence.

MV3 action opens a user-started export tab. Optional bounded scrolling and ordinary
sequential post-page reads; metadata from hydrated JSON/JSON-LD/visible article/OG
previews. No private API, cookies/passwords/tokens, DMs, raw script/HTML export or
server upload. Instagram/CDN-only permissions. Select posts and download standard
STORE ZIPs compatible with Social5 importer. Comments/profile details retained as
report, not native recreated objects. Best effort; some layouts/media unavailable.
Preserves dates/captions/order where accessible; explicit capture-time fallback,
2200-byte import captions with full text in report, 10-item splits, sub400MiB ZIP
parts,100MiB/file and350MiB/post caps. Incomplete carousel fragments/report retained
outside import manifest. Native90s video/image/storage validation still applies.
AUI export screen authoritative; generated TSX. React/152 dependency notices ship.

Five ZIP/permission unit tests passed. All4 native browser workflows and same-data
restart passed on packaged6; realMV3 synthetic Instagram capture/download/import,
carousel/emoji/dates/private default, 390px screenshots reviewed. Existing public5
importer also passed that workflow. First fixture charset failure was fixed; first
CI login timing failure fixed by settling initial session check and asserting POST
login200 before30s Home wait. A redundant full mod verify repeat was stopped after
only test synchronization changed; previous full build/mod verify/Go tests/native
FFmpeg/vet/frontend tests had passed. Final CI and exact EXE browser suite passed.
Playwright download lock fails on exFAT; initial browser download completed on NTFS,
then verified completed browser binaries moved to I tooling. Private synthetic test
state remained isolated NTFS. Named C fixture cache can be cleaned after processes
exit. Never use owner app data for fixtures.

Managed release follows standard native/source/license/manifest packaging because
an extension-only newest tag would obstruct Workspace's newest-release discovery.
Exact public Workspace80 validator requires five owned Social packages, including
social-contract at docs/MySocial/compatibility.json (newer than prior74). Release
manifest includes it and requires manager80. Validator confirms other apps/ports
unchanged and rejects another app package. Initial draft missing contract rejected
before publication, then corrected and passed; public release is immutable.

Verified public extension ZIP SHA256:
15fd07418baea0c17c48dde528575bcb06c8f090e5d6b32d1d7593f5590f14e6
Public install helper safely extracted to stable
%LOCALAPPDATA%/MySocial/InstagramCapture, with install metadata. Does not overwrite
existing install, edit browser prefs, policy/registry or auto-capture personal data.
**Not yet loaded into Chrome.** Normal Chrome profile picker has ambiguous duplicate
labels; async user question asks them to open intended Instagram profile and reply
ready. No profile selected, no extension install card verified. Continue via normal
Chrome chrome://extensions → Developer mode → Load unpacked → stable folder; inspect
enabled card/name/version0.1.0, then update acceptance/history and commit/push.
Use computer-use skill sky API only for native UI/file picker. Fresh list_apps and
returned window, observe then one action+refresh. No auth/security barrier bypass.
Do not repeat install-files helper: stable directory already exists correctly.

Native owner Social remains preview5, which already imports extension ZIPs.
No native Social6 update is needed for this requested Chrome workflow. C has low
free space, F nearly full; do not attempt owner native update without sufficient
NTFS rollback capacity, remove unknown caches/backups, or move unencrypted owner
backups onto exFAT. Current Workspace80 must not be downgraded. Final user-facing
status must distinguish released/installed files from pending Chrome enablement.
