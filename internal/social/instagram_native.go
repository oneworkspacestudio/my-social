package social

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/oneworkspacestudio/my-social/internal/uploader"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

const instagramNativeSchema = `CREATE TABLE instagram_native_uploads (
 id TEXT PRIMARY KEY, owner TEXT NOT NULL REFERENCES host_users(id), token TEXT NOT NULL,
 client TEXT NOT NULL DEFAULT '', job TEXT NOT NULL DEFAULT '', expires INTEGER NOT NULL);
 CREATE INDEX instagram_native_owner ON instagram_native_uploads(owner,expires);`

func (a *App) nativeUploadInfo(w http.ResponseWriter, r *http.Request) {
	apphost.JSON(w, 200, map[string]any{"installed": uploader.Installed(a.H.Config.Origin), "version": a.H.Config.Version, "download": "https://github.com/oneworkspacestudio/my-social/releases/download/v0.1.0-preview.9/MySocialUploader.exe"})
}
func (a *App) createNativeUpload(w http.ResponseWriter, r *http.Request) {
	if a.H.Config.WorkspaceURL != "" && os.Getenv("MYSOCIAL_NATIVE_UPLOADS") != "1" {
		apphost.Problem(w, 409, "Update Workspace before starting a background upload")
		return
	}
	base := strings.TrimRight(a.H.Config.Origin, "/")
	if _, e := uploader.ValidBase(base); e != nil {
		apphost.Problem(w, 503, "The native uploader needs a secure Social address")
		return
	}
	a.nativeMu.Lock()
	defer a.nativeMu.Unlock()
	now := a.now().Unix()
	a.H.DB.Exec("DELETE FROM instagram_native_uploads WHERE expires<?", now)
	var count int
	if e := a.H.DB.QueryRow("SELECT count(*) FROM instagram_native_uploads WHERE owner=? AND job=''", uid(r)).Scan(&count); e != nil {
		a.fail(w, e)
		return
	}
	if count >= 4 {
		apphost.Problem(w, 429, "Finish opening the Windows uploader before starting another")
		return
	}
	id, token := apphost.ID(), apphost.ID()
	_, e := a.H.DB.Exec("INSERT INTO instagram_native_uploads(id,owner,token,expires) VALUES(?,?,?,?)", id, uid(r), apphost.Hash(token), now+600)
	if e != nil {
		a.fail(w, e)
		return
	}
	link := "mysocial-upload://start/?" + url.Values{"server": {base}, "ticket": {id + "." + token}}.Encode()
	apphost.JSON(w, 201, map[string]any{"id": id, "launch": link, "expiresIn": 600})
}

