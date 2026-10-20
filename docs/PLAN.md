# My Social implementation plan and contract review

The supplied brief authorizes staged, tested releases. This is an evaluation
preview for a small, trusted, adult community; full Instagram parity remains an
open programme. No milestone is certified by the presence of a navigation item.

1. Reuse One Workspace's actual Go apphost, MyAccounts OIDC, stable identity,
   SQLite WAL/migrations, CSRF/session checks, launcher and design tokens.
2. Reserve ordinary disk storage at Files/My Photos/My Social. Quantum's current
   native ACL API installs denyAll; reject specific ancestor/descendant allow
   overrides. A loopback capability guard checks this policy for media operations.
   Social exposes authenticated owner originals and audience-checked renditions.
   No broad source sharing or folder-name-only isolation is accepted.
3. Implement camera and layered Fabric editor, independent carousel scenes,
   device crash drafts, private server drafts and idempotent publication.
4. Add Stories with server time, polls and Highlights; bounded structured FFmpeg
   Reel jobs; basic profiles/follow approval/blocks/comments/saves/Explore/Studio.
5. Add durable per-user event delivery, SSE reconnection, opt-in VAPID push and a
   public-shell-only service worker. Device media never enters CacheStorage.
6. Test the native app, ACL policy and an isolated suite; release both repositories
   only after local tests and exact-commit Windows GitHub Actions pass. Verify
   public asset sizes and SHA256 digests. Never replace immutable release bytes.

## Inspected ecosystem contracts

- `one-workspace/pkg/apphost`: confidential authorization-code/PKCE login;
  issuer, audience, signature, nonce, expiry checks; HttpOnly scoped sessions;
  immutable issuer+subject identity; owner and CSRF enforcement; protected NTFS
  state. New explicit MediaCapture capability keeps other apps' camera policy.
- `filebrowser-drive/PHOTOS.md`, `backend/internal/web/photos.go`,
  `backend/internal/web/access.go`, `backend/internal/database/access/access.go`:
  Photos uses existing scoped Drive resources. ACL-specific allows override
  denyAll, so the adapter checks ancestors and descendants too. Existing Photos
  content is imported only through the user's current Drive session, after
  comparing its OIDC subject. My Social does not reuse recovery credentials in
  the browser or retain an ordinary user's Drive password/token.
- `workspace-chat/docs/workspace/HOSTING.md` and Tinode API: separate-origin
  gateway sessions and short-lived tickets. Opening My Chat is available;
  social DMs/calls need a reviewed identity/session adapter and are not simulated.
- `one-workspace/internal/suite/navigation`: Segoe UI, blue `#345de5`, pale
  `#f5f7fc`, ink `#18233b`; dark `#151c2b`/`#1e293b`, shared browser theme.
  Original Social SVG follows the suite's purple tile/mint detail family.

## Current official product and browser references

Reviewed 20 October 2026. These define comparison capabilities, not completion:

- [Meta: Instagram reposts, Map and Friends in Reels](https://about.fb.com/news/2025/08/new-instagram-features-help-you-connect/)
- [Meta: Instagram Reels](https://about.fb.com/news/2020/08/introducing-instagram-reels/)
- [MediaRecorder MIME negotiation](https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder/isTypeSupported_static)
- [WebKit: Home Screen Web Push](https://webkit.org/blog/13878/)
- [Apple: sending web push](https://developer.apple.com/documentation/usernotifications/sending-web-push-notifications-in-web-apps-and-browsers)
- [Fabric.js documentation and licence](https://github.com/fabricjs/fabric.js)
- [FFmpeg legal considerations](https://ffmpeg.org/legal.html)

Recommendations, reposts, the Friends activity feed, opt-in Map, broadcast
channels, advanced DMs, native-only functions, licensed music and commerce are
comparison gaps. No proprietary Instagram code, logo or media is included.
The UI uses original app assets and synthetic QA content only.
