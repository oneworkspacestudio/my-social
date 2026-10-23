# Instagram ZIP importer

User requests an importer that reads Instagram export ZIPs and applies content
to My Social. Implementation only is authorized; no commit, push, release,
owner installation update or import of personal data was requested/performed.

The conversation workspace clone at
`C:/Users/fahim/Documents/ChatGPT/OneWorkspace/my-social` was stale at fdc8bab.
Current canonical source is `I:/Projects/my-social`, clean at 108cfbc before this
task, containing public preview4's profile-photo/restart fixes and delivery
history. All implementation changes are uncommitted in that canonical checkout.
Do not replace it with the older workspace clone or replay schema1 over schema2.

Added authenticated/CSRF-protected multipart POST
`/api/imports/instagram/preview` and `/apply`. Fields: `options` JSON first,
then `archive` ZIP. Preview returns ZIP digest and per-post content keys/status;
apply requires matching digest, selected keys (1–100), explicit audience. No
schema migration. Existing posts.request_key uniqueness provides per-account
deduplication using source kind, caption, original timestamp and media hashes.
Repacked exports deduplicate; changing that content can create a new import.

Recognizes all numbered posts manifests in old content/ and newer
your_instagram_activity/{content,media}/ layouts, with optional enclosing export
directory, plus reels, stories and archived_posts wrappers. Preserves captions,
dates/order, handles unambiguous Latin-1/UTF-8 caption repair, sanitizes images
(including WebP), transcodes bounded video with aspect/audio, and displays mixed
carousels using extras.instagram.mediaTypes. Stories/archived posts stay archived.
No follower relationships, messages, reactions, comments or profile overwrite.

AUI source/generated InstagramImport screen is linked from Settings at
`#instagram`; controller owns its own file/selection fields so ZIPs never enter
unrelated editor/device drafts. Preview is captions/dates/media counts/status;
no thumbnail preview. Starts with no posts selected and Only me audience.
Up to 100 selected posts per request; detailed partial results permit safe retry.
Each post/carousel commits atomically; historical imports send no notifications.

ZIP is spooled only under the configured guarded adapter, then removed per
request; entries never become extraction paths. Rejects traversal, links,
duplicate paths, unsupported ZIP structures, huge directories/files/expansion.
Limits: ZIP512MiB, expanded2GiB, entries20k, metadata16MiB, posts1000,
original100MiB, images24MP, videos90s, existing account2GiB quota. ZIP64,
encrypted/multi-volume and joining split ZIPs are unsupported. Missing media is
unavailable, never silently dropped from a carousel. Protected storage/identity,
owner-only originals and existing media sharing checks remain mandatory.
Import mutex serializes requests; uncommitted instagram-writing intents are
recovered before next import. ZIPs left by process crash are removed by a later
import after one hour. Apply reuploads the same ZIP; request processing15min,
per-video3min. Lost responses are recovered by preview/retry, not a durable job UI.

Verification passed: Go tests (including actual configured FFmpeg/FFprobe Reel,
mixed carousel and >90s rejection), vet, frontend editor tests, TypeScript/Vite,
and native Chrome3 workflows: existing camera/carousel/Story/PWA, profile photos,
new ZIP preview/selection/private import/duplicates/reload/mobile. Added Go tests
cover isolation, no preview content writes, captions/dates, numbered manifests,
archive state, repack retry/audience preservation, unsafe archives, malformed
JSON, digest/selection/CSRF/auth/guard, quota, partial rollback and crash recovery.
Initial E2E locator matched another existing carousel; corrected to scope the
imported card, then entire native harness passed. Same-data EXE restart healthy.
Desktop/390px screenshots visually reviewed; no overflow in automated checks.

Build/evidence: `I:/Projects/tooling/social-instagram-20261023/`, executable
MySocial.exe; screenshots/ and browser-process.json/restart-verification.json.
Only synthetic NTFS fixtures under C:/Projects/.cache/social-instagram-20261023.
All preview processes stopped and that exact NTFS fixture tree was verified and
removed after testing; cleanup.json records completion. I: evidence is retained.
No owner installation or personal export used; no physical-phone acceptance.
Usage/limits documented in docs/INSTAGRAM-IMPORT.md and README.

Resume: review uncommitted source. A future explicitly authorized release/install
must bump/package metadata, follow the existing release4/Workspace74 contracts,
and use the protected normal updater. Never overwrite the running owner payload
with this development executable. No source commit or external publication made.

## Publication follow-up

User explicitly authorizes committing/pushing, releasing and installing this
importer. Target preview5, preserving Workspace74 and current other apps.
Earlier implementation-only boundary above describes the previous turn.
Use the normal protected Social-only updater and verify actual process exit,
published/tested/installed hashes, frontend assets, rollback and suite health.
