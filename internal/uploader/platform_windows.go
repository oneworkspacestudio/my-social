//go:build windows

package uploader

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type installation struct {
	Version, Executable, SHA256 string
	Servers                     map[string]bool
	Receivers                   map[string]string
}

func home() (string, error) {
	p, e := os.UserCacheDir()
	if e != nil {
		return "", e
	}
	return filepath.Join(p, "MySocial", "Uploader"), nil
}
func noLinks(p string) error {
	for {
		i, e := os.Lstat(p)
		if e == nil && i.Mode()&os.ModeSymlink != 0 {
			return errors.New("Linked uploader files are not allowed")
		}
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
	}
	return nil
}
func privateDir(p string) error {
	if e := noLinks(p); e != nil {
		return e
	}
	if e := os.MkdirAll(p, 0700); e != nil {
		return e
	}
	nativePath, e := windows.UTF16PtrFromString(p)
	if e != nil {
		return e
	}
	volume := make([]uint16, 32768)
	if e = windows.GetVolumePathName(nativePath, &volume[0], uint32(len(volume))); e != nil {
		return e
	}
	var flags uint32
	fsName := make([]uint16, 260)
	if e = windows.GetVolumeInformation(&volume[0], nil, 0, nil, nil, &flags, &fsName[0], uint32(len(fsName))); e != nil {
		return e
	}
	if flags&windows.FILE_PERSISTENT_ACLS == 0 || windows.GetDriveType(&volume[0]) == windows.DRIVE_REMOTE {
		return errors.New("The Windows uploader requires a local NTFS folder for its private queue")
	}
	token, e := windows.OpenCurrentProcessToken()
	if e != nil {
		return e
	}
	defer token.Close()
	u, e := token.GetTokenUser()
	if e != nil {
		return e
	}
	sd, e := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;" + u.User.Sid.String() + ")")
	if e != nil {
		return e
	}
	acl, _, e := sd.DACL()
	if e != nil {
		return e
	}
	return windows.SetNamedSecurityInfo(p, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
}
func digest(p string) (string, error) {
	f, e := os.Open(p)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	_, e = io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), e
}
func atomicWrite(p string, b []byte) error {
	if e := noLinks(p); e != nil {
		return e
	}
	tmp := p + "." + secret() + ".tmp"
	if e := os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	defer os.Remove(tmp)
	f, e := os.OpenFile(tmp, os.O_RDWR, 0600)
	if e != nil {
		return e
	}
	e = f.Sync()
	f.Close()
	if e != nil {
		return e
	}
	return os.Rename(tmp, p)
}
func protect(b []byte, decrypt bool) ([]byte, error) {
	if len(b) == 0 {
		return nil, errors.New("Empty upload state")
	}
	in := windows.DataBlob{Size: uint32(len(b)), Data: &b[0]}
	var out windows.DataBlob
	var e error
	if decrypt {
		e = windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	} else {
		e = windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	}
	if e != nil {
		return nil, e
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}
func saveAt(root string, s *State) error {
	if !isSecret(s.ID) {
		return errors.New("Invalid queue identity")
	}
	b, e := json.Marshal(s)
	if e != nil {
		return e
	}
	b, e = protect(b, false)
	if e != nil {
		return e
	}
	return atomicWrite(filepath.Join(root, "queue", s.ID+".upload"), b)
}
func loadAt(p string) (State, error) {
	var s State
	if e := noLinks(p); e != nil {
		return s, e
	}
	b, e := os.ReadFile(p)
	if e != nil {
		return s, e
	}
	b, e = protect(b, true)
	if e != nil {
		return s, e
	}
	e = json.Unmarshal(b, &s)
	return s, e
}
func installAt(root, source, version, origin, protocol, runKey string) error {
	if origin != "" {
		if _, e := ValidBase(origin); e != nil {
			return e
		}
	}
	if e := privateDir(root); e != nil {
		return e
	}
	if e := privateDir(filepath.Join(root, "queue")); e != nil {
		return e
	}
	sum, e := digest(source)
	if e != nil {
		return e
	}
	bin := filepath.Join(root, "bin", sum)
	if e = privateDir(bin); e != nil {
		return e
	}
	target := filepath.Join(bin, "MySocial.exe")
	if got, _ := digest(target); got != sum {
		f, e := os.Open(source)
		if e != nil {
			return e
		}
		b, e := io.ReadAll(io.LimitReader(f, 128<<20))
		f.Close()
		if e != nil {
			return e
		}
		if e = atomicWrite(target, b); e != nil {
			return e
		}
		if got, _ = digest(target); got != sum {
			return errors.New("Uploader copy verification failed")
		}
	}
	var old installation
	b, _ := os.ReadFile(filepath.Join(root, "installation.json"))
	_ = json.Unmarshal(b, &old)
	if old.Servers == nil {
		old.Servers = map[string]bool{}
	}
	if origin != "" {
		old.Servers[origin] = true
	}
	old.Version = version
	old.Executable = target
	old.SHA256 = sum
	b, _ = json.Marshal(old)
	if e = atomicWrite(filepath.Join(root, "installation.json"), b); e != nil {
		return e
	}
	key, _, e := registry.CreateKey(registry.CURRENT_USER, protocol, registry.SET_VALUE)
	if e != nil {
		return e
	}
	e = key.SetStringValue("", "URL:My Social background upload")
	if e == nil {
		e = key.SetStringValue("URL Protocol", "")
	}
	key.Close()
	if e != nil {
		return e
	}
	key, _, e = registry.CreateKey(registry.CURRENT_USER, protocol+`\shell\open\command`, registry.SET_VALUE)
	if e != nil {
		return e
	}
	e = key.SetStringValue("", `"`+target+`" --mysocial-upload "%1"`)
	key.Close()
	if e != nil {
		return e
	}
	key, _, e = registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if e != nil {
		return e
	}
	e = key.SetStringValue("MySocialUploader", `"`+target+`" --mysocial-upload-worker`)
	key.Close()
	return e
}
func Install(version, origin string) error {
	root, e := home()
	if e != nil {
		return e
	}
	source, e := os.Executable()
	if e != nil {
		return e
	}
	return installAt(root, source, version, strings.TrimRight(origin, "/"), `Software\Classes\mysocial-upload`, `Software\Microsoft\Windows\CurrentVersion\Run`)
}
func AutoInstall(version, origin, workspace, issuer, subject string) error {
	if workspace == "" || os.Getenv("MYSOCIAL_UPLOADER_DISABLE") == "1" {
		return nil
	}
	if e := Install(version, origin); e != nil {
		return e
	}
	root, e := home()
	if e != nil {
		return e
	}
	if issuer != "" && subject != "" {
		var info installation
		b, e := os.ReadFile(filepath.Join(root, "installation.json"))
		if e != nil || json.Unmarshal(b, &info) != nil {
			return errors.New("Could not read uploader registration")
		}
		if info.Receivers == nil {
			info.Receivers = map[string]string{}
		}
		sum := sha256.Sum256([]byte(issuer + "\x00" + subject))
		info.Receivers[strings.TrimRight(origin, "/")] = hex.EncodeToString(sum[:])
		b, _ = json.Marshal(info)
		if e = atomicWrite(filepath.Join(root, "installation.json"), b); e != nil {
			return e
		}
	}
	entries, _ := os.ReadDir(filepath.Join(root, "queue"))
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".upload") {
			return spawnWorker(root)
		}
	}
	return nil
}
func Installed(origin string) bool {
	root, e := home()
	if e != nil {
		return false
	}
	if noLinks(root) != nil {
		return false
	}
	var i installation
	b, e := os.ReadFile(filepath.Join(root, "installation.json"))
	if e != nil || json.Unmarshal(b, &i) != nil || !i.Servers[strings.TrimRight(origin, "/")] {
		return false
	}
	sum, e := digest(i.Executable)
	source, sourceErr := os.Executable()
	current, currentErr := digest(source)
	if e != nil || sum != i.SHA256 || sourceErr != nil || currentErr != nil || current != sum {
		return false
	}
	k, e := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\mysocial-upload\shell\open\command`, registry.QUERY_VALUE)
	if e != nil {
		return false
	}
	defer k.Close()
	command, _, e := k.GetStringValue("")
	return e == nil && command == `"`+i.Executable+`" --mysocial-upload "%1"`
}
func openSource(p string) (*os.File, error) {
	if !filepath.IsAbs(p) || strings.HasPrefix(p, `\\`) {
		return nil, errors.New("Choose a local ZIP")
	}
	if e := noLinks(p); e != nil {
		return nil, e
	}
	name, e := windows.UTF16PtrFromString(p)
	if e != nil {
		return nil, e
	}
	h, e := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if e != nil {
		return nil, e
	}
	return os.NewFile(uintptr(h), p), nil
}
func message(text string, question bool) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	body, _ := windows.UTF16PtrFromString(text)
	title, _ := windows.UTF16PtrFromString("My Social background uploader")
	flags := uintptr(0x40)
	if question {
		flags = 0x124
	}
	v, _, _ := windows.NewLazySystemDLL("user32.dll").NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), flags)
	return v == 6
}

type openFilename struct {
	Size                  uint32
	Owner, Instance       uintptr
	Filter, Custom        *uint16
	MaxCustom, Index      uint32
	File                  *uint16
	MaxFile               uint32
	FileTitle             *uint16
	MaxTitle              uint32
	Initial, Title        *uint16
	Flags                 uint32
	Offset, Extension     uint16
	Default               *uint16
	Data, Hook            uintptr
	Template              *uint16
	Reserved              unsafe.Pointer
	ReservedSize, FlagsEx uint32
}

func chooseZIP(server string) (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	buf := make([]uint16, 32768)
	filter := []uint16{}
	for _, s := range []string{"Instagram ZIP (*.zip)", "*.zip", ""} {
		filter = append(filter, windows.StringToUTF16(s)...)
	}
	title, _ := windows.UTF16PtrFromString("Upload Instagram ZIP to " + server)
	ext, _ := windows.UTF16PtrFromString("zip")
	dialog := openFilename{File: &buf[0], MaxFile: uint32(len(buf)), Filter: &filter[0], Index: 1, Title: title, Default: ext, Flags: 0x80000 | 0x1000 | 0x800 | 0x4 | 0x8 | 0x2000000 | 0x100000 | 0x20000}
	dialog.Size = uint32(unsafe.Sizeof(dialog))
	lib := windows.NewLazySystemDLL("comdlg32.dll")
	ok, _, _ := lib.NewProc("GetOpenFileNameW").Call(uintptr(unsafe.Pointer(&dialog)))
	if ok == 0 {
		v, _, _ := lib.NewProc("CommDlgExtendedError").Call()
		if v != 0 {
			return "", errors.New("Windows could not open the ZIP picker")
		}
		return "", nil
	}
	return windows.UTF16ToString(buf), nil
}
func spawnWorker(root string) error {
	var i installation
	b, e := os.ReadFile(filepath.Join(root, "installation.json"))
	if e != nil || json.Unmarshal(b, &i) != nil {
		return errors.New("Reinstall the Social uploader")
	}
	sum, e := digest(i.Executable)
	if e != nil || sum != i.SHA256 {
		return errors.New("Uploader verification failed")
	}
	c := exec.Command(i.Executable, "--mysocial-upload-worker")
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	if e := c.Start(); e != nil {
		return e
	}
	_ = c.Process.Release()
	return nil
}

// A mutex belongs to one queue item. A second selected ZIP can claim its
// short-lived permission immediately instead of waiting behind a long upload.
func worker(root string) error {
	entries, e := os.ReadDir(filepath.Join(root, "queue"))
	if e != nil {
		return e
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".upload") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".upload")
		if !isSecret(id) {
			continue
		}
		func() {
			sum := sha256.Sum256([]byte(root + "\x00" + id))
			name, _ := windows.UTF16PtrFromString(`Local\MySocialUpload-` + hex.EncodeToString(sum[:16]))
			m, e := windows.CreateMutex(nil, false, name)
			if e != nil {
				return
			}
			defer windows.CloseHandle(m)
			v, e := windows.WaitForSingleObject(m, 0)
			if e != nil || v != windows.WAIT_OBJECT_0 && v != windows.WAIT_ABANDONED {
				return
			}
			defer windows.ReleaseMutex(m)
			p := filepath.Join(root, "queue", entry.Name())
			state, e := loadAt(p)
			if e != nil || state.ID != id {
				return
			}
			if state.Status == "uploaded" || state.Status == "cancelled" || state.Status == "failed" {
				if time.Now().Unix() > state.Expires {
					_ = os.Remove(p)
				}
				return
			}
			e = Transfer(context.Background(), &state, func(s *State) error { return saveAt(root, s) })
			if e != nil {
				state.Status = "failed"
				state.Error = e.Error()
				_ = saveAt(root, &state)
				reportProblem(&state)
			}
		}()
	}
	return nil
}
func Main(args []string, version string) (bool, error) {
	// An explicit CLI file argument is also a user selection. The registered URI
	// command cannot reach this mode or inject any additional arguments.
	if len(args) > 1 && args[1] == "--mysocial-upload-file" {
		if len(args) != 4 {
			return true, errors.New("Usage: --mysocial-upload-file <upload-link> <ZIP>")
		}
		s, e := ParseLaunch(args[2])
		if e != nil {
			return true, e
		}
		if e = SelectFile(&s, args[3]); e != nil {
			return true, e
		}
		root, e := home()
		if e != nil {
			return true, e
		}
		if e = privateDir(filepath.Join(root, "queue")); e != nil {
			return true, e
		}
		if e = saveAt(root, &s); e != nil {
			return true, e
		}
		return true, worker(root)
	}

	if len(args) == 1 && strings.EqualFold(filepath.Base(os.Args[0]), "MySocialUploader.exe") {
		e := Install(version, "")
		if e == nil {
			message("The Windows uploader is installed. Return to My Social and choose background upload.", false)
		}
		return true, e
	}
	if len(args) == 2 && args[1] == "--mysocial-upload-worker" {
		root, e := home()
		if e != nil {
			return true, e
		}
		return true, worker(root)
	}
	if len(args) < 2 || args[1] != "--mysocial-upload" {
		return false, nil
	}
	if len(args) != 3 {
		return true, errors.New("Invalid background upload command")
	}
	s, e := ParseLaunch(args[2])
	if e != nil {
		return true, e
	}
	if e = InteractiveBase(s); e != nil {
		return true, e
	}
	root, e := home()
	if e != nil {
		return true, e
	}
	var i installation
	b, e := os.ReadFile(filepath.Join(root, "installation.json"))
	if e != nil || json.Unmarshal(b, &i) != nil {
		return true, errors.New("Install the Windows uploader from My Social first")
	}
	account, e := Describe(s)
	if e != nil {
		return true, e
	}
	if !receiverPaired(i, s.Base, account.ID) {
		if !message("Confirm the receiving Social account\n\nAccount: "+safeLabel(account.Name)+"\nEmail: "+safeLabel(account.Email)+"\nAccount ID: "+account.ID[:12]+"\nServer: "+s.Base+"\n\nThis account is not paired with your Windows uploader. Continue only if you intend to send your ZIP to this account.", true) {
			return true, nil
		}
		if i.Receivers == nil {
			i.Receivers = map[string]string{}
		}
		i.Receivers[s.Base] = account.ID
		b, _ = json.Marshal(i)
		if e = atomicWrite(filepath.Join(root, "installation.json"), b); e != nil {
			return true, e
		}
	}
	file, e := chooseZIP(safeLabel(account.Name) + " <" + safeLabel(account.Email) + "> · " + s.Base)
	if e != nil || file == "" {
		return true, e
	}
	if e = SelectFile(&s, file); e != nil {
		return true, e
	}
	if e = saveAt(root, &s); e != nil {
		return true, e
	}
	if e = spawnWorker(root); e != nil {
		return true, e
	}
	message("Your ZIP is uploading in the background. You can close the browser.\nKeep this computer on and the ZIP in its current location. If the connection drops, the uploader retries automatically.\nReturn to My Social to review five sample posts and import.", false)
	return true, nil
}
func Report(err error) { message(err.Error(), false) }

func reportProblem(s *State) {
	if s.Job == "" {
		return
	}
	req, e := http.NewRequest("POST", s.Base+"/api/imports/instagram/native/"+s.Grant+"/problem", bytes.NewBufferString("{}"))
	if e != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+s.Client)
	req.Header.Set("X-Upload-Recipient", s.Recipient)
	client := http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, e := client.Do(req)
	if e == nil {
		resp.Body.Close()
	}
}

func safeLabel(s string) string {
	out := []rune{}
	for _, r := range s {
		if !unicode.IsControl(r) && !unicode.In(r, unicode.Cf) {
			out = append(out, r)
			if len(out) >= 128 {
				break
			}
		}
	}
	return string(out)
}
func receiverPaired(i installation, base, id string) bool {
	return isAccount(id) && i.Receivers[base] == id
}
