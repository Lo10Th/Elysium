# Contributing to Elysium

Thank you for your interest in contributing to Elysium! This guide is optimized for both human and LLM contributors.

## Quick Start for LLM Contributors

When contributing to this repository, please:

1. **Read this file first** - It contains all project-specific conventions
2. **Check PROJECT_STATUS.md** - Understand current implementation state
3. **Follow the patterns** - Match existing code style in nearby files
4. **Reference AGENTS.md** - For detailed technical context

## Development Setup

### Prerequisites

- Go 1.21+
- Python 3.11+
- Node.js 18+ (for Supabase CLI)
- Git

### Installation

```bash
# Clone repository
git clone https://github.com/Lo10Th/Elysium.git
cd Elysium

# Install Go CLI dependencies
cd cli && go mod download

# Install Python server dependencies
cd ../server && pip install -r requirements.txt

# Install Supabase CLI (for local development)
npm install -g supabase
```

### Environment Variables

Create `.env` files for local development:

**CLI (.env in cli/ directory):**
```env
ELYSIUM_API_URL=http://localhost:8000
ELYSIUM_REGISTRY_URL=http://localhost:8000
```

**Server (.env in server/ directory):**
```env
SUPABASE_URL=your_supabase_url
SUPABASE_ANON_KEY=your_supabase_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key
DATABASE_URL=your_database_url
SECRET_KEY=your_secret_key
```

## Code Style

### Go (CLI)

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Package names: lowercase, single word
- Exported functions: must have comments
- Error handling: return errors, don't panic
- Use `cobra` for CLI commands
- Use `viper` for configuration
- Use `bubbletea` for TUI components

Example:
```go
// Execute runs an emblem action with the given parameters.
func (e *Executor) Execute(ctx context.Context, action *Action, params map[string]interface{}) error {
    if err := e.validate(action); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return e.run(ctx, action, params)
}
```

### Python (Server)

- Use `black` for formatting
- Use `isort` for import sorting
- Use `ruff` for linting
- Type hints: required for all functions
- Docstrings: Google style
- Use Pydantic for data validation
- Use FastAPI decorators and dependency injection

Example:
```python
from typing import Optional
from pydantic import BaseModel

class EmblemCreate(BaseModel):
    """Request model for creating an emblem."""
    name: str
    version: str
    description: Optional[str] = None
    
    class Config:
        json_schema_extra = {
            "example": {
                "name": "clothing-shop",
                "version": "1.0.0",
                "description": "API for clothing store"
            }
        }
```

### YAML (Emblems)

- Use 2-space indentation
- Quote strings with special characters
- Include `version`, `name`, `description` at minimum
- Validate against `schemas/emblem.schema.json`

Example:
```yaml
version: "1.0"
name: my-api
description: My API description
baseUrl: https://api.example.com
auth:
  type: api_key
  location: header
  key_name: X-API-KEY
actions:
  list-items:
    method: GET
    path: /items
    description: List all items
```

## Testing

### Test Requirements

- **80% code coverage** is required for all new code
- All tests must pass before merging
- Unit tests for all public functions
- Integration tests for API endpoints

### Running Tests

```bash
# Go tests
cd cli
go test ./... -v -race -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out

# Python tests
cd server
pytest --cov=app --cov-report=html tests/
```

### Test Naming Convention

- Go: `Test<FunctionName>_<Scenario>_<ExpectedResult>`
- Python: `test_<function_name>_<scenario>_<expected_result>`

Example:
```go
func TestExecute_ValidAction_ReturnsNil(t *testing.T) {}
func TestExecute_InvalidAction_ReturnsError(t *testing.T) {}
```

```python
def test_create_emblem_valid_data_returns_emblem():
    pass

def test_create_emblem_missing_name_returns_error():
    pass
```

## Git Workflow

### Branch Naming

- `feat/issue-123-short-description` - New features
- `fix/issue-456-short-description` - Bug fixes
- `docs/issue-789-short-description` - Documentation
- `refactor/short-description` - Code refactoring
- `test/short-description` - Adding tests

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
```
feat(cli): add dynamic emblem execution

Implement executor.Run() to execute emblem actions dynamically.
Supports HTTP methods, parameter substitution, and auth injection.

Closes #1
```

```
fix(server): handle missing auth header gracefully

Return 401 instead of 500 when auth header is missing.

Fixes #3
```

### Pull Request Process

1. **Create branch** from `main`
2. **Make changes** following code style
3. **Add/update tests** maintaining 80% coverage
4. **Update documentation** if needed
5. **Run tests locally**: `./scripts/test.sh`
6. **Check coverage**: `./scripts/coverage.sh`
7. **Create PR** referencing the issue
8. **Ensure CI passes**
9. **Request review**
10. **Address feedback**

### PR Title Format

Use the same format as commit messages:
```
feat(cli): add dynamic emblem execution
fix(server): handle missing auth header gracefully
```

## Release Process

Releases are fully automated via GitHub Actions:

1. **Ensure main branch is stable**
2. **Update CHANGELOG.md**
3. **Create and push a tag**:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
4. **GitHub Actions will**:
   - Run tests
   - Build binaries for all platforms
   - Create GitHub release
   - Upload release assets

### Version Numbering

Following [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

Pre-release versions:
- `v0.1.0-alpha.1` - Internal testing
- `v0.1.0-beta.1` - Public testing
- `v0.1.0-rc.1` - Release candidate

## Project Structure

```
elysium/
├── cli/                    # Go CLI application
│   ├── cmd/               # Cobra commands
│   ├── internal/          # Business logic
│   │   ├── api/          # API client
│   │   ├── config/       # Configuration
│   │   ├── emblem/       # Emblem handling
│   │   └── executor/     # Execution engine
│   └── main.go           # Entry point
│
├── server/                # FastAPI server
│   ├── app/
│   │   ├── main.py       # FastAPI app
│   │   ├── config.py     # Settings
│   │   ├── models.py     # Pydantic models
│   │   ├── routes/       # API endpoints
│   │   └── services/     # Business logic
│   └── tests/            # Test files
│
├── schemas/              # JSON schemas
├── examples/             # Example emblems
├── docs/                 # Documentation
├── scripts/              # Automation scripts
└── .github/workflows/    ## CI/CD workflows
```

## Documentation Standards

### README.md

- Clear project description
- Installation instructions
- Quick start guide
- Link to detailed docs

### AGENTS.md

- Technical architecture overview
- Current implementation status
- Key design decisions
- Known issues and blockers

### Code Comments

- **Why**, not *what*
- Explain non-obvious logic
- Reference issues for workarounds

```go
// Using pointer to allow nil values in API responses
// See: https://github.com/Lo10Th/Elysium/issues/42
OptionalField *string `json:"optional_field"`
```

## Issue Management

### Issue Priorities

- **P0**: Critical, blocks core functionality
- **P1**: High, required for MVP
- **P2**: Medium, nice to have
- **P3**: Low, future enhancement

### Issue Template

```markdown
## Description
[Clear description of the feature/bug]

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2

## Technical Notes
[Implementation hints]

## Related Issues
[Link to related issues]
```

## Getting Help

- Check existing issues before creating new ones
- Reference related issues in PRs
- Tag maintainers for urgent reviews
- Follow the code of conduct

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn
- Celebrate contributions

---

For questions or clarifications, please open an issue or reach out to the maintainers.

Thank you for contributing to Elysium!