package social

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/my-social/internal/uploader"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type App struct {
	H                                        *apphost.Host
	storageMu                                sync.Mutex
	publishMu                                sync.Mutex
	importMu                                 sync.Mutex
	instagramMu                              sync.Mutex
	nativeMu                                 sync.Mutex
	instagramCancel                          map[string]context.CancelFunc
	Root, Guard, GuardToken, FFmpeg, FFprobe string
	now                                      func() time.Time
	limitsMu                                 sync.Mutex
	limits                                   map[string][]time.Time
}

func New(h *apphost.Host) (*App, error) {
	// Migrate rejects downgrades, so do not replay an earlier migration once a
	// later one has committed. Always call the latest migration for that guard.
	var current int
	if e := h.DB.QueryRow("SELECT coalesce(max(version),0) FROM host_migrations").Scan(&current); e != nil {
		return nil, e
	}
	if current < 1 {
		if e := h.Migrate(1, schema); e != nil {
			return nil, e
		}
	}
	if current < 2 {
		if e := h.Migrate(2, `CREATE TABLE profile_photos(user_id TEXT PRIMARY KEY REFERENCES profiles(user_id),media_id TEXT NOT NULL REFERENCES media(id));`); e != nil {
			return nil, e
		}
	}
	if current < 3 {
		if e := h.Migrate(3, instagramJobSchema); e != nil {
			return nil, e
		}
	}
	if e := h.Migrate(4, instagramNativeSchema); e != nil {
		return nil, e
	}
	a := &App{H: h, Root: os.Getenv("MYSOCIAL_PHOTOS_ROOT"), Guard: os.Getenv("MYSOCIAL_STORAGE_GUARD_URL"), GuardToken: os.Getenv("MYSOCIAL_STORAGE_GUARD_TOKEN"), FFmpeg: os.Getenv("MYSOCIAL_FFMPEG"), FFprobe: os.Getenv("MYSOCIAL_FFPROBE"), now: time.Now, limits: map[string][]time.Time{}}
	a.instagramCancel = map[string]context.CancelFunc{}
	if e := a.recoverJobs(); e != nil {
		return nil, e
	}
	if e := uploader.AutoInstall(h.Config.Version, h.Config.Origin, h.Config.WorkspaceURL, h.Config.Issuer, h.Config.Owner); e != nil {
		h.Log.Warn("Windows uploader setup needs attention")
	}
	a.routes()
	return a, nil
}
func uid(r *http.Request) string {
	if id := apphost.Current(r).User.ID; id != "" {
		return id
	}
	u, _ := r.Context().Value(instagramOwnerKey{}).(apphost.User)
	return u.ID
}
func encode(v any) string   { b, _ := json.Marshal(v); return string(b) }
func validID(s string) bool { return regexp.MustCompile(`^[a-f0-9]{48}$`).MatchString(s) }
func (a *App) fail(w http.ResponseWriter, e error) {
	a.H.Log.Error("social operation failed", "type", fmtType(e))
	apphost.Problem(w, 503, "This operation could not finish. Check My Photos and retry; nothing has been published.")
}
func fmtType(e error) string {
	if e == nil {
		return ""
	}
	return "operation"
}
func (a *App) route(method, path string, fn http.HandlerFunc) {
	a.H.Mux.HandleFunc(method+" "+path, a.H.Require(func(w http.ResponseWriter, r *http.Request) {
		a.ensureProfile(apphost.Current(r).User)
		key, limit := uid(r), 120
		if method == "PUT" && path == "/api/imports/instagram/jobs/{id}/chunks" {
			key += ":instagram-chunks"
			limit = 512
		}
		if r.Method != "GET" && !a.allowedLimit(key, limit) {
			w.Header().Set("Retry-After", "60")
			apphost.Problem(w, 429, "Too many changes. Wait a minute and retry.")
			return
		}
		fn(w, r)
	}))
}
func (a *App) allowed(user string) bool {
	return a.allowedLimit(user, 120)
}
func (a *App) allowedLimit(user string, limit int) bool {
	a.limitsMu.Lock()
	defer a.limitsMu.Unlock()
	now := a.now()
	xs := a.limits[user]
	kept := xs[:0]
	for _, v := range xs {
		if now.Sub(v) < time.Minute {
			kept = append(kept, v)
		}
	}
	if len(kept) >= limit {
		return false
	}
	a.limits[user] = append(kept, now)
	return true
}
func (a *App) routes() {
	a.route("POST", "/api/imports/instagram/native", a.createNativeUpload)
	a.route("GET", "/api/imports/instagram/native", a.nativeUploadInfo)
	for _, method := range []string{"GET", "POST", "PUT"} {
		a.H.Mux.HandleFunc(method+" /api/imports/instagram/native/{grant}/{action}", a.nativeUploadRequest)
	}
	a.route("POST", "/api/imports/instagram/jobs", a.createInstagramJob)
	a.route("GET", "/api/imports/instagram/jobs", a.listInstagramJobs)
	a.route("GET", "/api/imports/instagram/jobs/{id}", a.getInstagramJob)
	a.route("PUT", "/api/imports/instagram/jobs/{id}/chunks", a.uploadInstagramChunk)
	a.route("POST", "/api/imports/instagram/jobs/{id}/{action}", a.instagramJobAction)
	a.route("POST", "/api/imports/instagram/{action}", func(w http.ResponseWriter, r *http.Request) {
		if action := r.PathValue("action"); action != "preview" && action != "apply" {
			http.NotFound(w, r)
			return
		}
		a.instagram(w, r)
	})
	a.route("GET", "/api/social", a.bootstrap)
	a.route("GET", "/api/feed", a.feed)
	a.route("GET", "/api/posts/{id}", a.getPost)
	a.route("POST", "/api/posts", a.publish)
	a.route("PATCH", "/api/posts/{id}", a.editPost)
	a.route("POST", "/api/posts/{id}/{action}", a.postAction)
	a.route("GET", "/api/posts/{id}/comments", a.comments)
	a.route("POST", "/api/posts/{id}/comments", a.addComment)
	a.route("GET", "/api/profiles", a.profiles)
	a.route("GET", "/api/profiles/{id}", a.profile)
	a.route("PATCH", "/api/profile", a.editProfile)
	a.route("PUT", "/api/profile/photo", a.setProfilePhoto)
	a.route("DELETE", "/api/profile/photo", a.removeProfilePhoto)
	a.route("POST", "/api/profiles/{id}/relations", a.relation)
	a.route("GET", "/api/relations", a.relations)
	a.route("POST", "/api/media", a.upload)
	a.route("GET", "/api/media", a.library)
	a.route("GET", "/api/media/{id}", a.serveMedia)
	a.route("GET", "/api/drafts", a.drafts)
	a.route("PUT", "/api/drafts/{id}", a.saveDraft)
	a.route("DELETE", "/api/drafts/{id}", a.deleteDraft)
	a.route("GET", "/api/highlights", a.highlights)
	a.route("POST", "/api/highlights", a.addHighlight)
	a.route("GET", "/api/notifications", a.notifications)
	a.route("POST", "/api/notifications/read", a.readNotifications)
	a.route("GET", "/api/events", a.events)
	a.route("GET", "/api/preferences", a.preferences)
	a.route("PUT", "/api/preferences", a.setPreferences)
	a.route("POST", "/api/push/subscribe", a.subscribe)
	a.route("DELETE", "/api/push/subscribe", a.unsubscribe)
	a.route("POST", "/api/push/test", a.testPush)
	a.route("POST", "/api/jobs", a.createJob)
	a.route("GET", "/api/jobs", a.jobs)
	a.route("POST", "/api/reports", a.report)
	a.H.Mux.HandleFunc("GET /api/moderation", a.H.Owner(a.moderation))
	a.H.Mux.HandleFunc("POST /api/moderation/{id}", a.H.Owner(a.moderate))
	a.route("GET", "/api/export", a.exportAccount)
	a.route("POST", "/api/account/delete", a.deleteAccount)
}
func (a *App) bootstrap(w http.ResponseWriter, r *http.Request) {
	apphost.JSON(w, 200, map[string]any{"profile": a.getProfile(uid(r)), "photosConnected": a.storageCheck(r.Context()) == nil, "videoRendering": a.FFmpeg != "" && a.FFprobe != "", "pushKey": os.Getenv("MYSOCIAL_VAPID_PUBLIC"), "chatURL": a.H.Config.WorkspaceURL + "/_workspace/open/chat", "capabilities": map[string]bool{"live": false, "chatAdapter": false, "autoCaptions": false, "licensedMusic": false, "payments": false}})
}
func (a *App) Run(ctx context.Context) error {
	done := make(chan struct{})
	go func() { defer close(done); a.instagramWorker(ctx) }()
	defer func() { <-done }()
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			a.tick(ctx)
		}
	}
}
func (a *App) tick(ctx context.Context) {
	a.H.Freeze.RLock()
	defer a.H.Freeze.RUnlock()
	now := a.now().Unix()
	a.publishScheduled(ctx)
	a.H.DB.Exec("UPDATE posts SET state='archived' WHERE kind='story' AND state='published' AND expires<=?", now)
	a.runJob(ctx)
	a.deliverPush(ctx)
}
func safeRoot(root string) error {
	if !filepath.IsAbs(root) || !strings.EqualFold(filepath.Base(root), "My Social") || !strings.EqualFold(filepath.Base(filepath.Dir(root)), "My Photos") {
		return errors.New("choose My Photos/My Social")
	}
	for p := filepath.Clean(root); ; p = filepath.Dir(p) {
		if info, e := os.Lstat(p); e == nil && info.Mode()&os.ModeSymlink != 0 {
			return errors.New("linked storage is not allowed")
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	return nil
}
