# Omadev — Roadmap

> Status: Draft for review · Milestone structure from the product goals, refined by the
> decisions in [`DECISIONS.md`](DECISIONS.md). v0.1.0 is the first public release (end of M2).

Guiding rule: optimize for a small, trustworthy v0.1, not feature count. When choosing between
clever automation and predictable, transparent behavior, choose the latter.

---

## M0 — Foundation → `v0.0.1`

**Goal:** a clean, testable Go project that builds and runs `omadev --version` / `omadev --help`.

Deliverables:
- Go module `github.com/c3nk/omadev`, Go 1.23 (S1, S2).
- Cobra command skeleton; business logic separated from handlers (composition root in
  `cmd/root.go`).
- `log/slog` logging with `--verbose` / `--debug`; `--no-color` handling (A4, S4).
- Typed error model with exit-code mapping (ARCHITECTURE §13–14).
- Unit-test infrastructure and first fixtures scaffold under `testdata/`.
- GitHub Actions `ci.yml`: `go vet` + `go test` + `go build` on push/PR (S8).
- `Makefile` (`build`/`test`/`vet`/`release`), `README.md` (framed), MIT `LICENSE`.

Non-goals: any detection or execution logic.

---

## M1 — Detection Engine → `v0.0.2`

**Goal:** `omadev inspect` correctly understands conventionally structured repositories. No
execution.

Deliverables:
- Project model and upward-only root discovery (ARCHITECTURE §4, D5).
- Findings-based detector framework: `Detector`, `Finding`, `Registry`, aggregator (A1).
- Detectors: Docker, Docker Compose, Node.js, Python, Vite, FastAPI, PostgreSQL, mise.
- Read-only, Docker-free Compose parsing (D3); supported-layout / ambiguity handling (D9).
- Confidence computation HIGH/MEDIUM/LOW (SPEC §6.2).
- `omadev inspect` output via the `ui` layer, with the lightweight `docker info` probe for command
  rendering (D3/D4).
- Detector unit tests against every fixture.

Non-goals: `up`/`down`/`status`/`logs`, any state change.

---

## M2 — Docker Compose Orchestration → `v0.1.0` (first public release)

**Goal:** safely start, stop, inspect, and tail a Compose-backed environment, end to end.

Deliverables:
- `omadev up`, `omadev down`, `omadev status`, `omadev logs`.
- Planner + plan rendering + confirmation prompt (SPEC §4).
- Host prerequisite checks, including runtime privilege detection (D4) and Docker v2 requirement
  (A2).
- Executor with capture + stream modes and a fake for tests (A3).
- Compose v2 delegation for execution; service-state verification (D6); `down` never removes
  volumes.
- Bare `omadev` informational overview (D1).
- **Distribution (S11):** static binary release job attaching artifacts + `SHA256SUMS` to GitHub
  Releases; `packaging/aur/PKGBUILD` (`omadev-bin`, optional `omadev` from source).
- Docker integration CI job (behind a build tag) on Linux (S5).
- `SECURITY.md` reporting contact finalized.

Acceptance: the SPEC §11 happy path passes on an unseen Compose repository with no `.omadev.yaml`.

Non-goals: native (Docker-less) execution; anything in the deferred list.

---

## M3 — Native Development → `v0.2`

Docker-less projects using Node / Python / Go / Rust / mise. Introduces a second execution
strategy behind the same model→plan→executor pipeline. Runtime installation stays **explicit and
confirmed** (e.g. "mise can provide Node 24" — never automatic). Only build the interfaces this
requires; do not pre-build it during v0.1.

---

## M4 — Developer Workflow → `v0.3`

`omadev test`, `omadev shell`, `omadev open`, `omadev exec`; plus port-conflict detection, stronger
health checks, and better monorepo support.

---

## M5 — Project Creation → `v0.5`

`omadev new` with a small set of stacks (e.g. React + FastAPI, React + Node, Django, Rails, Go). No
template generation before this milestone.

---

## M6 — Omarchy Experience → `v1.0`

Long-term Omarchy-native integration: notifications, Hyprland/Quickshell, a terminal launcher and
project launcher, and workspace automation. Kept out of v0.1 entirely and isolated when built.

---

## Release and distribution process

1. Develop step by step; each commit is green locally (`go build ./... && go vet ./... && go test
   ./...`) before it lands (S7).
2. CI re-runs the same gate on Linux (and macOS for unit tests) as a safety net (S8).
3. Tagging a release triggers the GHA `release` job: build static binaries, generate
   `SHA256SUMS`, publish a GitHub Release.
4. Update `packaging/aur/PKGBUILD` + `.SRCINFO` to the new version/checksums and push to the AUR.
5. Omarchy users install via `yay -S omadev` (or the Omarchy "Install > AUR" flow). Inclusion in
   Omarchy's OPR / default package set remains an upstream maintainer decision, not a project
   deliverable (S11).

---

## Versioning

Semantic-ish, milestone-anchored: `v0.0.1` (M0), `v0.0.2` (M1), `v0.1.0` (M2, first public),
`v0.2` (M3), `v0.3` (M4), `v0.5` (M5), `v1.0` (M6). Pre-1.0 minors may carry breaking changes;
they will be noted in release notes.
