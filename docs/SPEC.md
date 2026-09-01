# Omadev — Product Specification (v0.1)

> Status: Draft for review · Derives from the product goals and the decisions in
> [`DECISIONS.md`](DECISIONS.md) (the D/A/S records). Where this document and the original prose
> disagree, the decision records win.

---

## 1. Overview

Omadev is a developer-experience CLI, designed first for Omarchy Linux and usable on other Linux
systems where practical. It is a layer over the development environment a repository **already**
defines — it does not create one.

Guiding sentence:

> Omadev doesn't define your development environment. It understands it.

Omadev inspects a repository, determines how its development environment is meant to run, checks
prerequisites, shows what it intends to do, and — only with confirmation — starts and inspects
that environment.

Omadev is: Omarchy-first, terminal-first, transparent, deterministic, safe by default,
non-destructive, and useful without configuration.

Omadev does **not** replace Docker, Docker Compose, mise, npm/pnpm/yarn, Python tooling, pytest,
Vitest, systemd, Git, or a project's own scripts. It delegates to them.

---

## 2. Principles (requirements)

1. **Understand, don't reinvent.** Use the environment the repository already defines. Never
   create an alternative one.
2. **Detection before configuration.** Work with zero Omadev-specific config wherever possible.
3. **Compose is authoritative (v0.1).** If a supported Compose file exists, Docker Compose is the
   execution strategy. Technology detection and execution strategy are separate concepts.
4. **Safe by default.** State-changing actions follow the full lifecycle (section 4); no
   `DETECT → EXECUTE` shortcut.
5. **Non-destructive.** Omadev never modifies the repository or the system to "fix" it (section 7).
6. **Transparent.** Always show the underlying command before running it. Teach, don't hide.
7. **Report uncertainty.** When detection is ambiguous, say so; do not guess.

---

## 3. Scope

### In scope for v0.1

- Read-only inspection of conventionally structured repositories.
- Docker Compose orchestration for supported Compose layouts (D9).
- Commands: `omadev`, `omadev inspect`, `omadev up`, `omadev down`, `omadev status`, `omadev logs`.
- Detection of: Docker, Docker Compose, Node.js, Python, Vite, FastAPI, PostgreSQL, mise.

### Explicitly out of scope for v0.1

- Native (Docker-less) execution of Node/Python/Go/Rust — deferred to M3 / v0.2.
- `omadev new`, `test`, `shell`, `open`, `exec`, `doctor`.
- `.omadev.yaml` parsing (D2 — architectural seam kept, no implementation).
- TUI, profiles, project templates, workspace automation, Quickshell/Hyprland integration.
- A `--yes` non-interactive flag as a central feature (may exist minimally, not emphasized).

### Execution scope (D11)

`up` / `down` / `status` / `logs` operate **only** on Compose-backed projects in v0.1. For a
non-Compose repository, `inspect` still reports findings, but `up` refuses to execute rather than
inventing a start command.

---

## 4. Safety model and lifecycle

Every state-changing command follows:

```
DETECT → CHECK → PLAN → SHOW → CONFIRM → EXECUTE → VERIFY
```

- **DETECT** — build the project model from repository files (read-only).
- **CHECK** — verify host prerequisites (e.g. Docker available and reachable).
- **PLAN** — construct the ordered list of actions.
- **SHOW** — print the execution strategy and the exact underlying command(s).
- **CONFIRM** — prompt `Continue? [Y/n]` (default yes). A future `--yes` may skip this.
- **EXECUTE** — run via structured subprocess invocation (no shell interpolation).
- **VERIFY** — inspect resulting service state; never claim success from exit code 0 alone.

Read-only commands (`inspect`, `status`, and the bare `omadev`) stop after DETECT/CHECK and never
reach EXECUTE.

---

## 5. Commands

### 5.1 `omadev inspect` (read-only, most important)

Detects the stack and prints a compact overview: detected technologies, runtime hints, services,
execution strategy, the underlying commands, and a confidence level. It never starts anything.

Parsing is Docker-free (D3). To render the correct command line (with or without `sudo`, per D4),
`inspect` performs a lightweight read-only `docker info` probe; if Docker is absent it still
succeeds and annotates the command with a `docker not available` prerequisite note.

Illustrative output:

```
Project: example-app

Detected
  ok  Docker Compose
  ok  React
  ok  TypeScript
  ok  Vite
  ok  FastAPI
  ok  PostgreSQL

Runtime
  Node       24
  Python     3.13

Services
  frontend
  backend
  postgres

Execution
  Docker Compose

Commands
  Up     docker compose up -d
  Down   docker compose down
  Logs   docker compose logs -f

Confidence: HIGH
```

(Formatting is indicative; symbols/colors follow S4 and degrade to plain ASCII off a TTY.)

### 5.2 `omadev up`

1. inspect the project, 2. determine the execution strategy, 3. check host tooling, 4. build the
plan, 5. show it, 6. confirm, 7. execute, 8. verify service state, 9. report.

`up` runs only at **HIGH** confidence (D10). Below HIGH it refuses and prints the "couldn't
confidently determine how to start this project" message (exit code 2).

```
Project: example-app

Plan
  Start Docker Compose environment:
    docker compose up -d

Continue? [Y/n]
```

After execution, verify and report per-service state (section 6.3):

```
Services
  running  frontend
  running  backend
  running  postgres

Environment started successfully.
```

### 5.3 `omadev down`

Stops the Compose environment. **Never** deletes volumes; development data is preserved. Volume
removal, if ever added, is a future explicit option.

