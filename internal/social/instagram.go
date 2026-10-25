package social

// Instagram exports are data, never filesystem destinations. Only media named
// by the supported content manifests is read; no archive entry is extracted.
import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	_ "golang.org/x/image/webp"
)

const instagramZIPLimit = 512 << 20
const instagramMediaLimit = 100 << 20
const instagramVideoSeconds = 10 * 60

type instagramMedia struct {
	URI     string `json:"uri"`
	Created int64  `json:"creation_timestamp"`
	Title   string `json:"title"`
}
type instagramEntry struct {
	Title   string           `json:"title"`
	Created int64            `json:"creation_timestamp"`
	Media   []instagramMedia `json:"media"`
}
type instagramItem struct {
	Key        string `json:"key"`
	Kind       string `json:"kind"`
	Caption    string `json:"caption"`
	Created    int64  `json:"created"`
	MediaCount int    `json:"mediaCount"`
	Archived   bool   `json:"archived"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
	files      []*zip.File
	types      []string
}
type instagramOptions struct {
	Stream     bool     `json:"stream"`
	Digest     string   `json:"digest"`
	Keys       []string `json:"keys"`
	Visibility string   `json:"visibility"`
}

func instagramPath(s string) bool {
	return s != "" && !strings.ContainsAny(s, "\\:\x00") && !strings.HasPrefix(s, "/") && path.Clean(s) == s && s != ".." && !strings.HasPrefix(s, "../")
}

// Meta's JSON has historically encoded UTF-8 bytes as Latin-1 characters.
// Repair only when that reversal yields valid UTF-8; real Unicode stays intact.
func instagramText(s string) string {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 255 {
			return s
		}
		b = append(b, byte(r))
	}
	if utf8.Valid(b) {
		return string(b)
	}
	return s
}

func readInstagramFile(f *zip.File, limit int64) ([]byte, error) {
	if f.UncompressedSize64 > uint64(limit) {
		return nil, errors.New("file exceeds the import limit")
	}
	r, e := f.Open()
	if e != nil {
		return nil, e
	}
	defer r.Close()
	b, e := io.ReadAll(io.LimitReader(r, limit+1))
	if e != nil || int64(len(b)) > limit {
		return nil, errors.New("file is damaged or too large")
	}
	return b, nil
}

var instagramPosts = regexp.MustCompile(`^posts(_[0-9]+)?\.json$`)

type instagramFileInfo struct{ hash, mime, reason string }

func inspectInstagramFile(f *zip.File, videoEnabled bool) instagramFileInfo {
	b, e := readInstagramFile(f, instagramMediaLimit)
	if e != nil {
		return instagramFileInfo{reason: "Media is damaged or exceeds 100 MiB"}
	}
	info := instagramFileInfo{mime: http.DetectContentType(b)}
	switch info.mime {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		cfg, _, e := image.DecodeConfig(bytes.NewReader(b))
		if e != nil || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 24000000 {
			info.reason = "Image is damaged or exceeds 24 megapixels"
		}
	case "video/mp4", "video/webm":
		if !videoEnabled {
			info.reason = "Video needs configured FFmpeg and FFprobe"
		}
	default:
		info.reason = "Unsupported media; use JPEG, PNG, GIF, WebP, MP4 or WebM"
	}
	sum := sha256.Sum256(b)
	info.hash = hex.EncodeToString(sum[:])
	return info
}

// Bound the central directory before archive/zip allocates its file records.
// These imports do not need ZIP64 (512 MiB compressed / 2 GiB expanded).
func instagramDirectory(r io.ReaderAt, size int64) error {
	if size < 22 {
		return errors.New("This is not a readable ZIP archive")
	}
	b := make([]byte, min(size, 65557))
	if _, e := r.ReadAt(b, size-int64(len(b))); e != nil {
		return e
	}
	for i := len(b) - 22; i >= 0; i-- {
		if binary.LittleEndian.Uint32(b[i:]) != 0x06054b50 || i+22+int(binary.LittleEndian.Uint16(b[i+20:])) != len(b) {
			continue
		}
		count := binary.LittleEndian.Uint16(b[i+10:])
		directorySize := uint64(binary.LittleEndian.Uint32(b[i+12:]))
		if directorySize > 16<<20 || directorySize > uint64(size) {
			return errors.New("ZIP directory exceeds the import limit")
		}
		if count > 20000 || binary.LittleEndian.Uint16(b[i+4:]) != 0 || binary.LittleEndian.Uint16(b[i+6:]) != 0 || count != binary.LittleEndian.Uint16(b[i+8:]) {
			return errors.New("Use a single ZIP with at most 20,000 files; ZIP64 and multi-volume archives are unsupported")
		}
		return nil
	}
	return errors.New("This is not a readable ZIP archive")
}

func instagramManifest(name string) (prefix, kind, wrapper string, archived, ok bool) {
	for _, folder := range []string{"your_instagram_activity/content/", "your_instagram_activity/media/", "content/"} {
		i := strings.Index(name, folder)
		if i < 0 || i > 0 && name[i-1] != '/' {
			continue
		}
		base := strings.TrimPrefix(name[i:], folder)
		if strings.Contains(base, "/") {
			continue
		}
		switch {
		case instagramPosts.MatchString(base):
			return name[:i], "post", "", false, true
		case base == "reels.json":
			return name[:i], "reel", "ig_reels_media", false, true
		case base == "stories.json":
			return name[:i], "story", "ig_stories", true, true
		case base == "archived_posts.json":
			return name[:i], "post", "ig_archived_posts", true, true
		}
	}
	return
}

func (a *App) instagramPlan(ctx context.Context, z *zip.Reader, owner string, progress ...func(string, int, int)) ([]instagramItem, error) {
	report := func(label string, done, total int) {
		for _, update := range progress {
			update(label, done, total)
		}
	}
	if len(z.File) > 20000 {
		return nil, errors.New("ZIP has too many files (maximum 20,000)")
	}
	files := map[string]*zip.File{}
	var expanded, metadata uint64
	manifests := []*zip.File{}
	for _, f := range z.File {
		name := strings.TrimSuffix(f.Name, "/")
		if !instagramPath(name) || f.Mode()&os.ModeSymlink != 0 || !f.Mode().IsRegular() && !f.FileInfo().IsDir() || f.Flags&1 != 0 {
			return nil, errors.New("ZIP contains unsafe paths, links or encrypted files")
		}
		if _, exists := files[name]; exists {
			return nil, errors.New("ZIP contains duplicate paths")
		}
		files[name] = f
		if f.UncompressedSize64 > 2<<30 || expanded > (2<<30)-f.UncompressedSize64 {
			return nil, errors.New("ZIP expands beyond the 2 GiB limit")
		}
		expanded += f.UncompressedSize64
		if _, _, _, _, ok := instagramManifest(name); ok {
			metadata += f.UncompressedSize64
			if metadata > 16<<20 {
				return nil, errors.New("Content metadata exceeds 16 MiB")
			}
			manifests = append(manifests, f)
		}
	}
	if len(manifests) == 0 {
		return nil, errors.New("No Instagram JSON content manifests found. Export in JSON format and include posts and media")
	}
	sort.Slice(manifests, func(i, j int) bool { return manifests[i].Name < manifests[j].Name })
	items := []instagramItem{}
	seen := map[string]bool{}
	inspected := map[*zip.File]instagramFileInfo{}
	for collection, f := range manifests {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		prefix, kind, wrapper, archived, _ := instagramManifest(f.Name)
		raw, e := readInstagramFile(f, 16<<20)
		if e != nil {
			return nil, errors.New("Content metadata is damaged")
		}
		var entries []instagramEntry
		if wrapper == "" || len(bytes.TrimSpace(raw)) > 0 && bytes.TrimSpace(raw)[0] == '[' {
			e = json.Unmarshal(raw, &entries)
		} else {
			var object map[string]json.RawMessage
			e = json.Unmarshal(raw, &object)
			if e == nil {
				value, ok := object[wrapper]
				if !ok {
					return nil, errors.New("Unrecognized Instagram content structure")
				}
				e = json.Unmarshal(value, &entries)
			}
		}
		if e != nil {
			return nil, errors.New("Invalid Instagram JSON. Request a fresh JSON export")
		}
		if len(items)+len(entries) > 1000 {
			return nil, errors.New("Import at most 1,000 posts per ZIP; export a smaller date range")
		}
		for index, entry := range entries {
			label := fmt.Sprintf("Checking posts · collection %d of %d", collection+1, len(manifests))
			report(label, index, len(entries))
			item := instagramItem{Kind: kind, Archived: archived, Caption: instagramText(entry.Title), Created: entry.Created, MediaCount: len(entry.Media), Status: "ready"}
			if len(entry.Media) > 0 {
				if item.Caption == "" {
					item.Caption = instagramText(entry.Media[0].Title)
				}
				if item.Created == 0 {
					item.Created = entry.Media[0].Created
				}
			}
			if item.Created < 1 || item.Created > a.now().Unix() {
				item.Reason = "Missing or invalid original date"
			}
			if len(item.Caption) > 2200 {
				item.Reason = "Caption exceeds My Social's 2,200 byte limit"
			}
			if item.MediaCount < 1 || item.MediaCount > 10 {
				item.Reason = "Posts need 1–10 media files"
			}
			hashes := []string{}
			used := map[string]bool{}
			for mediaIndex, m := range entry.Media {
				report(fmt.Sprintf("Checking post %d of %d · media %d of %d", index+1, len(entries), mediaIndex+1, len(entry.Media)), index, len(entries))
				if item.Reason != "" {
					break
				}
				if e := ctx.Err(); e != nil {
					return nil, e
				}
				if !instagramPath(m.URI) || !strings.HasPrefix(m.URI, "media/") || used[m.URI] {
					item.Reason = "Invalid or repeated media reference"
					break
				}
				used[m.URI] = true
				mf := files[prefix+m.URI]
				if mf == nil || mf.FileInfo().IsDir() {
					item.Reason = "Media missing from this ZIP; include the matching media export"
					break
				}
				info, ok := inspected[mf]
				if !ok {
					info = inspectInstagramFile(mf, a.FFmpeg != "" && a.FFprobe != "")
					inspected[mf] = info
				}
				item.Reason = info.reason
				hashes = append(hashes, info.hash)
				item.files = append(item.files, mf)
				item.types = append(item.types, info.mime)
			}
			if item.Kind == "reel" && (len(item.types) != 1 || !strings.HasPrefix(item.types[0], "video/")) {
				item.Reason = "A Reel needs one video"
			}
			key := sha256.Sum256([]byte(encode([]any{item.Kind, item.Caption, item.Created, hashes})))
			item.Key = hex.EncodeToString(key[:])
			if item.Reason != "" {
				item.Status = "unavailable"
				// Unavailable rows also need distinct, stable UI keys.
				key = sha256.Sum256([]byte(fmt.Sprintf("%s:%d", f.Name, index)))
				item.Key = hex.EncodeToString(key[:])
			} else if seen[item.Key] || a.exists("SELECT count(*) FROM posts WHERE owner=? AND request_key=?", owner, "instagram:"+item.Key) {
				item.Status, item.Reason = "duplicate", "Already imported or repeated in this ZIP"
			}
			seen[item.Key] = true
			items = append(items, item)
			report(label, index+1, len(entries))
		}
	}
	return items, nil
}

// Spool once per request inside the guarded adapter, not an OS temp/shared
// folder. Both preview and apply remove the ZIP on return. Nothing is extracted.
func (a *App) instagram(w http.ResponseWriter, r *http.Request) {
	if !a.importMu.TryLock() {
		apphost.Problem(w, 429, "Another import is running. Retry when it finishes")
		return
	}
	defer a.importMu.Unlock()
	if e := a.storageCheck(r.Context()); e != nil {
		apphost.Problem(w, 503, "Connect protected My Photos/My Social storage before importing")
		return
	}
	if e := os.MkdirAll(a.Root, 0700); e != nil {
		a.fail(w, e)
		return
	}
	root, e := os.OpenRoot(a.Root)
	if e != nil {
		a.fail(w, e)
		return
	}
	defer root.Close()
	if e = a.recoverInstagramMedia(root); e != nil {
		apphost.Problem(w, 503, "Could not recover interrupted import media; check protected storage and retry")
		return
	}
	if e = root.MkdirAll("generated/imports", 0700); e != nil {
		a.fail(w, e)
		return
	}
	// Recover temporary archives left by an interrupted process. Active requests
	// are serialized and bounded to 15 minutes, so old files cannot be in use.
	if dir, err := root.Open("generated/imports"); err == nil {
		entries, _ := dir.ReadDir(-1)
		dir.Close()
		for _, entry := range entries {
			if info, err := entry.Info(); err == nil && info.Mode().IsRegular() && !strings.HasPrefix(entry.Name(), "job-") && strings.HasSuffix(entry.Name(), ".zip") && time.Since(info.ModTime()) > time.Hour {
				root.Remove("generated/imports/" + entry.Name())
			}
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, instagramZIPLimit+(1<<20))
	mr, e := r.MultipartReader()
	if e != nil {
		apphost.Problem(w, 400, "Choose an Instagram JSON export ZIP")
		return
	}
	part, e := mr.NextPart()
	if e != nil || part.FormName() != "options" {
		apphost.Problem(w, 400, "Import options must precede the ZIP")
		return
	}
	b, e := io.ReadAll(io.LimitReader(part, 65537))
	part.Close()
	var options instagramOptions
	if e != nil || len(b) > 65536 || json.Unmarshal(b, &options) != nil {
		apphost.Problem(w, 400, "Invalid import options")
		return
	}
	apply := r.PathValue("action") == "apply"
	if apply && (len(options.Keys) < 1 || len(options.Keys) > 100 || len(options.Digest) != 64 || (options.Visibility != "private" && options.Visibility != "followers" && options.Visibility != "close-friends" && options.Visibility != "public")) {
		apphost.Problem(w, 400, "Choose 1–100 posts and an audience")
		return
	}
	part, e = mr.NextPart()
	if e != nil || part.FormName() != "archive" {
		apphost.Problem(w, 400, "Choose the Instagram ZIP")
		return
	}
	rel := "generated/imports/" + apphost.ID() + ".zip"
	f, e := root.OpenFile(rel, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if e != nil {
		a.fail(w, e)
		return
	}
	defer func() { f.Close(); root.Remove(rel) }()
	h := sha256.New()
	size, e := io.Copy(io.MultiWriter(f, h), io.LimitReader(part, instagramZIPLimit+1))
	part.Close()
	if e != nil || size > instagramZIPLimit {
		apphost.Problem(w, 413, "ZIP must be at most 512 MiB; export a smaller date range")
		return
	}
	if _, e = mr.NextPart(); e != io.EOF {
		apphost.Problem(w, 400, "Upload exactly one ZIP")
		return
	}
	digest := hex.EncodeToString(h.Sum(nil))
	if apply && digest != options.Digest {
		apphost.Problem(w, 409, "The ZIP changed. Preview it again before importing")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Minute)
	defer cancel()
	if options.Stream {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Cache-Control", "no-store, no-transform")
		w.Header().Set("X-Accel-Buffering", "no")
	}
	send := func(value any) {
		if e := json.NewEncoder(w).Encode(value); e != nil {
			cancel()
			return
		}
		_ = http.NewResponseController(w).Flush()
	}
	update := func(phase, label string, done, total int) {
		if options.Stream {
			send(map[string]any{"type": "progress", "phase": phase, "label": label, "done": done, "total": total})
		}
	}
	problem := func(status int, message string) {
		if options.Stream {
			send(map[string]any{"type": "error", "status": status, "error": message})
		} else {
			apphost.Problem(w, status, message)
		}
	}
	respond := func(result any) {
		if options.Stream {
			send(map[string]any{"type": "result", "result": result})
		} else {
			apphost.JSON(w, 200, result)
		}
	}
	update("checking", "Opening and checking your ZIP", 0, 0)
	if e = instagramDirectory(f, size); e != nil {
		problem(400, e.Error())
		return
	}
	z, e := zip.NewReader(f, size)
	if e != nil {
		problem(400, "This is not a readable ZIP archive")
		return
	}
	items, e := a.instagramPlan(ctx, z, uid(r), func(label string, done, total int) { update("checking", label, done, total) })
	if e != nil {
		problem(400, e.Error())
		return
	}
	if !apply {
		respond(map[string]any{"digest": digest, "items": items})
		return
	}
	byKey := map[string]instagramItem{}
	for _, item := range items {
		if _, ok := byKey[item.Key]; !ok {
			byKey[item.Key] = item
		}
	}
	selected := map[string]bool{}
	for _, key := range options.Keys {
		item, ok := byKey[key]
		if !ok || selected[key] || item.Status == "unavailable" {
			problem(400, "Selection is invalid. Preview the ZIP again")
			return
		}
		selected[key] = true
	}
	results := []map[string]string{}
	imported, duplicates, failed := 0, 0, 0
	for index, key := range options.Keys {
		if ctx.Err() != nil {
			problem(408, "Import stopped. Completed posts are kept; preview and retry this ZIP safely")
			return
		}
		item := byKey[key]
		update("importing", fmt.Sprintf("Importing post %d of %d", index+1, len(options.Keys)), index, len(options.Keys))
		id, duplicate, err := a.importInstagramPost(ctx, root, uid(r), options.Visibility, item, func(media, total int) {
			update("importing", fmt.Sprintf("Importing post %d of %d · media %d of %d", index+1, len(options.Keys), media, total), index, len(options.Keys))
		})
		status, message := "imported", "Imported"
		if err != nil {
			status, message = "failed", err.Error()
			failed++
		} else if duplicate {
			status, message = "duplicate", "Already imported"
			duplicates++
		} else {
			imported++
		}
		results = append(results, map[string]string{"key": key, "status": status, "message": message, "post": id})
		update("importing", "Importing posts", index+1, len(options.Keys))
	}
	respond(map[string]any{"imported": imported, "duplicates": duplicates, "failed": failed, "results": results})
}

func (a *App) importInstagramPost(ctx context.Context, root *os.Root, owner, visibility string, item instagramItem, progress ...func(int, int)) (id string, duplicate bool, err error) {
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	a.publishMu.Lock()
	defer a.publishMu.Unlock()
	var role string
	if a.H.DB.QueryRow("SELECT role FROM host_users WHERE id=?", owner).Scan(&role) != nil || role == "disabled" {
		return "", false, errors.New("This account is no longer active")
	}
	key := "instagram:" + item.Key
	if a.H.DB.QueryRow("SELECT id FROM posts WHERE owner=? AND request_key=?", owner, key).Scan(&id) == nil {
		return id, true, nil
	}
	if ctx.Err() != nil {
		return "", false, errors.New("Import timed out or was cancelled; retry this post")
	}
	if a.storageCheck(ctx) != nil {
		return "", false, errors.New("Protected storage unavailable; retry this post")
	}
	var used int64
	if e := a.H.DB.QueryRow("SELECT coalesce(sum(bytes),0) FROM media WHERE owner=?", owner).Scan(&used); e != nil {
		return "", false, errors.New("Could not check media quota")
	}
	for _, file := range item.files {
		used += int64(file.UncompressedSize64)
	}
	if used > 2<<30 {
		return "", false, errors.New("Your 2 GiB preview media quota is full")
	}
	media := []Media{}
	paths := []string{}
	committed := false
	defer func() {
		if !committed {
			for _, p := range paths {
				root.Remove(filepath.FromSlash(p))
			}
			for _, m := range media {
				a.H.DB.Exec("DELETE FROM upload_intents WHERE id=?", m.ID)
			}
		}
	}()
	for i, file := range item.files {
		for _, update := range progress {
			update(i+1, len(item.files))
		}
		b, e := readInstagramFile(file, instagramMediaLimit)
		if e != nil {
			return "", false, errors.New("Media is damaged; request a fresh export")
		}
		mid := apphost.ID()
		folder := map[string]string{"post": "posts", "reel": "reels", "story": "stories"}[item.Kind]
		original := path.Join("users", owner, folder, mid+".original")
		rendered := path.Join("generated", "previews", owner, mid+".jpg")
		mime := "image/jpeg"
		video := strings.HasPrefix(item.types[i], "video/")
		if video {
			rendered = path.Join("generated", "transcodes", owner, mid+".mp4")
			mime = "video/mp4"
		}
		sum := sha256.Sum256(b)
		m := Media{ID: mid, Owner: owner, Kind: folder, Original: original, Rendered: rendered, MIME: mime, SHA: hex.EncodeToString(sum[:]), Bytes: int64(len(b)), State: "ready", Created: item.Created, Project: json.RawMessage(`{"source":"instagram"}`)}
		if _, e = a.H.DB.Exec("INSERT INTO upload_intents VALUES(?,?,?,?,?)", mid, owner, "instagram-writing", original, a.now().Unix()); e != nil {
			return "", false, errors.New("Could not stage media; retry")
		}
		media = append(media, m)
		paths = append(paths, original, rendered)
		if e = writeRoot(root, original, b); e != nil {
			return "", false, errors.New("Could not save original; check storage and retry")
		}
		if video {
			e = a.instagramVideo(ctx, root, original, rendered)
		} else {
			var derivative []byte
			derivative, e = safeImage(b)
			if e == nil {
				e = writeRoot(root, rendered, derivative)
			}
		}
		if e != nil {
			return "", false, fmt.Errorf("Media %d: %s", i+1, e.Error())
		}
	}
	if a.storageCheck(ctx) != nil || ctx.Err() != nil {
		return "", false, errors.New("Storage or connection interrupted; retry this post")
	}
	tx, e := a.H.DB.BeginTx(ctx, nil)
	if e != nil {
		return "", false, errors.New("Could not save post; retry")
	}
	defer tx.Rollback()
	ids, types := []string{}, []string{}
	for _, m := range media {
		_, e = tx.Exec("INSERT INTO media VALUES(?,?,?,?,?,?,?,?,?,?,?)", m.ID, m.Owner, m.Kind, m.Original, m.Rendered, m.MIME, m.SHA, m.Bytes, m.State, m.Created, string(m.Project))
		if e != nil {
			return "", false, errors.New("Could not save media; retry")
		}
		if _, e = tx.Exec("UPDATE upload_intents SET state='complete' WHERE id=?", m.ID); e != nil {
			return "", false, errors.New("Could not finalize media; retry")
		}
		ids = append(ids, m.ID)
		types = append(types, m.MIME)
	}
	state, expires := "published", int64(0)
	if item.Archived {
		state = "archived"
	}
	if item.Kind == "story" {
		state = "archived"
		expires = item.Created + 86400
	}
	id = apphost.ID()
	extras := encode(map[string]any{"instagram": map[string]any{"importedAt": a.now().Unix(), "mediaTypes": types}})
	_, e = tx.Exec("INSERT INTO posts VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", id, owner, item.Kind, item.Caption, "", "", visibility, state, item.Created, item.Created, expires, encode(ids), extras, key)
	if e != nil {
		return "", false, errors.New("Could not save post; retry")
	}
	// Historical imports deliberately do not notify followers for every old post.
	if e = tx.Commit(); e != nil {
		return "", false, errors.New("Could not finish post; retry safely")
	}
	committed = true
	return id, false, nil
}

// The import mutex is held by the caller. A crash rolls back the post/media
// transaction; these intents identify only this importer's uncommitted files.
func (a *App) recoverInstagramMedia(root *os.Root) error {
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	rows, e := a.H.DB.Query("SELECT id,owner,path FROM upload_intents WHERE state='instagram-writing' AND NOT EXISTS(SELECT 1 FROM media WHERE media.id=upload_intents.id)")
	if e != nil {
		return e
	}
	type intent struct{ id, owner, original string }
	intents := []intent{}
	for rows.Next() {
		var item intent
		if e = rows.Scan(&item.id, &item.owner, &item.original); e != nil {
			rows.Close()
			return e
		}
		intents = append(intents, item)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, item := range intents {
		if !validID(item.id) || !instagramPath(item.original) || !instagramPath(item.owner) || !strings.HasPrefix(item.original, "users/"+item.owner+"/") {
			return errors.New("invalid import intent")
		}
		for _, p := range []string{item.original, path.Join("generated", "previews", item.owner, item.id+".jpg"), path.Join("generated", "transcodes", item.owner, item.id+".mp4")} {
			if e = root.Remove(filepath.FromSlash(p)); e != nil && !os.IsNotExist(e) {
				return e
			}
		}
		if _, e = a.H.DB.Exec("DELETE FROM upload_intents WHERE id=? AND state='instagram-writing'", item.id); e != nil {
			return e
		}
	}
	return nil
}

func (a *App) instagramVideo(ctx context.Context, root *os.Root, original, rendered string) error {
	if a.FFmpeg == "" || a.FFprobe == "" {
		return errors.New("video needs FFmpeg and FFprobe")
	}
	source := filepath.Join(a.Root, filepath.FromSlash(original))
	if !withinRoot(a.Root, source) {
		return errors.New("video unavailable")
	}
	probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, a.FFprobe, "-v", "error", "-protocol_whitelist", "file,pipe", "-show_entries", "format=duration", "-of", "json", source)
	cmd.Dir = a.Root
	b, e := cmd.Output()
	var info struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if e != nil || len(b) > 8192 || json.Unmarshal(b, &info) != nil {
		return errors.New("cannot read video duration")
	}
	duration, e := strconv.ParseFloat(info.Format.Duration, 64)
	if e != nil || math.IsNaN(duration) || math.IsInf(duration, 0) || duration <= 0 || duration > instagramVideoSeconds {
		return errors.New("imported videos must be at most 10 minutes; this clip was not imported")
	}
	if e = root.MkdirAll(filepath.FromSlash(path.Dir(rendered)), 0700); e != nil {
		return errors.New("cannot create video preview")
	}
	runCtx, stop := context.WithTimeout(ctx, 10*time.Minute)
	defer stop()
	cmd = exec.CommandContext(runCtx, a.FFmpeg, "-nostdin", "-hide_banner", "-loglevel", "error", "-max_alloc", "134217728", "-n", "-protocol_whitelist", "file,pipe", "-threads", "2", "-i", source, "-map", "0:v:0", "-map", "0:a:0?", "-vf", "scale=1280:1280:force_original_aspect_ratio=decrease:force_divisible_by=2,setsar=1", "-c:v", "libx264", "-preset", "veryfast", "-crf", "25", "-pix_fmt", "yuv420p", "-c:a", "aac", "-b:a", "128k", "-movflags", "+faststart", "-threads", "2", "-t", number(duration), "-map_metadata", "-1", filepath.Join(a.Root, filepath.FromSlash(rendered)))
	cmd.Dir = a.Root
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if e = cmd.Run(); e != nil {
		return errors.New("video conversion failed; check FFmpeg and retry")
	}
	infoFile, e := root.Stat(filepath.FromSlash(rendered))
	if e != nil || infoFile.Size() < 1 || infoFile.Size() > 120<<20 {
		return errors.New("converted video exceeds the output limit")
	}
	return nil
}
