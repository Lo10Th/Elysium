# Security Policy

## Supported Versions

Only the latest release receives security fixes.

| Version | Supported          |
|---------|--------------------|
| 0.2.x   | ✅ Yes             |
| < 0.2   | ❌ No              |

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report security issues privately by opening a [GitHub Security Advisory](https://github.com/Lo10Th/Elysium/security/advisories/new) or emailing the maintainers directly (see the GitHub profile).

Please include:

- A description of the vulnerability and its potential impact.
- Steps to reproduce (proof-of-concept code or commands, if available).
- Any suggested fix or mitigation.

### Response Timeline

| Stage                           | Target Time         |
|---------------------------------|---------------------|
| Acknowledgement                 | Within 48 hours     |
| Triage and severity assessment  | Within 3 days       |
| Fix or mitigation plan (critical/high) | Within 7 days |
| Fix or mitigation plan (medium/low)    | Within 30 days |
| Public advisory after patch     | After users have had reasonable time to update |

### Responsible Disclosure

We follow a coordinated disclosure process:

1. Reporter submits vulnerability privately.
2. Maintainers acknowledge within 48 hours.
3. Maintainers investigate and develop a fix.
4. Fix is released; reporter is credited (unless they prefer to remain anonymous).
5. A public advisory is published after users have had reasonable time to update.

Thank you for helping keep Elysium secure.

---

## Security Model

### What Elysium Protects

| Asset                | How it is protected                                                |
|----------------------|--------------------------------------------------------------------|
| Registry auth token  | Stored in `~/.elysium/config.yaml` with `0600` permissions        |
| Third-party API keys | **Never stored** — read from environment variables at runtime      |
| Emblem content       | Served over HTTPS; registry requires auth for writes               |
| User passwords       | Hashed by Supabase Auth (bcrypt) before storage                    |

### What Elysium Does NOT Protect

- Elysium does not sandbox emblem execution. An emblem can instruct the CLI to call any HTTPS endpoint. Only pull emblems from authors you trust.
- The CLI does not verify the TLS certificate chain beyond Go's standard `net/http` defaults.

---

## 1. Authentication Best Practices

### Token Storage

After `ely login`, a JWT access token is written to:

```
~/.elysium/config.yaml
```

This file is created with permissions `0600` (owner read/write only). It is **not** stored in the OS keyring; all tokens are stored as plain text in the config file, so protect this file accordingly.

**Do not** commit `~/.elysium/config.yaml` to version control.

### Token Refresh

Supabase issues short-lived access tokens alongside a refresh token. The CLI automatically uses the refresh token to obtain a new access token when the current one is close to expiry. If automatic refresh fails, re-run `ely login` to obtain fresh tokens.

### Token Revocation

To revoke the locally stored token:

```bash
ely logout
```

This removes the token from `~/.elysium/config.yaml`. To revoke a token from the server side, use the Supabase dashboard or the Supabase Auth admin API to sign out all sessions for the user.

### API Key Management

Third-party API keys referenced by emblems are **never** stored by Elysium. They are read from environment variables at execution time:

```yaml
auth:
  type: api_key
  keyEnv: STRIPE_API_KEY   # CLI reads os.Getenv("STRIPE_API_KEY")
  header: Authorization
```

Best practices for API keys:

- Rotate keys regularly and revoke unused keys.
- Use the least-privilege key for the task (read-only if only reads are needed).
- Store keys in a secrets manager (e.g. 1Password, AWS Secrets Manager, Vault) and inject them into the environment at runtime.
- Never pass keys via command-line flags or embed them in emblem YAML files.

---

## 2. Input Validation

### Emblem Name Validation

Emblem names must match the pattern `^[a-z0-9-]+$` (lowercase letters, digits, and hyphens only). This is enforced at both upload time by the registry and at local validation time by `ely validate`.

```bash
# Valid names
clothing-shop
my-api-v2

# Invalid names (rejected)
MyAPI          # uppercase letters
my_api         # underscores
../evil        # path traversal
```

### URL Validation

The emblem `baseUrl` field must begin with `http://` or `https://`. Before making any HTTP request on behalf of an emblem, the CLI executor validates that the target URL:

- Uses the `http` or `https` scheme (no `file://`, `ftp://`, etc.).
- Is formed from the `baseUrl` declared in the emblem plus the action's relative path.

This prevents Server-Side Request Forgery (SSRF) via non-HTTP protocols.

### Parameter Sanitization

Path parameters are URL-encoded with `url.PathEscape` before being interpolated into the request URL. Query parameters and body fields are passed through the HTTP client without shell expansion, preventing injection attacks.

```go
// Path parameters are percent-encoded
encoded := url.PathEscape(fmt.Sprintf("%v", value))
result = strings.ReplaceAll(result, placeholder, encoded)
```

All incoming registry data is validated by Pydantic models before reaching the database layer. The emblem YAML content is validated against `schemas/emblem.schema.json` on the server before storage.

### Rate Limiting

The registry server uses [SlowAPI](https://github.com/laurentS/slowapi) for per-IP rate limiting. The following limits apply:

| Endpoint category     | Default limit  |
|-----------------------|----------------|
| Public endpoints      | 60 / minute    |
| Authentication        | 30 / minute    |
| Strict (write ops)    | 10 / minute    |
| Registration          | 5 / minute     |
| Token refresh         | 20 / minute    |

Exceeded limits return `429 Too Many Requests`.

For production deployments, layer additional rate limiting at the reverse proxy (nginx, Caddy) or CDN/edge (Vercel Edge) level.

---

## 3. Data Protection

### Credential Handling

- Registry auth tokens are stored in `~/.elysium/config.yaml` with `0600` permissions.
- Third-party API keys are **never** written to disk by the CLI. They exist only in process memory during execution.
- Server-side, all credentials pass through Supabase Auth; the registry server never handles raw passwords.

### Sensitive Data in Logs

- API keys and tokens are **never** logged by the CLI or the registry server.
- Server-side services use `logger.error()` for internal errors and raise sanitised `HTTPException` responses — raw exception messages are never forwarded to the client.
- Do not enable `DEBUG=true` in production, as this can expose internal stack traces.

### Environment Variables

Sensitive configuration for the registry server must be supplied via environment variables, never hardcoded:

| Variable                    | Purpose                                                       |
|-----------------------------|---------------------------------------------------------------|
| `SUPABASE_URL`              | Supabase project URL                                          |
| `SUPABASE_ANON_KEY`         | Public anon key (also accepted as `SUPABASE_KEY`)             |
| `SUPABASE_SERVICE_ROLE_KEY` | Service role key (admin access — keep secret; also accepted as `SUPABASE_SERVICE_KEY`) |
| `SECRET_KEY`                | Application secret for signing                                |
| `ALLOWED_ORIGINS`           | CORS allowlist (comma-separated)                              |

Store these in a `.env` file locally (never commit it) or in your deployment platform's secret store.

### Database Security

- Supabase Row Level Security (RLS) policies enforce that users can only modify emblems they own; read access to emblems is public.
- All queries to the database use parameterised statements via the Supabase Python client, preventing SQL injection.
- The service role key is used only in server-side code and is never exposed to the CLI or browser clients.

---

## 4. Security Headers

### Headers Added by the Registry Server

The `SecurityHeadersMiddleware` in `server/app/main.py` adds the following headers to every response:

| Header                      | Value                                        | Purpose                                              |
|-----------------------------|----------------------------------------------|------------------------------------------------------|
| `X-Content-Type-Options`    | `nosniff`                                    | Prevents MIME-type sniffing                          |
| `X-Frame-Options`           | `DENY`                                       | Prevents clickjacking via iframes                    |
| `X-XSS-Protection`          | `1; mode=block`                              | Legacy XSS filter for older browsers                 |
| `Referrer-Policy`           | `strict-origin-when-cross-origin`            | Limits referrer information leakage                  |
| `Permissions-Policy`        | `geolocation=(), microphone=(), camera=()`   | Disables unused browser features                     |

### CORS Configuration

The `CORS_ORIGINS` environment variable (mapped to `settings.CORS_ORIGINS`) controls the allowed origins. In production, **always** set this to specific domains:

```env
ALLOWED_ORIGINS=https://yourdomain.com
```

Avoid `*` (allow-all) in production. The server accepts credentials (`allow_credentials=True`), which is incompatible with wildcard origins per the CORS specification.

### Rate Limiting Headers

When a rate limit is exceeded the server returns:

```
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{"error": "Rate limit exceeded", "detail": "..."}
```

### Security Headers Checklist (for Production Deployments)

Use this checklist when deploying behind a reverse proxy or CDN:

- [ ] `Strict-Transport-Security` header enabled (HSTS) with `max-age` ≥ 1 year
- [ ] `Content-Security-Policy` header configured for your frontend
- [ ] `X-Content-Type-Options: nosniff` ✅ (set by middleware)
- [ ] `X-Frame-Options: DENY` ✅ (set by middleware)
- [ ] `Referrer-Policy` ✅ (set by middleware)
- [ ] `Permissions-Policy` ✅ (set by middleware)
- [ ] TLS 1.2+ enforced; TLS 1.0/1.1 disabled
- [ ] Weak cipher suites disabled
- [ ] `Server` and `X-Powered-By` headers suppressed at reverse proxy
- [ ] CORS allowlist restricted to known origins

---

## 5. Dependency Security

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

Enable [Dependabot](https://docs.github.com/en/code-security/dependabot) alerts in repository settings to receive automatic notifications of vulnerable dependencies.

---

## 6. Request Safety

### Timeouts

All outbound HTTP requests made by the executor have a **30-second default timeout** to prevent hanging processes.

### Response Size Limit

Executor responses are capped at **10 MB**. Responses exceeding this limit are rejected with an error.

### Redirect Policy

The executor follows up to **5 redirects**. Redirect chains exceeding this limit are rejected.
