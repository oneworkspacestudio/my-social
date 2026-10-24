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


Follow-up: user selected their profile; normal Chrome window confirms the selected
profile. Native helper could open Load unpacked but rejected modal input with
geometry/target-window errors. No successful Chrome card verification yet. Asked
user to select the folder manually. AppData is hidden in their folder navigation,
so a second copy was installed using the verified public helper at visible
C:/Projects/MySocial/InstagramCapture; every ZIP member hash checked. Both copies
are intact. The latest async question asks user to paste that visible path in the
Folder box, click Select Folder and reply done. Resume with read-only extension
card/name/version/enabled verification, then mark acceptance complete and push.
Use unique task-scoped node_repl globals (e.g. mySocialInstall20261024) to avoid
state collisions; refresh exact returned windows. Do not keep retrying rejected
modal coordinates or alter browser security/preferences to force installation.


Installation completed: user selected the visible verified folder and replied done.
Read-only native Chrome UI confirms exactly My Social Instagram Capture, version
0.1.0, Unpacked extension, On/extension enabled, with capture toolbar button
available. No personal Instagram content was captured or imported by this task.
Acceptance now records successful browser installation; prior pending notes above
are historical and superseded. Public preview6 assets/tag remain unchanged. Both
verified file copies remain; do not remove either while browser source paths are
in use. Named synthetic C test cache was cleaned after process/reparse checks.
The full original request (build/push/release/install extension) is complete.


## Own uploads correction — implementation follow-up

User clarified that their own uploaded media is the purpose, not arbitrary feed
posts or importer instructions. Live profile read (no personal media fixture or
export) confirmed username-prefixed /username/p/code URLs.0.1.0 matched only root
/p/code URLs, so it could miss that grid entirely; scrolling was off by default.
Do not claim the user's complete archive or all their files have been downloaded.

Extension0.1.1 now accepts prefixed/long post and Reel links; explicit username,
own profile/Reels scope, other-owner filtering, default100-step bounded scroll,
rendered inactive-tab fallback and carousel-next traversal. Preview-only media is
reported incomplete and excluded from native manifests. Previews/photo-video
counts, clear Scan/Download actions, downloaded ZIP folder reveal, report file
paths and zero-media failure notice address the confusing importer-first flow.
Actual files are in Downloads/MySocial ZIPs: media/posts, partials capture/incomplete.
Profile/account source is runtime-only; no personal username/media is in fixtures.
AUI authoritative. Permission set unchanged. Temporary worker tab closed in finally.

Six unit tests and all4 native browser workflows/restart passed using unchanged
public Social6 EXE. Synthetic regression covers prefixed links, other-owner denial,
HTML-shell fallback, DOM carousel2slides, real ZIP4photos/native2post import and
390px previews. Initial assertion expected scan button ready before the longer
scan finished; updated wait to60s and passed. Desktop/phone screenshots reviewed.
Metadata/last polling robustness change will be covered by source CI.

Publishing a Chrome-component patch with unchanged native6 assets and a matching
managed Workspace manifest retains newest-release discovery compatibility. Keep
published preview6 immutable; tag/patch release notes must clearly state native
app version6 unchanged. Do not rebuild/install owner server for this browser fix.

Browser Use denies observing chrome-extension:// pages (allowed protocols HTTP/S
only), explicitly forbidding alternate-surface workaround. Respect that deny:
no native automation/CDP to inspect the blocked extension page. Public Instagram
profile HTTP/S was read normally and returned to its grid. After verified files
are updated in the known installed folders, user must Reload extension and reopen
its window. Source tests use synthetic isolated profiles, not owner browser state.
Next: commit/push, CI, verified Chrome patch release, update matching public files
in existing install directories without deleting old live hashed assets, request
manual reload. Keep Chrome runtime version pending until user confirms it.
