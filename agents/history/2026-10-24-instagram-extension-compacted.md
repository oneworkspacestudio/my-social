# Instagram capture extension — implementation in progress

User explicitly requests a Chrome extension inside My Social, maximum practical
capture from an open Instagram tab, ZIPs compatible with the existing importer,
and commit/push/release/install. Base cd2fa9f with installed/public Social5.
Canonical source I:/Projects/my-social; workspace mirror currently same base.

Implementation in extensions/instagram-capture: MV3 action opens an export tab;
user-started capture, optional bounded scroll and sequential normal post-page
GETs from the source tab, then selection and local ZIP downloads. Captures
hydrated JSON/JSON-LD/visible article/preview media without private API, cookies,
credentials, raw HTML/script export or uploads. Only Instagram pages and CDN
hosts, no all-sites/history/debugger permission. Source security/DM pages blocked.

AUI screen and generated TSX, React control logic, standalone ZIP/CRC writer,
capture reports, original dates or explicit capture-time fallback, full captions
in report/truncated import caption when needed, 10-media carousel splitting,
sub-400MiB self-contained ZIP parts, 100MiB/file and 350MiB/post capture budget.
Incomplete carousels excluded from native manifests; recovered files retained
under capture/incomplete. Existing server import limits still apply (90s video).
Nothing captured automatically on extension install. No personal content used
as a fixture. User initiates capture/export after installation.

Releasing Social6 with extension0.1.0 and an importer download link. A separate
extension-only GitHub release would obstruct the installed Workspace updater's
newest-release scan; preserve the managed manifest/native package release flow.
No new schema. Web/scripts/build-instagram-extension.mjs uses current web deps;
CI builds/types/tests extension and installs Chromium for its synthetic E2E.

Local ZIP tests5 passed. Initial I: browser download lock failed on exFAT; exact
Playwright Chromium now downloaded into named NTFS temporary cache under
C:/Projects/.cache/social-extension-20261024/browsers (required for its lock).
Keep build/evidence on I:/Projects/tooling/social-extension-20261024. Clean only
that named C fixture tree after browsers/server/tests exit; never owner data.
Current browser harness tests real MV3 in an isolated profile with synthetic
Instagram/CDN routes and imports its downloaded ZIP into a fresh native Social.
Initial missing-browser test failed before Chromium installation; existing3
app regression flows passed. Ongoing test exec session51001; inspect its result.

Real extension E2E now passes with the existing public native Social5 importer;
all4 browser workflows passed. Initial synthetic page omitted UTF-8 charset,
causing mojibake; corrected the fixture and verified emoji preservation end to end.
Phone/desktop screenshots reviewed. ZIP unit tests5 pass. Browser binaries are
being moved from the named NTFS cache to I tooling after successful download
(the I download lock failed, but running completed binaries from I is supported).

Next: final Social6 release/package and exact-executable tests, source/CI push,
public immutable assets, Chrome unpacked installation via normal UI.
The requested install is the Chrome extension. Existing installed Social5 already
accepts its ZIPs; a server update is not needed for that browser workflow. New6
also publishes a convenient extension download link in the app. C drive has less
than1GiB free before moving the task browser cache; F NTFS is also full. Do not
attempt a native owner update without enough NTFS rollback capacity, remove
unknown data/backups, or put unencrypted owner backups on exFAT. Browser extension
files are small and can be installed independently of a server update.
Owner Workspace is now80 (not74); inspect current public manager/installer/bundle
and all concurrent native updaters before installation. Do not downgrade it or
reuse the old74 installation driver unchanged. Existing Social5 must stay intact.
Use the computer-use skill sky API for native Chrome settings/file-picker UI.
Chrome window selected from list_apps has app Chrome, returned window197194;
refresh selection before input. No auth/security-barrier bypass, registry forced
extension installs, profile preference tampering or disabling protections.
Install files using verified ZIP into stable LocalAppData/MySocial/InstagramCapture;
then load that exact folder unpacked and verify the extension card/version.
No source commit or public release yet at this handoff; do not claim installation.
