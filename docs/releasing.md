# Releasing

`wails3 dev` and `./scripts/build.sh build` make an app you can open on this computer. [Building](building.md) covers those. They need no certificates.

The files other people install come from `./scripts/build.sh build --dist`. Run that through the Release workflow. A local `--dist` is the same script. Use it on a computer that already has the signing material, or when Actions is unavailable.

`--dist` checks certificates and tools before it compiles. A missing Apple notary profile, a missing `.pfx`, or `signtool` off PATH is that check. The log names each missing piece and then stops. The unsigned build is still `./scripts/build.sh build`.

## Release workflow

File: `.github/workflows/release.yml`.

You start it from the Actions tab. Pushing `main` leaves it idle. Pushing a tag leaves it idle. A clone is enough for `wails3 dev` and a local build.

1. **Actions**
2. **Release**
3. **Run workflow**. Branch: usually `main`.

| Job | Runner | Command | Output |
| --- | --- | --- | --- |
| macOS | `macos-latest` | `--dist --platforms darwin/arm64` | Notarized DMG. Apple silicon. |
| Linux and Windows | `ubuntu-latest` | `--dist --platforms linux/amd64,windows/amd64` | `.deb`, `.rpm`, `.AppImage`, Windows `.exe`, NSIS installer |

Artifacts are kept for 14 days. The `.app` is an Actions artifact. The Mac download is the DMG.

**Publish a GitHub Release named from info.version** is off by default. Turn it on to publish tag `v` + `info.version`, for example `v0.1.0`. Uploaded: `.dmg`, Windows files, `.deb`, `.rpm`, `.AppImage`, and `checksums.txt`.

