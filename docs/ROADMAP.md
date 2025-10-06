# Peared Roadmap

This roadmap captures the initial backlog for delivering a reliable Bluetooth
manager. Items are grouped by milestone to provide a guided development path.

## Milestone 0 – Project Bootstrap
- [ ] Publish repository metadata (README, architecture docs, contribution
      guidelines).
- [ ] Establish Go module layout and CI scaffolding.
- [ ] Provide configuration templates with placeholder addresses and adapter IDs.
- [ ] Adopt the GPL-3.0 license and document contributor guidance.

## Milestone 1 – Core Connectivity
- [ ] Implement adapter discovery and selection across multiple controllers.
- [ ] Surface radio block state (via BlueZ or rfkill) with clear error messaging
      and controlled unblocking.
- [ ] Provide reset workflows for bluetoothd (systemd), kernel modules, and
      BlueZ service restarts with retry semantics.
- [ ] Implement device connect/disconnect/trust flows with retry policies.
- [ ] Ensure PipeWire sink/source switching when connecting audio devices and
      fall back to ALSA where necessary.

## Milestone 2 – User Interface Layer
- [ ] Deliver a CLI client for the daemon (connect, disconnect, scan, status,
      reset, forget).
- [ ] Integrate desktop notifications via D-Bus (e.g., `notify-send`).
- [ ] Ship optional status bar outputs compatible with Waybar, Polybar, and
      similar projects without depending on any single compositor.

## Milestone 3 – Automation & Rules
- [ ] Create a rule engine for time-based, manual trigger, and proximity-based
      actions.
- [ ] Support per-device priorities and automatic reconnection policies.
- [ ] Offer hooks for custom scripts upon connection/disconnection events.

## Milestone 4 – Advanced Audio Controls
- [ ] Device-specific volume presets and smoothing curves.
- [ ] Microphone/monitor selection with failover detection.
- [ ] PipeWire graph inspection to ensure the selected profile matches device
      capabilities (HSP/HFP/A2DP).

## Milestone 5 – Extensibility & Distribution
- [ ] Document an extension API for community plugins.
- [ ] Package for Arch Linux (AUR), Debian packages, Flatpak, and generic
      tarballs.
- [ ] Provide migration tooling for existing bluetoothctl scripts.

## Non-Goals (for now)
- Managing mobile platforms or Windows/macOS.
- Providing a full graphical desktop environment applet (focus remains on CLI +
  optional tray bridges).

