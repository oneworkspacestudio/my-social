# Installation and operation

One Workspace preview48+ installs My Social as an independent optional app.
Default listener 127.0.0.1:9013, route /social; installation-specific ports are
allocated by Workspace. MyAccounts registers /social/auth/callback and supplies
the confidential OIDC credentials only to the server process. Existing proxy
hosting carries the route; no new domain or DNS write is needed. HTTPS is
required for remote camera, recording and push.

Standalone: set MYSOCIAL_PHOTOS_ROOT to an absolute dedicated path ending in
`My Photos/My Social`, and set MYSOCIAL_STANDALONE_STORAGE=isolated only when it
is outside all shared/static sources. Start `MySocial.exe --data C:\Social\Private
--listen 127.0.0.1:9013 --origin https://social.example/social --no-browser` behind
a trusted HTTPS proxy. With no OIDC configuration the initial local-only setup
link creates an owner. MyAccounts variables are WORKSPACE_OIDC_ISSUER,
WORKSPACE_OIDC_CLIENT_ID, WORKSPACE_OIDC_CLIENT_SECRET and optionally
WORKSPACE_OWNER_SUBJECT. Never put secrets in CLI arguments or source files.
An absent storage adapter leaves the app running but refuses uploads/publication.

## Media helpers

MYSOCIAL_FFMPEG and MYSOCIAL_FFPROBE are absolute paths to centrally installed,
trusted matching executables. They are optional runtime helpers, not bundled
Node services. Choose and maintain a legally appropriate FFmpeg build; features
such as libx264 can change its licence obligations. No FFmpeg binary is
redistributed here. Video uploads/rendering stay unavailable until configured.
Operators must validate their build and resource capacity. Renders use two
threads, a three-minute timeout, 720×1280 H.264 and at most 90 seconds. Single
clips can retain original audio; multi-clip audio mixing, voice-over, licensed
music, timed video overlays and automatic transcription remain open gates.

## Push

Configure MYSOCIAL_VAPID_PUBLIC, MYSOCIAL_VAPID_PRIVATE and
MYSOCIAL_VAPID_CONTACT securely. The user must opt in from Settings. Chromium
Android and supported browsers depend on OS permission and push-service reachability.
On supported iOS/iPadOS, use an installed Home Screen web app. Real device
delivery has not been verified by desktop emulation. Quiet hours use server time.
Payloads say only that activity is available. The default is no push.

## Recovery, deletion and backups

Stop My Social before copying its private folder plus the entire configured
My Photos/My Social root. Restore them together to a new protected NTFS location,
then set the same adapter mapping. Workspace's normal stopped, verified central
backup includes both its Private and Files trees. The apphost standalone `.owb`
backup contains app state only; it is insufficient for the external media root.
Keep originals and database snapshots at the same version. No clean second-host
or service/reboot restore certification is claimed.

Upload intents precede filesystem writes; successful media rows and completion
receipts commit together. Failed uploads remove newly created files. Interrupted
intents remain tracked for reconciliation; do not delete arbitrary directories.
Review upload_intents with state=writing while the app is stopped; verify the
recorded owner/path, absence of a matching media row and backups before removing
that exact file/intent. Completed media is never deleted by this reconciliation.
Failed renders can be retried from the draft. Originals remain private; Story
expiry removes audience access at server time, independently of cleanup timing.

Archive/delete hides posts immediately. Studio permits restore; expired Stories
need a Highlight for permanent audience access. Retention is conservative: all
originals and server drafts stay until the owner performs an explicit backup-aware
retention operation. Account deletion disables Social sessions and hides posts;
it does not erase the MyAccounts identity or pretend retained originals are gone.
Use the user export before an explicitly agreed permanent erasure procedure.
No automatic storage reclamation policy is claimed in this preview.

## Capacity and troubleshooting

This is a single-host SQLite application, not an Instagram-scale service. Monitor
disk space, upload memory, render time and database latency. Stop accepting media
before disk fills; configure 2 GiB/account quota changes in a reviewed release.
An external media worker/object store, database migration and federation would
need new contracts. Check healthz, app.log and Workspace app logs for readiness.
Logs contain no original content or login tokens. An unhealthy My Photos guard
fails closed; restore the dedicated private source and per-account scope mapping.
Do not work around a guard failure by publishing the private directory.

## Scoped storage correction (Workspace48 / Social preview2)

Supersedes the preview1/Workspace47 shared-folder adapter. Native Quantum folder
ACL keys preserve case on Windows; alternate casing can bypass a deny rule.
Do not use preview1 for private uploads. No local owner media existed when this
was discovered. Published predecessor artifacts remain immutable for provenance.

The authoritative My Photos/My Social source now resides under
`Private/photos-social/My Photos/My Social`, outside every broad Files source.
`My Social Originals` is disabled by default. The server binds a verified
MyAccounts issuer/subject to `/users/<sha256(issuer + NUL + subject)>`, with native
view/download only. Native canonical physical scope containment protects other
accounts, including casing and traversal variants. Generated derivatives stay
outside all user scopes; Social applies post audiences. Shared My Drive remains
unchanged. Never add a broad source overlapping this private source.

Upgrades refuse a nonempty legacy Social folder: stop services, back up app state
and both source roots, and perform a separately reviewed migration. Empty legacy
markers may remain inert. Back up the new private source together with Social DB.
