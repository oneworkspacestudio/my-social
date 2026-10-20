package social

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

func (a *App) readyMedia(m Media) bool {
	if m.State != "ready" || m.Rendered == "" {
		return false
	}
	root, e := os.OpenRoot(a.Root)
	if e != nil {
		return false
	}
	defer root.Close()
	for _, p := range []string{m.Original, m.Rendered} {
		f, e := root.Open(filepath.FromSlash(p))
		if e != nil {
			return false
		}
		info, e := f.Stat()
		f.Close()
		if e != nil || !info.Mode().IsRegular() || info.Size() < 1 {
			return false
		}
	}
	return true
}
func (a *App) publishScheduled(ctx context.Context) {
	if a.storageCheck(ctx) != nil {
		return
	}
	rows, e := a.H.DB.Query("SELECT id FROM posts WHERE state='scheduled' AND publish_at<=?", a.now().Unix())
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
	for _, id := range ids {
		p, e := a.post(id)
		if e != nil {
			continue
		}
		ready := true
		for _, id := range p.Media {
			m, e := a.media(id)
			if e != nil || !a.readyMedia(m) {
				ready = false
				break
			}
		}
		if !ready {
			continue
		}
		tx, e := a.H.DB.Begin()
		if e != nil {
			return
		}
		if _, e = tx.Exec("UPDATE posts SET state='published' WHERE id=? AND state='scheduled'", p.ID); e == nil {
			_, e = tx.Exec("INSERT OR IGNORE INTO events(user_id,actor,category,object,created) SELECT actor,?,'post',?,? FROM relations WHERE target=? AND kind='follow' AND state='accepted'", p.Owner, p.ID, a.now().Unix(), p.Owner)
		}
		if e == nil {
			e = tx.Commit()
		} else {
			tx.Rollback()
		}
		if e != nil {
			return
		}
	}
}
func (a *App) recoverJobs() error {
	_, e := a.H.DB.Exec("UPDATE jobs SET state='queued',progress=0,error='' WHERE state='running'")
	return e
}

var errInvalidStory = errors.New("invalid Story elements")
