# Changelog

All notable changes to Elysium will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-03-03

### Added
- **Testing Infrastructure** - 146 tests added, 50% Go coverage
  - Config package tests (78.8%)
  - Emblem package tests (92.6%)
  - Validator package tests (100%)
  - API package tests (73.5%)
  - Scaffold package tests (83.7%)
- **Shell Completion** - `ely completion [bash|zsh|fish|powershell]`
  - Dynamic completion for emblem names
  - Dynamic completion for action names
  - Dynamic completion for action parameters
- **Install Script** - One-line installation
  - `curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash`
  - OS/architecture auto-detection
  - Version selection (`--version v0.2.0`)
  - Shell completion setup
- **Uninstall Script** - Clean uninstall with `--purge` option
- **CI Status Badges** - Test and Release workflow badges in README

### Changed
- CI coverage threshold enforced (Go: 40%, Python: 80%)
- Tests now block PRs on failure (removed `continue-on-error`)

### Installation Methods
1. One-line: `curl -sSL https://raw.githubusercontent.com/Lo10Th/Elysium/main/scripts/install.sh | bash`
2. Go install: `go install github.com/Lo10Th/Elysium/cli/cmd/ely@latest`
3. Manual download from GitHub Releases

---

## [0.1.1] - 2026-03-03

### Added
- Email/password authentication for CLI (`ely login` now prompts for credentials)
- Registration flow within login (prompts to register if account doesn't exist)
- User info storage (email, username) in config

### Changed
- `ely login` now uses email/password instead of OAuth
- `ely logout` clears all stored auth data
- `ely whoami` shows stored user info

### Fixed
- OAuth endpoints now return helpful error messages (501 Not Implemented)
- Password input is hidden during login

---

## [0.1.0] - 2026-03-03

### Added
- Dynamic emblem execution - `ely <emblem> <action>` works
- Auth token management with email/password authentication
- CLI API key management commands (`ely keys`)
- Emblem scaffolding with `ely init`
- Local validation with `ely validate`
- Local testing with `ely test`
- Error message improvements with actionable suggestions
- Configuration management with `ely config`
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

## [Unreleased]

### Planned Features
- `ely update <name>` - Update emblem to latest version
- `ely remove <name>` - Uninstall emblem
- `ely publish` - Publish emblem to registry
- `ely completion` - Shell completion
- Web UI for browsing emblems
- Emblem marketplace with ratings

---

[Unreleased]: https://github.com/Lo10Th/Elysium/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/Lo10Th/Elysium/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/Lo10Th/Elysium/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Lo10Th/Elysium/releases/tag/v0.1.0