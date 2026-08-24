#!/usr/bin/env bash
set -euo pipefail

REPO="${GVM_REPO:-Vilayat-Ali/gvm}"
BIN_NAME="gvm"
INSTALL_DIR="${GVM_INSTALL_DIR:-$HOME/.local/bin}"
GVM_ROOT="${GVM_ROOT:-${XDG_DATA_HOME:-$HOME/.local/share}/gvm}"

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
    C_OK=$'\033[32m'; C_WARN=$'\033[33m'; C_ERR=$'\033[31m'; C_DIM=$'\033[2m'; C_RESET=$'\033[0m'
else
    C_OK=''; C_WARN=''; C_ERR=''; C_DIM=''; C_RESET=''
fi

step() { printf '\n%s==>%s %s\n' "$C_DIM" "$C_RESET" "$1"; }
ok()   { printf '  %s%s%s\n' "$C_OK" "$1" "$C_RESET"; }
warn() { printf '  %s! %s%s\n' "$C_WARN" "$1" "$C_RESET" >&2; }
die()  { printf '  %serror: %s%s\n' "$C_ERR" "$1" "$C_RESET" >&2; exit 1; }

TMP_DIR=""
cleanup() { [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ] && rm -rf "$TMP_DIR"; }
trap cleanup EXIT INT TERM

if [ "$(id -u)" -eq 0 ] && [ -z "${GVM_ALLOW_ROOT:-}" ]; then
    die "do not install gvm as root; it installs into your home directory. Set GVM_ALLOW_ROOT=1 to override"
fi

step "Checking dependencies"
for cmd in curl tar uname mktemp install; do
    command -v "$cmd" >/dev/null 2>&1 || die "missing required command: $cmd"
done
command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1 \
    || warn "no sha256sum or shasum found; the download cannot be verified"
ok "all required tools present"

step "Detecting platform"
case "$(uname -s)" in
    Linux)  OS=linux ;;
    Darwin) OS=darwin ;;
    *) die "unsupported operating system: $(uname -s)" ;;
esac
case "$(uname -m)" in
    x86_64|amd64)  ARCH=amd64 ;;
    aarch64|arm64) ARCH=arm64 ;;
    *) die "unsupported architecture: $(uname -m)" ;;
esac
ok "$OS/$ARCH"

step "Resolving the latest release"
VERSION="${GVM_VERSION:-}"
if [ -z "$VERSION" ] && [ -n "${GVM_BASE_URL:-}" ]; then
    die "set GVM_VERSION when using GVM_BASE_URL"
fi
if [ -z "$VERSION" ]; then
    VERSION=$(curl -fsSL --retry 3 --retry-delay 1 --max-time 30 \
        "https://api.github.com/repos/$REPO/releases/latest" \
        | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1) \
        || die "cannot reach the GitHub release API"
fi
[ -n "$VERSION" ] || die "could not determine the latest release; set GVM_VERSION=vX.Y.Z to pin one"
ok "$VERSION"

TARBALL="${BIN_NAME}-${OS}-${ARCH}-${VERSION}.tar.gz"
BASE_URL="${GVM_BASE_URL:-https://github.com/$REPO/releases/download/$VERSION}"
TMP_DIR=$(mktemp -d)

step "Downloading $TARBALL"
curl -fsSL --retry 3 --retry-delay 1 --max-time 300 "$BASE_URL/$TARBALL" -o "$TMP_DIR/$TARBALL" \
    || die "download failed: $BASE_URL/$TARBALL"
ok "downloaded"

step "Verifying checksum"
if curl -fsSL --retry 2 --max-time 30 "$BASE_URL/SHA256SUMS.txt" -o "$TMP_DIR/SHA256SUMS.txt" 2>/dev/null; then
    EXPECTED=$(awk -v f="$TARBALL" '$2 == f || $2 == "*"f {print $1}' "$TMP_DIR/SHA256SUMS.txt" | head -n 1)
    if [ -z "$EXPECTED" ]; then
        warn "$TARBALL is not listed in SHA256SUMS.txt; skipping verification"
    else
        if command -v sha256sum >/dev/null 2>&1; then
            ACTUAL=$(sha256sum "$TMP_DIR/$TARBALL" | awk '{print $1}')
        else
            ACTUAL=$(shasum -a 256 "$TMP_DIR/$TARBALL" | awk '{print $1}')
        fi
        [ "$ACTUAL" = "$EXPECTED" ] || die "checksum mismatch: expected $EXPECTED, got $ACTUAL"
        ok "checksum verified"
    fi
else
    warn "no SHA256SUMS.txt published for $VERSION; skipping verification"
fi

step "Installing"
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR" || die "cannot unpack $TARBALL"
BINARY=$(find "$TMP_DIR" -type f -name "$BIN_NAME" -perm -u+x | head -n 1)
[ -n "$BINARY" ] || die "no $BIN_NAME binary found inside $TARBALL"

"$BINARY" --version >/dev/null 2>&1 || die "the downloaded binary does not run on this machine"

mkdir -p "$INSTALL_DIR"
install -m 0755 "$BINARY" "$INSTALL_DIR/$BIN_NAME"
ok "installed $INSTALL_DIR/$BIN_NAME ($("$INSTALL_DIR/$BIN_NAME" --version))"

step "Configuring your shell"
SHIM_DIR="$GVM_ROOT/bin"
LINE="export PATH=\"$SHIM_DIR:$INSTALL_DIR:\$PATH\""
MARKER="# added by gvm installer"
UPDATED=0

for PROFILE in "$HOME/.bashrc" "$HOME/.zshrc"; do
    [ -f "$PROFILE" ] || continue
    if grep -qF "$MARKER" "$PROFILE"; then
        ok "$PROFILE already configured"
        UPDATED=1
        continue
    fi
    printf '\n%s\n%s\n' "$MARKER" "$LINE" >> "$PROFILE"
    ok "updated $PROFILE"
    UPDATED=1
done

if [ "$UPDATED" -eq 0 ]; then
    warn "no shell profile found; add this line to yours manually:"
    printf '    %s\n' "$LINE"
fi

step "Done"
printf '  Open a new shell, then run:\n\n'
printf '    gvm configure\n'
printf '    gvm use latest\n'
printf '    gvm doctor\n\n'
