//go:build !windows

package uploader

import "os"

func openSource(p string) (*os.File, error)    { return os.Open(p) }
func AutoInstall(string, string, string) error { return nil }
func Installed(string) bool                    { return false }
func Main([]string, string) (bool, error)      { return false, nil }
func Report(error)                             {}
