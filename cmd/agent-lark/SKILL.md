# Agent Lark Skill

## Overview

You have access to the `agent-lark` CLI for interacting with Feishu/Lark. Use it to authenticate, read documents and messages, create and edit content, manage comments and permissions.

The CLI follows a **stateless command pattern**: each call is fully self-contained. State (credentials, tokens) is persisted in `~/.agent-lark/profiles/<profile>.json` automatically.

> Full command reference is evolving with the CLI source; use `agent-lark --help` and `<subcommand> --help` for the latest flags.

---

## Agent Mode Flag

For any automated workflow, **always add `--agent` to every command**. It combines:
- `--format json` — machine-readable stdout
- `--yes` — skips all `[y/N]` confirmation prompts
- Errors → JSON on stderr: `{"error":"CODE","message":"...","hint":"..."}`
- `next_cursor` added to all paginated list responses (empty string `""` = last page)

```bash
# Always use --agent in automated pipelines:
agent-lark docs list --agent
agent-lark perms update "$URL" --user alice@company.com --role view --agent
agent-lark perms transfer "$URL" --to alice@company.com --agent
```

> Without `--agent`, commands that modify permissions or transfer ownership will block waiting for stdin and hang the workflow.

**Paginating through all results:**

```bash
cursor=""
while true; do
  result=$(agent-lark docs list --agent ${cursor:+--cursor "$cursor"})
  # process items...
  cursor=$(echo "$result" | jq -r '.next_cursor')
  [ -z "$cursor" ] && break
done
```

---

## Authentication

Before calling any Lark API, valid credentials are required.

```bash
agent-lark login          # Personal use: opens browser, no App ID needed
agent-lark setup          # Team bot / advanced: interactive wizard for App ID + Secret
agent-lark auth status    # Check current auth state
agent-lark doctor         # Diagnose issues with fix suggestions
agent-lark auth logout --all  # Full reset
```

**CI/CD — environment variables:**

```bash
LARK_APP_ID=cli_xxx LARK_APP_SECRET=yyy agent-lark docs list
AGENT_LARK_PROFILE=team-a agent-lark docs list
```

---

## URL Cheat Sheet

Most commands take a URL copied directly from the browser:

| Resource | URL Pattern |
|----------|-------------|
| Document | `https://company.feishu.cn/docx/doxcn…` |
| Wiki page | `https://company.feishu.cn/wiki/wikcn…` |
| Bitable | `https://company.feishu.cn/base/bascn…?table=tbl…` |
| Chat ID | `oc_XXXXXXXX` — from `im chats list/search` |
| User open_id | `ou_XXXXXXXX` — from `contact search/resolve` |

Lark international URLs (`larksuite.com`) are equally supported.

---

## Reading Content

### Documents

```bash
agent-lark docs list [--limit 20] [--since "7d"] [--all]
agent-lark docs get "<URL>"                        # full content, returns Markdown
agent-lark docs get "<URL>" --section "Solution"   # single section only
agent-lark docs get "<URL>" --metadata             # owner/date/block count, no body
agent-lark docs outline "<URL>"                    # heading tree only
agent-lark docs search "keyword" [--exists]        # --exists outputs "found N" or "not_found"
```

> Use `docs outline` first on large documents, then `docs get --section` to avoid loading the full body into context.

### Wiki

```bash
agent-lark wiki list "<URL>" [--depth 2]   # limit tree depth for large wikis
agent-lark wiki get "<URL>"
agent-lark wiki create "<URL>" --title "Title"
```

### Bitable (Multi-dimensional tables)

```bash
agent-lark base fields "<URL>"                             # inspect field schema before writing
agent-lark base records count "<URL>" [--filter '...']     # count only, no record data
agent-lark base records list "<URL>" \
  --select "Name,Status,Due" \                             # always select columns for large tables
  --filter '{"field_name":"Status","operator":"is","value":["Todo"]}' \
  [--limit 100 | --all]
agent-lark base records get "<URL>" <record_id>
```

> Always call `base fields` before `records create` or `records update` to verify field names and types.

### Messages & Contacts

