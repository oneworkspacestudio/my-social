# Profile photos

User requests Social profile-picture support, source push, release and local install.
Base fdc8bab / public preview2. Branch codex/profile-photos. Build artifacts onI in
named social-profile-20261023; isolated NTFS fixtures in matching named C cache.
No owner content or account used as test fixture. Preserve simultaneous suite,
My3D, Drive and Slicer work; never downgrade the installed Workspace manager.

Added additive migration2 profile_photos table; old profiles/posts preserved and
old app schema remains readable. PUT/DELETE /api/profile/photo binds only an owned,
ready avatars image. Existing protected photo adapter stores originals; only the
current sanitized512px JPEG derivative is visible to signed-in unblocked members.
Replacing/removing revokes profile sharing; originals remain private for retention.
Disabled account deletion clears binding. Storage guard remains fail-closed.
Browser normalizes orientation and square crop; server re-encodes/strips metadata,
validates originals,10MiB and24MP limits. No external image URLs are accepted.

AUI source/generated screens add Edit profile & photo, selection/preview/save,
replace/remove, profile header, post/suggestion/explore avatars and account button.
File-input selection is reset so the same file can be selected again; avatar file
is excluded from unrelated device drafts. Initials remain as fallback.

Local Go tests/vet and TypeScript/Vite build passed. Actual Windows candidate
passed existing camera/carousel/draft/publish/Story/PWA regression and new390px
photo upload/reload/replace/remove checks. Synthetic screenshots visually reviewed.
No physical-phone, clean-host/reboot or new production-security certification.
Source03a20b3 was pushed main with public release3; Windows CI39982593678 passed.
Normal protected Social-only offline update installed exact tested bytes under
Workspace73, retaining identity, paths, ports and other app packages. Initial
startup was healthy. Metadata-only3.1 reissue removes untracked social-contract
from the update manifest; exact Workspace73 validator accepts the four-package
manifest and rejects the original. Never mutate published release assets.

Subsequent Slicer update exposed a restart bug introduced here: New replayed
Migrate(1) before Migrate(2), and apphost rejects version1 once version2 exists.
Social3 executable hash remained correct; do not downgrade or edit owner data.
Release4 repair reads current migration version, applies initial schema only when
needed, always invokes latest migration to retain the future-version guard.
Additive migration test now closes/reopens the database and checks profile/version;
native browser harness also restarts the exact executable against its same fixture.
Go tests/vet pass. Next: build/test/publish release4 with four-package metadata,
coordinate owner updater hold with concurrent Slicer chat, apply normal protected
Social-only update after its updater exits, verify same-data restart and service
health. No actual personal photo was supplied; feature lets user choose one.
