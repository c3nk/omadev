# Omadev

A developer experience CLI designed for Omarchy.

> **Status: early development.** The foundation is in place; the detection engine and
> Docker Compose orchestration are being built milestone by milestone (see
> [`docs/ROADMAP.md`](docs/ROADMAP.md)). Not all commands below work yet.

Omadev understands the development environment a repository already defines and gives you
a consistent way to inspect and operate it. It does not replace Docker, Docker Compose,
mise, your package manager, or your project's own scripts — it delegates to them.

> Omadev doesn't define your development environment. It understands it.

## What it does

The goal is a simple entry point into an unfamiliar repository:

```bash
git clone <repository>
cd <repository>
omadev inspect   # understand the stack, read-only
omadev up        # show the plan, confirm, start, verify
omadev status    # see service state and ports
omadev logs      # tail development logs
omadev down      # stop, preserving your data
```

Omadev is transparent and safe by default. Before it changes anything it shows the exact
command it will run and asks for confirmation:

```
Execution strategy:
  Docker Compose

Command:
  docker compose up -d

Continue? [Y/n]
```

## Principles

- **Understand, don't reinvent.** Use the environment the repository already defines.
- **Detection before configuration.** Works with zero Omadev-specific config.
- **Safe and non-destructive.** No source changes, no generated files, no deleted volumes,
  no secrets printed, no privilege or security settings modified.
- **Report uncertainty.** When detection is ambiguous, Omadev says so instead of guessing.

## Build from source

Requires Go 1.24 or newer.

```bash
git clone https://github.com/c3nk/omadev.git
cd omadev
make build
./omadev --help
```

Prebuilt binaries and an AUR package are planned for the first public release (v0.1.0).

## Documentation

- [`docs/SPEC.md`](docs/SPEC.md) — product specification
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — internal design
- [`docs/SECURITY.md`](docs/SECURITY.md) — threat model and guarantees
- [`docs/ROADMAP.md`](docs/ROADMAP.md) — milestones
- [`docs/DECISIONS.md`](docs/DECISIONS.md) — decision records

## Relationship to Omarchy

Omadev is an independent tool designed to work well on Omarchy. It is not an official
Omarchy product and does not imply affiliation.

## License

[MIT](LICENSE)
