# Troubleshooting

This guide covers the most common problems encountered when installing, configuring, or using Elysium.

---

## Table of Contents

1. [Authentication Issues](#1-authentication-issues)
2. [Emblem Issues](#2-emblem-issues)
3. [Network Issues](#3-network-issues)
4. [Installation Issues](#4-installation-issues)
5. [Server / Registry Issues](#5-server--registry-issues)
6. [Go CLI Build Issues](#6-go-cli-build-issues)
7. [Python Server Issues](#7-python-server-issues)
8. [Error Message Reference](#8-error-message-reference)
9. [Getting More Help](#9-getting-more-help)

---

## 1. Authentication Issues

### Issue: `invalid credentials`

#### Symptoms
- `ely login` prompts for email and password, then prints `invalid credentials`
- You are not able to complete the login flow

#### Cause
- Wrong email or password entered
- The account does not exist in the registry

#### Solution
1. Double-check your email address and password for typos.
2. If you have not registered yet, answer `y` when `ely login` asks to create a new account.
3. If you forgot your password, reset it through the Supabase dashboard or the registry web UI.

#### Prevention
- Use a password manager to avoid transcription errors.
- Confirm your registration email before attempting to log in.

---

### Issue: `not logged in` / no stored token

#### Symptoms
- `ely whoami` prints `not logged in`
- Every command that requires authentication fails immediately

#### Cause
- No valid token is stored in `~/.elysium/config.yaml`
- A previous `ely logout` cleared the token without a subsequent login

#### Solution
```bash
ely login
```

If you were previously logged in and the token expired:

```bash
ely logout   # Clear the stale token
ely login    # Re-authenticate
```

#### Prevention
- Tokens expire after a set period. Re-authenticate before starting a long session.

---

### Issue: `API returned status 401` with a stored token

#### Symptoms
- You are logged in (`ely whoami` shows your account) but commands fail with HTTP 401
- The error reads `authentication required` or `API returned status 401`

#### Cause
- The stored JWT token has expired
- The registry URL in `~/.elysium/config.yaml` changed and the token no longer matches

#### Solution
```bash
# Verify the registry URL
ely config get registry

# Re-authenticate
ely logout && ely login
```

#### Prevention
- Run `ely logout && ely login` at the start of each day if you work with long-lived sessions.

---

### Issue: `permission denied` / `API returned status 403`

#### Symptoms
- The CLI reports `permission denied` or `Insufficient permissions for this resource`
- The action completes the HTTP call but is rejected by the target API

#### Cause
- The API key stored in the environment variable does not have the required scope/role
- The key was revoked on the target service

#### Solution
1. Check which environment variable the emblem uses:
   ```bash
   ely info <emblem>
   ```
2. Verify the key has the correct permissions on the target API's dashboard.
3. Regenerate the key if it was revoked, then update the environment variable.

#### Prevention
- Use keys with the minimum required permissions (principle of least privilege).
- Rotate keys periodically and update environment variables accordingly.

---

## 2. Emblem Issues

### Issue: `emblem 'X' not found`

#### Symptoms
- `ely pull <name>` or `ely <name> <action>` prints `emblem '<name>' not found`
- `ely info <name>` returns nothing

#### Cause
- The emblem name is misspelled
- The emblem has not been published to the registry yet
- The CLI is pointing at the wrong registry

#### Solution
```bash
# Search for the correct name
ely search <partial-name>

# Verify the registry URL
ely config get registry

# Point at the production registry if needed
ely config set registry https://ely.karlharrenga.com
```

#### Prevention
- Use `ely search` to discover emblem names before pulling.
- Keep `registry` in `~/.elysium/config.yaml` pointed at the correct server.

---

### Issue: `invalid YAML` / YAML parse error

#### Symptoms
- `ely validate ./emblem.yaml` or `ely pull` fails with a YAML parse error
- The error message includes a line number and column pointing to the syntax problem

#### Cause
- Tabs used instead of spaces for indentation (YAML requires spaces)
- Missing or extra colons, quotes, or indentation levels
- Special characters (`:`, `#`, `{`, `}`) in a string value not properly quoted

#### Solution
1. Open the file in a YAML-aware editor (VS Code, IntelliJ) that highlights syntax errors.
2. Validate online at [yamllint.com](https://www.yamllint.com/).
3. Run the built-in validator:
   ```bash
   ely validate ./emblem.yaml
   ```

#### Prevention
- Enable YAML linting in your editor.
- Use the `ely init <name>` scaffold as a starting point to avoid structural mistakes.

---

### Issue: `unknown action` / action not found

#### Symptoms
- `ely <emblem> <action>` prints `action not found` or `unknown action`
- `ely info <emblem>` does not list the action you expect

#### Cause
- The action name is misspelled (action names are case-sensitive and kebab-case)
- The locally cached emblem is an older version that does not include that action

#### Solution
```bash
# List all available actions for the emblem
ely info <emblem>

# Re-pull to get the latest version
ely pull <emblem>

# Try the action again
ely <emblem> <action>
```

#### Prevention
- Always `ely pull <emblem>` after the registry announces a new version.
- Use `ely outdated` to check which cached emblems have newer versions available.

---

### Issue: Missing required parameter / parameter validation error

#### Symptoms
- The CLI exits with an error such as `required flag "--<param>" not set` or `parameter '<name>' is required`
- The action does not execute at all

#### Cause
- A required parameter defined in the emblem's action was not supplied on the command line
- The parameter name was misspelled (flags are case-sensitive)

#### Solution
```bash
# Check the required parameters for an action
ely info <emblem>

# Supply all required flags
ely <emblem> <action> --<param1> value1 --<param2> value2

# Example
ely clothing-shop get-product --id 42
```

#### Prevention
- Run `ely <emblem> <action> --help` before executing to see all required and optional flags.

---

### Issue: `response too large` (exceeds 10 MB limit)

#### Symptoms
- The action appears to complete but the CLI prints `response too large: N bytes (limit 10485760 bytes)`
- No output data is displayed

#### Cause
- The target API returned a payload larger than the 10 MB safety limit enforced by the executor

#### Solution
1. Use server-side pagination parameters to request a smaller data set:
   ```bash
   ely <emblem> list-items --limit 100 --page 1
   ```
2. Contact the emblem author to add pagination support if it is missing.

#### Prevention
- Always use pagination when listing large collections.
- Prefer action parameters that filter results (date range, status, category) to reduce response size.

---

### Issue: Pulled emblem fails schema validation

#### Symptoms
- The emblem downloads successfully but `ely validate` or execution fails with a schema mismatch error
- Error message references a field name or type that is unexpected

#### Cause
- The emblem in the registry was published with an older schema version that is no longer compatible with your CLI version

#### Solution
1. Check your CLI version and the registry schema version:
   ```bash
   ely --version
   ```
2. Contact the emblem author to republish with the current schema.
3. As a temporary workaround, manually edit `~/.elysium/cache/<name>/emblem.yaml` to conform to the current schema.

#### Prevention
- Keep the CLI up to date: `ely self-update` or re-run the install script.
- Pin a specific emblem version if you need stability: `ely pull <name>@<version>`.

---

## 3. Network Issues

### Issue: `connection refused`

#### Symptoms
- Any command that contacts the registry or a target API fails with `connection refused`
- The error includes a URL such as `http://localhost:8000` or `https://api.example.com`

#### Cause
- The target server is not running (common for local development setups)
- The port in the URL is wrong
- A firewall rule is blocking the connection

#### Solution
```bash
# Check the registry URL
ely config get registry

# If pointing at localhost, start the local server
cd server
uvicorn app.main:app --reload --port 8000

# If the registry should be the production server
ely config set registry https://ely.karlharrenga.com
```

#### Prevention
- Always start the local server before running CLI commands against it.
- Double-check the port number in the registry URL matches the running server.

---

### Issue: Request timed out (30-second limit)

#### Symptoms
- A command hangs for ~30 seconds, then prints `Request timed out`
- The error message shows `Timeout: 30s`

#### Cause
- The target API server is overloaded, unreachable, or very slow
- A network proxy is intercepting and delaying the request

#### Solution
1. Check the target API's health status page or dashboard.
2. Retry after a short wait:
   ```bash
   ely <emblem> <action>
   ```
3. Test direct connectivity:
   ```bash
   curl -v <baseUrl>
   ```

#### Prevention
- Avoid running large batch operations during known peak hours.
- Contact the emblem author if the API consistently takes longer than 30 seconds for normal requests.

---

### Issue: SSL/TLS certificate validation errors

#### Symptoms
- The CLI fails with an error such as `x509: certificate signed by unknown authority`, `TLS handshake timeout`, or `certificate has expired`
- HTTPS connections to the registry or target API fail

#### Cause
- The server's TLS certificate is expired, self-signed, or issued by a CA not trusted by your system
- Your system clock is significantly wrong (causes valid certificates to appear expired)
- A corporate proxy is performing TLS inspection with its own CA

#### Solution

**Expired certificate:**
```bash
# Verify certificate validity dates (ensure basic connectivity first)
openssl s_client -connect <hostname>:443 -servername <hostname> | openssl x509 -noout -dates

# If the server is under your control, renew the certificate (e.g. via Let's Encrypt)
certbot renew
```

**Untrusted CA (`x509: certificate signed by unknown authority`):**
```bash
# Add the server's CA certificate to the system trust store
# Linux (Debian/Ubuntu)
sudo cp server-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# macOS
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain server-ca.crt
```

**System clock issue:**
```bash
# Check and synchronise the system clock
date
sudo ntpdate pool.ntp.org   # Linux
# macOS: System Preferences → Date & Time → Set automatically
```

**Corporate proxy with custom CA:**
```bash
# Add your organisation's root CA to the system trust store
# Linux (Debian/Ubuntu)
sudo cp your-corp-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# macOS
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain your-corp-ca.crt
```

#### Prevention
- Automate TLS certificate renewal (e.g. with `certbot --deploy-hook`).
- Keep your system clock synchronised with NTP.
- Document any custom CA requirements in your team's onboarding guide.

---

### Issue: `rate limit exceeded` / `API returned status 429`

#### Symptoms
- A command fails with `rate limit exceeded` and `Retry after: N seconds`
- Multiple rapid requests are rejected

#### Cause
- Too many requests were sent in a short period, exceeding the server's rate limit

#### Solution
```bash
# Wait the indicated number of seconds before retrying
sleep <retry_after>
ely <emblem> <action>
```

#### Prevention
- Add delays between repeated calls in automation scripts.
- Upgrade to a higher-tier plan if your workflow legitimately requires more requests per minute.

---

## 4. Installation Issues

### Issue: `ely: command not found`

#### Symptoms
- Typing `ely` in a new terminal session returns `command not found`
- `which ely` produces no output

#### Cause
- The install script placed the binary in `/usr/local/bin` but that path is not in your `PATH`
- The binary was installed in a user-local directory (`~/.local/bin`) that is not in `PATH`

#### Solution
```bash
# Locate the binary
which ely || ls /usr/local/bin/ely 2>/dev/null || ls ~/.local/bin/ely 2>/dev/null

# Temporarily add the directory to PATH
export PATH="$PATH:/usr/local/bin"

# Verify
ely --version

# Make the change permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.bashrc
source ~/.bashrc
```

#### Prevention
- After installation, always open a **new** terminal session so that shell profile changes take effect.
- Choose an install directory that is already on your `PATH`.

---

### Issue: `permission denied` during installation

#### Symptoms
- The install script exits with `permission denied` when trying to copy the binary to `/usr/local/bin`
- `go install` fails with a write-permission error on the `$GOPATH/bin` directory

#### Cause
- Your current user does not have write access to the target directory

#### Solution
```bash
# Option 1: Run the install script with elevated privileges
curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | sudo bash

# Option 2: Install to a user-writable directory
curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash -s -- --install-dir ~/.local/bin
export PATH="$PATH:$HOME/.local/bin"
```

#### Prevention
- Prefer user-local installation (`~/.local/bin`) to avoid requiring `sudo`.
- Set `GOPATH` to a directory your user owns when using `go install`.

---

### Issue: `go install` fails with Go version errors

#### Symptoms
- `go install` fails with a message like `note: module requires Go 1.21`
- `go build` prints version-related errors

#### Cause
- Your installed Go version is older than the minimum required version (Go 1.21+)

#### Solution
```bash
# Check the installed Go version
go version

# If below 1.21, install the latest release from https://go.dev/dl/
# Then retry
go install github.com/Lo10Th/Elysium/cli/cmd/ely@latest
```

#### Prevention
- Keep Go updated. Use a version manager such as [asdf](https://asdf-vm.com/) or [g](https://github.com/stefanmaric/g) to manage multiple Go versions.

---

## 5. Server / Registry Issues

### Issue: Server returns `500 Internal Server Error` on all endpoints

#### Symptoms
- Every API request to the registry returns HTTP 500
- The server log shows database connection errors

#### Cause
- Required Supabase environment variables are missing or incorrect in `server/.env`

#### Solution
```bash
# Check the .env file
cat server/.env

# Required variables:
# SUPABASE_URL=https://<project-ref>.supabase.co
# SUPABASE_ANON_KEY=<anon-key>
# SUPABASE_SERVICE_ROLE_KEY=<service-role-key>
# SECRET_KEY=<random-secret>
```

Refer to [SERVER_SETUP.md](SERVER_SETUP.md) for the full environment variable reference.

#### Prevention
- Use `.env.example` as a template and fill in all values before first launch.
- Add a startup health-check that validates required environment variables.

---

### Issue: `POST /api/auth/login` returns 401 with correct credentials

#### Symptoms
- The server accepts the request but returns 401 even though the credentials are correct
- Newly registered users cannot log in

#### Cause
- Supabase requires email confirmation and the user has not yet confirmed their email address

#### Solution
1. Check the registration email for a confirmation link and click it.
2. For development environments, disable email confirmation in the Supabase dashboard:
   **Authentication → Providers → Email → Confirm email** → toggle off.

#### Prevention
- Document the email confirmation requirement in onboarding materials.
- Use auto-confirm mode in local/staging environments to simplify development.

---

### Issue: Registry search returns no results

#### Symptoms
- `ely search <query>` returns an empty list for queries that should match emblems
- The registry server is reachable (no connection error)

#### Cause
- The `emblems` table in the database is empty
- The search index has not been populated

#### Solution
```bash
# Insert a test emblem via the Supabase dashboard:
# 1. Go to https://supabase.com/dashboard
# 2. Navigate to Table Editor → emblems
# 3. Insert a row with name, version, description, and content fields
```

#### Prevention
- Seed the database with sample emblems during development setup.
- Verify the emblems table has data before testing search functionality.

---

## 6. Go CLI Build Issues

### Issue: `go build` fails with `missing go.sum entry`

#### Symptoms
- `go build` or `go test` fails with `missing go.sum entry for module ...`

#### Cause
- Dependencies were added or changed without updating `go.sum`

#### Solution
```bash
cd cli
go mod tidy
go build -o ely ./cmd
```

#### Prevention
- Run `go mod tidy` after every dependency change and commit the updated `go.sum`.

---

### Issue: `go test ./...` fails with import errors

#### Symptoms
- Tests fail immediately with `cannot find package` errors
- Import paths are unresolved

#### Cause
- Module cache is incomplete or dependencies were not downloaded

#### Solution
```bash
cd cli
go mod download
go test ./...
```

#### Prevention
- Run `go mod download` as part of CI setup steps before running tests.

---

## 7. Python Server Issues

### Issue: `ModuleNotFoundError` when starting the server

#### Symptoms
- `uvicorn app.main:app` exits with `ModuleNotFoundError: No module named 'fastapi'` (or similar)

#### Cause
- Python dependencies are not installed, or the virtual environment is not activated

#### Solution
```bash
cd server
python -m venv venv
source venv/bin/activate   # Windows: venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --reload
```

#### Prevention
- Always activate the virtual environment before running server commands.
- Add the `venv/` directory to `.gitignore` so it is not accidentally committed.

---

### Issue: `pytest` cannot find the `app` module

#### Symptoms
- Running `pytest` from the repo root fails with `ModuleNotFoundError: No module named 'app'`

#### Cause
- pytest is run from a directory where `app/` is not importable (typically the repo root instead of `server/`)

#### Solution
```bash
# Always run pytest from inside the server/ directory
cd server
pytest tests/ -v
```

#### Prevention
- Add a `Makefile` target or script that changes into `server/` before running tests to prevent this mistake.

---

### Issue: `black` or `isort` not found

#### Symptoms
- Pre-commit hooks or CI steps fail with `black: command not found` or `isort: command not found`

#### Cause
- Development tools are not installed in the active virtual environment

#### Solution
```bash
pip install black isort ruff
```

#### Prevention
- Add `black`, `isort`, and `ruff` to `requirements-dev.txt` and install with `pip install -r requirements-dev.txt`.

---

## 8. Error Message Reference

This section maps the exact error strings produced by the CLI to their causes and resolution steps.

| Error Message | HTTP Status | Cause | Resolution |
|---|---|---|---|
| `authentication required` / `Missing API key: <ENV>` | 401 | The environment variable for the emblem's API key is not set | `export <ENV>=your-key-here` |
| `permission denied` / `Insufficient permissions for <resource>` | 403 | The API key does not have the required scope | Check key permissions on the target API dashboard |
| `emblem '<name>' not found` | — | Emblem name is wrong or not published | `ely search <name>` to find the correct name |
| `resource not found` | 404 | The requested resource does not exist on the target API | Verify the resource ID or URL path in the emblem |
| `rate limit exceeded` / `Too many requests` | 429 | Request rate exceeded the server's limit | Wait `Retry after: N seconds` before retrying |
| `API returned status 500: API server error` | 500 | The registry or target API has an internal error | Try again later; check the service status page |
| `Request timed out` / `Timeout: 30s` | — | The target API did not respond within 30 seconds | Check API health; use pagination to reduce payload size |
| `connection refused` | — | The target server is not running or unreachable | Start the server or fix the URL in `ely config` |
| `network connectivity issue` | — | General network failure (DNS, routing) | Check your internet connection; try `ping <host>` |
| `configuration not found` | — | `~/.elysium/config.yaml` does not exist | Run `ely login` to create it |
| `invalid URL scheme: only http and https are allowed` | — | The emblem's `baseUrl` uses a non-HTTP scheme | Ensure `baseUrl` starts with `http://` or `https://` |
| `response too large: N bytes (limit 10485760 bytes)` | — | API response exceeds the 10 MB safety limit | Use pagination or filtering parameters |
| `action not found` | — | The specified action does not exist in the emblem | Run `ely info <emblem>` to list valid actions |
| `unsupported HTTP method: <METHOD>` | — | The emblem defines an HTTP method the executor does not support | Only `GET`, `POST`, `PUT`, `DELETE`, `PATCH` are supported |
| `ModuleNotFoundError` | — | Python venv not activated or dependencies not installed | Run `pip install -r requirements.txt` inside an active venv |
| `x509: certificate signed by unknown authority` | — | TLS certificate is issued by a CA not trusted by your system | Install the server's CA certificate in your system trust store |
| `certificate has expired` | — | The server's TLS certificate has passed its expiry date | Renew the certificate (e.g. `certbot renew`) on the server |

---

## 9. Getting More Help

If your issue is not listed here:

1. Search [GitHub Issues](https://github.com/Lo10Th/Elysium/issues) — it may already be reported.
2. Run commands with the `--verbose` or `-v` flag to get more diagnostic output:
   ```bash
   ely pull <emblem> --verbose
   ely <emblem> <action> --verbose
   ```
3. Check the server logs for detailed error information:
   ```bash
   # If running locally
   uvicorn app.main:app --reload --log-level debug
   ```
4. Open a new [GitHub Issue](https://github.com/Lo10Th/Elysium/issues/new) with:
   - The exact command you ran
   - The full error output (including any `Reason:` and `Suggestion:` lines)
   - Your OS, `ely --version`, and Go/Python version if relevant
