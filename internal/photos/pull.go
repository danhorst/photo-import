package photos

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Pullable reports whether asset a should be pulled into the archive: a photo
// (not a movie) from an allowlisted device, not one of our own published
// derivatives (identified by uuid — the assets we stamped with a catalogKey
// are exactly those recorded at publish), and captured on/after since when
// since is set.
func Pullable(a Asset, allowedDevices map[string]bool, publishedUUIDs map[string]bool, since time.Time) bool {
	if a.IsMovie {
		return false
	}
	if !allowedDevices[strings.ToLower(a.Device())] {
		return false
	}
	if publishedUUIDs[a.UUID] {
		return false
	}
	if !since.IsZero() && a.CaptureTime.Before(since) {
		return false
	}
	return true
}

// AllowedDevices normalizes a device allowlist for Pullable.
func AllowedDevices(devices []string) map[string]bool {
	m := map[string]bool{}
	for _, d := range devices {
		m[strings.ToLower(strings.TrimSpace(d))] = true
	}
	return m
}

// Export exports the given asset uuids into dir via osxphotos, with --update
// so already-exported assets are skipped and --skip-live so Live Photo motion
// components are left behind.
func (OSXPhotos) Export(dir string, uuids []string) error {
	if len(uuids) == 0 {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp("", "photo-management-pull-*.uuids")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	for _, u := range uuids {
		if _, err := f.WriteString(u + "\n"); err != nil {
			f.Close()
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}

	out, err := exec.Command("osxphotos", "export", dir,
		"--uuid-from-file", f.Name(),
		"--update", "--skip-live").CombinedOutput()
	if err != nil {
		return fmt.Errorf("osxphotos export: %v: %s", err, out)
	}
	return nil
}
