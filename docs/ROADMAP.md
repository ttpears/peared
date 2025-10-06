# Peared Roadmap

This roadmap captures the initial backlog for delivering a reliable Bluetooth
manager. Items are grouped by milestone to provide a guided development path.
Each milestone now includes explicit deliverables, success criteria, and
cross-cutting considerations so contributors can quickly understand what "done"
looks like and how work items fit together.

## Milestone 0 – Project Bootstrap

**Objective:** Set up the project so new contributors can clone the repository
and run a minimal build/test loop without touching production systems.

**Deliverables**
- [ ] Publish repository metadata (README, architecture docs, contribution
      guidelines) with clear setup instructions for Arch and Debian developers.
- [ ] Establish Go module layout and CI scaffolding (linting, unit tests, and
      static analysis gates).
- [ ] Provide configuration templates with placeholder addresses and adapter IDs
      stored under `config/examples/`.
- [ ] Adopt the GPL-3.0 license and document contributor guidance, including the
      Developer Certificate of Origin (DCO) process.

**Success Criteria**
- [ ] `go test ./...` and `golangci-lint run` execute in CI on every pull
      request.
- [ ] A new developer can follow documented steps to build the project in under
      15 minutes on a fresh workstation.

## Milestone 1 – Core Connectivity

**Objective:** Provide a resilient daemon capable of discovering adapters and
maintaining device connectivity across restarts or system hiccups.

**Deliverables**
- [ ] Implement adapter discovery and selection across multiple controllers,
      exposing the active adapter through the daemon API.
- [ ] Surface radio block state (via BlueZ or rfkill) with clear error messaging
      and controlled unblocking, including audit logs of unblock attempts.
- [ ] Provide reset workflows for bluetoothd (systemd), kernel modules, and
      BlueZ service restarts with retry semantics and exponential backoff.
- [ ] Implement device connect/disconnect/trust flows with retry policies and
      persistence of known devices.
- [ ] Ensure PipeWire sink/source switching when connecting audio devices and
      fall back to ALSA where necessary, logging the chosen route.

**Success Criteria**
- [ ] Automated integration tests simulate adapter removal/insertion and verify
      stable reconnection.
- [ ] CLI status output reflects accurate adapter and device state within
      1 second of change events.

## Milestone 2 – User Interface Layer

**Objective:** Ship the first user-facing tools that interact with the daemon
and surface state changes in a desktop environment.

**Deliverables**
- [ ] Deliver a CLI client for the daemon (connect, disconnect, scan, status,
      reset, forget) with contextual help output.
- [ ] Integrate desktop notifications via D-Bus (e.g., `notify-send`) with
      user-configurable verbosity levels.
- [ ] Ship optional status bar outputs compatible with Waybar, Polybar, and
      similar projects without depending on any single compositor.

**Success Criteria**
- [ ] CLI commands return structured exit codes and machine-readable output
      (JSON) when requested.
- [ ] Notification system respects quiet hours and suppresses duplicate alerts
      during rapid reconnect attempts.

## Milestone 3 – Automation & Rules

**Objective:** Allow the daemon to respond to context and user-defined policies
without manual intervention.

**Deliverables**
- [ ] Create a rule engine for time-based, manual trigger, and proximity-based
      actions with YAML/TOML configuration support.
- [ ] Support per-device priorities and automatic reconnection policies that
      interact with multiple adapters.
- [ ] Offer hooks for custom scripts upon connection/disconnection events with
      sandboxing guidance.

**Success Criteria**
- [ ] Rule evaluations are observable via logs and optionally metrics so users
      can audit automation behavior.
- [ ] Integration tests cover at least one scenario for each trigger type.

## Milestone 4 – Advanced Audio Controls

**Objective:** Provide audio experiences tailored to device capabilities while
handling edge cases that typically require manual intervention.

**Deliverables**
- [ ] Device-specific volume presets and smoothing curves persisted per profile.
- [ ] Microphone/monitor selection with failover detection and user prompts when
      intervention is required.
- [ ] PipeWire graph inspection to ensure the selected profile matches device
      capabilities (HSP/HFP/A2DP) with fallback heuristics when mismatches occur.

**Success Criteria**
- [ ] Automated checks validate that profile switching restores the previous
      volume curves after reconnect.
- [ ] User feedback loop (CLI or notification) confirms when the system falls
      back to ALSA and why.

## Milestone 5 – Extensibility & Distribution

**Objective:** Ensure the project can be extended by the community and installed
across popular Linux distributions with minimal friction.

**Deliverables**
- [ ] Document an extension API for community plugins with versioning and
      stability guarantees.
- [ ] Package for Arch Linux (AUR), Debian packages, Flatpak, and generic
      tarballs with reproducible build instructions.
- [ ] Provide migration tooling for existing bluetoothctl scripts, including
      documentation on equivalent commands.

**Success Criteria**
- [ ] Binary releases are produced automatically on tagged commits and pass
      smoke tests in containerized environments.
- [ ] At least two community extensions are validated against the API contract
      before stabilizing it.

## Non-Goals (for now)
- Managing mobile platforms or Windows/macOS.
- Providing a full graphical desktop environment applet (focus remains on CLI +
  optional tray bridges).

