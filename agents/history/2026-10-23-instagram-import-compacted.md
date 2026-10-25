# Instagram ZIP importer — preview5 delivered

Human requested the importer and then explicitly authorized commit/push, release
and local installation. Implementation source40efe96675285f698791fba935080dcc3802757a
is pushed to main and codex/instagram-import; public immutable release:
https://github.com/oneworkspacestudio/my-social/releases/tag/v0.1.0-preview.5
Windows CI40028802687 passed. No personal Instagram export was supplied/imported.

Canonical source I:/Projects/my-social was based on108cfbc, retaining preview4's
profile-photo and restart fixes. The conversation's workspace mirror was stale
at fdc8bab; it was safely fast-forwarded from the public remote. Keep both source
copies synchronized. Preserve migration2; never replay schema1 on a newer DB.

Authenticated/CSRF-protected multipart POST /api/imports/instagram/preview and
/apply accept options JSON first, then archive ZIP. Preview returns ZIP digest
and per-post content keys/status. Apply validates the digest,1–100 selected keys
and explicit audience. No schema change. Existing posts.request_key uniqueness
provides per-account deduplication from source kind, caption, original timestamp
and media hashes. Repacking the same content deduplicates; changing it can create
a different import. Each post/carousel commits atomically; partial results permit
safe retry without altering existing audiences or notifying historical followers.

All numbered posts manifests in content/ and your_instagram_activity/{content,
media}/ are supported with an optional enclosing export directory, plus reels,
stories and archived_posts wrappers. Captions/dates/order are retained and
unambiguous Latin-1/UTF-8 caption encoding is repaired. Images include WebP and
are sanitized to JPEG. Native FFmpeg converts bounded video preserving aspect
and optional audio; extras.instagram.mediaTypes enables mixed carousels.
Stories/archived posts stay archived. Followers, messages, likes, comments and
profile details are excluded. No thumbnail preview; review shows captions,
dates, media counts and availability. No silent dropping of carousel media.

AUI-generated screen at #instagram is linked from Settings. Its controller owns
file/selection fields, keeping ZIPs out of editor/device drafts. Starts with no
posts selected and Only me. ZIP is spooled only in the guarded adapter, removed
on return, and never extracted using archive paths. Traversal, links, duplicate
paths and oversized structures are rejected. Interrupted instagram-writing
intents are recovered before the next import; crash-left ZIPs are removed by a
later import after1hour. Requests serialize, budget15min; videos3min each.

Limits: ZIP512MiB, expanded2GiB, entries20k, metadata16MiB, posts1000,
100 selected/request,1–10 media/post,100MiB/original, images24MP, videos90s,
existing2GiB/account preview quota. ZIP64/encrypted/multi-volume and joining split
ZIPs unsupported. Missing media is unavailable. Import requires the protected
scoped My Photos/My Social adapter, immutable account identity and existing media
authorization. Private originals never become raw shared-folder URLs.

Verification: module verification, Go tests/vet, TypeScript/Vite/frontend tests;
native FFmpeg Reel/mixed-carousel import and >90s rejection; malformed ZIP/JSON,
CSRF/auth/guard/audience, dates/captions, numbered manifests, archive state,
repack retry/audience preservation, quota, whole-carousel rollback and crash
recovery. Exact final packaged executable passed3Chrome flows (camera/editor/
Story/PWA, profile photos, ZIP preview/selection/private import/duplicates/reload/
390px layout), then restarted healthy against the same synthetic database.
An initial broad carousel test locator was corrected to scope the imported card;
final full harness passed. Desktop/mobile screenshots visually reviewed.

Release5 has12verified asset digests/sizes; anonymous public ZIP download and tag
target verified. Its4-owned-package manifest passed the exact public Workspace74
validator before/after publication, preserved other apps and rejected an unowned
contract. A derived public74 offline bundle verified all147payload hashes/sizes,
changing only Social. Installed using the normal protected Social-only updater,
with verified public74 installer bytes and full consistent rollback backup.

A concurrent Slicer7 update began just before our initial preflight. That attempt
stopped BEFORE invoking a native updater or changing owner data. Waited for the
other native updater's actual successful exit, then captured the new baseline.
Our native updater and driver both exited0. Final verification preserved
Workspace74, Slicer7, Mail18 and every other app contract/payload hash, immutable
identity, paths, ports, selection and autostart. All25local/public services healthy;
served JS/CSS match the tested build and contain the importer. Exact tested,
published and installed Social5 SHA256:
860428231184d124fc48f73da7119e9f0105053369d181cb1e7f03ecbcc2cf4c
Latest rollback manifest verified9de24de2422e43835e571cdfbb916e66003ea3918cffcbcfb4ad258ba650061a.
Normal cleanup removed the prior automatic backup and retained the latest
verified one; some older files were conservatively kept by the updater. Do not
sweep unrelated private state or remove the newest usable recovery copy.

