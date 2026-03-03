# AGENTS.md - Elysium Project

This document provides guidance for AI coding agents working in this repository.

## Project Overview

Elysium is an API app store consisting of:
- **Registry Server**: FastAPI backend with Supabase storage
- **CLI Tool (ely)**: Go client for interacting with emblems
- **Emblem Specification**: YAML format for describing APIs

## Project Structure

```
elysium/
├── server/              # FastAPI backend
│   ├── app/
│   │   ├── routes/     # API endpoints (auth, emblems, search)
│   │   ├── models.py   # Pydantic models
│   │   ├── database.py # Supabase connection
│   │   └── config.py   # Settings
│   ├── tests/
│   └── requirements.txt
│
├── cli/                 # Go CLI
│   ├── cmd/            # Cobra commands (login, pull, search, etc.)
│   ├── internal/
│   │   ├── api/       # Registry client
│   │   ├── config/    # State management
│   │   ├── emblem/    # Parser & validator
│   │   └── executor/  # HTTP requester
│   └── go.mod
│
├── schemas/
│   └── emblem.schema.json  # JSON Schema validation
│
├── examples/
│   └── clothing-shop/
│       └── emblem.yaml      # Example emblem
│
├── docs/
│   ├── EMBLEM_SPEC.md      # Complete specification
│   ├── GETTING_STARTED.md  # User guide
│   └── SERVER_SETUP.md    # Deployment guide
│
└── tests/
    └── e2e-test.sh       # End-to-end test suite
```

## Build/Run Commands

### Server (FastAPI)

```bash
# Setup
cd server
python -m venv venv
source venv/bin/activate  # Linux/Mac
pip install -r requirements.txt

# Environment
cp .env.example .env
# Edit .env with Supabase credentials

# Run development server
uvicorn app.main:app --reload --port 8000

# Run production
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4

# Run tests
pytest tests/ -v

# Type check
mypy app/ --ignore-missing-imports

# Format
black app/
isort app/
```

### CLI (Go)

```bash
# Setup
cd cli
go mod tidy
go mod download

# Build
go build -o ely ./cmd

# Install globally
go install ./cmd

# Run tests
go test ./... -v

# Format
go fmt ./...

# Lint
golangci-lint run
```

### Clothing Shop (Example API)

```bash
# Setup
cd ../clothing_shop
python -m venv env
source env/bin/activate
pip install -r requirements.txt

# Run
python app.py
# Runs on http://localhost:5000

# Generate API key
curl -X POST http://localhost:5000/api/auth/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-key"}'
```

### Test Suite

```bash
# Run comprehensive tests
./tests/e2e-test.sh

# Test specific components
cd server && pytest tests/ -k auth
cd cli && go test ./internal/emblem -v
```

## Code Style Guidelines

### Python (Server)

**Imports:**
```python
# Standard library
import os
from datetime import datetime

# Third-party
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

# Local
from app.config import settings
```

**Naming:**
- Files: `snake_case.py`
- Classes: `PascalCase`
- Functions/variables: `snake_case`
- Constants: `UPPER_SNAKE_CASE`

**Type Hints:**
```python
from typing import List, Optional

def get_emblems(category: Optional[str] = None) -> List[Emblem]:
    ...
```

**Error Handling:**
```python
from fastapi import HTTPException

if not emblem:
    raise HTTPException(status_code=404, detail="Emblem not found")
```

### Go (CLI)

**Imports:**
```go
import (
    // Standard library
    "fmt"
    "os"
    
    // Third-party
    "github.com/spf13/cobra"
    
    // Local
    "github.com/elysium/elysium/cli/internal/config"
)
```

**Naming:**
- Packages: lowercase, no underscores
- Exported functions: `PascalCase`
- Unexported functions: `camelCase`
- Constants: `UPPER_SNAKE_CASE`

**Error Handling:**
```go
if err != nil {
    return fmt.Errorf("failed to ...: %w", err)
}
```

**Commands:**
```go
var cmd = &cobra.Command{
    Use:   "pull <name>[@version]",
    Short: "Download an emblem",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
    },
}
```

### YAML (Emblems)

**Structure:**
```yaml
apiVersion: v1
name: emblem-name
version: 1.0.0

# Use 2 spaces for indentation
# Keys in alphabetical order where possible
```

