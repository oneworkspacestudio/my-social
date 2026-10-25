package social

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func restartInstagramApp(t *testing.T, f *fixture) {
	t.Helper()
	config := f.a.H.Config
	if e := f.a.H.Close(); e != nil {
		t.Fatal(e)
	}
	h, e := apphost.New(config, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { h.Close() })
	next, e := New(h)
	if e != nil {
		t.Fatal(e)
	}
	f.a = next
}

func jobUpload(t *testing.T, f *fixture, data []byte) instagramJob {
	t.Helper()
	return jobUploadReader(t, f, bytes.NewReader(data), int64(len(data)))
}
func jobUploadReader(t *testing.T, f *fixture, reader io.ReaderAt, size int64) instagramJob {
	t.Helper()
	w := f.call("POST", "/api/imports/instagram/jobs", "alice", map[string]any{"filename": "instagram-private-check.zip", "size": size, "fingerprint": strings.Repeat("a", 64)})
	if w.Code != 201 {
		t.Fatalf("create job status %d", w.Code)
	}
	var view struct {
		ID string `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &view)
	for offset := int64(0); offset < size; offset += instagramChunkSize {
		data := make([]byte, min(int64(instagramChunkSize), size-offset))
		if _, e := reader.ReadAt(data, offset); e != nil && e != io.EOF {
			t.Fatal("read private chunk")
		}
		w = jobChunk(f, view.ID, "alice", offset, data, "")
		if w.Code != 200 {
			t.Fatalf("chunk status %d", w.Code)
		}
	}
	w = f.call("POST", "/api/imports/instagram/jobs/"+view.ID+"/complete", "alice", map[string]any{})
	if w.Code != 202 {
		t.Fatalf("queue status %d", w.Code)
	}
	j, e := f.a.loadInstagramJob(view.ID, f.users["alice"])
	if e != nil {
		t.Fatal(e)
	}
	return j
}
func jobChunk(f *fixture, id, user string, offset int64, data []byte, checksum string) *httptest.ResponseRecorder {
	sum := sha256.Sum256(data)
	if checksum == "" {
		checksum = hex.EncodeToString(sum[:])
	}
	r := httptest.NewRequest("PUT", f.a.H.Config.Origin+"/api/imports/instagram/jobs/"+id+"/chunks?offset="+strconv.FormatInt(offset, 10), bytes.NewReader(data))
	r.Header.Set("Origin", f.a.H.Config.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf[user])
	r.Header.Set("X-Chunk-SHA256", checksum)
	r.Header.Set("Content-Type", "application/octet-stream")
	if c := f.cookies[user]; c != nil {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(w, r)
	return w
}
func manyInstagramPosts(t *testing.T, count int) []byte {
	rows := []map[string]any{}
	for i := 0; i < count; i++ {
		rows = append(rows, map[string]any{"title": fmt.Sprintf("Synthetic server import %d", i), "creation_timestamp": 1600000000 + i, "media": []map[string]string{{"uri": "media/posts/a.png"}}})
	}
	raw, _ := json.Marshal(rows)
	return instagramZIP(t, map[string][]byte{"content/posts_1.json": raw, "media/posts/a.png": sample()})
}

func TestInstagramJobRequestDisconnectAndServerResume(t *testing.T) {
	f := setup(t)
	data := manyInstagramPosts(t, 101)
	j := jobUpload(t, f, data)
	// The upload response is already over. No request, cookie or page is retained
	// by the worker; a fresh app instance also recovers its durable queue.
	restartInstagramApp(t, f)
	f.a.runInstagramJob(t.Context())
	j, e := f.a.loadInstagramJob(j.ID, j.Owner)
	if e != nil || j.State != "ready" || len(j.Items) != 101 {
		t.Fatalf("preview state %s", j.State)
	}
	view := instagramJobView(j)
	if len(view["items"].([]instagramItem)) != 5 || view["ready"] != 101 {
		t.Fatal("sample lost full ready count")
	}
	var count int
	f.a.H.DB.QueryRow("SELECT count(*) FROM posts").Scan(&count)
	if count != 0 {
		t.Fatal("preview wrote posts")
	}
	w := f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/import", "alice", map[string]string{"visibility": "private"})
	if w.Code != 202 {
		t.Fatal(w.Code)
	}
	server, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); f.a.runInstagramJob(server) }()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		current, _ := f.a.loadInstagramJob(j.ID, j.Owner)
		if len(current.Results) >= 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	<-done
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.State != "importing" || len(j.Results) == 0 || len(j.Results) >= 101 {
		t.Fatalf("shutdown did not pause durable work: %s/%d", j.State, len(j.Results))
	}
	restartInstagramApp(t, f)
	f.a.runInstagramJob(t.Context())
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.State != "completed" || len(j.Results) != 101 || j.Retained {
		t.Fatalf("resume state %s results %d retained %v", j.State, len(j.Results), j.Retained)
	}
	f.a.H.DB.QueryRow("SELECT count(*) FROM posts WHERE visibility='private'").Scan(&count)
	if count != 101 {
		t.Fatalf("exactly-once private posts: %d", count)
	}
	for _, r := range j.Results {
		if r.Status == "failed" {
			t.Fatal("synthetic post failed")
		}
	}
	if _, e = os.Stat(f.a.Root + "/" + instagramArchive(j.ID)); !os.IsNotExist(e) {
		t.Fatal("successful job retained its archive")
	}
}

func TestInstagramChunkIntegrityOwnershipAndInterruptedResume(t *testing.T) {
	f := setup(t)
	padding := make([]byte, instagramChunkSize+13)
	if _, e := rand.Read(padding); e != nil {
		t.Fatal(e)
	}
	data := instagramZIP(t, map[string][]byte{"content/posts_1.json": []byte(`[{"creation_timestamp":1600000000,"media":[{"uri":"media/posts/a.png"}]}]`), "media/posts/a.png": sample(), "padding.bin": padding})
	w := f.call("POST", "/api/imports/instagram/jobs", "alice", map[string]any{"filename": "resume.zip", "size": len(data), "fingerprint": strings.Repeat("b", 64)})
	var view struct {
		ID string `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &view)
	if w.Code != 201 {
		t.Fatal(w.Code)
	}
	if jobChunk(f, view.ID, "alice", 0, data[:instagramChunkSize], strings.Repeat("0", 64)).Code != 400 {
		t.Fatal("bad digest accepted")
	}
	j, _ := f.a.loadInstagramJob(view.ID, f.users["alice"])
	if j.Received != 0 {
		t.Fatal("bad chunk advanced durable offset")
	}
	if jobChunk(f, view.ID, "bob", 0, data[:instagramChunkSize], "").Code != 404 {
		t.Fatal("cross-account upload accepted")
	}
	if jobChunk(f, view.ID, "alice", 0, data[:instagramChunkSize], "").Code != 200 {
		t.Fatal("first chunk")
	}
	if jobChunk(f, view.ID, "alice", 0, data[:instagramChunkSize], "").Code != 409 {
		t.Fatal("replayed offset overwrote stored bytes")
	}
	if f.call("GET", "/api/imports/instagram/jobs/"+view.ID, "bob", nil).Code != 404 {
		t.Fatal("cross-account job readable")
	}
	w = f.call("GET", "/api/imports/instagram/jobs", "bob", nil)
	if strings.Contains(w.Body.String(), view.ID) {
		t.Fatal("job leaked through listing")
	}
	restartInstagramApp(t, f)
	j, _ = f.a.loadInstagramJob(view.ID, f.users["alice"])
	if j.Received != instagramChunkSize || len(j.Chunks) != 1 {
		t.Fatal("partial upload lost on restart")
	}
	// An interrupted suffix is not committed and is truncated on retry.
	suffix := data[instagramChunkSize:]
	if jobChunk(f, j.ID, "alice", j.Received, suffix[:len(suffix)-1], "").Code != 400 {
		t.Fatal("short final chunk accepted")
	}
	if jobChunk(f, j.ID, "alice", j.Received, suffix, "").Code != 200 {
		t.Fatal("resumed final chunk")
	}
	if f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/complete", "alice", map[string]any{}).Code != 202 {
		t.Fatal("queue resumed ZIP")
	}
	f.a.runInstagramJob(t.Context())
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.State != "ready" {
		t.Fatalf("resumed ZIP: %s", j.State)
	}
	f.csrf["alice"] = "wrong"
	if f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/import", "alice", map[string]string{"visibility": "private"}).Code != 403 {
		t.Fatal("CSRF accepted")
	}
}

