package social

import (
	"database/sql"
	"encoding/json"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"strings"
	"time"
)

type Post struct {
	ID         string          `json:"id"`
	Owner      string          `json:"owner"`
	Kind       string          `json:"kind"`
	Caption    string          `json:"caption"`
	Alt        string          `json:"alt"`
	Location   string          `json:"location"`
	Visibility string          `json:"visibility"`
	State      string          `json:"state"`
	Created    int64           `json:"created"`
	PublishAt  int64           `json:"publishAt"`
	Expires    int64           `json:"expires"`
	Media      []string        `json:"media"`
	Extras     json.RawMessage `json:"extras"`
	Profile    any             `json:"profile"`
	Likes      int             `json:"likes"`
	Comments   int             `json:"comments"`
	Liked      bool            `json:"liked"`
	Saved      bool            `json:"saved"`
}

func (a *App) post(id string) (Post, error) {
	var p Post
	var media, extras string
	e := a.H.DB.QueryRow("SELECT id,owner,kind,caption,alt,location,visibility,state,created,publish_at,expires,media,extras FROM posts WHERE id=?", id).Scan(&p.ID, &p.Owner, &p.Kind, &p.Caption, &p.Alt, &p.Location, &p.Visibility, &p.State, &p.Created, &p.PublishAt, &p.Expires, &media, &extras)
	json.Unmarshal([]byte(media), &p.Media)
	p.Extras = json.RawMessage(extras)
	return p, e
}
func (a *App) exists(q string, args ...any) bool {
	var n int
	a.H.DB.QueryRow(q, args...).Scan(&n)
	return n > 0
}
func (a *App) blocked(u, v string) bool {
	return a.exists("SELECT count(*) FROM relations WHERE kind='block' AND ((actor=? AND target=?) OR (actor=? AND target=?))", u, v, v, u)
}
func (a *App) visible(user string, p Post) bool {
	if user == p.Owner {
		return p.State != "removed"
	}
	if (p.State != "published" && !(p.State == "archived" && p.Kind == "story" && a.exists("SELECT count(*) FROM highlights WHERE owner=? AND EXISTS(SELECT 1 FROM json_each(highlights.posts) WHERE value=?)", p.Owner, p.ID))) || p.PublishAt > a.now().Unix() || a.blocked(user, p.Owner) {
		return false
	}
	if p.Kind == "story" && p.Expires <= a.now().Unix() {
		return a.exists("SELECT count(*) FROM highlights WHERE owner=? AND EXISTS(SELECT 1 FROM json_each(highlights.posts) WHERE value=?)", p.Owner, p.ID) && a.audience(user, p)
	}
	return a.audience(user, p)
}
func (a *App) audience(u string, p Post) bool {
	if p.Visibility == "private" {
		return false
	}
	if p.Visibility == "close-friends" {
		return a.exists("SELECT count(*) FROM relations WHERE actor=? AND target=? AND kind='close'", p.Owner, u)
	}
	if p.Visibility == "followers" || a.exists("SELECT count(*) FROM profiles WHERE user_id=? AND private=1", p.Owner) {
		return a.exists("SELECT count(*) FROM relations WHERE actor=? AND target=? AND kind='follow' AND state='accepted'", u, p.Owner)
	}
	return p.Visibility == "public"
}
func (a *App) decorate(p Post, user string) Post {
	p.Profile = a.getProfile(p.Owner)
	a.H.DB.QueryRow("SELECT count(*) FROM engagement WHERE post_id=? AND kind='like'", p.ID).Scan(&p.Likes)
	a.H.DB.QueryRow("SELECT count(*) FROM comments WHERE post_id=?", p.ID).Scan(&p.Comments)
	p.Liked = a.exists("SELECT count(*) FROM engagement WHERE user_id=? AND post_id=? AND kind='like'", user, p.ID)
	p.Saved = a.exists("SELECT count(*) FROM engagement WHERE user_id=? AND post_id=? AND kind='save'", user, p.ID)
	return p
}
func (a *App) getPost(w http.ResponseWriter, r *http.Request) {
	p, e := a.post(r.PathValue("id"))
	if e != nil || !a.visible(uid(r), p) {
		http.NotFound(w, r)
		return
	}
	apphost.JSON(w, 200, a.decorate(p, uid(r)))
}
func (a *App) feed(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT id FROM posts WHERE (state='published' OR owner=?) AND created<? ORDER BY created DESC,id DESC LIMIT 500", uid(r), parseBefore(r.URL.Query().Get("before")))
	if e != nil {
		a.fail(w, e)
		return
	}
	ids := []string{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	out := []Post{}
	mode := r.URL.Query().Get("mode")
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	for _, id := range ids {
		p, e := a.post(id)
		if e != nil || !a.visible(uid(r), p) {
			continue
		}
		if mode == "studio" {
			if p.Owner != uid(r) {
				continue
			}
		} else if p.State != "published" || p.Kind == "story" && p.Expires <= a.now().Unix() {
			continue
		}
		if mode == "stories" && p.Kind != "story" || mode == "reels" && p.Kind != "reel" {
			continue
		}
		if mode != "stories" && mode != "studio" && mode != "profile" && p.Kind == "story" {
			continue
		}
		if mode == "following" && p.Owner != uid(r) && !a.exists("SELECT count(*) FROM relations WHERE actor=? AND target=? AND kind='follow' AND state='accepted'", uid(r), p.Owner) {
			continue
		}
		if a.exists("SELECT count(*) FROM relations WHERE actor=? AND target=? AND kind='mute'", uid(r), p.Owner) && mode == "home" {
			continue
		}
		if owner := r.URL.Query().Get("owner"); owner != "" && p.Owner != owner {
			continue
		}
		if mode == "saved" && !a.exists("SELECT count(*) FROM engagement WHERE user_id=? AND post_id=? AND kind='save'", uid(r), p.ID) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(p.Caption+" "+encode(a.getProfile(p.Owner))), q) {
			continue
		}
		out = append(out, a.decorate(p, uid(r)))
		if len(out) == 30 {
			break
		}
	}
	next := int64(0)
	if len(out) > 0 {
		next = out[len(out)-1].Created
	}
	apphost.JSON(w, 200, map[string]any{"items": out, "next": next})
}
func parseBefore(s string) int64 {
	if s == "" {
		return 1 << 62
	}
	var n int64
	json.Unmarshal([]byte(s), &n)
	if n < 1 {
		return 1 << 62
	}
	return n
}
func (a *App) publish(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Kind, Caption, Alt, Location, Visibility, RequestKey string
		Media                                                []string
		Extras                                               json.RawMessage
		PublishAt                                            int64
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	if in.Kind != "post" && in.Kind != "story" && in.Kind != "reel" || len(in.Caption) > 2200 || len(in.Alt) > 1000 || len(in.Location) > 200 || len(in.Media) < 1 || len(in.Media) > 10 || !validID(in.RequestKey) {
		apphost.Problem(w, 400, "Choose 1–10 media items and a caption under 2,200 characters")
		return
	}
	if in.Visibility != "public" && in.Visibility != "followers" && in.Visibility != "close-friends" && in.Visibility != "private" {
		apphost.Problem(w, 400, "Choose an audience")
		return
	}
	if len(in.Extras) == 0 {
		in.Extras = json.RawMessage(`{}`)
	}
	if len(in.Extras) > 16384 || !json.Valid(in.Extras) {
		apphost.Problem(w, 400, "Invalid story elements")
		return
	}
	var elements struct {
		Poll *struct {
			Question string   `json:"question"`
			Choices  []string `json:"choices"`
		} `json:"poll"`
		Link     string `json:"link"`
		Captions string `json:"captions"`
	}
	if json.Unmarshal(in.Extras, &elements) != nil || (elements.Link != "" && !safeLink(elements.Link)) || len(elements.Captions) > 10000 {
		apphost.Problem(w, 400, "Invalid Story link or captions")
		return
	}
	if elements.Poll != nil {
		if len(elements.Poll.Question) > 200 || len(elements.Poll.Choices) < 2 || len(elements.Poll.Choices) > 4 {
			apphost.Problem(w, 400, "Polls need a question and 2–4 choices")
			return
		}
		for _, c := range elements.Poll.Choices {
			if len(c) > 100 {
				apphost.Problem(w, 400, "Poll choice is too long")
				return
			}
		}
	}
	in.Extras = json.RawMessage(encode(elements))
	a.publishMu.Lock()
	defer a.publishMu.Unlock()
	var prior string
	if a.H.DB.QueryRow("SELECT id FROM posts WHERE owner=? AND request_key=?", uid(r), in.RequestKey).Scan(&prior) == nil {
		p, _ := a.post(prior)
		apphost.JSON(w, 200, a.decorate(p, uid(r)))
		return
	}
	if e := a.storageCheck(r.Context()); e != nil {
		apphost.Problem(w, 503, "My Photos is unavailable; your media stays a draft")
		return
	}
	unique := map[string]bool{}
	for _, id := range in.Media {
		m, e := a.media(id)
		if e != nil || m.Owner != uid(r) || !a.readyMedia(m) || unique[id] {
			apphost.Problem(w, 409, "Finish uploading and rendering every selected item first")
			return
		}
		if in.Kind == "reel" && !strings.HasPrefix(m.MIME, "video/") {
			apphost.Problem(w, 400, "Reels require rendered video")
			return
		}
		unique[id] = true
	}
	now := a.now().Unix()
	at := in.PublishAt
	if at == 0 {
		at = now
	}
	if at < now-60 || at > now+366*86400 {
		apphost.Problem(w, 400, "Schedule within the next year")
		return
	}
	state := "published"
	if at > now {
		state = "scheduled"
	}
	expires := int64(0)
	if in.Kind == "story" {
		expires = at + 86400
	}
	id := apphost.ID()
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.fail(w, e)
		return
	}
	defer tx.Rollback()
	_, e = tx.Exec("INSERT INTO posts VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", id, uid(r), in.Kind, in.Caption, in.Alt, in.Location, in.Visibility, state, now, at, expires, encode(in.Media), string(in.Extras), in.RequestKey)
	if e != nil {
		a.fail(w, e)
		return
	}
	// The publication and durable event are one transaction.
	if state == "published" {
		_, e = tx.Exec("INSERT OR IGNORE INTO events(user_id,actor,category,object,created) SELECT actor,?,'post',?,? FROM relations WHERE target=? AND kind='follow' AND state='accepted'", uid(r), id, now, uid(r))
	}
	if e != nil {
		a.fail(w, e)
		return
	}
	if e = tx.Commit(); e != nil {
		a.fail(w, e)
		return
	}
	p, _ := a.post(id)
	apphost.JSON(w, 201, a.decorate(p, uid(r)))
}
func (a *App) editPost(w http.ResponseWriter, r *http.Request) {
	p, e := a.post(r.PathValue("id"))
	if e != nil || p.Owner != uid(r) {
		http.NotFound(w, r)
		return
	}
	var in struct{ Caption, Alt, Visibility string }
	if !apphost.Decode(w, r, &in) {
		return
	}
	if len(in.Caption) > 2200 || len(in.Alt) > 1000 || !strings.Contains("|public|followers|close-friends|private|", "|"+in.Visibility+"|") {
		apphost.Problem(w, 400, "Check caption and audience")
		return
	}
	_, e = a.H.DB.Exec("UPDATE posts SET caption=?,alt=?,visibility=? WHERE id=? AND owner=?", in.Caption, in.Alt, in.Visibility, p.ID, uid(r))
	if e != nil {
		a.fail(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) postAction(w http.ResponseWriter, r *http.Request) {
	p, e := a.post(r.PathValue("id"))
	if e != nil || !a.visible(uid(r), p) {
		http.NotFound(w, r)
		return
	}
	action := r.PathValue("action")
	switch action {
	case "like", "save":
		var in struct {
			Enabled    bool
			Collection string
		}
		if !apphost.Decode(w, r, &in) {
			return
		}
		if len(in.Collection) > 80 {
			apphost.Problem(w, 400, "Collection name is too long")
			return
		}
		if in.Enabled {
			_, e = a.H.DB.Exec("INSERT OR IGNORE INTO engagement VALUES(?,?,?,?)", uid(r), p.ID, action, in.Collection)
			if action == "like" {
				a.notify(p.Owner, uid(r), "like", p.ID)
			}
		} else {
			_, e = a.H.DB.Exec("DELETE FROM engagement WHERE user_id=? AND post_id=? AND kind=?", uid(r), p.ID, action)
		}
	case "archive", "restore", "delete":
		if p.Owner != uid(r) {
			http.NotFound(w, r)
			return
		}
		state := map[string]string{"archive": "archived", "restore": "published", "delete": "deleted"}[action]
		if action == "restore" && p.Kind == "story" && p.Expires <= a.now().Unix() {
			state = "archived"
		}
		_, e = a.H.DB.Exec("UPDATE posts SET state=? WHERE id=?", state, p.ID)
	case "view":
		if p.Kind != "story" {
			http.NotFound(w, r)
			return
		}
		_, e = a.H.DB.Exec("INSERT OR IGNORE INTO story_views VALUES(?,?,?)", p.ID, uid(r), a.now().Unix())
	case "vote":
		if p.Kind != "story" || p.Expires <= a.now().Unix() {
			http.NotFound(w, r)
			return
		}
		var in struct{ Value string }
		if !apphost.Decode(w, r, &in) {
			return
		}
		if len(in.Value) > 200 {
			apphost.Problem(w, 400, "Response too long")
			return
		}
		_, e = a.H.DB.Exec("INSERT INTO story_votes VALUES(?,?,?) ON CONFLICT(post_id,user_id) DO UPDATE SET value=excluded.value", p.ID, uid(r), in.Value)
		a.notify(p.Owner, uid(r), "story", p.ID)
	default:
		http.NotFound(w, r)
		return
	}
	if e != nil {
		a.fail(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}

type Comment struct {
	ID      string `json:"id"`
	Owner   string `json:"owner"`
	Parent  string `json:"parent"`
	Body    string `json:"body"`
	Created int64  `json:"created"`
	Profile any    `json:"profile"`
}

func (a *App) comments(w http.ResponseWriter, r *http.Request) {
	p, e := a.post(r.PathValue("id"))
	if e != nil || !a.visible(uid(r), p) {
		http.NotFound(w, r)
		return
	}
	rows, e := a.H.DB.Query("SELECT id,owner,parent,body,created FROM comments WHERE post_id=? ORDER BY created LIMIT 200", p.ID)
	if e != nil {
		a.fail(w, e)
		return
	}
	out := []Comment{}
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID, &c.Owner, &c.Parent, &c.Body, &c.Created)
		out = append(out, c)
	}
	rows.Close()
	for i := range out {
		if a.blocked(uid(r), out[i].Owner) {
			out[i].Body = "[Unavailable]"
		}
		out[i].Profile = a.getProfile(out[i].Owner)
	}
	apphost.JSON(w, 200, out)
}
func (a *App) addComment(w http.ResponseWriter, r *http.Request) {
	p, e := a.post(r.PathValue("id"))
	if e != nil || !a.visible(uid(r), p) {
		http.NotFound(w, r)
		return
	}
	var in struct{ Body, Parent string }
	if !apphost.Decode(w, r, &in) {
		return
	}
	in.Body = strings.TrimSpace(in.Body)
	if in.Body == "" || len(in.Body) > 2000 {
		apphost.Problem(w, 400, "Write a comment under 2,000 characters")
		return
	}
	if in.Parent != "" && !a.exists("SELECT count(*) FROM comments WHERE id=? AND post_id=?", in.Parent, p.ID) {
		apphost.Problem(w, 400, "Reply target not found")
		return
	}
	if a.exists("SELECT count(*) FROM relations WHERE actor=? AND target=? AND kind='restrict'", p.Owner, uid(r)) {
		apphost.Problem(w, 403, "Comments are restricted for this account")
		return
	}
	_, e = a.H.DB.Exec("INSERT INTO comments VALUES(?,?,?,?,?,?)", apphost.ID(), p.ID, uid(r), in.Parent, in.Body, a.now().Unix())
	if e != nil {
		a.fail(w, e)
		return
	}
	a.notify(p.Owner, uid(r), "comment", p.ID)
	apphost.JSON(w, 201, map[string]bool{"ok": true})
}

var _ = sql.ErrNoRows
var _ = time.Second
