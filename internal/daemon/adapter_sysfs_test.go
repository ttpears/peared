package daemon

import (
	"context"
	"errors"
	"io/fs"
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

	deviceDir := filepath.Join(adapterDir, "device")
	if err := os.Mkdir(deviceDir, 0o755); err != nil {
		t.Fatalf("failed to create device dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(deviceDir, "modalias"), []byte("usb:1234\n"), 0o644); err != nil {
		t.Fatalf("failed to write modalias: %v", err)
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

	if adapter.Transport != AdapterTransportUSB {
		t.Errorf("expected usb transport, got %s", adapter.Transport)
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

	if adapter.Transport != AdapterTransportUnknown {
		t.Errorf("expected unknown transport, got %s", adapter.Transport)
	}
}

func TestSysfsAdapterProviderFollowsSymlinks(t *testing.T) {
	dir := t.TempDir()

	realAdapterDir := filepath.Join(dir, "devices", "virtual", "bluetooth", "hci3")
	if err := os.MkdirAll(realAdapterDir, 0o755); err != nil {
		t.Fatalf("failed to create real adapter dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(realAdapterDir, "address"), []byte("11:22:33:44:55:66\n"), 0o644); err != nil {
		t.Fatalf("failed to write address: %v", err)
	}

	linkPath := filepath.Join(dir, "hci3")
	if err := os.Symlink(realAdapterDir, linkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	provider := NewSysfsAdapterProvider(dir)
	adapters, err := provider.ListAdapters(context.Background())
	if err != nil {
		t.Fatalf("ListAdapters returned error: %v", err)
	}

	if len(adapters) != 1 {
		t.Fatalf("expected 1 adapter, got %d", len(adapters))
	}

	if adapters[0].ID != "hci3" {
		t.Fatalf("expected adapter ID hci3, got %s", adapters[0].ID)
	}

	if adapters[0].Address != "11:22:33:44:55:66" {
		t.Fatalf("unexpected address: %q", adapters[0].Address)
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

func TestSysfsAdapterProviderExposesPermissionErrors(t *testing.T) {
	t.Cleanup(func() { readDir = os.ReadDir })
	readDir = func(string) ([]os.DirEntry, error) {
		return nil, fs.ErrPermission
	}

	provider := NewSysfsAdapterProvider("/sys/class/bluetooth")
	_, err := provider.ListAdapters(context.Background())
	if err == nil {
		t.Fatalf("expected permission error")
	}

	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("expected fs.ErrPermission, got %v", err)
	}

	var accessErr *AdapterAccessError
	if !errors.As(err, &accessErr) {
		t.Fatalf("expected AdapterAccessError, got %T", err)
	}

	if accessErr.Path == "" {
		t.Fatalf("expected path to be populated")
	}
}
