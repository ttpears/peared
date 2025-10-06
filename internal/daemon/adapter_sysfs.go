package daemon

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
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

// DefaultSysfsPath exposes the default sysfs root used to discover adapters.
// CLI consumers use this to provide guidance when permission issues prevent
// discovery from succeeding.
func DefaultSysfsPath() string {
	return defaultSysfsBluetoothPath
}

type sysfsAdapterProvider struct {
	root string
}

var readDir = os.ReadDir

func (p *sysfsAdapterProvider) ListAdapters(ctx context.Context) ([]Adapter, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}

	entries, err := readDir(p.root)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return nil, &AdapterAccessError{Path: p.root, Err: err}
		}
		return nil, fmt.Errorf("read sysfs adapters: %w", err)
	}

	var adapters []Adapter
	for _, entry := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		mode := entry.Type()
		if !mode.IsDir() && mode&fs.ModeSymlink == 0 {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "hci") {
			continue
		}
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

// AdapterAccessError indicates that the provider could not inspect the sysfs
// directory due to insufficient permissions. It unwraps to fs.ErrPermission so
// callers can detect the condition with errors.Is.
type AdapterAccessError struct {
	Path string
	Err  error
}

// Error implements error.
func (e *AdapterAccessError) Error() string {
	return fmt.Sprintf("read sysfs adapters at %s: %v", e.Path, e.Err)
}

// Unwrap exposes the underlying error.
func (e *AdapterAccessError) Unwrap() error {
	return e.Err
}
