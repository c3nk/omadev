# Omadev — Architecture (v0.1)

> Status: Draft for review · Implements [`SPEC.md`](SPEC.md) and the decisions in
> [`DECISIONS.md`](DECISIONS.md). Architecture forks resolved in this doc: findings-based
> aggregation (A), Compose v2 only (B), capture+stream executor (C), exit codes 0/2/3/4/5/6 (D).

---

## 1. Principles

- Business logic lives in `internal/` services; Cobra command handlers only parse flags and call
  those services.
- No global mutable state. Dependencies are constructed once in a composition root and passed in
.
- Interfaces at external boundaries only (subprocess execution, filesystem where useful) — not
  everywhere. No speculative plugin framework.
- Deterministic: same repository in, same model out. No network calls during detection. No execution of repository-defined scripts during inspection.

---

## 2. Package layout

```
omadev/
├── cmd/
│   ├── root.go        # composition root; bare `omadev`; global flags (--verbose/--debug/--no-color)
│   ├── inspect.go
│   ├── up.go
│   ├── down.go
│   ├── status.go
│   └── logs.go
│
├── internal/
│   ├── detect/        # detectors + findings + aggregator + registry
│   │   ├── detector.go    # Detector interface, Finding types, Registry
│   │   ├── aggregate.go   # findings -> Project model + confidence
│   │   ├── compose.go
│   │   ├── docker.go
│   │   ├── node.go        # Node + package manager + React/TypeScript/Vite findings
│   │   ├── python.go      # Python + dep manager + FastAPI/pytest findings
│   │   ├── postgres.go
│   │   └── mise.go
│   │
│   ├── project/       # normalized model + root discovery
│   │   ├── model.go
│   │   └── root.go        # upward-only project-root discovery (D5)
│   │
│   ├── plan/          # execution plan
│   │   ├── planner.go
│   │   └── action.go
│   │
│   ├── exec/          # subprocess abstraction (capture + stream)
│   │   ├── executor.go    # Executor interface, Command, Result
│   │   ├── os.go          # real os/exec implementation
│   │   └── fake.go        # test double
│   │
│   ├── compose/       # Docker Compose v2 integration (argv building, delegation)
│   │   └── compose.go
│   │
│   ├── checks/        # host prerequisite checks
│   │   ├── runtime.go     # docker availability + privilege detection (D4)
│   │   ├── env.go         # .env / .env.example reporting (never reads secret values)
│   │   └── ports.go
│   │
│   ├── ui/            # thin output layer (no TUI)
│   │   ├── output.go      # writers, symbols, color degrade (S4)
│   │   └── prompt.go      # Continue? [Y/n]
│   │
│   └── logging/       # slog setup, verbose/debug levels
│       └── logging.go
│
├── packaging/aur/     # PKGBUILD + .SRCINFO (S11)
├── testdata/          # fixture repositories
├── docs/
├── .github/workflows/ci.yml
├── go.mod             # module github.com/c3nk/omadev, go 1.23 (S1, S2)
├── Makefile
├── LICENSE            # MIT
└── README.md
```

A separate `config/` package is **omitted** in v0.1 (`.omadev.yaml` deferred,
D2); a seam is left in the aggregator where an override layer would merge.

---

## 3. Data flow

```
Repository files
      │  (read-only)
      ▼
  Detectors ──► []Finding
      │
      ▼
  Aggregator ──► project.Project (+ Confidence)
      │
      ▼
   Checks (host prerequisites: docker, privilege, ports, env)
      │
      ▼
   Planner ──► plan.Plan ([]Action)
      │
      ▼   (SHOW + CONFIRM)
   Executor ──► runs Command(s)
      │
      ▼
   Verify (re-query service state)
```

Read-only commands (`inspect`, `status`, bare `omadev`) stop before the Executor.

---

## 4. Project model

```go
package project

type Confidence int // Low, Medium, High

type ExecutionStrategy int // None, Compose  (v0.1: only Compose is executable)

type Project struct {
    Name              string
    Root              string            // absolute path, from upward-only discovery (D5)
    Components        []Component       // detected app parts (dir + technology)
    Services          []Service         // Docker Compose services
    Runtimes          []Runtime         // version hints (informational in v0.1)
    URLs              []URL             // evidence-backed, from published ports (D7)
    ExecutionStrategy ExecutionStrategy
    Confidence        Confidence
    Notes             []string          // e.g. "compose.override.yaml merged", ambiguity notes
}

type Component struct {
    Name       string   // e.g. "frontend"
    Path       string   // relative to Root
    Technology []string // e.g. ["React", "TypeScript", "Vite"]
}

type Service struct {
    Name        string
    Image       string
    Ports       []Port   // only published ports
    HasHealth   bool     // Compose healthcheck present
    Role        string   // "app"/"api"/"database" when evidence supports it, else ""
}

type Port struct {
    Published int
    Target    int
}

type Runtime struct {
    Name    string // "Node", "Python"
    Version string // "24", "3.13"
    Source  string // "mise.toml", ".tool-versions", "engines"
}

type URL struct {
    Label string // "app", "api"
    URL   string // http://localhost:<published>
}
```

