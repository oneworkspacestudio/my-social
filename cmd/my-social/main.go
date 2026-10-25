package main

import (
	"fmt"
	"github.com/oneworkspacestudio/my-social/internal/social"
	"github.com/oneworkspacestudio/my-social/internal/uploader"
	"github.com/oneworkspacestudio/my-social/internal/webui"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"os"
)

var version = "0.1.0-preview.10"

func main() {
	if handled, e := uploader.Main(os.Args, version); handled {
		if e != nil {
			uploader.Report(e)
			os.Exit(1)
		}
		return
	}
	if e := apphost.Main(apphost.Config{ID: "social", Name: "My Social", Version: version, Listen: "127.0.0.1:9013", MediaCapture: true}, webui.Files(), func(h *apphost.Host) (apphost.Application, error) { return social.New(h) }); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
