# Repository Guidelines

## Project Structure & Module Organization
- `cmd/pearedd` runs the daemon; `cmd/peared` delivers the CLI and interactive shell.
- Core logic lives in `internal/` packages (`bluetoothctl`, `cli`, `config`, `daemon`); each keeps tests in matching `_test.go` files.
- `docs/ARCHITECTURE.md` and `docs/ROADMAP.md` anchor long-term plans—sync large proposals with them.
- Starter config sits in `config/examples/minimal.yaml`; copy it into `$XDG_CONFIG_HOME/peared/` for local runs.
- Packaging experiments live in `packaging/arch`; touch them only when working on distro deliverables.

## Build, Test, and Development Commands
- `go test ./...` exercises every package; run it before commits and again prior to reviews.
- `go run ./cmd/pearedd --log-level debug` starts the daemon with verbose tracing for adapter discovery.
- `go run ./cmd/pearedd --config ./config/examples/minimal.yaml` checks config loading without polluting user directories.
- `go run ./cmd/peared adapters list` and `go run ./cmd/peared devices scan` cover the most common CLI flows.
- From `packaging/arch`, `makepkg -si` builds and installs the Arch development package for end-to-end smoke tests.

## Coding Style & Naming Conventions
- Use `gofmt`/`goimports`; keep tabs, grouped imports, and one statement per line.
- Employ clear CamelCase for exported structs and verbs for commands; keep helpers package-private unless the API requires exposure.
- Propagate `context.Context` explicitly and prefer dependency injection over globals.

## Testing Guidelines
- Mirror the existing table-driven approach in `internal/bluetoothctl/runner_test.go` when adding scenarios.
- Isolate external calls with fakes or interfaces; tests should not shell out to `bluetoothctl`.
- Document coverage gaps with TODOs tied to issues and highlight any manual validation in your PR notes.

## Commit & Pull Request Guidelines
- Write imperative, sentence-case subjects (e.g. `Fix tests to avoid passing nil contexts`) and include `Signed-off-by:` via `git commit -s`.
- Keep each commit focused; split refactors from feature work to ease review.
- Pull requests need a short problem statement, summary of changes, and proof of testing (`go test ./...`, manual steps, screenshots if necessary).
- Link related roadmap items or issues so maintainers can track progress.

## Configuration & Security Tips
- Do not commit real adapter IDs, secrets, or `/sys` output; scrub artifacts before pushing.
- When elevated privileges are required (`sudo bluetoothctl`), note why in the PR so reviewers can reproduce safely.
