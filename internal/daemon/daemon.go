package daemon

import (
	"context"
	"errors"
	"log/slog"
	"os"
)

// Options configures the behavior of the daemon when constructed.
type Options struct {
	// PreferredAdapter is the adapter identifier the daemon should try to use
	// first. Leaving it empty defers to automatic selection.
	PreferredAdapter string

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

	return &Daemon{
		preferredAdapter: opts.PreferredAdapter,
		log:              logger,
		configSource:     opts.ConfigSource,
		configLoaded:     opts.ConfigLoaded,
	}, nil
}

// Run starts the daemon loop and blocks until the context is cancelled or an
// unrecoverable error occurs.
func (d *Daemon) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("nil context passed to Run")
	}

	d.log.Info("daemon started", "preferred_adapter", d.preferredAdapter, "config_source", d.configSource, "config_loaded", d.configLoaded)
	<-ctx.Done()

	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		d.log.Error("daemon exiting due to context error", "error", err)
		return err
	}

	d.log.Info("daemon stopped")
	return nil
}