// A grant authorizes one archive transfer, never sign-in, publication, arbitrary
// files, another job, or general Social API access. Browser cookies are ignored.
func (a *App) nativeUploadRequest(w http.ResponseWriter, r *http.Request) {
	action := r.PathValue("action")
	if !((action == "claim" || action == "complete" || action == "problem") && r.Method == "POST" || action == "status" && r.Method == "GET" || action == "chunks" && r.Method == "PUT") {
		http.NotFound(w, r)
		return
	}
	if r.Header.Get("Origin") != "" || !validID(r.PathValue("grant")) {
		apphost.Problem(w, 403, "Native upload request rejected")
		return
	}
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !validID(token) {
		apphost.Problem(w, 401, "Upload permission is invalid or expired")
		return
	}
	if !a.allowedLimit("native-auth", 4096) {
		apphost.Problem(w, 429, "Wait before retrying the native upload")
		return
	}
	a.H.Freeze.RLock()
	defer a.H.Freeze.RUnlock()
	a.nativeMu.Lock()
	defer a.nativeMu.Unlock()
	var owner, initial, client, job string
	var expires int64
	e := a.H.DB.QueryRow("SELECT owner,token,client,job,expires FROM instagram_native_uploads WHERE id=?", r.PathValue("grant")).Scan(&owner, &initial, &client, &job, &expires)
	expected := client
	if action == "claim" {
		expected = initial
	}
	if e != nil || expires < a.now().Unix() || expected == "" || !apphost.Equal(expected, apphost.Hash(token)) {
		apphost.Problem(w, 401, "Upload permission is invalid or expired. Start a new background upload in Social.")
		return
	}
	ctx, e := a.instagramOwnerContext(r.Context(), owner)
	if e != nil {
		apphost.Problem(w, 403, "This upload account is no longer active")
		return
	}
	r = r.WithContext(ctx)
	if !a.allowedLimit(owner+":native-chunks", 512) {
		w.Header().Set("Retry-After", "60")
		apphost.Problem(w, 429, "Wait a minute before continuing the upload")
		return
	}
	if action == "claim" {
		next := r.Header.Get("X-Upload-Client")
		if !validID(next) || (client != "" && !apphost.Equal(client, apphost.Hash(next))) {
			apphost.Problem(w, 403, "This upload permission has already been claimed")
			return
		}
		var in instagramUploadInput
		if !apphost.Decode(w, r, &in) {
			return
		}
		var j instagramJob
		if job != "" {
			j, e = a.loadInstagramJob(job, owner)
			if e == nil && (j.Fingerprint != in.Fingerprint || j.Size != in.Size || j.Filename != in.Filename) {
				apphost.Problem(w, 409, "The selected ZIP changed")
				return
			}
		} else {
			var status int
			j, status, e = a.allocateInstagramJob(ctx, owner, in)
			if e != nil {
				apphost.Problem(w, status, e.Error())
				return
			}
		}
		if e != nil {
			a.fail(w, e)
			return
		}
		e = a.changeInstagramJob(j.ID, owner, func(j *instagramJob) {
			j.Native = true
			j.NativeGrant = r.PathValue("grant")
			j.Label = "Uploading ZIP in the Windows background uploader"
		})
		if e == nil {
			_, e = a.H.DB.Exec("UPDATE instagram_native_uploads SET client=?,job=?,expires=? WHERE id=?", apphost.Hash(next), j.ID, a.now().Add(24*time.Hour).Unix(), r.PathValue("grant"))
		}
		if e != nil {
			a.fail(w, e)
			return
		}
		j, _ = a.loadInstagramJob(j.ID, owner)
		apphost.JSON(w, 200, nativeJobView(j))
		return
	}
	if job == "" {
		apphost.Problem(w, 403, "Choose a ZIP in the Windows uploader first")
		return
	}
	r.SetPathValue("id", job)
	w = nativeReceiptWriter{w}
	switch action {
	case "status":
		j, e := a.loadInstagramJob(job, owner)
		if e != nil {
			http.NotFound(w, r)
			return
		}
		apphost.JSON(w, 200, nativeJobView(j))
	case "chunks":
		a.uploadInstagramChunk(w, r)
	case "problem":
		_ = a.changeInstagramJob(job, owner, func(j *instagramJob) {
			if j.State == "uploading" {
				j.Error = "The Windows uploader needs attention. Choose the original ZIP again or cancel this upload and start another."
				j.Label = "Background upload paused"
			}
		})
		j, _ := a.loadInstagramJob(job, owner)
		apphost.JSON(w, 200, nativeJobView(j))
	case "complete":
		r.SetPathValue("action", "complete")
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		a.instagramJobAction(w, r)
	}
}
func nativeJobView(j instagramJob) map[string]any {
	return map[string]any{"id": j.ID, "state": j.State, "received": j.Received, "size": j.Size, "chunks": j.Chunks, "chunkSize": instagramChunkSize}
}

// Keep the grant API's responses free of captions, media and import results.
// Chunk/action responses use this writer to retain only the transfer receipt.
type nativeReceiptWriter struct{ http.ResponseWriter }

func (w nativeReceiptWriter) Write(b []byte) (int, error) {
	var v map[string]json.RawMessage
	if json.Unmarshal(b, &v) != nil {
		return w.ResponseWriter.Write(b)
	}
	for k := range v {
		if k != "error" && k != "id" && k != "state" && k != "received" && k != "size" && k != "chunks" && k != "chunkSize" {
			delete(v, k)
		}
	}
	out, e := json.Marshal(v)
	if e != nil {
		return 0, e
	}
	_, e = w.ResponseWriter.Write(out)
	return len(b), e
}
