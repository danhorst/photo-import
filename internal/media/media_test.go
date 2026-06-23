package media

import "testing"

func TestIsMedia(t *testing.T) {
	cases := map[string]bool{
		"DSCF1297.RAF":      true,
		"IMG_0001.JPG":      true,
		"photo.jpeg":        true,
		"clip.MOV":          true,
		"pic.heic":          true,
		"scan.tiff":         true,
		"Photo Library.cop": false,
		"Thumbs.db":         false,
		".DS_Store":         false,
		"._IMG_0001.JPG":    false,
		"notes.txt":         false,
		"Photo Library.cof": false,
	}
	for name, want := range cases {
		if got := IsMedia(name); got != want {
			t.Errorf("IsMedia(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestIsExcludedDir(t *testing.T) {
	if !IsExcludedDir("Catalog") {
		t.Error("Catalog should be excluded")
	}
	if IsExcludedDir("2026") {
		t.Error("2026 should not be excluded")
	}
}
