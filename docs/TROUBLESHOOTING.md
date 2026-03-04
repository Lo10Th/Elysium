# Troubleshooting

This guide covers the most common problems encountered when installing, configuring, or using Elysium.

---

## Table of Contents

- [Installation Issues](#installation-issues)
- [Authentication Issues](#authentication-issues)
- [Pulling Emblems](#pulling-emblems)
- [Executing Actions](#executing-actions)
- [Server / Registry Issues](#server--registry-issues)
- [Go CLI Build Issues](#go-cli-build-issues)
- [Python Server Issues](#python-server-issues)
- [Getting More Help](#getting-more-help)

---

## Installation Issues

### `ely: command not found` after install script

**Cause:** The install script placed the binary in `/usr/local/bin` but that directory is not in your `PATH`.

**Fix:**

```bash
# Check where ely was installed
which ely || ls /usr/local/bin/ely

# Verify PATH
echo $PATH

# Add /usr/local/bin to PATH (add to ~/.bashrc or ~/.zshrc to make permanent)
export PATH="$PATH:/usr/local/bin"

# Then try again
ely --version
```

### Permission denied when running install script

**Cause:** The script tries to move the binary to `/usr/local/bin` and your user does not have write access.

**Fix:** Run with `sudo`:

```bash
curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | sudo bash
```

Or install to a user-writable directory:

```bash
curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash -s -- --install-dir ~/.local/bin
export PATH="$PATH:$HOME/.local/bin"
```

### `go install` fails with version errors

**Cause:** Your Go version is too old. Elysium requires Go 1.21+.

**Fix:**

```bash
go version   # Check current version

# Install the latest Go from https://go.dev/dl/
# Then retry:
go install github.com/Lo10Th/Elysium/cli/cmd/ely@latest
```

---

## Authentication Issues

### `ely login` prompts for credentials but returns "invalid credentials"

**Cause:** Wrong email or password, or the account does not exist yet.

**Fix:**

1. Double-check your email and password.
2. If you have not registered, `ely login` will ask if you want to create an account — answer `y`.
3. If you forgot your password, reset it through the Supabase dashboard or the registry web UI.

### `ely whoami` returns "not logged in"

**Cause:** No valid token is stored.

**Fix:**

```bash
ely login
```

If you were previously logged in but the token expired:

```bash
ely logout   # Clear stale token
ely login    # Re-authenticate
```

### Token stored but API returns 401

**Cause:** The token has expired or the registry URL changed.

**Fix:**

```bash
# Check which registry the CLI is pointing at
ely config get registry

# Re-authenticate
ely logout && ely login
```

---

## Pulling Emblems

### `ely pull <name>` returns "emblem not found"

**Cause:** The emblem name is misspelled, or the emblem has not been published to the registry yet.

**Fix:**

```bash
# Search for the correct name
ely search <partial-name>

# Example
ely search stripe
```

### `ely pull <name>` returns "connection refused"

**Cause:** The CLI is pointed at a local registry that is not running.

**Fix:**

```bash
# Check the configured registry URL
ely config get registry

# Point at the production registry
ely config set registry https://ely.karlharrenga.com

# Or start your local server
cd server
uvicorn app.main:app --reload --port 8000
```

### Pulled emblem fails schema validation

**Cause:** The emblem in the registry was published with an older schema version that is no longer compatible.

**Fix:** Contact the emblem author to republish, or manually edit `~/.elysium/cache/<name>/emblem.yaml` as a workaround.

---

## Executing Actions

### `ely <emblem> <action>` returns "API key not set"

**Cause:** The required environment variable for the emblem's authentication is not set.

**Fix:**

```bash
# Check which env var is required
ely info <emblem>

# Set the variable
export STRIPE_API_KEY=sk_test_...
ely stripe list-customers
```

To set it permanently, add the `export` line to `~/.bashrc` or `~/.zshrc`.

### `ely <emblem> <action>` returns "unknown action"

**Cause:** The action name is misspelled, or the emblem you have cached is an older version that does not include that action.

**Fix:**

```bash
# List available actions
ely info <emblem>

# Re-pull to get the latest version
ely pull <emblem>
```

### HTTP request fails with `connection refused` or `no such host`

**Cause:** The `baseUrl` in the emblem points to a local development server that is not running, or the domain does not exist.

**Fix:** Check the emblem's `baseUrl`:

```bash
ely info <emblem>
# Look for the baseUrl field
```

If the emblem is meant for local development (e.g. `http://localhost:5000`), start the target API first.

### Response is empty or garbled

**Cause:** The target API returned an unexpected content type (e.g. HTML error page instead of JSON).

**Fix:** Run with verbose output to see the raw response:

```bash
ely <emblem> <action> --verbose
```

### Action hangs with no output

**Cause:** The target API is not responding within the 30-second timeout.

**Fix:** Check the target API's health, then retry. If the service is slow by design, contact the emblem author to increase the timeout setting in the emblem definition.

---

## Server / Registry Issues

### Server starts but returns 500 on all endpoints

**Cause:** Supabase environment variables are not set or are incorrect.

**Fix:**

```bash
# Check the .env file
cat server/.env

# Required variables:
# SUPABASE_URL
# SUPABASE_ANON_KEY
# SUPABASE_SERVICE_ROLE_KEY
# SECRET_KEY
```

See [SERVER_SETUP.md](SERVER_SETUP.md) for full environment variable reference.

### `POST /api/auth/login` returns 401 even with correct credentials

**Cause:** The user was created in Supabase but email confirmation is still pending.

**Fix:** Check the user's email for a confirmation link, or disable email confirmation in Supabase Auth settings for development.

### Registry search returns no results

**Cause:** The emblems table is empty, or the search index has not been built.

**Fix:**

```bash
# Insert an emblem manually via the Supabase dashboard:
# 1. Open your project at https://supabase.com/dashboard
# 2. Navigate to Table Editor → emblems
# 3. Insert a row with name, version, description, and content fields
```

---

## Go CLI Build Issues

### `go build` fails with `missing go.sum entry`

**Fix:**

```bash
cd cli
go mod tidy
go build -o ely ./cmd
```

### `go test ./...` fails with import errors

**Fix:**

```bash
cd cli
go mod download
go test ./...
```

---

## Python Server Issues

### `ModuleNotFoundError` when running the server

**Fix:**

```bash
cd server
python -m venv venv
source venv/bin/activate   # Windows: venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --reload
```

### `pytest` cannot find the `app` module

**Cause:** Tests are run from the wrong directory.

**Fix:** Always run pytest from inside the `server/` directory:

```bash
cd server
pytest tests/ -v
```

### `black` or `isort` not found

**Fix:**

```bash
pip install black isort ruff
```

---

## Getting More Help

If your issue is not listed here:

1. Search [GitHub Issues](https://github.com/Lo10Th/Elysium/issues) — it may already be reported.
2. Run with `--verbose` or `-v` flag to get more diagnostic output.
3. Check the server logs:
   ```bash
   # If running locally
   uvicorn app.main:app --reload --log-level debug
   ```
4. Open a new [GitHub Issue](https://github.com/Lo10Th/Elysium/issues/new) with:
   - The exact command you ran
   - The full error output
   - Your OS and `ely --version` output
