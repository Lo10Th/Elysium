# Changelog

All notable changes to Elysium will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Vercel Emblem** - Deployments and project management for Vercel frontend cloud ([#113](https://github.com/Lo10Th/Elysium/pull/113))

### Planned Features
- `ely remove <name>` - Uninstall emblem
- `ely publish` - Publish emblem to registry
- Web UI for browsing emblems
- Emblem marketplace with ratings

---

## [1.0.0] - 2026-03-04

### Added
- **Self-Update Command** - `ely self-update` updates the CLI binary in place
  - `ely self-update --check` checks for updates without installing
  - `ely self-update --version v1.0.0` installs a specific version
  - `ely self-update --force` reinstalls even if already up to date
- **Browser-Based Authentication** - Device-flow OAuth via `ely login`
  - Displays a one-time device code and verification URL
  - Attempts to open the browser automatically
  - Polls for authorization with a live spinner
- **Update & Outdated Commands** - `ely update` and `ely outdated` are now fully implemented
  - `ely update [emblem...]` updates one or more installed emblems
  - `ely update --all` updates every installed emblem at once
  - `ely outdated` lists installed emblems that have newer versions available
- **Improved Error Messages** - Actionable suggestions on common failures (emblem not found, auth required, registry unreachable)
- **Better Performance** - Faster emblem resolution and reduced startup overhead
- **Integration Tests** - End-to-end CLI integration tests for the emblem pull → execute pipeline ([#102](https://github.com/Lo10Th/Elysium/pull/102))
  - Tests gated behind `//go:build integration`, using `httptest` servers and `t.TempDir()` — no external services required
  - Happy-path scenarios: cache write/load/execute, path param substitution, query param forwarding
  - Error scenarios: 404 emblems, auth failures, network timeouts, malformed responses
- **Shared HTTP Client Pool** - New `cli/internal/httpclient` package with a single `*http.Transport` shared across all CLI operations ([#90](https://github.com/Lo10Th/Elysium/pull/90))
  - Connection reuse (`MaxIdleConns=100`, `MaxIdleConnsPerHost=10`, `IdleConnTimeout=90s`)
  - Replaces per-request `http.Client` creation in `login.go`, `checker.go`, and `downloader.go`

### Changed
- Version bumped from 0.2.x to 1.0.0 (first stable release)
- `ely update` and `ely outdated` moved from *Planned* to *Implemented*
- Install script default version updated to `v1.0.0`
- **Server: Async Route Handlers** - All Supabase calls now run in a thread pool via `asyncio.to_thread()` through new `run_sync()` helper in `database.py`, preventing event-loop blocking under concurrency ([#97](https://github.com/Lo10Th/Elysium/pull/97))
- **Server: Pydantic Models Consolidated** - All auth request/response models moved from inline definitions in `routes/auth.py` to the canonical `app/models.py` ([#96](https://github.com/Lo10Th/Elysium/pull/96))
- **Server: Emblem Construction Unified** - All `Emblem(...)` constructions go through `_row_to_emblem()` helper, with improved `author_name` resolution for both joined rows and RPC responses ([#95](https://github.com/Lo10Th/Elysium/pull/95))
- **Server: Service Layer Extracted** - Business logic moved from route handlers into dedicated service classes (`AuthService`, `EmblemService`, `KeyService`) in new `server/app/services/` package ([#94](https://github.com/Lo10Th/Elysium/pull/94))
- **CLI: Login Split into Focused Modules** - `login.go` (506 lines) split into `login.go`, `login_password.go`, `login_device.go`, and `login_oauth.go` by concern ([#93](https://github.com/Lo10Th/Elysium/pull/93))
- **CLI: API Error Handling Consolidated** - Extracted `handleAPIError` helper in `cli/internal/api/client.go`, replacing 9 duplicate error-handling blocks ([#92](https://github.com/Lo10Th/Elysium/pull/92))
- **CLI: Regex Pre-Compilation** - All four validator regexes in `cli/internal/validator/validator.go` lifted to package-level `var` block using `regexp.MustCompile` for performance ([#91](https://github.com/Lo10Th/Elysium/pull/91))
- **CLI cmd package coverage** improved from 17.5% → 74.6% with 14 new test files ([#101](https://github.com/Lo10Th/Elysium/pull/101))
- **Executor package coverage** improved from 50.3% → 95.4% ([#100](https://github.com/Lo10Th/Elysium/pull/100))

### Fixed
- **Server: Test Suite** - 11 failing tests fixed by updating profile query mocks after auth refactor added `profiles` table lookups ([#89](https://github.com/Lo10Th/Elysium/pull/89))
- **Server: OAuth NameError** - Restored module-level `oauth_states: dict[str, str] = {}` in `routes/auth.py` that was accidentally removed during device-code flow refactor ([#88](https://github.com/Lo10Th/Elysium/pull/88))

### Breaking Changes
- None. All existing commands and emblem YAML files remain compatible.

---

## [0.2.1] - 2026-03-04

### Added
- **Self-Update Command** - `ely self-update` for in-tool upgrades without re-running the install script ([#45](https://github.com/Lo10Th/Elysium/pull/45))
  - `ely self-update` — download and install the latest release
  - `ely self-update --check` — report available update without installing
  - `ely self-update --version v0.3.0` — install a specific version
  - `ely self-update --force` — reinstall even if already on the current version
  - Proper semver comparison (`0.10.0 > 0.2.0` handled correctly)
  - Atomic binary replacement (Unix: `os.Rename`; Windows: rename-to-old pattern with restart prompt)
  - HTTPS enforced; 10-minute download timeout
  - New `cli/internal/selfupdate` package (`checker.go`, `downloader.go`, `replacer.go`)
- **Documentation Expansion** ([#46](https://github.com/Lo10Th/Elysium/pull/46))
  - `docs/ARCHITECTURE.md` — component overview, package tables, sequence diagrams for pull/publish/execute flows
  - `docs/SECURITY.md` — supported versions, vulnerability reporting, credential storage model, URL validation policy, CORS/rate-limiting guidance
  - `docs/TROUBLESHOOTING.md` — categorized troubleshooting (install, auth, pull, execute, server, Go build, Python server)
  - `examples/stripe/emblem.yaml` — Stripe customers CRUD and payment intents
  - `examples/github/emblem.yaml` — GitHub repos, issues, pull requests, users
  - `examples/slack/emblem.yaml` — Slack messages, channels, and users
  - README updated with Documentation and Example Emblems navigation tables

---

## [0.2.0] - 2026-03-03

### Added
- **Testing Infrastructure** - 146 tests added, 50% Go coverage ([#38](https://github.com/Lo10Th/Elysium/pull/38))
  - Config package tests (78.8%)
  - Emblem package tests (92.6%)
  - Validator package tests (100%)
  - API package tests (73.5%)
  - Scaffold package tests (83.7%)
- **Shell Completion** - `ely completion [bash|zsh|fish|powershell]` ([#40](https://github.com/Lo10Th/Elysium/pull/40))
  - Dynamic completion for emblem names
  - Dynamic completion for action names
  - Dynamic completion for action parameters
- **Install Script** - One-line installation ([#41](https://github.com/Lo10Th/Elysium/pull/41))
  - `curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash`
  - OS/architecture auto-detection
  - Version selection (`--version v0.2.0`)
  - Shell completion setup
- **Uninstall Script** - Clean uninstall with `--purge` option ([#41](https://github.com/Lo10Th/Elysium/pull/41))
- **Homebrew Distribution** - `brew install Lo10Th/tap/ely` ([#39](https://github.com/Lo10Th/Elysium/pull/39))
- **CI Status Badges** - Test and Release workflow badges in README ([#42](https://github.com/Lo10Th/Elysium/pull/42))
- **Batch Commands** - `ely pull <name1> <name2>`, `ely update [name]`, `ely outdated` ([#31](https://github.com/Lo10Th/Elysium/pull/31))
- **Enhanced Output Formatting** - CSV, YAML, pretty JSON, field selection, templates, color tables ([#30](https://github.com/Lo10Th/Elysium/pull/30))
- **Update Notifications** - `ely check-updates`, security advisories, version cache, `--no-check` flag ([#32](https://github.com/Lo10Th/Elysium/pull/32))
- **Security Hardening** - Rate limiting, security headers, input validation, and URL safety checks ([#34](https://github.com/Lo10Th/Elysium/pull/34))

### Changed
- CI coverage threshold enforced (Go: 40%, Python: 80%)
- Tests now block PRs on failure (removed `continue-on-error`)

### Installation Methods
1. One-line: `curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash`
2. Homebrew: `brew install Lo10Th/tap/ely`
3. Go install: `go install github.com/Lo10Th/Elysium/cli/cmd/ely@latest`
4. Manual download from GitHub Releases

---

## [0.1.1] - 2026-03-03

### Added
- Email/credential authentication for CLI (`ely login` now prompts for credentials)
- Registration flow within login (prompts to register if account doesn't exist)
- User info storage (email, username) in config

### Changed
- `ely login` now uses email/credential auth instead of OAuth
- `ely logout` clears all stored auth data
- `ely whoami` shows stored user info

### Fixed
- OAuth endpoints now return helpful error messages (501 Not Implemented)
- Credential input is hidden during login

---

## [0.1.0] - 2026-03-03

### Added
- Dynamic emblem execution - `ely <emblem> <action>` works
- Auth token management with email/credential authentication
- CLI API key management commands (`ely keys`)
- Emblem scaffolding with `ely init` ([#27](https://github.com/Lo10Th/Elysium/pull/27))
- Local validation with `ely validate` ([#28](https://github.com/Lo10Th/Elysium/pull/28))
- Local testing with `ely test` ([#28](https://github.com/Lo10Th/Elysium/pull/28))
- Error message improvements with actionable suggestions ([#26](https://github.com/Lo10Th/Elysium/pull/26))
- Configuration management with `ely config` ([#29](https://github.com/Lo10Th/Elysium/pull/29))
- Registry server deployed at https://ely.karlharrenga.com
- Supabase backend with authentication
- FastAPI registry server with Vercel deployment

### Changed
- Updated default registry to https://ely.karlharrenga.com
- Improved error messages with suggestions
- Fixed lazy initialization for Supabase client

### Fixed
- Vercel serverless function deployment issues
- Environment variable handling for Supabase keys

---

[Unreleased]: https://github.com/Lo10Th/Elysium/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Lo10Th/Elysium/compare/v0.2.1...v1.0.0
[0.2.1]: https://github.com/Lo10Th/Elysium/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/Lo10Th/Elysium/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/Lo10Th/Elysium/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Lo10Th/Elysium/releases/tag/v0.1.0