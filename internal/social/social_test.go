package social

import (
	"bytes"
	"encoding/json"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fixture struct {
	a       *App
	cookies map[string]*http.Cookie
	csrf    map[string]string
	users   map[string]string
}

func setup(t *testing.T) *fixture {
	t.Helper()
	root := t.TempDir()
	h, e := apphost.New(apphost.Config{ID: "social", Name: "My Social", Version: "test", Data: filepath.Join(root, "Private"), Listen: "127.0.0.1:29113", Origin: "http://localhost:29113", MediaCapture: true}, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { h.Close() })
	t.Setenv("MYSOCIAL_PHOTOS_ROOT", filepath.Join(root, "My Photos", "My Social"))
	t.Setenv("MYSOCIAL_STANDALONE_STORAGE", "isolated")
	a, e := New(h)
	if e != nil {
		t.Fatal(e)
	}
	f := &fixture{a: a, cookies: map[string]*http.Cookie{}, csrf: map[string]string{}, users: map[string]string{}}
	for _, name := range []string{"alice", "bob"} {
		id := apphost.Identity("local:social", name)
		role := "member"
		if name == "alice" {
			role = "owner"
		}
		_, e = h.DB.Exec("INSERT INTO host_users(id,issuer,subject,name,email,role) VALUES(?,?,?,?,?,?)", id, "local:social", name, name, name+"@example.test", role)
		if e != nil {
			t.Fatal(e)
		}
		token := strings.Repeat(name[:1], 96)
		csrf := apphost.ID()
		_, e = h.DB.Exec("INSERT INTO host_sessions VALUES(?,?,?,?)", apphost.Hash(token), id, csrf, time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano))
		if e != nil {
			t.Fatal(e)
		}
		f.cookies[name] = &http.Cookie{Name: "social_session", Value: token}
		f.csrf[name] = csrf
		f.users[name] = id
		a.ensureProfile(apphost.User{ID: id})
	}
	return f
}
func (f *fixture) call(method, path, user string, body any) *httptest.ResponseRecorder {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	r := httptest.NewRequest(method, f.a.H.Config.Origin+path, rd)
	r.RemoteAddr = "127.0.0.1:19000"
	r.Header.Set("Origin", f.a.H.Config.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf[user])
	if c := f.cookies[user]; c != nil {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(w, r)
	return w
}
func sample() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 32, 48))
	for y := 0; y < 48; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 7), uint8(y * 4), 160, 255})
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	return b.Bytes()
}
func (f *fixture) upload(t *testing.T, user string) Media {
	t.Helper()
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	p, _ := mw.CreateFormFile("original", "photo.png")
	p.Write(sample())
	mw.WriteField("kind", "posts")
	mw.WriteField("project", `{"scene":{"objects":[]}}`)
	mw.Close()
	r := httptest.NewRequest("POST", f.a.H.Config.Origin+"/api/media", &b)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.Header.Set("Origin", f.a.H.Config.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf[user])
	r.AddCookie(f.cookies[user])
	w := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(w, r)
	if w.Code != 201 {
		t.Fatal("upload", w.Code, w.Body.String())
	}
	var m Media
	if json.Unmarshal(w.Body.Bytes(), &m) != nil {
		t.Fatal(w.Body.String())
	}
	return m
}
func (f *fixture) publish(t *testing.T, user, visibility, kind string, ids []string, key string, at int64) Post {
	t.Helper()
	w := f.call("POST", "/api/posts", user, map[string]any{"kind": kind, "caption": "A synthetic moment #test", "alt": "A test pattern", "visibility": visibility, "media": ids, "requestKey": key, "publishAt": at})
	if w.Code != 201 && w.Code != 200 {
		t.Fatal("publish", w.Code, w.Body.String())
	}
	var p Post
	json.Unmarshal(w.Body.Bytes(), &p)
	return p
}
func TestOriginalsAudienceAndDerivativeIsolation(t *testing.T) {
	f := setup(t)
	m := f.upload(t, "alice")
	if f.call("GET", "/api/media/"+m.ID, "bob", nil).Code != 404 {
		t.Fatal("unpublished leaked")
	}
	p := f.publish(t, "alice", "private", "post", []string{m.ID}, apphost.ID(), 0)
	for _, route := range []string{"/api/posts/" + p.ID, "/api/media/" + m.ID, "/api/media/" + m.ID + "?original=1", "/api/posts/" + p.ID + "/comments"} {
		if w := f.call("GET", route, "bob", nil); w.Code != 404 {
			t.Fatal("private route exposed", route, w.Code)
		}
	}
	f.call("PATCH", "/api/posts/"+p.ID, "alice", map[string]string{"caption": "Public pattern", "alt": "pattern", "visibility": "public"})
	if f.call("GET", "/api/media/"+m.ID, "bob", nil).Code != 200 {
		t.Fatal("public derivative unavailable")
	}
	if f.call("GET", "/api/media/"+m.ID+"?original=1", "bob", nil).Code != 404 {
		t.Fatal("public original leaked")
	}
	f.call("POST", "/api/profiles/"+f.users["bob"]+"/relations", "alice", map[string]any{"kind": "block", "enabled": true})
	if f.call("GET", "/api/media/"+m.ID, "bob", nil).Code != 404 {
		t.Fatal("blocked derivative leaked")
	}
	stored, e := f.a.media(m.ID)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.HasPrefix(stored.Original, "users/"+f.users["alice"]+"/posts/") {
		t.Fatal("wrong originals layout")
	}
	if _, e = os.Stat(filepath.Join(f.a.Root, filepath.FromSlash(stored.Original))); e != nil {
		t.Fatal(e)
	}
	if strings.Contains(f.call("GET", "/api/media", "alice", nil).Body.String(), "original") {
		t.Fatal("filesystem path exposed")
	}
}
func TestRecoverablePublishingIdempotencyAndStorageFailure(t *testing.T) {
	f := setup(t)
	m := f.upload(t, "alice")
	key := apphost.ID()
	p := f.publish(t, "alice", "public", "post", []string{m.ID}, key, 0)
	again := f.publish(t, "alice", "public", "post", []string{m.ID}, key, 0)
	if p.ID != again.ID {
		t.Fatal("retry created phantom duplicate")
	}
	var count int
	f.a.H.DB.QueryRow("SELECT count(*) FROM posts").Scan(&count)
	if count != 1 {
		t.Fatal(count)
	}
	w := f.call("POST", "/api/posts", "alice", map[string]any{"kind": "post", "caption": "missing", "visibility": "public", "media": []string{apphost.ID()}, "requestKey": apphost.ID()})
	if w.Code != 409 {
		t.Fatal("published missing media", w.Code)
	}
	f.a.Root = ""
	w = f.call("POST", "/api/posts", "alice", map[string]any{"kind": "post", "caption": "unavailable", "visibility": "public", "media": []string{m.ID}, "requestKey": apphost.ID()})
	if w.Code != 503 {
		t.Fatal("storage failure was not safe", w.Code)
	}
	f.a.H.DB.QueryRow("SELECT count(*) FROM posts").Scan(&count)
	if count != 1 {
		t.Fatal("metadata published without storage")
	}
}
func TestStoryClockAndScheduling(t *testing.T) {
	f := setup(t)
	m := f.upload(t, "alice")
	now := time.Now()
	f.a.now = func() time.Time { return now }
	story := f.publish(t, "alice", "public", "story", []string{m.ID}, apphost.ID(), 0)
	if story.Expires != now.Unix()+86400 {
		t.Fatal(story.Expires)
	}
	if f.call("GET", "/api/posts/"+story.ID, "bob", nil).Code != 200 {
		t.Fatal("story not visible")
	}
	f.a.now = func() time.Time { return now.Add(24*time.Hour + time.Second) }
	if f.call("GET", "/api/posts/"+story.ID, "bob", nil).Code != 404 {
		t.Fatal("expired story leaked before cleanup")
	}
	f.a.now = func() time.Time { return now }
	scheduled := f.publish(t, "alice", "public", "post", []string{m.ID}, apphost.ID(), now.Add(time.Hour).Unix())
	if f.call("GET", "/api/posts/"+scheduled.ID, "bob", nil).Code != 404 {
		t.Fatal("schedule leaked")
	}
	f.a.now = func() time.Time { return now.Add(2 * time.Hour) }
	f.a.tick(t.Context())
	if f.call("GET", "/api/posts/"+scheduled.ID, "bob", nil).Code != 200 {
		t.Fatal("scheduled post did not publish")
	}
}
func TestPrivateFollowApprovalAndCloseFriends(t *testing.T) {
	f := setup(t)
	f.call("PATCH", "/api/profile", "alice", map[string]any{"handle": "alice", "bio": "", "private": true, "links": []string{}})
	m := f.upload(t, "alice")
	p := f.publish(t, "alice", "followers", "post", []string{m.ID}, apphost.ID(), 0)
	f.call("POST", "/api/profiles/"+f.users["alice"]+"/relations", "bob", map[string]any{"kind": "follow", "enabled": true})
	if f.call("GET", "/api/posts/"+p.ID, "bob", nil).Code != 404 {
		t.Fatal("unapproved follower admitted")
	}
	f.call("POST", "/api/profiles/"+f.users["bob"]+"/relations", "alice", map[string]any{"kind": "approve", "enabled": true})
	if f.call("GET", "/api/posts/"+p.ID, "bob", nil).Code != 200 {
		t.Fatal("approved follower rejected")
	}
	close := f.publish(t, "alice", "close-friends", "post", []string{m.ID}, apphost.ID(), 0)
	if f.call("GET", "/api/posts/"+close.ID, "bob", nil).Code != 404 {
		t.Fatal("close friends exposed")
	}
	f.call("POST", "/api/profiles/"+f.users["bob"]+"/relations", "alice", map[string]any{"kind": "close", "enabled": true})
	if f.call("GET", "/api/posts/"+close.ID, "bob", nil).Code != 200 {
		t.Fatal("close friend excluded")
	}
}
func TestDraftsNotificationsCSRFAndModerationIsolation(t *testing.T) {
	f := setup(t)
	id := apphost.ID()
	w := f.call("PUT", "/api/drafts/"+id, "alice", map[string]string{"text": "private"})
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	if strings.Contains(f.call("GET", "/api/drafts", "bob", nil).Body.String(), "private") {
		t.Fatal("draft leaked")
	}
	if f.call("PUT", "/api/drafts/"+id, "bob", map[string]string{"text": "overwrite"}).Code != 404 {
		t.Fatal("draft ownership bypass")
	}
	f.a.notify(f.users["alice"], f.users["bob"], "follow", f.users["bob"])
	if strings.Contains(f.call("GET", "/api/notifications", "bob", nil).Body.String(), "follow") {
		t.Fatal("event leaked")
	}
	f.csrf["alice"] = ""
	if f.call("PUT", "/api/drafts/"+apphost.ID(), "alice", map[string]string{}).Code != 403 {
		t.Fatal("CSRF bypass")
	}
	if f.call("GET", "/api/moderation", "bob", nil).Code != 403 {
		t.Fatal("member moderation bypass")
	}
	if f.call("GET", "/api/media", "", nil).Code != 401 {
		t.Fatal("unauthenticated media list")
	}
}
func TestPushAndLinkValidation(t *testing.T) {
	for _, u := range []string{"http://localhost", "https://localhost/a", "https://127.0.0.1/a", "https://fcm.googleapis.com.evil.test/a", "https://user:pass@fcm.googleapis.com/a"} {
		if validPush(u) {
			t.Fatal("unsafe push endpoint", u)
		}
	}
	if !validPush("https://fcm.googleapis.com/send/example") {
		t.Fatal("valid service rejected")
	}
	if safeLink("javascript:alert(1)") || safeLink("https://name:password@example.com") {
		t.Fatal("unsafe link")
	}
}
