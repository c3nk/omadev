# Omadev — Issue Backlog (M0–M2)

> Status: Draft for review · Small, independently implementable issues for the first public
> release (`v0.1.0`). Each issue is sized so one developer or coding agent can implement and
> review it on its own. IDs are stable references; convert to `gh` issues later (S6).

Format per issue: **Objective · Scope · Acceptance · Dependencies · Non-goals.**

---

## M0 — Foundation (`v0.0.1`)

### M0-1 · Bootstrap Go module and Cobra skeleton
- **Objective:** a buildable CLI answering `--version` and `--help`.
- **Scope:** `go mod init github.com/c3nk/omadev` (Go 1.23); Cobra root command; `cmd/root.go` as
  composition root; `--version` wired to a build-stamped version var.
- **Acceptance:** `go build ./...` produces `omadev`; `omadev --version` and `omadev --help` work;
  `go vet ./...` clean.
- **Dependencies:** none.
- **Non-goals:** any subcommand logic.

### M0-2 · Logging and global flags
- **Objective:** leveled logging and shared flags.
- **Scope:** `internal/logging` over `log/slog`; global `--verbose`, `--debug`, `--no-color`
  flags; debug never prints secrets.
- **Acceptance:** `--debug` raises level to debug on stderr; default is warn/error; unit test
  asserts level selection.
- **Dependencies:** M0-1.
- **Non-goals:** detector/plan debug content (added in M1/M2).

### M0-3 · Error model and exit codes
- **Objective:** one place that maps errors to process exit codes.
- **Scope:** typed `CodedError`; `cmd/root.go` translates returned errors to codes 0/2/3/4/5/6
  (ARCHITECTURE §13).
- **Acceptance:** table-driven test maps each error class to its code; `1` reserved for Cobra
  usage errors.
- **Dependencies:** M0-1.
- **Non-goals:** producing those errors from real logic yet.

### M0-4 · UI output layer and prompt
- **Objective:** consistent, script-friendly output.
- **Scope:** `internal/ui` writers + status markers; TTY color, degrade to plain ASCII on
  pipe/redirect or `NO_COLOR`/`--no-color` (S4); `Continue? [Y/n]` prompt (default yes).
- **Acceptance:** tests assert plain output when not a TTY and `[Y/n]` default-yes parsing
  (including empty input, `n`, `N`).
- **Dependencies:** M0-1.
- **Non-goals:** TUI; command-specific rendering.

### M0-5 · Makefile and static build
- **Objective:** reproducible local build/test targets.
- **Scope:** `Makefile` with `build` (`CGO_ENABLED=0`), `test`, `vet`, `release` (skeleton).
- **Acceptance:** `make build test vet` all succeed on Linux and macOS.
- **Dependencies:** M0-1.
- **Non-goals:** the real release job (M2-12).

### M0-6 · GitHub Actions CI
- **Objective:** Linux safety net matching the local gate.
- **Scope:** `.github/workflows/ci.yml` running `go vet` + `go test` + `go build`; unit tests on
  Linux and macOS (S5).
- **Acceptance:** CI is green on a trivial PR; fails on a deliberately broken test.
- **Dependencies:** M0-1.
- **Non-goals:** Docker integration job (M2-14); release job (M2-12).

### M0-7 · README, LICENSE, repo docs wiring
- **Objective:** project presentable and correctly framed.
- **Scope:** MIT `LICENSE`; `README.md` describing Omadev as "a developer experience CLI designed
  for Omarchy" (not "the Omarchy dev CLI"), no AI attribution; link the `docs/`.
- **Acceptance:** README states scope, install (placeholder), and safety stance; wording review
  passes the branding guideline.
- **Dependencies:** none.
- **Non-goals:** final install instructions (depend on M2-12/M2-13).

### M0-8 · Test fixtures scaffold
- **Objective:** fixture repositories in place for M1.
- **Scope:** `testdata/` skeleton: `docker-compose-fastapi-react/`, `node-vite/`,
  `python-fastapi/`, `monorepo/`, `invalid-compose/`, `missing-env/`, `unknown-project/`.
