# agent-lark

面向 AI Agent 的飞书 / Lark 命令行工具。覆盖文档、知识库、多维表格、消息、任务、权限等核心场景。

[English](README.md)

## 安装

**环境要求：** Go 1.22+，飞书自建应用（App ID + App Secret）

```bash
git clone https://github.com/wsafight/agent-lark.git
cd agent-lark
go build -o agent-lark ./cmd/agent-lark
mv agent-lark /usr/local/bin/   # 可选
```

## 快速开始

```bash
agent-lark init                 # 首次使用：安装 Skill + 配置凭据
agent-lark setup                # 配置应用凭据（交互式向导）
agent-lark auth oauth           # 追加 OAuth 用户授权（访问个人云盘等场景）
agent-lark auth status          # 查看认证状态

agent-lark docs list
agent-lark docs get "<文档URL>"
agent-lark im send --chat-id <id> --text "你好"
agent-lark task list
```

## 认证配置

### 应用凭据

```bash
agent-lark setup                                        # 交互式向导（自建应用）
agent-lark auth login                                   # 快速登录（内置公共应用，仅特定构建可用）
```

### OAuth 用户授权

代表用户操作（个人云盘、个人任务等）时需要用户 Token：

```bash
agent-lark auth oauth                                   # 浏览器授权，自动保存 Token
agent-lark auth oauth --scope "docx:readonly drive:readonly" --port 9999
```

### Token 模式

| 模式 | 行为 |
|------|------|
| `auto`（默认） | 有用户 Token 时优先使用，否则降级为租户 Token |
| `tenant` | 始终使用应用机器人身份（tenant_access_token） |
| `user` | 始终使用 OAuth 用户身份（user_access_token） |

```bash
agent-lark auth set-mode user
```

### 多 Profile 管理

适用于同时管理多个飞书应用或团队的场景。每个项目目录会根据路径自动分配隔离的 profile（基于 SHA-256 哈希）。

```bash
agent-lark setup --profile work       # 为指定 profile 配置凭据
agent-lark auth profile list          # 列出所有 profile
agent-lark --profile work docs list   # 临时指定 profile
```

### 其他认证命令

```bash
agent-lark auth status               # 查看当前认证状态
agent-lark auth logout --user        # 仅清除用户 Token
agent-lark auth logout --all         # 清除全部凭据
```

## 全局参数

所有参数均可通过环境变量设置，显式传入的参数优先级更高。

| 参数 | 环境变量 | 默认值 | 说明 |
|------|----------|--------|------|
| `--format text\|json\|md` | `AGENT_LARK_FORMAT` | `text` | 输出格式 |
| `--token-mode auto\|tenant\|user` | `AGENT_LARK_TOKEN_MODE` | `auto` | Token 模式 |
| `--profile <name>` | `AGENT_LARK_PROFILE` | — | 指定凭据 profile |
| `--config <path>` | `AGENT_LARK_CONFIG` | — | 显式指定凭据文件路径 |
| `--domain <domain>` | `AGENT_LARK_DOMAIN` | — | 覆盖 API 域名（Lark 国际版填 `open.larksuite.com`） |
| `--agent` | `AGENT_LARK_AGENT` | — | Agent 模式：强制 JSON 输出 + 结构化错误 + 自动确认 |
| `--quiet` | `AGENT_LARK_QUIET` | — | 静默模式，仅输出数据，不打印提示信息 |
| `--debug` | `AGENT_LARK_DEBUG` | — | 打印原始 HTTP 请求/响应 |
| `--yes` | — | — | 自动确认所有交互提示 |

## 命令参考

### auth

```bash
agent-lark setup                                        # 交互式凭据向导
agent-lark auth login                                   # 快速登录（内置公共应用）
agent-lark auth oauth                                   # 追加 OAuth 用户授权
agent-lark auth status                                  # 查看认证状态
agent-lark auth set-mode auto|tenant|user               # 修改默认 Token 模式
agent-lark auth profile list                            # 列出所有 profile
agent-lark auth logout --user                           # 清除用户 Token
agent-lark auth logout --all                            # 全部重置
```

### docs — 文档

