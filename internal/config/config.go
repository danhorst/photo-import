// Package config resolves the photo library and index database locations from
// a TOML file, command-line flags, and built-in defaults, and reads/writes that
// file.
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

// Keys lists the settable configuration keys.
var Keys = []string{"library", "database"}

// Config holds the library root and index database path. Empty fields are
// omitted when written so the file only records what the user set.
type Config struct {
	Library  string `toml:"library,omitempty"`
	Database string `toml:"database,omitempty"`
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

// LoadFile reads only the config file, without applying flags or defaults. A
// missing file yields a zero-value Config.
func LoadFile() (Config, error) {
	var cfg Config
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return cfg, fmt.Errorf("parsing %s: %w", Path(), err)
	}
	return cfg, nil
}

// Load resolves configuration. For each value the order is: flag > file >
// default. An empty flag string means "unset". The database defaults to a
// dotfile inside the library so the index travels with the drive.
func Load(libFlag, dbFlag string) (Config, error) {
	cfg, err := LoadFile()
	if err != nil {
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

// Save writes cfg to the config file, creating the directory if needed.
func Save(cfg Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

const defaultTemplate = `# photo-import configuration

# library: photo library root
library = "%s"

# database: content-hash index path
# Defaults to <library>/.photo-index.db. Uncomment to override.
# database = "%s"
`

// WriteDefault writes a commented config file populated with the default values.
func WriteDefault() error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	content := fmt.Sprintf(defaultTemplate, DefaultLibrary, filepath.Join(DefaultLibrary, indexFileName))
	return os.WriteFile(p, []byte(content), 0o644)
}

// Get returns the value of a named key.
func (c Config) Get(key string) (string, error) {
	switch key {
	case "library":
		return c.Library, nil
	case "database":
		return c.Database, nil
	default:
		return "", fmt.Errorf("unknown key %q (want library or database)", key)
	}
}

// Set assigns a named key.
func (c *Config) Set(key, value string) error {
	switch key {
	case "library":
		c.Library = value
	case "database":
		c.Database = value
	default:
		return fmt.Errorf("unknown key %q (want library or database)", key)
	}
	return nil
}