## Architecture Decisions

### Why FastAPI for Server?
- Async support for scalability
- Automatic OpenAPI documentation
- Pydantic integration for validation
- Modern Python with type hints

### Why Go for CLI?
- Fast startup time (< 100ms)
- Cross-platform compilation
- Single binary distribution
- Excellent CLI libraries (Cobra, Bubbletea)

### Why Supabase for Database?
- PostgreSQL relational database
- Built-in authentication
- Row Level Security
- Real-time subscriptions (future feature)

### Why YAML for Emblems?
- Human-readable
- Comments support
- Multi-document support
- Widely used in DevOps

## Testing Strategy

### Unit Tests
- **Server**: Pytest with fixtures
- **CLI**: Go testing package
- **Coverage**: 80%+ target

### Integration Tests
- Server endpoints with test database
- CLI commands against mock registry

### End-to-End Tests
- Full workflow: login → pull → execute
- Multiple emblem types
- Error scenarios

## Security Considerations

### Authentication
- JWT tokens stored in OS keyring
- Tokens never logged or displayed
- Secure transmission via HTTPS

### Emblem Execution
- URLs validated (http/https only)
- Request timeouts enforced
- Response size limits
- No arbitrary code execution

### Database
- Parameterized queries (SQL injection prevention)
- Row Level Security enabled
- API keys hashed (future)

## Common Tasks

### Add New Emblem Action Type
1. Update `schemas/emblem.schema.json`
2. Add parsing in `cli/internal/emblem/parser.go`
3. Add execution in `cli/internal/executor/runner.go`
4. Add tests
5. Update docs/EMBLEM_SPEC.md

### Add New CLI Command
1. Create `cli/cmd/newcommand.go`
2. Implement with `cobra.Command`
3. Add to `init()` in `cmd/root.go`
4. Add tests
5. Update README

### Add New Server Endpoint
1. Create route in `server/app/routes/`
2. Define Pydantic models in `models.py`
3. Add Supabase query logic
4. Add tests
5. Update OpenAPI docs

### Deploy Server
1. Set environment variables (Supabase, CORS origins)
2. Run database migrations
3. Start with Gunicorn/Uvicorn
4. Configure rate limiting
5. Enable logging

### Release CLI
1. Update version in `cmd/root.go`
2. Run `go mod tidy`
3. Run `./scripts/build-all.sh`
4. Create GitHub release
5. Update Homebrew formula

## Debugging

### Server
```bash
# Enable debug logging
export DEBUG=true
uvicorn app.main:app --reload --log-level debug

# View logs
tail -f logs/app.log
```

### CLI
```bash
# Enable verbose output
ely pull emblem-name -v

# Check config
cat ~/.elysium/config.yaml

# View cached emblems
ls ~/.elysium/cache/
```

### Emblem Validation
```bash
# Validate emblem YAML
python3 -c "
import yaml, json
schema = json.load(open('schemas/emblem.schema.json'))
from jsonschema import validate
with open('examples/clothing-shop/emblem.yaml') as f:
    validate(yaml.safe_load(f), schema)
    print('Valid!')
"
```

## Performance Optimization

### Server
- Use async endpoints
- Cache frequent queries
- Implement rate limiting
- Connection pooling

### CLI
- Lazy load emblems
- Cache registry responses
- Parallel downloads for multiple emblems
- Minimize startup time

## Known Issues

1. **Import errors in IDE**: Dependencies not installed - run `pip install` and `go mod download`
2. **Supabase connection fails**: Check `.env` configuration
3. **Go build fails**: Run `go mod tidy` first
4. **Emblem validation fails**: Check YAML syntax and schema compliance

## Future Improvements

- [ ] GraphQL support for complex queries
- [ ] Web UI for browsing emblems
- [ ] Private namespaces
- [ ] Team collaboration
- [ ] API usage analytics
- [ ] Code generation (SDKs)

## Resources

- [FastAPI Docs](https://fastapi.tiangolo.com/)
- [Cobra Docs](https://github.com/spf13/cobra)
- [Supabase Docs](https://supabase.com/docs)
- [JSON Schema](https://json-schema.org/)

## Contact

- GitHub Issues: https://github.com/Lo10Th/Elysium/issues