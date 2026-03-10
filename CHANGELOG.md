# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- MIT LICENSE file
- Makefile with build/test/lint/cover/install targets
- golangci-lint configuration (`.golangci.yml`)
- CI lint job via `golangci-lint-action`
- Unit tests for `im`, `comments`, `contact`, `doctor` packages
- `CHANGELOG.md` and `CONTRIBUTING.md`

## [0.1.0] - 2026-03-11

### Added
- Core CLI framework with Cobra (`agent-lark` binary)
- Document operations: list, get, create, update, search, outline, table insertion
- Wiki operations: list, get, create with hierarchical support
- Bitable operations: list, get, create, update, batch, count, field inspection
- Instant messaging: send, list, get, chat management, reactions
- Task management: list, get, create, update
- Contact search and resolution (email -> open_id lookup)
- Document comment management: list, add, reply, resolve
- Permission management: list, add, remove, update, transfer, public sharing, check
- Template system: save, apply, variable resolution, local storage
- Authentication: app credentials, OAuth flow, encrypted credential storage (AES-256-GCM)
- Profile system: SHA-256 hash-based auto project profiles
- Agent mode (`--agent`): JSON output, structured error to stderr
- Doctor command for configuration diagnostics
- Self-upgrade mechanism
- Multi-platform release CI (linux/darwin/windows, amd64/arm64)
- `install.sh` for quick installation
- Documentation: README.md (English), README.zh.md (Chinese), SKILL.md (agent reference)
