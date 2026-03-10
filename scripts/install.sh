#!/bin/bash

# ============================================================================
# GVM INSTALLER: CYBER-CORE EDITION
# Minimalist, high-contrast, and robust.
# ============================================================================

set -e

# --- Configuration & Styling ---
VERBOSE=false
[[ "$1" == "--verbose" || "$1" == "-v" ]] && VERBOSE=true

# Professional Neon Palette
CLR_PRIME='\033[38;5;51m'   # Cyan
CLR_ACCENT='\033[38;5;208m'  # Orange
CLR_SUCCESS='\033[38;5;82m'  # Bright Green
CLR_ERROR='\033[38;5;197m'    # Rose Red
CLR_MUTE='\033[38;5;244m'    # Medium Gray
CLR_DIM='\033[2m'
BOLD='\033[1m'
RESET='\033[0m'

# Symbols
ICON_GEAR="⚙"
ICON_DOWN="󰇚"
ICON_SHIELD="󰒃"
ICON_LINK=""
ICON_CHECK="✔"

# --- UI Components ---
log_step() {
    echo -e "\n${CLR_PRIME}${BOLD}${ICON_GEAR} $1${RESET}"
}

log_sub() {
    echo -e "  ${CLR_MUTE}→ $1${RESET}"
}

log_ok() {
    echo -e "  ${CLR_SUCCESS}${ICON_CHECK} $1${RESET}"
}

log_fail() {
    echo -e "  ${CLR_ERROR}✘ $1${RESET}"
}

spinner() {
    local pid=$1
    local msg=$2
    local frames='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
    
    if [ "$VERBOSE" = true ]; then wait $pid; return; fi
    
    while kill -0 $pid 2>/dev/null; do
        for i in {0..9}; do
            printf "\r  ${CLR_ACCENT}${frames:$i:1}${RESET} ${msg}..."
            sleep 0.1
        done
    done
    printf "\r  ${CLR_SUCCESS}${ICON_CHECK}${RESET} ${msg} [COMPLETE]\n"
}

# --- Header ---
clear
echo -e "${CLR_PRIME}${BOLD}"
echo " ┌───────────────────────────────────────┐"
echo " │       GVM | Go Version Manager        │"
echo " └───────────────────────────────────────┘"
echo -e "${RESET}${CLR_MUTE}   Initializing secure installation...${RESET}\n"

# --- 1. System Check & Auth ---
log_step "Elevating Privileges"
if sudo -v; then
    log_ok "Auth confirmed"
else
    log_fail "Sudo required for system integration"
    exit 1
fi

# --- 2. Workspace Prep ---
log_step "Cleaning Environment"
if [ -d "/usr/local/gvm" ]; then
    log_sub "Removing stale files at /usr/local/gvm"
    sudo rm -rf /usr/local/gvm
fi
sudo rm -f /usr/local/bin/gvm
log_ok "Workspace ready"

# --- 3. Versioning ---
log_step "Fetching Release Data"
QUERY_URL="https://api.github.com/repos/Vilayat-Ali/gvm/releases/latest"
LATEST_TAG=$(curl -s "$QUERY_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
LATEST_TAG=${LATEST_TAG:-"v1.0.0"}

log_sub "Target Version: ${CLR_ACCENT}${LATEST_TAG}${RESET}"
BINARY_URL="https://github.com/Vilayat-Ali/gvm/releases/download/$LATEST_TAG/gvm-linux-x86.tar.gz"

# --- 4. Retrieval ---
log_step "Synchronizing Binary"
TMP_DIR=$(mktemp -d)

if [ "$VERBOSE" = true ]; then
    curl -L "$BINARY_URL" | tar -xzv -C "$TMP_DIR"
else
    (curl -Lfs "$BINARY_URL" | tar -xz -C "$TMP_DIR") &
    spinner $! "Downloading assets"
fi

# --- 5. Deployment ---
log_step "Deploying Core"
GVM_SOURCE=$(find "$TMP_DIR" -type f -name "gvm" | head -n 1)

if [ -z "$GVM_SOURCE" ]; then
    log_fail "Binary extraction failed"
    rm -rf "$TMP_DIR"
    exit 1
fi

sudo mkdir -p /usr/local/gvm
sudo mv "$GVM_SOURCE" /usr/local/gvm/gvm
log_ok "Binary moved to /usr/local/gvm"

# --- 6. Security Hardening ---
# Applying the specific permissions requested
log_step "Hardening Permissions"
log_sub "Applying SUID & Root Ownership"

sudo chown root:root /usr/local/gvm/gvm
sudo chmod u+s /usr/local/gvm/gvm

log_ok "Permissions secured (root:root, u+s)"

# --- 7. Linking ---
log_step "Global Integration"
sudo ln -sf /usr/local/gvm/gvm /usr/local/bin/gvm
log_ok "Symlink created in /usr/local/bin"

# --- 8. Path Configuration ---
log_step "Configuring Shell"
for PROFILE in "$HOME/.zshrc" "$HOME/.bashrc"; do
    if [ -f "$PROFILE" ]; then
        if ! grep -q "/usr/local/gvm" "$PROFILE"; then
            echo -e "\n# GVM Configuration\nexport PATH=\"\$PATH:/usr/local/gvm\"" >> "$PROFILE"
            log_ok "Added to $(basename $PROFILE)"
        else
            log_sub "$(basename $PROFILE) already configured"
        fi
    fi
done

# Cleanup
rm -rf "$TMP_DIR"

# --- Success Footer ---
echo -e "\n${CLR_SUCCESS}${BOLD}  DEPLOYMENT SUCCESSFUL${RESET}"
echo -e "  ${CLR_MUTE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "  ${BOLD}1.${RESET} Refresh:   ${CLR_ACCENT}source ~/.zshrc${RESET} (or .bashrc)"
echo -e "  ${BOLD}2.${RESET} Check:     ${CLR_ACCENT}gvm --version${RESET}"
echo -e "  ${BOLD}3.${RESET} Help:      ${CLR_ACCENT}gvm --help${RESET}"
echo -e "  ${CLR_MUTE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "  ${CLR_DIM}Crafted by Vilayat Ali | github.com/Vilayat-Ali${RESET}\n"