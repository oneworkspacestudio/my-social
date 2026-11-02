package social

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"
)

// A stalled sender must not reserve the single storage writer for everyone.
func TestSlowUploadDoesNotBlockAnotherCreator(t *testing.T) {
	f := setup(t)
	reader, writer := io.Pipe()
	req := httptest.NewRequest("POST", f.a.H.Config.Origin+"/api/media", reader)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=slow-phone")
	req.Header.Set("Origin", f.a.H.Config.Origin)
	req.Header.Set("X-CSRF-Token", f.csrf["alice"])
	req.AddCookie(f.cookies["alice"])
	finished := make(chan struct{})
	go func() { defer close(finished); f.a.H.Handler().ServeHTTP(httptest.NewRecorder(), req) }()
	// Pipe.Write waits until the server has begun consuming the request body.
	if _, e := writer.Write([]byte("--slow-phone\r\n")); e != nil {
		t.Fatal(e)
	}
	defer func() { writer.Close(); <-finished }()
	other := make(chan int, 1)
	go func() {
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		p, _ := mw.CreateFormFile("original", "photo.png")
		p.Write(sample())
		mw.Close()
		r := httptest.NewRequest("POST", f.a.H.Config.Origin+"/api/media", &b)
		r.Header.Set("Content-Type", mw.FormDataContentType())
		r.Header.Set("Origin", f.a.H.Config.Origin)
		r.Header.Set("X-CSRF-Token", f.csrf["bob"])
		r.AddCookie(f.cookies["bob"])
		w := httptest.NewRecorder()
		f.a.H.Handler().ServeHTTP(w, r)
		other <- w.Code
	}()
	select {
	case code := <-other:
		if code != 201 {
			t.Fatalf("second creator upload: %d", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("stalled sender blocked a different creator")
	}
}

func TestRenderedPhotoCannotHideCorruptOriginal(t *testing.T) {
	f := setup(t)
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	p, _ := mw.CreateFormFile("original", "broken.png")
	p.Write(sample()[:40]) // valid PNG dimensions, truncated pixel data
	p, _ = mw.CreateFormFile("rendered", "preview.png")
	p.Write(sample())
	mw.Close()
	r := httptest.NewRequest("POST", f.a.H.Config.Origin+"/api/media", &b)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.Header.Set("Origin", f.a.H.Config.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf["alice"])
	r.AddCookie(f.cookies["alice"])
	w := httptest.NewRecorder()
	f.a.H.Handler().ServeHTTP(w, r)
	if w.Code != 415 {
		t.Fatalf("corrupt original accepted: %d %s", w.Code, w.Body.String())
	}
	var n int
	f.a.H.DB.QueryRow("SELECT count(*) FROM media").Scan(&n)
	if n != 0 {
		t.Fatal("invalid original was committed")
	}
}
