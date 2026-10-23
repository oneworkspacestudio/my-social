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
Next: commit source, build exact release3 with notices/source pins, pass CI and
verify all public asset hashes, then use normal backed-up Social-only update.
Record actual installed version/hash and retain newest verified rollback.