- **Acceptance:** each fixture contains the minimal files its name implies; a smoke test lists
  them.
- **Dependencies:** none.
- **Non-goals:** detector assertions (M1).

---

## M1 — Detection Engine (`v0.0.2`)

### M1-1 · Project model
- **Objective:** the normalized types.
- **Scope:** `internal/project` model (`Project`, `Component`, `Service`, `Port`, `Runtime`,
  `URL`, `Confidence`, `ExecutionStrategy`) per ARCHITECTURE §4.
- **Acceptance:** types compile; zero-value semantics documented; no behavior yet.
- **Dependencies:** M0-1.
- **Non-goals:** populating them.

### M1-2 · Project-root discovery (upward-only)
- **Objective:** deterministic root detection.
- **Scope:** `project/root.go` walks upward from CWD to the first recognized Compose file (or
  nearest project marker), stops at Git/filesystem root; no downward or filesystem-wide search
  (D5).
- **Acceptance:** tests cover: root at CWD, root above CWD, no root found, stop-at-git-root.
- **Dependencies:** M1-1.
- **Non-goals:** monorepo child-Compose discovery (out of scope, D5).

### M1-3 · Detector framework
- **Objective:** the findings-based contract (A1).
- **Scope:** `Detector`, `Finding`, `FindingKind`, `Context` (rooted read-only `fs.FS`),
  `Registry`; ignore `.git/`, `node_modules/`, `.venv/`, `venv/`, `dist/`, `build/`.
- **Acceptance:** a stub detector runs through the registry; `fs.FS` refuses paths outside root
  (symlink-escape test).
- **Dependencies:** M1-1.
- **Non-goals:** real detectors.

### M1-4 · Aggregator and confidence
- **Objective:** compose findings into a `Project` + confidence.
- **Scope:** `aggregate.go` merges `[]Finding` into `Project`; computes HIGH/MEDIUM/LOW
  (SPEC §6.2); ambiguity warnings prevent HIGH (D9/D10); leaves the config-override seam (D2).
- **Acceptance:** table-driven tests: findings in → expected `Project` + confidence out, including
  ambiguity → not HIGH.
- **Dependencies:** M1-3.
- **Non-goals:** detector-specific parsing.

### M1-5 · Docker Compose detector
- **Objective:** read-only, Docker-free Compose understanding.
- **Scope:** parse `compose.yaml`/`.yml`/`docker-compose.*` with `yaml.v3`; extract
  services/images/published ports/healthcheck presence; supported = single file or
  `compose.yaml`+`compose.override.yaml` (D9); do **not** resolve `.env`/interpolation (D3);
  invalid file → `KindWarning` + exit-4 path; multiple/override/`profiles` → ambiguity warning.
- **Acceptance:** fixtures `docker-compose-fastapi-react/` and `invalid-compose/` produce expected
  findings; no secret interpolation occurs.
- **Dependencies:** M1-3, M1-4.
- **Non-goals:** executing Compose.

### M1-6 · Dockerfile detector
- **Objective:** note Dockerfile presence.
- **Scope:** detect `Dockerfile`(s) and emit a technology finding; no build, no parse of RUN
  commands as instructions.
- **Acceptance:** fixture with a Dockerfile yields the finding; content is never executed.
- **Dependencies:** M1-3.
- **Non-goals:** image analysis.

### M1-7 · Node detector (incl. Vite/React/TypeScript)
- **Objective:** understand a Node project from evidence.
- **Scope:** parse `package.json`; detect package manager (from lockfile: `package-lock.json`/
  `pnpm-lock.yaml`/`yarn.lock`), `dev`/`test`/`build` scripts, and React/TypeScript/Vite from
  dependencies; no package-manager execution.
- **Acceptance:** `node-vite/` yields Node + manager + Vite + dev-script findings; MEDIUM
  confidence when standalone.
- **Dependencies:** M1-3, M1-4.
- **Non-goals:** running scripts; native execution.

