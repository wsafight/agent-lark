# AGENTS.md

Instructions for AI coding agents working with this codebase.

## Build and Test

```bash
go build ./...                   # Compile all packages
go build -o agent-lark ./cmd/agent-lark   # Build CLI binary
go test ./...                    # Run all tests
go vet ./...                     # Static analysis
```

## Code Style

- Error strings must use the `UPPER_SNAKE_CASE：message` convention (full-width colon `：`).
  Examples: `"INVALID_URL：无法解析文档 token"`, `"MISSING_FLAG：请提供 --file 或 --content"`.
- All global flags are read via `cmdutil.ResolveGlobalFlags(cmd)` (applies agent/format side effects) or `cmdutil.GetGlobalFlags(cmd)` (pure read). Never read root flags directly from `cmd.Root().PersistentFlags()` in command handlers. Use `g.ClientOptions()` to pass options to business logic, and `g.NewClient()` to create a Feishu client.
- Paginated list commands in `--agent` mode must return `cmdutil.PagedResponse{Items: items, NextCursor: token}`. Non-agent mode returns the slice directly.
- CLI flags use kebab-case (`--token-mode`, `--chat-id`). Never camelCase.
- Do not add emojis to CLI output or source code. Unicode symbols (✓, ✗) are acceptable.
- Do not add unnecessary error handling for internal invariants. Validate only at system boundaries (user input, API responses).

## Documentation

When adding or changing user-facing features (new flags, commands, behaviors, environment variables):

1. `README.md` — English command reference table and examples
2. `README.zh.md` — Chinese mirror, kept in sync with README.md
3. `cmd/agent-lark/SKILL.md` — AI agent skill reference (so agents know about the feature)

## Architecture

- `cmd/agent-lark/` — entry point and root cobra command (`main.go` registers all subcommands)
- `internal/cmdutil/` — global flag resolution and shared pagination types
- `internal/client/` — Feishu SDK client factory; resolves token mode and calls `auth.EnsureUserTokenValid`
- `internal/auth/` — credential storage (`~/.agent-lark/profiles/<name>.json`), OAuth flow, token refresh
- `internal/output/` — Markdown rendering, JSON printing, `GlobalAgent` flag
- `internal/docxutil/` — block fetching and Markdown↔block conversion shared by docs, wiki, template

## Environment Variables

All global flags have `AGENT_LARK_*` environment variable equivalents. Explicit flags take precedence.

| Env var | Equivalent flag |
|---------|----------------|
| `AGENT_LARK_FORMAT` | `--format` |
| `AGENT_LARK_TOKEN_MODE` | `--token-mode` |
| `AGENT_LARK_PROFILE` | `--profile` |
| `AGENT_LARK_CONFIG` | `--config` |
| `AGENT_LARK_DOMAIN` | `--domain` |
| `AGENT_LARK_DEBUG` | `--debug` |
| `AGENT_LARK_QUIET` | `--quiet` |
| `AGENT_LARK_AGENT` | `--agent` |
