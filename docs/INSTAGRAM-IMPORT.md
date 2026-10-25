# Import an Instagram export

Open **Settings → Import Instagram ZIP**. In Instagram's Accounts Center,
request an export in **JSON** format, including your posts and media. Upload
the downloaded ZIP and review a sample of up to five posts with captions, dates,
carousel sizes and availability. Choose an audience and select **Import N posts**
to import all available posts, including those beyond the sample. Upload bytes, checking counts, import counts and elapsed time show progress.
On Windows, select **Choose ZIP for background upload** and allow the Windows
uploader to open. Choose the ZIP in the native file picker. You can then close
the browser: the native uploader sends the file independently and retries broken
connections automatically. Keep the computer on and the ZIP in place. Return to
Social for progress, the five-post sample and the Import button. The uploader
cannot publish posts; you still choose an audience and import in Social.

Workspace89 and later installs the uploader on its Windows PC. On another
Windows computer, use **Install Windows uploader on this computer**, run the
public MySocialUploader.exe once and return to Social. A new destination requires
native confirmation before selecting a ZIP. Only HTTPS Social addresses are
accepted, with loopback HTTP reserved for local installations. The Windows
uploader stores its queue encrypted for the current Windows user and resumes
pending transfers at the next Windows sign-in. Files moved or changed require
selection again; canceling the Social job stops the native transfer.

**Or upload in this tab** remains available on other devices. Those browser
uploads use resumable 2 MiB chunks and require the page to stay open until upload
finishes. If interrupted, choose the same ZIP to resume saved chunks; checksums
are verified first. **Pause upload** keeps saved chunks.

After upload, checking and importing run as persistent server jobs. You can leave
the page or close the tab and return to the progress and results. A server restart
resumes interrupted processing without duplicating completed posts. One Import
click processes every available post directly from the stored ZIP; the ZIP is
not uploaded again. **Cancel import** stops that job and keeps completed posts.
The default audience is **Only me**. You can view imported posts on your profile
and manage them in Creator Studio. Stories and archived posts remain archived;
importing an old Story does not publish it again for another 24 hours.

The importer supports numbered `posts_N.json` manifests under `content/`,
`your_instagram_activity/content/` and `your_instagram_activity/media/`, including
an enclosing export directory. It also accepts `reels.json` (`ig_reels_media`),
`stories.json` (`ig_stories`) and `archived_posts.json` (`ig_archived_posts`).
All numbered manifest parts in the ZIP are read. Export layouts vary; unknown
structures are rejected rather than guessed. HTML exports are unsupported.

Captions, original timestamps and carousel ordering are preserved. Instagram's
historical Latin-1 encoding of UTF-8 captions is repaired when unambiguous.
Images are sanitized into JPEG derivatives. Animated GIFs become still images.
MP4/WebM videos are transcoded to MP4 with their aspect ratio and optional audio;
the owner must configure FFmpeg and FFprobe. Video duration is checked during
import, with a 10-minute limit for imported clips and no silent trimming. Mixed image/video
carousels are supported. No external media URLs are fetched.

Limits: 512 MiB ZIP, 2 GiB expanded ZIP, 20,000 entries, 16 MiB combined content
metadata, 1,000 posts in a ZIP, 1,000 posts per server job (the older multipart API keeps its 100-post request limit), 1–10 media items
per post, 100 MiB per original, 24 megapixels per image, and the existing 2 GiB
per-account preview media quota. Export smaller date ranges for larger accounts.
ZIP64, encrypted and multi-volume archives are unsupported. Media absent from a
split export is reported as unavailable; each ZIP needs the manifests and media
for the posts it imports. Multiple ZIP parts cannot currently be merged in-app.

Each post and all its media commit together. The results distinguish successful,
duplicate and failed posts; successful posts remain saved if another fails.
Retrying the same content skips completed imports without changing their audience.
Duplicate identity includes the content hashes, original date, caption and kind,
so re-uploading/repacking the same export is safe. Editing those fields can make
the content a different import. Imports do not create historical notifications.
Return to the import page to read server results. Failed server checks can retry
the stored ZIP; failed posts can retry without uploading again. Jobs have a
12-hour processing budget per run and a 10-minute conversion budget per video.
Account deletion cancels its jobs; authenticated immutable ownership and storage
isolation are checked for every upload and background import.

The ZIP stays inside the guarded My Photos/My Social adapter and is never
extracted using archive paths or exposed through a download URL. Upload and ready
ZIPs expire after 24 hours without activity. Successful imports delete their ZIP;
archives with failed posts remain available for retry for up to 24 hours. Results
are retained briefly (up to eight days including cleanup). Two retained ZIPs per
account and four globally bound staging storage to at most 2 GiB.

SQLite migration 4 stores resumable upload receipts, checksum lists, job progress
and results. Migration4 adds hashed, expiring, single-file native upload permissions.
The native client has no general API or account/session access. Jobs pause at post boundaries for normal backup/checkpointing.
Existing per-post atomic commits, quota, identity, CSRF and storage guards remain.
Completed imports deduplicate after an interrupted process. Older request-bound
API clients still work, with their existing request lifetime and 15-minute budget.

Followers, follow relationships, likes, comments, private messages, profile
details, stickers, music licences and original Instagram URLs are not migrated.
Only media referenced by the recognized content manifests is imported. Import
your own exports; a filename is not proof of ownership of someone else's work.

Verification uses synthetic JSON/ZIPs and media, including native FFmpeg checks
when configured. `go test ./...`, `go vet ./...`, frontend checks and the Chrome
import flow cover preview/no-write behavior, account boundaries, duplicate retry,
carousel rollback, malformed archives, limits, interruption recovery and mobile
layout. Published browser fixtures use synthetic data; additional authorized import
and installation checks are recorded in [server import acceptance](instagram-server-jobs-acceptance.json).
Physical-phone acceptance remains open.
