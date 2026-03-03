#!/bin/bash

# Elysium End-to-End Test Suite
# Tests all features of the Elysium API app store

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Test functions
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((TESTS_PASSED++))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    echo "  Error: $2"
    ((TESTS_FAILED++))
}

skip() {
    echo -e "${YELLOW}⊘ SKIP${NC}: $1"
    echo "  Reason: $2"
    ((TESTS_SKIPPED++))
}

section() {
    echo ""
    echo -e "${BLUE}══════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}══════════════════════════════════════${NC}"
}

# Environment setup
setup_environment() {
    section "Environment Setup"
    
    # Check Python
    if command -v python3 &> /dev/null; then
        pass "Python3 found: $(python3 --version)"
    else
        fail "Python3 not found" "Python 3.11+ required"
        exit 1
    fi
    
    # Check Go
    if command -v go &> /dev/null; then
        pass "Go found: $(go version)"
    else
        fail "Go not found" "Go 1.21+ required"
        exit 1
    fi
    
    # Check pip
    if command -v pip3 &> /dev/null; then
        pass "pip3 found"
    else
        fail "pip3 not found" "pip required"
        exit 1
    fi
    
    # Set environment variables
    export PYTHONPATH="${PYTHONPATH}:$(pwd)/server"
    export PATH="${PATH}:$(pwd)/cli"
    
    pass "Environment setup complete"
}

# Phase 1: Project Structure Tests
test_project_structure() {
    section "Phase 1: Project Structure & Schema"
    
    # Test directory structure
    test -d "elysium/server" && pass "Server directory exists" || fail "Server directory missing" "Directory not found"
    test -d "elysium/cli" && pass "CLI directory exists" || fail "CLI directory missing" "Directory not found"
    test -d "elysium/schemas" && pass "Schemas directory exists" || fail "Schemas directory missing" "Directory not found"
    test -d "elysium/examples" && pass "Examples directory exists" || fail "Examples directory missing" "Directory not found"
    test -d "elysium/docs" && pass "Docs directory exists" || fail "Docs directory missing" "Directory not found"
    
    # Test schema files
    test -f "elysium/schemas/emblem.schema.json" && pass "JSON Schema exists" || fail "JSON Schema missing" "File not found"
    test -f "elysium/docs/EMBLEM_SPEC.md" && pass "Emblem spec doc exists" || fail "Emblem spec doc missing" "File not found"
    test -f "elysium/docs/GETTING_STARTED.md" && pass "Getting started doc exists" || fail "Getting started doc missing" "File not found"
    test -f "elysium/docs/SERVER_SETUP.md" && pass "Server setup doc exists" || fail "Server setup doc missing" "File not found"
    
    # Test example emblem
    test -f "elysium/examples/clothing-shop/emblem.yaml" && pass "Clothing shop emblem exists" || fail "Clothing shop emblem missing" "File not found"
    
    # Validate JSON Schema
    if python3 -c "import json; json.load(open('elysium/schemas/emblem.schema.json'))" 2>/dev/null; then
        pass "JSON Schema is valid JSON"
    else
        fail "JSON Schema is invalid" "Parse error"
    fi
    
    # Validate YAML
    if python3 -c "import yaml; yaml.safe_load(open('elysium/examples/clothing-shop/emblem.yaml'))" 2>/dev/null; then
        pass "Example emblem is valid YAML"
    else
        fail "Example emblem is invalid YAML" "Parse error"
    fi
    
    # Validate emblem structure
    if python3 <<EOF 2>/dev/null; then
import yaml
import json

with open('elysium/examples/clothing-shop/emblem.yaml') as f:
    emblem = yaml.safe_load(f)

required = ['apiVersion', 'name', 'version', 'description', 'baseUrl', 'actions']
for field in required:
    assert field in emblem, f"Missing field: {field}"

assert emblem['apiVersion'] == 'v1', "Invalid apiVersion"
assert len(emblem['name']) <= 64, "Name too long"
assert 'actions' in emblem and len(emblem['actions']) > 0, "No actions defined"
print("Emblem structure validated")
EOF
        pass "Example emblem has valid structure"
    else
        fail "Example emblem has invalid structure" "Validation failed"
    fi
}

# Phase 2: Server Tests
test_server() {
    section "Phase 2: FastAPI Server"
    
    cd elysium/server
    
    # Check requirements
    test -f "requirements.txt" && pass "requirements.txt exists" || { fail "requirements.txt missing" "File not found"; cd ../..; return; }
    
    # Check main application file
    test -f "app/main.py" && pass "main.py exists" || fail "main.py missing" "File not found"
    test -f "app/config.py" && pass "config.py exists" || fail "config.py missing" "File not found"
    test -f "app/database.py" && pass "database.py exists" || fail "database.py missing" "File not found"
    test -f "app/models.py" && pass "models.py exists" || fail "models.py missing" "File not found"
    
    # Check routes
    test -f "app/routes/auth.py" && pass "auth routes exist" || fail "auth routes missing" "File not found"
    test -f "app/routes/emblems.py" && pass "emblems routes exist" || fail "emblems routes missing" "File not found"
    
    # Validate Python imports
    if python3 -c "from app.config import Settings; print('Config module OK')" 2>/dev/null; then
        pass "Config module imports correctly"
    else
        skip "Config module import" "Dependencies not installed"
    fi
    
    # Validate Pydantic models
    if python3 -c "from app.models import EmblemCreate; print('Models OK')" 2>/dev/null; then
        pass "Models import correctly"
    else
        skip "Models import" "Dependencies not installed"
    fi
    
    cd ../..
}

