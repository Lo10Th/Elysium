#!/bin/bash
# scripts/test.sh
# Runs all tests for Go and Python
# Usage: ./scripts/test.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🧪 Running all tests${NC}"
echo "=================="
echo ""

# Track overall success
OVERALL_SUCCESS=true

# Test Go CLI
echo -e "${YELLOW}Testing Go CLI...${NC}"
if [ -d "cli" ]; then
    cd cli
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        # Download dependencies
        echo "Downloading dependencies..."
        go mod download 2>/dev/null || true
        
        # Run tests
        if go test ./... -v -count=1 2>&1 | tee /tmp/go-test-output; then
            echo -e "${GREEN}✓ Go tests passed${NC}"
        else
            echo -e "${RED}✗ Go tests failed${NC}"
            OVERALL_SUCCESS=false
        fi
    else
        echo -e "${YELLOW}⚠️  go.mod not found, skipping Go tests${NC}"
    fi
    
    cd ..
else
    echo -e "${YELLOW}⚠️  cli/ directory not found, skipping Go tests${NC}"
fi

echo ""

# Test Python Server
echo -e "${YELLOW}Testing Python Server...${NC}"
if [ -d "server" ]; then
    cd server
    
    # Check if requirements.txt exists
    if [ -f "requirements.txt" ]; then
        # Check if pytest is available
        if command -v pytest &> /dev/null; then
            if pytest tests/ -v 2>&1 | tee /tmp/python-test-output; then
                echo -e "${GREEN}✓ Python tests passed${NC}"
            else
                echo -e "${RED}✗ Python tests failed${NC}"
                OVERALL_SUCCESS=false
            fi
        else
            echo -e "${YELLOW}⚠️  pytest not installed, skipping Python tests${NC}"
            echo "   Install with: pip install pytest"
        fi
    else
        echo -e "${YELLOW}⚠️  requirements.txt not found, skipping Python tests${NC}"
    fi
    
    cd ..
else
    echo -e "${YELLOW}⚠️  server/ directory not found, skipping Python tests${NC}"
fi

echo ""

# Test E2E (if test script exists)
if [ -f "tests/e2e-test.sh" ]; then
    echo -e "${YELLOW}Running E2E tests...${NC}"
    chmod +x tests/e2e-test.sh
    if ./tests/e2e-test.sh; then
        echo -e "${GREEN}✓ E2E tests passed${NC}"
    else
        echo -e "${RED}✗ E2E tests failed${NC}"
        OVERALL_SUCCESS=false
    fi
fi

echo ""
echo "================================"

if [ "$OVERALL_SUCCESS" = true ]; then
    echo -e "${GREEN}✅ ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
    exit 1
fi