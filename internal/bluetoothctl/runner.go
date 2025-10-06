package bluetoothctl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var geteuid = os.Geteuid

type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Runner executes bluetoothctl commands while handling privilege escalation when
// necessary. The CLI relies on it to provide a thin wrapper around common
// pairing and connection operations without forcing users to drop into the
// interactive bluetoothctl shell.
type Runner struct {
	// Binary is the bluetoothctl executable path.
	Binary string

	// SudoPath is the sudo executable path used for privilege escalation.
	SudoPath string

	// UseSudo indicates whether commands should be executed through sudo.
	UseSudo bool

	// Adapter identifies the adapter bluetoothctl commands should target when
	// multiple controllers are present. Leave empty to defer to bluetoothctl's
	// default selection.
	Adapter string

	useSudoSet bool
	sudoSet    bool

	run commandRunner
}

// RunnerOption customises Runner construction.
type RunnerOption func(*Runner)

// WithBinary sets the bluetoothctl binary path used by the runner.
func WithBinary(path string) RunnerOption {
	return func(r *Runner) {
		r.Binary = path
	}
}

// WithSudoPath overrides the sudo binary path used for escalation.
func WithSudoPath(path string) RunnerOption {
	return func(r *Runner) {
		r.SudoPath = path
		r.sudoSet = true
	}
}

// WithUseSudo forces whether sudo should be used when invoking bluetoothctl.
func WithUseSudo(use bool) RunnerOption {
	return func(r *Runner) {
		r.UseSudo = use
		r.useSudoSet = true
	}
}

// WithAdapter configures the adapter bluetoothctl commands should target.
func WithAdapter(adapter string) RunnerOption {
	return func(r *Runner) {
		r.Adapter = strings.TrimSpace(adapter)
	}
}

// WithCommandRunner allows tests to replace the command execution primitive.
func WithCommandRunner(run commandRunner) RunnerOption {
	return func(r *Runner) {
		r.run = run
	}
}

// NewRunner constructs a Runner configured to execute bluetoothctl commands.
// When no overrides are provided the runner automatically discovers the
// bluetoothctl and sudo binaries on PATH. Non-root users default to executing
// commands through sudo so operations that require elevated privileges succeed
// without additional flags.
func NewRunner(opts ...RunnerOption) (*Runner, error) {
	r := &Runner{}

	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}

	if r.Binary == "" {
		path, err := exec.LookPath("bluetoothctl")
		if err != nil {
			return nil, fmt.Errorf("locate bluetoothctl: %w", err)
		}
		r.Binary = path
	}

	if r.SudoPath == "" && !r.sudoSet {
		if path, err := exec.LookPath("sudo"); err == nil {
			r.SudoPath = path
		}
	}

	if !r.useSudoSet {
		if geteuid() != 0 {
			r.UseSudo = true
		}
	}

	if r.UseSudo && r.SudoPath == "" {
		return nil, errors.New("sudo binary not found while privilege escalation is required; run the command as root or install sudo")
	}

	if r.run == nil {
		r.run = defaultCommandRunner
	}

	return r, nil
}

// Scan enables adapter discovery for the provided duration and returns the raw
// bluetoothctl output. A zero or negative duration falls back to a 15 second
// scan window.
func (r *Runner) Scan(ctx context.Context, duration time.Duration) (string, error) {
	if ctx == nil {
		return "", errors.New("nil context passed to Scan")
	}

	if duration <= 0 {
		duration = 15 * time.Second
	}

	secs := int(duration / time.Second)
	if secs <= 0 {
		secs = 1
	}

	args := []string{"--timeout", fmt.Sprintf("%d", secs), "scan", "on"}
	output, err := r.exec(ctx, args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// Pair attempts to pair with the provided device address and returns the raw
// bluetoothctl output.
func (r *Runner) Pair(ctx context.Context, address string) (string, error) {
	return r.simpleDeviceCommand(ctx, "pair", address)
}

// Connect attempts to connect to the provided device address and returns the
// raw bluetoothctl output.
func (r *Runner) Connect(ctx context.Context, address string) (string, error) {
	return r.simpleDeviceCommand(ctx, "connect", address)
}

// Disconnect attempts to disconnect from the provided device address and
// returns the raw bluetoothctl output.
func (r *Runner) Disconnect(ctx context.Context, address string) (string, error) {
	return r.simpleDeviceCommand(ctx, "disconnect", address)
}

func (r *Runner) simpleDeviceCommand(ctx context.Context, command, address string) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("nil context passed to %s", command)
	}

	addr := strings.TrimSpace(address)
	if addr == "" {
		return "", fmt.Errorf("device address required for %s", command)
	}

	args := []string{command, addr}
	output, err := r.exec(ctx, args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func (r *Runner) exec(ctx context.Context, args ...string) (string, error) {
	if ctx == nil {
		return "", errors.New("nil context passed to exec")
	}

	var name string
	var finalArgs []string

	if r.Adapter != "" {
		finalArgs = append(finalArgs, "--adapter", r.Adapter)
	}

	finalArgs = append(finalArgs, args...)
	if r.UseSudo {
		name = r.SudoPath
		finalArgs = append([]string{r.Binary}, finalArgs...)
	} else {
		name = r.Binary
		// finalArgs already contains the command arguments.
	}

	out, err := r.run(ctx, name, finalArgs...)
	if err != nil {
		return "", &CommandError{Args: args, Output: string(out), Err: err}
	}

	return string(out), nil
}

func defaultCommandRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// CommandError captures failures when invoking bluetoothctl. Callers can inspect
// the output to provide helpful user guidance.
type CommandError struct {
	Args   []string
	Output string
	Err    error
}

// Error implements error.
func (e *CommandError) Error() string {
	return fmt.Sprintf("bluetoothctl %s failed: %v", strings.Join(e.Args, " "), e.Err)
}

// Unwrap allows errors.Is / errors.As to inspect the root cause.
func (e *CommandError) Unwrap() error {
	return e.Err
}
