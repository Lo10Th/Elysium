# Elysium Development Roadmap

## Priority Levels
- 🔴 **P0 - Critical**: Blocks core functionality
- 🟠 **P1 - High**: Essential for MVP
- 🟡 **P2 - Medium**: Important for polish
- 🟢 **P3 - Low**: Nice to have

---

## Phase 1: Core Functionality (Week 1-2) 🔴

### 1.1 Dynamic Emblem Execution 🔴
**Current State**: Executor code exists but not integrated  
**What's Needed**:

```go
// Add to cmd/root.go - Dynamic command generation
func ExecuteEmblemCommand(emblemName string, action string, flags map[string]string) error {
    // 1. Load emblem from cache
    def, err := emblem.LoadFromCache(emblemName, "latest")
    
    // 2. Get action definition
    act, err := def.GetAction(action)
    
    // 3. Parse flags into parameters
    params := executor.ParseParams(flags)
    
    // 4. Inject auth credentials
    creds, err := def.GetAuthCredentials()
    headers := buildAuthHeaders(creds)
    
    // 5. Execute HTTP request
    result, err := executor.New(def).Execute(action, params, headers)
    
    // 6. Format and display output
    return formatOutput(result, outputFormat)
}
```

**Files to Create/Modify**:
- `cli/cmd/exec.go` - New command for `ely <emblem> <action>`
- `cli/cmd/root.go` - Add dynamic command routing
- `cli/internal/executor/runner.go` - Fix and complete
- `cli/internal/output/*.go` - Format output beautifully

**Tasks**:
- [ ] Create `cmd/exec.go` with dynamic subcommand routing
- [ ] Wire executor into CLI commands
- [ ] Add flag generation from emblem parameters
- [ ] Implement output formatting (table/JSON)
- [ ] Test with clothing-shop emblem
- [ ] Add error handling and validation

---

### 1.2 Interactive Mode 🔴
**Problem**: Complex actions with many parameters are hard to use

```bash
# Current: Error-prone typing
ely clothing-shop create-order \
  --customer-name "John Doe" \
  --customer-email "john@example.com" \
  --items '[{"product_id": 1, "quantity": 2}]'  # JSON on CLI?!
  
# Better: Interactive mode
ely clothing-shop create-order --interactive
? Customer name: John Doe
? Customer email: john@example.com  
? Select products:
  ✓ Product 1 (Vintage T-Shirt) - Qty: 2
  ✓ Product 3 (Hoodie) - Qty: 1
```

**Implementation**:
```go
// Use bubbletea for interactive prompts
package interactive

import (
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
)

func PromptForParams(params []emblem.Parameter) map[string]interface{} {
    // For each parameter:
    // - If enum: show multi-select
    // - If boolean: show toggle
    // - If string/number: prompt
    // - If array: show multi-input
}
```

**Tasks**:
- [ ] Create `cli/internal/interactive/prompt.go`
- [ ] Implement parameter prompting
- [ ] Add `--interactive` flag support
- [ ] Add validation during prompting
- [ ] Test complex parameters

---

### 1.3 Auth Token Management 🔴
**Current State**: `ely login` opens browser but doesn't save token properly

**What's Needed**:
```
Flow:
1. ely login
   → Opens browser to /auth/login
   → Local server receives callback with token
   → Saves to keyring
   
2. Token usage
   → Reads from keyring on every request
   → Refreshes if expired
   → Prompts re-login if invalid

3. Token refresh
   → Automatic background refresh
   → Retry with refreshed token on 401
```

**Files to Create**:
- `cli/internal/auth/token.go` - Token management
- `cli/internal/auth/oauth.go` - OAuth callback server
- `cli/cmd/login.go` - Fix to actually work

**Tasks**:
- [ ] Implement OAuth callback server (localhost:port)
- [ ] Save token to OS keyring
- [ ] Implement token refresh logic
- [ ] Add token expiry checking
- [ ] Handle 401 errors gracefully

---

## Phase 2: Developer Experience (Week 3-4) 🟠

### 2.1 Emblem Scaffolding 🟠
**Problem**: Creating emblems from scratch is tedious

```bash
ely init my-api --category payments
# Creates:
my-api/
├── emblem.yaml          # Template
├── README.md            # Usage docs
├── examples/            # Example requests
└── .emblemignore       # Ignore patterns
```

**Implementation**:
- [ ] Create `cmd/init.go`
- [ ] Generate template from category
- [ ] Include common actions
- [ ] Add validation comments

---

### 2.2 Local Emblem Development 🟠
**Problem**: Can't test emblems locally before publishing

