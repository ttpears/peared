package daemon

import (
	"context"
	"strings"
)

// Adapter represents a Bluetooth controller the daemon can manage. The fields
// focus on identifiers that are stable across restarts so selection logic can
// favour consistent hardware when multiple adapters are present.
type Adapter struct {
	// ID is a stable identifier for the adapter (e.g. D-Bus object path or
	// kernel name like hci0).
	ID string

	// Address is the MAC address associated with the adapter.
	Address string

	// Alias is a human-friendly label surfaced by BlueZ.
	Alias string

	// Powered indicates whether the adapter radio is currently powered on.
	Powered bool
}

// Matches returns true when the adapter corresponds to the provided identifier.
// It performs case-insensitive comparisons across the adapter's ID, address,
// and alias so callers can provide whichever value they have available.
func (a Adapter) Matches(identifier string) bool {
	id := strings.TrimSpace(identifier)
	if id == "" {
		return false
	}

	if strings.EqualFold(a.ID, id) {
		return true
	}

	if strings.EqualFold(a.Address, id) {
		return true
	}

	if strings.EqualFold(a.Alias, id) {
		return true
	}

	return false
}

// AdapterProvider knows how to discover adapters that are currently available
// on the system.
type AdapterProvider interface {
	ListAdapters(ctx context.Context) ([]Adapter, error)
}

// AdapterProviderFunc adapts a function to the AdapterProvider interface so
// tests can easily stub discovery logic.
type AdapterProviderFunc func(ctx context.Context) ([]Adapter, error)

// ListAdapters implements AdapterProvider.
func (f AdapterProviderFunc) ListAdapters(ctx context.Context) ([]Adapter, error) {
	return f(ctx)
}
