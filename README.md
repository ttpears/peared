# Peared

Peared is an in-progress Bluetooth management toolkit for Linux systems with an
initial focus on Arch Linux while targeting compatibility with other
distributions such as Debian. The goal is to replace brittle scripts and tray
applets with a dependable daemon + CLI combination that understands multiple
controllers, audio routing, automation, and status bar integrations.

## Project Status
This repository currently contains planning documentation only. Refer to the
[architecture](docs/ARCHITECTURE.md) and [roadmap](docs/ROADMAP.md) documents for
up-to-date implementation plans. The initial Go module and service code will be
added once the bootstrap milestone is complete.

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

