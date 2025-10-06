package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath(t *testing.T) {
	t.Setenv("PEARED_CONFIG", "")

	explicit, err := ResolvePath("/tmp/peared.yaml")
	if err != nil {
		t.Fatalf("ResolvePath explicit: %v", err)
	}
	if explicit != "/tmp/peared.yaml" {
		t.Fatalf("unexpected explicit path: %q", explicit)
	}

	t.Setenv("PEARED_CONFIG", "/custom/path.yaml")
	resolved, err := ResolvePath("")
	if err != nil {
		t.Fatalf("ResolvePath env: %v", err)
	}
	if resolved != "/custom/path.yaml" {
		t.Fatalf("unexpected env path: %q", resolved)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("PEARED_CONFIG", "")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load missing file: %v", err)
	}
	if cfg.Source != path {
		t.Fatalf("expected source %q, got %q", path, cfg.Source)
	}
	if cfg.Loaded {
		t.Fatalf("expected Loaded to be false")
	}
	if cfg.Daemon.PreferredAdapter != "" {
		t.Fatalf("expected zero PreferredAdapter, got %q", cfg.Daemon.PreferredAdapter)
	}
}

func TestLoadExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "daemon:\n  preferred_adapter: test-adapter\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load existing file: %v", err)
	}
	if !cfg.Loaded {
		t.Fatalf("expected Loaded to be true")
	}
	if cfg.Daemon.PreferredAdapter != "test-adapter" {
		t.Fatalf("unexpected PreferredAdapter: %q", cfg.Daemon.PreferredAdapter)
	}
}
