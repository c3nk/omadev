# Omadev — Security Model (v0.1)

> Status: Draft for review · Codifies the threat model behind [`SPEC.md`](SPEC.md) and
> [`ARCHITECTURE.md`](ARCHITECTURE.md). Core stance: **repository contents are untrusted input.**

---

## 1. Threat model

Omadev runs inside repositories a developer has just cloned and may not have read. A repository can
be crafted by an attacker. Therefore Omadev treats **all repository content as untrusted data, not
as instructions**: file names, file contents, `package.json` scripts, `Makefile` targets,
`pyproject.toml` entries, Compose labels, environment interpolation, and symlinks.

Assets Omadev protects:

- The developer's machine (no unintended command execution, no state changes).
- Secrets (never read, never displayed, never logged).
- The repository's integrity (never modified to "fix" it).

Out of scope for v0.1: sandboxing Docker itself, auditing the images a Compose file pulls, and
defending against a malicious Docker daemon. Omadev delegates execution to Docker Compose and does
not add a security layer over container runtime.

---

## 2. Untrusted-input attack surfaces and mitigations

### 2.1 Arbitrary command execution from repository files
A repository may define commands (`"dev": "curl … | sh"`, a `Makefile` target, a Compose
`command:`). **Reading such a value never means running it.** During inspection Omadev only reads
and classifies; it never executes repository-defined scripts. Execution happens only
through the explicitly supported strategy (Docker Compose v2), only after DETECT→CHECK→PLAN→SHOW→
CONFIRM (SPEC §4).

### 2.2 Shell injection / malicious filenames
Omadev builds subprocess invocations as **structured argv** (`exec.Command{Name, Args}`) and never
passes strings through a shell. There is no `sh -c`, no string interpolation of repository-derived
values into a command line. Filenames containing shell metacharacters, spaces, or newlines are
data, carried as argv elements, never interpreted.

### 2.3 Compose interpolation and environment values
Omadev's own Compose parser (used for read-only `inspect`) does **not** resolve `${VAR}` from
`.env` or the environment (D3). This avoids pulling secret values into the model or debug output.
When actual execution is needed, the real `docker compose` binary performs interpolation itself —
Omadev does not reimplement it and does not echo the resolved result.

### 2.4 Symlinks and paths outside the repository root
Detection runs against a **root-scoped, read-only filesystem view** (`fs.FS` rooted at the project
root). Omadev does not follow symlinks that resolve outside the root and does not read paths above
the root. Project-root discovery walks **upward only**, stops at the Git root or filesystem root,
and never performs a filesystem-wide scan (D5).

### 2.5 Resource / performance abuse
Detection inspects a bounded set of known project files and skips `.git/`, `node_modules/`,
`.venv/`, `venv/`, `dist/`, `build/`. No recursive parsing of arbitrary trees, no expensive
commands, and **no network calls** during detection. A malformed Compose file is reported
as invalid (exit 4), not retried or worked around.

---

## 3. Secret handling

- Omadev never reads the **values** in `.env` or any secret file. `checks/env.go` reports only the
  variable **names** implied by `.env.example`, and flags a missing `.env` (SPEC §7).
- Secrets are never printed in normal output and never in `--verbose` / `--debug` output.
- Omadev never generates secrets and never writes `.env`.

---

## 4. Non-destructive guarantees

Omadev v0.1 does not, under any code path:

- modify source code, or generate/modify `Dockerfile`, Compose files, `package.json`, or
  `pyproject.toml`;
- install application dependencies or silently install packages;
- create, modify, or delete `.env`, secrets, or system/firewall/Docker security settings;
- add the user to the `docker` group or change Docker permissions;
- delete Compose volumes (`down` never passes `--volumes`; development data is preserved);
- auto-fix repositories.

If something is missing, Omadev reports it and stops. The only state changes Omadev makes are
the Docker Compose lifecycle actions the user explicitly confirms.

---

## 5. Privilege

Docker privilege is **detected at runtime, never hardcoded** (D4). Omadev probes `docker info`
without `sudo`; if that fails with a permission error it reports that elevated access (or a
rootless/`docker`-group setup the user controls) is required. Omadev never elevates privileges on
its own, never modifies group membership, and never weakens security settings.

---

## 6. Confirmation boundary

Every state-changing command shows the exact underlying command(s) and requires confirmation
before execution (SPEC §4). A future `--yes` flag may skip the prompt for automation; it is not a
v0.1 emphasis and does not change what is shown, only whether the prompt blocks.

---

## 7. Supply chain

- Minimal dependencies: Cobra and `gopkg.in/yaml.v3` are the only mandatory third-party modules
  (S3). Versions are pinned in `go.mod`/`go.sum`; `go.sum` integrity is enforced by the
  toolchain.
- Builds are static (`CGO_ENABLED=0`) to avoid dynamic-linking surprises on target systems.
- Dependency updates are reviewed, not auto-merged blindly.

---

## 8. Distribution integrity

- GitHub Release binaries are published with checksums (e.g. `SHA256SUMS`).
- The AUR `PKGBUILD` pins release artifacts by `sha256sums`, so `omadev-bin` installs verify the
  downloaded binary (S11).
- No affiliation with Omarchy is implied; packaging is standard Arch-native.

---

## 9. Debug output safety

`--debug` may reveal detectors run, files inspected, findings, confidence reasoning, the plan, and
the underlying commands — because those are exactly what a developer needs to trust the tool. It
must never reveal secret values, resolved `.env` interpolation, or credentials embedded in a
Compose file. This is a review checkpoint for any change that adds debug logging.

---

## 10. Reporting security issues

Security reports should go to the project maintainer privately rather than via a public issue,
until a fix is available. A `SECURITY.md` reporting section / contact will accompany the first
public release (M2).
