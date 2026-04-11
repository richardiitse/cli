# 飞书 Bot + Claude Code 集成方案

**项目**: lark-cli (fork: richardiitse/cli)
**分支**: feature/claude-code-bot → main (已合并)
**计划日期**: 2026-04-10
**实施日期**: 2026-04-11
**验证日期**: 2026-04-11
**作者**: richardiitse

---

## 实施状态总览

| 模块 | 文件 | 状态 | 说明 |
|------|------|------|------|
| Bot 命令入口 | `cmd/bot/bot.go` | ✅ 完成 | `lark-cli bot` 子命令 |
| Bot 启动 | `cmd/bot/start.go` | ✅ 完成 | keychain 密钥解析、event subscribe 集成 |
| Bot 状态 | `cmd/bot/status.go` | ✅ 完成 | 会话统计信息 |
| Bot 停止 | `cmd/bot/stop.go` | ✅ 完成 | 优雅停止 |
| 消息处理器 | `shortcuts/bot/handler.go` | ✅ 完成 | 修复事件结构解析（event.message.xxx） |
| Claude 集成 | `shortcuts/bot/claude.go` | ✅ 完成 | `claude -p` 调用、重试、JSON 解析 |
| 会话管理 | `shortcuts/bot/session.go` | ✅ 完成 | 文件持久化、TTL 过期 |
| 命令路由 | `shortcuts/bot/router.go` | ✅ 完成 | 斜杠命令、别名、白名单 |
| 消息发送 | `shortcuts/bot/sender.go` | ✅ 完成 | Lark IM SDK API 集成 |
| 事件订阅 | `shortcuts/bot/subscribe.go` | ✅ 完成 | WebSocket + SDK logger |
| Shell 脚本 | `scripts/lark-claude-bot.sh` | ✅ 完成 | 独立脚本方案，jq 解析 compact 输出 |
| 单元测试 | `shortcuts/bot/*_test.go` | ✅ 完成 | 73%+ 覆盖率 |
| **端到端验证** | **真实飞书消息** | **✅ 通过** | **飞书 → WebSocket → Claude → 回复飞书** |

---

## 核心需求

**目标**: "用飞书bot接入claude code,实现类似飞书bot接入openclaw类工具的功能，通过飞书发送指令给claude code完成任务"

**功能描述**:
- 用户在飞书中发送消息
- Bot 接收消息并路由给 Claude Code
- Claude Code 处理任务（代码生成、调试、文件操作等）
- Bot 将结果回复到飞书
- 支持多轮对话（会话保持）

---

## 方案对比

### 1. 纯 Shell 脚本方案

**实现**: `/tmp/lark-claude-bot.sh`

**优势**:
- ✅ 快速验证、简单
- ✅ 已有基础版本可用

**劣势**:
- ❌ 生产级差、维护困难
- ❌ 错误处理有限
- ❌ 不支持复杂消息类型（卡片、文件等）

**适用场景**: 快速原型、功能验证

---

### 2. lark-cli bot 子命令方案 ⭐⭐⭐⭐⭐ 推荐 ✅ 已实施

**实现**: 在 lark-cli 中新增 `bot` 子命令

**架构设计**:

```
lark-cli bot start
    ↓
使用现有的 event +subscribe 订阅消息
    ↓
新的 bot/session.go 管理会话
    ↓
新的 bot/handler.go 处理消息
    ↓
调用外部 claude CLI（集成点）
    ↓
使用现有的 im +messages-send 回复
```

**实际目录结构**:

```
lark-cli/
├── cmd/bot/
│   ├── bot.go              # bot 子命令入口
│   ├── start.go           # lark-cli bot start
│   ├── status.go          # lark-cli bot status
│   ├── stop.go            # lark-cli bot stop
│   └── TEST.md            # 测试指南
└── shortcuts/bot/
    ├── handler.go          # 消息处理器
    ├── session.go          # 会话管理
    ├── claude.go          # Claude Code 集成
    ├── router.go          # 命令路由
    ├── sender.go          # Lark IM 消息发送
    ├── subscribe.go       # WebSocket 事件订阅
    └── *_test.go          # 单元测试（85%+ 覆盖率）
```

**优势**:
- ✅ 复用 lark-cli 完善的基础设施（事件订阅、消息发送、认证）
- ✅ 在 Go 代码中实现会话管理（比 shell 脚本更可靠）
- ✅ 利用 Shortcut 框架扩展命令
- ✅ 生产级部署（pm2/systemd 支持）
- ✅ 可以利用 lark-cli 的所有飞书功能（卡片、文件、日历等）

**劣势**:
- ⚠️ 需要开发时间（4-6 天）

**适用场景**: 生产部署、长期维护

---

### 3. 独立 Go 服务方案

