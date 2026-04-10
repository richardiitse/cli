# Architecture Overview

<!-- Generated: 2026-04-10 | Files scanned: 520 | Token estimate: ~600 -->

## Project Type

**Lark CLI Tool** - Command-line interface for Feishu/Lark Open Platform APIs

- **Language**: Go 1.23+
- **Architecture**: Three-layer command system
- **Scope**: 12 business domains, 200+ commands, 20 AI Agent Skills

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        User/AI Agent                        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────────┐
│                     Root Command (cmd/root.go)              │
│  - Global flags management                                  │
│  - Command routing                                           │
│  - Profile/Config initialization                             │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         ↓                       ↓
┌──────────────────┐    ┌──────────────────────────────────┐
│ Built-in Commands│    │     Shortcuts Framework         │
│ (cmd/*)          │    │  (shortcuts/*)                   │
├──────────────────┤    ├──────────────────────────────────┤
│ auth             │    │ calendar +agenda                │
│ config           │    │ im +messages-send               │
│ doctor           │    │ doc +create                     │
│ profile          │    │ event +subscribe (WebSocket)    │
│ schema           │    │ ... (200+ shortcuts)             │
│ api              │    │                                  │
│ bot              │    │ - Human-friendly shortcuts       │
│                  │    │ - AI-optimized parameters        │
│                  │    │ - Dry-run previews              │
└──────────────────┘    └──────────────────────────────────┘
         │                         │
         └───────────┬───────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│                   Internal Layers                          │
├─────────────────────────────────────────────────────────────┤
│ internal/auth/      - OAuth, token management             │
│ internal/client/    - Lark SDK wrapper                    │
│ internal/core/      - Config, endpoints, runtime          │
│ internal/cmdutil/   - Factory, helpers                     │
│ internal/output/    - JSON, table, pretty formatting       │
│ internal/registry/  - API metadata registry                │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         ↓                       ↓
┌──────────────────┐    ┌──────────────────────────────────┐
│  Lark SDK        │    │   Extension System              │
│  (oapi-sdk-go)   │    │  (extension/*)                   │
├──────────────────┤    ├──────────────────────────────────┤
│ - API calls      │    │ - Credential interface           │
│ - WebSocket      │    │ - File I/O abstraction           │
│ - Auth handling  │    │ - Transport abstraction          │
└──────────────────┘    └──────────────────────────────────┘
         │
         ↓
┌─────────────────────────────────────────────────────────────┐
│              Feishu/Lark Open Platform APIs               │
│  - Messenger, Docs, Base, Sheets, Calendar, Mail, etc.   │
└─────────────────────────────────────────────────────────────┘
```

---

## Three-Layer Command System

### Layer 1: Shortcuts (AI/Human Friendly)
- **Format**: `lark-cli <service> +<verb> [flags]`
- **Examples**: `calendar +agenda`, `im +messages-send`
- **Features**: Smart defaults, table output, dry-run

### Layer 2: API Commands (Platform-Synced)
- **Format**: `lark-cli <service> <resource> <method> [flags]`
- **Examples**: `calendar events instance_view`
- **Source**: Auto-generated from Lark OAPI metadata

### Layer 3: Raw API (Full Coverage)
- **Format**: `lark-cli api <method> <path> [--params] [--data]`
- **Examples**: `api GET /open-apis/calendar/v4/calendars`
- **Coverage**: 2500+ API endpoints

---

## Key Entry Points

| File | Purpose |
|------|---------|
| `main.go` | Go build entry point |
| `cmd/root.go` | Cobra root command, CLI bootstrap |
| `cmd/bootstrap.go` | Initialization sequence |
| `shortcuts/register.go` | Shortcut registration hub |

---

## Module Boundaries

### Commands Layer (`cmd/`)
- **Responsibility**: CLI interface, command parsing, user interaction
- **Dependencies**: `internal/` packages, `shortcuts/`
- **Size**: 57 files, ~3000 LOC

### Shortcuts Layer (`shortcuts/`)
- **Responsibility**: Business logic, API orchestration, human-friendly UX
- **Dependencies**: `internal/`, Lark SDK
- **Size**: 322 files, ~15000 LOC
- **Domains**: 12 business domains (calendar, im, doc, etc.)

### Internal Layer (`internal/`)
- **Responsibility**: Core utilities, shared infrastructure
- **Size**: 141 files, ~8000 LOC
- **Key Packages**:
  - `auth/` - OAuth flows, token storage (keychain)
  - `client/` - Lark SDK client factory
  - `core/` - Config loading, endpoint resolution
  - `cmdutil/` - Factory pattern, helpers
  - `output/` - Multi-format output (JSON/table/pretty)

### Extension Layer (`extension/`)
- **Responsibility**: Pluggable interfaces for credentials, file I/O, transport
- **Size**: 5 packages
- **Interfaces**:
  - `credential.CredentialProvider` - Token storage abstraction
  - `fileio.FileHandler` - File upload/download
  - `transport.Transport` - HTTP client abstraction

---

## Data Flow

### Typical Command Execution

```
User input: "lark-cli calendar +agenda"
    ↓
cmd/root.go: Parse arguments, load config
    ↓
cmdutil.Factory: Initialize runtime context
    ↓
shortcuts/calendar/agenda.go: Execute shortcut
    ↓
internal/client/: Get Lark SDK client
    ↓
internal/auth/: Get access token
    ↓
Lark SDK: API call to /open-apis/calendar/v4/calendar_events/list
    ↓
internal/output/: Format response as table
    ↓
User sees: Agenda table
```

### Event Subscription Flow (WebSocket)

```
User: "lark-cli event +subscribe --event-types im.message.receive_v1"
    ↓
shortcuts/event/subscribe.go: Establish WebSocket connection
    ↓
shortcuts/event/pipeline.go: Process events
    ↓
Event Filter → Dedup → Transform → Output
    ↓
Output: NDJSON stream to stdout
    ↓
User can pipe to other tools (e.g., bot handler)
```

---

## AI Agent Integration

### Skills System (`skills/`)
- **20 AI Agent Skills** - Teach LLMs how to use lark-cli
- **Format**: Structured SKILL.md files
- **Installation**: `npx skills add larksuite/cli -g -y`

### Key Skills
- `lark-shared` - Auth, config, scope management (auto-loaded)
- `lark-calendar` - Calendar operations
- `lark-im` - Messaging, chat management
- `lark-event` - WebSocket subscriptions
- ... (17 more)

---

## New: Bot Integration (Feature Branch)

### Location: `cmd/bot/`
- **Purpose**: Claude Code Bot - "Feishu → Claude Code" integration
- **Branch**: `feature/claude-code-bot`
- **Status**: Framework implemented, core logic pending

### Architecture
```
Feishu message
    ↓
lark-cli event +subscribe (WebSocket)
    ↓
bot/handler.go (message processing)
    ↓
bot/router.go (command routing)
    ↓
bot/claude.go (Claude Code CLI integration)
    ↓
bot/session.go (session_id persistence)
    ↓
lark-cli im +messages-send (reply)
```

### Key Files
- `cmd/bot/bot.go` - Bot command entry
- `cmd/bot/start.go` - Start bot (TODO: implement)
- `cmd/bot/status.go` - Check status (TODO: implement)
- `cmd/bot/stop.go` - Stop bot (TODO: implement)

---

## Security Layers

1. **Input Sanitization**: All user input validated, injection protected
2. **Token Storage**: OS-native keychain (Keychain on macOS, wincred on Windows)
3. **Scope Management**: User can limit granted permissions
4. **Risk Levels**: Commands marked as "read"/"write"/"high-risk-write"
5. **Dry Run Mode**: Preview requests without execution

---

## Configuration

### Location: `~/.lark-cli/`
- `config.json` - Multi-app configuration (app_id, app_secret, brand)
- `profiles/` - Named profiles (dev, staging, prod)
- `*.keychain` - Encrypted tokens (OS keychain)

### Environment Variables
- `LARK_CLI_PROFILE` - Active profile
- `LARK_CLI_CONFIG_DIR` - Custom config directory
