package main

import (
	"fmt"
	"github.com/oneworkspacestudio/my-social/internal/social"
	"github.com/oneworkspacestudio/my-social/internal/webui"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"os"
)

var version = "0.1.0-preview.3"

func main() {
	if e := apphost.Main(apphost.Config{ID: "social", Name: "My Social", Version: version, Listen: "127.0.0.1:9013", MediaCapture: true}, webui.Files(), func(h *apphost.Host) (apphost.Application, error) { return social.New(h) }); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
