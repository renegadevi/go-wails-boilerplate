#!/usr/bin/env bash
# build produces a local app. On a Mac that is an ad-hoc signed .app, which
# opens on this computer and is rejected by Gatekeeper anywhere else.
# build --dist produces the files you ship: a notarized macOS DMG, a signed
# Windows exe and NSIS installer, and on Linux a deb, an rpm, and an AppImage.
# --dist reads build/dist.env and stops before compiling when signing
# credentials or tools are missing. A normal build does not use them.

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

die() {
  echo "error: $*" >&2
  exit 1
}

read_yaml_value() {
  local file="$1"
  local pattern="$2"
  python3 - "$file" "$pattern" <<'PY'
import re, sys
file, pattern = sys.argv[1], sys.argv[2]
text = open(file, encoding="utf-8").read()
match = re.search(pattern, text)
print(match.group(1) if match else "")
PY
}

load_dist_env() {
  local file="$root/build/dist.env"
  [[ -f "$file" ]] || return 0
  local line key value
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line#"${line%%[![:space:]]*}"}"
    [[ -z "$line" || "$line" == \#* ]] && continue
    case "$line" in
      export\ *) line="${line#export }" ;;
    esac
    [[ "$line" == *=* ]] || continue
    key="${line%%=*}"
    value="${line#*=}"
    case "$key" in
      APPLE_SIGNING_IDENTITY|APPLE_NOTARY_KEYCHAIN_PROFILE|WINDOWS_CERT_PATH|WINDOWS_CERT_PASSWORD|WINDOWS_TIMESTAMP_URL|LINUX_PGP_KEY|LINUX_PGP_PASSWORD)
        ;;
      *)
        continue
        ;;
    esac
    if [[ "$value" == \"*\" && "$value" == *\" ]]; then
      value="${value#\"}"
      value="${value%\"}"
    elif [[ "$value" == \'*\' && "$value" == *\' ]]; then
      value="${value#\'}"
      value="${value%\'}"
    fi
    if [[ -z "${!key:-}" ]]; then
      export "$key=$value"
    fi
  done < "$file"
}

host_os="$(uname -s)"
case "$host_os" in
  Darwin) host_os="darwin" ;;
  Linux) host_os="linux" ;;
  MINGW*|MSYS*|CYGWIN*) host_os="windows" ;;
  *) die "unsupported host $(uname -s)" ;;
esac

host_arch="$(uname -m)"
case "$host_arch" in
  x86_64|amd64) host_arch="amd64" ;;
  arm64|aarch64) host_arch="arm64" ;;
  *) die "unsupported architecture $(uname -m)" ;;
esac

default_platforms() {
  case "$host_os" in
    darwin) printf '%s\n' "darwin/${host_arch}" "windows/amd64" ;;
    linux) printf '%s\n' "linux/${host_arch}" ;;
    windows) printf '%s\n' "windows/${host_arch}" ;;
  esac
}

usage() {
  cat <<EOF
Usage: $0 build [--dist] [--platforms os/arch,os/arch]

  build       Local binaries. The Mac app is ad-hoc signed.
  --dist      Signed packages. Fails if certificates or tools are missing.
  --platforms Comma-separated list. Defaults to what this computer can build.

Examples:
  $0 build
  $0 build --platforms darwin/arm64
  $0 build --dist
  $0 build --dist --platforms darwin/arm64,windows/amd64
EOF
}

