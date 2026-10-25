// Package uploader owns file transfer outside the browser. Its credential is
// bound to a single ZIP job and has no sign-in or publication privileges.
package uploader

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const Limit = int64(512 << 20)

type State struct {
	ID, Base, Grant, Token, Client, File, Fingerprint, Filename, Job, Status, Error string
	Size, Modified, Expires                                                         int64
}
type Chunk struct {
	Offset int64  `json:"offset"`
	Size   int64  `json:"size"`
	SHA    string `json:"sha256"`
}
type Receipt struct {
	ID        string  `json:"id"`
	State     string  `json:"state"`
	Received  int64   `json:"received"`
	Size      int64   `json:"size"`
	ChunkSize int64   `json:"chunkSize"`
	Chunks    []Chunk `json:"chunks"`
}
type HTTPError struct{ Status int }

func (e HTTPError) Error() string { return fmt.Sprintf("The upload server returned HTTP %d", e.Status) }
func secret() string {
	b := make([]byte, 24)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func isSecret(s string) bool {
	if len(s) != 48 {
		return false
	}
	_, e := hex.DecodeString(s)
	return e == nil && strings.ToLower(s) == s
}
func ValidBase(raw string) (*url.URL, error) {
	u, e := url.Parse(raw)
	if e != nil || u.User != nil || u.Hostname() == "" || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" || strings.ContainsAny(raw, "\r\n\\\x00") {
		return nil, errors.New("Invalid Social server address")
	}
	local := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "[::1]" || u.Hostname() == "::1"
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return nil, errors.New("The uploader requires HTTPS")
	}
	if u.Path != "/social" && !(local && u.Path == "") {
		return nil, errors.New("Choose the Social address, ending in /social")
	}
	return u, nil
}
func ParseLaunch(raw string) (State, error) {
	var s State
	u, e := url.Parse(raw)
	if e != nil || len(raw) > 4096 || u.Scheme != "mysocial-upload" || u.Host != "start" || u.Path != "/" || u.User != nil || u.Fragment != "" {
		return s, errors.New("Open the background uploader from My Social")
	}
	q, e := url.ParseQuery(u.RawQuery)
	if e != nil || len(q) != 2 || len(q["server"]) != 1 || len(q["ticket"]) != 1 {
		return s, errors.New("Invalid upload link")
	}
	s.Base = q.Get("server")
	if _, e = ValidBase(s.Base); e != nil {
		return s, e
	}
	ticket := strings.Split(q.Get("ticket"), ".")
	if len(ticket) != 2 || !isSecret(ticket[0]) || !isSecret(ticket[1]) {
		return s, errors.New("Invalid upload permission")
	}
	s.ID = secret()
	s.Grant = ticket[0]
	s.Token = ticket[1]
	s.Client = secret()
	s.Status = "queued"
	s.Expires = time.Now().Add(24 * time.Hour).Unix()
	return s, nil
}
func Fingerprint(f *os.File, name string, size int64) (string, error) {
	h := sha256.New()
	h.Write([]byte(name + "|" + strconv.FormatInt(size, 10)))
	for _, off := range []int64{0, max(int64(0), size-65536)} {
		n := min(int64(65536), size-off)
		if _, e := io.CopyN(h, io.NewSectionReader(f, off, n), n); e != nil {
			return "", e
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func SelectFile(s *State, file string) error {
	if !filepath.IsAbs(file) || strings.HasPrefix(file, `\\`) || strings.ToLower(filepath.Ext(file)) != ".zip" {
		return errors.New("Choose a ZIP on a local drive")
	}
	f, e := openSource(file)
	if e != nil {
		return errors.New("The selected ZIP cannot be opened")
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Size() < 22 || info.Size() > Limit {
		return errors.New("Choose a ZIP up to 512 MiB")
	}
	s.File = file
	s.Filename = filepath.Base(file)
	s.Size = info.Size()
	s.Modified = info.ModTime().UnixNano()
	s.Fingerprint, e = Fingerprint(f, s.Filename, s.Size)
	return e
}

// Transfer may be restarted with the encrypted State. It reads server receipts
// before every resumed chunk and verifies all committed bytes against the file.
func Transfer(ctx context.Context, s *State, save func(*State) error) error {
	if _, e := ValidBase(s.Base); e != nil {
		return e
	}
	if !isSecret(s.Grant) || !isSecret(s.Token) || !isSecret(s.Client) {
		return errors.New("Invalid saved upload permission")
	}
	f, e := openSource(s.File)
	if e != nil {
		return errors.New("The selected ZIP was moved, removed or is unavailable")
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || info.Size() != s.Size || info.ModTime().UnixNano() != s.Modified {
		return errors.New("The ZIP changed. Choose it again in My Social")
	}
	fp, e := Fingerprint(f, s.Filename, s.Size)
	if e != nil || fp != s.Fingerprint {
		return errors.New("The ZIP changed. Choose it again in My Social")
	}
	client := http.Client{Timeout: 5 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	call := func(action, method, token string, body []byte, digest string) (Receipt, error) {
		var receipt Receipt
		req, e := http.NewRequestWithContext(ctx, method, s.Base+"/api/imports/instagram/native/"+s.Grant+"/"+action, bytes.NewReader(body))
		if e != nil {
			return receipt, errors.New("Invalid upload request")
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "MySocial-Uploader/1")
		req.Header.Set("Content-Type", "application/json")
		if action == "claim" {
			req.Header.Set("X-Upload-Client", s.Client)
		}
		if digest != "" {
			req.Header.Set("X-Chunk-SHA256", digest)
			req.Header.Set("Content-Type", "application/octet-stream")
		}
		resp, e := client.Do(req)
		if e != nil {
			if ctx.Err() != nil {
				return receipt, ctx.Err()
			}
			return receipt, errors.New("Upload connection interrupted")
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return receipt, HTTPError{resp.StatusCode}
		}
		e = json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&receipt)
		if e != nil || receipt.Size != s.Size || receipt.Received < 0 || receipt.Received > s.Size || receipt.ChunkSize != 2<<20 || !isSecret(receipt.ID) {
			return receipt, errors.New("Invalid upload receipt from server")
		}
		return receipt, nil
	}
	verify := func(r Receipt) error {
		var offset int64
		for _, c := range r.Chunks {
			if c.Offset != offset || c.Size <= 0 || c.Size > 2<<20 || c.Offset+c.Size > s.Size {
				return errors.New("Invalid saved chunk receipt")
			}
			h := sha256.New()
			if _, e := io.CopyN(h, io.NewSectionReader(f, c.Offset, c.Size), c.Size); e != nil {
				return e
			}
			if hex.EncodeToString(h.Sum(nil)) != c.SHA {
				return errors.New("The ZIP differs from the saved upload")
			}
			offset += c.Size
		}
		if offset != r.Received {
			return errors.New("Incomplete server receipt")
		}
		return nil
	}
	var receipt Receipt
	delay := time.Second
	retry := func(err error) error {
		var h HTTPError
		if errors.As(err, &h) && h.Status < 500 && h.Status != 429 && h.Status != 408 && h.Status != 409 {
			return err
		}
		if time.Now().Unix() >= s.Expires {
			return errors.New("Upload permission expired. Start a new upload in Social")
		}
		s.Status = "reconnecting"
		s.Error = "Waiting for the upload server to reconnect"
		if e := save(s); e != nil {
			return e
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay = min(delay*2, 30*time.Second)
		return nil
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if s.Job == "" {
			body, _ := json.Marshal(map[string]any{"filename": s.Filename, "size": s.Size, "fingerprint": s.Fingerprint})
			receipt, e = call("claim", "POST", s.Token, body, "")
			if e == nil {
				s.Job = receipt.ID
				if e = save(s); e != nil {
					return e
				}
			}
		} else {
			receipt, e = call("status", "GET", s.Client, nil, "")
		}
		if e != nil {
			if e = retry(e); e != nil {
				return e
			}
			continue
		}
		if receipt.ID != s.Job {
			return errors.New("The upload job changed")
		}
		if receipt.State == "cancelled" {
			s.Status = "cancelled"
			s.Error = ""
			return save(s)
		}
		if receipt.State != "uploading" {
			s.Status = "uploaded"
			s.Error = ""
			return save(s)
		}
		if e = verify(receipt); e != nil {
			return e
		}
		s.Status = "uploading"
		s.Error = ""
		if e = save(s); e != nil {
			return e
		}
		for receipt.Received < s.Size {
			off := receipt.Received
			n := min(receipt.ChunkSize, s.Size-off)
			buf := make([]byte, n)
			if _, e = io.ReadFull(io.NewSectionReader(f, off, n), buf); e != nil {
				return errors.New("The ZIP could not be read")
			}
			hash := sha256.Sum256(buf)
			receipt, e = call("chunks?offset="+strconv.FormatInt(off, 10), "PUT", s.Client, buf, hex.EncodeToString(hash[:]))
			if e != nil {
				break
			}
			if receipt.ID != s.Job || receipt.Received != off+n {
				return errors.New("Upload receipt did not confirm this chunk")
			}
			delay = time.Second
		}
		if e != nil {
			if e = retry(e); e != nil {
				return e
			}
			continue
		}
		_, e = call("complete", "POST", s.Client, []byte("{}"), "")
		if e != nil {
			if e = retry(e); e != nil {
				return e
			}
			continue
		}
		s.Status = "uploaded"
		s.Error = ""
		return save(s)
	}
}
