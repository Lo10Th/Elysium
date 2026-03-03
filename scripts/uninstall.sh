#!/bin/bash
# scripts/uninstall.sh
# Removes the Elysium CLI (ely) and optionally its config and cache
# Usage: ./scripts/uninstall.sh [--purge]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Defaults
BINARY_NAME="ely"
CONFIG_DIR="${HOME}/.elysium"
PURGE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --purge)
            PURGE=true
            shift
            ;;
        -h|--help)
            echo "Usage: uninstall.sh [--purge]"
            echo ""
            echo "Options:"
            echo "  --purge   Also remove config files and cached emblems (~/.elysium)"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}🗑️  Elysium CLI Uninstaller${NC}"
echo "==========================="
echo ""

# Find and remove binary
remove_binary() {
    local found=false
    local binary_path

    # Common installation directories to check
    for dir in /usr/local/bin /usr/bin "${HOME}/.local/bin" "${HOME}/bin"; do
        binary_path="${dir}/${BINARY_NAME}"
        if [ -f "$binary_path" ]; then
            echo -e "${BLUE}Found ${BINARY_NAME} at ${binary_path}${NC}"
            if [ -w "$dir" ]; then
                rm -f "$binary_path"
            elif command -v sudo &> /dev/null; then
                echo -e "${YELLOW}  ${dir} is not writable; requesting sudo to remove binary...${NC}"
                sudo rm -f "$binary_path"
            else
                echo -e "${RED}❌ Cannot remove ${binary_path}: permission denied${NC}"
                exit 1
            fi
            echo -e "${GREEN}✓ Removed ${binary_path}${NC}"
            found=true
        fi
    done

    if [ "$found" = false ]; then
        # Try to locate with 'which'
        binary_path=$(which "${BINARY_NAME}" 2>/dev/null || true)
        if [ -n "$binary_path" ] && [ -f "$binary_path" ]; then
            echo -e "${BLUE}Found ${BINARY_NAME} at ${binary_path}${NC}"
            if [ -w "$(dirname "$binary_path")" ]; then
                rm -f "$binary_path"
            elif command -v sudo &> /dev/null; then
                echo -e "${YELLOW}  $(dirname "$binary_path") is not writable; requesting sudo to remove binary...${NC}"
                sudo rm -f "$binary_path"
            else
                echo -e "${RED}❌ Cannot remove ${binary_path}: permission denied${NC}"
                exit 1
            fi
            echo -e "${GREEN}✓ Removed ${binary_path}${NC}"
            found=true
        fi
    fi

    if [ "$found" = false ]; then
        echo -e "${YELLOW}⚠️  ${BINARY_NAME} binary not found in standard locations${NC}"
    fi
}

# Remove shell completion files
remove_completions() {
    echo ""
    echo -e "${BLUE}Removing shell completions...${NC}"
    local removed=false

    # Bash
    for f in /etc/bash_completion.d/ely "${HOME}/.bash_completion.d/ely"; do
        if [ -f "$f" ]; then
            if [ -w "$(dirname "$f")" ]; then
                rm -f "$f"
            else
                echo -e "${YELLOW}  $(dirname "$f") is not writable; requesting sudo...${NC}"
                sudo rm -f "$f" 2>/dev/null || true
            fi
            echo -e "${GREEN}✓ Removed bash completion: ${f}${NC}"
            removed=true
        fi
    done

    # Zsh
    if [ -f "${HOME}/.zsh/completions/_ely" ]; then
        rm -f "${HOME}/.zsh/completions/_ely"
        echo -e "${GREEN}✓ Removed zsh completion: ${HOME}/.zsh/completions/_ely${NC}"
        removed=true
    fi

    # Fish
    if [ -f "${HOME}/.config/fish/completions/ely.fish" ]; then
        rm -f "${HOME}/.config/fish/completions/ely.fish"
        echo -e "${GREEN}✓ Removed fish completion: ${HOME}/.config/fish/completions/ely.fish${NC}"
        removed=true
    fi

    if [ "$removed" = false ]; then
        echo -e "${YELLOW}  No completion files found${NC}"
    fi
}

# Remove config and cache
remove_config() {
    if [ -d "$CONFIG_DIR" ]; then
        rm -rf "$CONFIG_DIR"
        echo -e "${GREEN}✓ Removed config directory: ${CONFIG_DIR}${NC}"
    else
        echo -e "${YELLOW}  Config directory not found: ${CONFIG_DIR}${NC}"
    fi
}

main() {
    remove_binary
    remove_completions

    if [ "$PURGE" = true ]; then
        echo ""
        echo -e "${BLUE}Removing config and cache (--purge)...${NC}"
        remove_config
    else
        if [ -d "$CONFIG_DIR" ]; then
            echo ""
            echo -e "${YELLOW}  Config directory retained: ${CONFIG_DIR}${NC}"
            echo "  Run with --purge to also remove config and cached emblems."
        fi
    fi

    echo ""
    echo -e "${GREEN}✅ Ely uninstalled successfully!${NC}"
}

main "$@"
