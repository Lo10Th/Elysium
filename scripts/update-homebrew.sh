#!/bin/bash
# scripts/update-homebrew.sh
# Updates the Homebrew formula with new version and checksums after a release.
# Usage: ./scripts/update-homebrew.sh v0.1.1

set -e

VERSION=$1

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

if [ -z "$VERSION" ]; then
    echo -e "${RED}❌ Error: No version specified${NC}"
    echo "Usage: ./scripts/update-homebrew.sh v0.1.0"
    exit 1
fi

if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}❌ Error: Invalid version format${NC}"
    echo "Version must be in format: v0.1.0"
    exit 1
fi

VERSION_NUM="${VERSION#v}"
FORMULA="Formula/ely.rb"
BASE_URL="https://github.com/Lo10Th/Elysium/releases/download/$VERSION"

echo -e "${GREEN}🍺 Updating Homebrew Formula${NC}"
echo "================================"
echo -e "Version: ${YELLOW}$VERSION${NC}"
echo ""

if [ ! -f "$FORMULA" ]; then
    echo -e "${RED}❌ Formula not found: $FORMULA${NC}"
    echo "Run this script from the repository root."
    exit 1
fi

# Download release tarballs and compute SHA256 checksums
PLATFORMS=("darwin-amd64" "darwin-arm64" "linux-amd64" "linux-arm64")
declare -A CHECKSUMS

for PLATFORM in "${PLATFORMS[@]}"; do
    TARBALL="ely-${PLATFORM}.tar.gz"
    URL="$BASE_URL/$TARBALL"
    TMP_FILE="/tmp/$TARBALL"

    echo -e "${BLUE}Downloading $TARBALL...${NC}"
    if ! curl -fsSL "$URL" -o "$TMP_FILE"; then
        echo -e "${RED}❌ Failed to download $URL${NC}"
        echo "Make sure the release $VERSION exists on GitHub."
        exit 1
    fi

    if command -v sha256sum &> /dev/null; then
        SHA=$(sha256sum "$TMP_FILE" | awk '{print $1}')
    elif command -v shasum &> /dev/null; then
        SHA=$(shasum -a 256 "$TMP_FILE" | awk '{print $1}')
    else
        echo -e "${RED}❌ sha256sum / shasum not found${NC}"
        exit 1
    fi

    CHECKSUMS[$PLATFORM]="$SHA"
    echo -e "${GREEN}✓ $PLATFORM: $SHA${NC}"
    rm -f "$TMP_FILE"
done

echo ""
echo -e "${BLUE}Updating $FORMULA...${NC}"

# Update version
sed -i.bak "s/version \".*\"/version \"$VERSION_NUM\"/" "$FORMULA" && rm -f "${FORMULA}.bak"

# Update SHA256 placeholders / existing values
sed -i.bak "/darwin-amd64/{n;s/sha256 \".*\"/sha256 \"${CHECKSUMS[darwin-amd64]}\"/}" "$FORMULA" && rm -f "${FORMULA}.bak"
sed -i.bak "/darwin-arm64/{n;s/sha256 \".*\"/sha256 \"${CHECKSUMS[darwin-arm64]}\"/}" "$FORMULA" && rm -f "${FORMULA}.bak"
sed -i.bak "/linux-amd64/{n;s/sha256 \".*\"/sha256 \"${CHECKSUMS[linux-amd64]}\"/}" "$FORMULA" && rm -f "${FORMULA}.bak"
sed -i.bak "/linux-arm64/{n;s/sha256 \".*\"/sha256 \"${CHECKSUMS[linux-arm64]}\"/}" "$FORMULA" && rm -f "${FORMULA}.bak"

echo -e "${GREEN}✓ Formula updated${NC}"
echo ""
echo -e "${BLUE}Updated $FORMULA:${NC}"
cat "$FORMULA"
echo ""
echo -e "${GREEN}✅ Done! Commit and push the updated formula:${NC}"
echo "  git add $FORMULA"
echo "  git commit -m \"chore: update Homebrew formula to $VERSION\""
echo "  git push origin main"
