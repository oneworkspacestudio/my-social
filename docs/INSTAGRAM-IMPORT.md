# Import an Instagram export

Open **Settings → Import Instagram ZIP**. In Instagram's Accounts Center,
request an export in **JSON** format, including your posts and media. Upload
the downloaded ZIP and review a sample of up to five posts with captions, dates,
carousel sizes and availability. Choose an audience and select **Import N posts**
to import all available posts, including those beyond the sample. Upload bytes,
checking counts, import counts and elapsed time show progress. **Stop** cancels
the current operation while keeping completed posts.
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
import, with a 90-second limit and no silent trimming. Mixed image/video
carousels are supported. No external media URLs are fetched.

Limits: 512 MiB ZIP, 2 GiB expanded ZIP, 20,000 entries, 16 MiB combined content
metadata, 1,000 posts in a ZIP, 100 posts per request (automatically batched after one Import click), 1–10 media items
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
Keep the page open during import; after a lost response or cancellation, preview
and retry the ZIP to find completed posts. The request has a 15-minute processing
budget, with a three-minute conversion budget per video.

The ZIP is temporarily spooled inside the guarded My Photos/My Social adapter
and removed when preview/import returns. Entries are never extracted to paths
from the ZIP. Interrupted media writes are recovered before the next import;
temporary ZIPs left by a process crash are removed on a subsequent import once
an hour old. The importer uses the existing account identity, CSRF checks,
media authorization, storage isolation guard and quota. Private originals are
never exposed as shared folder links. No database schema change is needed.

Followers, follow relationships, likes, comments, private messages, profile
details, stickers, music licences and original Instagram URLs are not migrated.
Only media referenced by the recognized content manifests is imported. Import
your own exports; a filename is not proof of ownership of someone else's work.

Verification uses synthetic JSON/ZIPs and media, including native FFmpeg checks
when configured. `go test ./...`, `go vet ./...`, frontend checks and the Chrome
import flow cover preview/no-write behavior, account boundaries, duplicate retry,
carousel rollback, malformed archives, limits, interruption recovery and mobile
layout. No personal export or physical-phone acceptance is implied.