Semantics (P4): `Service` = Compose service; `Component` = detected app part; `Runtime` = version
hint; `URL` = evidence-backed address from a published port.

---

## 5. Detector architecture (fork A — findings-based)

Detectors are pure functions: read the repository, return typed findings, never mutate, never
execute.

```go
package detect

type Context struct {
    Root string          // project root
    FS   fs.FS           // rooted, read-only filesystem (aids testing, blocks escapes)
}

type Detector interface {
    Name() string
    Detect(ctx Context) ([]Finding, error)
}

type FindingKind int
const (
    KindTechnology FindingKind = iota // "React", "FastAPI"
    KindService                       // a Compose service
    KindRuntime                       // "Node 24"
    KindExecution                     // "Docker Compose is authoritative"
    KindPort                          // published port
    KindWarning                       // "compose.yaml unparseable", "multiple override files"
)

type Finding struct {
    Kind       FindingKind
    Detector   string
    Value      string
    Confidence project.Confidence
    Data       map[string]string // structured extras (image, path, version, ...)
}
```

A `Registry` holds the ordered detector set; detectors run independently, order does not affect
the result. The **aggregator** (`aggregate.go`) composes `[]Finding` into a `project.Project` and
computes overall confidence (section 6). Adding a technology (Go, Rust) means adding a detector
and, if needed, an aggregation rule — the core is untouched.

Initial detectors: Docker, Docker Compose, Node.js, Python, PostgreSQL, mise. Vite/React/
TypeScript are emitted by the Node detector and FastAPI/pytest by the Python detector (as
`KindTechnology` findings) rather than as separate detectors, to avoid parsing the same file
twice.

---

## 6. Confidence computation

Overall project confidence (SPEC §6.2):

- **HIGH** — a supported Compose file (D9) parsed successfully, services extracted, execution
  strategy = Compose. Only HIGH permits automatic `up` (D10).
- **MEDIUM** — application evidence but no authoritative execution strategy.
- **LOW** — a project type recognized, no reliable development command.

Ambiguity (multiple/override beyond the conventional pair, `profiles`, arbitrary `-f`) produces a
`KindWarning` finding and prevents HIGH, so execution is refused rather than guessed (D9).

---

## 7. Checks (host prerequisites)

- `runtime.go` — is Docker available and reachable? **Privilege detection (D4):** run
  `docker info` (structured argv, no shell) without `sudo`; on success → no elevation needed; on a
  permission error → elevation required, reflected in shown commands. Nothing hardcoded; the
  `docker` group is never modified.
- `env.go` — reports a missing `.env` and, from `.env.example`, the variable **names** it implies.
  It never reads or prints secret values (SPEC §7).
- `ports.go` — reports published ports from Compose metadata (D7). Port-conflict *resolution* is
  out of scope for v0.1 (roadmap M4).

Checks return structured results that the planner and UI consume; a failed prerequisite maps to
exit code 3.

---

## 8. Planner

```go
package plan

type Action struct {
    Title   string        // human summary, e.g. "Start Docker Compose environment"
    Command exec.Command  // the exact argv to run (may be empty for non-exec steps)
    Kind    ActionKind    // Exec, Wait, Inspect
}

type Plan struct {
    Strategy project.ExecutionStrategy
    Actions  []Action
}
```

The planner turns a `Project` + checks into an ordered `Plan`. The UI renders the plan (SHOW) and
asks for confirmation (CONFIRM) before the executor runs any `Exec` action.

---

## 9. Executor (fork C — capture + stream)

```go
package exec

type Command struct {
    Name string   // "docker"
    Args []string // ["compose", "up", "-d"]  — structured, never shell-interpolated
    Dir  string   // project root
    Env  []string // explicit; no inherited secret leakage in debug output
}

type Result struct {
    ExitCode int
    Stdout   string // populated in capture mode
    Stderr   string
}

type Executor interface {
    // Capture runs to completion and returns output (up/down/status verification).
    Capture(ctx context.Context, c Command) (Result, error)
    // Stream wires child stdio to the user's terminal (logs -f, interactive).
    Stream(ctx context.Context, c Command, io Stdio) (int, error)
}
```

