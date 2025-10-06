package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/peared/peared/internal/config"
	"github.com/peared/peared/internal/daemon"
)

func main() {
	var adapter string
	var configPath string
	var logLevel string

	flag.StringVar(&adapter, "adapter", "", "Preferred adapter name or MAC address to prioritize")
	flag.StringVar(&configPath, "config", "", "Path to configuration file (defaults to XDG config directory)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: parseLevel(logLevel)}))

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if adapter == "" {
		adapter = cfg.Daemon.PreferredAdapter
	}

	d, err := daemon.New(daemon.Options{
		PreferredAdapter: adapter,
		Logger:           logger,
		ConfigSource:     cfg.Source,
		ConfigLoaded:     cfg.Loaded,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure daemon: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := d.Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		fmt.Fprintf(os.Stderr, "daemon exited with error: %v\n", err)
		os.Exit(1)
	}
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(level) {
	case "debug":
		lvl := slog.LevelDebug
		return lvl
	case "warn", "warning":
		lvl := slog.LevelWarn
		return lvl
	case "error", "err":
		lvl := slog.LevelError
		return lvl
	default:
		lvl := slog.LevelInfo
		return lvl
	}
}
