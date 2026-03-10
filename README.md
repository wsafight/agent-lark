# agent-lark

Feishu / Lark CLI for AI agents. Docs, wiki, bitable, messages, tasks, permissions, and more.

[中文文档](README.zh.md)

## Installation

**Requirements:** Go 1.22+, a Feishu / Lark app (App ID + App Secret)

```bash
git clone https://github.com/wsafight/agent-lark.git
cd agent-lark
go build -o agent-lark ./cmd/agent-lark
mv agent-lark /usr/local/bin/   # optional
```

## Quick Start

```bash
agent-lark init                 # First-time setup: install Skill + configure credentials
agent-lark setup                # Configure app credentials (interactive wizard)
agent-lark auth oauth           # Add OAuth user token (for personal Drive access)
agent-lark auth status          # Check auth status

agent-lark docs list
agent-lark docs get "<doc-url>"
agent-lark im send --chat-id <id> --text "Hello"
agent-lark task list
```

## Authentication

### App credentials

```bash
agent-lark setup                                        # Interactive wizard (self-built apps)
agent-lark auth login                                   # Quick login with built-in public app (special builds only)
```

### OAuth user token

Required for operations that act on behalf of a user (personal Drive, user tasks, etc.):

```bash
agent-lark auth oauth                                   # Opens browser, saves token
agent-lark auth oauth --scope "docx:readonly drive:readonly" --port 9999
```

### Token mode

| Mode | Behavior |
|------|----------|
| `auto` (default) | User token when available, falls back to tenant token |
| `tenant` | Always use app bot identity (tenant_access_token) |
| `user` | Always use OAuth user identity (user_access_token) |

```bash
agent-lark auth set-mode user
```

### Profiles

Manage multiple apps or teams with named profiles. Each project directory is automatically assigned an isolated profile based on its path (SHA-256 hash).

```bash
agent-lark setup --profile work       # Configure credentials for a named profile
agent-lark auth profile list          # List all profiles
agent-lark --profile work docs list   # One-off override
```

### Other auth commands

```bash
agent-lark auth status               # Show current auth status
agent-lark auth logout --user        # Clear user token only
agent-lark auth logout --all         # Clear all credentials
```

## Global Flags

All flags can also be set via environment variables (explicit flag takes precedence).

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--format text\|json\|md` | `AGENT_LARK_FORMAT` | `text` | Output format |
| `--token-mode auto\|tenant\|user` | `AGENT_LARK_TOKEN_MODE` | `auto` | Token mode |
| `--profile <name>` | `AGENT_LARK_PROFILE` | — | Named credential profile |
| `--config <path>` | `AGENT_LARK_CONFIG` | — | Explicit credential file path |
| `--domain <domain>` | `AGENT_LARK_DOMAIN` | — | Override API domain (e.g. `open.larksuite.com` for Lark international) |
| `--agent` | `AGENT_LARK_AGENT` | — | Agent mode: force JSON + structured errors + auto-confirm |
| `--quiet` | `AGENT_LARK_QUIET` | — | Suppress informational output, print data only |
| `--debug` | `AGENT_LARK_DEBUG` | — | Print raw HTTP requests and responses |
| `--yes` | — | — | Auto-confirm all prompts |

## Commands

### auth

```bash
agent-lark setup                                        # Interactive credential wizard
agent-lark auth login                                   # Quick login (built-in public app)
agent-lark auth oauth                                   # Authorize OAuth user token
agent-lark auth status                                  # Show auth status
agent-lark auth set-mode auto|tenant|user               # Set default token mode
agent-lark auth profile list                            # List profiles
agent-lark auth logout --user                           # Clear user token
agent-lark auth logout --all                            # Full reset
```

### docs

```bash
agent-lark docs list                                    # List files in Drive root
agent-lark docs list --folder "<url>"                   # List files in folder
agent-lark docs list --since 7d                         # Filter by modified time (24h / 7d / today / YYYY-MM-DD)
agent-lark docs list --all                              # Auto-paginate
agent-lark docs list --limit 50

agent-lark docs get "<url>"                             # Get doc content (Markdown by default)
agent-lark docs get "<url>" --format json
agent-lark docs get "<url>" --section "Background"      # Return only the matching section
agent-lark docs get "<url>" --metadata                  # Metadata only (no content)
agent-lark docs get "<url>" --content-boundaries        # Wrap output in <document> tags (prompt injection protection)
agent-lark docs get "<url>" --max-chars 8000            # Truncate output to N characters

