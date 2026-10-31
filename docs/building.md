# Building and distributing

Output: `dist/v<version>_<timestamp>/`, with `LICENSE` and `checksums.txt`. Example: `dist/v0.1.0_261027220841/`. `dist/` is gitignored.

`build/` holds icons, plists, and installer config. `bin/` is Wails scratch, gitignored. `frontend/dist` is the Vite build embedded in the binary.

`go build` has no window.

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

Signed release packages:

```sh
./scripts/build.sh build --dist
```

```bat
scripts\build.cmd build --dist
```

Targets:

```sh
./scripts/build.sh build --platforms darwin/arm64
./scripts/build.sh build --platforms darwin/arm64,windows/amd64
./scripts/build.sh build --dist --platforms darwin/arm64,windows/amd64
```

On Windows, pass the same arguments to `scripts\build.cmd`.

`darwin/arm64`, `darwin/amd64`, `windows/amd64`, `windows/arm64`, `linux/amd64`, `linux/arm64`.

Linux builds on Linux, for the host architecture. On macOS the script skips Linux. No Docker cross-compile. Use a Linux machine, or the GitHub Release workflow.

## Local build

| Host | Output |
| --- | --- |
| Mac | `macos/Boilerplate.app`, `windows/boilerplate-v0.1.0-amd64.exe` |
| Linux | `linux/boilerplate-v0.1.0-amd64` |
| Windows | `windows/boilerplate-v0.1.0-<arch>.exe` |

`Boilerplate` is `info.productName`. `boilerplate` is `APP_NAME`. `0.1.0` is `info.version`. The build copies those into the Windows manifest, the Mac bundle, and the Linux packages.

Local Mac `.app`: ad-hoc signature. Gatekeeper blocks it on other Macs.