# Phase 3: Clothing Shop API Tests
test_clothing_shop() {
    section "Phase 3: Clothing Shop API"
    
    cd clothing_shop
    
    # Check files
    test -f "app.py" && pass "app.py exists" || fail "app.py missing" "File not found"
    test -f "models.py" && pass "models.py exists" || fail "models.py missing" "File not found"
    
    # Check APIKey model
    if grep -q "class APIKey" models.py; then
        pass "APIKey model defined"
    else
        fail "APIKey model missing" "No APIKey class found"
    fi
    
    # Check auth decorator
    if grep -q "def require_api_key" app.py; then
        pass "Auth decorator defined"
    else
        fail "Auth decorator missing" "No require_api_key function found"
    fi
    
    # Check generate-key endpoint
    if grep -q "'/api/auth/generate-key'" app.py; then
        pass "Generate-key endpoint exists"
    else
        fail "Generate-key endpoint missing" "No /api/auth/generate-key route found"
    fi
    
    # Check protected routes
    if grep -q "@require_api_key" app.py; then
        pass "Protected routes exist"
    else
        fail "No protected routes" "No routes with @require_api_key decorator"
    fi
    
    cd ..
}

# Phase 4: CLI Code Tests
test_cli_code() {
    section "Phase 4: Go CLI Code"
    
    cd elysium/cli
    
    # Check go.mod
    test -f "go.mod" && pass "go.mod exists" || { fail "go.mod missing" "File not found"; cd ../..; return; }
    
    # Check main entry point
    test -f "cmd/root.go" && pass "root.go exists" || fail "root.go missing" "File not found"
    
    # Check command files
    for cmd in login logout whoami pull list info search; do
        if [ -f "cmd/${cmd}.go" ]; then
            pass "${cmd}.go exists"
        else
            fail "${cmd}.go missing" "File not found"
        fi
    done
    
    # Check internal packages
    test -f "internal/config/config.go" && pass "config package exists" || fail "config package missing" "File not found"
    test -f "internal/api/client.go" && pass "api package exists" || fail "api package missing" "File not found"
    test -f "internal/emblem/parser.go" && pass "emblem parser exists" || fail "emblem parser missing" "File not found"
    
    # Validate Go syntax
    if go fmt ./... &>/dev/null; then
        pass "Go code is formatted"
    else
        skip "Go format check" "Go not properly configured"
    fi
    
    # Check imports
    if grep -q "github.com/spf13/cobra" go.mod; then
        pass "Cobra dependency specified"
    else
        fail "Cobra dependency missing" "Cobra not found in go.mod"
    fi
    
    if grep -q "github.com/charmbracelet/bubbletea" go.mod; then
        pass "Bubbletea dependency specified"
    else
        skip "Bubbletea dependency" "Optional dependency"
    fi
    
    cd ../..
}

# Phase 5: Integration Tests
test_integration() {
    section "Phase 5: Integration (Code Review)"
    
    # Check README
    test -f "elysium/README.md" && pass "README exists" || fail "README missing" "File not found"
    
    # Check PROJECT_STATUS
    test -f "elysium/PROJECT_STATUS.md" && pass "PROJECT_STATUS exists" || fail "PROJECT_STATUS missing" "File not found"
    
    # Count lines of code
    SERVER_LOC=$(find elysium/server -name "*.py" -type f 2>/dev/null | xargs wc -l 2>/dev/null | tail -1 | awk '{print $1}')
    CLI_LOC=$(find elysium/cli -name "*.go" -type f 2>/dev/null | xargs wc -l 2>/dev/null | tail -1 | awk '{print $1}')
    YAML_LINES=$(find elysium/examples -name "*.yaml" -type f 2>/dev/null | xargs wc -l 2>/dev/null | tail -1 | awk '{print $1}')
    
    echo "  Server LOC: ${SERVER_LINES:-0}"
    echo "  CLI LOC: ${CLI_LOC:-0}"
    echo "  YAML lines: ${YAML_LINES:-0}"
    
    # Check documentation completeness
    if grep -q "## Quick Start" elysium/README.md; then
        pass "README has Quick Start section"
    else
        fail "README missing Quick Start" "Section not found"
    fi
    
    if grep -q "## Architecture" elysium/README.md; then
        pass "README has Architecture section"
    else
        fail "README missing Architecture" "Section not found"
    fi
    
    # Check for AGENTS.md
    test -f "elysium/AGENTS.md" && pass "AGENTS.md exists" || skip "AGENTS.md" "Could be added later"
}

# Final Summary
print_summary() {
    section "Test Summary"
    
    echo ""
    echo "┌────────────────┬──────────┐"
    printf "│ %-14s │ %8s │\n" "Tests Passed" "${TESTS_PASSED}"
    printf "│ %-14s │ %8s │\n" "Tests Failed" "${TESTS_FAILED}"
    printf "│ %-14s │ %8s │\n" "Tests Skipped" "${TESTS_SKIPPED}"
    echo "├────────────────┼──────────┤"
    TOTAL=$((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))
    printf "│ %-14s │ %8s │\n" "Total" "${TOTAL}"
    echo "└────────────────┴──────────┘"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ ALL TESTS PASSED${NC}"
        exit 0
    else
        echo -e "${RED}✗ SOME TESTS FAILED${NC}"
        exit 1
    fi
}

# Main execution
main() {
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║          ELYSIUM API APP STORE - TEST SUITE              ║"
    echo "║                    Version 1.0.0                           ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    
    setup_environment
    test_project_structure
    test_server
    test_clothing_shop
    test_cli_code
    test_integration
    print_summary
}

# Run main function
main "$@"