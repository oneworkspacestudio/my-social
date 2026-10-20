package social

import (
	"context"
	"encoding/json"
	"fmt"
	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Event struct {
	ID       int64  `json:"id"`
	Actor    string `json:"actor"`
	Category string `json:"category"`
	Object   string `json:"object"`
	Created  int64  `json:"created"`
	Seen     bool   `json:"seen"`
}

func (a *App) notify(user, actor, category, object string) {
	if user == actor && category != "test" {
		return
	}
	a.H.DB.Exec("INSERT OR IGNORE INTO events(user_id,actor,category,object,created) VALUES(?,?,?,?,?)", user, actor, category, object, a.now().Unix())
}
func (a *App) userEvents(user string, after int64) []Event {
	rows, e := a.H.DB.Query("SELECT id,actor,category,object,created,seen FROM events WHERE user_id=? AND id>? ORDER BY id DESC LIMIT 100", user, after)
	if e != nil {
		return []Event{}
	}
	out := []Event{}
	for rows.Next() {
		var x Event
		rows.Scan(&x.ID, &x.Actor, &x.Category, &x.Object, &x.Created, &x.Seen)
		out = append(out, x)
	}
	rows.Close()
	filtered := []Event{}
	for _, x := range out {
		if x.Category != "test" && a.blocked(user, x.Actor) {
			continue
		}
		if x.Category == "post" || x.Category == "like" || x.Category == "comment" || x.Category == "story" {
			p, e := a.post(x.Object)
			if e != nil || !a.visible(user, p) {
				continue
			}
		}
		filtered = append(filtered, x)
	}
	return filtered
}
func (a *App) notifications(w http.ResponseWriter, r *http.Request) {
	apphost.JSON(w, 200, a.userEvents(uid(r), 0))
}
func (a *App) readNotifications(w http.ResponseWriter, r *http.Request) {
	a.H.DB.Exec("UPDATE events SET seen=1 WHERE user_id=?", uid(r))
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) events(w http.ResponseWriter, r *http.Request) {
	// A short SSE response revalidates the HttpOnly session on every reconnect.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	after := parseBefore(r.URL.Query().Get("after"))
	if after == 1<<62 {
		after = 0
	}
	fmt.Fprintf(w, "event: activity\ndata: %s\n\nretry: 5000\n\n", encode(a.userEvents(uid(r), after)))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
func (a *App) pref(user string) map[string]any {
	var raw string
	out := map[string]any{"push": false, "quietStart": "", "quietEnd": ""}
	if a.H.DB.QueryRow("SELECT data FROM preferences WHERE user_id=?", user).Scan(&raw) == nil {
		json.Unmarshal([]byte(raw), &out)
	}
	return out
}
func (a *App) preferences(w http.ResponseWriter, r *http.Request) {
	apphost.JSON(w, 200, a.pref(uid(r)))
}
func (a *App) setPreferences(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Push       bool            `json:"push"`
		QuietStart string          `json:"quietStart"`
		QuietEnd   string          `json:"quietEnd"`
		Categories map[string]bool `json:"categories"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	for _, v := range []string{in.QuietStart, in.QuietEnd} {
		if v != "" {
			if _, e := time.Parse("15:04", v); e != nil {
				apphost.Problem(w, 400, "Use HH:MM for quiet hours (server time)")
				return
			}
		}
	}
	if len(in.Categories) > 20 {
		apphost.Problem(w, 400, "Too many categories")
		return
	}
	_, e := a.H.DB.Exec("INSERT INTO preferences VALUES(?,?) ON CONFLICT(user_id) DO UPDATE SET data=excluded.data", uid(r), encode(in))
	if e != nil {
		a.fail(w, e)
		return
	}
	if !in.Push {
		a.H.DB.Exec("DELETE FROM push_subscriptions WHERE owner=?", uid(r))
	}
	apphost.JSON(w, 200, in)
}
func safeLink(s string) bool {
	u, e := url.Parse(s)
	return e == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && len(s) < 2000 && !strings.ContainsAny(s, "\r\n")
}
func validPush(endpoint string) bool {
	u, e := url.Parse(endpoint)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return false
	}
	h := u.Hostname()
	return h == "fcm.googleapis.com" || h == "updates.push.services.mozilla.com" || h == "web.push.apple.com" || strings.HasSuffix(h, ".notify.windows.com")
}
func (a *App) subscribe(w http.ResponseWriter, r *http.Request) {
	var in webpush.Subscription
	if !apphost.Decode(w, r, &in) {
		return
	}
	if os.Getenv("MYSOCIAL_VAPID_PRIVATE") == "" || !validPush(in.Endpoint) || len(in.Keys.Auth) < 10 || len(in.Keys.P256dh) < 40 {
		apphost.Problem(w, 400, "Configure VAPID and use a supported browser push service")
		return
	}
	id := apphost.Hash(uid(r) + in.Endpoint)
	a.H.DB.Exec("INSERT INTO push_subscriptions VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET data=excluded.data", id, uid(r), encode(in), a.now().Unix())
	apphost.JSON(w, 201, map[string]bool{"ok": true})
}
func (a *App) unsubscribe(w http.ResponseWriter, r *http.Request) {
	a.H.DB.Exec("DELETE FROM push_subscriptions WHERE owner=?", uid(r))
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) testPush(w http.ResponseWriter, r *http.Request) {
	a.notify(uid(r), uid(r), "test", apphost.ID())
	apphost.JSON(w, 202, map[string]bool{"queued": true})
}
func (a *App) deliverPush(ctx context.Context) {
	if os.Getenv("MYSOCIAL_VAPID_PRIVATE") == "" {
		return
	}
	rows, e := a.H.DB.Query("SELECT id,user_id,category,object FROM events WHERE delivered=0 AND attempts<5 AND next_attempt<=? ORDER BY id LIMIT 10", a.now().Unix())
	if e != nil {
		return
	}
	type item struct {
		id                     int64
		user, category, object string
	}
	items := []item{}
	for rows.Next() {
		var x item
		rows.Scan(&x.id, &x.user, &x.category, &x.object)
		items = append(items, x)
	}
	rows.Close()
	for _, x := range items {
		p := a.pref(x.user)
		if p["push"] != true {
			a.H.DB.Exec("UPDATE events SET delivered=1 WHERE id=?", x.id)
			continue
		}
		now := a.now().Format("15:04")
		start, _ := p["quietStart"].(string)
		end, _ := p["quietEnd"].(string)
		if start != "" && end != "" && ((start < end && now >= start && now < end) || (start > end && (now >= start || now < end))) {
			continue
		}
		if cs, ok := p["categories"].(map[string]any); ok && cs[x.category] == false {
			a.H.DB.Exec("UPDATE events SET delivered=1 WHERE id=?", x.id)
			continue
		}
		allowed := false
		for _, ev := range a.userEvents(x.user, 0) {
			if ev.ID == x.id {
				allowed = true
				break
			}
		}
		if !allowed {
			a.H.DB.Exec("UPDATE events SET delivered=1 WHERE id=?", x.id)
			continue
		}
		rs, e := a.H.DB.Query("SELECT id,data FROM push_subscriptions WHERE owner=?", x.user)
		if e != nil {
			continue
		}
		subs := map[string]webpush.Subscription{}
		for rs.Next() {
			var id, raw string
			rs.Scan(&id, &raw)
			var s webpush.Subscription
			if json.Unmarshal([]byte(raw), &s) == nil && validPush(s.Endpoint) {
				subs[id] = s
			}
		}
		rs.Close()
		failed := false
		for id, s := range subs {
			sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			resp, e := webpush.SendNotificationWithContext(sendCtx, []byte(`{"title":"My Social","body":"You have new activity. Open My Social to view it.","path":"#notifications"}`), &s, &webpush.Options{Subscriber: os.Getenv("MYSOCIAL_VAPID_CONTACT"), VAPIDPublicKey: os.Getenv("MYSOCIAL_VAPID_PUBLIC"), VAPIDPrivateKey: os.Getenv("MYSOCIAL_VAPID_PRIVATE"), TTL: 60, HTTPClient: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}})
			cancel()
			if e != nil {
				failed = true
				continue
			}
			io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
			if resp.StatusCode == 404 || resp.StatusCode == 410 {
				a.H.DB.Exec("DELETE FROM push_subscriptions WHERE id=?", id)
			} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				failed = true
			}
		}
		if failed {
			a.H.DB.Exec("UPDATE events SET attempts=attempts+1,next_attempt=? WHERE id=?", a.now().Add(time.Minute).Unix(), x.id)
		} else {
			a.H.DB.Exec("UPDATE events SET delivered=1 WHERE id=?", x.id)
		}
	}
}
