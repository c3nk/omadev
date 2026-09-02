# Omadev — Omarchy Compatibility Policy

> How Omadev stays compatible with Omarchy as Omarchy evolves, without becoming coupled to any
> single version of it. Omadev is an independent tool designed to work well on Omarchy; it is not
> an official Omarchy product and implies no affiliation.

## Stance: Omarchy-first, not Omarchy-coupled

Omadev targets Omarchy first and runs on other Linux systems where practical. It does not depend on
Omarchy internals. As of v0.1 there is essentially **no Omarchy-specific code** — the tool is
generic Linux + Docker — so Omarchy can change without breaking Omadev's core. Coupling is
introduced deliberately and only where it earns its place (see "Deep integration (M6)").

## Core principle: detect at runtime, never hardcode

The primary way Omadev absorbs Omarchy's changes is by **inspecting the actual environment rather
than assuming its state**:

- **Docker privilege** is probed with `docker info` at runtime (D4), so a change in Omarchy's
  Docker setup (rootless, group membership, socket permissions) is handled automatically — nothing
  is hardcoded and the `docker` group is never modified.
- **Tool availability** (Docker Compose v2, and later mise/runtimes) is checked before use, not
  assumed.
- **Detection** is driven by repository evidence, not by the host distribution.

Because Omadev looks instead of assuming, most Omarchy changes need no corresponding change here.

## Namespace policy

The executable stays **`omadev`**, a standalone binary. Omadev never uses the `omarchy dev`
namespace (which Omarchy may use or expand for its own commands). This avoids collision if Omarchy
adds its own developer commands.

## Delegate to Omarchy tools; do not duplicate them

Omadev does not reimplement what Omarchy already provides (its package TUI, its `omarchy-*`
helpers, its system management). Where responsibilities overlap, Omadev delegates to the host tool.
When Omarchy adds a capability, the goal is to **discover and use it**, not copy it.

## Deep integration (M6) is isolated and capability-gated

The real Omarchy coupling arrives in M6 (notifications, Hyprland, Quickshell, launchers). Two rules
contain the drift risk:

1. **Isolation** — all Omarchy-specific behavior lives in one place (e.g. an `internal/omarchy/`
   package), never scattered through detection or execution (§34.14).
2. **Capability detection** — each integration is gated on a runtime check ("Is Hyprland running?
   Is Quickshell present?"). If the capability is absent — an older Omarchy, or a non-Omarchy Linux
   — Omadev skips it silently and keeps working. Integrations degrade, they do not break.

## Version targeting

- Omarchy is rolling (Arch-based), so Omadev targets the current **Omarchy stable channel** as the
  primary reference, and avoids assumptions that would break on the `rc`/`edge` channels.
- Minimum Go is 1.24; on rolling Arch/Omarchy this is not a practical constraint.
- Distribution is via GitHub Releases + AUR (`omadev-bin`/`omadev`). Inclusion in Omarchy's OPR or
  default package set is an upstream maintainer decision, not something Omadev asserts (S11).

## What to watch upstream

Changes in Omarchy that could require action here:

| Omarchy change | Effect on Omadev |
|----------------|------------------|
| Adds an `omarchy dev` command | Namespace — confirm no collision (we stay `omadev`) |
| Changes Docker setup (rootless/group/socket) | Absorbed by runtime privilege detection (D4); verify) |
| Changes OPR / packaging conventions | Distribution path; update `packaging/` if we pursue OPR |
| Changes Hyprland/Quickshell/notification interfaces | M6 integrations — update the isolated `internal/omarchy/` layer |
| Ships an overlapping tool | Prefer delegation; drop any now-duplicated behavior |

## Testing for compatibility

- Unit and integration tests run on generic Linux (CI), which covers the core.
- Before an Omarchy-affecting change (especially M6), validate on an Omarchy install (VM or
  container) on the stable channel.
- Keep the Omarchy-specific surface small enough that this manual validation stays cheap.