```bash
ely validate ./my-api/emblem.yaml
ely test ./my-api/ --action list-users
ely run ./my-api/ --local  # Run without publishing
```

**Tasks**:
- [ ] Create `cmd/validate.go`
- [ ] Implement schema validation
- [ ] Add `--local` flag to run without registry
- [ ] Test against local servers

---

### 2.3 Better Error Messages 🟠
**Current**: Generic errors  
**Desired**: Contextual, actionable errors

```bash
# ❌ Current
Error: request failed

# ✅ Better
Error: Failed to connect to API (http://localhost:5000)
  Reason: Connection refused
  Suggestion: Is the API server running?
               Try: python clothing_shop/app.py

# ❌ Current  
Error: authentication required

# ✅ Better
Error: API key required for this action
  Required: CLOTHING_SHOP_API_KEY
  Suggestion: Generate key with:
               curl -X POST http://localhost:5000/api/auth/generate-keys
               Then: export CLOTHING_SHOP_API_KEY=your-key
```

**Tasks**:
- [ ] Create error type hierarchy
- [ ] Add context to all errors
- [ ] Include suggestions
- [ ] Add error codes for scripts

---

### 2.4 Configuration Management 🟠
**Problem**: Can't configure per-emblem settings

```bash
# ~/.elysium/config.yaml
registry: https://registry.elysium.dev
output: table
cache_dir: ~/.elysium/cache

# ~/.elysium/emblems.yaml
defaults:
  clothing-shop:
    api_key: ${CLOTHING_SHOP_API_KEY}
    timeout: 30s
  stripe:
    api_key: ${STRIPE_API_KEY}
    sandbox: true
```

**Tasks**:
- [ ] Create per-emblem config
- [ ] Add `ely config` command
- [ ] Support environment variable interpolation
- [ ] Add configuration validation

---

## Phase 3: Production Readiness (Week 5-6) 🟡

### 3.1 Backend Improvements 🟡

#### Unify to FastAPI
```bash
# Migrate clothing-shop to FastAPI
clothing_shop/
├── app/
│   ├── main.py          # FastAPI app
│   ├── models.py        # SQLAlchemy models
│   ├── auth.py          # Auth middleware
│   └── routers/
│       ├── products.py
│       ├── orders.py
│       └── auth.py
├── requirements.txt     # FastAPI stack
└── tests/
```

**Tasks**:
- [ ] Convert clothing-shop from Flask to FastAPI
- [ ] Add OpenAPI documentation
- [ ] Add rate limiting
- [ ] Add request validation
- [ ] Add async support

---

#### Add Rate Limiting
```python
from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

@app.route("/api/emblems")
@limiter.limit("100/minute")
def list_emblems():
    ...
```

**Tasks**:
- [ ] Add slowapi to server
- [ ] Configure per-endpoint limits
- [ ] Add rate limit headers
- [ ] Document limits

---

### 3.2 Testing Infrastructure 🟡

#### Go CLI Tests
```go
// cli/internal/emblem/parser_test.go
func TestParseEmblem(t *testing.T) {
    tests := []struct{
        name string
        yaml string
        wantErr bool
    }{
        {"valid", "apiVersion: v1\nname: test", false},
        {"invalid", "apiVersion: v2", true},
    }
    // ...
}
```

**Tasks**:
- [ ] Add unit tests for each package
- [ ] Add integration tests
- [ ] Set up CI testing
- [ ] Add coverage reporting

#### Python Server Tests
```python
# tests/test_emblems.py
def test_create_emblem(client, auth_header):
    response = client.post("/api/emblems", 
        headers=auth_header,
        json={"name": "test", ...}
    )
    assert response.status_code == 201
```

**Tasks**:
- [ ] Add pytest fixtures
- [ ] Add test database
- [ ] Test all endpoints
- [ ] Add test coverage

---

### 3.3 Security Hardening 🟡

#### Input Validation
```go
// Validate all emblem inputs
func ValidateParam(param Parameter, value interface{}) error {
    if param.Required && value == nil {
        return RequiredError{Field: param.Name}
    }
    if param.Enum != nil {
        if !contains(param.Enum, value) {
            return InvalidEnumError{...}
        }
    }
    // Type checking, range checking, etc.
}
```

**Tasks**:
- [ ] Add input validation layer
- [ ] Sanitize all user inputs
- [ ] Add SQL injection prevention
- [ ] Add XSS prevention
- [ ] Audit dependencies

#### Token Security
- [ ] Implement token rotation
- [ ] Add scope-based permissions
- [ ] Audit token storage
- [ ] Add session management

