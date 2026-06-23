package config

import (
	"path/filepath"
	"testing"
)

func TestPathHonorsXDG(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	want := filepath.Join(dir, "photo-import", "photo-import.toml")
	if got := Path(); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

func TestSaveLoadFileRoundtrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	var c Config
	if err := c.Set("library", "/Volumes/Archive"); err != nil {
		t.Fatal(err)
	}
	if err := Save(c); err != nil {
		t.Fatal(err)
	}

	got, err := LoadFile()
	if err != nil {
		t.Fatal(err)
	}
	if got.Library != "/Volumes/Archive" {
		t.Errorf("library = %q, want /Volumes/Archive", got.Library)
	}
	if got.Database != "" {
		t.Errorf("database should be unset (omitempty), got %q", got.Database)
	}
}

func TestLoadDerivesDatabaseFromLibrary(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var c Config
	c.Set("library", "/Volumes/Archive")
	if err := Save(c); err != nil {
		t.Fatal(err)
	}

	resolved, err := Load("", "")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/Volumes/Archive", indexFileName)
	if resolved.Database != want {
		t.Errorf("database = %q, want %q", resolved.Database, want)
	}
}

func TestLoadFileMissingIsEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	got, err := LoadFile()
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if got != (Config{}) {
		t.Errorf("missing file should yield zero Config, got %+v", got)
	}
}

func TestSetUnknownKey(t *testing.T) {
	var c Config
	if err := c.Set("bogus", "x"); err == nil {
		t.Error("expected error for unknown key")
	}
}
