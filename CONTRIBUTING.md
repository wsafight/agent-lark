# Contributing to agent-lark

Thanks for your interest in contributing!

## Getting Started

```bash
git clone https://github.com/wsafight/agent-lark.git
cd agent-lark
make build
make test
```

## Development Workflow

1. Fork the repository and create a feature branch from `main`
2. Make your changes
3. Run tests: `make test`
4. Run linter: `make lint` (requires [golangci-lint](https://golangci-lint.run/usage/install/))
5. Commit with a clear message following [Conventional Commits](https://www.conventionalcommits.org/) (e.g. `feat:`, `fix:`, `refactor:`)
6. Open a pull request

## Code Style

- Error codes use the format `"UPPER_SNAKE_CASE：message"` (full-width Chinese colon)
- Use `cmdutil.ResolveGlobalFlags(cmd)` for global flag resolution in all commands
- Use `g.NewClient()` for creating Feishu API clients
- Follow existing patterns in `internal/` packages

## Adding a New Command

1. Create a new package under `internal/`
2. Implement `NewCommand() *cobra.Command`
3. Register it in `cmd/agent-lark/main.go`
4. Add tests (at minimum: command registration and flag validation)
5. Update documentation: `README.md`, `README.zh.md`, `cmd/agent-lark/SKILL.md`

## Reporting Issues

Open an issue at https://github.com/wsafight/agent-lark/issues with:
- Steps to reproduce
- Expected vs actual behavior
- `agent-lark version` output