Local Windows `.exe`: unsigned. SmartScreen warns. Requires WebView2, which current Windows 11 includes. `scripts\build.cmd build --dist` is the signed exe and the NSIS installer, built on Windows. See [Signing](#signing).

## Release

`--dist` checks certificates and tools before compiling.

macOS host:

```text
dist/v0.1.0_261027220841/
  LICENSE
  checksums.txt
  macos/
    Boilerplate.app
    Boilerplate.dmg
  windows/
    boilerplate-v0.1.0-amd64.exe
    boilerplate-v0.1.0-amd64-setup.exe
```

Windows host:

```text
dist/v0.1.0_261027220841/
  LICENSE
  checksums.txt
  windows/
    boilerplate-v0.1.0-amd64.exe
    boilerplate-v0.1.0-amd64-setup.exe
```

An arm64 PC writes `arm64` in those file names. The script builds the Windows app for this PC. It skips macOS and Linux.

Linux host:

```text
dist/v0.1.0_261027220841/
  LICENSE
  checksums.txt
  linux/
    boilerplate_0.1.0_amd64.deb
    boilerplate-0.1.0-amd64.rpm
    boilerplate-v0.1.0-amd64.AppImage
```

| File | Use |
| --- | --- |
| `macos/Boilerplate.dmg` | macOS download. Drag to Applications. |
| `windows/boilerplate-v0.1.0-amd64.exe` | Portable. WebView2 must already be installed. |
| `windows/boilerplate-v0.1.0-amd64-setup.exe` | NSIS installer. Installs WebView2 when it is missing. |
| `linux/boilerplate_0.1.0_amd64.deb` | Debian, Ubuntu (`apt` / `dpkg`) |
| `linux/boilerplate-0.1.0-amd64.rpm` | Fedora (`dnf`). Depends on `gtk4` and `webkitgtk6.0`. |
| `linux/boilerplate-v0.1.0-amd64.AppImage` | Other Linux |

Mac release: Developer ID, hardened runtime, notarization ticket on the `.app` and the `.dmg`. App Sandbox is off. This is a direct download, not a Mac App Store build. Gatekeeper reports `App sandbox: not enabled`.

The `.rpm` targets Fedora. RHEL and clones do not ship WebKitGTK 6.0 in the default repos. openSUSE uses different package names. Use the AppImage there.

Arch: `wails3 task linux:create:aur`. Not part of `--dist`.

## Signing

```sh
cp build/dist.env.example build/dist.env
```

`build/dist.env` is gitignored. `./scripts/build.sh build` ignores it. Environment variables already set override the file.

### macOS

Apple Developer Program. **Developer ID Application** certificate in the login keychain.

```sh
security find-identity -v -p codesigning
```

Create an app-specific password at [appleid.apple.com](https://appleid.apple.com). Store the notary credentials once. The profile name is the value in `build/dist.env`.

```sh
xcrun notarytool store-credentials "your-notary-profile" \
  --apple-id "you@example.com" \
  --team-id "TEAMID" \
  --password "app-specific-password"
```

```sh
APPLE_NOTARY_KEYCHAIN_PROFILE="your-notary-profile"
```

Set `APPLE_SIGNING_IDENTITY` when more than one Developer ID is installed. With one certificate, the script uses that one.

`codesign` and `notarytool` ship with the Xcode command line tools. Notarization needs a network connection and takes a few minutes. The script adds a secure timestamp.

### Windows

On Windows, `scripts\build.cmd build` writes an unsigned exe. `scripts\build.cmd build --dist` writes the signed exe and the NSIS installer. Use the second command to test the installer on a Windows PC, or to produce the files you ship.

`--dist` stops before compiling until `signtool` and `makensis` are on PATH and `build/dist.env` points at a `.pfx`. Open a new terminal after changing PATH. Git Bash reads the Windows PATH.

`signtool` comes with the Windows SDK. The installer leaves it off PATH. In the Visual Studio Installer, add **Windows SDK** or **Windows SDK Signing Tools for Desktop Apps**. The tool is `C:\Program Files (x86)\Windows Kits\10\bin\<version>\x64\signtool.exe`. Add that `x64` folder to PATH. `where signtool` prints the path.

`makensis` is [NSIS](https://nsis.sourceforge.io/).

```bat
winget install NSIS.NSIS
```

`where makensis` prints the path.

Copy `build/dist.env.example` to `build/dist.env`. Keep the `.pfx` outside the repository.

```sh
WINDOWS_CERT_PATH="C:\Users\you\certificate.pfx"
WINDOWS_CERT_PASSWORD="the-password"
```

A self-signed `.pfx` is enough to run the signed installer on this PC. SmartScreen still warns. Create one in PowerShell:

```powershell
$cert = New-SelfSignedCertificate -Type CodeSigningCert -Subject "CN=Dev Code Signing" -CertStoreLocation Cert:\CurrentUser\My
$pwd = ConvertTo-SecureString -String "the-password" -Force -AsPlainText
Export-PfxCertificate -Cert $cert -FilePath "$env:USERPROFILE\dev-codesign.pfx" -Password $pwd
```

A purchased Authenticode `.pfx` uses the same two variables. A new standard certificate can still trigger SmartScreen until that publisher has reputation. EV or Azure Trusted Signing is faster. A USB token or Azure Trusted Signing needs `signtool` on Windows.

Timestamp server default: `http://timestamp.digicert.com`. Signing contacts that server, so `--dist` needs a network connection.

Order: sign the `.exe`, build the NSIS installer around that file, sign the installer. SmartScreen checks the download and the program the installer writes.

On a Mac, the same `WINDOWS_CERT_PATH` and `WINDOWS_CERT_PASSWORD` apply. The script signs with `osslsigncode` instead of `signtool`:

```sh
brew install makensis osslsigncode
```

On Linux, install `nsis` and `osslsigncode`. The Release workflow does this on Ubuntu.

### Linux

`--dist` builds `.deb`, `.rpm`, and `.AppImage` for the host architecture. Version: `info.version` (`X.Y.Z`).

Opening a `.deb` in GNOME Software reads the project website and the license from the package control data. Those are `info.homepage` and `info.license`. The Wails template used `https://wails.io`; the package now uses the project URL.

The `.deb` and `.rpm` also install AppStream metadata written by `scripts/write-linux-metainfo.py` from `info.license`, `info.description`, `info.version`, and `info.homepage`. After the package is installed, Software uses that file for the license, the website, and the release notes. The version row you see before install still says "No details for this release", because that screen only has the package version.

`LINUX_PGP_KEY`: optional exported private key. Unsigned `.deb` and `.rpm` packages install.

GTK 4 and WebKitGTK 6.0: see the [README](../README.md).

## GitHub Actions

Manual. A push to `main` does not run it. A tag does not run it. A clone is enough for `wails3 dev` and a local build.

1. **Actions**
2. **Release**
3. **Run workflow**. Branch: usually `main`.

macOS runner: notarized DMG. Ubuntu runner: Linux packages, Windows `.exe`, NSIS installer. Artifacts are kept for 14 days.

**Publish a GitHub Release named from info.version** is off by default. Turn it on to publish tag `v` + `info.version`, for example `v0.1.0`. Uploaded: `.dmg`, Windows files, `.deb`, `.rpm`, `.AppImage`. The `.app` is an Actions artifact only. The Mac download is the DMG.

Secrets: **Settings, Secrets and variables, Actions**. An organization can set the same names once. Later repos inherit them.

| Secret | Value |
| --- | --- |
| `APPLE_CERTIFICATE_BASE64` | Developer ID `.p12`, base64 |
| `APPLE_CERTIFICATE_PASSWORD` | `.p12` password |
| `APPLE_SIGNING_IDENTITY` | `Developer ID Application: Your Name (TEAMID)` |
| `APPLE_ID` | Apple ID |
| `APPLE_TEAM_ID` | Team ID |
| `APPLE_APP_SPECIFIC_PASSWORD` | App-specific password |
| `WINDOWS_CERTIFICATE_BASE64` | Authenticode `.pfx`, base64 |
| `WINDOWS_CERTIFICATE_PASSWORD` | `.pfx` password |
| `LINUX_PGP_KEY` | Optional. Packages are published without it. |
| `LINUX_PGP_PASSWORD` | Optional password for that key. |

A missing Apple or Windows secret fails the job and points at `./scripts/build.sh build --dist`.

Public repositories: standard runners are free. Private repositories share 2,000 Actions minutes per month across the account. A macOS minute costs about ten Linux minutes. One signed release is about 200 to 300 minutes. The workflow is manual so a push to `main` does not spend them. Runners: `macos-latest` and `ubuntu-latest`. Larger runners are billed on public repositories too.

When the minutes run out, hosted jobs stop. `./scripts/build.sh build --dist` on a Mac still signs from `build/dist.env`. Linux packages then need a Linux machine, or the next month's quota.

## Setup

| Goal | Also required |
| --- | --- |
| `wails3 dev`, `./scripts/build.sh build` | Go, Node, Wails CLI. No `build/dist.env`. No Actions secrets. On Windows: `scripts\build.cmd build`. |
| `./scripts/build.sh build --dist` | `build/dist.env`. Linux packages need a Linux machine. On Windows: `scripts\build.cmd build --dist`, plus `signtool`, NSIS, and `WINDOWS_CERT_PATH`. |
| Actions, including Linux from a Mac | Secrets above. The workflow file is already in the repo. Renaming the app does not require a YAML edit. |
