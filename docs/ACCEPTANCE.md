# Preview acceptance and parity gaps

The brief is implemented as staged work. This release is a functional evaluation
preview, not a declaration of full Instagram feature parity or production safety.

| Area | Implemented | Remaining acceptance / gaps |
| --- | --- | --- |
| Native foundation | Embedded React/TypeScript Go EXE, SQLite WAL/migrations, stable MyAccounts identity, isolated standalone users | Clean second Windows host, SCM and reboot |
| Camera | Explicit permission, front/rear request, photo shutter, supported MediaRecorder MIME choice, visible timer, opt-in mic, pause/resume, close/background cleanup; capability-only zoom/torch | Physical Android and iOS, device loss/rotation/memory/torch/focus/exposure coverage |
| Photo composition | Fabric scenes; aspect crop/zoom/straighten/rotate/flip, brightness/contrast/saturation/warmth/blur, presets, multiline text with handles/color/font/highlight, emoji stickers, brush, layers, undo/redo, independent carousels | Freeform crop handles, pan crop, highlights/shadows/sharpness/vignette, gradients/outline/alignment UI, eraser, snapping, approved GIFs, cross-browser pixel comparisons |
| Draft/publish | Account-keyed IndexedDB recovery, private server drafts and originals, preview, caption/alt/location/audience/scheduling, retry-idempotent posts, edit/archive/delete/restore | Resumable chunk upload after reload; scanner/moderation-before-publication; tags/collaborator invitation workflows |
| Stories | Server-controlled 24h access, image/video media, text/brush scene, polls and replies, pausable viewer/navigation, owner-created Highlights | Viewer-list UI, multi-frame progress grouping, question/countdown/sliders, licensed music; replies currently audience-visible comments, not private DMs |
| Reels | Browser capture, multi-clip imports, trim/reorder/remove/speed, bounded structured FFmpeg queue, single-clip original audio or muted multi-clip, vertical player | Clip splitting, timed overlays, captions track display, covers, audio upload/mixing/fades/voice-over/transcription; no licensed library |
| Social | Profiles/bio/privacy, follow approval, blocks/mutes/restrict/Close Friends API, chronological Following, permission-filtered feed/search, likes/comments/threaded replies/saves and collection API | Recommendations are basic chronological discovery; reposts, collaborator posts, reaction palettes, QR profiles, broadcast channels, verification, friends activity/map |
| Storage | Authoritative originals under My Photos/My Social; owner originals, per-post derivatives and range checks, protected drafts and retained projects, guarded Quantum ACL | Broad-source aliases/custom Drive configurations need review; automated retention/orphan cleanup and second-host coordinated backup/restore |
| Notifications | Durable per-user deduplicated outbox, read state, SSE reconnect, opt-in Push/VAPID, retry/expired subscriptions, generic payload, toggles/quiet hours | Physical Android/iOS delivery, OS revocation and production retry/batching load acceptance; no invented DM/call events |
| Messaging/live | Inspected actual Tinode/My Chat gateway; explicit link to My Chat | No Social DM adapter, calls, ephemeral/group DM semantics, streaming backend or Live UI implementation |
| Safety/admin | Owner moderation, private reports, write throttling, upload/quota limits, account export and explicit disable | Teen/minor regional protections, spam/impersonation operations, appeals, malware scanning, independent security review, permanent erasure tooling |
| Commerce | No active payment/subscription or commercial music integration | Provider contracts and approvals required |

## Reproducible checks

`go test ./...` covers originals/derivatives/audience/blocks/follow approval,
draft/event/owner isolation, CSRF, idempotent publication, unavailable My Photos,
server Story expiry and scheduling, unsafe push URLs and link schemes.
With MYSOCIAL_TEST_FFMPEG and MYSOCIAL_TEST_FFPROBE, the optional native test
creates two real clips, reorders/trims/speeds them, renders and publishes a Reel.

`npm --prefix web test` verifies crop geometry and URL safety. `node
scripts/browser-verify.mjs` creates a fresh protected standalone fixture, starts
the actual EXE, and exercises Chrome with a synthetic camera, text/filters,
independent carousel draft/reload/preview/publish, likes/comments, Story poll,
390px layout, manifest and service worker. Screenshots contain only synthetic
test content. A mocked camera is not physical-phone permission/capture proof.

Release evidence and exact native-suite verification are appended after their
successful execution. No physical-device, clean-host, service, security-audit or
full Instagram parity certification should be inferred from a passing build.