---

## Phase 4: Distribution (Week 7-8) 🟡

### 4.1 Build Pipeline 🟡

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        arch: [amd64, arm64]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go build -o ely-${{ matrix.os }}-${{ matrix.arch }}
      - uses: actions/upload-artifact@v3
```

**Tasks**:
- [ ] Create GitHub Actions workflow
- [ ] Build for all platforms
- [ ] Upload to GitHub Releases
- [ ] Create SHA256 checksums
- [ ] Sign binaries

---

### 4.2 Homebrew Formula 🟡

```ruby
# Formula/ely.rb
class Ely < Formula
  desc "API App Store CLI"
  homepage "https://elysium.dev"
  version "1.0.0"
  url "https://github.com/elysium/elysium/releases/download/v1.0.0/ely-darwin-amd64"
  sha256 "..."
  
  def install
    bin.install "ely"
  end
  
  test do
    assert_match "Elysium", shell_output("#{bin}/ely --version")
  end
end
```

**Tasks**:
- [ ] Create Homebrew tap
- [ ] Write formula
- [ ] Test installation
- [ ] Document in README

---

### 4.3 Install Script 🟢

```bash
#!/bin/bash
# install.sh
set -e

OS=$(uname -s)
ARCH=$(uname -m)
VERSION="latest"

# Download correct binary
URL="https://get.elysium.dev/${OS}/${ARCH}/ely"
curl -sSL $URL -o ely
chmod +x ely
sudo mv ely /usr/local/bin/

echo "✓ Ely installed successfully!"
echo "Run 'ely --help' to get started."
```

---

## Phase 5: Advanced Features (Week 9-12) 🟢

### 5.1 Web UI 🟢
- Browse emblems in browser
- View documentation
- Test actions interactively
- Manage API keys

### 5.2 Emblem Marketplace 🟢
- Featured emblems
- Categories and tags
- Reviews and ratings
- Usage statistics

### 5.3 Advanced Execution 🟢

#### Chaining Actions
```bash
ely clothing-shop get-product --id 1 --output json | \
  jq '.id' | \
  xargs -I {} ely clothing-shop create-order --product-id {}
```

#### Batch Operations
```bash
ely clothing-shop batch-create-orders --file orders.csv
```

#### Workflow Automation
```yaml
# workflow.yaml
name: Test Order Flow
steps:
  - action: list-products
    save: products
  
  - action: create-order
    with:
      product_id: ${products[0].id}
  
  - action: get-order
    with:
      id: ${create-order.id}
```

---

### 5.4 Private Registries 🟢

```bash
ely config set registry https://my-company.registry.dev
ely login --company acme
ely pull internal-api
```

---

### 5.5 Plugin System 🟢

```go
// Plugins extend functionality
type Plugin interface {
    Name() string
    Execute(ctx context.Context, emblem string, action string, params map[string]interface{}) error
}

// Example: Data transformation plugin
func TransformOutput(data []byte, format string) []byte {
    // Convert, filter, format
}
```

---

## Success Metrics

### Week 1-2 (Core)
- [ ] Can execute `ely clothing-shop list-products`
- [ ] Can create product interactively
- [ ] Auth flow works end-to-end

### Week 3-4 (DX)
- [ ] Can create emblem in < 2 minutes
- [ ] All errors have suggestions
- [ ] Config file management works

### Week 5-6 (Production)
- [ ] 80%+ test coverage
- [ ] Security audit passed
- [ ] Performance benchmarks

### Week 7-8 (Distribution)
- [ ] Binary works on Linux/Mac/Windows
- [ ] Homebrew install works
- [ ] Install script works

### Week 9+ (Advanced)
- [ ] Web UI live
- [ ] 10+ example emblems
- [ ] Plugin system documented

---

## Technical Debt to Address

1. **Unused imports in Go** - Run `goimports`
2. **Python import errors** - Fix venv setup
3. **Emblem validation failure** - Fix YAML parsing
4. **No graceful shutdown** - Add signal handling
5. **Hardcoded values** - Environment config
6. **Missing Docker setup** - Add Dockerfile
7. **No database migrations** - Add Alembic
8. **Missing API docs** - Add OpenAPI descriptions

---

## Maintenance Schedule

### Daily
- Monitor error rates
- Check security alerts
- Review issues

### Weekly
- Update dependencies
- Review PRs
- Update docs

### Monthly
- Security audit
- Performance review
- Feature prioritization

### Quarterly
- Major version planning
- Architecture review
- Community feedback analysis