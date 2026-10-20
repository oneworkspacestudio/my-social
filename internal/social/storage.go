package social

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"golang.org/x/image/draw"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Media struct {
	ID       string          `json:"id"`
	Owner    string          `json:"owner"`
	Kind     string          `json:"kind"`
	Original string          `json:"-"`
	Rendered string          `json:"-"`
	MIME     string          `json:"mime"`
	SHA      string          `json:"sha256"`
	Bytes    int64           `json:"bytes"`
	State    string          `json:"state"`
	Created  int64           `json:"created"`
	Project  json.RawMessage `json:"project"`
}

func (a *App) storageCheck(ctx context.Context) error {
	if e := safeRoot(a.Root); e != nil {
		return e
	}
	if a.H.Config.WorkspaceURL == "" && os.Getenv("MYSOCIAL_STANDALONE_STORAGE") == "isolated" {
		return nil
	}
	u, e := url.Parse(a.Guard)
	if e != nil || u.Scheme != "http" || u.User != nil || u.Hostname() == "" || net.ParseIP(u.Hostname()) == nil || !net.ParseIP(u.Hostname()).IsLoopback() || a.GuardToken == "" {
		return errors.New("storage guard unavailable")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", a.Guard, nil)
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", "Bearer "+a.GuardToken)
	user := apphost.Current((&http.Request{}).WithContext(ctx)).User
	if user.Subject != "" {
		req.Header.Set("X-Social-Subject", user.Subject)
		req.Header.Set("X-Social-Issuer", user.Issuer)
	}
	c := http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, e := c.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("storage isolation not confirmed")
	}
	var contract struct {
		Contract string `json:"contract"`
	}
	if json.NewDecoder(resp.Body).Decode(&contract) != nil || contract.Contract != "quantum-social-v2" {
		return errors.New("Workspace 48 scoped storage is required")
	}
	return nil
}
func (a *App) media(id string) (Media, error) {
	var m Media
	var raw string
	e := a.H.DB.QueryRow("SELECT id,owner,kind,original,rendered,mime,sha,bytes,state,created,project FROM media WHERE id=?", id).Scan(&m.ID, &m.Owner, &m.Kind, &m.Original, &m.Rendered, &m.MIME, &m.SHA, &m.Bytes, &m.State, &m.Created, &raw)
	m.Project = json.RawMessage(raw)
	return m, e
}
func safeImage(data []byte) ([]byte, error) {
	cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 24000000 {
		return nil, errors.New("use a JPEG, PNG or GIF up to 24 megapixels")
	}
	img, _, e := image.Decode(bytes.NewReader(data))
	if e != nil {
		return nil, e
	}
	w, h := cfg.Width, cfg.Height
	if w > 2048 || h > 2048 {
		ratio := 2048.0 / float64(max(w, h))
		w = int(float64(w) * ratio)
		h = int(float64(h) * ratio)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(canvas, canvas.Bounds(), img, img.Bounds(), draw.Src, nil)
	var out bytes.Buffer
	e = jpeg.Encode(&out, canvas, &jpeg.Options{Quality: 90})
	return out.Bytes(), e
}
func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	if e := a.storageCheck(r.Context()); e != nil {
		apphost.Problem(w, 503, "Connect the protected My Photos/My Social storage before uploading.")
		return
	}
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	var used int64
	a.H.DB.QueryRow("SELECT coalesce(sum(bytes),0) FROM media WHERE owner=?", uid(r)).Scan(&used)
	if used > 2<<30 {
		apphost.Problem(w, 413, "Your 2 GiB preview media quota is full.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 150<<20)
	reader, e := r.MultipartReader()
	if e != nil {
		apphost.Problem(w, 400, "Choose media to upload")
		return
	}
	var original, rendered []byte
	project := "{}"
	kind := "posts"
	for {
		p, e := reader.NextPart()
		if e == io.EOF {
			break
		}
		if e != nil {
			apphost.Problem(w, 400, "Upload interrupted; retry the file")
			return
		}
		limit := int64(100 << 20)
		if p.FormName() == "project" {
			limit = 512 << 10
		}
		if p.FormName() == "kind" {
			limit = 30
		}
		b, e := io.ReadAll(io.LimitReader(p, limit+1))
		p.Close()
		if e != nil || int64(len(b)) > limit {
			apphost.Problem(w, 413, "Media exceeds the upload limit")
			return
		}
		switch p.FormName() {
		case "original":
			original = b
		case "rendered":
			rendered = b
		case "project":
			if !json.Valid(b) {
				apphost.Problem(w, 400, "Invalid editor project")
				return
			}
			project = string(b)
		case "kind":
			kind = string(b)
		default:
			apphost.Problem(w, 400, "Unknown media field")
			return
		}
	}
	if len(original) == 0 || len(original) > 100<<20 {
		apphost.Problem(w, 400, "Choose an original file up to 100 MiB")
		return
	}
	if kind != "posts" && kind != "stories" && kind != "reels" && kind != "avatars" {
		apphost.Problem(w, 400, "Choose Post, Story or Reel")
		return
	}
	mime := http.DetectContentType(original)
	ext := ".jpg"
	state := "ready"
	if strings.HasPrefix(mime, "image/") {
		source := original
		if len(rendered) > 0 {
			source = rendered
		}
		rendered, e = safeImage(source)
		if e != nil {
			apphost.Problem(w, 415, e.Error())
			return
		}
		if _, e = safeImage(original); e != nil {
			apphost.Problem(w, 415, e.Error())
			return
		}
		mime = "image/jpeg"
	} else if mime == "video/mp4" || mime == "video/webm" {
		if a.FFprobe == "" || a.FFmpeg == "" {
			apphost.Problem(w, 503, "Video publishing needs configured FFmpeg and FFprobe. Keep this clip in your device draft.")
			return
		}
		ext = ".video"
		rendered = nil
		state = "uploaded"
		mime = "video/mp4"
	} else {
		apphost.Problem(w, 415, "Use JPEG, PNG, GIF, MP4 or WebM media")
		return
	}
	if used+int64(len(original)) > 2<<30 {
		apphost.Problem(w, 413, "Your 2 GiB preview media quota is full.")
		return
	}
	id := apphost.ID()
	rel := path.Join("users", uid(r), kind, id+ext)
	derivative := path.Join("generated", "previews", uid(r), id+".jpg")
	if state != "ready" {
		derivative = ""
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
	if _, e = a.H.DB.Exec("INSERT INTO upload_intents VALUES(?,?,?,?,?)", id, uid(r), "writing", rel, a.now().Unix()); e != nil {
		a.fail(w, e)
		return
	}
	committed := false
	defer func() {
		if !committed {
			root.Remove(rel)
			if derivative != "" {
				root.Remove(derivative)
			}
			a.H.DB.Exec("DELETE FROM upload_intents WHERE id=?", id)
		}
	}()
	if e = writeRoot(root, rel, original); e != nil {
		a.fail(w, e)
		return
	}
	if len(rendered) > 0 {
		if e = writeRoot(root, derivative, rendered); e != nil {
			a.fail(w, e)
			return
		}
	}
	sum := sha256.Sum256(original)
	m := Media{ID: id, Owner: uid(r), Kind: kind, Original: rel, Rendered: derivative, MIME: mime, SHA: hex.EncodeToString(sum[:]), Bytes: int64(len(original)), State: state, Created: a.now().Unix(), Project: json.RawMessage(project)}
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.fail(w, e)
		return
	}
	defer tx.Rollback()
	_, e = tx.Exec("INSERT INTO media VALUES(?,?,?,?,?,?,?,?,?,?,?)", m.ID, m.Owner, m.Kind, m.Original, m.Rendered, m.MIME, m.SHA, m.Bytes, m.State, m.Created, project)
	if e != nil {
		a.fail(w, e)
		return
	}
	if _, e = tx.Exec("UPDATE upload_intents SET state='complete' WHERE id=?", id); e != nil {
		a.fail(w, e)
		return
	}
	if e = tx.Commit(); e != nil {
		a.fail(w, e)
		return
	}
	committed = true
	apphost.JSON(w, 201, m)
}
func writeRoot(root *os.Root, rel string, b []byte) error {
	if e := root.MkdirAll(filepath.FromSlash(path.Dir(rel)), 0700); e != nil {
		return e
	}
	f, e := root.OpenFile(filepath.FromSlash(rel), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil {
		return e
	}
	return closeErr
}
func (a *App) serveMedia(w http.ResponseWriter, r *http.Request) {
	m, e := a.media(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	original := r.URL.Query().Get("original") == "1"
	if original && m.Owner != uid(r) {
		http.NotFound(w, r)
		return
	}
	if m.Owner != uid(r) && !a.canSeeMedia(uid(r), m.ID) {
		http.NotFound(w, r)
		return
	}
	if e = a.storageCheck(r.Context()); e != nil {
		apphost.Problem(w, 503, "My Photos is unavailable. Retry when it reconnects.")
		return
	}
	rel := m.Rendered
	if original {
		rel = m.Original
	}
	if rel == "" {
		apphost.Problem(w, 409, "This clip is waiting for rendering")
		return
	}
	root, e := os.OpenRoot(a.Root)
	if e != nil {
		a.fail(w, e)
		return
	}
	defer root.Close()
	f, e := root.Open(filepath.FromSlash(rel))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		a.fail(w, e)
		return
	}
	w.Header().Set("Content-Type", m.MIME)
	w.Header().Set("Content-Disposition", `inline; filename="media`+path.Ext(rel)+`"`)
	w.Header().Set("Cache-Control", "private, no-store")
	if original {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="original`+path.Ext(rel)+`"`)
	}
	http.ServeContent(w, r, "media", info.ModTime(), f)
}
func (a *App) canSeeMedia(user, id string) bool {
	rows, e := a.H.DB.Query("SELECT id FROM posts WHERE EXISTS(SELECT 1 FROM json_each(posts.media) WHERE value=?)", id)
	if e != nil {
		return false
	}
	ids := []string{}
	for rows.Next() {
		var x string
		rows.Scan(&x)
		ids = append(ids, x)
	}
	rows.Close()
	for _, x := range ids {
		if p, e := a.post(x); e == nil && a.visible(user, p) {
			return true
		}
	}
	return false
}
func (a *App) library(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT id FROM media WHERE owner=? ORDER BY created DESC LIMIT 200", uid(r))
	if e != nil {
		a.fail(w, e)
		return
	}
	ids := []string{}
	for rows.Next() {
		var x string
		rows.Scan(&x)
		ids = append(ids, x)
	}
	rows.Close()
	out := []Media{}
	for _, x := range ids {
		m, e := a.media(x)
		if e == nil {
			out = append(out, m)
		}
	}
	apphost.JSON(w, 200, out)
}
