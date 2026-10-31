# Building

Local app for this computer. No certificates.

Output: `dist/v<version>_<timestamp>/`, with `LICENSE` and `checksums.txt`. Example: `dist/v0.1.0_261027220841/`. `dist/` is gitignored.

`build/` holds icons, plists, and installer config. `bin/` is Wails scratch, gitignored. `frontend/dist` is the Vite build embedded in the binary.

`go build` has no window.

Packages for other people, signing, and the GitHub Release workflow: [Releasing](releasing.md).

## Commands

Development:

```sh
wails3 dev
```

`.env` is optional. Copy `.env.sample` to `.env` to change the app log. Defaults: console debug, rotated debug file. `dev_mode.log_level` in `build/config.yml` is the Wails watcher.

Local packages for this machine:

```sh
./scripts/build.sh build
```

On Windows, PowerShell and Command Prompt do not run `.sh` files. Windows asks which program should open the file, and the build never starts. Use the launcher instead. It runs `scripts/build.sh` with Git Bash:

```bat
scripts\build.cmd build
```

Targets:

```sh
./scripts/build.sh build --platforms darwin/arm64
./scripts/build.sh build --platforms darwin/arm64,windows/amd64
```

On Windows, pass the same arguments to `scripts\build.cmd`.

`darwin/arm64`, `darwin/amd64`, `windows/amd64`, `windows/arm64`, `linux/amd64`, `linux/arm64`.

Linux builds on Linux, for the host architecture. On macOS the script skips Linux. No Docker cross-compile. Linux packages for other people come from the [Release workflow](releasing.md).

## Local build

| Host | Output |
| --- | --- |
| Mac | `macos/Boilerplate.app`, `windows/boilerplate-v0.1.0-amd64.exe` |
| Linux | `linux/boilerplate-v0.1.0-amd64` |
| Windows | `windows/boilerplate-v0.1.0-<arch>.exe` |

`Boilerplate` is `info.productName`. `boilerplate` is `APP_NAME`. `0.1.0` is `info.version`. The build copies those into the Windows manifest, the Mac bundle, and the Linux packages.

Local Mac `.app`: ad-hoc signature. It opens on this Mac. Gatekeeper blocks it on other Macs.

Local Windows `.exe`: unsigned. SmartScreen warns. Requires WebView2, which current Windows 11 includes.