agent-lark docs outline "<url>"                         # Return heading outline

agent-lark docs search "keyword"
agent-lark docs search "keyword" --limit 10
agent-lark docs search "keyword" --exists               # Exit 0 if found, print found/not_found

agent-lark docs create --title "Weekly Report W26"
agent-lark docs create --title "Design Doc" --folder "<url>"

agent-lark docs update "<url>" --content "Appended paragraph"
agent-lark docs update "<url>" --file content.md
agent-lark docs update "<url>" --stdin < content.md

agent-lark docs table "<url>" --rows 3 --cols 4
agent-lark docs table "<url>" --after "Summary" --data '[["A","B"],["1","2"]]'
agent-lark docs table "<url>" --file data.csv --headers
agent-lark docs table "<url>" --dry-run                 # Preview without writing
```

### wiki

```bash
agent-lark wiki list "<url>"                            # List wiki space nodes
agent-lark wiki list "<url>" --depth 2                  # Expand up to 2 levels deep
agent-lark wiki list "<url>" --all                      # Auto-paginate

agent-lark wiki get "<url>"                             # Get wiki page content
agent-lark wiki get "<url>" --format json
agent-lark wiki get "<url>" --content-boundaries        # Wrap output in <document> tags
agent-lark wiki get "<url>" --max-chars 8000            # Truncate output to N characters

agent-lark wiki create "<space-url>" --title "New Page"
```

### base (Bitable)

The URL argument is the full Feishu Bitable page URL — `appToken` and `tableId` are parsed automatically.

```bash
agent-lark base fields "<url>"                          # Show table field schema

agent-lark base records list "<url>"
agent-lark base records list "<url>" --limit 50 --all
agent-lark base records list "<url>" --filter 'CurrentValue.[Status]="In Progress"'
agent-lark base records list "<url>" --select "Name,Status,Due"
agent-lark base records list "<url>" --cursor "<token>"

agent-lark base records get "<url>" <record_id>

agent-lark base records count "<url>"
agent-lark base records count "<url>" --filter 'CurrentValue.[Status]="Done"'

agent-lark base records create "<url>" --field "Name=Alice" --field "Age=30"
agent-lark base records create "<url>" --fields '{"Name":"Bob","Status":"Todo"}'

agent-lark base records update "<url>" <record_id> --field "Status=Done"

agent-lark base records batch-create "<url>" --file records.json
# records.json: [{"Name":"Alice","Age":25}, {"Name":"Bob","Age":30}]
```

### im

```bash
agent-lark im send --chat-id <id> --text "Hello"
agent-lark im send --user-id <open_id> --text "DM"     # Creates 1-on-1 chat automatically
agent-lark im send --chat-id <id> --card '<json>'
agent-lark im send --chat-id <id> --card-file card.json

agent-lark im chats list
agent-lark im chats list --limit 50 --all
agent-lark im chats search "keyword"

agent-lark im messages list --chat-id <id>
agent-lark im messages list --chat-id <id> --limit 20
agent-lark im messages get --message-id <id>

agent-lark im react add --message-id <id> --emoji THUMBSUP
agent-lark im react remove --message-id <id> --emoji THUMBSUP
```

### task

```bash
agent-lark task list
agent-lark task list --status todo|done
agent-lark task list --assignee <open_id>
agent-lark task list --limit 30 --cursor "<token>"

agent-lark task get --task-id <guid>

agent-lark task create --title "Finish spec review"
agent-lark task create --title "Launch" --due 2025-06-30
agent-lark task create --title "Follow up" --due 2025-06-30T18:00:00+08:00 --assignee <open_id>

agent-lark task update --task-id <guid> --title "New title"
agent-lark task update --task-id <guid> --status done
agent-lark task update --task-id <guid> --due 2025-07-01
```

### comments

```bash
agent-lark comments list "<url>"
agent-lark comments list "<url>" --format json

agent-lark comments add "<url>" --content "Needs revision"
agent-lark comments add "<url>" --content "See here" --quote '<selection-json>'

agent-lark comments reply "<url>" --comment-id <id> --content "Fixed"

