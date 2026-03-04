# Getting Started

Welcome to **Elysium** - the API app store. This guide will help you get up and running quickly.

## What is Elysium?

Elysium is a registry and CLI tool for discovering and using APIs. Think of it as "npm for APIs" or "Homebrew for web services."

**Core Concepts:**

- **Emblem**: A YAML file that describes an API's endpoints, parameters, authentication, and types
- **ely**: The CLI tool used to pull and execute emblems
- **Registry**: The server that stores and serves emblem definitions

## Installation

### One-line install (Linux/macOS) — Recommended

```bash
curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash
```

Or with wget:

```bash
wget -qO- https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash
```

Install a specific version:

```bash
curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash -s -- --version v0.2.0
```

### Using Go install

```bash
go install github.com/Lo10Th/Elysium/cli/cmd/ely@latest
```

### Download binary directly

Download the latest release from [GitHub Releases](https://github.com/Lo10Th/Elysium/releases):

```bash
# Linux
curl -L -o ely https://github.com/Lo10Th/Elysium/releases/latest/download/ely-linux-amd64
chmod +x ely
sudo mv ely /usr/local/bin/

# macOS
curl -L -o ely https://github.com/Lo10Th/Elysium/releases/latest/download/ely-darwin-arm64
chmod +x ely
sudo mv ely /usr/local/bin/

# Windows (run in PowerShell as Administrator)
# Download from https://github.com/Lo10Th/Elysium/releases/latest
```

## Quick Start

### 1. Verify Installation

```bash
ely --version
```

### 2. Authenticate

```bash
ely login
```

This will open your browser to authenticate with the Elysium registry. Once authenticated, your credentials are stored securely in the system keyring.

### 3. Pull an emblem

```bash
ely pull clothing-shop
```

This downloads the emblem to your local cache (`~/.elysium/cache/`).

### 4. Configure authentication

If the emblem requires API keys, set the environment variable:

```bash
export CLOTHING_SHOP_API_KEY=your-api-key-here
```

> ⚠️ **Security Note**: Never commit API keys to version control. Use environment variables or a secrets manager.

The emblem definition specifies which environment variable to use via the `auth.keyEnv` field.

### 5. Use the emblem

List available actions:

```bash
ely clothing-shop --help
```

Execute an action:

```bash
ely clothing-shop list-products
```

Execute with parameters:

```bash
ely clothing-shop get-product --id 1
```

Create a resource:

```bash
ely clothing-shop create-product \
  --name "Vintage T-Shirt" \
  --price 29.99 \
  --category shirts \
  --size M \
  --color blue
```

## Available Commands

### Authentication

```bash
ely login        # Browser-based OAuth login
ely logout       # Remove stored credentials
ely whoami       # Show current user info
```

### Discovery

```bash
ely search <query>           # Search emblems in registry
ely info <name>              # View emblem details
ely info <name>@<version>    # View specific version
ely list                      # List installed emblems
```

### Installation

```bash
ely pull <name>              # Pull latest version
ely pull <name>@<version>    # Pull specific version
```

### Development

```bash
ely init <name>              # Create new emblem scaffold
ely validate ./emblem.yaml   # Validate emblem YAML
ely test ./<dir>/            # Test emblem locally
```

### API Keys

```bash
ely keys list               # List your API keys
ely keys create <name>      # Create new API key
ely keys create <name> --expires 30  # Key expires in 30 days
ely keys delete <id>        # Delete an API key
ely keys show <id>          # Show key details
```

### Execution

```bash
ely <emblem-name> <action> [flags]

# Examples:
ely clothing-shop list-products
ely clothing-shop get-product --id 1
```

## Configuration

Elysium stores configuration in `~/.elysium/config.yaml`:

```yaml
registry: https://ely.karlharrenga.com
token: eyJhbGciOiJIUzI1NiIs...
cache_dir: ~/.elysium/cache
installed:
  clothing-shop: 1.0.0
```

### Environment Variables

- `ELYSIUM_REGISTRY` - Override registry URL (default: https://ely.karlharrenga.com)

## Example Workflows

### Search for emblems

```bash
ely search "payment"
ely search "database" --limit 10
ely search "ai" --sort downloads
```

### View emblem information

```bash
ely info clothing-shop
ely info clothing-shop@1.0.0
```

### Install specific version

```bash
ely pull clothing-shop@1.2.0
```

### View installed emblems

```bash
ely list
```

## Creating Your Own Emblem

### 1. Initialize a new emblem

```bash
ely init my-api
cd my-api
```

This creates an `emblem.yaml` template.

### 2. Edit the emblem

Edit `emblem.yaml` with your API details (see [EMBLEM_SPEC.md](./EMBLEM_SPEC.md)).

### 3. Validate

```bash
ely validate ./emblem.yaml
```

### 4. Test locally

```bash
ely test ./
```

### 5. Publish (Planned Feature)

```bash
# Coming soon: ely publish ./my-api/
```

> **Note**: The `ely publish` command is not yet implemented. For now, emblems can be added to the registry by submitting them via the API or creating a pull request.

## Running Tests

### Unit Tests

```bash
cd cli
go test ./...
```

### Integration Tests

Integration tests exercise the full emblem execution flow against local mock
HTTP servers. They are kept behind the `integration` build tag so they do not
run during ordinary `go test ./...` passes.

```bash
cd cli
go test -v -tags=integration ./test/
```

The tests cover three scenarios:

| Scenario | Tests |
|----------|-------|
| **Happy path** | `TestFullEmblemFlow*` — cache write → load → execute |
| **Error handling** | `TestErrorHandling_*` — missing emblem, bad YAML, API errors, connection failure |
| **Auth integration** | `TestAuthIntegration_*` — missing/wrong/correct key, no-auth emblem |

No external services or credentials are required; all HTTP calls are made to
`net/http/httptest` servers that spin up and tear down within each test.

## Troubleshooting

### "API key required" error

Set the required environment variable:

```bash
export CLOTHING_SHOP_API_KEY=your-key
```

Check emblem info for required env vars:

```bash
ely info clothing-shop
```

### "Not authenticated" error

Run login:

```bash
ely login
```

### "Emblem not found" error

Check the emblem name:

```bash
ely search clothing-shop
```

Make sure you're connected to the correct registry:

```bash
ely config get registry
```

### "Invalid emblem" error

Validate your emblem:

```bash
ely validate ./emblem.yaml
```

### Connection errors

If you're having trouble connecting to the registry:

1. Check your internet connection
2. Verify the registry URL: `ely config get registry`
3. Try with verbose output: `ely search <query> --verbose`

## Next Steps

- Read the [Emblem Specification](./EMBLEM_SPEC.md) to create your own emblems
- Check out [example emblems](../examples/) for inspiration
- Report issues on [GitHub](https://github.com/Lo10Th/Elysium/issues)