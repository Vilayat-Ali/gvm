#!/usr/bin/env bash

# ============================================================================
# GVM INSTALLER — CYBER CORE EDITION
# Robust • Deterministic • Minimal
# ============================================================================

set -euo pipefail

# ----------------------------------------------------------------------------
# Configuration
# ----------------------------------------------------------------------------

VERBOSE=false
[[ "${1:-}" == "--verbose" || "${1:-}" == "-v" ]] && VERBOSE=true

REPO="Vilayat-Ali/gvm"
INSTALL_DIR="/usr/local/gvm"
BIN_NAME="gvm"

# ----------------------------------------------------------------------------
# Colors & Icons
# ----------------------------------------------------------------------------

CLR_PRIME='\033[38;5;51m'
CLR_ACCENT='\033[38;5;208m'
CLR_SUCCESS='\033[38;5;82m'
CLR_ERROR='\033[38;5;197m'
CLR_MUTE='\033[38;5;244m'
BOLD='\033[1m'
RESET='\033[0m'

log_step() { echo -e "\n${CLR_PRIME}${BOLD}⚙ $1${RESET}"; }
log_ok()   { echo -e "  ${CLR_SUCCESS}✔ $1${RESET}"; }
log_fail() { echo -e "  ${CLR_ERROR}✘ $1${RESET}"; exit 1; }

# ----------------------------------------------------------------------------
# Installation Process
# ----------------------------------------------------------------------------

# 1. Dependency/Privilege Check
log_step "Elevating Privileges & Checking Dependencies"
sudo -v || log_fail "Sudo access required"
for cmd in curl tar; do command -v "$cmd" >/dev/null || log_fail "Missing: $cmd"; done

# 2. Architecture & Cleanup
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
log_step "Cleaning Workspace"
[ -d "$INSTALL_DIR" ] && sudo rm -rf "$INSTALL_DIR"
TMP_DIR=$(mktemp -d)

# 3. Fetch & Extract
log_step "Downloading & Extracting Assets"
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
TARBALL="gvm-linux-${ARCH}.tar.gz"
curl -Lfs "https://github.com/$REPO/releases/download/${LATEST_TAG:-v1.0.0}/$TARBALL" -o "$TMP_DIR/$TARBALL"
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

# 4. Locate & Deploy
log_step "Deploying Binary"
# Dynamically find the executable inside the extract
BINARY_PATH=$(find "$TMP_DIR" -maxdepth 2 -type f -executable | head -n 1)

if [ -z "$BINARY_PATH" ]; then
    log_fail "No executable binary found in archive."
fi

sudo mkdir -p "$INSTALL_DIR"
sudo mv "$BINARY_PATH" "$INSTALL_DIR/$BIN_NAME"
sudo chmod 755 "$INSTALL_DIR/$BIN_NAME"
sudo ln -sf "$INSTALL_DIR/$BIN_NAME" "/usr/local/bin/$BIN_NAME"
log_ok "Binary installed to $INSTALL_DIR/$BIN_NAME"

# 5. Shell Configuration
log_step "Configuring Path"
for PROFILE in "$HOME/.zshrc" "$HOME/.bashrc"; do
    if [ -f "$PROFILE" ] && ! grep -q 'GVM_PATH_SETUP' "$PROFILE"; then
        echo -e '\n# GVM_PATH_SETUP\nexport PATH="/usr/local/gvm:$PATH"' >> "$PROFILE"
        log_ok "Updated $PROFILE"
    fi
done

# Cleanup
rm -rf "$TMP_DIR"

echo -e "\n${CLR_SUCCESS}${BOLD}INSTALLATION SUCCESSFUL${RESET}"