```bash
agent-lark docs list                                    # 列举云盘根目录文件
agent-lark docs list --folder "<url>"                   # 列举指定文件夹
agent-lark docs list --since 7d                         # 按修改时间过滤（24h / 7d / today / YYYY-MM-DD）
agent-lark docs list --all                              # 自动翻页获取全部
agent-lark docs list --limit 50

agent-lark docs get "<url>"                             # 读取文档内容（默认 Markdown）
agent-lark docs get "<url>" --format json
agent-lark docs get "<url>" --section "需求背景"         # 只返回匹配的章节
agent-lark docs get "<url>" --metadata                  # 只返回元信息
agent-lark docs get "<url>" --content-boundaries        # 用 <document> 标签包裹输出（防止提示注入）
agent-lark docs get "<url>" --max-chars 8000            # 截断输出，最多 N 个字符

agent-lark docs outline "<url>"                         # 返回文档标题大纲

agent-lark docs search "关键词"
agent-lark docs search "关键词" --limit 10
agent-lark docs search "关键词" --exists                 # 存在时输出 found，否则 not_found

agent-lark docs create --title "周报 W26"
agent-lark docs create --title "设计稿" --folder "<url>"

agent-lark docs update "<url>" --content "追加一段内容"
agent-lark docs update "<url>" --file content.md
agent-lark docs update "<url>" --stdin < content.md

agent-lark docs table "<url>" --rows 3 --cols 4
agent-lark docs table "<url>" --after "总结" --data '[["A","B"],["1","2"]]'
agent-lark docs table "<url>" --file data.csv --headers
agent-lark docs table "<url>" --dry-run                 # 预览，不实际写入
```

### wiki — 知识库

```bash
agent-lark wiki list "<url>"                            # 列举知识空间节点
agent-lark wiki list "<url>" --depth 2                  # 最多展开 2 层
agent-lark wiki list "<url>" --all                      # 自动翻页

agent-lark wiki get "<url>"                             # 读取知识库页面内容
agent-lark wiki get "<url>" --format json
agent-lark wiki get "<url>" --content-boundaries        # 用 <document> 标签包裹输出
agent-lark wiki get "<url>" --max-chars 8000            # 截断输出，最多 N 个字符

agent-lark wiki create "<space-url>" --title "新页面"
```

### base — 多维表格

URL 参数传入飞书多维表格完整页面地址，`appToken` 和 `tableId` 自动解析。

```bash
agent-lark base fields "<url>"                          # 查看字段结构

agent-lark base records list "<url>"
agent-lark base records list "<url>" --limit 50 --all
agent-lark base records list "<url>" --filter 'CurrentValue.[状态]="进行中"'
agent-lark base records list "<url>" --select "姓名,状态,截止日期"
agent-lark base records list "<url>" --cursor "<token>"

agent-lark base records get "<url>" <record_id>

agent-lark base records count "<url>"
agent-lark base records count "<url>" --filter 'CurrentValue.[状态]="完成"'

agent-lark base records create "<url>" --field "姓名=张三" --field "年龄=25"
agent-lark base records create "<url>" --fields '{"姓名":"李四","状态":"进行中"}'

agent-lark base records update "<url>" <record_id> --field "状态=完成"

agent-lark base records batch-create "<url>" --file records.json
# records.json 格式：[{"姓名":"张三","年龄":25}, {"姓名":"李四","年龄":30}]
```

### im — 即时消息

```bash
agent-lark im send --chat-id <id> --text "你好"
agent-lark im send --user-id <open_id> --text "单聊"   # 自动创建单聊
agent-lark im send --chat-id <id> --card '<json>'
agent-lark im send --chat-id <id> --card-file card.json

agent-lark im chats list
agent-lark im chats list --limit 50 --all
agent-lark im chats search "关键词"

agent-lark im messages list --chat-id <id>
agent-lark im messages list --chat-id <id> --limit 20
agent-lark im messages get --message-id <id>

agent-lark im react add --message-id <id> --emoji THUMBSUP
agent-lark im react remove --message-id <id> --emoji THUMBSUP
```

### task — 任务

```bash
agent-lark task list
agent-lark task list --status todo|done
agent-lark task list --assignee <open_id>
agent-lark task list --limit 30 --cursor "<token>"

agent-lark task get --task-id <guid>

agent-lark task create --title "完成需求评审"
agent-lark task create --title "上线" --due 2025-06-30
agent-lark task create --title "跟进" --due 2025-06-30T18:00:00+08:00 --assignee <open_id>

agent-lark task update --task-id <guid> --title "新标题"
agent-lark task update --task-id <guid> --status done
agent-lark task update --task-id <guid> --due 2025-07-01
```

### comments — 文档评论

```bash
agent-lark comments list "<url>"
agent-lark comments list "<url>" --format json

agent-lark comments add "<url>" --content "这里需要修改"
agent-lark comments add "<url>" --content "看这里" --quote '<选区JSON>'

agent-lark comments reply "<url>" --comment-id <id> --content "已修改"

agent-lark comments resolve "<url>" --comment-id <id>
```

### perms — 权限

