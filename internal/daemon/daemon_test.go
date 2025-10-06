package daemon

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

func TestNewUsesDefaultLogger(t *testing.T) {
	d, err := New(Options{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d.log == nil {
		t.Fatalf("expected daemon to configure a logger")
	}
}

func TestRunRespectsCancellation(t *testing.T) {
	d, err := New(Options{Logger: slog.New(slog.NewTextHandler(testWriter{t}, nil))})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := d.Run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation error, got %v", err)
		}
	}
}

func TestRunPropagatesContextError(t *testing.T) {
	d, err := New(Options{Logger: slog.New(slog.NewTextHandler(testWriter{t}, nil))})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errors.New("boom"))

	err = d.Run(ctx)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

// testWriter implements io.Writer using testing.T logging so that slog output
// is visible when tests run with -v.
type testWriter struct {
	t *testing.T
}

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Helper()
	w.t.Logf("daemon log: %s", string(p))
	return len(p), nil
}

func TestRunNilContext(t *testing.T) {
	d, err := New(Options{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := d.Run(nil); err == nil {
		t.Fatalf("expected error when context is nil")
	}
}

func TestRunBlocksUntilCancelled(t *testing.T) {
	d, err := New(Options{Logger: slog.New(slog.NewTextHandler(testWriter{t}, nil))})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		if err := d.Run(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("daemon returned before cancellation")
	case <-time.After(25 * time.Millisecond):
	}

	cancel()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("daemon did not return after cancellation")
	}
}
