package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSysfsAdapterProviderListsAdapters(t *testing.T) {
	dir := t.TempDir()

	adapterDir := filepath.Join(dir, "hci0")
	if err := os.Mkdir(adapterDir, 0o755); err != nil {
		t.Fatalf("failed to create adapter dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(adapterDir, "address"), []byte("AA:BB:CC:DD:EE:FF\n"), 0o644); err != nil {
		t.Fatalf("failed to write address: %v", err)
	}

	if err := os.WriteFile(filepath.Join(adapterDir, "name"), []byte("Test Adapter\n"), 0o644); err != nil {
		t.Fatalf("failed to write name: %v", err)
	}

	if err := os.WriteFile(filepath.Join(adapterDir, "powered"), []byte("1\n"), 0o644); err != nil {
		t.Fatalf("failed to write powered: %v", err)
	}

	provider := NewSysfsAdapterProvider(dir)
	adapters, err := provider.ListAdapters(context.Background())
	if err != nil {
		t.Fatalf("ListAdapters returned error: %v", err)
	}

	if len(adapters) != 1 {
		t.Fatalf("expected 1 adapter, got %d", len(adapters))
	}

	adapter := adapters[0]
	if adapter.ID != "hci0" {
		t.Errorf("unexpected adapter ID: %q", adapter.ID)
	}

	if adapter.Address != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("unexpected address: %q", adapter.Address)
	}

	if adapter.Alias != "Test Adapter" {
		t.Errorf("unexpected alias: %q", adapter.Alias)
	}

	if !adapter.Powered {
		t.Errorf("expected adapter to be powered")
	}
}

func TestSysfsAdapterProviderHandlesMissingFiles(t *testing.T) {
	dir := t.TempDir()

	adapterDir := filepath.Join(dir, "hci1")
	if err := os.Mkdir(adapterDir, 0o755); err != nil {
		t.Fatalf("failed to create adapter dir: %v", err)
	}

	provider := NewSysfsAdapterProvider(dir)
	adapters, err := provider.ListAdapters(context.Background())
	if err != nil {
		t.Fatalf("ListAdapters returned error: %v", err)
	}

	if len(adapters) != 1 {
		t.Fatalf("expected 1 adapter, got %d", len(adapters))
	}

	adapter := adapters[0]
	if adapter.ID != "hci1" {
		t.Errorf("unexpected adapter ID: %q", adapter.ID)
	}

	if adapter.Address != "" {
		t.Errorf("expected empty address, got %q", adapter.Address)
	}

	if adapter.Alias != "" {
		t.Errorf("expected empty alias, got %q", adapter.Alias)
	}

	if adapter.Powered {
		t.Errorf("expected adapter to be unpowered by default")
	}
}

func TestSysfsAdapterProviderHonorsContextCancellation(t *testing.T) {
	dir := t.TempDir()

	adapterDir := filepath.Join(dir, "hci2")
	if err := os.Mkdir(adapterDir, 0o755); err != nil {
		t.Fatalf("failed to create adapter dir: %v", err)
	}

	provider := NewSysfsAdapterProvider(dir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := provider.ListAdapters(ctx); err == nil {
		t.Fatalf("expected error when context cancelled")
	}
}
