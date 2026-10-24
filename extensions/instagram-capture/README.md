# My Social Instagram Capture

A personal-use Manifest V3 Chrome extension that downloads **your own Instagram
uploads** as actual photo/video files in ZIP archives. Open your profile in
Instagram, click the extension and confirm the detected username. Choose **Scan my
uploads**. Scrolling older uploads and scanning your Reels grid are enabled by
default. Each discovered upload is inspected for photos, videos and carousel
slides; a temporary background tab is used when a normal page response is only a
shell. Feed, saved, tagged and other accounts' posts are excluded.

Review the thumbnails/counts, then click **Download my media ZIPs**. Scanning alone
does not download files. Chrome saves ZIPs in **Downloads/MySocial**. Unzip them and
open **media/posts/** for the photos/videos. The **Show downloaded ZIP in folder**
button reveals the download. Partial recoveries are under **capture/incomplete/**;
**capture/report.json** lists file paths and failures. A ZIP with zero media files
is explicitly reported as a failure report, not a backup.

The same ZIPs can also be imported through My Social **Settings → Import Instagram
ZIP**. Importing is the optional destination step after downloading from Instagram.
Comments/profile details in the report are not recreated as native Social data.
The file layout remains compatible with the existing importer.

## Installation

Download the extension ZIP from the My Social release, extract it to a permanent
folder, open `chrome://extensions`, turn on Developer mode, select **Load unpacked**
and choose that folder (the one with `manifest.json`). Pin the extension if useful.
Reload its card after replacing files for an update. This GitHub package is not a
Chrome Web Store listing and has no automatic browser update service.

## What can be captured

- Accessible hydrated JSON, JSON-LD, visible post elements and page-preview media.
- Best available image/video URLs, captions, original dates when available,
  carousel order, alt text, source links and bounded visible comments/profile text.
- Up to 1,000 posts, 100 scroll steps and 1,000 discovered post pages per run.
  Page reads are sequential and stop when Instagram returns HTTP 403/429.
- No login automation, cookies/passwords/tokens, private API, DRM/DASH assembly,
  messages, follower graph, hidden/unavailable posts or server uploads.

Capture is best effort, not an account backup. Dynamic Instagram layouts can
change. A visible preview can be lower quality or omit a carousel/video. Rendered upload pages and carousel controls are inspected when page HTML is incomplete.
Some player/blob videos still have no downloadable source; those are marked incomplete. User-facing notes
identify preview-only posts, missing dates and failed page reads. The official
Instagram JSON export remains the more complete account migration option.

Missing dates use the run's capture date; reports identify this explicitly.
Captions exceeding My Social's byte limit are shortened only in import manifests;
the full captured caption remains in the report. Carousels split at 10 items.
Downloads are split below 400 MiB with each ZIP's own complete manifest/media.
A thumbnail-only upload is marked incomplete and is not presented as a complete original.
Each media file is limited to 100 MiB and each captured post to 350 MiB. A failed
carousel is excluded from import manifests; recovered files remain under
`capture/incomplete/` and the report explains the failure. Native import still
checks image limits and the 90-second video limit. No video is silently trimmed.

## Build and verify

From the repository root after installing `web` dependencies:

```
node web/scripts/build-instagram-extension.mjs
node web/node_modules/typescript/bin/tsc -p extensions/instagram-capture/tsconfig.json
node --test extensions/instagram-capture/tests/archive.test.mjs
```

The AUI screen is authoritative; regenerate its TSX. Build output is `dist/` in
this directory and is ignored by Git. Extension browser checks use a fresh
Chromium profile and synthetic Instagram/media routes, then import the resulting
ZIP into a fresh My Social fixture. Never use owner installation data as fixtures.

Original My Social branding/code is Apache-2.0. Bundled React and build-dependency
notices accompany release packages. Instagram is a third-party service; this
extension is not affiliated with Meta. Capture content you own or may copy.

## 0.1.1 correction

Recognises both `/p/code/` and `/username/p/code/` links (and corresponding Reels),
including long codes. Defaults to an explicit profile scan, skips other owners,
shows photo/video previews and actual download locations, and distinguishes
incomplete previews from full media. No account-wide completeness claim is made.
