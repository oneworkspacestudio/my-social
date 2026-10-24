# Privacy

Capture begins only after clicking Scan and download my uploads. The extension reads the
selected Instagram tab and, if enabled, scrolls it and requests discovered post
pages using that tab's normal Instagram session. Media downloads omit credentials
and use only HTTPS Instagram CDN hosts. Captures are kept in memory in the export
tab and, by default, downloaded immediately after each post is read. Local ZIP
parts save every 10 uploads or 400 MiB, plus at completion or Stop. Optional manual
review mode saves only after clicking Download my media ZIPs. Closing the export tab clears
its capture state. Already downloaded files remain under your control.

Nothing is sent to My Social or another server by this extension. Importing a ZIP
into your own My Social server is a separate action. There is no analytics,
telemetry, account login, remote executable code or extension storage database.
Cookies, credentials, access tokens, raw page HTML/scripts and private messages
are not exported. Signed media URL query strings are not written into reports.

Permissions:

- activeTab/scripting: read the chosen Instagram tab when invoked.
- Instagram host access: read accessible post pages from that tab.
- Instagram CDN host access: retrieve referenced media files.
- downloads: save ZIPs and wait for their completion before releasing buffers.

No all-sites, cookies, history, debugger or external-messaging permission.

The own-uploads scan limits candidates to the selected username and profile grids.
It may open one temporary inactive Instagram tab for the Reels grid or rendered
post/carousel pages. That tab is closed after scanning, failure or cancellation.
It does not crawl feed recommendations, saved or tagged content.
