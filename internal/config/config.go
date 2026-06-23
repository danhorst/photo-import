// Package config resolves the photo library and index database locations from
// a TOML file, command-line flags, and built-in defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultLibrary is the photo library root used when nothing else is set.
	DefaultLibrary = "/Volumes/Photos"
	indexFileName  = ".photo-index.db"
)

// Config holds the resolved library root and index database path.
type Config struct {
	Library  string `toml:"library"`
	Database string `toml:"database"`
}

// Path returns the config file location, honoring XDG_CONFIG_HOME and falling
// back to ~/.config (not os.UserConfigDir, which on macOS points elsewhere).
func Path() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "photo-import", "photo-import.toml")
}

// Load resolves configuration. For each value the order is: flag > file > default.
// An empty flag string means "unset". The database defaults to a dotfile inside
// the library so the index travels with the drive.
func Load(libFlag, dbFlag string) (Config, error) {
	cfg := Config{Library: DefaultLibrary}

	if data, err := os.ReadFile(Path()); err == nil {
		if _, err := toml.Decode(string(data), &cfg); err != nil {
			return cfg, fmt.Errorf("parsing %s: %w", Path(), err)
		}
	} else if !os.IsNotExist(err) {
		return cfg, err
	}

	if libFlag != "" {
		cfg.Library = libFlag
	}
	if dbFlag != "" {
		cfg.Database = dbFlag
	}
	if cfg.Library == "" {
		cfg.Library = DefaultLibrary
	}
	if cfg.Database == "" {
		cfg.Database = filepath.Join(cfg.Library, indexFileName)
	}
	return cfg, nil
}
