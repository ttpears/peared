package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
)

// Shell provides a minimal interactive prompt for controlling the daemon. It is
// intentionally small for now so we can iterate quickly while we flesh out the
// underlying control plane.
type Shell struct {
	reader *bufio.Reader
	writer io.Writer
	prompt string

	closed atomic.Bool
}

// ShellOption allows customization of the shell when constructed.
type ShellOption func(*Shell)

// WithPrompt customizes the shell prompt string.
func WithPrompt(prompt string) ShellOption {
	return func(s *Shell) {
		s.prompt = prompt
	}
}

// NewShell creates a Shell instance that reads commands from r and writes
// output to w. Callers may provide additional options to tweak defaults.
func NewShell(r io.Reader, w io.Writer, opts ...ShellOption) *Shell {
	shell := &Shell{
		reader: bufio.NewReader(r),
		writer: w,
		prompt: "peared> ",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(shell)
		}
	}

	return shell
}

// Run executes the interactive loop until the user exits or the context is
// cancelled. Unknown commands result in a helpful message so early adopters can
// understand the current surface area.
func (s *Shell) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("nil context passed to Shell.Run")
	}

	if s.closed.Load() {
		return errors.New("shell already closed")
	}

	fmt.Fprintf(s.writer, "Welcome to the Peared shell! Type 'help' to see available commands.\n")

	type input struct {
		line string
		err  error
	}

	inputCh := make(chan input, 1)
	go func() {
		defer close(inputCh)
		for {
			line, err := s.reader.ReadString('\n')
			select {
			case <-ctx.Done():
				return
			case inputCh <- input{line: line, err: err}:
			}

			if err != nil {
				return
			}
		}
	}()

	for {
		fmt.Fprint(s.writer, s.prompt)

		select {
		case <-ctx.Done():
			s.closed.Store(true)
			if cause := context.Cause(ctx); cause != nil {
				return cause
			}
			return context.Canceled
		case in, ok := <-inputCh:
			if !ok {
				fmt.Fprintln(s.writer)
				s.closed.Store(true)
				return nil
			}

			if in.err != nil {
				if errors.Is(in.err, io.EOF) {
					fmt.Fprintln(s.writer)
					s.closed.Store(true)
					return nil
				}
				s.closed.Store(true)
				return in.err
			}

			cmd := strings.TrimSpace(in.line)
			switch strings.ToLower(cmd) {
			case "", "#":
				continue
			case "exit", "quit":
				fmt.Fprintln(s.writer, "Goodbye!")
				s.closed.Store(true)
				return nil
			case "help":
				s.writeHelp()
			default:
				fmt.Fprintf(s.writer, "Unknown command: %s\n", cmd)
			}
		}
	}
}

func (s *Shell) writeHelp() {
	fmt.Fprintf(s.writer, "Available commands:\n")
	fmt.Fprintf(s.writer, "  help  - show this message\n")
	fmt.Fprintf(s.writer, "  exit  - leave the shell\n")
	fmt.Fprintf(s.writer, "  quit  - alias for exit\n")
}