func TestInstagramExplicitCancelKeepsPostsAndCannotCrossAccounts(t *testing.T) {
	f := setup(t)
	j := jobUpload(t, f, manyInstagramPosts(t, 100))
	f.a.runInstagramJob(t.Context())
	f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/import", "alice", map[string]string{"visibility": "private"})
	if f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/cancel", "bob", map[string]any{}).Code != 404 {
		t.Fatal("another account cancelled job")
	}
	done := make(chan struct{})
	go func() { defer close(done); f.a.runInstagramJob(t.Context()) }()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		current, _ := f.a.loadInstagramJob(j.ID, j.Owner)
		if len(current.Results) >= 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/cancel", "alice", map[string]any{}).Code != 202 {
		t.Fatal("cancel failed")
	}
	<-done
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.State != "cancelled" {
		t.Fatal("worker overwrote explicit cancellation")
	}
	var before int
	f.a.H.DB.QueryRow("SELECT count(*) FROM posts").Scan(&before)
	if before < 1 || before >= 100 {
		t.Fatal("cancel boundary")
	}
	f.a.runInstagramJob(t.Context())
	var after int
	f.a.H.DB.QueryRow("SELECT count(*) FROM posts").Scan(&after)
	if before != after {
		t.Fatal("cancelled job continued writing")
	}
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.Retained {
		t.Fatal("cancelled archive not cleaned")
	}
}

