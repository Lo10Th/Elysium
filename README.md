# Elysium - The API App Store

[![Test](https://github.com/Lo10Th/Elysium/actions/workflows/test.yml/badge.svg)](https://github.com/Lo10Th/Elysium/actions/workflows/test.yml)
[![Release](https://github.com/Lo10Th/Elysium/actions/workflows/release.yml/badge.svg)](https://github.com/Lo10Th/Elysium/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Lo10Th/Elysium)](https://goreportcard.com/report/github.com/Lo10Th/Elysium)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Version**: 0.1.0  
**Status**: In Development

Elysium is an API app store that allows developers and AI agents to discover, download, and use APIs programmatically through defined emblems—YAML files that describe an API's endpoints, parameters, authentication, and types.

## The Problem

APIs are everywhere, but using them requires:
- Reading extensive documentation
- Understanding authentication flows
- Writing boilerplate HTTP client code
- Managing API keys and environment variables
- Handling rate limits, errors, and edge cases

## The Solution

**Emblems** are machine-readable API definitions that enable:
- 🤖 **AI Agents** to use APIs without human intervention
- 👨‍💻 **Developers** to skip documentation dive
- 🔄 **Automation** of complex API workflows
- 📦 **Version control** for API integrations

## Quick Start

### Installation

```bash
# One-line install (Linux/macOS)
curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash

# Or with wget
wget -qO- https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash

# Install a specific version
curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash -s -- --version v0.1.0

# Using Homebrew (macOS / Linux)
brew tap Lo10Th/tap
brew install ely

# Using Go install
go install github.com/Lo10Th/Elysium/cli/cmd/ely@latest

# Download binary from GitHub Releases
# Visit: https://github.com/Lo10Th/Elysium/releases
```

### Pull and Use an Emblem

```bash
# Authenticate
ely login

# Pull emblem
ely pull clothing-shop

# View available actions
ely info clothing-shop

# Execute actions
ely pull clothing-shop
export CLOTHING_SHOP_API_KEY=your-api-key

# List products
ely clothing-shop list-products

# Create product
ely clothing-shop create-product \
  --name "Vintage T-Shirt" \
  --price 29.99 \
  --category shirts \
  --size M \
  --color blue

# Place order
ely clothing-shop create-order \
  --customer-name "John Doe" \
  --customer-email "john@example.com" \
  --customer-address "123 Main St" \
  --items '[{"product_id": 1, "quantity": 2}]'
```

## Architecture

```
┌─────────────────┐      ┌──────────────────┐      ┌─────────────────┐
│                 │      │                  │      │                 │
│   CLI (ely)     │──────▶   Registry      │◀──────│   Developer     │
│                 │      │   (Supabase)     │      │   (You)         │
│   - Pull        │      │                  │      │                 │
│   - Execute     │      │   - Store        │      │   - Publish     │
│   - Search      │      │   - Search       │      │   - Version     │
│                 │      │   - Auth          │      │                 │
└─────────────────┘      └──────────────────┘      └─────────────────┘
        │                         │                         
        │                         │                         
        ▼                         │                         
┌─────────────────┐               │                         
│   Emblem YAML   │               │                         
│                 │               │                         
│   - Actions     │               │                         
│   - Types       │               │                         
│   - Auth        │               │                         
└─────────────────┘               │                         
        │                         │                         
        ▼                         │                         
┌─────────────────┐               │                         
│   Any API       │◀──────────────┘                         
│                 │                                         
│   - REST        │                                         
│   - GraphQL     │                                         
│   - WebSocket   │                                         
└─────────────────┘                                         
```

## Project Structure

```
elysium/
├── server/                 # FastAPI registry backend
│   ├── app/
│   │   ├── routes/         # API endpoints
│   │   ├── models.py       # Pydantic models
│   │   ├── database.py     # Supabase client
│   │   └── config.py       # Settings
│   └── requirements.txt
│
├── cli/                    # Go CLI tool
│   ├── cmd/               # Cobra commands
│   ├── internal/
│   │   ├── api/          # Registry client
│   │   ├── config/       # State management
│   │   ├── emblem/       # Parser & validator
│   │   └── executor/     # HTTP requester
│   └── go.mod
│
├── schemas/
│   └── emblem.schema.json # JSON Schema for validation
│
├── examples/
│   └── clothing-shop/
│       └── emblem.yaml    # Complete example
│
└── docs/
    ├── EMBLEM_SPEC.md     # Full specification
    ├── GETTING_STARTED.md # User guide
    └── SERVER_SETUP.md    # Deployment guide
```

## Emblem Specification

```yaml
apiVersion: v1
name: clothing-shop
version: 1.0.0
description: REST API for online clothing store
baseUrl: https://api.clothing-shop.example.com

auth:
  type: api_key
  keyEnv: CLOTHING_SHOP_API_KEY
  header: X-API-Key

types:
  Product:
    properties:
      id: { type: integer }
      name: { type: string }
      price: { type: number }

actions:
  list-products:
    description: List all products
    method: GET
    path: /products
    parameters:
      - name: category
        type: string
        in: query
        
  create-product:
    description: Create a product
    method: POST
    path: /products
    parameters:
      - name: name
        type: string
        in: body
        required: true
```

**Full specification**: [docs/EMBLEM_SPEC.md](docs/EMBLEM_SPEC.md)

## CLI Reference

### Authentication

```bash
ely login                 # Browser-based login
ely logout                # Remove credentials
ely whoami               # Show current user
```

### Discovery

```bash
ely search <query>       # Search emblems
ely info <name>          # View emblem details
ely list                  # List installed emblems
```

### Installation

```bash
ely pull <name>[@version] # Pull emblem
ely init <name>           # Create new emblem scaffold
ely validate ./emblem.yaml # Validate emblem
ely test ./<dir>/         # Test emblem locally
```

### Execution

```bash
ely <emblem-name> <action> [flags]

# Examples:
ely clothing-shop list-products
ely clothing-shop get-product --id 1
ely stripe create-customer --email "test@example.com"
```

### API Keys

```bash
ely keys list              # List your API keys
ely keys create <name>     # Create new API key
ely keys delete <id>       # Delete API key
ely keys show <id>         # Show key details
```

### Planned Commands (Not Yet Implemented)

The following commands are planned for future releases:

```bash
ely update <name>         # Update emblem to latest version
ely remove <name>         # Uninstall emblem
ely publish ./<dir>/      # Publish emblem to registry
ely completion bash       # Shell completion
```

## Server Endpoints

### Authentication

```
POST /api/auth/register   # Create account
POST /api/auth/login      # Login
POST /api/auth/logout     # Logout
POST /api/auth/refresh    # Refresh token
GET  /api/auth/me         # Current user
```

### Emblems

```
GET  /api/emblems                    # List all
GET  /api/emblems?category=payments  # Filter
GET  /api/emblems/:name              # Get emblem
GET  /api/emblems/:name/:version     # Get version
POST /api/emblems                    # Publish new (planned)
PUT  /api/emblems/:name              # New version (planned)
DELETE /api/emblems/:name           # Delete (planned)
```

### Search

```
GET /api/emblems/search?q=query&sort=downloads&limit=20
```

## Setup Guide

### Server (Backend)

```bash
cd server
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
# Edit .env with Supabase credentials
uvicorn app.main:app --reload
```

**Full setup**: [docs/SERVER_SETUP.md](docs/SERVER_SETUP.md)

### CLI (Development)

```bash
cd cli
go mod tidy
go build -o ely ./cmd
./ely --help
```

### Clothing Shop (Example API)

```bash
cd ../clothing_shop
python -m venv env
source env/bin/activate
pip install -r requirements.txt
python app.py
# API runs on http://localhost:5000
```

Generate API key:

```bash
curl -X POST http://localhost:5000/api/auth/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-key"}'
```

## Development

### Run Tests

```bash
# Server tests
cd server
pytest tests/ -v

# CLI tests
cd cli
go test ./... -v

# End-to-end
./scripts/e2e-test.sh
```

### Build Distribution

```bash
# Build all platforms
./scripts/build-all.sh

# Creates:
# - ely-linux-amd64
# - ely-linux-arm64
# - ely-darwin-amd64
# - ely-darwin-arm64
# - ely-windows-amd64.exe
```

### Create Release

```bash
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions builds and uploads binaries
```

## Roadmap

- [x] Emblem specification and schema
- [x] Registry backend (Supabase + FastAPI)
- [x] Go CLI core commands
- [x] Authentication (OAuth + API keys)
- [x] Dynamic emblem execution
- [x] Local validation and testing
- [ ] Web UI for browsing emblems
- [ ] Emblem marketplace
- [ ] Private namespaces
- [ ] Team collaboration
- [ ] API monitoring integration
- [ ] Code generation (SDKs)

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Make changes
4. Run tests (`go test ./...` and `pytest tests/`)
5. Commit (`git commit -m 'Add amazing feature'`)
6. Push (`git push origin feature/amazing-feature`)
7. Open Pull Request

## Code Style

### Go
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Add comments for exported functions
- Use meaningful variable names

### Python
- Use `black` for formatting
- Use `flake8` for linting
- Use type hints
- Keep functions under 50 lines

## License

MIT License - see [LICENSE](LICENSE)

## Support

- 📚 Documentation: [docs/](docs/)
- 🐛 Issues: [GitHub Issues](https://github.com/Lo10Th/Elysium/issues)

---

Built with ❤️ for the API economy