### M1-8 · Python detector (incl. FastAPI/pytest)
- **Objective:** understand a Python project from evidence.
- **Scope:** inspect `pyproject.toml`/`requirements.txt`/`uv.lock`/`poetry.lock`/`Pipfile`; detect
  dependency manager, FastAPI, pytest, and Python version hints; infer no startup command without
  evidence.
- **Acceptance:** `python-fastapi/` yields Python + manager + FastAPI + pytest findings.
- **Dependencies:** M1-3, M1-4.
- **Non-goals:** running Python tooling.

### M1-9 · PostgreSQL detector
- **Objective:** identify Postgres from strong evidence.
- **Scope:** primarily from a Compose service image (`image: postgres:...`); secondary dependency
  evidence; never read/print DB credentials.
- **Acceptance:** the Compose fixture yields a PostgreSQL finding with `Role: database`; no
  credentials touched.
- **Dependencies:** M1-5.
- **Non-goals:** connecting to a database.

### M1-10 · mise detector
- **Objective:** informational runtime requirements.
- **Scope:** parse `mise.toml` / `.tool-versions` into `Runtime` findings (e.g. Node 24, Python
  3.13); no installation.
- **Acceptance:** a fixture with mise config yields expected runtime findings.
- **Dependencies:** M1-3.
- **Non-goals:** installing runtimes (M3).

### M1-11 · `omadev inspect` command
- **Objective:** the read-only overview command.
- **Scope:** wire detectors → aggregator → `ui` rendering (detected, runtime, services, execution,
  commands, confidence); lightweight `docker info` probe for command rendering (D3/D4); succeeds
  without Docker.
- **Acceptance:** on the Compose fixture, output matches SPEC §5.1 shape with `Confidence: HIGH`;
  on `unknown-project/`, reports unknown (exit 2) without guessing.
- **Dependencies:** M1-4 through M1-10.
- **Non-goals:** any execution.

---

## M2 — Docker Compose Orchestration (`v0.1.0`)

### M2-1 · Executor abstraction
- **Objective:** testable subprocess execution (A3).
- **Scope:** `internal/exec` `Executor` interface with `Capture` and `Stream`; `os.go` real impl;
  `fake.go` test double; structured argv only, no shell.
- **Acceptance:** fake records exact argv; a capture test and a stream test pass with no Docker.
- **Dependencies:** M0-1.
- **Non-goals:** Compose-specific argv.

### M2-2 · Compose v2 integration
- **Objective:** build and delegate Compose commands.
- **Scope:** `internal/compose` builds `docker compose up -d|down|logs -f|ps` argv for v2 only
  (A2); routes through the executor; maps missing v2 (or only legacy `docker-compose`) to a
  prerequisite error.
- **Acceptance:** unit tests assert generated argv per action; legacy-only environment yields
  exit-3 error text.
- **Dependencies:** M2-1.
- **Non-goals:** running against a real daemon in unit tests.

### M2-3 · Runtime checks and privilege detection
- **Objective:** verify Docker and decide on elevation (D4).
- **Scope:** `checks/runtime.go` probes `docker info` without `sudo`; classifies
  available/unavailable/permission-denied; never modifies groups or settings.
- **Acceptance:** tests (via fake executor) cover the three outcomes and the resulting command
  rendering / exit-3 path.
- **Dependencies:** M2-1.
- **Non-goals:** elevating privileges.

### M2-4 · Environment checks
- **Objective:** report env prerequisites without leaking secrets.
- **Scope:** `checks/env.go` reports missing `.env` and variable **names** implied by
  `.env.example`; never reads values.
- **Acceptance:** `missing-env/` fixture yields the expected variable-name report; no value is
  ever read.
- **Dependencies:** M0-1.
- **Non-goals:** creating or modifying `.env`.

### M2-5 · Planner and plan rendering
- **Objective:** turn project+checks into a shown plan.
- **Scope:** `internal/plan` `Action`/`Plan`; planner builds the Compose plan; `ui` renders
  strategy + exact command(s) (SPEC §4).
