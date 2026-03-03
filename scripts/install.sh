#!/bin/bash
# scripts/install.sh
# Installs the Elysium CLI (ely) for the current platform
# Usage: curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash
#        Or: ./scripts/install.sh [--version v0.1.0] [--install-dir /usr/local/bin]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Defaults
REPO="Lo10Th/Elysium"
BINARY_NAME="ely"
INSTALL_DIR="/usr/local/bin"
VERSION=""
CONFIG_DIR="${HOME}/.elysium"
GITHUB_API="https://api.github.com"
GITHUB_RELEASES="https://github.com/${REPO}/releases"
SKIP_SHELL_CONFIG=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --skip-shell-config)
            SKIP_SHELL_CONFIG=true
            shift
            ;;
        -h|--help)
            echo "Usage: install.sh [--version v0.1.0] [--install-dir /usr/local/bin] [--skip-shell-config]"
            echo ""
            echo "Options:"
            echo "  --version           Specific version to install (default: latest)"
            echo "  --install-dir       Installation directory (default: /usr/local/bin)"
            echo "  --skip-shell-config Skip automatic shell completion and PATH setup"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

echo -e "${GREEN}🚀 Elysium CLI Installer${NC}"
echo "=========================="
echo ""

# Detect OS
detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux*)  echo "linux" ;;
        darwin*) echo "darwin" ;;
        msys*|mingw*|cygwin*) echo "windows" ;;
        *)
            echo -e "${RED}❌ Unsupported operating system: $os${NC}" >&2
            exit 1
            ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)
            echo -e "${RED}❌ Unsupported architecture: $arch${NC}" >&2
            exit 1
            ;;
    esac
}

# Get the latest release version from GitHub
get_latest_version() {
    local version
    if command -v curl &> /dev/null; then
        version=$(curl -sSL "${GITHUB_API}/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        version=$(wget -qO- "${GITHUB_API}/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        echo -e "${RED}❌ curl or wget is required${NC}" >&2
        exit 1
    fi

    if [ -z "$version" ]; then
        echo -e "${YELLOW}⚠️  Could not fetch latest version, defaulting to v0.1.0${NC}" >&2
        version="v0.1.0"
    fi
    echo "$version"
}

# Download a file
download_file() {
    local url="$1"
    local dest="$2"
    if command -v curl &> /dev/null; then
        curl -sSL "$url" -o "$dest"
    elif command -v wget &> /dev/null; then
        wget -qO "$dest" "$url"
    else
        echo -e "${RED}❌ curl or wget is required${NC}" >&2
        exit 1
    fi
}

# Check if a command exists
require_cmd() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}❌ Required command not found: $1${NC}" >&2
        exit 1
    fi
}

# Setup shell completion
setup_completion() {
    if [ "$SKIP_SHELL_CONFIG" = true ]; then
        echo ""
        echo -e "${YELLOW}⚠️  Skipping shell completion setup (--skip-shell-config).${NC}"
        echo "   To set it up manually, run: ${BINARY_NAME} completion --help"
        return
    fi

    local shell_name
    shell_name=$(basename "${SHELL:-bash}")
    local completion_installed=false

    echo ""
    echo -e "${BLUE}🔧 Setting up shell completion...${NC}"

    case "$shell_name" in
        bash)
            local bash_completion_dir
            if [ -d "/etc/bash_completion.d" ] && [ -w "/etc/bash_completion.d" ]; then
                bash_completion_dir="/etc/bash_completion.d"
            else
                mkdir -p "${HOME}/.bash_completion.d"
                bash_completion_dir="${HOME}/.bash_completion.d"
            fi

            if "${INSTALL_DIR}/${BINARY_NAME}" completion bash > "${bash_completion_dir}/ely" 2>/dev/null; then
                echo -e "${GREEN}✓ Bash completion installed to ${bash_completion_dir}/ely${NC}"
                if [ "$bash_completion_dir" = "${HOME}/.bash_completion.d" ]; then
                    # Ensure the completion dir is sourced
                    local bashrc="${HOME}/.bashrc"
                    if ! grep -q "bash_completion.d/ely" "$bashrc" 2>/dev/null; then
                        echo -e "${BLUE}  Adding completion source to ${bashrc}${NC}"
                        echo "" >> "$bashrc"
                        echo "# Elysium CLI completion" >> "$bashrc"
                        echo "[ -f \"\${HOME}/.bash_completion.d/ely\" ] && source \"\${HOME}/.bash_completion.d/ely\"" >> "$bashrc"
                    fi
                fi
                completion_installed=true
            fi
            ;;
        zsh)
            local zsh_completion_dir="${HOME}/.zsh/completions"
            mkdir -p "$zsh_completion_dir"
            if "${INSTALL_DIR}/${BINARY_NAME}" completion zsh > "${zsh_completion_dir}/_ely" 2>/dev/null; then
                echo -e "${GREEN}✓ Zsh completion installed to ${zsh_completion_dir}/_ely${NC}"
                local zshrc="${HOME}/.zshrc"
                if ! grep -q "${HOME}/.zsh/completions" "$zshrc" 2>/dev/null; then
                    echo -e "${BLUE}  Adding fpath entry to ${zshrc}${NC}"
                    echo "" >> "$zshrc"
                    echo "# Elysium CLI completion" >> "$zshrc"
                    echo "fpath=(\"\${HOME}/.zsh/completions\" \$fpath)" >> "$zshrc"
                    # Only add compinit if it's not already present in .zshrc
                    if ! grep -q "compinit" "$zshrc" 2>/dev/null; then
                        echo "autoload -U compinit && compinit" >> "$zshrc"
                    fi
                fi
                completion_installed=true
            fi
            ;;
        fish)
            local fish_completion_dir="${HOME}/.config/fish/completions"
            mkdir -p "$fish_completion_dir"
            if "${INSTALL_DIR}/${BINARY_NAME}" completion fish > "${fish_completion_dir}/ely.fish" 2>/dev/null; then
                echo -e "${GREEN}✓ Fish completion installed to ${fish_completion_dir}/ely.fish${NC}"
                completion_installed=true
            fi
            ;;
    esac

    if [ "$completion_installed" = false ]; then
        echo -e "${YELLOW}⚠️  Shell completion could not be installed automatically.${NC}"
        echo "   To set it up manually, run: ${BINARY_NAME} completion --help"
    fi
}

