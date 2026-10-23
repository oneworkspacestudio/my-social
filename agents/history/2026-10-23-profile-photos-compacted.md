# Profile photos

User requests Social profile-picture support, source push, release and local install.
Base fdc8bab / public preview2. Branch codex/profile-photos. Build artifacts onI in
named social-profile-20261023; isolated NTFS fixtures in matching named C cache.
No owner content or account used as test fixture. Preserve simultaneous suite,
My3D, Drive and Slicer work; never downgrade the installed Workspace manager.

Added additive migration2 profile_photos table; old profiles/posts preserved and
existing profile/post tables remain intact. Older executables must not downgrade the host schema. PUT/DELETE /api/profile/photo binds only an owned,
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
Release4 source a346fcbc94d5d62c26cef29ae93aac64ce1f7771 pushed main/tag.
Go module verification, tests/vet, frontend build/tests/notices passed. Windows CI
39987921383 and local native browser flows plus same-data executable restart pass.
Exact packaged/tested/installed SHA256:
ccb7b352c112ea1350c399e1c27795c37446b7b3fab0dcc9ed3a187667bd914b.
Public release4 eleven asset hashes/sizes verified; exact Workspace73 manifest
validator run uncached against actual public metadata accepted all four owned
packages and rejected the old unowned contract. Older3/3.1 release notes point4.

After coordinated Slicer update exited, normal protected offline Social-only
repair installed4, with a full consistent verified backup and actual updater
exit0. Kept Workspace73, Slicer5, every non-Social contract/package hash, identity,
paths, ports and selection. Final local/public services all running, exact served
frontend hashes match tested build, autostart and latest recovery manifest verified.
After Social updater exit, a separate authorized update installed Workspace74 and
Mail18. Final verification accepts only those specific catalog/package changes
against the verified public74 bundle. Social4 and Slicer5 actual executable hashes
remain correct; all25 services and public Workspace health are healthy under74.
No owner content used as fixtures; isolated previews stopped and own C caches
removed. No physical phone/reboot acceptance claimed. No personal photo supplied;
user chooses one at Profile -> Edit profile & photo. No remaining task work.