```bash
agent-lark im messages list --chat-id <chat_id> [--limit 20]
agent-lark contact search "Alice"                    # returns candidate list
agent-lark contact resolve "alice@company.com"       # returns single open_id; errors on ambiguity
agent-lark im chats list [--limit 20]
agent-lark im chats search "weekly sync"
```

`im messages get <message_id>` is available for fetching a single message detail.

---

## Writing Content

### Documents

```bash
agent-lark docs create --title "My Document" [--folder "<folder_url>"]
agent-lark docs update "<URL>" --content "New paragraph text"
echo "## Section\nBody." | agent-lark docs update "<URL>" --stdin
```

### Tables in documents

```bash
# Insert after a paragraph containing the given text
agent-lark docs table "<URL>" --after "Q3 Summary" --rows 4 --cols 3 --headers

# From CSV file
agent-lark docs table "<URL>" --after "Data Overview" --data ./report.csv

# --dry-run to preview before writing (recommended for irreversible inserts)
agent-lark docs table "<URL>" --after "Summary" --rows 3 --cols 3 --dry-run
```

If `--after` matches multiple paragraphs: re-run with `--format json` to get structured `matches`, pick the right `index`, then add `--match-index N`.

### Messages

```bash
agent-lark im send --chat-id <chat_id> --text "Hello"
agent-lark im send --user-id <open_id> --text "Hi"       # auto-creates DM
agent-lark im send --chat-id <chat_id> --card-file ./card.json
agent-lark im react add --message-id <msg_id> --emoji THUMBSUP
agent-lark im react remove --message-id <msg_id> --reaction-id <reaction_id>
```

Common emoji keys: `THUMBSUP` `THUMBSDOWN` `FIRE` `CLAP` `EYES` `ROCKET` `HEART`

### Bitable records

```bash
agent-lark base records create "<URL>" --fields '{"Name":"Alice","Status":"Todo"}'
agent-lark base records update "<URL>" <record_id> --fields '{"Status":"Done"}'
agent-lark base records batch-create "<URL>" --file ./records.json
```

### Tasks

```bash
agent-lark task list [--status todo|done]
agent-lark task create --title "Review PR" --due "2024-12-31" [--assignee <open_id>]
agent-lark task update --task-id <id> --status done
```

---

## Templates

Templates are stored locally at `~/.agent-lark/templates/`. Placeholders like `{{date}}` are replaced at apply time.

**Built-in variables:** `{{date}}` `{{datetime}}` `{{author}}` `{{title}}` `{{week}}`
**Custom variables:** `--var key=value` (repeatable)

```bash
agent-lark template list
agent-lark template save weekly-report --file ./weekly.md
agent-lark template save meeting-notes --from "<feishu_doc_url>"
agent-lark template vars weekly-report [--format json]    # inspect before apply
agent-lark template apply weekly-report --new --title "2024-W48 Weekly Report" \
  --var "team=Engineering"
agent-lark template apply meeting-notes --to "<URL>" --after "Agenda" [--dry-run]
agent-lark template delete weekly-report
```

> Always run `template vars` before `template apply`. Variables with `resolved: null` will be written as literal `{{name}}` in the document.

---

## Comments

```bash
agent-lark comments list "<URL>"                          # outputs #N index per comment
agent-lark comments add "<URL>" --content "Looks good!"
agent-lark comments reply "<URL>" --to "#1" --content "Agreed."
agent-lark comments resolve "<URL>" --to "#1"

# Batch: pipe URL list to comments add
agent-lark docs search "sprint-" --format json | \
  jq -r '.[].url' | \
  agent-lark comments add --content "Archived." --batch --agent
```

> `comments unresolve` and `comments delete` are **not yet implemented**.

---

## Permissions

```bash
agent-lark perms list "<URL>"
agent-lark perms check "<URL>" --user alice@company.com    # outputs role or "no_access"
agent-lark perms add "<URL>" --user alice@company.com --role edit [--no-notify]
agent-lark perms update "<URL>" --user alice@company.com --role view [--yes]
agent-lark perms remove "<URL>" --user alice@company.com [--yes]
agent-lark perms transfer "<URL>" --to alice@company.com [--yes]   # irreversible
agent-lark perms public "<URL>"                                     # view current settings
agent-lark perms public "<URL>" --link-share tenant --comment deny [--yes]
```

Permission roles: `view` `edit` `full_access`

