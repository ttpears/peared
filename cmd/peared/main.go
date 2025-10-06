package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/peared/peared/internal/cli"
	"github.com/peared/peared/internal/daemon"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "shell":
		runShell(os.Args[2:])
	case "adapters":
		runAdapters(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func runShell(args []string) {
	fs := flag.NewFlagSet("shell", flag.ExitOnError)
	logLevel := fs.String("log-level", "info", "Log level (debug, info, warn, error)")
	prompt := fs.String("prompt", "peared> ", "Prompt to display for the interactive shell")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
		os.Exit(2)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: parseLevel(*logLevel)}))
	logger.Info("starting interactive shell")
	defer logger.Info("shell session ended")

	shell := cli.NewShell(os.Stdin, os.Stdout, cli.WithPrompt(*prompt))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := shell.Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info("shell interrupted by context cancellation")
			return
		}
		fmt.Fprintf(os.Stderr, "shell exited with error: %v\n", err)
		os.Exit(1)
	}

	logger.Info("shell exited normally")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Peared CLI\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  peared <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Available Commands:\n")
	fmt.Fprintf(os.Stderr, "  adapters  Inspect Bluetooth adapters available on the host\n")
	fmt.Fprintf(os.Stderr, "  shell     Start an interactive shell session\n")
	fmt.Fprintf(os.Stderr, "  help      Show this message\n")
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

func runAdapters(args []string) {
	if len(args) == 0 {
		adaptersUsage()
		os.Exit(2)
	}

	switch args[0] {
	case "list":
		listAdapters(args[1:])
	case "help", "-h", "--help":
		adaptersUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown adapters command: %s\n\n", args[0])
		adaptersUsage()
		os.Exit(2)
	}
}

func adaptersUsage() {
	fmt.Fprintf(os.Stderr, "Usage: peared adapters <command>\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  list   Discover Bluetooth adapters managed by the host\n")
}

func listAdapters(args []string) {
	flagSet := flag.NewFlagSet("adapters list", flag.ExitOnError)
	sysfsPath := flagSet.String("sysfs", "", "Override the sysfs root used to discover adapters (advanced)")
	if err := flagSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse adapters flags: %v\n", err)
		os.Exit(2)
	}

	ctx := context.Background()

	provider := daemon.NewSysfsAdapterProvider(*sysfsPath)
	adapters, err := provider.ListAdapters(ctx)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			path := *sysfsPath
			if path == "" {
				path = daemon.DefaultSysfsPath()
			}
			fmt.Fprintf(os.Stderr, "failed to list adapters: insufficient permissions to read %s. Run the command with elevated privileges or adjust udev rules to grant access.\n", path)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "failed to list adapters: %v\n", err)
		os.Exit(1)
	}

	if len(adapters) == 0 {
		fmt.Fprintf(os.Stdout, "No adapters detected.\n")
		return
	}

	for _, adapter := range adapters {
		powered := "off"
		if adapter.Powered {
			powered = "on"
		}

		alias := adapter.Alias
		if alias == "" {
			alias = "(no alias)"
		}

		fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%s\n", adapter.ID, adapter.Address, alias, powered)
	}
}