# Create default config directory and file
create_config() {
    echo ""
    echo -e "${BLUE}📁 Creating config directory...${NC}"
    mkdir -p "${CONFIG_DIR}"
    mkdir -p "${CONFIG_DIR}/cache"

    local config_file="${CONFIG_DIR}/config.yaml"
    if [ ! -f "$config_file" ]; then
        cat > "$config_file" << 'EOF'
# Elysium CLI Configuration
# Documentation: https://github.com/Lo10Th/Elysium/blob/main/docs/GETTING_STARTED.md

# Registry server URL (update with your registry endpoint if self-hosting)
registry_url: https://registry.elysium.dev

# Default output format: text, json
output_format: text

# Cache settings
cache:
  enabled: true
  ttl: 3600  # seconds
EOF
        echo -e "${GREEN}✓ Config created at ${config_file}${NC}"
    else
        echo -e "${GREEN}✓ Config already exists at ${config_file}${NC}"
    fi
}

# Main installation
main() {
    # Detect platform
    local os arch
    os=$(detect_os)
    arch=$(detect_arch)

    echo -e "Platform: ${YELLOW}${os}/${arch}${NC}"

    # Resolve version
    if [ -z "$VERSION" ]; then
        echo -e "${BLUE}Fetching latest version...${NC}"
        VERSION=$(get_latest_version)
    fi
    echo -e "Version:  ${YELLOW}${VERSION}${NC}"
    echo ""

    # Build download URL
    # Release assets are named: ely-{os}-{arch}.tar.gz (e.g. ely-linux-amd64.tar.gz)
    local asset_name="ely-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        asset_name="${asset_name}.exe.tar.gz"
    else
        asset_name="${asset_name}.tar.gz"
    fi
    local download_url="${GITHUB_RELEASES}/download/${VERSION}/${asset_name}"

    # Create a temporary directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    # Download the binary
    echo -e "${BLUE}Downloading ${asset_name}...${NC}"
    local archive="${tmp_dir}/${asset_name}"
    if ! download_file "$download_url" "$archive"; then
        echo -e "${RED}❌ Failed to download from: ${download_url}${NC}"
        echo "   Please check your network connection or visit:"
        echo "   ${GITHUB_RELEASES}"
        exit 1
    fi

    # Verify the download is non-empty
    if [ ! -s "$archive" ]; then
        echo -e "${RED}❌ Downloaded file is empty. The release asset may not exist yet.${NC}"
        echo "   Visit ${GITHUB_RELEASES} to download manually."
        exit 1
    fi

    # Extract binary
    echo -e "${BLUE}Extracting binary...${NC}"
    tar -xzf "$archive" -C "$tmp_dir"

    local binary_path="${tmp_dir}/${BINARY_NAME}"
    if [ "$os" = "windows" ]; then
        binary_path="${tmp_dir}/${BINARY_NAME}.exe"
    fi

    if [ ! -f "$binary_path" ]; then
        # Try finding the binary in subdirectories
        binary_path=$(find "$tmp_dir" -name "${BINARY_NAME}" -o -name "${BINARY_NAME}.exe" 2>/dev/null | head -1)
        if [ -z "$binary_path" ]; then
            echo -e "${RED}❌ Binary not found in archive${NC}"
            exit 1
        fi
    fi

    chmod +x "$binary_path"

    # Install binary
    echo -e "${BLUE}Installing to ${INSTALL_DIR}/${BINARY_NAME}...${NC}"
    if [ -w "$INSTALL_DIR" ]; then
        mv "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        # Try sudo if the directory is not writable
        if command -v sudo &> /dev/null; then
            sudo mv "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        else
            echo -e "${RED}❌ Cannot write to ${INSTALL_DIR}. Try running with sudo or specify a writable --install-dir.${NC}"
            exit 1
        fi
    fi

    echo -e "${GREEN}✓ Installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}${NC}"

    # Verify installation
    if command -v "${BINARY_NAME}" &> /dev/null; then
        local installed_version
        installed_version=$("${BINARY_NAME}" --version 2>/dev/null || echo "unknown")
        echo -e "${GREEN}✓ Verified: ${installed_version}${NC}"
    fi

    # Create config directory and default config
    create_config

    # Setup shell completion
    setup_completion

    # Final message
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}✅ Ely installed successfully!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
    echo "Run 'ely --help' to get started."
    echo ""
    echo "Next steps:"
    echo "  1. Authenticate:       ely login"
    echo "  2. Search emblems:     ely search <query>"
    echo "  3. Pull an emblem:     ely pull <name>"
    echo "  4. Use an emblem:      ely <name> <action>"
    echo ""
    echo "Documentation: https://github.com/Lo10Th/Elysium/blob/main/docs/GETTING_STARTED.md"

    # Warn if install dir is not in PATH
    if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
        echo ""
        echo -e "${YELLOW}⚠️  ${INSTALL_DIR} is not in your PATH.${NC}"
        echo "   Add it to your shell profile:"
        echo "   export PATH=\"\$PATH:${INSTALL_DIR}\""
    fi
}

main "$@"
