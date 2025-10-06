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
	"sync"
	"syscall"
	"time"

	"github.com/peared/peared/internal/bluetoothctl"
	"github.com/peared/peared/internal/cli"
	"github.com/peared/peared/internal/config"
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
	case "devices":
		runDevices(os.Args[2:])
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
	fmt.Fprintf(os.Stderr, "  devices   Manage Bluetooth devices (scan, pair, connect, disconnect)\n")
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

func runDevices(args []string) {
	if len(args) == 0 {
		devicesUsage()
		os.Exit(2)
	}

	switch args[0] {
	case "scan":
		scanDevices(args[1:])
	case "pair":
		pairDevice(args[1:])
	case "connect":
		connectDevice(args[1:])
	case "disconnect":
		disconnectDevice(args[1:])
	case "help", "-h", "--help":
		devicesUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown devices command: %s\n\n", args[0])
		devicesUsage()
		os.Exit(2)
	}
}

func devicesUsage() {
	fmt.Fprintf(os.Stderr, "Usage: peared devices <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  scan         Discover nearby devices using bluetoothctl\n")
	fmt.Fprintf(os.Stderr, "  pair <addr>  Pair with the specified device\n")
	fmt.Fprintf(os.Stderr, "  connect <addr>   Connect to the specified device\n")
	fmt.Fprintf(os.Stderr, "  disconnect <addr> Disconnect the specified device\n")
}

func scanDevices(args []string) {
	flagSet := flag.NewFlagSet("devices scan", flag.ExitOnError)
	duration := flagSet.Duration("duration", 15*time.Second, "Duration to scan for devices")
	noSudo := flagSet.Bool("no-sudo", false, "Disable automatic sudo escalation (advanced)")
	adapter := flagSet.String("adapter", "", "Adapter identifier (ID, address, or alias) to target")
	configPath := flagSet.String("config", "", "Path to configuration file (defaults to XDG config directory)")
	if err := flagSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse devices flags: %v\n", err)
		os.Exit(2)
	}

	runner, selectedAdapter, err := newBluetoothRunner(*noSudo, *adapter, *configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up bluetoothctl runner: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	adapterDescription := "the default adapter"
	if selectedAdapter != "" {
		adapterDescription = fmt.Sprintf("adapter %s", selectedAdapter)
	}

	fmt.Fprintf(os.Stderr, "Scanning for devices using %s for up to %s...\n", adapterDescription, formatDuration(*duration))
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to cancel.\n")
	fmt.Fprintf(os.Stderr, "Progress: ")

	start := time.Now()
	progressDone := make(chan struct{})
	var progressWG sync.WaitGroup
	progressWG.Add(1)
	go func() {
		defer progressWG.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-progressDone:
				return
			case <-ticker.C:
				fmt.Fprint(os.Stderr, ".")
			}
		}
	}()

	output, err := runner.Scan(ctx, *duration)
	close(progressDone)
	progressWG.Wait()
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Scan finished after %s.\n", time.Since(start).Round(time.Second))

	if err != nil {
		handleDeviceCommandError("scan", err)
		os.Exit(1)
	}

	if output != "" {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
}

func pairDevice(args []string) {
	flagSet := flag.NewFlagSet("devices pair", flag.ExitOnError)
	noSudo := flagSet.Bool("no-sudo", false, "Disable automatic sudo escalation (advanced)")
	adapter := flagSet.String("adapter", "", "Adapter identifier (ID, address, or alias) to target")
	configPath := flagSet.String("config", "", "Path to configuration file (defaults to XDG config directory)")
	if err := flagSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse devices flags: %v\n", err)
		os.Exit(2)
	}

	if flagSet.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "pair requires a device address\n")
		os.Exit(2)
	}

	address := flagSet.Arg(0)
	runner, _, err := newBluetoothRunner(*noSudo, *adapter, *configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up bluetoothctl runner: %v\n", err)
		os.Exit(1)
	}

	output, err := runner.Pair(context.Background(), address)
	if err != nil {
		handleDeviceCommandError(fmt.Sprintf("pair %s", address), err)
		os.Exit(1)
	}

	if output != "" {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
}

