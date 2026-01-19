#!/bin/bash

# ============================================================================
# GVM Installation Script
# A beautiful, modern installer with smooth progress indicators
# ============================================================================

set -e

# Configuration
VERBOSE=false
[[ "$1" == "--verbose" || "$1" == "-v" ]] && VERBOSE=true

# Color palette - Modern minimalist theme
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'
BLUE='\033[38;5;39m'
GREEN='\033[38;5;42m'
YELLOW='\033[38;5;220m'
RED='\033[38;5;196m'
GRAY='\033[38;5;240m'
WHITE='\033[38;5;255m'

# Unicode symbols
CHECK="✓"
CROSS="✗"
ARROW="→"
DOTS="..."

# Logging functions
log_verbose() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${DIM}${GRAY}  ↳ $1${RESET}"
    fi
}

step() {
    echo -e "\n${BLUE}${BOLD}$1${RESET}"
}

success() {
    echo -e "${GREEN}${CHECK}${RESET} $1"
}

error() {
    echo -e "${RED}${CROSS}${RESET} $1"
}

info() {
    echo -e "${GRAY}${ARROW}${RESET} $1"
}

spinner() {
    local pid=$1
    local msg=$2
    local spin='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
    local i=0
    
    if [ "$VERBOSE" = true ]; then
        wait $pid
        return
    fi
    
    echo -n "  "
    while kill -0 $pid 2>/dev/null; do
        i=$(( (i+1) %10 ))
        printf "\r  ${BLUE}${spin:$i:1}${RESET} ${msg}${DOTS}"
        sleep 0.1
    done
    printf "\r  ${GREEN}${CHECK}${RESET} ${msg}\n"
}

# Header
clear
echo ""
echo -e "${BLUE}${BOLD}"
echo "  ╔═══════════════════════════════════════╗"
echo "  ║                                       ║"
echo "  ║         GVM INSTALLER SCRIPT          ║"
echo "  ║                                       ║"
echo "  ╚═══════════════════════════════════════╝"
echo -e "${RESET}"
echo -e "${DIM}${GRAY}  Go Version Manager Installation${RESET}"
echo ""

if [ "$VERBOSE" = true ]; then
    info "Verbose mode enabled"
fi

# Step 1: Authentication
step "Authentication"
log_verbose "Requesting administrative privileges"

if sudo -v; then
    success "Administrative access granted"
    log_verbose "Sudo session cached successfully"
else
    error "Failed to obtain administrative privileges"
    exit 1
fi

# Step 2: Cleanup
step "Environment Preparation"
log_verbose "Checking for existing GVM installation"

if [ -d "/usr/local/gvm" ]; then
    info "Found previous installation, removing..."
    log_verbose "Purging directory: /usr/local/gvm"
    sudo rm -rf /usr/local/gvm
    success "Previous installation removed"
fi

if [ -L "/usr/local/bin/gvm" ]; then
    log_verbose "Removing existing symlink"
    sudo rm -f /usr/local/bin/gvm
fi

# Step 3: Fetch latest version
step "Version Detection"
info "Querying GitHub for latest release"

QUERY_URL="https://api.github.com/repos/Vilayat-Ali/gvm/releases/latest"
log_verbose "API endpoint: $QUERY_URL"

LATEST_TAG=$(curl -s "$QUERY_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
LATEST_TAG=${LATEST_TAG:-"v1.0.0"}

success "Latest version: ${BOLD}${LATEST_TAG}${RESET}"
log_verbose "Download URL constructed"

BINARY_URL="https://github.com/Vilayat-Ali/gvm/releases/download/$LATEST_TAG/gvm-linux-x86.tar.gz"

# Step 4: Download
step "Package Download"
TMP_DIR=$(mktemp -d)
log_verbose "Temporary directory: $TMP_DIR"

if [ "$VERBOSE" = true ]; then
    info "Downloading and extracting GVM binary..."
    curl -L "$BINARY_URL" | tar -xzv -C "$TMP_DIR"
    success "Download complete"
else
    (curl -Lfs "$BINARY_URL" | tar -xz -C "$TMP_DIR") &
    spinner $! "Downloading GVM binary"
fi

# Step 5: Installation
step "Binary Installation"

GVM_SOURCE=$(find "$TMP_DIR" -type f -name "gvm" | head -n 1)

if [ -z "$GVM_SOURCE" ]; then
    error "Binary not found in archive"
    log_verbose "Archive contents: $(ls -R "$TMP_DIR")"
    rm -rf "$TMP_DIR"
    exit 1
fi

log_verbose "Binary located: $GVM_SOURCE"
info "Installing to /usr/local/gvm"

sudo mkdir -p /usr/local/gvm
sudo mv "$GVM_SOURCE" /usr/local/gvm/gvm

success "Binary installed"

# Step 6: Permissions
step "System Integration"

log_verbose "Setting ownership to root:root"
sudo chown root:root /usr/local/gvm/gvm

log_verbose "Applying executable permissions (4755)"
sudo chmod 4755 /usr/local/gvm/gvm

info "Creating symlink in /usr/local/bin"
sudo ln -sf /usr/local/gvm/gvm /usr/local/bin/gvm

success "System integration complete"

# Step 7: PATH configuration
step "Shell Configuration"

for PROFILE in "$HOME/.zshrc" "$HOME/.bashrc"; do
    if [ -f "$PROFILE" ]; then
        if ! grep -q "/usr/local/gvm" "$PROFILE"; then
            log_verbose "Updating $PROFILE"
            echo -e "\n# GVM - Go Version Manager\nexport PATH=\"\$PATH:/usr/local/gvm\"" >> "$PROFILE"
            success "Updated $(basename $PROFILE)"
        else
            info "$(basename $PROFILE) already configured"
        fi
    fi
done

# Cleanup
log_verbose "Removing temporary files"
rm -rf "$TMP_DIR"

# Success message
echo ""
echo -e "${GREEN}${BOLD}"
echo "  ╔═══════════════════════════════════════╗"
echo "  ║                                       ║"
echo "  ║     Installation Successful! 🎉       ║"
echo "  ║                                       ║"
echo "  ╚═══════════════════════════════════════╝"
echo -e "${RESET}"
echo ""
echo -e "${YELLOW}${BOLD}Next Steps:${RESET}"
echo -e "  ${GRAY}1.${RESET} Run: ${BOLD}source ~/.zshrc${RESET} ${DIM}(or ~/.bashrc)${RESET}"
echo -e "  ${GRAY}2.${RESET} Verify: ${BOLD}gvm --version${RESET}"
echo -e "  ${GRAY}3.${RESET} Get started: ${BOLD}gvm --help${RESET}"
echo ""
echo -e "${DIM}${GRAY}─────────────────────────────────────────${RESET}"
echo -e "${DIM}Built with ❤️  by ${RESET}${BOLD}Vilayat${RESET}"
echo -e "${DIM}GitHub: ${RESET}${BLUE}https://github.com/Vilayat-Ali${RESET}"
echo -e "${DIM}${GRAY}─────────────────────────────────────────${RESET}"
echo ""