dist=false
platforms_arg=""
if [[ $# -eq 0 ]]; then
  usage >&2
  exit 1
fi
if [[ "$1" != "build" ]]; then
  usage >&2
  exit 1
fi
shift
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dist)
      dist=true
      shift
      ;;
    --platforms)
      [[ $# -ge 2 ]] || die "--platforms needs a value"
      platforms_arg="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac
done

load_dist_env

app_name="$(read_yaml_value Taskfile.yml '(?m)^  APP_NAME: \"([^\"]+)\"')"
app_version="$(read_yaml_value build/config.yml '(?m)^  version: \"([^\"]+)\"')"
product_name="$(read_yaml_value build/config.yml '(?m)^  productName: \"([^\"]+)\"')"
[[ -n "$app_name" && -n "$app_version" && -n "$product_name" ]] || die "APP_NAME, info.productName, and info.version must be set"
[[ "$product_name" != */* ]] || die "info.productName cannot contain a slash"

if command -v wails3 >/dev/null 2>&1; then
  wails3="$(command -v wails3)"
else
  wails3="$(go env GOPATH)/bin/wails3"
fi
[[ -x "$wails3" ]] || die "wails3 is not installed. Run: go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0"

requested=()
if [[ -n "$platforms_arg" ]]; then
  old_ifs="$IFS"
  IFS=','
  # shellcheck disable=SC2086
  set -- $platforms_arg
  IFS="$old_ifs"
  for item in "$@"; do
    item="${item#"${item%%[![:space:]]*}"}"
    item="${item%"${item##*[![:space:]]}"}"
    [[ -n "$item" ]] || continue
    requested+=("$item")
  done
else
  while IFS= read -r item; do
    [[ -n "$item" ]] && requested+=("$item")
  done <<EOF
$(default_platforms)
EOF
fi

valid_platform() {
  case "$1" in
    darwin/amd64|darwin/arm64|windows/amd64|windows/arm64|linux/amd64|linux/arm64) return 0 ;;
    *) return 1 ;;
  esac
}

problems=""
add_problem() {
  problems="${problems}error: $*"$'\n'
}

developer_id_identities() {
  security find-identity -v -p codesigning 2>/dev/null | sed -n 's/.*"\(Developer ID Application:[^"]*\)".*/\1/p'
}

resolve_apple_identity() {
  if [[ -n "${APPLE_SIGNING_IDENTITY:-}" ]]; then
    if ! developer_id_identities | grep -F -x "$APPLE_SIGNING_IDENTITY" >/dev/null; then
      add_problem "APPLE_SIGNING_IDENTITY is not a Developer ID Application certificate in the keychain: ${APPLE_SIGNING_IDENTITY}"
      return
    fi
    return
  fi
  local ids="" count
  ids="$(developer_id_identities || true)"
  count="$(printf '%s\n' "$ids" | sed '/^$/d' | wc -l | tr -d ' ')"
  if [[ "$count" == "1" ]]; then
    APPLE_SIGNING_IDENTITY="$(printf '%s\n' "$ids" | sed '/^$/d')"
    export APPLE_SIGNING_IDENTITY
    return
  fi
  if [[ "$count" == "0" ]]; then
    add_problem "No Developer ID Application certificate in the keychain. Install one, or set APPLE_SIGNING_IDENTITY in build/dist.env."
    return
  fi
  add_problem "Several Developer ID certificates are installed. Set APPLE_SIGNING_IDENTITY in build/dist.env to the one to use."
}

preflight_darwin_dist() {
  [[ -f "$root/build/darwin/entitlements.plist" ]] || add_problem "missing build/darwin/entitlements.plist"
  command -v /usr/bin/codesign >/dev/null 2>&1 || add_problem "codesign is not available. Install the Xcode command line tools."
  xcrun --find notarytool >/dev/null 2>&1 || add_problem "notarytool is not available. Install the Xcode command line tools."
  resolve_apple_identity
  if [[ -z "${APPLE_NOTARY_KEYCHAIN_PROFILE:-}" ]]; then
    add_problem "Set APPLE_NOTARY_KEYCHAIN_PROFILE in build/dist.env. Create it once with: xcrun notarytool store-credentials \"PROFILE\" --apple-id \"you@example.com\" --team-id \"TEAMID\" --password \"app-specific-password\""
    return
  fi
  if ! xcrun notarytool history --keychain-profile "$APPLE_NOTARY_KEYCHAIN_PROFILE" >/dev/null </dev/null; then
    add_problem "Notary profile '${APPLE_NOTARY_KEYCHAIN_PROFILE}' failed. Check APPLE_NOTARY_KEYCHAIN_PROFILE, or create it with xcrun notarytool store-credentials."
  fi
}

preflight_windows_dist() {
  local ts_tool=""
  if [[ "$host_os" == "windows" ]]; then
    if command -v signtool.exe >/dev/null 2>&1 || command -v signtool >/dev/null 2>&1; then
      ts_tool="signtool"
    fi
  elif command -v osslsigncode >/dev/null 2>&1; then
    ts_tool="osslsigncode"
  fi
  if [[ -z "$ts_tool" ]]; then
    if [[ "$host_os" == "windows" ]]; then
      add_problem "signtool is not on PATH. Install the Windows SDK."
    else
      add_problem "osslsigncode is not installed. On macOS: brew install osslsigncode"
    fi
  fi
  command -v makensis >/dev/null 2>&1 || add_problem "makensis is not installed. On macOS: brew install makensis"
  if [[ -z "${WINDOWS_CERT_PATH:-}" || ! -f "${WINDOWS_CERT_PATH}" ]]; then
    add_problem "Set WINDOWS_CERT_PATH in build/dist.env to a code signing .pfx file. A self-signed certificate still shows SmartScreen."
  fi
}

if [[ ${#requested[@]} -eq 0 ]]; then
  die "no platforms selected"
fi

buildable=()
skipped=0
checked_darwin=false
checked_windows=false
for plat in "${requested[@]}"; do
  if ! valid_platform "$plat"; then
    die "unsupported platform ${plat}"
  fi
  os_name="${plat%%/*}"
  arch="${plat##*/}"
  if [[ "$os_name" == "darwin" && "$host_os" != "darwin" ]]; then
    echo "Skipping ${plat}: build the Mac app on macOS."
    skipped=$((skipped + 1))
    continue
  fi
  if [[ "$os_name" == "linux" && "$host_os" != "linux" ]]; then
    echo "Skipping ${plat}: build Linux packages on Linux. From a Mac, run the Release workflow in GitHub Actions."
    skipped=$((skipped + 1))
    continue
  fi
  if [[ "$os_name" == "linux" && "$arch" != "$host_arch" ]]; then
    echo "Skipping ${plat}: the Linux packages are built for the host architecture (${host_arch})."
    skipped=$((skipped + 1))
    continue
  fi
  if [[ "$dist" == true && "$os_name" == "darwin" && "$checked_darwin" != true ]]; then
    preflight_darwin_dist
    checked_darwin=true
  fi
  if [[ "$dist" == true && "$os_name" == "windows" && "$checked_windows" != true ]]; then
    preflight_windows_dist
    checked_windows=true
  fi
  buildable+=("$plat")
done

if [[ -n "$problems" ]]; then
  printf '%s' "$problems" >&2
  echo "Copy build/dist.env.example to build/dist.env and fill in the values. A local build does not need them: $0 build" >&2
  exit 1
fi

if [[ ${#buildable[@]} -eq 0 ]]; then
  if [[ "$skipped" -gt 0 ]]; then
    die "no platforms left to build"
  fi
  die "no platforms selected"
fi

stamp="$(date +%y%m%d%H%M%S)"
out="dist/v${app_version}_${stamp}"
mkdir -p "$out"
cp LICENSE "$out/LICENSE"

sign_windows_file() {
  local input="$1"
  local ts="${WINDOWS_TIMESTAMP_URL:-http://timestamp.digicert.com}"
  local signed="${input}.signed"
  if [[ "$host_os" == "windows" ]]; then
    local tool="signtool"
    if command -v signtool.exe >/dev/null 2>&1; then
      tool="signtool.exe"
    fi
    if [[ -n "${WINDOWS_CERT_PASSWORD:-}" ]]; then
      "$tool" sign /fd SHA256 /td SHA256 /tr "$ts" /f "$WINDOWS_CERT_PATH" /p "$WINDOWS_CERT_PASSWORD" "$input"
    else
      "$tool" sign /fd SHA256 /td SHA256 /tr "$ts" /f "$WINDOWS_CERT_PATH" "$input"
    fi
    return
  fi
  rm -f "$signed"
  if [[ -n "${WINDOWS_CERT_PASSWORD:-}" ]]; then
    osslsigncode sign -pkcs12 "$WINDOWS_CERT_PATH" -pass "$WINDOWS_CERT_PASSWORD" -n "$product_name" -h sha256 -ts "$ts" -in "$input" -out "$signed"
  else
    osslsigncode sign -pkcs12 "$WINDOWS_CERT_PATH" -n "$product_name" -h sha256 -ts "$ts" -in "$input" -out "$signed"
  fi
  mv -f "$signed" "$input"
}

notarize_and_staple() {
  local target="$1"
  local submit="$target"
  local zip=""
  if [[ -d "$target" ]]; then
    zip="$(mktemp -t notary)"
    rm -f "$zip"
    zip="${zip}.zip"
    ditto -c -k --keepParent "$target" "$zip"
    submit="$zip"
  fi
  xcrun notarytool submit "$submit" --keychain-profile "$APPLE_NOTARY_KEYCHAIN_PROFILE" --wait
  xcrun stapler staple "$target"
  [[ -n "$zip" ]] && rm -f "$zip"
  xcrun stapler validate "$target"
}

build_darwin() {
  local arch="$1"
  echo "building darwin/${arch}"
  "$wails3" package GOOS=darwin ARCH="$arch"
  local app_path="bin/${app_name}.app"
  [[ -d "$app_path" ]] || die "missing Mac app ${app_path}"
  mkdir -p "$out/macos"
  local app_dest="$out/macos/${product_name}.app"
  rm -rf "$app_dest"

  if [[ "$dist" != true ]]; then
    mv "$app_path" "$app_dest"
    /usr/bin/xattr -cr "$app_dest"
    /usr/bin/codesign --force --deep --sign - "$app_dest"
    /usr/bin/codesign --verify --deep --strict "$app_dest"
    return
  fi

  /usr/bin/xattr -cr "$app_path"
  /usr/bin/codesign --force --timestamp --options runtime \
    --entitlements "$root/build/darwin/entitlements.plist" \
    --sign "$APPLE_SIGNING_IDENTITY" \
    "$app_path"
  /usr/bin/codesign --verify --deep --strict "$app_path"
  echo "notarizing ${app_path}"
  notarize_and_staple "$app_path"

  rm -f "bin/${app_name}.dmg"
  "$wails3" tool package --format dmg --name "$app_name" --out "bin" \
    --background "build/darwin/dmg-background.png" \
    --volume-icon "build/darwin/icons.icns" \
    --file-icon "build/darwin/dmg-file-icon.icns" \
    --window-width 540 \
    --window-height 380
  local dmg_path="bin/${app_name}.dmg"
  [[ -f "$dmg_path" ]] || die "missing DMG ${dmg_path}"
  /usr/bin/codesign --force --timestamp --sign "$APPLE_SIGNING_IDENTITY" "$dmg_path"
  echo "notarizing ${dmg_path}"
  notarize_and_staple "$dmg_path"

  ditto "$app_path" "$app_dest"
  local dmg_dest="$out/macos/${product_name}.dmg"
  cp "$dmg_path" "$dmg_dest"
  /usr/bin/codesign --verify --deep --strict "$app_dest"
  xcrun stapler validate "$app_dest"
  xcrun stapler validate "$dmg_dest"
  spctl --assess --verbose=2 "$app_dest"
  spctl --assess --type open --context context:primary-signature --verbose "$dmg_dest"
}

build_windows() {
  local arch="$1"
  echo "building windows/${arch}"
  "$wails3" build GOOS=windows ARCH="$arch"
  local exe="bin/${app_name}.exe"
  [[ -f "$exe" ]] || die "missing ${exe}"
  mkdir -p "$out/windows"
  local portable="$out/windows/${app_name}-v${app_version}-${arch}.exe"

  if [[ "$dist" != true ]]; then
    cp "$exe" "$portable"
    return
  fi

  echo "signing ${exe}"
  sign_windows_file "$exe"
  "$wails3" generate webview2bootstrapper -dir "$root/build/windows/nsis"
  local arg_name="ARG_WAILS_AMD64_BINARY"
  if [[ "$arch" == "arm64" ]]; then
    arg_name="ARG_WAILS_ARM64_BINARY"
  fi
  (
    cd "$root/build/windows/nsis"
    makensis -D"${arg_name}=${root}/bin/${app_name}.exe" project.nsi
  )
  local installer="bin/${app_name}-${arch}-installer.exe"
  [[ -f "$installer" ]] || die "missing installer ${installer}"
  echo "signing ${installer}"
  sign_windows_file "$installer"
  cp "$exe" "$portable"
  cp "$installer" "$out/windows/${app_name}-v${app_version}-${arch}-setup.exe"
}

write_nfpm_config() {
  local dest="$1"
  local arch="$2"
  APP_NAME="$app_name" APP_VERSION="$app_version" GOARCH="$arch" NFPM_SRC="build/linux/nfpm/nfpm.yaml" NFPM_DEST="$dest" python3 - <<'PY'
import os, re
from pathlib import Path
text = Path(os.environ["NFPM_SRC"]).read_text(encoding="utf-8")
app = os.environ["APP_NAME"]
version = os.environ["APP_VERSION"]
text = re.sub(r'(?m)^name:.*$', f'name: "{app}"', text, count=1)
text = re.sub(r'(?m)^version:.*$', f'version: "{version}"', text, count=1)
text = re.sub(r'src: "\./bin/[^"]+"', f'src: "./bin/{app}"', text, count=1)
text = re.sub(r'dst: "/usr/local/bin/[^"]+"', f'dst: "/usr/local/bin/{app}"', text, count=1)
text = re.sub(
    r'dst: "/usr/share/icons/hicolor/128x128/apps/[^"]+"',
    f'dst: "/usr/share/icons/hicolor/128x128/apps/{app}.png"',
    text,
    count=1,
)
text = re.sub(r'src: "\./build/linux/[^"]+\.desktop"', f'src: "./build/linux/{app}.desktop"', text, count=1)
text = re.sub(r'dst: "/usr/share/applications/[^"]+"', f'dst: "/usr/share/applications/{app}.desktop"', text, count=1)
Path(os.environ["NFPM_DEST"]).write_text(text, encoding="utf-8")
PY
}

sign_linux_package() {
  local file="$1"
  [[ -n "${LINUX_PGP_KEY:-}" ]] || return 0
  [[ -f "$LINUX_PGP_KEY" ]] || die "LINUX_PGP_KEY is not a file: ${LINUX_PGP_KEY}"
  local password="${LINUX_PGP_PASSWORD:-${WAILS_PGP_PASSWORD:-}}"
  if [[ -n "$password" ]]; then
    "$wails3" tool sign --input "$file" --pgp-key "$LINUX_PGP_KEY" --pgp-password "$password"
  else
    "$wails3" tool sign --input "$file" --pgp-key "$LINUX_PGP_KEY"
  fi
}

build_linux() {
  local arch="$1"
  echo "building linux/${arch}"
  "$wails3" build GOOS=linux ARCH="$arch"
  local bin_path="bin/${app_name}"
  [[ -f "$bin_path" ]] || die "missing ${bin_path}"
  mkdir -p "$out/linux"

  if [[ "$dist" != true ]]; then
    cp "$bin_path" "$out/linux/${app_name}-v${app_version}-${arch}"
    return
  fi

  export GIT_COMMITTER_NAME="${GIT_COMMITTER_NAME:-$(git -C "$root" log -1 --format=%an 2>/dev/null || true)}"
  GIT_COMMITTER_NAME="${GIT_COMMITTER_NAME:-Builder}"
  export GIT_COMMITTER_NAME
  export GIT_COMMITTER_EMAIL="${GIT_COMMITTER_EMAIL:-$(git -C "$root" log -1 --format=%ae 2>/dev/null || true)}"
  GIT_COMMITTER_EMAIL="${GIT_COMMITTER_EMAIL:-builder@localhost}"
  export GIT_COMMITTER_EMAIL
  export GOARCH="$arch"

  "$wails3" task linux:generate:dotdesktop
  local nfpm_tmp
  nfpm_tmp="$(mktemp "${TMPDIR:-/tmp}/nfpm.XXXXXX")"
  write_nfpm_config "$nfpm_tmp" "$arch"
  "$wails3" tool package -name "$app_name" -format deb -config "$nfpm_tmp" -out "$root/bin"
  "$wails3" tool package -name "$app_name" -format rpm -config "$nfpm_tmp" -out "$root/bin"
  rm -f "$nfpm_tmp"

  local deb="bin/${app_name}.deb"
  local rpm="bin/${app_name}.rpm"
  [[ -f "$deb" && -f "$rpm" ]] || die "missing deb or rpm package in bin/"
  sign_linux_package "$deb"
  sign_linux_package "$rpm"

  (
    cd "$root/build/linux/appimage"
    cp "$root/bin/${app_name}" "$app_name"
    cp "$root/build/appicon.png" "${app_name}.png"
    "$wails3" generate appimage \
      -binary "$app_name" \
      -icon "${app_name}.png" \
      -desktopfile "../${app_name}.desktop" \
      -outputdir "$root/bin" \
      -builddir "$root/build/linux/appimage/build"
  )
  local appimage_arch="x86_64"
  if [[ "$arch" == "arm64" ]]; then
    appimage_arch="aarch64"
  fi
  # wails lowercases the binary name and replaces spaces with hyphens.
  local appimage_base
  appimage_base="$(printf '%s' "$app_name" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')"
  local appimage="bin/${appimage_base}-${appimage_arch}.AppImage"
  [[ -f "$appimage" ]] || die "missing AppImage ${appimage}"

  cp "$deb" "$out/linux/${app_name}_${app_version}_${arch}.deb"
  cp "$rpm" "$out/linux/${app_name}-${app_version}-${arch}.rpm"
  cp "$appimage" "$out/linux/${app_name}-v${app_version}-${arch}.AppImage"
}

for plat in "${buildable[@]}"; do
  os_name="${plat%%/*}"
  arch="${plat##*/}"
  case "$os_name" in
    darwin) build_darwin "$arch" ;;
    windows) build_windows "$arch" ;;
    linux) build_linux "$arch" ;;
  esac
done

hash_line() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1"
  else
    sha256sum "$1"
  fi
}

(
  cd "$out"
  find . -type f ! -name checksums.txt -print | LC_ALL=C sort | while IFS= read -r file; do
    hash_line "$file"
  done > checksums.txt
)

echo "Output: $out"
