package social

import (
	"encoding/json"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestStructuredTwoClipRender(t *testing.T) {
	ff := os.Getenv("MYSOCIAL_TEST_FFMPEG")
	fp := os.Getenv("MYSOCIAL_TEST_FFPROBE")
	if ff == "" || fp == "" {
		t.Skip("optional operator FFmpeg integration test; configure both helper paths")
	}
	f := setup(t)
	f.a.FFmpeg = ff
	f.a.FFprobe = fp
	os.MkdirAll(f.a.Root, 0700)
	ids := []string{}
	for _, color := range []string{"blue", "green"} {
		id := apphost.ID()
		rel := filepath.Join("users", f.users["alice"], "reels", id+".mp4")
		dest := filepath.Join(f.a.Root, rel)
		os.MkdirAll(filepath.Dir(dest), 0700)
		cmd := exec.Command(ff, "-nostdin", "-v", "error", "-f", "lavfi", "-i", "color=c="+color+":s=320x480:d=1", "-f", "lavfi", "-i", "sine=frequency=440:duration=1", "-shortest", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-y", dest)
		if out, e := cmd.CombinedOutput(); e != nil {
			t.Fatal(e, string(out))
		}
		info, _ := os.Stat(dest)
		_, e := f.a.H.DB.Exec("INSERT INTO media VALUES(?,?,?,?,?,?,?,?,?,?,?)", id, f.users["alice"], "reels", filepath.ToSlash(rel), "", "video/mp4", "fixture", info.Size(), "uploaded", f.a.now().Unix(), "{}")
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, id)
	}
	w := f.call("POST", "/api/jobs", "alice", EDL{Clips: []Clip{{Media: ids[1], Start: 0.1, End: 0.8, Speed: 1}, {Media: ids[0], Start: 0.2, End: 0.9, Speed: 1.5}}, Mute: true})
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	var job map[string]string
	json.Unmarshal(w.Body.Bytes(), &job)
	f.a.runJob(t.Context())
	var state, err string
	f.a.H.DB.QueryRow("SELECT state,error FROM jobs WHERE id=?", job["id"]).Scan(&state, &err)
	if state != "complete" {
		t.Fatal(state, err)
	}
	m, e := f.a.media(job["media"])
	if e != nil || !f.a.readyMedia(m) {
		t.Fatal("render output not ready", e)
	}
	p := f.publish(t, "alice", "public", "reel", []string{m.ID}, apphost.ID(), 0)
	if f.call("GET", "/api/posts/"+p.ID, "bob", nil).Code != 200 {
		t.Fatal("rendered Reel not published")
	}
}
