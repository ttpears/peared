package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestShellRunExit(t *testing.T) {
	input := strings.NewReader("help\nexit\n")
	var output bytes.Buffer

	shell := NewShell(input, &output)
	if err := shell.Run(context.Background()); err != nil {
		t.Fatalf("shell.Run returned error: %v", err)
	}

	out := output.String()
	if !strings.Contains(out, "Welcome to the Peared shell") {
		t.Fatalf("expected welcome message, got: %q", out)
	}
	if !strings.Contains(out, "Available commands") {
		t.Fatalf("expected help output, got: %q", out)
	}
	if !strings.Contains(out, "Goodbye!") {
		t.Fatalf("expected goodbye message, got: %q", out)
	}
}

func TestShellRunUnknownCommand(t *testing.T) {
	input := strings.NewReader("foo\nexit\n")
	var output bytes.Buffer

	shell := NewShell(input, &output)
	if err := shell.Run(context.Background()); err != nil {
		t.Fatalf("shell.Run returned error: %v", err)
	}

	out := output.String()
	if !strings.Contains(out, "Unknown command: foo") {
		t.Fatalf("expected unknown command message, got: %q", out)
	}
}

func TestShellRunContextCancel(t *testing.T) {
	r, w := io.Pipe()
	var output bytes.Buffer

	shell := NewShell(r, &output)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		done <- shell.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shell did not return after context cancellation")
	}

	_ = w.Close()
}

func TestShellRunTwice(t *testing.T) {
	input := strings.NewReader("exit\n")
	var output bytes.Buffer

	shell := NewShell(input, &output)
	if err := shell.Run(context.Background()); err != nil {
		t.Fatalf("first run returned error: %v", err)
	}

	if err := shell.Run(context.Background()); err == nil {
		t.Fatal("expected error on second run but got nil")
	}
}
