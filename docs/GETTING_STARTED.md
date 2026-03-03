# Getting Started

Welcome to **Elysium** - the API app store. This guide will help you get up and running quickly.

## What is Elysium?

Elysium is a registry and CLI tool for discovering and using APIs. Think of it as "npm for APIs" or "Homebrew for web services."

**Core Concepts:**

- **Emblem**: A YAML file that describes an API's endpoints, parameters, authentication, and types
- **ely**: The CLI tool used to pull and execute emblems
- **Registry**: The server that stores and serves emblem definitions

## Installation

### Using the install script (recommended)

```bash
curl -sSL https://get.elysium.dev | bash
```

### Using Homebrew (macOS/Linux)

```bash
brew tap elysium/tap
brew install ely
```

### From source (Go install)

```bash
go install github.com/elysium/ely/cmd/ely@latest
```

### Download binary directly

Download the latest release from [GitHub Releases](https://github.com/elysium/elysium/releases):

```bash
# Linux
curl -L -o ely https://github.com/elysium/elysium/releases/latest/download/ely-linux-amd64
chmod +x ely
sudo mv ely /usr/local/bin/

# macOS
curl -L -o ely https://github.com/elysium/elysium/releases/latest/download/ely-darwin-arm64
chmod +x ely
sudo mv ely /usr/local/bin/

# Windows (run in PowerShell as Administrator)
# Download from https://github.com/elysium/elysium/releases/latest
```

## Quick Start

### 1. Install

```bash
ely --version
```

### 2. Authenticate

```bash
ely login
```

This will open your browser to authenticate with the Elysium registry. Once authenticated, your credentials are stored securely in `~/.elysium/`.

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
ely pull clothing-shop@^1.0.0
```

### View installed emblems

```bash
ely list
```

### Update to latest version

```bash
ely update clothing-shop
```

### Remove an emblem

```bash
ely remove clothing-shop
```

## Publishing Your Own Emblem

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

### 4. Publish

```bash
ely publish ./my-api/
```

## Configuration

Elysium stores configuration in `~/.elysium/config.json`:

```json
{
  "registry": "https://registry.elysium.dev",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "cache_dir": "/home/user/.elysium/cache",
  "installed": {
    "clothing-shop": "1.0.0",
    "stripe": "2.1.3"
  }
}
```

### Environment Variables

- `ELYSIUM_REGISTRY` - Override registry URL
- `ELYSIUM_CACHE_DIR` - Override cache directory
- `NO_COLOR` - Disable colored output
- `ELYSIUM_DEBUG` - Enable debug logging

## Advanced Usage

### Run without installing

You can run emblems directly from the registry without installing:

```bash
ely --no-install clothing-shop list-products
```

### Use with CI/CD

```bash
# In CI pipeline
ely pull clothing-shop@1.0.0
ely clothing-shop create-order \
  --customer-name "CI Test" \
  --customer-email "ci@test.com" \
  --customer-address "Test Address" \
  --items '[{"product_id": 1, "quantity": 1}]'
```

### JSON output for scripting

```bash
ely clothing-shop list-products --output json | jq '.[0].name'
```

### Quiet mode

```bash
ely clothing-shop list-products --quiet
```

### Verbose mode

```bash
ely clothing-shop list-products --verbose
```

## Shell Completion

Enable shell completion for bash, zsh, or fish:

### Bash

```bash
source <(ely completion bash)
# Add to ~/.bashrc
echo 'source <(ely completion bash)' >> ~/.bashrc
```

### Zsh

```bash
source <(ely completion zsh)
# Add to ~/.zshrc
echo 'source <(ely completion zsh)' >> ~/.zshrc
```

### Fish

```bash
ely completion fish > ~/.config/fish/completions/ely.fish
```

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

Or set token manually:

```bash
export ELYSIUM_TOKEN=your-token
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

### Clear cache

```bash
ely cache clean
```

## Next Steps

- Read the [Emblem Specification](./EMBLEM_SPEC.md) to create your own emblems
- Check out [example emblems](../examples/) for inspiration
- Join the community on [Discord](https://discord.gg/elysium)
- Report issues on [GitHub](https://github.com/elysium/elysium/issues)