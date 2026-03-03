#!/bin/bash
# scripts/coverage.sh
# Calculates and returns test coverage percentage
# Usage: ./scripts/coverage.sh

set -e

echo "📊 Calculating test coverage..."

# Initialize coverage values
GO_COVERAGE="0"
PYTHON_COVERAGE="0"

# Calculate Go coverage
if [ -d "cli" ] && [ -f "cli/go.mod" ]; then
    cd cli
    GO_COVERAGE=$(go test ./... -coverprofile=coverage.out 2>&1 | grep -oP 'coverage: \K[0-9.]+' || echo "0")
    echo "Go coverage: ${GO_COVERAGE}%"
    cd ..
else
    echo "⚠️  Go CLI not found, skipping"
fi

# Calculate Python coverage
if [ -d "server" ] && [ -f "server/requirements.txt" ]; then
    cd server
    if command -v pytest &> /dev/null; then
        PYTHON_COVERAGE=$(pytest tests/ --cov=app --cov-report=term 2>&1 | grep TOTAL | awk '{print $4}' | sed 's/%//' || echo "0")
        echo "Python coverage: ${PYTHON_COVERAGE}%"
    else
        echo "⚠️  pytest not installed, skipping"
    fi
    cd ..
else
    echo "⚠️  Python server not found, skipping"
fi

# Calculate average (if both exist)
if [ "$GO_COVERAGE" != "0" ] && [ "$PYTHON_COVERAGE" != "0" ]; then
    # Use awk for floating point arithmetic
    AVERAGE=$(awk "BEGIN {printf \"%.1f\", ($GO_COVERAGE + $PYTHON_COVERAGE) / 2}")
    echo ""
    echo "📊 Average coverage: ${AVERAGE}%"
    echo "$AVERAGE"
else
    # Return what we have
    if [ "$GO_COVERAGE" != "0" ]; then
        echo "$GO_COVERAGE"
    elif [ "$PYTHON_COVERAGE" != "0" ]; then
        echo "$PYTHON_COVERAGE"
    else
        echo "0"
    fi
fi