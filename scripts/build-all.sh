#!/bin/bash
# scripts/build-all.sh
# Cross-compiles binaries for all platforms
# Usage: ./scripts/build-all.sh v0.1.0

set -e

VERSION=$1

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

if [ -z "$VERSION" ]; then
    # Try to get version from git tag
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev")
fi

echo -e "${GREEN}🔨 Building Elysium Binaries${NC}"
echo "=========================="
echo -e "Version: ${YELLOW}$VERSION${NC}"
echo ""

# Extract version without 'v' for ldflags
VERSION_NUM="${VERSION#v}"

# Create build directory
BUILD_DIR="build/$VERSION"
mkdir -p "$BUILD_DIR"

echo "Build directory: $BUILD_DIR"
echo ""

# Change to CLI directory
cd cli

# Check if go.mod exists
if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ go.mod not found in cli/ directory${NC}"
    echo "Make sure you're in the project root"
    exit 1
fi

# Download dependencies
echo -e "${BLUE}Downloading Go modules...${NC}"
go mod download || {
    echo -e "${RED}❌ Failed to download modules${NC}"
    exit 1
}
echo -e "${GREEN}✓ Modules downloaded${NC}"
echo ""

# Define build targets
declare -A TARGETS=(
    ["linux-amd64"]="linux/amd64"
    ["linux-arm64"]="linux/arm64"
    ["darwin-amd64"]="darwin/amd64"
    ["darwin-arm64"]="darwin/arm64"
    ["windows-amd64"]="windows/amd64"
)

# Build for each target
for TARGET_NAME in "${!TARGETS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "${TARGETS[$TARGET_NAME]}"
    
    OUTPUT_NAME="ely-$TARGET_NAME"
    if [[ $TARGET_NAME == *"windows"* ]]; then
        OUTPUT_NAME="$OUTPUT_NAME.exe"
    fi
    
    echo -e "${BLUE}Building for $TARGET_NAME...${NC}"
    
    GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="-s -w -X 'github.com/elysium/elysium/cli/cmd.Version=$VERSION_NUM'" \
        -o "../$BUILD_DIR/$OUTPUT_NAME" \
        . || {
            echo -e "${RED}❌ Failed to build $TARGET_NAME${NC}"
            exit 1
        }
    
    # Get file size
    SIZE=$(ls -lh "../$BUILD_DIR/$OUTPUT_NAME" | awk '{print $5}')
    echo -e "${GREEN}✓ Built $OUTPUT_NAME ($SIZE)${NC}"
done

cd ..

echo ""
echo -e "${BLUE}📝 Generating checksums...${NC}"
cd "$BUILD_DIR"

# Generate SHA256 checksums
if command -v sha256sum &> /dev/null; then
    sha256sum ely-* > checksums.sha256
elif command -v shasum &> /dev/null; then
    shasum -a 256 ely-* > checksums.sha256
else
    echo -e "${YELLOW}⚠️  sha256sum not found, skipping checksums${NC}"
fi

cd - > /dev/null

echo ""
echo -e "${GREEN}═══════════════════════════════${NC}"
echo -e "${GREEN}✅ Build Complete!${NC}"
echo -e "${GREEN}═══════════════════════════════${NC}"
echo ""
echo "Binaries:"
ls -lh "$BUILD_DIR"
echo ""
echo "Checksums:"
cat "$BUILD_DIR/checksums.sha256" 2>/dev/null || echo "No checksums generated"
echo ""
echo "Next step: Run ./scripts/create-release.sh $VERSION"