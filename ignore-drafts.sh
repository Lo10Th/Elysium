#!/bin/bash

# If no PR ID is present, it's not a PR. Proceed with the build.
if [[ -z "$VERCEL_GIT_PULL_REQUEST_ID" ]]; then
  echo "✅ - Not a Pull Request. Proceeding with build."
  exit 1
fi

# Fetch PR status from the GitHub API (Public repo, no token needed)
HTTP_STATUS=$(curl -s -o pr_response.json -w "%{http_code}" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/Lo10Th/Elysium/pulls/$VERCEL_GIT_PULL_REQUEST_ID")

if [[ "$HTTP_STATUS" -ne 200 ]]; then
  echo "⚠️ - Could not fetch PR data. Proceeding with build as a fallback."
  exit 1
fi

# Parse 'draft' status
if grep -q '"draft": true' pr_response.json; then
  echo "🛑 - Draft PR detected. Cancelling build."
  exit 0
else
  echo "✅ - PR is ready for review. Proceeding with build."
  exit 1
fi