Evidence I:/Projects/tooling/social-instagram-20261023: release5/,
native5/, publication-verification.json inside release5, anonymous-download5.json,
offline-scope-verification5.json, owner-update5.log, updater-exit5.json and
delivery-verification5.json. Synthetic NTFS release fixtures were stopped and
removed; cleanup-release5.json. No owner data used as a development fixture;
no physical Android/iOS, clean-host or reboot acceptance claimed. Public release
assets/tag stay immutable. docs/instagram-import-acceptance.json records delivery.

Usage: Settings -> Import Instagram ZIP. No remaining implementation/publication/
installation work. Future fixes require a new version and protected update.


## Preview7 progress and compact review (2026-10-24)

User requests visible progress and a five-post preview followed by one Import
click for all available posts. Opt-in NDJSON events report ZIP checking and
per-post/media import; XHR reports actual upload bytes. Elapsed timer, native
progress bar and Stop are shown. Existing nonstream JSON API remains compatible.
UI renders first five rows only, keeps Only me default, and automatically applies
all ready keys in batches of100. Completed posts survive cancellation/retry;
existing dedupe and per-post atomic commit remain. No schema or identity change.
AUI source regenerated; capture ZIP link points to public Chrome0.1.2.

Go1.27.3 full tests/vet, frontend build/editor tests, exact packaged native7
five-flow Chromium harness and same-private-data restart pass. Synthetic 101-post
case proves requests100+1 and all101 imported despite only five review rows.
Throttled preview verifies checking progress before final review; mobile390px
no horizontal overflow. Initial multipart request observer missed Chrome's omitted
file bodies; fixed test instrumentation observes sent FormData, no runtime change.
Desktop/mobile screenshots reviewed. Owner media/data never fixtures.

Next: publish native7, validate fresh five-package and legacy four-package Social
contracts with exact Workspace83, promote public channel, normal backed-up
Social-only installation preserving manager83 and every other app. Evidence
I:/Projects/tooling/social-import-progress-20261024. Do not mutate old releases.


### Preview7 delivered

Public source2ce4a4aa83ba034404d1d0a423417011ca8823b8, release
https://github.com/oneworkspacestudio/my-social/releases/tag/v0.1.0-preview.7
Core CI40134982983 and capture-only CI40134982994 pass. All10 release assets
match GitHub digests/lengths; anonymous native ZIP/EXE/manifest and source tag
verified. Native7 SHA6c9fd7be7f9092b93c4fd518d0e0416a34005c7cbfb53fc88ee9c304780364f8.
Chrome0.1.2 is unchanged. Workspace channel promotion merged concurrent Cinema
records at1f147be; raw anonymous Social7 and withheld builds verified.

Exact83 fresh5/legacy4 manifest checks and negative ownership regressions pass.
Normal isolated83/Social6->Social7 backed-up upgrade, real Accounts/OIDC sign-in
and same-data restart pass. Synthetic fixtures stopped and removed.

A concurrent Cinema owner update held the native lock and upgraded Workspace84.
Waited for its actual exit/healthy27-service baseline. Our83-bundle preflight
correctly rejected a version regression BEFORE mutation. Rebuilt a Social-only
bundle from the locally built84 bundle, verified every166 payload hashes/sizes
and current manager match; exact84 updater source matches native commit083716e.
No claim that84 was publicly released at this checkpoint. Preflight and final
receipt prove manager84, Cinema and every other app/package/executable retained.

Normal protected Social-only update exit0, full5.2GB recovery, preserved owner
identity/OIDC/recovery credentials, roots/ports/preferences/selection and legacy
document registration. All27 runtime services and direct app probes healthy;
public runtime healthy and Social keeps its sign-in gate. Exact deployed JS/CSS
match tested build. Unauthenticated public app probes require browser-compatible
headers (Cloudflare1010 otherwise) and sign-in; Vault has a special route, not a
generic public health endpoint. Do not treat those auth responses as failed apps.

Newest recovery manifest matches its registered checksum. Normal cleanup removed
one superseded automatic backup; older unknown partial files conservatively kept
with pre-existing warning. Keep newest usable recovery; do not sweep owner state.
Idle83 maintenance helper closed through authenticated local API and current84
helper reopened normally. Fresh check confirms Social7 current, unavailable=false,
no reason/review warning. Local Python POST can reset its response connection
after accepting the operation; verify fresh operation timestamp/completion/state.
Evidence docs/instagram-import-progress-acceptance.json and task tooling.

No required implementation/publication/installation work remains. Refresh My
Social and choose the ZIP again to use the progress bar and five-post sample.
Do not import owner content as test fixtures or change immutable published assets.


