# Mobile Create flow and posting latency

User requests faster posting, a cleaner Instagram-style home/header, mobile-first
select → edit → share creation, source push, immutable release and installations
on both PCs. The user explicitly answered Both PCs, superseding the earlier
leave-primary instruction for this new release. No subagents or Actions used.

Canonical source I:/Projects/my-social, branch codex/mobile-post-flow, based on
7f6db79. Workspace checkout I:/Projects/one-workspace-update-docs fast-forwarded
to current public main d5c4bce before integration work. Owner baseline manager109,
Social10. Source mirror under the user's OneWorkspace remains to be synchronized.

AUI screens implement distinct select/edit/share steps, compact home header,
edge-to-edge mobile posts, profile photo in bottom navigation, existing editing
tools grouped by Crop/Adjust/Text/Draw, and management actions in a Post options
dialog. The system file picker supplies device photos; no fabricated full-device
gallery or unsupported music/tagging/live controls. Existing audiences, protected
originals, drafts, Stories, video helper requirements and imports are retained.

Photo rendering/upload begins on the caption step with two bounded workers,
shared in-flight render/upload promises, revision-aware caches and explicit
progress/failure/retry state. Caption remains editable. Completed photos are
reused; an edited photo uploads again. Share retains the existing idempotent
publication key and single-operation guard. Fabric skips zero-value filters and
uses native toBlob encoding. Server uploadSlots bounds bodies/decoders to two;
the storage writer mutex covers quota/commit rather than slow network reading
or codecs. Derivatives are sanitized once, original validity still checked, and
the unused JPEG conversion is removed. No database migration or manager change.

Version11 packaging adds scripts/package-managed.py: the same five owned paths,
schema2/integration3, minimum Workspace90 and scoped-social-storage-v2, direct EXE
asset plus archive notices/contract. A temporary current-SDK validation test
verifies fresh/legacy manifests, unsupported manager rejection and preservation
of all other apps/manager; remove its exact temporary source before publication.

Go tests/vet, frontend/editor and extension archive checks passed. Browser checks
on the actual Windows EXE verify pre-Share preparation, editable caption during
upload, partial retry reuse, changed-image reupload and exactly one post on rapid
double-tap. Local prepared-media share measured66–100ms; not a remote performance
guarantee. Camera/layered carousel/draft recovery/Story and all four importer flows
passed an initial run. Final full eight-flow run is pending at this checkpoint:
old profile test expected the intentionally removed mobile header avatar; it now
checks the visible bottom-nav photo. Extension needed Playwright Chromium1248,
installed in the named C cache after I-drive directory-lock/Go loading stalls.

Build/evidence I:/Projects/tooling/social-mobile-post-20261102. Isolated private
NTFS/cache C:/Projects/.cache/social-mobile-post-20261102. A partial fallback
source/module copy exists there (task-owned copy process stopped); no owner data.
Track/stop fixture processes and verify path/reparse/process boundaries before
cleanup. Preserve owner media, identities/credentials, unrelated apps and newest
usable recovery. Owner installs have NOT started. Primary public tunnel1033 and
no manager observed; transaction is complete, so use normal supported local
updater when ready. Secondary109 normal Updates UI is reachable/idle.

Next: finish exact packaged eight-flow regression, commit source, repackage from
that commit preserving tested EXE hash, publish/verify immutable11 and promote
public channel/legacy Social pin. Verify a fresh isolated managed install/upgrade
with real Accounts/Drive and the mobile flow, then normal Social-only updates and
independent verification on both PCs. No Workspace/Android suite release needed
for an independent app update. No physical-phone/native-picker claim.

## Release freeze checkpoint

Final exact Windows binary841364eac3320d53f9a4408dcc6bc08fe1a62dd4ed12b0c570d028163766922b
passes all eight native browser workflows in one run and actual same-data restart.
Final prepared-media Share144ms locally; synthetic/loopback only. The action guard
now returns ignored taps before entering finally, and only its owning action can
clear publication busy state. Go tests/vet, frontend/editor, nine extension unit
checks and current-SDK fresh/legacy manifest validation pass. Actions are disabled
on both repositories. Source/release/channel and both normal installs follow;
no owner install or public release11 claimed at this checkpoint.
