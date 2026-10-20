# My Social initial staged preview

User authorizes the full Instagram-style PWA brief, local ecosystem update,
organization source push and verified GitHub releases of Social and Workspace,
with local tests and GitHub Actions required. Full parity is unfinished: exact
implemented features and gaps are in docs/ACCEPTANCE.md; do not claim production,
physical-device, clean-second-host or service/reboot acceptance.

New original Apache-2.0 source I:/Projects/my-social. Shared apphost is pinned to
public Workspace31171aa1970e (explicit camera opt-in, self/blob media CSP), whose
Windows CI succeeded. No local replacements or go.work in release builds.
Source/build/downloads on I:/Projects/tooling/my-social-20261020; protected test
state C:/Projects/.cache/my-social-20261020. I: is exFAT, so GOTMPDIR for protected
Go tests must point to the named C test temp. Exact Git safe.directory exception
was added only for this new checkout. No owner content used as a fixture.

Implemented native SQLite API + AUI-generated React PWA, Fabric scenes/camera,
private originals under My Photos/My Social, independent carousel editor/drafts,
idempotent publishing/scheduling, Story expiry/polls/Highlights, bounded FFmpeg
Reel jobs, basic social/Studio/moderation, durable SSE/VAPID notifications.
My Chat is a launcher only after inspecting real Tinode/session contracts.

Local Go tests/vet and crop/link checks passed. Optional native FFmpeg acceptance
created two actual clips, reordered/trimmed/speed-adjusted, rendered and published.
Synthetic Chrome flow passed on actual EXE: fake camera without audio, layered
multiline text/filter/undo, independent two-image drafts/reload/preview/publish,
like/comment, 390px layout, Story poll/pause, PWA manifest/worker. Corrected text
origin after visual QA; final screenshots are synthetic. CI and managed native
Quantum installation verification/publication remain pending at this handoff.

Suite integration is in existing I:/Projects/one-workspace. Reserve Files/My
Photos/My Social with Quantum denyAll; reject all relevant explicit allow rules
because those override denyAll. Loopback capability guard rechecks isolation.
Do not expose originals through raw Drive/static routes; owner-only originals
and per-post derivatives are mediated by Social. Standalone isolated adapter is
explicit; absent storage fails uploads/publication. Other app contracts preserved.

Next: finish fresh managed OIDC/Quantum denial test, commit/push Social, pass CI,
package and verify preview1, integrate its exact public pin in Workspace47,
test final full installer, pass suite CI, publish both immutable releases and
use the normal backed-up updater for the authorized local installation. Preserve
four unrelated untracked suite histories and the newest verified owner rollback.

Preview1 published418292513/targeta7dfda3 with9 matching assets; Workspace47
public418297021/target5b83518,21 matching assets, full104-package bundle. Normal
owner47 update is underway via approved public-digest-equivalent local bytes.
Current app follow-up2 fixes actual camera zoom binding and fences device draft
storage by its active account. Browser capture/editor/Story/offline and old-account
draft-removal checks passed locally; server schema remains1 and minimum manager47.
Do not rewrite preview1/47 tags or assets. Publish2 as a compatible independent
app update, then use the normal Social update group after owner47 completes.

Security correction: Windows native logical ACLs preserve casing. An isolated
real Drive probe returned404 for exact denied path but200 and identical marker
for lower/upper-case paths. Earlier exact-path acceptance was insufficient.
Do not use preview1/Workspace47 storage for private media. Published tags/assets
remain immutable; release notes mark predecessors superseded. The local owner
root contained only its marker, no media, and uploads were paused by renaming
that marker. The new adapter uses Data/photos-social/My Photos/My Social outside
all shared sources, default-disabled My Social Originals native source, and
read-only per-subject scopes; physical native scope containment provides the
boundary. Social2 requires Workspace48 and scoped-social-storage-v2.
Existing legacy media is refused, never silently moved. Managed tests and final
release48/publication/local update are pending this correction handoff.

Delivery complete: Social v0.1.0-preview.2 target0744a86c826a978c76cbfe521b00336dbe187dc0
(release418427788,9 verified assets, anonymous public ZIP verified) and Workspace
v0.3.0-preview.48 target1e589609b86b96ca2aa2a4e61b0d5430893b21ce
(release418432236,21 verified assets) are public. Source CI39429278014 and
39430126063 passed, as did packaged Chrome and native managed acceptance.
A real published47-to48 backed-up upgrade preserved identity/issuer/paths,
ordinary bytes, Accounts/Drive hashes and18 app opt-outs, then passed own native
original identity and cross-account/case/traversal denial. Existing20 other
component payloads/versions remain unchanged.

The authorized real local installation now runs48/Social2, public-digest-matched
manager1c459041 and Social3f938d8f, with21 apps and health verified. Normal updater
created a verified rollback; preserve the newest recovery copy. Owner data was
never a development fixture. Existing operator FFmpeg/FFprobe are configured; no
helper binaries redistributed. Synthetic C fixtures were stopped and removed.
Source clones in the user workspace are synchronized. Full Instagram parity,
advanced video/audio, Social messaging/live, physical Android/iOS push/capture,
clean-host restore and independent security review remain documented gates.
The predecessor1/47 tags/assets stay immutable and their notes are superseded.
Do not revive shared-folder ACL storage. Existing nonempty legacy media still
requires a separate explicit migration; this local owner had no legacy media.
