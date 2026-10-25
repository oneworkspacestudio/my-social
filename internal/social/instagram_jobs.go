package social

// Uploads are short, resumable requests. Checking and importing belong to the
// application worker, never to a browser request or its cancellation context.
import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

const instagramChunkSize = 2 << 20
const instagramJobSchema = `CREATE TABLE instagram_import_jobs (
 id TEXT PRIMARY KEY, owner TEXT NOT NULL REFERENCES host_users(id),
 state TEXT NOT NULL, data TEXT NOT NULL, created INTEGER NOT NULL, updated INTEGER NOT NULL, retained INTEGER NOT NULL DEFAULT 1);
 CREATE INDEX instagram_jobs_owner ON instagram_import_jobs(owner,created);
 CREATE INDEX instagram_jobs_state ON instagram_import_jobs(state,created);`

type instagramOwnerKey struct{}
type instagramChunk struct {
	Offset int64  `json:"offset"`
	Size   int64  `json:"size"`
	SHA    string `json:"sha256"`
}
type instagramResult struct {
	Key     string `json:"key"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
type instagramJob struct {
	ID          string            `json:"id"`
	Owner       string            `json:"owner"`
	State       string            `json:"state"`
	Filename    string            `json:"filename"`
	Size        int64             `json:"size"`
	Received    int64             `json:"received"`
	Fingerprint string            `json:"fingerprint"`
	Chunks      []instagramChunk  `json:"chunks"`
	Digest      string            `json:"digest"`
	Items       []instagramItem   `json:"items"`
	Keys        []string          `json:"keys"`
	Visibility  string            `json:"visibility"`
	Results     []instagramResult `json:"results"`
	Label       string            `json:"label"`
	Done        int               `json:"done"`
	Total       int               `json:"total"`
	Error       string            `json:"error"`
	Created     int64             `json:"created"`
	Updated     int64             `json:"updated"`
	Retained    bool              `json:"retained"`
	Revision    int64             `json:"revision"`
	NativeGrant string            `json:"nativeGrant"`
	Native      bool              `json:"native"`
}

func instagramArchive(id string) string { return "generated/imports/job-" + id + ".zip" }
func instagramActive(state string) bool {
	return state == "uploading" || state == "checking" || state == "ready" || state == "queued_import" || state == "importing"
}
func (a *App) loadInstagramJob(id, owner string) (j instagramJob, e error) {
	var raw, state string
	e = a.H.DB.QueryRow("SELECT data,state FROM instagram_import_jobs WHERE id=? AND owner=?", id, owner).Scan(&raw, &state)
	if e == nil {
		e = json.Unmarshal([]byte(raw), &j)
		j.State = state
		if e == nil && (j.ID != id || j.Owner != owner || !validID(j.ID) || j.Size < 22 || j.Size > instagramZIPLimit || j.Received < 0 || j.Received > j.Size) {
			e = errors.New("invalid stored upload identity or size")
		}
	}
	return
}
func (a *App) saveInstagramJob(j *instagramJob) error {
	j.Updated = a.now().Unix()
	j.Revision++
	_, e := a.H.DB.Exec("UPDATE instagram_import_jobs SET state=?,data=?,updated=?,retained=? WHERE id=? AND owner=?", j.State, encode(j), j.Updated, j.Retained, j.ID, j.Owner)
	return e
}

// Callers hold the host's backup read lock; the mutex orders worker/API changes.
func (a *App) changeInstagramJob(id, owner string, change func(*instagramJob)) error {
	a.instagramMu.Lock()
	defer a.instagramMu.Unlock()
	j, e := a.loadInstagramJob(id, owner)
	if e != nil {
		return e
	}
	change(&j)
	return a.saveInstagramJob(&j)
}
func instagramJobView(j instagramJob) map[string]any {
	sample := j.Items
	if len(sample) > 5 {
		sample = sample[:5]
	}
	if sample == nil {
		sample = []instagramItem{}
	}
	ready, unavailable := 0, 0
	for _, i := range j.Items {
		if i.Status == "ready" {
			ready++
		}
		if i.Status == "unavailable" {
			unavailable++
		}
	}
	imported, duplicates, failed := 0, 0, 0
	for _, r := range j.Results {
		switch r.Status {
		case "imported":
			imported++
		case "duplicate":
			duplicates++
		case "failed":
			failed++
		}
	}
	label := j.Label
	if j.State == "cancelled" {
		label = "Cancelled. Completed posts are kept."
	}
	return map[string]any{"nativeGrant": j.NativeGrant, "nativeUpload": j.Native, "id": j.ID, "state": j.State, "filename": j.Filename, "size": j.Size, "received": j.Received, "chunks": j.Chunks, "fingerprint": j.Fingerprint, "revision": j.Revision, "retained": j.Retained, "chunkSize": instagramChunkSize, "digest": j.Digest, "items": sample, "postCount": len(j.Items), "ready": ready, "unavailable": unavailable, "label": label, "done": j.Done, "total": j.Total, "error": j.Error, "imported": imported, "duplicates": duplicates, "failed": failed, "created": j.Created, "updated": j.Updated, "visibility": j.Visibility}
}

type instagramUploadInput struct {
	Filename    string
	Size        int64
	Fingerprint string
}

func (a *App) createInstagramJob(w http.ResponseWriter, r *http.Request) {
	var in instagramUploadInput
	if !apphost.Decode(w, r, &in) {
		return
	}
	j, status, e := a.allocateInstagramJob(r.Context(), uid(r), in)
	if e != nil {
		apphost.Problem(w, status, e.Error())
		return
	}
	apphost.JSON(w, status, instagramJobView(j))
}
func (a *App) allocateInstagramJob(ctx context.Context, owner string, in instagramUploadInput) (instagramJob, int, error) {
	var empty instagramJob
	bad := func(status int, msg string) (instagramJob, int, error) { return empty, status, errors.New(msg) }
	if in.Size < 22 || in.Size > instagramZIPLimit || len(in.Filename) > 255 || in.Filename == "" || path.Base(in.Filename) != in.Filename || strings.ContainsAny(in.Filename, "\\\x00") || len(in.Fingerprint) != 64 {
		return bad(400, "Choose an Instagram ZIP up to 512 MiB")
	}
	if _, e := hex.DecodeString(in.Fingerprint); e != nil {
		return bad(400, "Invalid upload fingerprint")
	}
	if a.storageCheck(ctx) != nil {
		return bad(503, "Connect protected My Photos/My Social storage before uploading")
	}
	a.instagramMu.Lock()
	defer a.instagramMu.Unlock()
	rows, e := a.H.DB.Query("SELECT id FROM instagram_import_jobs WHERE owner=? AND state='uploading' ORDER BY created DESC", owner)
	if e != nil {
		return bad(503, "Could not read saved uploads")
	}
	ids := []string{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		j, e := a.loadInstagramJob(id, owner)
		if e == nil && j.Size == in.Size && j.Filename == in.Filename && j.Fingerprint == in.Fingerprint {
			return j, 200, nil
		}
	}
	var count, global int
	if a.H.DB.QueryRow("SELECT count(*) FROM instagram_import_jobs WHERE owner=? AND retained=1", owner).Scan(&count) != nil || a.H.DB.QueryRow("SELECT count(*) FROM instagram_import_jobs WHERE retained=1").Scan(&global) != nil {
		return bad(503, "Could not check saved uploads")
	}
	if count >= 2 || global >= 4 {
		return bad(429, "Finish or cancel an existing ZIP before uploading another")
	}
	j := instagramJob{ID: apphost.ID(), Owner: owner, State: "uploading", Filename: in.Filename, Size: in.Size, Fingerprint: in.Fingerprint, Visibility: "private", Label: "Waiting for ZIP upload", Created: a.now().Unix(), Updated: a.now().Unix(), Chunks: []instagramChunk{}, Results: []instagramResult{}, Retained: true}
	_, e = a.H.DB.Exec("INSERT INTO instagram_import_jobs VALUES(?,?,?,?,?,?,?)", j.ID, j.Owner, j.State, encode(j), j.Created, j.Updated, 1)
	if e != nil {
		return bad(503, "Could not save upload")
	}
	return j, 201, nil
}
func (a *App) listInstagramJobs(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT id FROM instagram_import_jobs WHERE owner=? ORDER BY created DESC,rowid DESC LIMIT 10", uid(r))
	if e != nil {
		a.fail(w, e)
		return
	}
	ids := []string{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	out := []map[string]any{}
	for _, id := range ids {
		if j, e := a.loadInstagramJob(id, uid(r)); e == nil {
			out = append(out, instagramJobView(j))
		}
	}
	apphost.JSON(w, 200, map[string]any{"jobs": out})
}
func (a *App) getInstagramJob(w http.ResponseWriter, r *http.Request) {
	j, e := a.loadInstagramJob(r.PathValue("id"), uid(r))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	apphost.JSON(w, 200, instagramJobView(j))
}
func (a *App) uploadInstagramChunk(w http.ResponseWriter, r *http.Request) {
	offset, e := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	expected := r.Header.Get("X-Chunk-SHA256")
	if e != nil || offset < 0 || len(expected) != 64 {
		apphost.Problem(w, 400, "Invalid upload chunk")
		return
	}
	if _, e = hex.DecodeString(expected); e != nil {
		apphost.Problem(w, 400, "Invalid chunk checksum")
		return
	}
	a.instagramMu.Lock()
	defer a.instagramMu.Unlock()
	j, e := a.loadInstagramJob(r.PathValue("id"), uid(r))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	if j.State != "uploading" || offset != j.Received {
		apphost.Problem(w, 409, "Upload offset changed. Reconnect and resume this ZIP")
		return
	}
	if e = a.storageCheck(r.Context()); e != nil {
		apphost.Problem(w, 503, "Protected storage is unavailable; resume when it reconnects")
		return
	}
	size := min(int64(instagramChunkSize), j.Size-j.Received)
	if size <= 0 {
		apphost.Problem(w, 409, "ZIP upload is already complete")
		return
	}
	if e = os.MkdirAll(a.Root, 0700); e != nil {
		a.fail(w, e)
		return
	}
	root, e := os.OpenRoot(a.Root)
	if e != nil {
		a.fail(w, e)
		return
	}
	defer root.Close()
	if e = root.MkdirAll("generated/imports", 0700); e != nil {
		a.fail(w, e)
		return
	}
	f, e := root.OpenFile(instagramArchive(j.ID), os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		a.fail(w, e)
		return
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Size() < j.Received {
		apphost.Problem(w, 409, "Stored upload is incomplete. Cancel it and choose the ZIP again")
		return
	}
	if e = f.Truncate(j.Received); e != nil {
		a.fail(w, e)
		return
	}
	if _, e = f.Seek(j.Received, io.SeekStart); e != nil {
		a.fail(w, e)
		return
	}
	committed := false
	defer func() {
		if !committed {
			_ = f.Truncate(j.Received)
		}
	}()
	r.Body = http.MaxBytesReader(w, r.Body, size+1)
	h := sha256.New()
	n, e := io.Copy(io.MultiWriter(f, h), r.Body)
	if e != nil || n != size || hex.EncodeToString(h.Sum(nil)) != expected {
		apphost.Problem(w, 400, "Upload chunk was interrupted or damaged. Resume the same ZIP")
		return
	}
	if e = f.Sync(); e != nil {
		a.fail(w, e)
		return
	}
	j.Chunks = append(j.Chunks, instagramChunk{Offset: offset, Size: n, SHA: expected})
	previous := j.Received
	j.Received += n
	j.Label = "Uploading ZIP"
	j.Error = ""
	if e = a.saveInstagramJob(&j); e != nil {
		j.Received = previous
		a.fail(w, e)
		return
	}
	committed = true
	apphost.JSON(w, 200, instagramJobView(j))
}
func (a *App) instagramJobAction(w http.ResponseWriter, r *http.Request) {
	a.instagramMu.Lock()
	defer a.instagramMu.Unlock()
	j, e := a.loadInstagramJob(r.PathValue("id"), uid(r))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	switch r.PathValue("action") {
	case "complete":
		if j.State == "uploading" {
			if j.Received != j.Size {
				apphost.Problem(w, 409, "Finish uploading the ZIP before checking it")
				return
			}
			j.State = "checking"
			j.Label = "ZIP saved. Waiting for server checks"
		}
	case "import":
		var in struct{ Visibility string }
		if !apphost.Decode(w, r, &in) {
			return
		}
		if in.Visibility != "private" && in.Visibility != "followers" && in.Visibility != "close-friends" && in.Visibility != "public" {
			apphost.Problem(w, 400, "Choose an audience")
			return
		}
		if j.State == "ready" {
			j.Keys = nil
			for _, item := range j.Items {
				if item.Status == "ready" {
					j.Keys = append(j.Keys, item.Key)
				}
			}
			if len(j.Keys) == 0 {
				apphost.Problem(w, 409, "There are no available posts to import")
				return
			}
			j.Visibility = in.Visibility
			j.Results = nil
			j.State = "queued_import"
			j.Error = ""
			j.Done = 0
			j.Total = len(j.Keys)
			j.Label = "Import queued on the server"
		} else if j.State == "completed" {
			failed := false
			kept := []instagramResult{}
			for _, result := range j.Results {
				if result.Status == "failed" {
					failed = true
				} else {
					kept = append(kept, result)
				}
			}
			if failed {
				j.Results = kept
				j.Visibility = in.Visibility
				j.State = "queued_import"
				j.Error = ""
				j.Label = "Retry queued on the server"
			}
		} else if j.State != "queued_import" && j.State != "importing" {
			apphost.Problem(w, 409, "Wait for the ZIP check before importing")
			return
		}
	case "cancel":
		if instagramActive(j.State) || j.Retained {
			j.State = "cancelled"
			j.Label = "Cancelled. Completed posts are kept."
			if cancel := a.instagramCancel[j.ID]; cancel != nil {
				cancel()
			}
			if j.Received == 0 {
				j.Retained = false
			}
		}
	case "retry":
		if j.State != "failed" || !j.Retained || j.Received != j.Size {
			apphost.Problem(w, 409, "Choose the ZIP again to resume uploading")
			return
		}
		j.Error = ""
		j.Label = "Retry queued on the server"
		if len(j.Keys) > 0 {
			j.State = "queued_import"
		} else {
			j.State = "checking"
		}
	default:
		http.NotFound(w, r)
		return
	}
	if e = a.saveInstagramJob(&j); e != nil {
		a.fail(w, e)
		return
	}
	apphost.JSON(w, 202, instagramJobView(j))
}

func (a *App) instagramWorker(ctx context.Context) {
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		a.runInstagramJob(ctx)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}
func (a *App) instagramOwnerContext(ctx context.Context, owner string, cleanup ...bool) (context.Context, error) {
	var u apphost.User
	e := a.H.DB.QueryRow("SELECT id,issuer,subject,role FROM host_users WHERE id=?", owner).Scan(&u.ID, &u.Issuer, &u.Subject, &u.Role)
	if e != nil || (u.Role == "disabled" && (len(cleanup) == 0 || !cleanup[0])) || u.Subject == "" || (a.H.Config.Issuer != "" && u.Issuer != a.H.Config.Issuer) {
		return ctx, errors.New("This account is no longer active")
	}
	return context.WithValue(ctx, instagramOwnerKey{}, u), nil
}
func (a *App) runInstagramJob(server context.Context) {
	if !a.importMu.TryLock() {
		return
	}
	defer a.importMu.Unlock()
	var id, owner string
	if e := a.H.DB.QueryRow("SELECT id,owner FROM instagram_import_jobs WHERE state IN ('checking','queued_import','importing') ORDER BY created LIMIT 1").Scan(&id, &owner); e != nil {
		a.cleanInstagramJobs(server)
		return
	}
	a.instagramMu.Lock()
	j, e := a.loadInstagramJob(id, owner)
	if e != nil {
		a.instagramMu.Unlock()
		return
	}
	ctx, cancel := context.WithTimeout(server, 12*time.Hour)
	if a.instagramCancel == nil {
		a.instagramCancel = map[string]context.CancelFunc{}
	}
	a.instagramCancel[id] = cancel
	a.instagramMu.Unlock()
	defer func() { cancel(); a.instagramMu.Lock(); delete(a.instagramCancel, id); a.instagramMu.Unlock() }()
	ctx, e = a.instagramOwnerContext(ctx, owner)
	// Request cancellation never reaches here. Server shutdown leaves the stored
	// phase intact so the next process can resume using per-post deduplication.
	fail := func(err error) {
		if server.Err() != nil {
			return
		}
		a.H.Freeze.RLock()
		defer a.H.Freeze.RUnlock()
		_ = a.changeInstagramJob(id, owner, func(current *instagramJob) {
			if current.State != "cancelled" {
				current.State = "failed"
				current.Error = err.Error()
				current.Label = "Import needs attention. Completed posts are kept."
			}
		})
	}
	if e != nil {
		fail(e)
		return
	}
	var root *os.Root
	var file *os.File
	var items []instagramItem
	e = func() error {
		a.H.Freeze.RLock()
		defer a.H.Freeze.RUnlock()
		if e := a.storageCheck(ctx); e != nil {
			return errors.New("Protected storage is unavailable. Reconnect it and retry")
		}
		var err error
		root, err = os.OpenRoot(a.Root)
		if err != nil {
			return errors.New("Stored ZIP is unavailable. Choose the ZIP again")
		}
		if err = a.recoverInstagramMedia(root); err != nil {
			return errors.New("Could not recover interrupted media safely")
		}
		file, err = root.Open(instagramArchive(id))
		if err != nil {
			return errors.New("Stored ZIP is unavailable. Choose the ZIP again")
		}
		stat, err := file.Stat()
		if err != nil || stat.Size() != j.Size {
			return errors.New("Stored ZIP size changed. Choose the ZIP again")
		}
		h := sha256.New()
		if _, err = io.Copy(h, file); err != nil {
			return errors.New("Could not read the stored ZIP")
		}
		digest := hex.EncodeToString(h.Sum(nil))
		if j.Digest != "" && digest != j.Digest {
			return errors.New("Stored ZIP checksum changed. Choose the ZIP again")
		}
		if err = instagramDirectory(file, j.Size); err != nil {
			return err
		}
		archive, err := zip.NewReader(file, j.Size)
		if err != nil {
			return errors.New("This is not a readable ZIP archive")
		}
		last := time.Time{}
		items, err = a.instagramPlan(ctx, archive, owner, func(label string, done, total int) {
			if time.Since(last) < 250*time.Millisecond && done != total {
				return
			}
			last = time.Now()
			_ = a.changeInstagramJob(id, owner, func(current *instagramJob) {
				if current.State == "cancelled" {
					cancel()
					return
				}
				current.Label = label
				current.Done = done
				current.Total = total
			})
		})
		if err != nil {
			return err
		}
		return a.changeInstagramJob(id, owner, func(current *instagramJob) {
			if current.State == "cancelled" {
				cancel()
				return
			}
			current.Digest = digest
			current.Items = items
			if current.State == "checking" {
				current.State = "ready"
				current.Label = "ZIP checked. Ready to import"
				current.Error = ""
				current.Done = len(items)
				current.Total = len(items)
			} else {
				current.State = "importing"
				current.Label = "Importing on the server"
				current.Done = len(current.Results)
				current.Total = len(current.Keys)
			}
		})
	}()
	if root != nil {
		defer root.Close()
	}
	if file != nil {
		defer file.Close()
	}
	if e != nil {
		fail(e)
		return
	}
	j, e = a.loadInstagramJob(id, owner)
	if e != nil || j.State == "ready" || j.State == "cancelled" {
		return
	}
	byKey := map[string]instagramItem{}
	for _, item := range items {
		byKey[item.Key] = item
	}
	finished := map[string]bool{}
	for _, r := range j.Results {
		finished[r.Key] = true
	}
	for index, key := range j.Keys {
		if server.Err() != nil {
			return
		}
		if finished[key] {
			continue
		}
		a.H.Freeze.RLock()
		current, err := a.loadInstagramJob(id, owner)
		if err != nil || current.State == "cancelled" || ctx.Err() != nil {
			a.H.Freeze.RUnlock()
			if server.Err() == nil && ctx.Err() == context.DeadlineExceeded {
				fail(errors.New("Server processing time limit reached. Completed posts are kept"))
			}
			return
		}
		item, ok := byKey[key]
		result := instagramResult{Key: key, Status: "failed", Message: "This post is no longer available in the ZIP"}
		if ok && item.Status != "unavailable" {
			_ = a.changeInstagramJob(id, owner, func(current *instagramJob) {
				current.Label = fmt.Sprintf("Importing post %d of %d", index+1, len(j.Keys))
				current.Done = len(current.Results)
				current.Total = len(j.Keys)
			})
			_, duplicate, err := a.importInstagramPost(ctx, root, owner, j.Visibility, item, func(media, total int) {
				_ = a.changeInstagramJob(id, owner, func(current *instagramJob) {
					if current.State != "cancelled" {
						current.Label = fmt.Sprintf("Importing post %d of %d · media %d of %d", index+1, len(j.Keys), media, total)
					}
				})
			})
			if server.Err() != nil {
				a.H.Freeze.RUnlock()
				return
			}
			if err != nil {
				result.Message = err.Error()
			} else if duplicate {
				result.Status = "duplicate"
				result.Message = "Already imported"
			} else {
				result.Status = "imported"
				result.Message = "Imported"
			}
		}
		err = a.changeInstagramJob(id, owner, func(current *instagramJob) {
			current.Results = append(current.Results, result)
			current.Done = len(current.Results)
			current.Total = len(j.Keys)
		})
		a.H.Freeze.RUnlock()
		if err != nil {
			return
		}
	}
	a.H.Freeze.RLock()
	defer a.H.Freeze.RUnlock()
	_ = a.changeInstagramJob(id, owner, func(current *instagramJob) {
		if current.State != "cancelled" {
			current.State = "completed"
			current.Label = "Import complete"
			current.Done = len(current.Keys)
			current.Total = len(current.Keys)
			for _, r := range current.Results {
				for n := range current.Items {
					if current.Items[n].Key == r.Key {
						current.Items[n].Status = r.Status
						current.Items[n].Reason = r.Message
						break
					}
				}
			}
		}
	})
	// Keep the archive only when a failed post can still be retried.
	current, err := a.loadInstagramJob(id, owner)
	if err == nil {
		failed := false
		for _, r := range current.Results {
			if r.Status == "failed" {
				failed = true
			}
		}
		if !failed {
			file.Close()
			if e := root.Remove(instagramArchive(id)); e == nil || errors.Is(e, os.ErrNotExist) {
				_ = a.changeInstagramJob(id, owner, func(j *instagramJob) { j.Retained = false })
			}
		}
	}
}

// Expired archives are removed only through the protected adapter. Failed and
// completed job records remain briefly for a returning browser to read results.
func (a *App) cleanInstagramJobs(ctx context.Context) {
	a.H.Freeze.RLock()
	defer a.H.Freeze.RUnlock()
	rows, e := a.H.DB.Query("SELECT id,owner,state,updated FROM instagram_import_jobs WHERE state NOT IN ('checking','queued_import','importing') AND (retained=1 OR updated<?)", a.now().Unix()-7*86400)
	if e != nil {
		return
	}
	type expired struct {
		id, owner, state string
		updated          int64
	}
	list := []expired{}
	for rows.Next() {
		var j expired
		rows.Scan(&j.id, &j.owner, &j.state, &j.updated)
		if j.state == "cancelled" || a.now().Unix()-j.updated > 86400 {
			list = append(list, j)
		}
	}
	rows.Close()
	for _, j := range list {
		if ctx.Err() != nil {
			return
		}
		ownerCtx, e := a.instagramOwnerContext(ctx, j.owner, true)
		if e != nil {
			continue
		}
		if a.storageCheck(ownerCtx) != nil {
			continue
		}
		a.instagramMu.Lock()
		// A fresh upload/import action wins over a stale cleanup snapshot.
		current, e := a.loadInstagramJob(j.id, j.owner)
		if e == nil && current.Updated == j.updated && current.State == j.state {
			if current.Retained {
				root, err := os.OpenRoot(a.Root)
				if err != nil && !(errors.Is(err, os.ErrNotExist) && current.Received == 0) {
					a.instagramMu.Unlock()
					continue
				}
				if root != nil {
					err = root.Remove(instagramArchive(j.id))
					root.Close()
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						a.instagramMu.Unlock()
						continue
					}
				}
			}
			current.Retained = false
			if current.State == "uploading" || current.State == "ready" {
				current.State = "failed"
				current.Error = "This stored ZIP expired after 24 hours. Choose the ZIP again"
			}
			_ = a.saveInstagramJob(&current)
			if a.now().Unix()-j.updated > 7*86400 {
				a.H.DB.Exec("DELETE FROM instagram_import_jobs WHERE id=? AND owner=?", j.id, j.owner)
			}
		}
		a.instagramMu.Unlock()
	}
}
