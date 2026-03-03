#!/bin/bash
set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.1.0"
    exit 1
fi

if ! command -v gh &> /dev/null; then
    echo "Error: gh (GitHub CLI) is required"
    echo "Install from: https://cli.github.com/"
    exit 1
fi

if ! gh auth status &> /dev/null; then
    echo "Error: Not authenticated with GitHub"
    echo "Run: gh auth login"
    exit 1
fi

cd "$(dirname "$0")/.."

TAG_EXISTS=$(git tag -l "$VERSION")
if [ -n "$TAG_EXISTS" ]; then
    echo "Error: Tag $VERSION already exists"
    exit 1
fi

echo "Creating release for $VERSION..."

echo "Running tests..."
./scripts/test.sh

echo "Checking coverage..."
COVERAGE_PASSED=$(./scripts/coverage.sh | grep "Coverage threshold met" || true)
if [ -z "$COVERAGE_PASSED" ]; then
    echo "Error: Coverage below 80% threshold"
    exit 1
fi

echo "Generating changelog..."
./scripts/generate-changelog.sh "$VERSION"

echo "Building binaries..."
./scripts/build-all.sh

echo "Creating git tag..."
git tag -a "$VERSION" -m "Release $VERSION"

echo "Pushing tag to remote..."
git push origin "$VERSION"

echo "Creating GitHub release..."
gh release create "$VERSION" \
    --title "Elysium $VERSION" \
    --notes-file release_notes.md \
    cli/bin/linux_amd64/ely \
    cli/bin/linux_arm64/ely \
    cli/bin/darwin_amd64/ely \
    cli/bin/darwin_arm64/ely \
    cli/bin/windows_amd64/ely.exe 2>/dev/null || {
    echo "Warning: Could not upload all assets. Creating release without assets..."
    gh release create "$VERSION" \
        --title "Elysium $VERSION" \
        --notes-file release_notes.md
}

echo "Release $VERSION created successfully!"
echo "View at: https://github.com/Lo10Th/Elysium/releases/tag/$VERSION"