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

	"github.com/peared/peared/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "shell":
		runShell(os.Args[2:])
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
	fmt.Fprintf(os.Stderr, "  shell   Start an interactive shell session\n")
	fmt.Fprintf(os.Stderr, "  help    Show this message\n")
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
