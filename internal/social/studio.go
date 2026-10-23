package social

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Clip struct {
	Media string  `json:"media"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Speed float64 `json:"speed"`
}
type EDL struct {
	Clips   []Clip `json:"clips"`
	Mute    bool   `json:"mute"`
	Quality string `json:"quality"`
}

func (a *App) createJob(w http.ResponseWriter, r *http.Request) {
	var edl EDL
	if !apphost.Decode(w, r, &edl) {
		return
	}
	if a.FFmpeg == "" || a.FFprobe == "" {
		apphost.Problem(w, 503, "Ask the owner to configure FFmpeg and FFprobe")
		return
	}
	if len(edl.Clips) < 1 || len(edl.Clips) > 10 {
		apphost.Problem(w, 400, "Choose 1–10 clips")
		return
	}
	duration := 0.0
	for _, c := range edl.Clips {
		m, e := a.media(c.Media)
		if e != nil || m.Owner != uid(r) || !strings.HasPrefix(m.MIME, "video/") || c.Start < 0 || c.End <= c.Start || c.End > 600 || c.Speed < 0.5 || c.Speed > 2 {
			apphost.Problem(w, 400, "Use your own video clips with valid trim times and speeds 0.5–2")
			return
		}
		duration += (c.End - c.Start) / c.Speed
	}
	if duration > 90 {
		apphost.Problem(w, 400, "This preview supports Reels up to 90 seconds")
		return
	}
	if a.exists("SELECT count(*) FROM jobs WHERE owner=? AND state IN ('queued','running') HAVING count(*)>=3", uid(r)) {
		apphost.Problem(w, 429, "Wait for your current renders to finish")
		return
	}
	id := apphost.ID()
	output := apphost.ID()
	_, e := a.H.DB.Exec("INSERT INTO jobs VALUES(?,?,?,? ,'queued',0,'',?)", id, uid(r), output, encode(edl), a.now().Unix())
	if e != nil {
		a.fail(w, e)
		return
	}
	apphost.JSON(w, 202, map[string]string{"id": id, "media": output})
}
func (a *App) jobs(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT id,media_id,state,progress,error FROM jobs WHERE owner=? ORDER BY created DESC LIMIT 30", uid(r))
	if e != nil {
		a.fail(w, e)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, media, state, err string
		var progress int
		rows.Scan(&id, &media, &state, &progress, &err)
		out = append(out, map[string]any{"id": id, "media": media, "state": state, "progress": progress, "error": err})
	}
	apphost.JSON(w, 200, out)
}
func (a *App) runJob(ctx context.Context) {
	var id, owner, media, raw string
	if a.H.DB.QueryRow("SELECT id,owner,media_id,edl FROM jobs WHERE state='queued' ORDER BY created LIMIT 1").Scan(&id, &owner, &media, &raw) != nil {
		return
	}
	if a.storageCheck(ctx) != nil {
		return
	}
	res, e := a.H.DB.Exec("UPDATE jobs SET state='running',progress=5 WHERE id=? AND state='queued'", id)
	if e != nil {
		return
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return
	}
	var edl EDL
	if json.Unmarshal([]byte(raw), &edl) != nil {
		return
	}
	err := a.render(ctx, id, owner, media, edl)
	if err != nil {
		a.H.DB.Exec("UPDATE jobs SET state='failed',error='Render failed. Check clip duration, FFmpeg availability and free storage; retry from the draft.' WHERE id=?", id)
	} else {
		a.H.DB.Exec("UPDATE jobs SET state='complete',progress=100 WHERE id=?", id)
	}
}
func number(f float64) string { return strconv.FormatFloat(f, 'f', 3, 64) }
func (a *App) render(ctx context.Context, job, owner, id string, edl EDL) error {
	// Inputs are owned immutable objects; every argument is separate. No shell.
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	if len(edl.Clips) == 0 {
		return errors.New("no clips")
	}
	root, e := os.OpenRoot(a.Root)
	if e != nil {
		return e
	}
	defer root.Close()
	rel := filepath.Join("generated", "transcodes", owner, id+".mp4")
	if e = root.MkdirAll(filepath.Dir(rel), 0700); e != nil {
		return e
	}
	output := filepath.Join(a.Root, rel)
	args := []string{"-nostdin", "-hide_banner", "-loglevel", "error", "-max_alloc", "134217728", "-y"}
	filters := []string{}
	inputs := ""
	total := 0.0
	for i, c := range edl.Clips {
		m, e := a.media(c.Media)
		if e != nil || m.Owner != owner {
			return errors.New("media unavailable")
		}
		source := filepath.Join(a.Root, filepath.FromSlash(m.Original))
		if !withinRoot(a.Root, source) {
			return errors.New("unsafe input")
		}
		probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		probe := exec.CommandContext(probeCtx, a.FFprobe, "-v", "error", "-protocol_whitelist", "file,pipe", "-show_entries", "format=duration", "-of", "json", source)
		probe.Dir = a.Root
		b, e := probe.Output()
		cancel()
		if e != nil || len(b) > 8192 {
			return errors.New("cannot probe")
		}
		var info struct {
			Format struct {
				Duration string `json:"duration"`
			} `json:"format"`
		}
		json.Unmarshal(b, &info)
		duration, _ := strconv.ParseFloat(info.Format.Duration, 64)
		if c.End > duration+0.1 || c.End <= c.Start {
			return errors.New("trim beyond clip")
		}
		args = append(args, "-protocol_whitelist", "file,pipe", "-threads", "2", "-ss", number(c.Start), "-t", number(c.End-c.Start), "-i", source)
		filters = append(filters, fmt.Sprintf("[%d:v]scale=720:1280:force_original_aspect_ratio=decrease,pad=720:1280:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=30,setpts=(PTS-STARTPTS)/%s[v%d]", i, number(c.Speed), i))
		inputs += fmt.Sprintf("[v%d]", i)
		total += (c.End - c.Start) / c.Speed
	}
	filters = append(filters, inputs+fmt.Sprintf("concat=n=%d:v=1:a=0[outv]", len(edl.Clips)))
	args = append(args, "-filter_complex_threads", "1", "-filter_complex", strings.Join(filters, ";"), "-map", "[outv]")
	// Single-clip original audio is supported. Multi-clip mixing is an open gate.
	if !edl.Mute && len(edl.Clips) == 1 {
		args = append(args, "-map", "0:a?", "-af", "atempo="+number(edl.Clips[0].Speed), "-c:a", "aac", "-b:a", "128k")
	} else {
		args = append(args, "-an")
	}
	crf := "25"
	if edl.Quality == "low" {
		crf = "30"
	}
	args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-crf", crf, "-pix_fmt", "yuv420p", "-movflags", "+faststart", "-threads", "2", "-t", number(total), "-map_metadata", "-1", output)
	runCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(runCtx, a.FFmpeg, args...)
	cmd.Dir = a.Root
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	e = cmd.Run()
	if e != nil {
		root.Remove(rel)
		return e
	}
	info, e := root.Stat(rel)
	if e != nil || info.Size() == 0 || info.Size() > 120<<20 {
		root.Remove(rel)
		return errors.New("invalid output")
	}
	// Render records reference an existing original, never copy it elsewhere.
	first, _ := a.media(edl.Clips[0].Media)
	_, e = a.H.DB.Exec("INSERT INTO media VALUES(?,?,?,?,?,?,?,?,?,?,?)", id, owner, "reels", first.Original, filepath.ToSlash(rel), "video/mp4", first.SHA, int64(0), "ready", a.now().Unix(), encode(edl))
	return e
}
func withinRoot(root, p string) bool {
	resolved, e := filepath.EvalSymlinks(p)
	if e != nil {
		return false
	}
	rel, e := filepath.Rel(root, resolved)
	return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func (a *App) report(w http.ResponseWriter, r *http.Request) {
	var in struct{ Post, Reason string }
	if !apphost.Decode(w, r, &in) {
		return
	}
	p, e := a.post(in.Post)
	if e != nil || !a.visible(uid(r), p) {
		http.NotFound(w, r)
		return
	}
	if len(in.Reason) < 3 || len(in.Reason) > 1000 {
		apphost.Problem(w, 400, "Explain the report in 3–1,000 characters")
		return
	}
	a.H.DB.Exec("INSERT INTO reports VALUES(?,?,?,?,'open',?)", apphost.ID(), uid(r), p.ID, in.Reason, a.now().Unix())
	apphost.JSON(w, 201, map[string]bool{"ok": true})
}
func (a *App) moderation(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT id,post_id,reason,created FROM reports WHERE state='open' ORDER BY created LIMIT 100")
	if e != nil {
		a.fail(w, e)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, post, reason string
		var at int64
		rows.Scan(&id, &post, &reason, &at)
		out = append(out, map[string]any{"id": id, "post": post, "reason": reason, "created": at})
	}
	apphost.JSON(w, 200, out)
}
func (a *App) moderate(w http.ResponseWriter, r *http.Request) {
	var in struct{ Action string }
	if !apphost.Decode(w, r, &in) {
		return
	}
	if in.Action != "remove" && in.Action != "dismiss" {
		apphost.Problem(w, 400, "Choose Remove or Dismiss")
		return
	}
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.fail(w, e)
		return
	}
	defer tx.Rollback()
	if in.Action == "remove" {
		_, e = tx.Exec("UPDATE posts SET state='removed' WHERE id=(SELECT post_id FROM reports WHERE id=?)", r.PathValue("id"))
		if e != nil {
			a.fail(w, e)
			return
		}
	}
	_, e = tx.Exec("UPDATE reports SET state=? WHERE id=?", in.Action, r.PathValue("id"))
	if e != nil {
		a.fail(w, e)
		return
	}
	if e = tx.Commit(); e != nil {
		a.fail(w, e)
		return
	}
	a.H.Audit(uid(r), "moderation."+in.Action, r.PathValue("id"))
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) exportAccount(w http.ResponseWriter, r *http.Request) {
	if e := a.storageCheck(r.Context()); e != nil {
		a.fail(w, e)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="my-social-export.zip"`)
	z := zip.NewWriter(w)
	defer z.Close()
	f, e := z.Create("profile.json")
	if e != nil {
		return
	}
	json.NewEncoder(f).Encode(a.getProfile(uid(r)))
	rows, e := a.H.DB.Query("SELECT id FROM posts WHERE owner=?", uid(r))
	if e != nil {
		return
	}
	ids := []string{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	posts := []Post{}
	for _, id := range ids {
		p, e := a.post(id)
		if e == nil {
			posts = append(posts, p)
		}
	}
	f, e = z.Create("posts.json")
	if e != nil {
		return
	}
	json.NewEncoder(f).Encode(posts)
	rows, e = a.H.DB.Query("SELECT DISTINCT original FROM media WHERE owner=?", uid(r))
	if e != nil {
		return
	}
	paths := []string{}
	for rows.Next() {
		var p string
		rows.Scan(&p)
		paths = append(paths, p)
	}
	rows.Close()
	root, e := os.OpenRoot(a.Root)
	if e != nil {
		return
	}
	defer root.Close()
	for _, p := range paths {
		f, e := root.Open(filepath.FromSlash(p))
		if e != nil {
			continue
		}
		out, e := z.Create("originals/" + filepath.Base(p))
		if e == nil {
			io.Copy(out, io.LimitReader(f, 101<<20))
		}
		f.Close()
	}
}
func (a *App) deleteAccount(w http.ResponseWriter, r *http.Request) {
	var in struct{ Confirmation string }
	if !apphost.Decode(w, r, &in) {
		return
	}
	if in.Confirmation != "DELETE MY SOCIAL ACCOUNT" {
		apphost.Problem(w, 400, "Type DELETE MY SOCIAL ACCOUNT to confirm")
		return
	}
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.fail(w, e)
		return
	}
	defer tx.Rollback()
	for _, q := range []string{"UPDATE posts SET state='deleted' WHERE owner=?", "DELETE FROM host_sessions WHERE user_id=?", "DELETE FROM drafts WHERE owner=?", "DELETE FROM push_subscriptions WHERE owner=?", "DELETE FROM preferences WHERE user_id=?", "DELETE FROM profile_photos WHERE user_id=?", "DELETE FROM highlights WHERE owner=?", "DELETE FROM events WHERE user_id=?", "UPDATE profiles SET bio='',links='[]',private=1 WHERE user_id=?", "UPDATE host_users SET role='disabled',name='Deleted account',email='' WHERE id=?"} {
		if _, e = tx.Exec(q, uid(r)); e != nil {
			a.fail(w, e)
			return
		}
	}
	if e = tx.Commit(); e != nil {
		a.fail(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]string{"status": "disabled", "retention": "Originals remain protected for owner-managed retention and backup. This does not delete your MyAccounts identity."})
}
