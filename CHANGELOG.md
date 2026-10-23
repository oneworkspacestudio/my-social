# Changelog

## 0.1.0-preview.2

Connect capability-based camera zoom to the actual video-track constraints.
Clear previous accounts' private device drafts when activating another account,
and fence stale tabs from writing a draft for the previous identity. Existing
server state is retained. Workspace48 is required for scoped private media.

Replace the original shared-folder ACL adapter: Windows path casing could bypass
its native Drive deny rules. Originals now use a dedicated private source outside
shared Files, with each MyAccounts subject bound to its own read-only native
scope. Existing legacy media fails closed and requires explicit migration.

## 0.1.0-preview.1

First functional Windows PWA preview: MyAccounts/native standalone identity,
guarded My Photos originals, camera/photo/video recording, layered photo editor,
carousel drafts and publishing, Story expiry/polls/Highlights, bounded Reel
render queue, feed/profiles/follow approval/likes/saves/comments, Studio,
moderation, export and opt-in notifications. See ACCEPTANCE.md for exact gates
and feature gaps. No full Instagram parity or production acceptance claim.

## 0.1.0-preview.3

Add, replace and remove a profile photo in Edit profile & photo. Photos appear on profiles, posts, suggestions and the account button. Images are cropped to a small square; original files stay protected. Existing profiles and posts are preserved.