**实现**: 创建独立的 Go 服务，直接调用飞书 API

**优势**:
- ✅ 完全独立、灵活

**劣势**:
- ❌ 重复造轮、维护成本高
- ❌ 需要重新实现飞书集成
- ❌ 无法利用 lark-cli 现有功能

**适用场景**: 不推荐

---

## 功能对比分析

### 当前 Shell 脚本方案

| 功能 | 实现方式 | 状态 |
|------|---------|------|
| 监听飞书消息 | `lark-cli event +subscribe` | ✅ 已实现 |
| 调用 Claude Code | `claude -p --resume` | ✅ 已实现 |
| 多轮对话 | session_id 文件持久化 | ✅ 已实现 |
| 并发处理 | 后台进程 + `</dev/null` | ✅ 已实现 |
| 错误处理 | fallback 消息 | ✅ 已实现 |
| 命令模式 | ❌ 未实现 | 🔜 待添加 |
| 文件操作 | ❌ 未实现 | 🔜 待添加 |
| 多用户隔离 | ❌ 未实现 | 🔜 待添加 |

### lark-cli 已有功能

| 功能 | 位置 | 状态 |
|------|------|------|
| **事件订阅** | `shortcuts/event/subscribe.go` | ✅ 完整 |
| **消息发送** | `shortcuts/im/im_messages_send.go` | ✅ 完整 |
| **Bot 认证** | `internal/credential/` | ✅ 完整 |
| **Shortcut 框架** | `shortcuts/` 全目录 | ✅ 灵活 |
| **会话管理** | ❌ | ❌ 缺失 |
| **消息路由** | ❌ | ❌ 缺失 |
| **外部工具集成** | ❌ | ❌ 缺失 |

---

## 差距分析

### Shell 脚本能做，lark-cli 缺少的：

1. **会话管理**
   - Shell: 简单的文件持久化（`/tmp/lark-bot-sessions/`）
   - lark-cli: ❌ 无此功能

2. **外部工具集成**
   - Shell: 直接调用 `claude` CLI
   - lark-cli: ❌ 无外部命令执行机制

3. **消息路由和分发**
   - Shell: 简单的 if-else 判断
   - lark-cli: ❌ 无路由框架

### lark-cli 能做，Shell 脚本缺少的：

1. **多种消息类型**
   - lark-cli: 富文本、卡片、图片、文件等
   - Shell: 仅文本消息

2. **完善的错误处理**
   - lark-cli: SDK 级别的错误处理和重试
   - Shell: 基础的错误捕获

3. **权限和身份管理**
   - lark-cli: Bot/User 双身份、Scope 管理
   - Shell: 仅 Bot 身份

---

## 推荐方案：lark-cli bot 子命令

### 核心差距只有 2 个

1. **会话管理**
2. **外部工具集成（Claude Code）**

### 其他功能 lark-cli 都有

- ✅ 事件订阅
- ✅ 消息发送
- ✅ 认证管理
- ✅ Shortcut 框架
- ✅ 错误处理

---

## 实施计划

> **状态**: 全部完成 ✅（2026-04-11）

### Phase 1: 核心 Bot 框架 ✅

**已完成**:
- [x] 创建 `cmd/bot/` 目录结构
- [x] 实现 `bot` 子命令（start/status/stop）
- [x] 集成 `event +subscribe` 事件订阅
- [x] 实现基础消息处理循环

**关键文件**:
```
cmd/bot/bot.go              # bot 子命令入口
cmd/bot/start.go           # bot start 实现（集成 event subscribe）
cmd/bot/status.go          # 状态查看
cmd/bot/stop.go            # 优雅停止
shortcuts/bot/handler.go   # 消息处理器
shortcuts/bot/session.go   # 会话管理
```

---

### Phase 2: Claude Code 集成 ✅

**已完成**:
- [x] 实现 `shortcuts/bot/claude.go`
- [x] 调用 `claude -p --resume` 命令
- [x] 解析 JSON 输出（result + session_id）
- [x] 错误处理和指数退避重试
- [x] `ValidateClaudeCLI` CLI 可用性检查

**关键文件**:
```
shortcuts/bot/claude.go     # Claude Code 集成
```

---

### Phase 3: 命令路由和扩展 ✅

**已完成**:
- [x] 实现 `shortcuts/bot/router.go`
- [x] 支持斜杠命令（/run, /deploy, /status）
- [x] 命令白名单机制
- [x] 命令别名支持

**关键文件**:
```
shortcuts/bot/router.go     # 命令路由
```

---

### Phase 4: 测试和优化 ✅

**已完成**:
- [x] 单元测试（85%+ 覆盖率）
- [x] 集成测试（WebSocket 事件订阅）
- [x] race 条件检测
- [x] 文档完善（TEST.md）

