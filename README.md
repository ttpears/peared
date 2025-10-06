# Peared

Peared is an in-progress Bluetooth management toolkit for Linux systems with an
initial focus on Arch Linux while targeting compatibility with other
distributions such as Debian. The goal is to replace brittle scripts and tray
applets with a dependable daemon + CLI combination that understands multiple
controllers, audio routing, automation, and status bar integrations.

## Project Status
This repository now includes the first pass at the Go module that will back the
project. The `pearedd` daemon currently starts, discovers Bluetooth adapters via
sysfs, selects an active controller, waits for a cancellation signal, and shuts
down cleanly—providing the scaffolding required for future Bluetooth management
features. Refer to the [architecture](docs/ARCHITECTURE.md) and
[roadmap](docs/ROADMAP.md) documents for the broader implementation plan.

```bash
go test ./...
go run ./cmd/pearedd --log-level debug
go run ./cmd/pearedd --config /path/to/config.yaml
go run ./cmd/peared shell
go run ./cmd/peared adapters list
```

The daemon exits when it receives `SIGINT`/`SIGTERM` or when the provided
context is cancelled. It now consumes configuration from the standard XDG
location (`$XDG_CONFIG_HOME/peared/config.yaml`) or a path supplied via
`--config`. Adapter discovery happens automatically on startup so the daemon can
select a controller without requiring you to manually gather MAC addresses.

The companion CLI ships with an early interactive shell so you can validate
that the binary launches and cleanly exits on your workstation. Type `help`
inside the shell to see the available commands and use `exit` when you're
finished testing. The CLI also exposes `peared adapters list` to query detected
controllers and surface their IDs, addresses, and power state. Inspecting the
sysfs hierarchy usually works without additional setup, but some distributions
restrict access to `/sys/class/bluetooth`. If you encounter a permission error,
run the command with elevated privileges or add your user to the `bluetooth`
group so discovery can proceed.

Copy `config/examples/minimal.yaml` into your configuration directory to get
started. You can optionally set `preferred_adapter` in the file using the values
reported by `peared adapters list` to persist a controller choice across
restarts.

### Arch Linux packaging

Early Arch Linux packaging bits live under `packaging/arch/PKGBUILD`. The
PKGBUILD follows the usual `-git` convention so maintainers can build the latest
commit without waiting for formal releases:

```bash
cd packaging/arch
makepkg -si
```

The resulting package installs both the `pearedd` daemon and the `peared` CLI
into `/usr/bin`.

## Configuration Hygiene
Configuration templates and examples will always use placeholder adapter IDs and
MAC addresses. Real user data stays in XDG config/runtime directories and is not
committed to the repository.

## UI Integrations
Wayland and X11 users alike should be able to surface Bluetooth state in their
preferred status bars. Planned integrations include Waybar, Polybar, and other
popular bars through the daemon's public API.

## License
The project is licensed under the [GNU General Public License v3.0](LICENSE).