The workflow builds Apple silicon, Windows x64, and Linux x64. `windows/arm64` and `linux/arm64` are local targets. See [Local release](#local-release).

### Secrets

**Settings, Secrets and variables, Actions**. An organization can set the same names once. Later repos inherit them.

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

A missing Apple secret fails the macOS job. A missing Windows secret fails the Linux and Windows job. The log names the secret. `LINUX_PGP_KEY` can stay empty: a `.deb` and an `.rpm` install without a signature.

Encode a certificate for the secret:

```sh
base64 < "path/to/certificate.p12" | pbcopy
```

On Linux, `base64 -w 0` prints one line. Paste that line into the secret. Keep the file itself outside the repository.

### Minutes

Public repositories: standard runners are free. Private repositories share 2,000 Actions minutes per month across the account. A macOS minute costs about ten Linux minutes. One signed release is about 200 to 300 minutes. The workflow stays manual so a push to `main` does not spend them. Runners: `macos-latest` and `ubuntu-latest`. Larger runners are billed on public repositories too.

When the minutes run out, hosted jobs stop. A [local release](#local-release) on a Mac still signs from `build/dist.env`. Linux packages then need a Linux machine, or the next month's quota.

## Packages

`Boilerplate` is `info.productName`. `boilerplate` is `APP_NAME`. `0.1.0` is `info.version`.

| File | Use |
| --- | --- |
| `Boilerplate.dmg` | macOS download. Drag to Applications. |
| `boilerplate-v0.1.0-amd64.exe` | Portable. WebView2 must already be installed. |
| `boilerplate-v0.1.0-amd64-setup.exe` | NSIS installer. Installs WebView2 when it is missing. |
| `boilerplate_0.1.0_amd64.deb` | Debian, Ubuntu (`apt` / `dpkg`) |
| `boilerplate-0.1.0-amd64.rpm` | Fedora (`dnf`). Depends on `gtk4` and `webkitgtk6.0`. |
| `boilerplate-v0.1.0-amd64.AppImage` | Other Linux |

Mac release: Developer ID, hardened runtime, notarization ticket on the `.app` and the `.dmg`. App Sandbox is off. This is a direct download, not a Mac App Store build. Gatekeeper reports `App sandbox: not enabled`. Without the ticket, another Mac blocks the download until someone allows it in System Settings.

Windows release: the `.exe` is signed, the NSIS installer is built around that file, then the installer is signed. SmartScreen checks the download and the program the installer writes. A new standard certificate can still show the warning until that publisher has reputation. EV or Azure Trusted Signing is faster. A USB token or Azure Trusted Signing needs `signtool` on Windows.

Linux release: unsigned `.deb` and `.rpm` packages install and launch. A PGP signature is a checksum a person can verify. The `.rpm` targets Fedora. RHEL and clones do not ship WebKitGTK 6.0 in the default repos. openSUSE uses different package names. Use the AppImage there.

Opening a `.deb` in GNOME Software reads the project website and the license from the package control data. Those are `info.homepage` and `info.license`. The `.deb` and `.rpm` also install AppStream metadata written by `scripts/write-linux-metainfo.py` from `info.license`, `info.description`, `info.version`, and `info.homepage`. After the package is installed, Software uses that file for the license, the website, and the release notes. The version row before install still says "No details for this release", because that screen only has the package version.

GTK 4 and WebKitGTK 6.0: see the [README](../README.md).

Arch: `wails3 task linux:create:aur`. That task is separate from `--dist`.

## Local release

Use this when the Release workflow is unavailable. The script is the one the workflow runs.

```sh
cp build/dist.env.example build/dist.env
```

`build/dist.env` is gitignored. `./scripts/build.sh build` ignores it. Environment variables already set override the file.

On Windows, PowerShell and Command Prompt do not run `.sh` files. Use the launcher:

```bat
scripts\build.cmd build --dist
```

`--dist` with no `--platforms` signs whatever this computer was going to build. On a Mac that is the Mac app and a Windows exe, so the check asks for an Apple notary profile and a Windows `.pfx` before it compiles anything.

| Host | Default `--dist` | The check requires |
| --- | --- | --- |
| Mac | `darwin/<this Mac>`, `windows/amd64` | Developer ID, notary profile, `osslsigncode`, `makensis`, `WINDOWS_CERT_PATH` |
| Linux | `linux/<this machine>` | GTK, WebKit, `rpm`. A PGP key is optional. |
| Windows | `windows/<this PC>` | `signtool`, `makensis`, `WINDOWS_CERT_PATH` |

Linux is skipped on a Mac. macOS is skipped on Linux and Windows. Limit the check to one side:

```sh
./scripts/build.sh build --dist --platforms darwin/arm64
./scripts/build.sh build --dist --platforms windows/amd64
```

On Windows, pass the same arguments to `scripts\build.cmd`. An arm64 PC writes `arm64` in the Windows file names. An arm64 Linux machine writes `arm64` in the package names. Pass `linux/arm64` or `windows/arm64` explicitly when the default is the other architecture.

`darwin/arm64`, `darwin/amd64`, `windows/amd64`, `windows/arm64`, `linux/amd64`, `linux/arm64`.

### On a Mac

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

A Mac `--dist` that includes Windows (the default) also needs the Windows certificate and:

```sh
brew install makensis osslsigncode
```

The script signs that Windows build with `osslsigncode`.

Output:

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

`--platforms darwin/arm64` leaves out the `windows/` directory and skips the Windows check.

### On Windows

`scripts\build.cmd build` writes an unsigned exe. `scripts\build.cmd build --dist` writes the signed exe and the NSIS installer. Use the second command to test the installer on a Windows PC, or to produce the files you ship.

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

A purchased Authenticode `.pfx` uses the same two variables.

Timestamp server default: `http://timestamp.digicert.com`. Signing contacts that server, so `--dist` needs a network connection.

Output:

```text
dist/v0.1.0_261027220841/
  LICENSE
  checksums.txt
  windows/
    boilerplate-v0.1.0-amd64.exe
    boilerplate-v0.1.0-amd64-setup.exe
```

The script builds the Windows app for this PC. It skips macOS and Linux.

### On Linux

`--dist` builds `.deb`, `.rpm`, and `.AppImage` for this machine's architecture. Version: `info.version` (`X.Y.Z`).

No certificate is required. `LINUX_PGP_KEY` in `build/dist.env` is an optional exported private key. `LINUX_PGP_PASSWORD` is the password for that key, when the key has one.

The libraries are the ones in the [README](../README.md). The package step also needs `rpm`, and FUSE libraries for the AppImage (`libfuse2` on Debian and Ubuntu). Adding Windows to `--platforms` on Linux also needs `nsis`, `osslsigncode`, and `WINDOWS_CERT_PATH`. The Release workflow installs that set on Ubuntu.

Output:

```text
dist/v0.1.0_261027220841/
  LICENSE
  checksums.txt
  linux/
    boilerplate_0.1.0_amd64.deb
    boilerplate-0.1.0-amd64.rpm
    boilerplate-v0.1.0-amd64.AppImage
```
