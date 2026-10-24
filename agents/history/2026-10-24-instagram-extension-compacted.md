# Instagram capture extension — own-uploads fix0.1.1 released

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


## Final0.1.1 delivery

Released v0.1.0-preview.6-capture.1 at source8cdd07fafdd812bd0d7d075c0ec479fe43964af8:
https://github.com/oneworkspacestudio/my-social/releases/tag/v0.1.0-preview.6-capture.1
Source CI40127887717 and Chrome/public-native-importer CI40127887780 both passed.
Six unit tests, four browser flows, restart, exact Workspace80 validator, public
asset digests/anonymous extension ZIP retrieval and source tag verified. Native6
EXE/ZIP are unchanged verified preview6 assets; owner server was not updated.

CI initially failed on the new Reels/worker navigation. Longer waits did not fix
it. Synthetic-only diagnostics showed the renderer waiting for the full tab load,
blocking enrichment after finding only grid previews. Final worker starts blank,
navigates normally, verifies the committed URL and interactive/complete document
through scripting, and lets the collector wait for media. Optional Reels failure
is nonfatal to main-upload enrichment. Final local/CI tests pass; do not revert to
waiting only for tabs.status complete or removing all discovered posts on error.
Both explicit-other-owner and unknown-owner posts are excluded; prefixed links
normalise to canonical post URLs for fetch/navigation. A standalone fast Chrome CI
uses the hash-pinned public native6 ZIP and repeats all4 flows without rebuilding
Go; full source CI separately passes Go/build and the same browser checks.

Both installed directories updated from the anonymous/hash-verified public0.1.1
ZIP. Every old file was checked before replacement, no permission expansion or
reparse paths, no old hashed assets deleted. First candidate ZIP retained locally
as a public-code rollback. User confirmed Reload of the first candidate. The final
readiness/recovery patch changed the export page script afterward; user has been
told to refresh the capture window once to load it. Browser policy still forbids
agent observation of chrome-extension pages, so no alternate-surface workaround
or claim of live-personal-scan verification. Source acceptance explicitly records
that final refresh is required. Old browser-install0.1.0 notes above are history.

The user's existing local download was located and linked back to them privately;
its content/counts/path are not repository fixtures or release evidence. No owner
media was downloaded or imported by this agent. Do not claim all account uploads
were backed up. Actual media remains in downloaded ZIPs (media/posts; partials
capture/incomplete; report lists files/failures). New workflow: refresh capture
window, Scan my uploads, then Download my media ZIPs. Existing Social importer
accepts those ZIPs; importing into Social is the destination step.

Remaining product maintenance: native preview6's convenience download link still
points to its original immutable0.1.0 extension asset. The new patch release lists
the correct0.1.1 asset and current source guide. Update that convenience link in a
future native release; do not mutate already-public preview6 assets or needlessly
install a different native server for this browser-component correction.

## Immediate media downloads0.1.2

User requested downloading media while each upload is read, avoiding a second
manual pass. Default now Scan and download my uploads. After grid discovery, an
async generator yields one enriched upload to the archive writer; its media is
fetched before the reader advances. Deduplicated candidate codes include records
found in structured data even without a grid link. Manual selection remains an
optional checkbox mode. ZIP checkpoints save every10 uploads or bounded part size;
final buffers flush on Stop and reader failure. Already saved ZIPs remain usable.
Closing the tab can still lose buffers; keep it open until saving completes.

Nine unit checks pass, including read/fetch ordering, early ZIP checkpoints, Stop
and reader-error preservation. All4 native browser flows plus same-data restart
pass against the unchanged public native6. Real MV3 test verifies the first post's
media fetch precedes the second page read, default one-click ZIP download4photos,
private2post native import, owner exclusions and390px layout. Synthetic data only.
AUI regenerated; permission set and archive format unchanged.
Next: push, verify CI, publish component patch capture.2 with unchanged native6
and validated Workspace manifest; replace matching installed public files and
leave Chrome Reload/refresh to user due to the extension-page browser restriction.
