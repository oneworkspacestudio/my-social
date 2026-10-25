package social

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oneworkspacestudio/my-social/internal/uploader"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

func nativePermission(t *testing.T, f *fixture) uploader.State {
	t.Helper()
	w := f.call("POST", "/api/imports/instagram/native", "alice", map[string]any{})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var v struct {
		Launch string `json:"launch"`
	}
	json.Unmarshal(w.Body.Bytes(), &v)
	s, e := uploader.ParseLaunch(v.Launch)
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestNativeUploadContinuesWithoutSessionAndRecoversHostRestart(t *testing.T) {
	f := setup(t)
	var cancel context.CancelFunc
	first := true
	offsets := []string{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			offsets = append(offsets, r.URL.Query().Get("offset"))
		}
		f.a.H.Handler().ServeHTTP(w, r)
		if first && r.Method == "PUT" {
			first = false
			cancel()
		}
	})
	server := httptest.NewServer(handler)
	f.a.H.Config.Origin = server.URL
	s := nativePermission(t, f)
	file := filepath.Join(t.TempDir(), "native.zip")
	padding := make([]byte, 6<<20)
	rand.Read(padding)
	data := instagramZIP(t, map[string][]byte{"padding": padding})
	if e := os.WriteFile(file, data, 0600); e != nil {
		t.Fatal(e)
	}
	if e := uploader.SelectFile(&s, file); e != nil {
		t.Fatal(e)
	}
	ctx, stop := context.WithCancel(context.Background())
	cancel = stop
	e := uploader.Transfer(ctx, &s, func(*uploader.State) error { return nil })
	stop()
	if e == nil {
		t.Fatal("interruption not exercised")
	}
	var job string
	f.a.H.DB.QueryRow("SELECT job FROM instagram_native_uploads WHERE id=?", s.Grant).Scan(&job)
	j, e := f.a.loadInstagramJob(job, f.users["alice"])
	if e != nil || j.Received < 2<<20 || j.Received >= j.Size {
		t.Fatal("first receipt missing", j.Received, e)
	}
	// A real Host close/reopen and complete removal of browser sessions cannot
	// remove the encrypted client's scoped permission or committed server bytes.
	addr := server.Listener.Addr().String()
	server.Close()
	f.a.H.DB.Exec("DELETE FROM host_sessions")
	restartInstagramApp(t, f)
	listener, e := net.Listen("tcp", addr)
	if e != nil {
		t.Fatal(e)
	}
	server = httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	defer server.Close()
	offsets = nil
	if e = uploader.Transfer(context.Background(), &s, func(*uploader.State) error { return nil }); e != nil {
		t.Fatal(e)
	}
	if s.Status != "uploaded" || len(offsets) == 0 || offsets[0] == "0" {
		t.Fatal("did not resume committed bytes", s.Status, offsets)
	}
	j, e = f.a.loadInstagramJob(job, f.users["alice"])
	if e != nil || j.Received != j.Size || j.State != "checking" || !j.Native {
		t.Fatal("not saved for server checking", e, j.State, j.Received)
	}
	r := httptest.NewRequest("GET", f.a.H.Config.Origin+"/api/imports/instagram/native/"+s.Grant+"/status", nil)
	r.Header.Set("Authorization", "Bearer "+s.Client)
	w := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(w, r)
	if w.Code != 200 || strings.Contains(w.Body.String(), "items") || strings.Contains(w.Body.String(), "caption") {
		t.Fatal("transfer capability exposed content", w.Code)
	}
}
func nativeCall(f *fixture, s uploader.State, action, method, token string, body []byte, extra string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, f.a.H.Config.Origin+"/api/imports/instagram/native/"+s.Grant+"/"+action, bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("X-Upload-Client", extra)
	w := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(w, r)
	return w
}
func TestNativePermissionIsSingleFileAccountAndTransferOnly(t *testing.T) {
	f := setup(t)
	s := nativePermission(t, f)
	in, _ := json.Marshal(map[string]any{"filename": "a.zip", "size": 22, "fingerprint": strings.Repeat("a", 64)})
	if w := nativeCall(f, s, "claim", "POST", s.Token, in, s.Client); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := nativeCall(f, s, "claim", "POST", s.Token, in, apphost.ID()); w.Code != 403 {
		t.Fatal("grant replay", w.Code)
	}
	if w := nativeCall(f, s, "status", "GET", s.Token, nil, ""); w.Code != 401 {
		t.Fatal("initial permission reused as client", w.Code)
	}
	if w := nativeCall(f, s, "import", "POST", s.Client, []byte(`{"visibility":"public"}`), ""); w.Code != 404 {
		t.Fatal("native client published", w.Code)
	}
	var job string
	f.a.H.DB.QueryRow("SELECT job FROM instagram_native_uploads WHERE id=?", s.Grant).Scan(&job)
	if w := f.call("GET", "/api/imports/instagram/jobs/"+job, "bob", nil); w.Code != 404 {
		t.Fatal("other account sees upload", w.Code)
	}
	f.a.H.DB.Exec("UPDATE host_users SET role='disabled' WHERE id=?", f.users["alice"])
	if w := nativeCall(f, s, "status", "GET", s.Client, nil, ""); w.Code != 403 {
		t.Fatal("disabled account native access", w.Code)
	}
	f.a.H.DB.Exec("UPDATE host_users SET role='owner' WHERE id=?", f.users["alice"])
	f.a.H.DB.Exec("UPDATE instagram_native_uploads SET expires=0 WHERE id=?", s.Grant)
	if w := nativeCall(f, s, "status", "GET", s.Client, nil, ""); w.Code != 401 {
		t.Fatal("expired grant", w.Code)
	}
	r := httptest.NewRequest("POST", f.a.H.Config.Origin+"/api/imports/instagram/native", io.NopCloser(strings.NewReader("{}")))
	r.AddCookie(f.cookies["alice"])
	w := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("grant bypassed CSRF", w.Code)
	}
}

func TestNativeUploadAuthorizedPrivateArchive(t *testing.T) {
	file := os.Getenv("MYSOCIAL_PRIVATE_TEST_ZIP")
	if file == "" {
		t.Skip("private archive test requires explicit user-authorized input")
	}
	f := setup(t)
	server := httptest.NewServer(f.a.H.Handler())
	defer server.Close()
	f.a.H.Config.Origin = server.URL
	s := nativePermission(t, f)
	if e := uploader.SelectFile(&s, file); e != nil {
		t.Fatal("private archive selection failed")
	}
	if e := uploader.Transfer(context.Background(), &s, func(*uploader.State) error { return nil }); e != nil {
		t.Fatal("private native archive transfer failed")
	}
	j, e := f.a.loadInstagramJob(s.Job, f.users["alice"])
	if e != nil || j.Received != 221742001 || j.State != "checking" {
		t.Fatal("private transfer receipt mismatch")
	}
	f.a.runInstagramJob(context.Background())
	j, e = f.a.loadInstagramJob(s.Job, f.users["alice"])
	if e != nil || j.State != "ready" || len(j.Items) != 99 {
		t.Fatal("private server preview did not recognize all posts")
	}
	t.Log("Authorized private archive transferred 221742001 bytes; server recognized99 posts; no owner installation used")
}
