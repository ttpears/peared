# Contributing to Peared

Thank you for your interest in improving Peared! The project is still early in
its roadmap, so contributions that improve tooling, documentation, or core
functionality are especially appreciated.

## Getting Started
- Install Go 1.22 or newer.
- Clone the repository and run `go test ./...` to verify your environment.
- Review the [architecture](docs/ARCHITECTURE.md) and [roadmap](docs/ROADMAP.md)
  documents to understand the current priorities.

## Developer Certificate of Origin (DCO)

Peared uses the [Developer Certificate of Origin](https://developercertificate.org)
for all contributions. When opening a pull request, make sure each commit is
signed off by including a `Signed-off-by` trailer:

```
Signed-off-by: Your Name <you@example.com>
```

You can add this automatically with:

```
git commit -s
```

Commits without a valid sign-off will need to be amended before they can be
merged.

## Testing & CI

- Run `go test ./...` before submitting a change.
- Keep `golangci-lint` happy once the linting workflow lands; lint clean merges
  faster.
- Include unit tests whenever practical so we can confidently evolve the codebase.

## Communication

- File issues for bugs or feature requests.
- Draft pull requests are welcome for early feedback.
- Be kind and respectful when interacting with other contributors—we're all here
  to build a better Bluetooth experience.
