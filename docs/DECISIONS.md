# Omadev — Decision Records

> Standalone architecture and product decisions for Omadev v0.1. Each record states the decision
> and the reason. Other documents ([`SPEC.md`](SPEC.md), [`ARCHITECTURE.md`](ARCHITECTURE.md),
> [`SECURITY.md`](SECURITY.md), [`ROADMAP.md`](ROADMAP.md)) reference these by ID.

---

## Product decisions

### D1 — Root `omadev` is informational-only
Bare `omadev` inspects and prints an overview; when the environment is not running it shows the
start command but does not offer to start it. Starting requires an explicit `omadev up`. Keeps the
safety lifecycle intact and avoids a side-effect entry point on the default command.

### D2 — `.omadev.yaml` is deferred
No configuration file is parsed in v0.1. The project model keeps an enrichment seam so an override
layer can be added later. No v0.1 acceptance criterion needs config, so it is not built
speculatively. The happy path must work with zero configuration.

### D3 — Hybrid Compose handling
`inspect` parses Compose YAML itself with a safe parser that does not resolve `.env`/secret
interpolation and does not require Docker. Execution (`up`/`down`/`logs`/`status`) delegates to the
real `docker compose` binary rather than reimplementing it. `inspect` runs a lightweight read-only
`docker info` probe only to render the correct command line; it still succeeds when Docker is
absent, annotating a prerequisite note.

### D4 — Docker privilege is detected at runtime
Omadev probes `docker info` without `sudo` first; on success it uses commands without `sudo`, and
on a permission error it reports that elevated access (or a rootless/`docker`-group setup the user
controls) is required. Nothing is hardcoded. Omadev never modifies the `docker` group or Docker
security settings.

### D5 — Project root via upward-only search
Root is found by walking upward from the working directory to the first directory with a
recognized Compose file (or nearest project marker), stopping at the Git root or filesystem root.
No downward search, no filesystem-wide scan. Deterministic and fast.

### D6 — Readiness uses Compose-substantiated states only
States: `running`, `healthy`, `unhealthy`, `stopped`, `unknown`. `healthy`/`unhealthy` appear only
when a Compose healthcheck exists. Container liveness is never presented as application readiness.

### D7 — URLs and ports are evidence-only
Ports and URLs come only from Compose published-port metadata. No assumed role-to-port mappings
(e.g. `5173 = frontend`) without evidence.

### D8 — Exit codes
`0` success, `2` unsupported/unknown project, `3` missing prerequisite, `4` invalid project
configuration, `5` user cancellation, `6` execution failure. `1` is reserved for Cobra's generic
usage errors.

### D9 — Compose scope: single file + conventional override
Supported: a single Compose file, or `compose.yaml` + `compose.override.yaml` (the conventional
pair). Reported as ambiguous (not guessed): arbitrary `-f` chains, extra override files, and
`profiles`. The actual merge is always delegated to `docker compose`.

### D10 — Execution requires HIGH confidence
Automatic `omadev up` runs only at HIGH confidence (a supported Compose file parsed successfully
with services extracted). At MEDIUM/LOW it refuses and explains, rather than guessing a command.

### D11 — v0.1 execution is Compose-only
`up`/`down`/`status`/`logs` operate only on Compose-backed projects. Non-Compose repositories are
inspect-only; native (Docker-less) execution is deferred to a later milestone.

---

## Architecture decisions

### A1 — Findings-based detection
Detectors are pure functions that read the repository and return typed findings; a separate
aggregator composes the normalized project model and computes confidence. Detectors stay small,
independently testable, and order-independent, so new technologies can be added without touching
the core.

### A2 — Docker Compose v2 only
Omadev targets `docker compose` (v2). If only legacy `docker-compose` (v1) is present, it reports a
clear prerequisite error rather than using it. One command form, deterministic behavior, lower
maintenance.

### A3 — Executor with capture and stream modes
The subprocess boundary is an interface with two modes: capture (run to completion, return output
— used to verify state) and stream (wire child stdio to the terminal — used for `logs -f` and
interactive output). A single fake implementation covers both for tests, so planning and command
logic run without a Docker daemon.

### A4 — Standard-library logging
Logging uses `log/slog`, leveled via `--verbose`/`--debug`. Debug output may reveal detectors,
files, findings, the plan, and underlying commands, but never secret values.

---

## Project and setup decisions

### S1 — Module path
`github.com/c3nk/omadev`.

### S2 — Minimum Go version
Go 1.23 (the `go` directive in `go.mod` and the version tested in CI).

### S3 — Dependencies
Standard library first; Cobra for the CLI, `gopkg.in/yaml.v3` for the Compose parser. No other
mandatory third-party modules.

### S4 — Output style
Color and status symbols on a TTY; automatic degrade to plain ASCII on a pipe/redirect or under
`NO_COLOR`/`--no-color`. No TUI framework in v0.1; output stays behind a thin `ui` layer.

### S5 — CI platform matrix
Unit tests on Linux and macOS; release binaries are Linux-first; Docker integration tests run in a
separate Linux job.

### S6 — Issue backlog
Tracked initially as an in-repo document, convertible to GitHub Issues once a remote exists.

### S7 — Version control workflow
Small, logical commits. Every commit must be green locally (`go build ./... && go vet ./... && go
test ./...`) before it lands; the default branch is never left in a red state.

### S8 — Continuous integration
A GitHub Actions workflow runs `go vet` + `go test` + `go build` on push/PR as a Linux safety net,
added in the foundation milestone; a Docker integration job is added with orchestration.

### S11 — Distribution
v0.1.0 ships via GitHub Releases (static `CGO_ENABLED=0` Linux binary, amd64; arm64 optional) and
the AUR (`omadev-bin` prebuilt, plus an optional `omadev` from source), with release checksums that
the AUR package pins. Packaging lives under `packaging/aur/`. Inclusion in Omarchy's own package
repository or default package set is an upstream maintainer decision and is out of v0.1 scope; no
official affiliation is implied.
