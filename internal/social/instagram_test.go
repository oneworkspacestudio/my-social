package social

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func instagramZIP(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	names := []string{}
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		w, e := z.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write(files[name]); e != nil {
			t.Fatal(e)
		}
	}
	if e := z.Close(); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func (f *fixture) instagramCall(t *testing.T, action, user string, data []byte, options any) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("options", encode(options))
	p, _ := w.CreateFormFile("archive", "instagram.zip")
	p.Write(data)
	w.Close()
	r := httptest.NewRequest("POST", f.a.H.Config.Origin+"/api/imports/instagram/"+action, &body)
	r.Header.Set("Content-Type", w.FormDataContentType())
	r.Header.Set("Origin", f.a.H.Config.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf[user])
	if c := f.cookies[user]; c != nil {
		r.AddCookie(c)
	}
	out := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(out, r)
	return out
}

type instagramPreview struct {
	Digest string          `json:"digest"`
	Items  []instagramItem `json:"items"`
}

func (f *fixture) instagramPreview(t *testing.T, data []byte) instagramPreview {
	t.Helper()
	w := f.instagramCall(t, "preview", "alice", data, map[string]any{})
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var p instagramPreview
	if e := json.Unmarshal(w.Body.Bytes(), &p); e != nil {
		t.Fatal(e)
	}
	return p
}
func countRows(t *testing.T, f *fixture, table string) int {
	t.Helper()
	var n int
	if e := f.a.H.DB.QueryRow("SELECT count(*) FROM " + table).Scan(&n); e != nil {
		t.Fatal(e)
	}
	return n
}
func assertNoImportTemps(t *testing.T, f *fixture) {
	t.Helper()
	entries, e := os.ReadDir(filepath.Join(f.a.Root, "generated", "imports"))
	if e != nil && !os.IsNotExist(e) {
		t.Fatal(e)
	}
	if len(entries) != 0 {
		t.Fatal("temporary ZIP retained")
	}
}
func TestInstagramPreviewSelectionDatesIsolationAndRetry(t *testing.T) {
	f := setup(t)
	files := map[string][]byte{
		"export/your_instagram_activity/media/posts_1.json": []byte(`[{"media":[{"uri":"media/posts/first.png","creation_timestamp":1600000000,"title":"Caf\u00c3\u00a9 with friends"},{"uri":"media/posts/second.png","creation_timestamp":1600000001}]}]`),
		"export/your_instagram_activity/media/posts_2.json": []byte(`[{"title":"Not selected","creation_timestamp":1600000002,"media":[{"uri":"media/posts/first.png"}]},{"title":"Missing","creation_timestamp":1600000003,"media":[{"uri":"media/posts/missing.png"}]}]`),
		"export/your_instagram_activity/media/stories.json": []byte(`{"ig_stories":[{"title":"Old Story","creation_timestamp":1600000004,"media":[{"uri":"media/posts/first.png"}]}]}`),
		"export/messages/inbox/private.json":                []byte(`{"private":"Never import messages"}`),
		"export/media/posts/first.png":                      sample(), "export/media/posts/second.png": sample(),
	}
	data := instagramZIP(t, files)
	preview := f.instagramPreview(t, data)
	if len(preview.Items) != 4 || countRows(t, f, "posts") != 0 || countRows(t, f, "media") != 0 {
		t.Fatal("preview wrote content or missed manifest parts", preview)
	}
	if preview.Items[0].Caption != "Café with friends" || preview.Items[2].Status != "unavailable" {
		t.Fatal(preview.Items)
	}
	assertNoImportTemps(t, f)
	keys := []string{preview.Items[0].Key, preview.Items[3].Key}
	options := instagramOptions{Digest: preview.Digest, Keys: keys, Visibility: "private"}
	w := f.instagramCall(t, "apply", "alice", data, options)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"imported":2`) {
		t.Fatal(w.Code, w.Body.String())
	}
	if countRows(t, f, "posts") != 2 || countRows(t, f, "media") != 3 || countRows(t, f, "events") != 0 {
		t.Fatal("unexpected content counts")
	}
	var id string
	f.a.H.DB.QueryRow("SELECT id FROM posts WHERE request_key=?", "instagram:"+keys[0]).Scan(&id)
	p, e := f.a.post(id)
	if e != nil || p.Caption != "Café with friends" || p.Created != 1600000000 || p.PublishAt != 1600000000 || len(p.Media) != 2 || p.Visibility != "private" {
		t.Fatal(p, e)
	}
	for _, mid := range p.Media {
		if f.call("GET", "/api/media/"+mid, "bob", nil).Code != 404 || f.call("GET", "/api/media/"+mid+"?original=1", "bob", nil).Code != 404 {
			t.Fatal("private import leaked")
		}
	}
	options.Visibility = "public"
	w = f.instagramCall(t, "apply", "alice", data, options)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"duplicates":2`) || countRows(t, f, "media") != 3 {
		t.Fatal("retry duplicated", w.Body.String())
	}
	p, _ = f.a.post(id)
	if p.Visibility != "private" {
		t.Fatal("retry changed audience")
	}
	var state string
	var expires int64
	f.a.H.DB.QueryRow("SELECT state,expires FROM posts WHERE kind='story'").Scan(&state, &expires)
	if state != "archived" || expires != 1600000004+86400 {
		t.Fatal("old Story resurfaced")
	}
	f.call("PATCH", "/api/posts/"+p.ID, "alice", map[string]string{"caption": p.Caption, "visibility": "public"})
	if f.call("GET", "/api/media/"+p.Media[0], "bob", nil).Code != 200 || f.call("GET", "/api/media/"+p.Media[0]+"?original=1", "bob", nil).Code != 404 {
		t.Fatal("rendition/original boundary failed")
	}
	assertNoImportTemps(t, f)
	// Duplicate identity is independent of the ZIP container/wrapper path.
	delete(files, "export/messages/inbox/private.json")
	preview = f.instagramPreview(t, instagramZIP(t, files))
	if preview.Items[0].Status != "duplicate" {
		t.Fatal("repacked export duplicated")
	}
}
func TestInstagramRejectsUnsafeArchivesAndMetadata(t *testing.T) {
	f := setup(t)
	for _, name := range []string{"../outside", "/absolute", "C:/absolute", "media\\bad", "media/../outside"} {
		w := f.instagramCall(t, "preview", "alice", instagramZIP(t, map[string][]byte{name: []byte("unsafe")}), map[string]any{})
		if w.Code != 400 {
			t.Fatal(name, w.Code, w.Body.String())
		}
	}
	for _, raw := range [][]byte{[]byte(`not a ZIP`), instagramZIP(t, map[string][]byte{"content/posts_1.html": []byte("HTML export")}), instagramZIP(t, map[string][]byte{"content/posts_1.json": []byte("broken JSON")})} {
		if w := f.instagramCall(t, "preview", "alice", raw, map[string]any{}); w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	header := &zip.FileHeader{Name: "media/link"}
	header.SetMode(os.ModeSymlink | 0700)
	out, _ := z.CreateHeader(header)
	io.WriteString(out, "../outside")
	z.Close()
	if w := f.instagramCall(t, "preview", "alice", b.Bytes(), map[string]any{}); w.Code != 400 {
		t.Fatal("symlink accepted")
	}
	// Validate size claims before opening/decompressing entries.
	if _, e := f.a.instagramPlan(t.Context(), &zip.Reader{File: []*zip.File{{FileHeader: zip.FileHeader{Name: "huge", UncompressedSize64: 3 << 30}}}}, f.users["alice"]); e == nil {
		t.Fatal("ZIP bomb size claim accepted")
	}
	assertNoImportTemps(t, f)
}
func TestInstagramApplyValidationGuardQuotaAndCSRF(t *testing.T) {
	f := setup(t)
	data := instagramZIP(t, map[string][]byte{"content/posts_1.json": []byte(`[{"creation_timestamp":1600000000,"media":[{"uri":"media/posts/a.png"}]}]`), "media/posts/a.png": sample()})
	p := f.instagramPreview(t, data)
	for _, opts := range []instagramOptions{{Digest: strings.Repeat("0", 64), Keys: []string{p.Items[0].Key}, Visibility: "private"}, {Digest: p.Digest, Keys: []string{"unknown"}, Visibility: "private"}, {Digest: p.Digest, Keys: []string{p.Items[0].Key, p.Items[0].Key}, Visibility: "private"}, {Digest: p.Digest, Keys: []string{p.Items[0].Key}, Visibility: "invalid"}} {
		w := f.instagramCall(t, "apply", "alice", data, opts)
		if w.Code != 400 && w.Code != 409 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if countRows(t, f, "media") != 0 || countRows(t, f, "posts") != 0 {
		t.Fatal("invalid selection mutated content")
	}
	f.csrf["alice"] = ""
	if w := f.instagramCall(t, "preview", "alice", data, map[string]any{}); w.Code != 403 {
		t.Fatal("CSRF bypass", w.Code)
	}
	if w := f.instagramCall(t, "preview", "", data, map[string]any{}); w.Code != 401 {
		t.Fatal("unauthenticated import", w.Code)
	}
	f.csrf["alice"] = "unused" // Guard test happens before decoding request data.
	f.a.Root = ""
	if w := f.instagramCall(t, "preview", "bob", data, map[string]any{}); w.Code != 503 {
		t.Fatal("unguarded import", w.Code)
	}
}
func TestInstagramPartialFailureRollsBackWholeCarousel(t *testing.T) {
	f := setup(t)
	broken := sample()[:40]
	data := instagramZIP(t, map[string][]byte{"content/posts_1.json": []byte(`[{"title":"Good","creation_timestamp":1600000000,"media":[{"uri":"media/posts/a.png"}]},{"title":"Bad carousel","creation_timestamp":1600000001,"media":[{"uri":"media/posts/a.png"},{"uri":"media/posts/broken.png"}]}]`), "media/posts/a.png": sample(), "media/posts/broken.png": broken})
	p := f.instagramPreview(t, data)
	if p.Items[1].Status != "ready" {
		t.Fatal("fixture must pass header check before full decode", p.Items[1])
	}
	w := f.instagramCall(t, "apply", "alice", data, instagramOptions{Digest: p.Digest, Keys: []string{p.Items[0].Key, p.Items[1].Key}, Visibility: "private"})
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"imported":1`) || !strings.Contains(w.Body.String(), `"failed":1`) {
		t.Fatal(w.Code, w.Body.String())
	}
	if countRows(t, f, "posts") != 1 || countRows(t, f, "media") != 1 || countRows(t, f, "upload_intents") != 1 {
		t.Fatal("failed carousel left partial database content")
	}
	fileCount := 0
	filepath.WalkDir(f.a.Root, func(_ string, d os.DirEntry, e error) error {
		if e == nil && !d.IsDir() {
			fileCount++
		}
		return e
	})
	if fileCount != 2 {
		t.Fatal("failed carousel left files", fileCount)
	}
	assertNoImportTemps(t, f)
}
func TestInstagramVideoAndMixedCarousel(t *testing.T) {
	ff, fp := os.Getenv("MYSOCIAL_TEST_FFMPEG"), os.Getenv("MYSOCIAL_TEST_FFPROBE")
	if ff == "" || fp == "" {
		t.Skip("configure operator FFmpeg/FFprobe for native video import checks")
	}
	f := setup(t)
	f.a.FFmpeg = ff
	f.a.FFprobe = fp
	video := filepath.Join(t.TempDir(), "clip.mp4")
	if output, e := exec.Command(ff, "-nostdin", "-v", "error", "-f", "lavfi", "-i", "color=c=blue:s=320x240:d=1", "-c:v", "libx264", "-pix_fmt", "yuv420p", video).CombinedOutput(); e != nil {
		t.Fatal(e, string(output))
	}
	b, e := os.ReadFile(video)
	if e != nil {
		t.Fatal(e)
	}
	data := instagramZIP(t, map[string][]byte{"content/posts_1.json": []byte(`[{"title":"Mixed carousel","creation_timestamp":1600000000,"media":[{"uri":"media/posts/a.png"},{"uri":"media/reels/a.mp4"}]}]`), "content/reels.json": []byte(`{"ig_reels_media":[{"title":"Reel","creation_timestamp":1600000001,"media":[{"uri":"media/reels/a.mp4"}]}]}`), "media/posts/a.png": sample(), "media/reels/a.mp4": b})
	p := f.instagramPreview(t, data)
	if len(p.Items) != 2 || p.Items[0].Status != "ready" || p.Items[1].Status != "ready" {
		t.Fatal(p)
	}
	w := f.instagramCall(t, "apply", "alice", data, instagramOptions{Digest: p.Digest, Keys: []string{p.Items[0].Key, p.Items[1].Key}, Visibility: "private"})
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"imported":2`) {
		t.Fatal(w.Code, w.Body.String())
	}
	var id string
	f.a.H.DB.QueryRow("SELECT id FROM posts WHERE kind='post'").Scan(&id)
	post, _ := f.a.post(id)
	var extras struct {
		Instagram struct {
			MediaTypes []string `json:"mediaTypes"`
		} `json:"instagram"`
	}
	json.Unmarshal(post.Extras, &extras)
	if len(extras.Instagram.MediaTypes) != 2 || extras.Instagram.MediaTypes[1] != "video/mp4" {
		t.Fatal(string(post.Extras))
	}
	m, e := f.a.media(post.Media[1])
	if e != nil || !f.a.readyMedia(m) {
		t.Fatal(m, e)
	}
	if w := f.call("GET", "/api/media/"+m.ID, "alice", nil); w.Code != 200 {
		t.Fatal(w.Code)
	}
	longVideo := filepath.Join(t.TempDir(), "too-long.mp4")
	if output, e := exec.Command(ff, "-nostdin", "-v", "error", "-f", "lavfi", "-i", "color=c=blue:s=16x16:r=1:d=91", "-c:v", "libx264", "-pix_fmt", "yuv420p", longVideo).CombinedOutput(); e != nil {
		t.Fatal(e, string(output))
	}
	longBytes, e := os.ReadFile(longVideo)
	if e != nil {
		t.Fatal(e)
	}
	tooLong := instagramZIP(t, map[string][]byte{"content/reels.json": []byte(`{"ig_reels_media":[{"title":"Too long","creation_timestamp":1600000002,"media":[{"uri":"media/reels/long.mp4"}]}]}`), "media/reels/long.mp4": longBytes})
	longPlan := f.instagramPreview(t, tooLong)
	w = f.instagramCall(t, "apply", "alice", tooLong, instagramOptions{Digest: longPlan.Digest, Keys: []string{longPlan.Items[0].Key}, Visibility: "private"})
	if w.Code != 200 || !strings.Contains(w.Body.String(), "at most 90 seconds") || countRows(t, f, "posts") != 2 || countRows(t, f, "media") != 3 {
		t.Fatal("long video was imported or left partial data", w.Body.String())
	}
	f.a.FFmpeg = ""
	noVideo := f.instagramPreview(t, data)
	if noVideo.Items[0].Status != "unavailable" || noVideo.Items[1].Status != "unavailable" {
		t.Fatal("video accepted without helpers")
	}
}

func TestInstagramRecoversInterruptedFilesAndHonorsQuota(t *testing.T) {
	f := setup(t)
	if e := os.MkdirAll(f.a.Root, 0700); e != nil {
		t.Fatal(e)
	}
	root, e := os.OpenRoot(f.a.Root)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	id, owner := strings.Repeat("a", 48), f.users["alice"]
	original := "users/" + owner + "/posts/" + id + ".original"
	for _, p := range []string{original, "generated/previews/" + owner + "/" + id + ".jpg", "generated/imports/stale.zip"} {
		if e := writeRoot(root, p, []byte("synthetic interrupted bytes")); e != nil {
			t.Fatal(e)
		}
	}
	if e := os.Chtimes(filepath.Join(f.a.Root, "generated", "imports", "stale.zip"), time.Now().Add(-2*time.Hour), time.Now().Add(-2*time.Hour)); e != nil {
		t.Fatal(e)
	}
	if _, e = f.a.H.DB.Exec("INSERT INTO upload_intents VALUES(?,?,?,?,?)", id, owner, "instagram-writing", original, 1600000000); e != nil {
		t.Fatal(e)
	}
	data := instagramZIP(t, map[string][]byte{"content/posts_1.json": []byte(`[{"creation_timestamp":1600000000,"media":[{"uri":"media/posts/a.png"}]}]`), "media/posts/a.png": sample()})
	p := f.instagramPreview(t, data)
	if countRows(t, f, "upload_intents") != 0 {
		t.Fatal("interrupted import not recovered")
	}
	if _, e = root.Stat(original); !os.IsNotExist(e) {
		t.Fatal("interrupted original retained", e)
	}
	assertNoImportTemps(t, f)
	m := f.upload(t, "alice")
	if _, e = f.a.H.DB.Exec("UPDATE media SET bytes=? WHERE id=?", 2<<30, m.ID); e != nil {
		t.Fatal(e)
	}
	w := f.instagramCall(t, "apply", "alice", data, instagramOptions{Digest: p.Digest, Keys: []string{p.Items[0].Key}, Visibility: "private"})
	if w.Code != 200 || !strings.Contains(w.Body.String(), "quota is full") || countRows(t, f, "posts") != 0 || countRows(t, f, "media") != 1 {
		t.Fatal("quota ignored", w.Body.String())
	}
}

func instagramPackets(t *testing.T, w *httptest.ResponseRecorder) []map[string]json.RawMessage {
	t.Helper()
	if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Type"), "application/x-ndjson") || !w.Flushed {
		t.Fatal("progress was not flushed", w.Code, w.Header())
	}
	out := []map[string]json.RawMessage{}
	decoder := json.NewDecoder(w.Body)
	for {
		var p map[string]json.RawMessage
		e := decoder.Decode(&p)
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		out = append(out, p)
	}
	return out
}
func TestInstagramStreamingProgressAndResultPreserveImportContract(t *testing.T) {
	f := setup(t)
	data := instagramZIP(t, map[string][]byte{"content/posts_1.json": []byte(`[{"title":"Progress one","creation_timestamp":1600000000,"media":[{"uri":"media/posts/a.png"}]},{"title":"Progress two","creation_timestamp":1600000001,"media":[{"uri":"media/posts/a.png"}]}]`), "media/posts/a.png": sample()})
	packets := instagramPackets(t, f.instagramCall(t, "preview", "alice", data, map[string]any{"stream": true}))
	var preview instagramPreview
	checks := 0
	for _, p := range packets {
		if string(p["type"]) == `"progress"` {
			checks++
			if string(p["phase"]) != `"checking"` {
				t.Fatal(p)
			}
		}
		if string(p["type"]) == `"result"` {
			if e := json.Unmarshal(p["result"], &preview); e != nil {
				t.Fatal(e)
			}
		}
	}
	if checks < 3 || len(preview.Items) != 2 {
		t.Fatal("missing live checks/final preview", checks, preview)
	}
	packets = instagramPackets(t, f.instagramCall(t, "apply", "alice", data, map[string]any{"stream": true, "digest": preview.Digest, "keys": []string{preview.Items[0].Key, preview.Items[1].Key}, "visibility": "private"}))
	completed := 0
	result := false
	for _, p := range packets {
		if string(p["phase"]) == `"importing"` {
			var done int
			json.Unmarshal(p["done"], &done)
			if done < completed {
				t.Fatal("progress moved backward")
			}
			completed = done
		}
		if string(p["type"]) == `"result"` {
			var v struct {
				Imported int `json:"imported"`
			}
			json.Unmarshal(p["result"], &v)
			result = v.Imported == 2
		}
	}
	if completed != 2 || !result {
		t.Fatal("missing actual post completion/result")
	}
	assertNoImportTemps(t, f)
	// Streamed validation failures remain explicit, rather than false successes.
	bad := instagramZIP(t, map[string][]byte{"content/posts_1.json": []byte("broken")})
	packets = instagramPackets(t, f.instagramCall(t, "preview", "alice", bad, map[string]any{"stream": true}))
	last := packets[len(packets)-1]
	if string(last["type"]) != `"error"` || string(last["status"]) != "400" {
		t.Fatal("stream validation failure hidden", last)
	}
	assertNoImportTemps(t, f)
}
