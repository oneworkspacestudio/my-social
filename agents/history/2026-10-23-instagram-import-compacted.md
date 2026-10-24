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
