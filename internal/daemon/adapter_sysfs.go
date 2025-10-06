package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultSysfsBluetoothPath = "/sys/class/bluetooth"

// DefaultAdapterProvider returns an AdapterProvider that discovers adapters
// using the Linux sysfs hierarchy. The provider inspects /sys/class/bluetooth
// by default so the daemon can automatically locate controllers on common
// distributions without additional configuration.
func DefaultAdapterProvider() AdapterProvider {
	return NewSysfsAdapterProvider(defaultSysfsBluetoothPath)
}

// NewSysfsAdapterProvider creates an AdapterProvider that looks for adapters
// under the supplied sysfs path. A zero path falls back to the default
// location.
func NewSysfsAdapterProvider(path string) AdapterProvider {
	if path == "" {
		path = defaultSysfsBluetoothPath
	}

	return &sysfsAdapterProvider{root: path}
}

type sysfsAdapterProvider struct {
	root string
}

func (p *sysfsAdapterProvider) ListAdapters(ctx context.Context) ([]Adapter, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}

	entries, err := os.ReadDir(p.root)
	if err != nil {
		return nil, fmt.Errorf("read sysfs adapters: %w", err)
	}

	var adapters []Adapter
	for _, entry := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		adapterPath := filepath.Join(p.root, name)

		address := readTrimmedFile(filepath.Join(adapterPath, "address"))
		alias := readTrimmedFile(filepath.Join(adapterPath, "name"))
		powered := parseBool(readTrimmedFile(filepath.Join(adapterPath, "powered")))

		adapters = append(adapters, Adapter{
			ID:      name,
			Address: address,
			Alias:   alias,
			Powered: powered,
		})
	}

	return adapters, nil
}

func readTrimmedFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func parseBool(value string) bool {
	if value == "" {
		return false
	}

	switch strings.ToLower(value) {
	case "1", "true", "on", "yes", "y":
		return true
	}

	return false
}
