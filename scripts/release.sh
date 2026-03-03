#!/bin/bash
# scripts/release.sh
# Main release script - fully automated release process
# Usage: ./scripts/release.sh v0.1.0

set -e

VERSION=$1

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Elysium Release Process${NC}"
echo "================================"
echo ""

# Step 1: Validate version format
if [ -z "$VERSION" ]; then
    echo -e "${RED}❌ Error: No version specified${NC}"
    echo "Usage: ./scripts/release.sh v0.1.0"
    exit 1
fi

if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}❌ Error: Invalid version format${NC}"
    echo "Version must be in format: v0.1.0"
    exit 1
fi

echo -e "📋 Version: ${YELLOW}$VERSION${NC}"

# Step 2: Ensure we're on main branch
BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ]; then
    echo -e "${RED}❌ Error: Must be on main branch${NC}"
    echo "Current branch: $BRANCH"
    echo "Run: git checkout main"
    exit 1
fi

echo -e "✓ On main branch"

# Step 3: Pull latest changes
echo -e "\n📥 Pulling latest changes..."
git pull origin main || {
    echo -e "${RED}❌ Failed to pull from main${NC}"
    exit 1
}

# Step 4: Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${RED}❌ Error: Uncommitted changes found${NC}"
    echo "Commit your changes first:"
    git status --short
    exit 1
fi

echo -e "✓ Working directory clean"

# Step 5: Ensure scripts are executable
echo -e "\n🔧 Ensuring scripts are executable..."
chmod +x scripts/*.sh 2>/dev/null || true
echo -e "✓ Scripts are executable"

# Step 6: Run tests
echo -e "\n🧪 Running tests..."
./scripts/test.sh || {
    echo -e "${RED}❌ Tests failed${NC}"
    echo "Fix failing tests before releasing"
    exit 1
}

echo -e "${GREEN}✓ All tests passed${NC}"

# Step 7: Check test coverage (only for v1.0.0 and later)
MAJOR_VERSION=$(echo $VERSION | sed 's/v\([0-9]*\).*/\1/')
if [ "$MAJOR_VERSION" -ge 1 ]; then
    echo -e "\n📊 Checking test coverage..."
    COVERAGE=$(./scripts/coverage.sh 2>/dev/null || echo "0")
    
    if [ "$COVERAGE" -lt 80 ]; then
        echo -e "${RED}❌ Test coverage (${COVERAGE}%) is below 80%${NC}"
        echo "Increase test coverage before v1.0.0 release"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Test coverage: ${COVERAGE}%${NC}"
else
    echo -e "\n⚠️  Skipping coverage check for alpha/beta releases"
fi

# Step 8: Generate CHANGELOG
echo -e "\n📝 Generating CHANGELOG..."
./scripts/generate-changelog.sh "$VERSION" || {
    echo -e "${YELLOW}⚠️  Warning: Could not generate CHANGELOG${NC}"
    echo "Continuing without CHANGELOG update..."
}

# Step 9: Update version in code
echo -e "\n💾 Updating version in code..."

# Update Go version
sed -i "s/var Version = \".*\"/var Version = \"$VERSION\"/" cli/cmd/root.go

# Update Python version
sed -i "s/VERSION = \".*\"/VERSION = \"$VERSION\"/" server/app/config.py 2>/dev/null || true

echo -e "✓ Version updated to $VERSION"

# Step 10: Commit version bump
echo -e "\n💾 Committing version bump..."
git add -A
git commit -m "chore: release $VERSION" || {
    echo -e "${YELLOW}⚠️  No changes to commit${NC}"
}
git push origin main || {
    echo -e "${RED}❌ Failed to push to main${NC}"
    exit 1
}

echo -e "${GREEN}✓ Version bump committed${NC}"

# Step 11: Create git tag
echo -e "\n🏷️  Creating git tag..."
git tag -a "$VERSION" -m "Release $VERSION" || {
    echo -e "${RED}❌ Failed to create tag${NC}"
    exit 1
}
git push origin "$VERSION" || {
    echo -e "${RED}❌ Failed to push tag${NC}"
    exit 1
}

echo -e "${GREEN}✓ Tag $VERSION created${NC}"

# Step 12: Build binaries
echo -e "\n🔨 Building binaries..."
./scripts/build-all.sh "$VERSION" || {
    echo -e "${RED}❌ Binary build failed${NC}"
    exit 1
}

# Step 13: Create GitHub release
echo -e "\n📦 Creating GitHub release..."
./scripts/create-release.sh "$VERSION" || {
    echo -e "${RED}❌ Failed to create GitHub release${NC}"
    exit 1
}

# Step 14: Success message
echo -e "\n${GREEN}══════════════════════════════════${NC}"
echo -e "${GREEN}✅ Release $VERSION created successfully!${NC}"
echo -e "${GREEN}══════════════════════════════════${NC}"
echo ""
echo "Release URL: https://github.com/Lo10Th/Elysium/releases/tag/$VERSION"
echo ""
echo "Next steps:"
echo "  1. Check the release page for correctness"
echo "  2. Announce on Discord/Twitter (if public)"
echo "  3. Update Homebrew formula (if v1.0.0+)"
echo ""
echo "To install this release:"
echo "  curl -sSL https://github.com/Lo10Th/Elysium/releases/download/$VERSION/ely-linux-amd64 -o ely"
echo "  chmod +x ely"
echo "  sudo mv ely /usr/local/bin/"