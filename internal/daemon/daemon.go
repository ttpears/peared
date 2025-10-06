package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// Options configures the behavior of the daemon when constructed.
type Options struct {
	// PreferredAdapter is the adapter identifier the daemon should try to use
	// first. Leaving it empty defers to automatic selection.
	PreferredAdapter string

	// AdapterProvider discovers adapters present on the system.
	AdapterProvider AdapterProvider

	// Logger allows callers to provide a slog.Logger configured with project
	// defaults. A sensible default logger is used when nil.
	Logger *slog.Logger

	// ConfigSource is the path used to load configuration. It is surfaced in logs
	// to help troubleshoot configuration resolution.
	ConfigSource string

	// ConfigLoaded indicates whether a configuration file was found on disk.
	ConfigLoaded bool
}

// Daemon represents the long-running coordination process that will manage
// Bluetooth adapters and connections.
type Daemon struct {
	preferredAdapter string
	log              *slog.Logger
	configSource     string
	configLoaded     bool

	mu            sync.RWMutex
	adapterProv   AdapterProvider
	activeAdapter *Adapter
}

// New constructs a Daemon from the provided options.
func New(opts Options) (*Daemon, error) {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
		if logger == nil {
			logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
		}
	}

	provider := opts.AdapterProvider
	if provider == nil {
		provider = DefaultAdapterProvider()
	}

	return &Daemon{
		preferredAdapter: opts.PreferredAdapter,
		log:              logger,
		configSource:     opts.ConfigSource,
		configLoaded:     opts.ConfigLoaded,
		adapterProv:      provider,
	}, nil
}

// Run starts the daemon loop and blocks until the context is cancelled or an
// unrecoverable error occurs.
func (d *Daemon) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("nil context passed to Run")
	}

	if d.adapterProv != nil {
		if err := d.refreshAdapters(ctx); err != nil {
			return err
		}
	}

	activeAdapter := ""
	if adapter, ok := d.ActiveAdapter(); ok {
		activeAdapter = adapter.ID
		if activeAdapter == "" {
			activeAdapter = adapter.Address
		}
	}

	d.log.Info("daemon started", "preferred_adapter", d.preferredAdapter, "config_source", d.configSource, "config_loaded", d.configLoaded, "active_adapter", activeAdapter)
	<-ctx.Done()

	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		d.log.Error("daemon exiting due to context error", "error", err)
		return err
	}

	d.log.Info("daemon stopped")
	return nil
}

// ActiveAdapter returns the adapter currently selected by the daemon.
func (d *Daemon) ActiveAdapter() (Adapter, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.activeAdapter == nil {
		return Adapter{}, false
	}

	adapter := *d.activeAdapter
	return adapter, true
}

func (d *Daemon) refreshAdapters(ctx context.Context) error {
	if d.adapterProv == nil {
		return errors.New("adapter provider not configured")
	}

	adapters, err := d.adapterProv.ListAdapters(ctx)
	if err != nil {
		return fmt.Errorf("list adapters: %w", err)
	}

	if len(adapters) == 0 {
		return errors.New("no adapters discovered")
	}

	chosen, err := SelectAdapter(d.preferredAdapter, adapters)
	if err != nil {
		return err
	}

	d.mu.Lock()
	d.activeAdapter = &Adapter{
		ID:        chosen.ID,
		Address:   chosen.Address,
		Alias:     chosen.Alias,
		Powered:   chosen.Powered,
		Transport: chosen.Transport,
	}
	d.mu.Unlock()

	return nil
}
