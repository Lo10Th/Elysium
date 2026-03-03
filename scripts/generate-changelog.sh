#!/bin/bash
# scripts/generate-changelog.sh
# Auto-generates CHANGELOG.md from GitHub issues in milestone
# Usage: ./scripts/generate-changelog.sh v0.1.0

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: ./scripts/generate-changelog.sh v0.1.0"
    exit 1
fi

# Extract version without 'v' prefix
MILESTONE_NAME="${VERSION#v}"

echo "📝 Generating CHANGELOG for $VERSION..."

# Create CHANGELOG if it doesn't exist
if [ ! -f CHANGELOG.md ]; then
    cat > CHANGELOG.md << 'EOF'
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

EOF
fi

# Create temporary file for new version
TEMP_FILE=$(mktemp)

# Add new version header
cat > "$TEMP_FILE" << EOF
## [$VERSION] - $(date +%Y-%m-%d)

EOF

# Extract milestone name from version
# v0.1.0 -> "v0.1.0 - Internal Alpha" or just the version
if [[ $MILESTONE_NAME == "0.1.0" ]]; then
    MILESTONE="v0.1.0 - Internal Alpha"
elif [[ $MILESTONE_NAME == "0.2.0" ]]; then
    MILESTONE="v0.2.0 - Private Beta"
elif [[ $MILESTONE_NAME == "0.3.0" ]]; then
    MILESTONE="v0.3.0 - Public Beta"
elif [[ $MILESTONE_NAME == "1.0.0" ]]; then
    MILESTONE="v1.0.0 - First Stable Release"
else
    MILESTONE="$VERSION"
fi

# Try to get issues from milestone
echo "### Added" >> "$TEMP_FILE"

# Fetch closed issues from milestone (if GitHub CLI available)
if command -v gh &> /dev/null; then
    # Get P0/P1 issues (features)
    gh issue list \
        --state closed \
        --limit 100 \
        --json number,title,labels \
        --jq '.[] | select(.labels[]?.name | contains("P0") or contains("P1")) | "- \(.title) (#\(.number))"' \
        2>/dev/null >> "$TEMP_FILE" || echo "- (See GitHub for details)" >> "$TEMP_FILE"
    
    echo "" >> "$TEMP_FILE"
    echo "### Fixed" >> "$TEMP_FILE"
    
    # Get bug issues
    gh issue list \
        --state closed \
        --limit 100 \
        --json number,title,labels \
        --jq '.[] | select(.labels[]?.name == "bug") | "- \(.title) (#\(.number))"' \
        2>/dev/null >> "$TEMP_FILE" || true
    
    echo "" >> "$TEMP_FILE"
    echo "### Changed" >> "$TEMP_FILE"
    
    # Get P2/P3 issues (improvements)
    gh issue list \
        --state closed \
        --limit 100 \
        --json number,title,labels \
        --jq '.[] | select(.labels[]?.name | contains("P2") or contains("P3")) | "- \(.title) (#\(.number))"' \
        2>/dev/null >> "$TEMP_FILE" || true
else
    # GitHub CLI not available, add placeholder
    echo "- See GitHub issues for details" >> "$TEMP_FILE"
fi

echo "" >> "$TEMP_FILE"

# Combine with existing CHANGELOG (after header)
{
    # Read existing header (first 7 lines)
    head -n 7 CHANGELOG.md 2>/dev/null || true
    
    # Add new version
    cat "$TEMP_FILE"
    
    # Add rest of existing CHANGELOG
    tail -n +8 CHANGELOG.md 2>/dev/null || true
} > CHANGELOG.md.new

# Replace old CHANGELOG
mv CHANGELOG.md.new CHANGELOG.md

# Cleanup
rm -f "$TEMP_FILE"

echo "✅ CHANGELOG.md updated"
echo ""
echo "Preview:"
echo "----------------------------------------"
head -n 20 CHANGELOG.md