# Peared Bluetooth Manager Architecture

## Vision
Peared is a Go-based Bluetooth orchestration service designed for Linux (tested
primarily on Arch but aiming for distro-agnostic behavior) that offers
dependable management of audio devices, controllers, system integrations such as
PipeWire, and optional status bar bridges. The project aims to provide a
resilient alternative to brittle tray applets by combining command-line
ergonomics with optional UI integrations.

## Guiding Principles
- **Configuration hygiene:** Keep real adapter addresses and secrets out of the
  repository while providing clear template data for users to copy.
- **Deterministic operations:** Prefer idempotent commands and robust retry logic
  to guarantee predictable outcomes.
- **Observable system:** Expose structured logs and metrics to make debugging
  reliable without temporary throwaway files.
- **Extensible rules engine:** Build the core in a modular fashion so proximity
  triggers and automation rules can be layered on later without refactoring the
  base.

## High-Level Components
| Component | Responsibility |
|-----------|----------------|
| **Command Layer** | CLI commands (and optional TUI/UI adapters) that expose features like connect, disconnect, scan, reset, and status inspection. |
| **Controller Manager** | Enumerates adapters, tracks their capabilities, toggles power state, and handles module reload/unload workflows. |
| **Device Orchestrator** | Performs pairing, trusting, connection prioritisation, and audio routing; holds retry policies. |
| **Audio Integration** | Bridges to PipeWire/ALSA to switch profiles, default sinks/sources, and device-specific volume curves. |
| **Automation & Rules** | Evaluates user-defined triggers (time, proximity, host state) and executes actions; pluggable backend for future expansions. |
| **State Persistence** | Stores known devices, preferred adapters, and automation rules in a user config directory using a human-readable format (YAML/TOML). |
| **Notification & Status Layer** | Sends desktop notifications, surfaces metrics/logs, and feeds data to status bars such as Waybar or Polybar without binding to any single compositor. |

## Process Model
1. **Daemon** (long-running service) maintains adapter/device state, listens for
   system events (D-Bus, PipeWire, radio block state), and exposes a local API
   (gRPC/REST/DBus).
2. **CLI Client** communicates with the daemon for user commands. This keeps
   operations fast and allows future GUI integrations.

## Integration Points
- **D-Bus API:** For bluetoothd, NetworkManager (if needed), and PipeWire.
- **Systemd:** Restarting bluetooth, loading kernel modules, and managing user
  services (e.g., PipeWire). Prefer `systemd-run --user` where appropriate.
- **Radio control:** Query and adjust adapter block state through BlueZ or
  rfkill tooling when necessary, keeping changes auditable.
- **Status bars:** JSON socket or file outputs that can be consumed by Waybar,
  Polybar, and similar projects.

## Configuration Strategy
- Use `$XDG_CONFIG_HOME/peared/config.yaml` for user-visible settings.
- Store secrets (PINs, passkeys) in `$XDG_RUNTIME_DIR/peared/` with strict
  permissions.
- Provide sample configuration templates with placeholder addresses only.

## Error Handling & Observability
- Structured logs (logfmt or JSON) with log levels.
- Rich error wrapping using Go 1.20+ `errors` features.
- Optional tracing hooks for future integration with OpenTelemetry.

## Testing Strategy
- Unit tests for module logic with mocked system interfaces.
- Integration tests executed inside containers with BlueZ running in simulation
  mode.
- Contract tests for the CLI/daemon API boundary.

## Roadmap Snapshot
The roadmap is maintained in [`docs/ROADMAP.md`](ROADMAP.md).

