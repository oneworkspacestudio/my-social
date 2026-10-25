# Changelog

## 0.1.0-preview.10

- Verify and pair the immutable receiving account before opening the native file picker.
- Windows background ZIP uploads keep transferring after the browser closes, retry connection failures and resume the encrypted queue at sign-in.
- Native ZIP picker, scoped single-file permissions and live server progress retain the five-post sample and explicit Import action.
- Requires Workspace90 for the dedicated authenticated native transfer route. Browser uploads remain available on other devices.


## 0.1.0-preview.8

- Upload Instagram ZIPs in resumable, checksum-verified chunks.
- Run checking and imports as persistent server jobs that continue after leaving the page and resume after server restart.
- Keep the five-post sample, import all available posts from the stored ZIP, and restore progress/results when returning.
- Support imported videos up to ten minutes; retain completed posts on cancellation and retries.
- Add migration 3 for private job state without changing existing accounts or media.

## 0.1.0-preview.7

- Show upload bytes, ZIP checking and import progress with elapsed time and Stop.
- Preview at most five posts and import all available posts with one click.
- Automatically batch larger imports and keep completed posts safe on retry.

## 0.1.0-preview.6

- Add a personal-use Chrome extension that captures accessible Instagram tab content into self-contained My Social import ZIPs.
- Include optional scrolling/post-page enrichment, selection, local downloads and capture reports.
- Link the extension download from the Instagram importer.

## 0.1.0-preview.5

- Import Instagram JSON export ZIPs from Settings with preview, selection and an Only me default.
- Preserve captions, original dates and carousels; support bounded videos and archived Stories.
- Skip duplicate imports on retry and recover interrupted media writes.
- Keep protected originals, existing accounts and profile photos intact.

## 0.1.0-preview.4

- Keep profile photos and existing profiles available after restarting or updating another Workspace app.
- Correct automatic-update package ownership metadata.

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

## Instagram Capture 0.1.1

Fix username-prefixed Instagram post/Reel URLs being missed. Target a selected
profile's own uploads, scroll older uploads by default, inspect rendered post
pages/carousel slides when HTML is only a shell, show media previews/counts and
clearly locate downloaded photo/video ZIPs. Preview-only media is reported as
incomplete and kept out of native import manifests.