agent-lark comments resolve "<url>" --comment-id <id>
```

### perms

```bash
agent-lark perms list "<url>"                           # List collaborators
agent-lark perms check "<url>"                          # Check current user's permission

agent-lark perms add "<url>" --user <open_id> --role view|edit|full_access
agent-lark perms add "<url>" --user <id1> --user <id2> --role edit
agent-lark perms add "<url>" --user <open_id> --role edit --no-notify

agent-lark perms update "<url>" --user <open_id> --role view

agent-lark perms remove "<url>" --user <open_id>

agent-lark perms transfer "<url>" --user <open_id>      # Transfer ownership

agent-lark perms public "<url>"                         # Show public access settings
agent-lark perms public "<url>" --link-share anyone_readable
```

### contact

```bash
agent-lark contact search user@example.com              # Lookup user by email

agent-lark contact list
agent-lark contact list --dept "Product"                # Filter by department name
agent-lark contact list --limit 100 --all

agent-lark contact resolve --email user@example.com     # Resolve user identifiers
```

### template

Templates are stored locally in Markdown format and support `{{variable}}` substitution.

**Built-in variables**

| Variable | Value |
|----------|-------|
| `{{date}}` | Current date (`2025-06-30`) |
| `{{datetime}}` | Current date and time (`2025-06-30 14:30`) |
| `{{week}}` | ISO week number (`W26`) |
| `{{author}}` | Name of the authenticated user |
| `{{title}}` | Document title (from `--title`) |

```bash
agent-lark template list

agent-lark template save weekly --file weekly.md
agent-lark template save weekly --file weekly.md --force   # Overwrite
agent-lark template save weekly --from "<url>"             # Pull from Feishu doc

agent-lark template get weekly
agent-lark template vars weekly                            # List variables used

agent-lark template apply weekly --new --title "W26 Weekly Report"
agent-lark template apply weekly --new --title "Report" --folder "<url>"
agent-lark template apply weekly --to "<url>"              # Append to existing doc
agent-lark template apply weekly --to "<url>" --after "Last week"
agent-lark template apply weekly --to "<url>" --before "Next week"
agent-lark template apply weekly --new --title "Report" --var "project=Rocket" --var "owner=Alice"
agent-lark template apply weekly --new --title "Report" --dry-run   # Preview only

agent-lark template delete weekly
```

### doctor

```bash
agent-lark doctor    # Check config file, credentials, API reachability, token expiry
```

## Agent Mode

`--agent` is designed for use inside AI agent pipelines:

- Forces `--format json` on all output
- Errors are written to stderr as structured JSON: `{"error": "CODE", "message": "...", "hint": "..."}`
- List commands return `{"items": [...], "next_cursor": "..."}` for stateless pagination
- Skips all interactive prompts (implies `--yes`)

```bash
# Basic usage
agent-lark --agent docs list --folder "<url>"

# Paginated reads — pass next_cursor until it is empty
agent-lark --agent docs list --limit 20
agent-lark --agent docs list --limit 20 --cursor "<next_cursor from previous response>"

# Error output (stderr)
{"error": "AUTH_REQUIRED", "message": "credentials not configured", "hint": "run: agent-lark auth login"}
```

## Error Codes

| Code | Meaning | Suggested action |
|------|---------|-----------------|
| `AUTH_REQUIRED` | No credentials configured | `agent-lark auth login` |
| `TOKEN_EXPIRED` | OAuth user token expired | `agent-lark auth oauth` |
| `MISSING_FLAG` | Required flag not provided | Check `--help` |
| `MISSING_ARG` | Required positional argument missing | Check `--help` |
| `INVALID_URL` | Cannot parse Feishu URL | Use the full browser URL of the resource |
| `INVALID_JSON` | Malformed JSON argument | Check `--fields` or `--data` value |
| `INVALID_STATUS` | Unsupported status value | See `--help` for allowed values |
| `API_ERROR` | Feishu API returned an error | Check app permissions and request parameters |
| `CLIENT_ERROR` | Client initialization failed | Run `agent-lark doctor` to diagnose |
| `FILE_ERROR` | Cannot read file | Verify the path and read permissions |
| `NOT_FOUND` | Resource does not exist | Check the URL or search keyword |
