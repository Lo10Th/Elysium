# Changelog

All notable changes to Elysium will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planning
- `ely remove <name>` - Uninstall emblem
- `ely publish` - Publish emblem to registry
- Web UI for browsing emblems
- Emblem marketplace with ratings

---

## [0.2.5] - 2026-03-06

### Security
- **Device Code Permissions** - Fixed PostgreSQL permissions for anonymous device code creation ([#126](https://github.com/Lo10Th/Elysium/issues/126))
  - Added GRANT permissions for `anon` role on `device_codes` and `emblem_pulls` tables
  - Improved UPDATE policy to restrict modifications to unverified codes only
- **Config File Permissions** - CLI config now uses 0600 permissions to restrict token access to owner only

### Fixed
- **CLI Login** - Device code creation now works with proper PostgreSQL RLS policies ([#124](https://github.com/Lo10Th/Elysium/issues/124))
- **CHANGELOG** - Corrected version history to reflect actual releases (removed fictional v1.0.0 entry)

---

## [0.2.4] - 2026-03-06

### Fixed
- **CLI Login** - Device code timestamp calculation (ISO 8601 format instead of raw SQL)
- **Self-Update** - Binary extraction from tar.gz archives

---

## [0.2.3] - 2026-03-06

### Fixed
- **Self-Update** - Binary extraction from tar.gz archives

---

## [0.2.2] - 2026-03-05

### Added
- **OpenAI Emblem** - Chat completions (GPT-4, GPT-3.5-turbo), model listing, embeddings, and DALL-E image generation ([#109](https://github.com/Lo10Th/Elysium/pull/109))
- **Anthropic Emblem** - Claude message creation and model listing ([#109](https://github.com/Lo10Th/Elysium/pull/109))
- **Twilio API Emblem** - SMS and phone verification ([#115](https://github.com/Lo10Th/Elysium/pull/115))
- **AWS S3 Emblem** - File storage operations ([#114](https://github.com/Lo10Th/Elysium/pull/114))
- **Vercel Emblem** - Deployments and project management ([#113](https://github.com/Lo10Th/Elysium/pull/113))
- **SendGrid Emblem** - Transactional email API ([#111](https://github.com/Lo10Th/Elysium/pull/111))
- **Supabase Emblem** - Database and auth operations ([#110](https://github.com/Lo10Th/Elysium/pull/110))
- **Auth0 Emblem** - Management API v2 ([#117](https://github.com/Lo10Th/Elysium/pull/117))
- **Notion Emblem** - API integration ([#119](https://github.com/Lo10Th/Elysium/pull/119))
- **Cloudflare Emblem** - DNS/CDN management ([#121](https://github.com/Lo10Th/Elysium/pull/121))
- **Sentry Emblem** - Issue tracking API ([#123](https://github.com/Lo10Th/Elysium/pull/123))
- **Linear Emblem** - Issue tracking integration
- **GitHub Emblem** - Repository and PR operations ([#120](https://github.com/Lo10Th/Elysium/pull/120))
- **Stripe Emblem Enhancement** - Payment intents, improved schemas ([#112](https://github.com/Lo10Th/Elysium/pull/112))
- **Mapbox Emblem** - Location services ([#122](https://github.com/Lo10Th/Elysium/pull/122))
- **Verified Author Badges** - CLI search output shows verified status for authors
- **Integration Tests** - End-to-end CLI tests for emblem execution ([#102](https://github.com/Lo10Th/Elysium/pull/102))
- **Testing Agents** - Felix agent for test generation, Markus agent for emblem creation

### Documentation
- Added READMEs for all example emblems
- Added ARCHITECTURE.md, SECURITY.md, TROUBLESHOOTING.md, CONTRIBUTING.md

### Changed
- **Server: Async Route Handlers** - Supabase calls run in thread pool via `run_sync()` ([#97](https://github.com/Lo10Th/Elysium/pull/97))
- **Server: Pydantic Models Consolidated** - All models in `app/models.py` ([#96](https://github.com/Lo10Th/Elysium/pull/96))
- **Server: Emblem Construction Unified** - `_row_to_emblem()` helper ([#95](https://github.com/Lo10Th/Elysium/pull/95))
- **Server: Service Layer** - Business logic in `server/app/services/` ([#94](https://github.com/Lo10Th/Elysium/pull/94))
- **CLI: Login Modules** - Split into focused files ([#93](https://github.com/Lo10Th/Elysium/pull/93))
- **CLI: Error Handling** - Consolidated `handleAPIError` helper ([#92](https://github.com/Lo10Th/Elysium/pull/92))
- **CLI: Regex Pre-Compilation** - Validators compiled at init ([#91](https://github.com/Lo10Th/Elysium/pull/91))
- **CLI: HTTP Client Pool** - Shared transport for connection reuse ([#90](https://github.com/Lo10Th/Elysium/pull/90))
- **Test Coverage** - cmd package: 17.5% → 74.6%, executor: 50.3% → 95.4% ([#100](https://github.com/Lo10Th/Elysium/pull/100), [#101](https://github.com/Lo10Th/Elysium/pull/101))

### Fixed
- OAuth states NameError after device-code refactor ([#88](https://github.com/Lo10Th/Elysium/pull/88))
- Test failures after auth refactor ([#89](https://github.com/Lo10Th/Elysium/pull/89))

---

## [0.2.1] - 2026-03-04

### Added
- **Self-Update Command** - `ely self-update` for in-tool upgrades ([#45](https://github.com/Lo10Th/Elysium/pull/45))
  - `ely self-update --check` reports available updates
  - `ely self-update --version v0.3.0` installs specific version
  - `ely self-update --force` reinstalls current version
  - Proper semver comparison
  - Atomic binary replacement
- **Documentation** - ARCHITECTURE.md, SECURITY.md, TROUBLESHOOTING.md
- **Example Emblems** - Stripe, GitHub, Slack YAML examples

### Fixed
- Test failures after auth refactor
- Profile query mocks updated

---

## [0.2.0] - 2026-03-03

### Added
- **Testing Infrastructure** - 146 tests, 50% Go coverage ([#38](https://github.com/Lo10Th/Elysium/pull/38))
- **Shell Completion** - `ely completion [bash|zsh|fish|powershell]` ([#40](https://github.com/Lo10Th/Elysium/pull/40))
  - Dynamic completion for emblems, actions, and parameters
- **Install Script** - One-line installation ([#41](https://github.com/Lo10Th/Elysium/pull/41))
- **Uninstall Script** - Clean removal with `--purge` option ([#41](https://github.com/Lo10Th/Elysium/pull/41))
- **Homebrew Distribution** - `brew install Lo10Th/tap/ely` ([#39](https://github.com/Lo10Th/Elysium/pull/39))
- **CI Status Badges** - Test and Release workflow badges ([#42](https://github.com/Lo10Th/Elysium/pull/42))
- **Batch Commands** - `ely pull <name1> <name2>`, `ely update`, `ely outdated` ([#31](https://github.com/Lo10Th/Elysium/pull/31))
- **Enhanced Output** - CSV, YAML, JSON, templates, tables ([#30](https://github.com/Lo10Th/Elysium/pull/30))
- **Update Notifications** - `ely check-updates`, version cache ([#32](https://github.com/Lo10Th/Elysium/pull/32))
- **Security Hardening** - Rate limiting, input validation ([#34](https://github.com/Lo10Th/Elysium/pull/34))

### Changed
- CI coverage threshold enforced (Go: 40%, Python: 80%)
- Tests now block PRs on failure

---

## [0.1.1] - 2026-03-03

### Added
- Email/credential authentication for CLI (`ely login`)
- Registration flow within login

### Changed
- `ely login` uses email/credential auth instead of OAuth
- `ely whoami` shows stored user info

---

## [0.1.0] - 2026-03-03

### Added
- Dynamic emblem execution - `ely <emblem> <action>` works
- Auth token management with OAuth flow and keyring storage
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
- Default registry set to https://ely.karlharrenga.com
- Improved error messages with suggestions
- Fixed lazy initialization for Supabase client

### Fixed
- Vercel serverless function deployment issues
- Environment variable handling for Supabase keys

---

[Unreleased]: https://github.com/Lo10Th/Elysium/compare/v0.2.5...HEAD
[0.2.5]: https://github.com/Lo10Th/Elysium/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/Lo10Th/Elysium/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/Lo10Th/Elysium/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/Lo10Th/Elysium/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/Lo10Th/Elysium/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/Lo10Th/Elysium/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/Lo10Th/Elysium/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Lo10Th/Elysium/releases/tag/v0.1.0