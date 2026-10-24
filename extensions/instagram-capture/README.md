# My Social Instagram Capture

A personal-use Manifest V3 Chrome extension. Open an Instagram profile, feed,
post, Reel or accessible Story, click the extension, then **Capture content**.
Optionally scroll for more content and read discovered post pages. Review the
captured posts and choose **Download import ZIPs**. Keep both tabs open.

Import each downloaded ZIP through My Social **Settings → Import Instagram ZIP**.
The ZIPs contain `content/posts_1.json`, Reel/Story manifests, local `media/posts/`
files, and `capture/report.json`. Stories remain archived. Comments/profile
details retained in the capture report are not recreated as native Social data.

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
change. A visible preview can be lower quality or omit a carousel/video. Open a
post if a profile grid does not provide complete metadata. User-facing notes
identify preview-only posts, missing dates and failed page reads. The official
Instagram JSON export remains the more complete account migration option.

Missing dates use the run's capture date; reports identify this explicitly.
Captions exceeding My Social's byte limit are shortened only in import manifests;
the full captured caption remains in the report. Carousels split at 10 items.
Downloads are split below 400 MiB with each ZIP's own complete manifest/media.
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
