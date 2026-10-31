# Boilerplate

A boilerplate for production ready cross-platform desktop applications, written in Go with Wails and a Vue, TypeScript, and Tailwind frontend. One build script produces signed, packaged applications for macOS (`.app` and `.dmg`), Windows (`.exe` and installer), and Linux (`.deb`, `.rpm`, and AppImage), locally or from the GitHub Release workflow.

(Todo: Needs more verification testing in the wild for signed builds)

Example **WindowFramelessMenu** on macOS

![WindowFramelessMenu on macOS](docs/screenshot.png)

Every window mode on macOS, Windows, and Linux: [Window modes](docs/window-modes.md).

## Stack

- [Go](https://go.dev)
- [Wails v3](https://v3.wails.io)
- [Vue 3](https://vuejs.org) (`<script setup>`)
- [TypeScript](https://www.typescriptlang.org)
- [Vite](https://vite.dev)
- [Tailwind CSS v4](https://tailwindcss.com)
- [zerolog](https://github.com/rs/zerolog) (`prezerolog`), rotated logs in `internal/applog`

## Requirements

- Go 1.26+
- Node.js 24+ with npm
- Wails v3 CLI
- macOS 14+ to run the Mac app

### Linux libraries

On Linux, install these before the Wails CLI. The CLI build needs them.

Debian/Ubuntu:

```sh
sudo apt install build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev
```

Fedora: `gtk4-devel`, `webkitgtk6.0-devel`. Arch: `gtk4`, `webkitgtk-6.0`.

### Wails CLI

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
```

Binary: `$(go env GOPATH)/bin`.

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

Check with `wails3 version`. The v2 `wails` binary is separate.

## Build

Development:

```sh
wails3 dev
```

Build on macOS or Linux:

```sh
./scripts/build.sh build
```

Build on Windows, from PowerShell or Command Prompt:

```bat
scripts\build.cmd build
```

With no `--platforms` flag, the script builds the default for this computer. macOS builds the Mac app and a Windows exe. Linux builds the Linux binary. Windows builds the Windows exe.

To build other platforms, pass a comma-separated list. On Linux, the Linux app and a Windows exe:

```sh
./scripts/build.sh build --platforms linux/amd64,windows/amd64
```

On an arm64 Linux machine, use `linux/arm64` in that list. The Mac app builds on macOS. Linux packages build on Linux, for this computer's architecture. On Windows, pass the same list to `scripts\build.cmd`.

Packages for other people: [Releasing](docs/releasing.md). That is the GitHub Release workflow. It runs this script with `--dist` on the runners. A local `--dist` is the same command when Actions is unavailable, and it stops until the certificates and tools for that computer are present.

## Test

```sh
go test ./...
```

## Layout

- `main.go` - window bootstrap
- `app.go` - bound methods: `Version`, `Greet`, `Frameless`, `ContextMenu`
- `config.go` - name, description, and version from `build/config.yml`
- `internal/window` - window template, native menu, right-click mode
- `internal/inspector` - dev-only web inspector
- `internal/applog` - process log
- `frontend/` - Vue app
- `Taskfile.yml` - Wails tasks
- `scripts/build.sh` - local build and `--dist`. On Windows: `scripts\build.cmd`
- `docs/building.md` - local build
- `docs/releasing.md` - signed packages and the Release workflow
- `docs/window-modes.md` - window screenshots per OS

## Window

`template` in `internal/window/window.go`. Current value: `WindowFramelessMenu`. Screenshots: [Window modes](docs/window-modes.md).

| Constant | Result |
| --- | --- |
| `WindowFrame` | System frame. No menu. |
| `WindowFrameMenu` | System frame and native menu. |
| `WindowFramelessMenu` | Custom toolbar. macOS: native menu in the screen menu bar. Windows and Linux: minimise, maximise, and close in the toolbar. |
| `WindowFrameless` | Custom toolbar. No native menu. |

Frameless windows draw `frontend/src/Toolbar.vue`. macOS traffic lights are in that toolbar. On Windows and Linux the menu bar is part of the system frame, so `WindowFramelessMenu` and `WindowFrameless` both show the toolbar and no menu bar.

`internal/window/menu.go`: macOS gets App, Edit, and Window. Windows and Linux get Edit and Window. On Windows the menu bar follows the system theme.

## Right-click

`contextMenu` in `internal/window/contextmenu.go`. Current value: `ContextMenuDefault`. `frontend/src/contextMenu.ts` applies it in dev and release builds.

| Constant | Result |
| --- | --- |
| `ContextMenuDefault` | Wails behaviour. Release: the webview menu on inputs, editable content, and selected text only. Dev: the webview menu everywhere, with Inspect Element. |
| `ContextMenuDisabled` | No right-click anywhere, inputs included. |
| `ContextMenuCustom` | No webview menu. The event still reaches Vue `@contextmenu` handlers and Wails `--custom-contextmenu` menus, so the app draws its own. |

Dev builds open the web inspector with `Cmd+Option+Shift+I` on macOS and `Ctrl+Shift+Alt+I` on Windows and Linux, in every mode. Release builds (`-tags production`) have no inspector and no shortcut. On macOS, `internal/inspector/darwin_dev.go` adds Inspect Element and the shortcut with private WebKit calls. It is not compiled into release builds.

## Configuration

- `build/config.yml` `info.productName` - window title and UI heading
- `build/config.yml` `info.description` - app description and UI subtitle
- `build/config.yml` `info.version` - UI and About (`X.Y.Z`)
- `build/config.yml` `info.homepage` - project website in the Linux packages
- `build/config.yml` `info.license` - SPDX license id in the Linux packages, such as `MIT`
- `Taskfile.yml` `APP_NAME` - binary and package name
- `build/config.yml` `dev_mode.log_level` - Wails watcher
- `.env` `LOG_CONSOLE_LEVEL`, `LOG_FILE_LEVEL` - app log. Defaults are in `.env.sample`

`config.go` and the frontend load `productName`, `description`, and `version` from `build/config.yml`. A build copies the `info` fields into the Windows manifest, version info, NSIS defines, Mac plists, and Linux package file. `APP_NAME` in `Taskfile.yml` is the binary name. That copy writes it into the executable name, the package name, and the installer. `wails3 tool msix` reads `build/config.yml` when it packages. `build/windows/msix` is not an input.

`frontend/package.json` is the Vite package. Its name and version are separate from the app.

## Rename

1. `build/config.yml`: `info.productName`, `info.description`, `info.companyName`, `info.copyright`, `info.comments`, `info.version`, `info.productIdentifier`, `info.homepage`, `info.license`.
2. `APP_NAME` in `Taskfile.yml`, when the binary name should change.
3. Module path in `go.mod`, and `boilerplate/...` imports.

The next build writes the `config.yml` fields and `APP_NAME` into the platform files. `LSMinimumSystemVersion` stays `14.0.0`. `wails3 update build-assets` sets the Mac minimum back to 12 and drops the Linux AppStream entry.

Icon: `build/appicon.png`.