func TestInstagramExpiredUploadsReleaseOnlyTheirOwnArchive(t *testing.T) {
	f := setup(t)
	old := jobUpload(t, f, manyInstagramPosts(t, 1))
	f.a.runInstagramJob(t.Context())
	fresh := jobUpload(t, f, manyInstagramPosts(t, 2))
	f.a.runInstagramJob(t.Context())
	_, e := f.a.H.DB.Exec("UPDATE instagram_import_jobs SET updated=? WHERE id=?", time.Now().Add(-25*time.Hour).Unix(), old.ID)
	if e != nil {
		t.Fatal(e)
	}
	// Keep JSON timestamps in step with the actual durable metadata.
	f.a.instagramMu.Lock()
	j, _ := f.a.loadInstagramJob(old.ID, old.Owner)
	j.Updated = time.Now().Add(-25 * time.Hour).Unix()
	f.a.H.DB.Exec("UPDATE instagram_import_jobs SET data=?,updated=? WHERE id=?", encode(j), j.Updated, j.ID)
	f.a.instagramMu.Unlock()
	f.a.cleanInstagramJobs(t.Context())
	j, _ = f.a.loadInstagramJob(old.ID, old.Owner)
	if j.State != "failed" || j.Retained {
		t.Fatal("expired ZIP not released")
	}
	j, _ = f.a.loadInstagramJob(fresh.ID, fresh.Owner)
	if j.State != "ready" || !j.Retained {
		t.Fatal("fresh ZIP damaged by cleanup")
	}
}