```
Plan
  Stop Docker Compose environment:
    docker compose down

Continue? [Y/n]
```

### 5.4 `omadev status`

Read-only. Shows current service state and ports without changing anything.

```
Project: example-app

Services
  running   frontend
  running   backend
  running   postgres

Ports
  frontend   5173
  backend    8000
  postgres   5432
```

State values: `running`, `healthy`, `unhealthy`, `stopped`, `unknown` (D6). `healthy`/`unhealthy`
appear only when a Compose healthcheck exists. Ports come only from published-port metadata (D7).

### 5.5 `omadev logs`

For v0.1 this transparently delegates to Docker Compose (`docker compose logs -f`). No separate
logging abstraction is built.

### 5.6 Bare `omadev` (root)

Informational only (D1). Inspects the current project and prints an overview. If the environment
is running, it shows status and the relevant next commands. If it is not running, it shows the
detected start command but **does not** offer to start it — starting requires an explicit
`omadev up`.

```
OMADEV

Project
  example-app

Stack
  React + FastAPI + PostgreSQL

Environment
  Docker Compose

Status
  running  frontend   :5173
  running  backend    :8000
  running  postgres   :5432

Commands
  omadev logs
  omadev down
```

---

## 6. Detection, confidence, and readiness

### 6.1 Detection

Detectors inspect known project files, return structured findings, never execute repository
scripts, and never mutate the repository. Inspected files include: `compose.yaml`, `compose.yml`,
`docker-compose.yaml`, `docker-compose.yml`, `Dockerfile`, `package.json`, `pyproject.toml`,
`requirements.txt`, `uv.lock`, `poetry.lock`, `Pipfile`, `mise.toml`, `.tool-versions`, `go.mod`,
`Cargo.toml`. Ignored directories: `.git/`, `node_modules/`, `.venv/`, `venv/`, `dist/`,
`build/`.

Project root is found by walking **upward only** from the working directory to the first directory
with a recognized Compose file (or nearest project marker), stopping at the Git root or filesystem
root. No downward search, no filesystem-wide scan (D5).

### 6.2 Confidence

Three levels: `HIGH`, `MEDIUM`, `LOW`. No fake numeric precision.

- **HIGH** — a supported Compose file exists, parsed successfully, services extracted, execution
  strategy explicit.
- **MEDIUM** — application evidence exists (e.g. `package.json` + Vite dependency + dev script)
  but no authoritative execution strategy.
- **LOW** — a project type is recognized but no reliable development command is found.

Automatic execution requires HIGH (D10).

### 6.3 Readiness (D6)

Omadev reports only what Docker Compose can substantiate:

| State | Meaning |
|-------|---------|
| `running` | container is up; no healthcheck defined |
| `healthy` | container is up and its Compose healthcheck passes |
| `unhealthy` | container is up but its Compose healthcheck fails |
| `stopped` | container is not running |
| `unknown` | state could not be determined |

Container liveness is never presented as application readiness.

---

## 7. Non-destructive guarantees

Omadev v0.1 must NOT: modify source code; generate or modify Dockerfiles, Compose files,
`package.json`, or `pyproject.toml`; install application dependencies; modify or generate `.env`
or secrets; modify system, firewall, or Docker security settings; add the user to the `docker`
group; silently install packages; or auto-fix repositories.

If something is missing, Omadev **reports** it (e.g. a missing `.env` with the variable names it
can infer from `.env.example`) and does not repair it. Omadev never displays secret values.

Docker privilege is detected at runtime, never hardcoded (D4): Omadev probes `docker info` without
`sudo` first; if that fails with a permission error it reports that elevated access is required.

---

## 8. Configuration (deferred — D2)

`.omadev.yaml` is **not** parsed in v0.1. The project model keeps an enrichment seam so an override
layer can be added later. The v0.1 happy path must work with zero Omadev-specific configuration.

---

## 9. Errors and exit codes

Errors are actionable: they state what was detected, what is required, and the next command to
run. Underlying command failures are surfaced (failed command + exit status), not hidden.

Exit codes (finalized in [`ARCHITECTURE.md`](ARCHITECTURE.md)):

| Code | Meaning |
|------|---------|
| 0 | success |
| 2 | unsupported / unknown project |
| 3 | missing prerequisite (e.g. Docker unavailable) |
| 4 | invalid project configuration (e.g. unparseable Compose) |
| 5 | user cancellation |
| 6 | execution failure |

---

## 10. Output conventions

- Terminal output goes through a thin `ui` layer; no TUI framework in v0.1 (S3).
- On a TTY: color and `ok` / `running` / `warn`-style markers. Off a TTY, or under `NO_COLOR` /
  `--no-color`: plain ASCII, script-friendly (S4).
- A `--verbose` / `--debug` flag may print detectors run, files inspected, findings, confidence
  reasoning, the generated plan, and underlying commands. Debug output must never expose secrets.

---

## 11. Acceptance criteria (v0.1 happy path)

Given a previously unseen, conventionally structured Docker Compose repository, with no
Omadev-specific configuration:

```
git clone <repository>
cd <repository>
omadev inspect   # correctly identifies the stack
omadev up        # shows strategy + command, confirms, starts, verifies, reports
omadev status    # shows meaningful per-service state and ports
omadev logs      # streams development logs
omadev down      # stops the environment without destroying persistent volumes
```

Success means a developer can enter an unfamiliar Omarchy development repository, understand it
through Omadev, and safely start and inspect it — transparently and without configuration.
