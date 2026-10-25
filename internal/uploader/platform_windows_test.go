//go:build windows

package uploader

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestPrivateEncryptedQueueAndIsolatedProtocolRegistration(t *testing.T) {
	root := t.TempDir()
	if e := privateDir(filepath.Join(root, "queue")); e != nil {
		t.Fatal(e)
	}
	s := State{ID: secret(), Token: secret(), Client: secret(), File: `C:\private\instagram.zip`, Status: "queued"}
	if e := saveAt(root, &s); e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(root, "queue", s.ID+".upload")
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(b, []byte(s.Token)) || bytes.Contains(b, []byte(s.File)) || bytes.Contains(b, []byte("queued")) {
		t.Fatal("queue contains plaintext secrets or file paths")
	}
	restored, e := loadAt(p)
	if e != nil || restored != s {
		t.Fatal("encrypted queue cannot resume", e)
	}
	source, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	key := `Software\MySocialUploaderTests\` + secret()
	t.Cleanup(func() {
		registry.DeleteKey(registry.CURRENT_USER, key+`\protocol\shell\open\command`)
		registry.DeleteKey(registry.CURRENT_USER, key+`\protocol\shell\open`)
		registry.DeleteKey(registry.CURRENT_USER, key+`\protocol\shell`)
		registry.DeleteKey(registry.CURRENT_USER, key+`\protocol`)
		registry.DeleteKey(registry.CURRENT_USER, key+`\run`)
		registry.DeleteKey(registry.CURRENT_USER, key)
	})
	if e = installAt(root, source, "test", "https://social.example/social", key+`\protocol`, key+`\run`); e != nil {
		t.Fatal(e)
	}
	k, e := registry.OpenKey(registry.CURRENT_USER, key+`\protocol\shell\open\command`, registry.QUERY_VALUE)
	if e != nil {
		t.Fatal(e)
	}
	defer k.Close()
	command, _, e := k.GetStringValue("")
	if e != nil || !strings.HasSuffix(command, `" --mysocial-upload "%1"`) {
		t.Fatal("unsafe protocol command")
	}
	// The selected original is held with read sharing only, preventing mutation
	// while an unattended upload is in progress.
	file := filepath.Join(root, "locked.zip")
	os.WriteFile(file, bytes.Repeat([]byte{1}, 22), 0600)
	f, e := openSource(file)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if e = os.WriteFile(file, []byte("changed"), 0600); e == nil {
		t.Fatal("upload original could change while held open")
	}
}