> Use `--agent` (preferred) or `--yes` for all `perms update/remove/transfer/public` in automated workflows — without it the command waits for stdin and hangs.

---

## Workflow Patterns

### Create a document from a template

```bash
agent-lark template save okr-review --from "https://company.feishu.cn/docx/doxcnTEMPLATE"
agent-lark template apply okr-review --new --title "2024 Q4 OKR Review" \
  --var "quarter=Q4" --var "author=Alice"
```

### Weekly report automation

```bash
TITLE="$(date +%Y)-W$(date +%V) Weekly Report"
agent-lark template apply weekly-report \
  --new --title "$TITLE" \
  --folder "https://company.feishu.cn/drive/folder/fldcnWEEKLY"
```

### Read → Edit → Verify

```bash
URL="https://company.feishu.cn/docx/doxcnXXXXXX"
agent-lark docs get "$URL"
agent-lark docs update "$URL" --content "Updated section content"
agent-lark docs get "$URL" | grep "Updated section"
```

### Summarize and comment

```bash
URL="https://company.feishu.cn/docx/doxcnXXXXXX"
CONTENT=$(agent-lark docs get "$URL")
# ... summarize $CONTENT with Claude ...
agent-lark comments add "$URL" --content "$SUMMARY"
```

### Search and batch comment

```bash
agent-lark docs search "sprint notes" --format json | \
  jq -r '.[].url' | \
  agent-lark comments add --content "Archived." --batch --agent
```

### Create document and share with team

```bash
URL=$(agent-lark docs create --title "Q4 OKR Review" --format json | jq -r '.url')
agent-lark perms add "$URL" --user alice@company.com --role edit
agent-lark perms add "$URL" --user bob@company.com --role view
agent-lark perms public "$URL" --link-share tenant --yes
```

### Lock down a document after publishing

```bash
URL="https://company.feishu.cn/docx/doxcnXXXXXX"
agent-lark perms list "$URL" --format json | \
  jq -r '.[] | select(.role=="edit") | .member_id' | \
  while read uid; do
    agent-lark perms update "$URL" --user "$uid" --role view --yes
  done
agent-lark perms public "$URL" --link-share off --copy deny --yes
```

### Append a section to multiple documents

```bash
agent-lark docs search "sprint-" --format json | \
  jq -r '.[].url' | \
  while read url; do
    agent-lark template apply retro --to "$url" --after "Sprint End"
  done
```

---

## Error Handling

All commands exit `0` on success, non-zero on failure.

| Code | Meaning | Fix |
|------|---------|-----|
| `AUTH_REQUIRED` | No credentials | Run `agent-lark login` or `agent-lark setup` |
| `TOKEN_EXPIRED` | User token expired | Run `agent-lark auth oauth` |
| `PERMISSION_DENIED` | Wrong token mode or missing scope | Try `--token-mode user`; check dev console permissions |
| `NOT_FOUND` | URL/ID missing or inaccessible | Verify URL, check sharing settings |
| `RATE_LIMITED` | Too many requests | Wait a few seconds and retry |
| `AMBIGUOUS_MATCH` | `--after` keyword matched multiple paragraphs | Add `--match-index N`; use `--format json` for structured `matches` array |
| `TEMPLATE_NOT_FOUND` | Template name not found locally | Run `agent-lark template list` |
| `TEMPLATE_EXISTS` | Template name already taken | Use `--force` to overwrite |
| `TEMPLATE_VAR_MISSING` | `{{var}}` placeholder has no value | Run `template vars <name> --format json`, then pass `--var key=value` |
| `MEMBER_NOT_FOUND` | User not found by email/open_id | Verify with `contact search` first |
| `CANNOT_REMOVE_OWNER` | Tried to remove document owner | Use `perms transfer` first, then remove |

The CLI always includes a **suggested fix** in the error message — read it before retrying.

---

## Output Formats

| Flag | Output | Use for |
|------|--------|---------|
| `--format text` | Human-readable (default) | Quick inspection |
| `--format json` | Machine-readable | Pipelines with `jq` |
| `--format md` | Markdown | Feeding into documents or Claude context |
| `--agent` | JSON + no prompts + `next_cursor` | **All automated workflows** |
