// Package exif extracts capture timestamps from media files via a single
// batched exiftool invocation.
package exif

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"
)

// layout is the Go time layout used to parse exiftool's output; strftimeFmt is
// the matching strftime format passed to exiftool's -d flag.
const (
	layout      = "2006-01-02 15:04:05"
	strftimeFmt = "%Y-%m-%d %H:%M:%S"
)

type entry struct {
	SourceFile       string `json:"SourceFile"`
	DateTimeOriginal string `json:"DateTimeOriginal"`
	CreateDate       string `json:"CreateDate"`
}

// Dates returns capture timestamps keyed by the input path. Files whose date is
// missing or unparseable are simply absent from the map; callers fall back to
// the filesystem mtime. The whole batch is read in one exiftool call.
func Dates(paths []string) (map[string]time.Time, error) {
	out := map[string]time.Time{}
	if len(paths) == 0 {
		return out, nil
	}

	// Pass the file list through an args file to avoid command-line length limits.
	args, err := os.CreateTemp("", "photo-import-exif-*.args")
	if err != nil {
		return nil, err
	}
	defer os.Remove(args.Name())
	for _, p := range paths {
		if _, err := args.WriteString(p + "\n"); err != nil {
			args.Close()
			return nil, err
		}
	}
	args.Close()

	cmd := exec.Command("exiftool",
		"-json", "-q",
		"-d", strftimeFmt,
		"-DateTimeOriginal", "-CreateDate",
		"-@", args.Name(),
	)
	stdout, err := cmd.Output()
	// exiftool exits non-zero when some files lack metadata but still emits JSON
	// for the rest, so only treat an empty output as a hard failure.
	if err != nil && len(stdout) == 0 {
		return nil, err
	}

	var entries []entry
	if err := json.Unmarshal(stdout, &entries); err != nil {
		return nil, err
	}
	for _, e := range entries {
		ds := e.DateTimeOriginal
		if ds == "" {
			ds = e.CreateDate
		}
		if ds == "" || strings.HasPrefix(ds, "0000") {
			continue
		}
		t, err := time.ParseInLocation(layout, ds, time.Local)
		if err != nil {
			continue
		}
		out[e.SourceFile] = t
	}
	return out, nil
}
