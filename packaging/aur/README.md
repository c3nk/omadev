# AUR packaging

Two packages let Omarchy/Arch users install Omadev via `yay -S` (or the Omarchy
"Install > AUR" flow):

- `omadev-bin` — installs the prebuilt release binary (fast; recommended).
- `omadev` — builds from the release source tarball (needs Go).

## Release-time steps

At each release, for the package being published:

1. Bump `pkgver` (and reset `pkgrel=1`).
2. Replace the `SKIP` checksums with the real `sha256` values:
   - `omadev-bin`: take them from the release `SHA256SUMS` file.
   - `omadev`: `sha256sum` of the release source tarball.
3. Regenerate `.SRCINFO`:
   ```bash
   makepkg --printsrcinfo > .SRCINFO
   ```
4. Push to the AUR (`ssh aur@aur.archlinux.org`) repository for the package.

Omadev is an independent tool designed to work well on Omarchy; publishing to the AUR
does not imply any official affiliation.
