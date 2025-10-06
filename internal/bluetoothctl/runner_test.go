package bluetoothctl

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunnerScanUsesSudoWhenConfigured(t *testing.T) {
	ctx := context.Background()
	var gotName string
	var gotArgs []string
	runner, err := NewRunner(
		WithBinary("bluetoothctl"),
		WithSudoPath("sudo"),
		WithUseSudo(true),
		WithCommandRunner(func(_ context.Context, name string, args ...string) ([]byte, error) {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return []byte("scan output\n"), nil
		}),
	)
	if err != nil {
		t.Fatalf("NewRunner returned error: %v", err)
	}

	out, err := runner.Scan(ctx, 3*time.Second)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if out != "scan output" {
		t.Fatalf("unexpected output: %q", out)
	}

	if gotName != "sudo" {
		t.Fatalf("expected sudo to be used, got %q", gotName)
	}

	wantArgs := []string{"bluetoothctl", "--timeout", "3", "scan", "on"}
	if len(gotArgs) != len(wantArgs) {
		t.Fatalf("expected %d args, got %d", len(wantArgs), len(gotArgs))
	}
	for i := range wantArgs {
		if gotArgs[i] != wantArgs[i] {
			t.Fatalf("arg %d mismatch: want %q got %q", i, wantArgs[i], gotArgs[i])
		}
	}
}

func TestRunnerPairReturnsCommandErrorOnFailure(t *testing.T) {
	runner, err := NewRunner(
		WithBinary("bluetoothctl"),
		WithUseSudo(false),
		WithCommandRunner(func(_ context.Context, name string, args ...string) ([]byte, error) {
			if name != "bluetoothctl" {
				t.Fatalf("expected bluetoothctl executable, got %q", name)
			}
			return []byte("Failed to pair"), errors.New("exit status 1")
		}),
	)
	if err != nil {
		t.Fatalf("NewRunner returned error: %v", err)
	}

	_, err = runner.Pair(context.Background(), "AA:BB:CC:DD:EE:FF")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	cmdErr := &CommandError{}
	if !errors.As(err, &cmdErr) {
		t.Fatalf("expected CommandError, got %T", err)
	}

	if cmdErr.Output != "Failed to pair" {
		t.Fatalf("unexpected output: %q", cmdErr.Output)
	}
}

func TestRunnerConnectValidatesInput(t *testing.T) {
	runner, err := NewRunner(
		WithBinary("bluetoothctl"),
		WithUseSudo(false),
	)
	if err != nil {
		t.Fatalf("NewRunner returned error: %v", err)
	}

	if _, err := runner.Connect(context.Background(), ""); err == nil {
		t.Fatalf("expected error for empty address")
	}
}

func TestRunnerDisconnectNilContext(t *testing.T) {
	runner, err := NewRunner(
		WithBinary("bluetoothctl"),
		WithUseSudo(false),
	)
	if err != nil {
		t.Fatalf("NewRunner returned error: %v", err)
	}

	var nilCtx context.Context
	if _, err := runner.Disconnect(nilCtx, "AA:BB"); err == nil {
		t.Fatalf("expected error for nil context")
	}
}

func TestRunnerScanDefaultsDuration(t *testing.T) {
	ctx := context.Background()
	var gotArgs []string
	runner, err := NewRunner(
		WithBinary("bluetoothctl"),
		WithUseSudo(false),
		WithCommandRunner(func(_ context.Context, name string, args ...string) ([]byte, error) {
			if name != "bluetoothctl" {
				t.Fatalf("unexpected executable %q", name)
			}
			gotArgs = append([]string(nil), args...)
			return []byte("ok"), nil
		}),
	)
	if err != nil {
		t.Fatalf("NewRunner returned error: %v", err)
	}

	if _, err := runner.Scan(ctx, 0); err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if len(gotArgs) < 4 {
		t.Fatalf("expected at least 4 arguments, got %v", gotArgs)
	}

	if gotArgs[0] != "--timeout" {
		t.Fatalf("expected timeout flag, got %q", gotArgs[0])
	}

	if gotArgs[2] != "scan" || gotArgs[3] != "on" {
		t.Fatalf("unexpected scan arguments: %v", gotArgs)
	}
}

func TestNewRunnerRequiresSudoWhenNonRoot(t *testing.T) {
	originalGeteuid := geteuid
	defer func() { geteuid = originalGeteuid }()

	geteuid = func() int {
		return 1000
	}

	_, err := NewRunner(
		WithBinary("bluetoothctl"),
		WithSudoPath(""),
	)
	if err == nil {
		t.Fatalf("expected error when sudo missing for non-root user")
	}

	if !strings.Contains(err.Error(), "sudo binary not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
