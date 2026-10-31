# Build directory

Packager files for `wails3 build` and `wails3 package` (`Taskfile.yml`).

- `config.yml` - name, version, dev watcher
- `appicon.png` - source icon. Build output: `darwin/icons.icns`, `windows/icon.ico`
- `darwin` - plist, entitlements, Mac tasks. Minimum OS 14. A build fills the bundle name, id, and version from `config.yml`. App Sandbox off (direct download).
- `windows` - manifest, version info, NSIS. A build fills these from `config.yml`. MSIX packaging reads `config.yml`. `windows/msix` is not an input.
- `linux` - package metadata, including the AppStream metainfo. A build fills the name, version, homepage, and license from `Taskfile.yml` and `config.yml`.
- `dist.env.example` - signing values for `--dist`. Copy to `dist.env` (gitignored)

`bin/` at the repo root: Wails intermediates, gitignored.
`dist/` at the repo root: `scripts/build.sh` output, gitignored. On Windows, run `scripts\build.cmd`.

Details: [Building and distributing](../docs/building.md).
