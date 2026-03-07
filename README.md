# agent-lark

`agent-lark` is a Go-based CLI for working with Feishu/Lark resources, including docs, wiki pages, bitable records, messages, comments, tasks, templates, and permissions.

## Features

- Authentication management (app credentials + OAuth user session)
- Read and update docs/wiki/bitable data
- Send IM messages and reactions
- Manage comments and permissions
- Template-based document workflows
- JSON/text/Markdown output and agent-friendly mode

## Requirements

- Go 1.22+
- A Feishu/Lark app (`App ID` and `App Secret`) for API access

## Build

```bash
go build -o agent-lark ./cmd/agent-lark
```

## Quick Start

```bash
# 1) Configure credentials interactively
./agent-lark setup

# 2) Check auth status
./agent-lark auth status

# 3) List docs
./agent-lark docs list
```

## Common Commands

```bash
./agent-lark login
./agent-lark setup
./agent-lark auth status
./agent-lark docs list
./agent-lark docs get "<doc-url-or-token>"
./agent-lark docs update "<doc-url-or-token>" --content "Updated content"
./agent-lark wiki list "<wiki-url-or-token>"
./agent-lark base records list "<base-url>"
./agent-lark im send --chat-id <chat_id> --text "Hello"
./agent-lark comments add "<doc-url>" --content "Looks good"
./agent-lark perms list "<doc-url>"
./agent-lark task list
./agent-lark template list
./agent-lark version
```

## Global Flags

- `--format text|json|md`
- `--token-mode auto|tenant|user`
- `--profile <name>`
- `--config <path>`
- `--domain <api-domain>`
- `--debug`
- `--quiet`
- `--yes`
- `--agent` (equivalent to agent-friendly JSON + auto-confirm behavior)

## Notes

- Credentials are stored in local profile files managed by the CLI.
- Use `--agent` for automation pipelines.

