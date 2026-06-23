// Package media classifies files as managed photo/video media and identifies
// directories to skip when scanning a library.
package media

import (
	"path/filepath"
	"strings"
)

// mediaExt is the set of managed media extensions (lowercase, with dot).
var mediaExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".heic": true, ".png": true,
	".gif": true, ".tif": true, ".tiff": true,
	".cr2": true, ".raf": true, ".dng": true, ".crw": true,
	".mov": true, ".mp4": true, ".avi": true,
}

// excludeDir names directories that are never descended into.
var excludeDir = map[string]bool{
	"Catalog":                   true,
	"$RECYCLE.BIN":              true,
	"System Volume Information": true,
	".Spotlight-V100":           true,
	".fseventsd":                true,
	".Trashes":                  true,
}

// IsMedia reports whether a filename is a managed media file, excluding
// AppleDouble and OS sidecar files.
func IsMedia(name string) bool {
	if strings.HasPrefix(name, "._") || name == ".DS_Store" || name == "Thumbs.db" {
		return false
	}
	return mediaExt[strings.ToLower(filepath.Ext(name))]
}

// IsExcludedDir reports whether a directory should be skipped during a walk.
func IsExcludedDir(name string) bool {
	return excludeDir[name]
}