func connectDevice(args []string) {
	flagSet := flag.NewFlagSet("devices connect", flag.ExitOnError)
	noSudo := flagSet.Bool("no-sudo", false, "Disable automatic sudo escalation (advanced)")
	adapter := flagSet.String("adapter", "", "Adapter identifier (ID, address, or alias) to target")
	configPath := flagSet.String("config", "", "Path to configuration file (defaults to XDG config directory)")
	if err := flagSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse devices flags: %v\n", err)
		os.Exit(2)
	}

	if flagSet.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "connect requires a device address\n")
		os.Exit(2)
	}

	address := flagSet.Arg(0)
	runner, _, err := newBluetoothRunner(*noSudo, *adapter, *configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up bluetoothctl runner: %v\n", err)
		os.Exit(1)
	}

	output, err := runner.Connect(context.Background(), address)
	if err != nil {
		handleDeviceCommandError(fmt.Sprintf("connect %s", address), err)
		os.Exit(1)
	}

	if output != "" {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
}

func disconnectDevice(args []string) {
	flagSet := flag.NewFlagSet("devices disconnect", flag.ExitOnError)
	noSudo := flagSet.Bool("no-sudo", false, "Disable automatic sudo escalation (advanced)")
	adapter := flagSet.String("adapter", "", "Adapter identifier (ID, address, or alias) to target")
	configPath := flagSet.String("config", "", "Path to configuration file (defaults to XDG config directory)")
	if err := flagSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse devices flags: %v\n", err)
		os.Exit(2)
	}

	if flagSet.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "disconnect requires a device address\n")
		os.Exit(2)
	}

	address := flagSet.Arg(0)
	runner, _, err := newBluetoothRunner(*noSudo, *adapter, *configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up bluetoothctl runner: %v\n", err)
		os.Exit(1)
	}

	output, err := runner.Disconnect(context.Background(), address)
	if err != nil {
		handleDeviceCommandError(fmt.Sprintf("disconnect %s", address), err)
		os.Exit(1)
	}

	if output != "" {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	rounded := d.Round(time.Second)
	if rounded == 0 {
		rounded = time.Second
	}

	return rounded.String()
}

func newBluetoothRunner(disableSudo bool, adapterOverride, configPath string) (*bluetoothctl.Runner, string, error) {
	var opts []bluetoothctl.RunnerOption
	if disableSudo {
		opts = append(opts, bluetoothctl.WithUseSudo(false))
	}

	adapter, err := determineAdapter(context.Background(), adapterOverride, configPath)
	if err != nil {
		return nil, "", fmt.Errorf("determine adapter: %w", err)
	}

	if adapter != "" {
		opts = append(opts, bluetoothctl.WithAdapter(adapter))
	}

	runner, err := bluetoothctl.NewRunner(opts...)
	if err != nil {
		return nil, "", err
	}

	return runner, adapter, nil
}

func determineAdapter(ctx context.Context, override, configPath string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return override, nil
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}

	provider := daemon.DefaultAdapterProvider()
	adapters, err := provider.ListAdapters(ctx)
	if err != nil {
		return "", fmt.Errorf("discover adapters: %w", err)
	}

	if len(adapters) == 0 {
		return "", errors.New("no adapters detected")
	}

	selected, err := daemon.SelectAdapter(cfg.Daemon.PreferredAdapter, adapters)
	if err != nil {
		return "", err
	}

	return selected.ID, nil
}

func handleDeviceCommandError(operation string, err error) {
	var cmdErr *bluetoothctl.CommandError
	if errors.As(err, &cmdErr) {
		trimmed := strings.TrimSpace(cmdErr.Output)
		if trimmed != "" {
			fmt.Fprintf(os.Stderr, "%s\n", trimmed)
		}
	}
	fmt.Fprintf(os.Stderr, "failed to execute %s: %v\n", operation, err)
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
