// Package organize computes destination paths in the YYYY/MM library layout and
// places files there, moving within a filesystem and copying across.
package organize

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const tsLayout = "2006-01-02--15-04-05-"

// Dest returns <library>/YYYY/MM/YYYY-MM-DD--HH-MM-SS-<origName>.
func Dest(library string, t time.Time, origName string) string {
	return filepath.Join(library,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", int(t.Month())),
		t.Format(tsLayout)+origName,
	)
}

// Place moves src to dst when they share a filesystem, otherwise copies src to
// dst (preserving mtime) and leaves src in place. It creates dst's parent
// directory. The returned bool reports whether the file was moved.
func Place(src, dst string) (moved bool, err error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return false, err
	}
	if err := os.Rename(src, dst); err == nil {
		return true, nil
	} else if !isCrossDevice(err) {
		return false, err
	}
	return false, copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if fi, err := os.Stat(src); err == nil {
		_ = os.Chtimes(dst, fi.ModTime(), fi.ModTime())
	}
	return nil
}

func isCrossDevice(err error) bool {
	var le *os.LinkError
	if errors.As(err, &le) {
		return errors.Is(le.Err, syscall.EXDEV)
	}
	return false
}
