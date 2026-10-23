package social

import (
	"encoding/json"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"regexp"
	"strings"
)

type Profile struct {
	Avatar         string          `json:"avatar,omitempty"`
	ID             string          `json:"id"`
	Handle         string          `json:"handle"`
	Name           string          `json:"name"`
	Bio            string          `json:"bio"`
	Private        bool            `json:"private"`
	Followers      int             `json:"followers"`
	Following      int             `json:"following"`
	Posts          int             `json:"posts"`
	Links          json.RawMessage `json:"links"`
	FollowingState string          `json:"followingState,omitempty"`
}

func (a *App) ensureProfile(u apphost.User) {
	a.H.DB.Exec("INSERT OR IGNORE INTO profiles(user_id,handle) VALUES(?,?)", u.ID, "member_"+u.ID[:12])
}
func (a *App) getProfile(id string) Profile {
	p := Profile{ID: id, Links: json.RawMessage(`[]`)}
	var links string
	a.H.DB.QueryRow("SELECT p.handle,u.name,p.bio,p.private,p.links FROM profiles p JOIN host_users u ON u.id=p.user_id WHERE p.user_id=?", id).Scan(&p.Handle, &p.Name, &p.Bio, &p.Private, &links)
	if json.Valid([]byte(links)) {
		p.Links = json.RawMessage(links)
	}
	a.H.DB.QueryRow("SELECT count(*) FROM relations WHERE target=? AND kind='follow' AND state='accepted'", id).Scan(&p.Followers)
	a.H.DB.QueryRow("SELECT count(*) FROM relations WHERE actor=? AND kind='follow' AND state='accepted'", id).Scan(&p.Following)
	a.H.DB.QueryRow("SELECT count(*) FROM posts WHERE owner=? AND state='published' AND kind<>'story'", id).Scan(&p.Posts)
	a.H.DB.QueryRow("SELECT media_id FROM profile_photos WHERE user_id=?", id).Scan(&p.Avatar)
	return p
}
func (a *App) profiles(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	if len(q) > 100 {
		apphost.Problem(w, 400, "Search is too long")
		return
	}
	rows, e := a.H.DB.Query("SELECT user_id FROM profiles WHERE handle LIKE ? OR bio LIKE ? ORDER BY handle LIMIT 100", "%"+q+"%", "%"+q+"%")
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
	out := []Profile{}
	for _, x := range ids {
		if a.blocked(uid(r), x) {
			continue
		}
		p := a.getProfile(x)
		a.H.DB.QueryRow("SELECT state FROM relations WHERE actor=? AND target=? AND kind='follow'", uid(r), x).Scan(&p.FollowingState)
		out = append(out, p)
	}
	apphost.JSON(w, 200, out)
}
func (a *App) profile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !a.exists("SELECT count(*) FROM profiles WHERE user_id=?", id) || a.blocked(uid(r), id) {
		http.NotFound(w, r)
		return
	}
	p := a.getProfile(id)
	a.H.DB.QueryRow("SELECT state FROM relations WHERE actor=? AND target=? AND kind='follow'", uid(r), id).Scan(&p.FollowingState)
	apphost.JSON(w, 200, p)
}
func (a *App) editProfile(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Handle, Bio string
		Private     bool
		Links       []string
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	in.Handle = strings.ToLower(in.Handle)
	if !regexp.MustCompile(`^[a-z0-9_\.]{3,30}$`).MatchString(in.Handle) || len(in.Bio) > 500 || len(in.Links) > 5 {
		apphost.Problem(w, 400, "Use a 3–30 character handle and bio under 500 characters")
		return
	}
	for _, l := range in.Links {
		if !safeLink(l) {
			apphost.Problem(w, 400, "Profile links must be HTTPS without credentials")
			return
		}
	}
	_, e := a.H.DB.Exec("UPDATE profiles SET handle=?,bio=?,private=?,links=? WHERE user_id=?", in.Handle, in.Bio, in.Private, encode(in.Links), uid(r))
	if e != nil {
		apphost.Problem(w, 409, "That handle is already in use")
		return
	}
	apphost.JSON(w, 200, a.getProfile(uid(r)))
}
func (a *App) relation(w http.ResponseWriter, r *http.Request) {
	target := r.PathValue("id")
	if target == uid(r) || !a.exists("SELECT count(*) FROM profiles WHERE user_id=?", target) {
		http.NotFound(w, r)
		return
	}
	var in struct {
		Kind    string
		Enabled bool
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	if in.Kind == "approve" || in.Kind == "reject" {
		if in.Kind == "approve" {
			a.H.DB.Exec("UPDATE relations SET state='accepted' WHERE actor=? AND target=? AND kind='follow' AND state='pending'", target, uid(r))
			a.notify(target, uid(r), "follow-approved", uid(r))
		} else {
			a.H.DB.Exec("DELETE FROM relations WHERE actor=? AND target=? AND kind='follow' AND state='pending'", target, uid(r))
		}
		apphost.JSON(w, 200, map[string]bool{"ok": true})
		return
	}
	if in.Kind != "follow" && in.Kind != "block" && in.Kind != "mute" && in.Kind != "restrict" && in.Kind != "close" {
		apphost.Problem(w, 400, "Unsupported relationship")
		return
	}
	if in.Kind == "follow" && a.blocked(uid(r), target) {
		http.NotFound(w, r)
		return
	}
	state := "accepted"
	if in.Kind == "follow" && a.getProfile(target).Private {
		state = "pending"
	}
	var e error
	if in.Enabled {
		_, e = a.H.DB.Exec("INSERT INTO relations VALUES(?,?,?,?) ON CONFLICT(actor,target,kind) DO UPDATE SET state=excluded.state", uid(r), target, in.Kind, state)
		if in.Kind == "follow" {
			a.notify(target, uid(r), "follow", uid(r))
		}
		if in.Kind == "block" {
			a.H.DB.Exec("DELETE FROM relations WHERE kind='follow' AND ((actor=? AND target=?) OR (actor=? AND target=?))", uid(r), target, target, uid(r))
		}
	} else {
		_, e = a.H.DB.Exec("DELETE FROM relations WHERE actor=? AND target=? AND kind=?", uid(r), target, in.Kind)
	}
	if e != nil {
		a.fail(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]string{"state": state})
}
func (a *App) relations(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT actor,target,kind,state FROM relations WHERE actor=? OR (target=? AND kind='follow' AND state='pending')", uid(r), uid(r))
	if e != nil {
		a.fail(w, e)
		return
	}
	defer rows.Close()
	out := []map[string]string{}
	for rows.Next() {
		var actor, target, kind, state string
		rows.Scan(&actor, &target, &kind, &state)
		out = append(out, map[string]string{"actor": actor, "target": target, "kind": kind, "state": state})
	}
	apphost.JSON(w, 200, out)
}
func (a *App) drafts(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT id,document,updated FROM drafts WHERE owner=? ORDER BY updated DESC LIMIT 50", uid(r))
	if e != nil {
		a.fail(w, e)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, doc string
		var at int64
		rows.Scan(&id, &doc, &at)
		out = append(out, map[string]any{"id": id, "document": json.RawMessage(doc), "updated": at})
	}
	apphost.JSON(w, 200, out)
}
func (a *App) saveDraft(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var doc json.RawMessage
	if !validID(id) || !apphost.Decode(w, r, &doc) {
		return
	}
	if len(doc) > 700000 {
		apphost.Problem(w, 413, "Editor project is too large")
		return
	}
	if a.exists("SELECT count(*) FROM drafts WHERE id=? AND owner<>?", id, uid(r)) {
		http.NotFound(w, r)
		return
	}
	if !a.exists("SELECT count(*) FROM drafts WHERE id=?", id) && a.exists("SELECT count(*) FROM drafts WHERE owner=? HAVING count(*)>=50", uid(r)) {
		apphost.Problem(w, 413, "Remove an old draft first")
		return
	}
	_, e := a.H.DB.Exec("INSERT INTO drafts VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET document=excluded.document,updated=excluded.updated WHERE owner=excluded.owner", id, uid(r), string(doc), a.now().Unix())
	if e != nil {
		a.fail(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) deleteDraft(w http.ResponseWriter, r *http.Request) {
	a.H.DB.Exec("DELETE FROM drafts WHERE id=? AND owner=?", r.PathValue("id"), uid(r))
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) highlights(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("owner")
	if owner == "" {
		owner = uid(r)
	}
	if a.blocked(uid(r), owner) {
		http.NotFound(w, r)
		return
	}
	rows, e := a.H.DB.Query("SELECT id,title,posts FROM highlights WHERE owner=?", owner)
	if e != nil {
		a.fail(w, e)
		return
	}
	out := []map[string]any{}
	for rows.Next() {
		var id, title, raw string
		rows.Scan(&id, &title, &raw)
		var ids []string
		json.Unmarshal([]byte(raw), &ids)
		out = append(out, map[string]any{"id": id, "title": title, "posts": ids})
	}
	rows.Close()
	for _, h := range out {
		visible := []string{}
		for _, id := range h["posts"].([]string) {
			p, e := a.post(id)
			if e == nil && (uid(r) == owner || a.audience(uid(r), p)) && p.State != "deleted" && p.State != "removed" {
				visible = append(visible, id)
			}
		}
		h["posts"] = visible
	}
	apphost.JSON(w, 200, out)
}
func (a *App) addHighlight(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title string
		Posts []string
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	if len(in.Title) > 60 || len(in.Posts) > 100 || len(in.Posts) == 0 {
		apphost.Problem(w, 400, "Name the Highlight and select up to 100 Stories")
		return
	}
	for _, id := range in.Posts {
		p, e := a.post(id)
		if e != nil || p.Owner != uid(r) || p.Kind != "story" || p.State == "deleted" || p.State == "removed" {
			http.NotFound(w, r)
			return
		}
	}
	id := apphost.ID()
	_, e := a.H.DB.Exec("INSERT INTO highlights VALUES(?,?,?,?)", id, uid(r), in.Title, encode(in.Posts))
	if e != nil {
		a.fail(w, e)
		return
	}
	apphost.JSON(w, 201, map[string]string{"id": id})
}