- `os.go` — real implementation over `os/exec`.
- `fake.go` — records commands and returns scripted results, so planner/command tests run with no
  Docker daemon and never touch the developer machine.

`up` must not report success on exit code 0 alone; it re-queries service state afterwards
(VERIFY, SPEC §5.2).

---

## 10. Compose integration (fork B — v2 only)

- **Parsing (inspect, read-only, Docker-free):** `internal/detect/compose.go` parses the Compose
  YAML with `gopkg.in/yaml.v3` (S3) into services/images/published-ports/healthcheck presence. It
  does **not** resolve `.env`/interpolation and does not run Docker (D3). Unresolved `${VAR}` in a
  security-relevant position is left symbolic and, where it blocks understanding, downgrades
  confidence.
- **Execution (up/down/logs/status):** `internal/compose/compose.go` builds argv for **Docker
  Compose v2** (`docker compose ...`) and delegates through the executor. Only `docker
  compose` is supported; if only legacy `docker-compose` v1 is present, Omadev reports a clear
  prerequisite error (exit code 3) rather than using it.
- **Supported layouts (D9):** a single Compose file, or `compose.yaml` + `compose.override.yaml`
  (the conventional pair). The actual merge is always done by the real `docker compose`, not
  reimplemented. Profiles, arbitrary `-f` chains, and extra override files are reported ambiguous.

---

## 11. UI layer

`internal/ui` is a thin abstraction (no TUI framework). `output.go` provides leveled/labelled
writers and status markers; on a TTY it uses color and symbols, and it degrades to plain ASCII on
a pipe/redirect or under `NO_COLOR` / `--no-color` (S4). `prompt.go` implements `Continue? [Y/n]`
(default yes). Keeping this behind an interface leaves room for a future TUI without touching
business logic.

---

## 12. Logging and debug

`internal/logging` configures `log/slog` (stdlib). Default: warnings/errors to stderr. `--verbose`
raises to info; `--debug` raises to debug and includes detectors run, files inspected, findings,
confidence reasoning, the generated plan, and underlying commands. Debug output must never
include secret values.

---

## 13. Exit codes (fork D — final)

| Code | Meaning | Example |
|------|---------|---------|
| 0 | success | inspect/up/down/status/logs completed |
| 2 | unsupported / unknown project | no recognized project, or confidence below HIGH for `up` |
| 3 | missing prerequisite | Docker unavailable, or only legacy `docker-compose` present |
| 4 | invalid project configuration | Compose file cannot be parsed |
| 5 | user cancellation | user answered `n` at confirmation |
| 6 | execution failure | `docker compose up` returned non-zero |

`1` is reserved for Cobra's generic usage/argument errors.

---

## 14. Error model

Errors carry an exit code and an actionable message (what was detected, what is required, the next
command). Underlying failures are surfaced with the failed command and exit status, not hidden
. A small typed error (e.g. `type CodedError struct { Code int; ... }`) lets `cmd/root.go`
translate a returned error into the correct process exit code in one place.

---

## 15. Testing

- **Detector unit tests** against fixtures in `testdata/` (`docker-compose-fastapi-react/`,
  `node-vite/`, `python-fastapi/`, `monorepo/`, `invalid-compose/`, `missing-env/`,
  `unknown-project/`).
- **Aggregator/model tests** — findings in, expected `Project` + confidence out.
- **Planner tests** — project in, expected `Plan` out.
- **Executor tests** — via `fake.go`; assert the exact argv, never run Docker.
- **Integration tests** — real `docker compose` behind a build tag, run only in a separate Linux
  CI job (S5).

---

## 16. Omarchy isolation

Omarchy-specific behavior (if any is needed beyond standard Linux) is kept in a single, clearly
named place rather than scattered through detection/execution. v0.1 avoids
hardcoding assumptions that would block broader Linux compatibility.

---

## 17. Build and distribution

Static binary via `CGO_ENABLED=0 go build` (amd64; arm64 optional), reproducible where practical.
A `Makefile` provides `build` / `test` / `vet` / `release` targets; the GHA `release` job attaches
binaries to GitHub Releases and the `packaging/aur/` PKGBUILD tracks them (S8, S11). Detailed
packaging/release flow lives in [`ROADMAP.md`](ROADMAP.md).
