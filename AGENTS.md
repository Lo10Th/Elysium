# AGENTS.md - Elysium Project

Guidance for AI coding agents working in this repository.

## Project Overview

Elysium is an API app store with:
- **Registry Server**: FastAPI + Supabase (Python)
- **CLI Tool (ely)**: Go client using Cobra

## Build/Run/Test Commands

### Server (Python/FastAPI)

```bash
cd server

# Setup
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt

# Run dev server
uvicorn app.main:app --reload --port 8000

# Run all tests
pytest tests/ -v

# Run single test file
pytest tests/test_auth.py -v

# Run single test function
pytest tests/test_auth.py::TestAuthRoutes::test_register_success -v

# Run tests matching pattern
pytest tests/ -k "auth" -v

# Run with coverage
pytest tests/ --cov=app --cov-report=term-missing

# Type check
mypy app/ --ignore-missing-imports

# Format/lint
black app/ && isort app/
ruff check app/
```

### CLI (Go)

```bash
cd cli

# Build
go build -o ely ./cmd

# Run all tests
go test ./... -v

# Run single test file
go test ./cmd/pull_test.go -v

# Run single test function
go test ./cmd -run TestPullSingleEmblem_Success -v

# Run tests in package
go test ./internal/emblem -v

# Format
go fmt ./...

# Lint (if golangci-lint installed)
golangci-lint run
```

### E2E Tests

```bash
./tests/e2e-test.sh
```

## Code Style Guidelines

### Python (Server)

**Imports** (standard library → third-party → local):
```python
import os
from datetime import datetime

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from app.config import settings
```

**Naming**:
- Files: `snake_case.py`
- Classes: `PascalCase`
- Functions/variables: `snake_case`
- Constants: `UPPER_SNAKE_CASE`

**Type Hints** (use modern syntax):
```python
from typing import Optional, Any

def get_emblems(category: str | None = None) -> list[Emblem]:
    ...

class User(BaseModel):
    username: str | None = None
```

**Error Handling**:
```python
from fastapi import HTTPException

if not emblem:
    raise HTTPException(status_code=404, detail="Emblem not found")
```

**Tests**:
- Class name: `TestAuthRoutes` (PascalCase)
- Method name: `test_register_success` (snake_case)
- Use pytest fixtures from `conftest.py`

### Go (CLI)

**Imports** (standard library → third-party → local):
```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/elysium/elysium/cli/internal/config"
)
```

**Naming**:
- Packages: lowercase, no underscores
- Exported functions: `PascalCase`
- Unexported functions: `camelCase`
- Constants: `UPPER_SNAKE_CASE`

**Error Handling** (wrap errors with context):
```go
if err != nil {
    return fmt.Errorf("failed to fetch emblem: %w", err)
}
```

**Cobra Commands**:
```go
var myCmd = &cobra.Command{
    Use:   "mycommand <arg>",
    Short: "Brief description",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
        return nil
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

**Tests**:
- File: `pull_test.go` (matches `pull.go`)
- Function: `TestPullSingleEmblem_Success` (PascalCase)
- Use `t.Helper()` for helper functions
- Table-driven tests preferred for multiple cases

### YAML (Emblems)

```yaml
apiVersion: v1
name: emblem-name
version: 1.0.0

# Use 2 spaces for indentation
```

## Project Structure

```
elysium/
├── server/              # FastAPI backend
│   ├── app/
│   │   ├── routes/     # API endpoints
│   │   ├── services/   # Business logic
│   │   ├── models.py   # Pydantic models
│   │   └── database.py # Supabase client
│   └── tests/          # Pytest tests
│
├── cli/                # Go CLI
│   ├── cmd/            # Cobra commands
│   └── internal/       # Internal packages
│       ├── api/        # Registry client
│       ├── config/     # State management
│       ├── emblem/     # YAML parser
│       └── executor/   # HTTP executor
│
└── schemas/
    └── emblem.schema.json
```

## Common Tasks

### Add New CLI Command
1. Create `cli/cmd/newcommand.go`
2. Add `RunE` function with `cobra.Command`
3. Register in `init()` with `rootCmd.AddCommand()`
4. Create `cli/cmd/newcommand_test.go`
5. Run tests: `go test ./cmd -v`

### Add New Server Endpoint
1. Create route in `server/app/routes/`
2. Define Pydantic models in `models.py`
3. Add service logic in `services/`
4. Create test in `server/tests/`
5. Run tests: `pytest tests/ -v`

### Add New Emblem Action Type
1. Update `schemas/emblem.schema.json`
2. Add parsing in `cli/internal/emblem/parser.go`
3. Add execution in `cli/internal/executor/runner.go`
4. Add unit tests

## Security Notes

- Never log/auth tokens or secrets
- URLs must be http/https only (SSRF prevention)
- Use parameterized queries (SQL injection prevention)
- Input validation with Pydantic validators
- Request timeouts enforced in executor

## Debugging

### Server
```bash
uvicorn app.main:app --reload --log-level debug
```

### CLI
```bash
ely pull emblem-name -v
cat ~/.elysium/config.yaml
```