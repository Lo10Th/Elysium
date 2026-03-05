# Security

> **The canonical security policy lives at [`/SECURITY.md`](../SECURITY.md)** at the repository root.  
> GitHub displays that file automatically on the repository's Security tab.
> The content below is kept as a convenience reference for readers browsing the `docs/` directory.

---

# Security Policy

## Supported Versions

Only the latest release receives security fixes.

| Version | Supported |
|---------|-----------|
| 0.2.x   | ✅ Yes    |
| < 0.2   | ❌ No     |

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report security issues privately by opening a [GitHub Security Advisory](https://github.com/Lo10Th/Elysium/security/advisories/new) or emailing the maintainers directly (see the GitHub profile).

Please include:
- A description of the vulnerability and its potential impact.
- Steps to reproduce (proof-of-concept code or commands, if available).
- Any suggested fix or mitigation.

You can expect an acknowledgement within **48 hours** and a fix or mitigation plan within **7 days** for critical issues.

## Security Model

### What Elysium protects

| Asset | How it is protected |
|---|---|
| Registry auth token | Stored in `~/.elysium/config.yaml` with `0600` permissions |
| Third-party API keys | **Never stored** — read from environment variables at runtime |
| Emblem content | Served over HTTPS; registry requires auth for writes |
| User passwords | Hashed by Supabase Auth (bcrypt) before storage |

### What Elysium does NOT protect

- Elysium does not sandbox emblem execution. An emblem can instruct the CLI to call any HTTPS endpoint. Only pull emblems from authors you trust.
- The CLI does not verify the TLS certificate chain beyond Go's standard `net/http` defaults.

## CLI Security Details

### Credential Storage

After `ely login`, a JWT access token is written to:

```
~/.elysium/config.yaml
```

This file is created with permissions `0600` (owner read/write only). It is **not** stored in the OS keyring; all tokens are stored as plain text in the config file, so protect this file accordingly.

**Do not** commit `~/.elysium/config.yaml` to version control. The file contains your Elysium auth token.

### Environment Variables for API Keys

Emblem auth is always configured via environment variables:

```yaml
auth:
  type: api_key
  keyEnv: STRIPE_API_KEY   # CLI reads os.Getenv("STRIPE_API_KEY")
  header: Authorization
```

The key is read at execution time with `os.Getenv`. It is never logged, written to disk, or sent to the Elysium registry.

### URL Validation

Before making any HTTP request on behalf of an emblem, the executor validates that the target URL:

- Uses the `http` or `https` scheme (no `file://`, `ftp://`, etc.).
- Matches the `baseUrl` declared in the emblem (or is a relative path from it).

### Request Timeouts

All outbound HTTP requests made by the executor have a 30-second default timeout to prevent hanging processes.

## Registry Server Security Details

### Authentication

The registry uses Supabase Auth to issue JWTs. All write endpoints (`POST`, `PUT`, `DELETE` on `/api/emblems`) require a valid JWT in the `Authorization: Bearer <token>` header.

### Row Level Security

Supabase Row Level Security (RLS) policies enforce that:

- Users can only modify emblems they own.
- Read access to emblems is public.

### Input Validation

All incoming data is validated by Pydantic models before reaching the database layer. The emblem content is validated against `schemas/emblem.schema.json`.

### Rate Limiting

The server should be deployed behind a reverse proxy (nginx, Caddy, Vercel Edge) that enforces rate limiting. The FastAPI app itself does not implement rate limiting.

### CORS

The `ALLOWED_ORIGINS` environment variable controls the CORS allowlist. In production, set this to specific domains only:

```env
ALLOWED_ORIGINS=https://yourdomain.com
```

Avoid using `*` (allow-all) in production.

## Dependency Security

### CLI (Go)

Go dependencies are pinned in `go.sum`. To check for known vulnerabilities:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### Server (Python)

Python dependencies are listed in `server/requirements.txt`. To check for vulnerabilities:

```bash
pip install pip-audit
pip-audit -r server/requirements.txt
```

### Staying Up to Date

Enable [Dependabot](https://docs.github.com/en/code-security/dependabot) alerts in the repository settings to receive automatic notifications of vulnerable dependencies.

## Responsible Disclosure

We follow a coordinated disclosure process:

1. Reporter submits vulnerability privately.
2. Maintainers acknowledge within 48 hours.
3. Maintainers investigate and develop a fix.
4. Fix is released and reporter is credited (unless they prefer to remain anonymous).
5. A public advisory is published after users have had reasonable time to update.

Thank you for helping keep Elysium secure.