- **Acceptance:** planner test: HIGH-confidence Compose project → expected ordered plan; render
  matches SPEC §5.2 shape.
- **Dependencies:** M1-11, M2-2.
- **Non-goals:** execution.

### M2-6 · `omadev up`
- **Objective:** the full safe lifecycle.
- **Scope:** inspect → check → plan → show → confirm → execute → verify; refuse below HIGH (D10,
  exit 2); verify service state, not just exit 0 (D6); report per-service result.
- **Acceptance:** with the fake executor, a full up flow verifies state and reports success;
  cancel at prompt → exit 5; non-zero from Compose → exit 6.
- **Dependencies:** M2-5, M2-3, M2-7-status-query (M2-8).
- **Non-goals:** non-Compose execution.

### M2-7 · `omadev down`
- **Objective:** stop safely, preserve data.
- **Scope:** plan/confirm/execute `docker compose down`; **never** `--volumes`.
- **Acceptance:** generated argv contains no volume-removal flag; confirm/cancel paths covered.
- **Dependencies:** M2-5.
- **Non-goals:** volume removal option (future).

### M2-8 · `omadev status`
- **Objective:** read-only state and ports.
- **Scope:** query `docker compose ps`; map to `running`/`healthy`/`unhealthy`/`stopped`/
  `unknown` (D6); show ports from published metadata (D7).
- **Acceptance:** parser test maps `ps` output to the state set; ports come only from Compose.
- **Dependencies:** M2-2.
- **Non-goals:** inventing application health.

### M2-9 · `omadev logs`
- **Objective:** transparent log tailing.
- **Scope:** stream `docker compose logs -f` via the executor's Stream mode; no logging
  abstraction.
- **Acceptance:** command wires to Stream with correct argv; manual check tails a running stack.
- **Dependencies:** M2-2, M2-1.
- **Non-goals:** filtering/formatting logs.

### M2-10 · Bare `omadev` overview
- **Objective:** informational root command (D1).
- **Scope:** inspect + show status/next-commands; if not running, show the start command but do
  **not** offer to start.
- **Acceptance:** running vs not-running fixtures render the two SPEC §5.6 shapes; no execution
  path from the root command.
- **Dependencies:** M1-11, M2-8.
- **Non-goals:** auto-start.

### M2-11 · Release job (binaries + checksums)
- **Objective:** publish installable artifacts.
- **Scope:** GHA `release` job on tag: build static binaries (amd64; arm64 optional),
  `SHA256SUMS`, attach to a GitHub Release.
- **Acceptance:** a test tag produces a draft release with binaries and checksums.
- **Dependencies:** M0-5, M0-6.
- **Non-goals:** AUR (M2-12).

### M2-12 · AUR packaging
- **Objective:** Omarchy-idiomatic install (S11).
- **Scope:** `packaging/aur/PKGBUILD` + `.SRCINFO` for `omadev-bin` (pins release `sha256sums`);
  optional `omadev` from source; no implied Omarchy affiliation.
- **Acceptance:** `PKGBUILD` builds/installs in a clean Arch container; checksums verify.
- **Dependencies:** M2-11.
- **Non-goals:** OPR / default-package inclusion (upstream decision).

### M2-13 · Docker integration CI job
- **Objective:** exercise real Compose behind a build tag.
- **Scope:** Linux-only CI job running integration tests against a real `docker compose` on the
  fixtures (S5).
- **Acceptance:** integration job green on `docker-compose-fastapi-react/`; skipped in the default
  unit-test job.
- **Dependencies:** M2-6..M2-9.
- **Non-goals:** macOS/Windows Docker testing.

---

## Sequencing notes

- M0 issues are largely parallel; M0-1 unblocks the rest.
- M1 must fully pass before M2 begins (acceptance-gated per the process phases).
- Within M2, executor (M2-1) and Compose integration (M2-2) unblock the commands; distribution
  (M2-11/M2-12) can proceed in parallel once M0-5/M0-6 exist.
