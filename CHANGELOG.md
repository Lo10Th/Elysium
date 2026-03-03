# Changelog

All notable changes to Elysium will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[Unreleased]: https://github.com/Lo10Th/Elysium/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/Lo10Th/Elysium/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Lo10Th/Elysium/releases/tag/v0.1.0