func TestInstagramEmptyExpiredUploadReleasesReservation(t *testing.T) {
	f := setup(t)
	w := f.call("POST", "/api/imports/instagram/jobs", "alice", map[string]any{"filename": "never-started.zip", "size": 100, "fingerprint": strings.Repeat("a", 64)})
	if w.Code != 201 {
		t.Fatal(w.Code)
	}
	var view struct {
		ID string `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &view)
	j, _ := f.a.loadInstagramJob(view.ID, f.users["alice"])
	j.Updated = time.Now().Add(-25 * time.Hour).Unix()
	f.a.H.DB.Exec("UPDATE instagram_import_jobs SET data=?,updated=? WHERE id=?", encode(j), j.Updated, j.ID)
	f.a.cleanInstagramJobs(t.Context())
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.Retained || j.State != "failed" {
		t.Fatal("empty upload reservation never expired")
	}
}

func TestInstagramBackgroundStorageUsesPersistedOwnerIdentity(t *testing.T) {
	f := setup(t)
	expected := "alice"
	guard := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Social-Subject") != expected || r.Header.Get("X-Social-Issuer") != "local:social" || r.Header.Get("Authorization") != "Bearer synthetic-guard" {
			http.Error(w, "denied", 403)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"contract":"quantum-social-v2"}`)
	}))
	defer guard.Close()
	f.a.H.Config.WorkspaceURL = "http://workspace.example.test"
	f.a.Guard = guard.URL
	f.a.GuardToken = "synthetic-guard"
	j := jobUpload(t, f, manyInstagramPosts(t, 1))
	f.a.runInstagramJob(t.Context())
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.State != "ready" {
		t.Fatal("worker lost immutable identity for storage guard")
	}
	f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/import", "alice", map[string]string{"visibility": "private"})
	expected = "another-subject"
	f.a.runInstagramJob(t.Context())
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.State != "failed" {
		t.Fatal("guard failure did not fail closed")
	}
	expected = "alice"
	if f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/retry", "alice", map[string]any{}).Code != 202 {
		t.Fatal("retry stored ZIP")
	}
	f.a.runInstagramJob(t.Context())
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.State != "completed" {
		t.Fatal("retry did not reuse stored ZIP")
	}
}

// Explicit user-authorized check only. Private contents stay in the isolated
// NTFS test root and are never printed, copied to source, or publicly shared.
func TestInstagramPrivateDownloadArchiveOnServer(t *testing.T) {
	filename := os.Getenv("MYSOCIAL_PRIVATE_TEST_ZIP")
	if filename == "" {
		t.Skip("optional user-authorized private archive check")
	}
	file, e := os.Open(filename)
	if e != nil {
		t.Fatal("private ZIP unavailable")
	}
	defer file.Close()
	info, e := file.Stat()
	if e != nil {
		t.Fatal("private ZIP stat")
	}
	f := setup(t)
	j := jobUploadReader(t, f, file, info.Size())
	f.a.runInstagramJob(t.Context())
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	if j.State != "ready" {
		t.Fatalf("private ZIP check failed: %s", j.Error)
	}
	if f.call("POST", "/api/imports/instagram/jobs/"+j.ID+"/import", "alice", map[string]string{"visibility": "private"}).Code != 202 {
		t.Fatal("private import queue failed")
	}
	f.a.runInstagramJob(t.Context())
	j, _ = f.a.loadInstagramJob(j.ID, j.Owner)
	view := instagramJobView(j)
	t.Logf("Private ZIP: bytes=%d posts=%d imported=%v duplicates=%v failed=%v state=%s", info.Size(), len(j.Items), view["imported"], view["duplicates"], view["failed"], j.State)
	if j.State != "completed" || view["failed"].(int) > 0 {
		reasons := map[string]int{}
		for _, r := range j.Results {
			if r.Status == "failed" {
				reasons[r.Message]++
			}
		}
		t.Fatalf("private import failures (media contents omitted): %v", reasons)
	}
}
