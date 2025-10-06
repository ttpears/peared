package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the on-disk configuration for the daemon and ancillary tools.
type Config struct {
	// Source tracks the path used to load the configuration. It is informational only.
	Source string `yaml:"-"`

	// Loaded reports whether the configuration file existed and was decoded.
	Loaded bool `yaml:"-"`

	Daemon DaemonConfig `yaml:"daemon"`
}

// DaemonConfig holds daemon-specific options from the configuration file.
type DaemonConfig struct {
	PreferredAdapter string `yaml:"preferred_adapter"`
}

// ResolvePath determines the configuration path to use. Explicit paths are honored first,
// followed by the PEARED_CONFIG environment variable, and finally the default XDG location.
func ResolvePath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	if env := os.Getenv("PEARED_CONFIG"); env != "" {
		return env, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	return filepath.Join(configDir, "peared", "config.yaml"), nil
}

// Load reads configuration from disk, returning default values if the file does not exist.
func Load(path string) (*Config, error) {
	resolved, err := ResolvePath(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{Source: resolved}

	data, err := os.ReadFile(resolved)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config %q: %w", resolved, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", resolved, err)
	}

	cfg.Source = resolved
	cfg.Loaded = true
	return cfg, nil
}