## Preview8 persistent server imports (25 October, delivery in progress)

User reports no successful ZIP upload and explicitly authorizes testing a ZIP in
Downloads, fix/push/release/install on both existing PCs. Use private source data
only in isolated NTFS tests, never publish/archive its captions/media/screenshots
or seed synthetic posts into owner installations. The supplied221,742,001-byte
ZIP has99 recognized posts; initial new-job test imported97 with2 failures due
to the former90-second cap. Imported clips now allow10minutes (100MiB originals,
120MiB outputs,10minute conversion budget); a second full private test imported
all99 with0failed in128seconds. Ordinary editor/video limits remain unchanged.

Migration3 stores instagram_import_jobs (private JSON, indexed owner/state,
retained staging reservations). h.Migrate2 is conditional, latest3 still rejects
future schemas. Two retained archives per account/four globally (2GiB max) include
failed-post retry archives. Upload chunks2MiB have SHA256/offset receipts, Sync
before DB commit, truncate uncommitted suffixes after interruption. Chunk requests
have a separate512/minute budget; other mutations retain120/minute. Uploads still
need an open tab until complete; after closure, reselect identical ZIP and verify
all stored chunk checksums before resuming. Do not promise browser-closed bytes
can be sent. No giant request, no re-upload on Import, no client-side batching.

New owner/CSRF-protected /api/imports/instagram/jobs create/list/detail/chunks/
complete/import/cancel/retry routes; older multipart API remains compatible.
Run(ctx) worker persists checking/importing, resumes across actual Host close/New
with existing per-post dedupe. Each post and all worker writes hold backup read
locks, released between posts. Durable owner issuer/subject feeds storage guard;
new sessions are not required for background work. Disabled accounts cannot
import; account deletion orders publish/job locks and cancels jobs. Explicit
Cancel differs from pausing a browser upload; completed posts remain.

Server results expose only five sample items/full counts, chunks and actual job
progress; monotonic revisions stop stale polls reverting state. AUI source owns
UI; leaving/unmount aborts only in-flight upload, never server jobs. Return restores
server jobs and audience. Completed ZIP removed, retry failures/ready/upload expire
after24h, result records after7days plus cleanup allowance. Empty and canceled
reservations expire safely; legacy cleanup never touches job-* archives.

Six native Chromium flows passed final candidate8 including actual tab close
mid-upload, resume from saved offsets and import complete with no page open.
Empty-upload expiry was tightened and the exact final artifact then passed all6
flows again plus same-data restart. Go tests/vet and frontend build/editor tests
pass. Current evidence I:/Projects/tooling/social-server-jobs-20261025.

Source canonical I:/Projects/my-social, mirror under OneWorkspace/my-social.
Workspace paired worktree one-workspace-update-docs fast-forwarded to6830a20;
AGENTS says no Actions for Workspace and promote every app release. Native owners
advanced to86; remote Social6, local Social7 before this update. Another MyAI task
is preparing87. Never downgrade/update unselected apps: use current public native
maintenance Social-only selection, verify actual baseline/locks immediately before
install. SSH alias m8e-pc authentication fails; use existing authenticated Chrome3
workspace-2.m8e.co.uk/updates page. A new task-owned tab585734234 is bound as
remoteSocialUpdates in cua_repl; don't commandeer other task tab585734233. It showed
86/Cinema5 and Social7 offered over6; do not apply7 while preparing8. No operation
started there. User RDP tab585733890 exists but normal Updates UI is sufficient.

Next: finish exact final8 six-flow browser check; commit/push, CI, package verified
EXE/source/notices/managed5-package manifest (minimum83), publish immutable8,
promote raw channel preserving concurrent/withheld releases, run normal backed-up
Social-only updates on both PCs, verify versions/bytes/other apps/current UI.
Keep newest owner recovery; C space is constrained by other apps and recovery,
never bypass normal backup guards or delete/move owner content.


## Preview8 delivery completed (25 October)

Published immutable v0.1.0-preview.8 at source bfcc8d4d44336d6a67930bdadae76d38c64bc4b6;
all10 assets/digests and anonymous native ZIP/EXE/manifest verified. Native SHA
683895f65bb9a9feac518011b4490705a5b13fe4914e0acb6794b8083cc79215.
Core CI40179327321 and capture CI40179327364 pass. Channel promoted by Workspace
b40fbc6, retains concurrent87/MyAI4 and all other records/withholdings; anonymous
raw channel/tag verified. Exact current manifest selector passes fresh5/legacy4
Social layouts, retaining manager, other apps, runtime paths and registration.

