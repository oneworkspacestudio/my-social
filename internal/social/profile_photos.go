package social

import (
	"bytes"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"net/http"
)

// Only a sanitized, bounded derivative becomes visible to other signed-in members.
// The original remains in the existing isolated owner-only photo adapter.
func profileImage(data []byte) ([]byte, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 24000000 {
		return nil, &photoError{}
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	size := min(cfg.Width, cfg.Height)
	bounds := img.Bounds()
	x := bounds.Min.X + (cfg.Width-size)/2
	y := bounds.Min.Y + (cfg.Height-size)/2
	out := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.CatmullRom.Scale(out, out.Bounds(), img, image.Rect(x, y, x+size, y+size), draw.Src, nil)
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, out, &jpeg.Options{Quality: 88})
	return buf.Bytes(), err
}

type photoError struct{}

func (*photoError) Error() string { return "Choose a JPEG, PNG or GIF up to 24 megapixels" }
func (a *App) setProfilePhoto(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Media string `json:"media"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	m, err := a.media(in.Media)
	if err != nil || m.Owner != uid(r) || m.Kind != "avatars" || m.State != "ready" || m.MIME != "image/jpeg" || m.Rendered == "" {
		apphost.Problem(w, 400, "Upload your own profile photo first")
		return
	}
	if err = a.storageCheck(r.Context()); err != nil {
		apphost.Problem(w, 503, "Your photo storage is unavailable. Please retry when it reconnects.")
		return
	}
	_, err = a.H.DB.Exec("INSERT INTO profile_photos(user_id,media_id) VALUES(?,?) ON CONFLICT(user_id) DO UPDATE SET media_id=excluded.media_id", uid(r), m.ID)
	if err != nil {
		a.fail(w, err)
		return
	}
	apphost.JSON(w, 200, a.getProfile(uid(r)))
}
func (a *App) removeProfilePhoto(w http.ResponseWriter, r *http.Request) {
	if _, err := a.H.DB.Exec("DELETE FROM profile_photos WHERE user_id=?", uid(r)); err != nil {
		a.fail(w, err)
		return
	}
	apphost.JSON(w, 200, a.getProfile(uid(r)))
}