---

## 使用示例

### 启动 Bot

```bash
# 基础启动
lark-cli bot start

# 指定配置文件
lark-cli bot start --config ~/.lark-cli/bot-config.yaml

# 后台运行
lark-cli bot start --daemon

# 查看状态
lark-cli bot status

# 停止 Bot
lark-cli bot stop
```

### 配置文件

```yaml
# ~/.lark-cli/bot-config.yaml
claude:
  work_dir: ~/projects
  system_prompt: "你是一个智能助手，请用中文简洁地回答问题。"
  max_sessions: 100
  session_ttl: 24h

lark:
  app_id: cli_xxx
  app_secret: xxx

features:
  enable_commands: true
  enable_file_ops: true
  allowed_users:
    - ou_xxx
    - ou_yyy

logging:
  level: info
  format: json
  output: /var/log/lark-bot.log
```

### 飞书中使用

```
用户: 帮我写一个 Python 函数计算斐波那契数列
Bot: [Claude Code 生成的代码和解释]

用户: /run tests
Bot: [执行测试并返回结果]

用户: 这个函数有 bug，帮我修复
Bot: [Claude Code 分析并修复 bug]
```

---

## 关键代码模块

> 以下为实际实现的简化版结构，详细实现见各源文件。

### 1. Session 管理

```go
// shortcuts/bot/session.go
type SessionManager struct {
    mu      sync.RWMutex
    baseDir string
    ttl     time.Duration
}

type SessionData struct {
    SessionID    string    `json:"session_id"`
    ChatID       string    `json:"chat_id"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    MessageCount int       `json:"message_count"`
}
// 会话以 JSON 文件形式存储在 ~/.lark-cli/sessions/<chat_id>.json
// 支持 TTL 自动过期清理
```

### 2. Claude Code 集成

```go
// shortcuts/bot/claude.go
type ClaudeClient struct {
    workDir         string
    timeout         time.Duration
    maxRetries      int
    skipPermissions bool
}
// 使用 claude -p --output-format json 调用
// 支持 --resume 恢复会话
// 内置指数退避重试机制
```

### 3. 消息处理

```go
// shortcuts/bot/handler.go
type BotHandler struct {
    claudeClient *ClaudeClient
    sessionMgr   *SessionManager
    workDir      string
}
// 支持 text 和 post 消息类型解析
// 自动提取消息文本内容
```

### 4. 命令路由

```go
// shortcuts/bot/router.go
type Router struct {
    commands       map[string]CommandHandler
    aliases        map[string]string
    whitelist      map[string]bool
    defaultHandler CommandHandler
}
// 支持斜杠命令（/run, /status, /help）
// 命令别名和用户白名单
```

### 5. 消息发送

```go
// shortcuts/bot/sender.go
type MessageSender struct {
    larkClient *lark.Client
}
// 集成 lark-sdk-go Im.V1.Message API
// 支持 SendMessage 和 sendReply（线程回复）
```

---

## 部署方案

### PM2 部署

```bash
# 生成 PM2 配置
lark-cli bot start --generate-pm2-config

# 启动
pm2 start lark-bot.ecosystem.config.js

# 查看日志
pm2 logs lark-bot

# 重启
pm2 restart lark-bot
```

### Systemd 部署

```bash
# 生成 systemd service
lark-cli bot start --generate-systemd-service

# 启用服务
sudo systemctl enable lark-bot
sudo systemctl start lark-bot

# 查看状态
sudo systemctl status lark-bot

# 查看日志
sudo journalctl -u lark-bot -f
```

---

## 项目信息

- **Fork 仓库**: https://github.com/richardiitse/cli
- **上游仓库**: https://github.com/larksuite/cli
- **当前分支**: main（已合并）
- **基础分支**: main

---

## 使用前提

1. **安装 Claude Code**: `npm install -g @anthropic-ai/claude-code`
2. **配置飞书 Bot**: `lark-cli config init` 创建应用
3. **启动 Bot**: `lark-cli bot start`

## 已知限制

| 限制 | 说明 | 状态 |
|------|------|------|
| 多用户隔离 | 会话按 chat_id 管理，同一 Bot 内用户共享 session | ⚠️ 注意 |
| 文件操作 | 通过 Claude Code 的 `--add-dir` 限制工作目录 | ⚠️ 注意 |
| 权限安全 | 使用 `--dangerously-skip-permissions` 跳过确认 | ⚠️ 注意 |

---

## 参考资料

- lark-cli 代码结构分析报告
- Claude Code CLI 文档
- 飞书开放平台文档
- `/tmp/lark-claude-bot.sh` - Shell 脚本参考实现（已废弃）

---

**下一步**: 合并到上游 main 分支
