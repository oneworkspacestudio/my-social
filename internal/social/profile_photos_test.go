package social

import (
	"bytes"
	"encoding/json"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"image"
	"mime/multipart"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func photoUpload(f *fixture, user string, data []byte) *httptest.ResponseRecorder {
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	mw.WriteField("kind", "avatars")
	p, _ := mw.CreateFormFile("original", "photo.png")
	p.Write(data)
	mw.Close()
	r := httptest.NewRequest("POST", f.a.H.Config.Origin+"/api/media", &b)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.Header.Set("Origin", f.a.H.Config.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf[user])
	r.AddCookie(f.cookies[user])
	w := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(w, r)
	return w
}
func TestProfilePhotoOwnershipVisibilityAndReplacement(t *testing.T) {
	f := setup(t)
	upload := photoUpload(f, "alice", sample())
	if upload.Code != 201 {
		t.Fatal(upload.Code, upload.Body.String())
	}
	var m Media
	json.Unmarshal(upload.Body.Bytes(), &m)
	if w := f.call("GET", "/api/media/"+m.ID, "bob", nil); w.Code != 404 {
		t.Fatal("unselected photo exposed", w.Code)
	}
	if w := f.call("PUT", "/api/profile/photo", "bob", map[string]string{"media": m.ID}); w.Code != 400 {
		t.Fatal("another member selected owner's photo", w.Code)
	}
	post := f.upload(t, "alice")
	if f.call("PUT", "/api/profile/photo", "alice", map[string]string{"media": post.ID}).Code != 400 {
		t.Fatal("non-avatar accepted")
	}
	if w := f.call("PUT", "/api/profile/photo", "alice", map[string]string{"media": m.ID}); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if f.a.getProfile(f.users["alice"]).Avatar != m.ID {
		t.Fatal("photo not persisted")
	}
	public := f.call("GET", "/api/media/"+m.ID, "bob", nil)
	if public.Code != 200 || public.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("profile photo unavailable", public.Code)
	}
	cfg, _, e := image.DecodeConfig(bytes.NewReader(public.Body.Bytes()))
	if e != nil || cfg.Width != 512 || cfg.Height != 512 {
		t.Fatal("not bounded square", cfg, e)
	}
	if f.call("GET", "/api/media/"+m.ID+"?original=1", "bob", nil).Code != 404 {
		t.Fatal("original leaked")
	}
	if f.call("GET", "/api/media/"+m.ID, "", nil).Code != 401 {
		t.Fatal("anonymous avatar access")
	}
	f.call("POST", "/api/profiles/"+f.users["bob"]+"/relations", "alice", map[string]any{"kind": "block", "enabled": true})
	if f.call("GET", "/api/media/"+m.ID, "bob", nil).Code != 404 {
		t.Fatal("blocked avatar access")
	}
	f.call("POST", "/api/profiles/"+f.users["bob"]+"/relations", "alice", map[string]any{"kind": "block", "enabled": false})
	next := photoUpload(f, "alice", sample())
	var replacement Media
	json.Unmarshal(next.Body.Bytes(), &replacement)
	f.call("PUT", "/api/profile/photo", "alice", map[string]string{"media": replacement.ID})
	if f.call("GET", "/api/media/"+m.ID, "bob", nil).Code != 404 {
		t.Fatal("replaced photo still shared")
	}
	f.call("DELETE", "/api/profile/photo", "alice", nil)
	if f.a.getProfile(f.users["alice"]).Avatar != "" || f.call("GET", "/api/media/"+replacement.ID, "bob", nil).Code != 404 {
		t.Fatal("removal did not revoke sharing")
	}
	if f.call("GET", "/api/media/"+m.ID+"?original=1", "alice", nil).Code != 200 {
		t.Fatal("owner's original was removed")
	}
	if photoUpload(f, "alice", []byte("not a photo")).Code != 415 {
		t.Fatal("invalid photo accepted")
	}
	if photoUpload(f, "alice", append(sample(), make([]byte, 10<<20)...)).Code != 415 {
		t.Fatal("oversized avatar accepted")
	}
}
func TestProfilePhotoAdditiveMigration(t *testing.T) {
	h, e := apphost.New(apphost.Config{ID: "social", Name: "My Social", Data: filepath.Join(t.TempDir(), "Private"), Listen: "127.0.0.1:29113", Origin: "http://localhost:29113"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	if e = h.Migrate(1, schema); e != nil {
		t.Fatal(e)
	}
	id := apphost.Identity("local:social", "existing")
	_, e = h.DB.Exec("INSERT INTO host_users(id,issuer,subject,name,email,role) VALUES(?,?,?,?,?,?)", id, "local:social", "existing", "Existing", "existing@example.test", "owner")
	if e != nil {
		t.Fatal(e)
	}
	h.DB.Exec("INSERT INTO profiles(user_id,handle,bio) VALUES(?,?,?)", id, "existing", "Kept bio")
	a, e := New(h)
	if e != nil {
		t.Fatal(e)
	}
	p := a.getProfile(id)
	if p.Bio != "Kept bio" || p.Avatar != "" {
		t.Fatal("existing profile changed", p)
	}
	config := h.Config
	if e = h.Close(); e != nil {
		t.Fatal(e)
	}
	h, e = apphost.New(config, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	a, e = New(h)
	if e != nil {
		t.Fatal("upgraded application must restart without replaying migration 1", e)
	}
	if p = a.getProfile(id); p.Bio != "Kept bio" || p.Handle != "existing" {
		t.Fatal("restart changed existing profile", p)
	}
	var version int
	if e = h.DB.QueryRow("SELECT max(version) FROM host_migrations").Scan(&version); e != nil || version != 4 {
		t.Fatal("restart changed migration version", version, e)
	}
}