Normal public native Social-only updates completed on BOTH PCs: primary7→8 and
second6→8, both retain Workspace87. Primary operator exit0 independently verifies
public Social bytes, manager/25 other process hashes, all contracts/package pins,
owner identity/paths/ports/preferences/selection and every decrypted credential/
client. All27 services/direct local apps and public runtime healthy; exact served
JS/CSS matches final tested build. Fresh current-controller check:8 up to date,
no Social review warning. Latest registered recovery manifest hash verified;
normal cleanup removed one superseded automatic backup, kept unknown old partial.

The second PC's prior Workspace/MyAI update completed before our separate update.
Our single Social-only transaction finished all9 stages in19m2s, reports all apps
healthy and newest verified recovery retained, removed one older automatic copy.
Fresh installed release UI shows Social8 up to date; all25 other displayed versions
(including87/MyAI4/Cinema5) retained. Remote evidence is normal native verification
and installed-version UI, NOT an independent filesystem/decrypted-secret audit.
No SSH credentials available; no repeated apply, guard bypass or owner fixtures.

Six final native synthetic flows include actual tab close during upload/resume and
server import with no browser page open. Authorized private222MB ZIP integration
imported99/0failed in isolated NTFS. Those posts were test imports, not inserted
into the owner's live Social profile. Successful upload must finish before tab
closure can leave checking/import running unattended; partial uploads resume after
reselecting identical ZIP. Ten-minute import-only video limit fixed2long clips.
Task-owned stopped NTFS fixtures removed after process/reparse/path checks. Public
proof: docs/instagram-server-jobs-acceptance.json; local detailed evidence under
I:/Projects/tooling/social-server-jobs-20261025. Do not move runtime tag/assets.


## Native Windows upload implementation (25 October, publication pending)

User explicitly authorizes native background transfer, push/release/install both.
Candidate Social9 / Workspace89 (88 reserved by concurrent MyAI managed-projects
source6f14ff1; preserve that source and wait for its public bundle as baseline).
Social adds single-EXE modes for URI activation, Windows ZIP picker, encrypted
current-user DPAPI queue, independent chunk worker, login resume and source file
read sharing that prevents mutation. Per-user registered mysocial-upload protocol
and HKCU Run use a verified protected content-addressed copy, so server updates
can replace MySocial.exe without locking the active uploader. Standalone fixtures
skip registration; Windows registry tests use private test keys only. Direct CLI
file-selection mode enables exact-native acceptance without automating owner UI.

Migration4 stores hashed expiring grants. Browser-authenticated/CSRF grant creation,
10minute unclaimed ticket, rotated client key bound to ONE job for24h; only claim,
receipt/chunk/complete/problem allowed. No sign-in, publication/generalAPI rights.
Native endpoints recheck immutable account/storage guard, keep backup locks and
existing chunk checksum/offset behavior. Initial credential cannot access client
API, replay with another client fails, disabled/expired/cross-user checks pass.
Workspace89 narrow public proxy boundary delegates validation to Social; ordinary
routes retain portal sign-in. Native/client records never expose captions/media.

All7 candidate native browser checks pass, including starting actual packaged
Windows transfer with all tabs closed, returning to review/import and phone layout.
Core transfer passes actual Host close/reopen with all browser sessions removed,
nonzero resume offsets, receipt/source matching and encrypted queue/registry tests.
Authorized original221742001-byte ZIP transfers fully via new client core and
server recognizes99 posts in isolated NTFS; source media/captions not public or
inserted into live owner profile. Final tiny timeout/startup-recovery refinements
require final build/browser check before publication. Full Social tests/vet and
front-end build/tests pass; gateway/portal/proxy/distribution/catalog tests/vet pass.

Next: final checks/source commits/CI; publish9(minimum89), build89 from public88
preserving concurrent MyAI5, fresh real Accounts/Drive/Social native-grant flow,
verify public assets/promote both; normal backed-up BOTH-PC updates selecting only
Workspace+Social with actual latest baseline/locks. Preserve every unselected app,
credential/config/path and latest verified rollback; never downgrade88/MyAI5.
Artifacts I:/Projects/tooling/social-native-upload-20261025, private fixtures only
C:/Projects/.cache/social-native-upload-20261025. No owner installation as fixture.

Final Social9 candidate re-passed all7 native flows and same-data executable restart after timeout/startup refinements. Workspace88/MyAI5 public assets are now verified and promoted (source6f14ff1, channel5353b2a); derive89 from that exact public88 bundle. Native grant creation refuses older managers without the explicit89 runtime capability instead of offering a failing transfer. Source/publication/normal two-PC delivery remain next.

Final queue refinement uses a Windows mutex per transfer, so selecting another ZIP can claim its ten-minute permission immediately while an earlier upload is still running. Queue inspection also occurs under that mutex, avoiding concurrent receipt-file replacement. Rechecking the final native artifact before publication.
