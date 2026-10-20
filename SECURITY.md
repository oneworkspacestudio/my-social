# Security and supported use

This unsigned evaluation preview is intended for small, trusted adult communities
on a single Windows host. Report a vulnerability privately to the repository
owner; do not include credentials or private originals in an issue.

MyAccounts provides OIDC code/PKCE sign-in. Application sessions are HttpOnly,
SameSite cookies, secure on HTTPS; identity is issuer plus immutable subject.
Every media request checks the current user and the post audience. Originals,
drafts, events, subscriptions and exports are owner scoped. Public means the
signed-in community; anonymous public media links are not implemented.

Managed originals reside in Private/photos-social/My Photos/My Social. Quantum's denyAll policy
is verified through a loopback HMAC-derived capability; conflicting ancestor or
descendant allows stop media access. Never expose this root as static files.
Administrators and Windows host owners are trusted and can read host storage.
Standalone storage must be explicitly marked isolated and never be a web root.

Uploads accept a bounded set of decoded image or probed video formats. Raster
renditions strip metadata, including GPS. Limits are 100 MiB/file, 24 MP/image,
10 media items/post, 2 GiB originals/account, 90 s/render and 120 writes/minute.
Video uses structured owned media IDs and numeric EDL fields, fixed argument
vectors and a timeout. No user-controlled shell command is executed. FFmpeg
must be operator supplied and updated. It is not an OS security sandbox.
Malware scanning, hostile-community moderation, appeals, child/teen compliance,
independent penetration review and service/reboot qualification remain gates.

The worker never publishes an incomplete upload/render. Idempotency keys prevent
post retry duplicates. Failed uploads are removed where possible; interrupted
filesystem/DB writes need reconciliation described in OPERATIONS.md.
No private responses or media enter the service-worker cache. Device drafts use
account-keyed IndexedDB and are cleared on sign-out/session expiry/account change.
Browser/Windows profile compromise is outside this local-storage boundary.

Push endpoint hosts are allowlisted. Payloads contain generic activity only;
delivery requires opt-in, a valid VAPID configuration, subscription and enabled
category. Permissions revocation and subscription failures retire endpoints.
Do not configure private network push endpoints or expose secrets to client JS.