```bash
agent-lark perms list "<url>"                           # 查看协作者列表
agent-lark perms check "<url>"                          # 检查当前用户权限

agent-lark perms add "<url>" --user <open_id> --role view|edit|full_access
agent-lark perms add "<url>" --user <id1> --user <id2> --role edit
agent-lark perms add "<url>" --user <open_id> --role edit --no-notify

agent-lark perms update "<url>" --user <open_id> --role view

agent-lark perms remove "<url>" --user <open_id>

agent-lark perms transfer "<url>" --user <open_id>      # 转让所有权

agent-lark perms public "<url>"                         # 查看公开访问设置
agent-lark perms public "<url>" --link-share anyone_readable
```

### contact — 通讯录

```bash
agent-lark contact search user@example.com              # 通过邮箱查询用户

agent-lark contact list
agent-lark contact list --dept "产品部"                  # 按部门过滤
agent-lark contact list --limit 100 --all

agent-lark contact resolve --email user@example.com     # 解析用户标识
```

### template — 本地模板

模板以 Markdown 格式存储在本地，支持 `{{变量}}` 占位符替换。

**内置变量**

| 变量 | 值 |
|------|----|
| `{{date}}` | 当前日期（`2025-06-30`） |
| `{{datetime}}` | 当前日期时间（`2025-06-30 14:30`） |
| `{{week}}` | ISO 周数（`W26`） |
| `{{author}}` | 当前登录用户姓名 |
| `{{title}}` | 文档标题（来自 `--title` 参数） |

```bash
agent-lark template list

agent-lark template save 周报 --file weekly.md
agent-lark template save 周报 --file weekly.md --force   # 覆盖已有模板
agent-lark template save 周报 --from "<url>"             # 从飞书文档拉取

agent-lark template get 周报
agent-lark template vars 周报                            # 列出模板中使用的变量

agent-lark template apply 周报 --new --title "2025-W26 周报"
agent-lark template apply 周报 --new --title "周报" --folder "<url>"
agent-lark template apply 周报 --to "<url>"              # 追加到已有文档
agent-lark template apply 周报 --to "<url>" --after "上周工作"
agent-lark template apply 周报 --to "<url>" --before "下周计划"
agent-lark template apply 周报 --new --title "周报" --var "project=飞船" --var "owner=张三"
agent-lark template apply 周报 --new --title "周报" --dry-run   # 预览，不写入

agent-lark template delete 周报
```

### doctor — 诊断

```bash
agent-lark doctor    # 检查配置文件、凭据有效性、API 可达性、Token 过期状态
```

## Agent 模式

`--agent` 专为在 AI Agent 流水线中使用而设计：

- 强制所有输出使用 `--format json`
- 错误写入 stderr，格式为结构化 JSON：`{"error": "CODE", "message": "...", "hint": "..."}`
- 列表命令返回 `{"items": [...], "next_cursor": "..."}` 支持无状态分页续读
- 自动跳过所有交互式确认（隐含 `--yes`）

```bash
# 基本用法
agent-lark --agent docs list --folder "<url>"

# 分页读取 — 将 next_cursor 传回直到为空
agent-lark --agent docs list --limit 20
agent-lark --agent docs list --limit 20 --cursor "<上次返回的 next_cursor>"

# 错误输出示例（stderr）
{"error": "AUTH_REQUIRED", "message": "credentials not configured", "hint": "run: agent-lark auth login"}
```

## 错误码

| 错误码 | 含义 | 建议操作 |
|--------|------|----------|
| `AUTH_REQUIRED` | 未配置凭据 | 运行 `agent-lark auth login` |
| `TOKEN_EXPIRED` | OAuth 用户 Token 已过期 | 运行 `agent-lark auth oauth` |
| `MISSING_FLAG` | 缺少必填参数 | 查看 `--help` |
| `MISSING_ARG` | 缺少位置参数 | 查看 `--help` |
| `INVALID_URL` | 无法解析飞书 URL | 使用完整的飞书资源页面地址 |
| `INVALID_JSON` | JSON 格式错误 | 检查 `--fields` 或 `--data` 参数 |
| `INVALID_STATUS` | 无效的状态值 | 参考 `--help` 中的可选值 |
| `API_ERROR` | 飞书 API 返回错误 | 检查应用权限及请求参数 |
| `CLIENT_ERROR` | 客户端初始化失败 | 运行 `agent-lark doctor` 诊断 |
| `FILE_ERROR` | 文件读取失败 | 确认文件路径和读取权限 |
| `NOT_FOUND` | 资源不存在 | 检查 URL 或搜索关键词 |
