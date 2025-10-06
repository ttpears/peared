package daemon

import (
	"context"
	"fmt"
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

	// Transport attempts to describe the bus used by the adapter (usb, pci,
	// platform, etc.). Selection logic can prefer specific transports when a
	// preferred adapter is not explicitly configured.
	Transport AdapterTransport
}

// AdapterTransport identifies the bus type used by an adapter. Values are best
// effort because sysfs metadata is not always present or consistent across
// kernels.
type AdapterTransport string

const (
	AdapterTransportUnknown  AdapterTransport = "unknown"
	AdapterTransportUSB      AdapterTransport = "usb"
	AdapterTransportPCI      AdapterTransport = "pci"
	AdapterTransportPlatform AdapterTransport = "platform"
)

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

// SelectAdapter chooses the most appropriate adapter from the supplied list.
// A preferred adapter identifier is honoured when provided; otherwise adapters
// attached via USB are prioritised. If no USB adapter is present, the first
// entry is returned.
func SelectAdapter(preferred string, adapters []Adapter) (Adapter, error) {
	if len(adapters) == 0 {
		return Adapter{}, fmt.Errorf("select adapter: no adapters supplied")
	}

	if preferred = strings.TrimSpace(preferred); preferred != "" {
		for _, adapter := range adapters {
			if adapter.Matches(preferred) {
				return adapter, nil
			}
		}
	}

	for _, adapter := range adapters {
		if adapter.Transport == AdapterTransportUSB {
			return adapter, nil
		}
	}

	return adapters[0], nil
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
