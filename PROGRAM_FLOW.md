# Crush Program Flow - Complete Trace

This document traces the complete execution flow from startup to tool execution in Crush.

## Directory Structure

```
crush/
├── main.go                      # Entry point, .env loading, profiling setup
├── go.mod / go.sum              # Go module dependencies
├── Taskfile.yml                 # Task runner config (build, test, lint)
├── .crush.json / crush.json     # Optional config (LSP, MCP, options)
├── CLAUDE.md                    # AI development guidance
├── PROGRAM_FLOW.md              # This file
├── README.md                    # User-facing documentation
│
├── internal/                    # Private implementation (not importable externally)
│   │
│   ├── cmd/                     # CLI command definitions
│   │   ├── root.go              # Root command, setupApp(), flag parsing
│   │   ├── run.go               # Non-interactive run command
│   │   ├── dirs.go              # Data directory management
│   │   ├── logs.go              # Log viewing command
│   │   ├── providers.go         # Provider management commands
│   │   └── version.go           # Version information
│   │
│   ├── app/                     # Main application orchestration
│   │   └── app.go               # App struct, service initialization, event setup
│   │
│   ├── agent/                   # Agent system (core AI orchestration)
│   │   ├── coordinator.go       # Agent lifecycle, mode selection, tool building
│   │   ├── agent.go             # SessionAgent, turn execution, streaming
│   │   ├── prompts.go           # Prompt loader functions
│   │   ├── templates/           # System prompt templates
│   │   │   ├── coder.md.tpl     # Coder mode system prompt
│   │   │   ├── research.md.tpl  # Research mode system prompt
│   │   │   ├── task.md.tpl      # Task mode system prompt
│   │   │   └── agent.md.tpl     # Meta-agent prompt (for agent tool)
│   │   └── tools/               # Tool implementations
│   │       ├── view.go          # Read files (with UTF-8 check)
│   │       ├── edit.go          # Edit files with exact string replacement
│   │       ├── multiedit.go     # Multi-file batch editing
│   │       ├── write.go         # Create/overwrite files
│   │       ├── bash.go          # Shell command execution
│   │       ├── ls.go            # Directory listing
│   │       ├── glob.go          # Pattern matching file search
│   │       ├── grep.go          # Content search with ripgrep
│   │       ├── journal.go       # Research journal entries (JOURNAL.md)
│   │       ├── fetch.go         # HTTP requests
│   │       ├── download.go      # File downloads
│   │       ├── sourcegraph.go   # Code search integration
│   │       ├── diagnostics.go   # LSP diagnostics
│   │       ├── references.go    # LSP find references
│   │       └── mcp.go           # MCP tool proxy
│   │
│   ├── config/                  # Configuration system
│   │   ├── config.go            # Config struct, agent setup, tool names
│   │   ├── load.go              # Config loading cascade (JSON → env → defaults)
│   │   ├── load_test.go         # Configuration tests
│   │   └── lsp.go               # LSP server configuration
│   │
│   ├── tui/                     # Terminal User Interface (Bubble Tea)
│   │   ├── tui.go               # Main TUI model, Update(), View()
│   │   ├── page/                # Pages (screens)
│   │   │   ├── chat/            # Main chat interface
│   │   │   ├── sessions/        # Session history view
│   │   │   ├── files/           # File history view
│   │   │   ├── config/          # Configuration editor
│   │   │   └── providers/       # Provider setup wizard
│   │   ├── dialog/              # Modal dialogs
│   │   │   ├── permission.go    # Tool permission approval
│   │   │   ├── modelpicker.go   # Model selection
│   │   │   └── confirm.go       # Generic confirmation
│   │   └── component/           # Reusable UI components
│   │       ├── message.go       # Message rendering
│   │       ├── statusbar.go     # Bottom status bar
│   │       └── textarea.go      # Multi-line input
│   │
│   ├── session/                 # Session management service
│   │   └── service.go           # CRUD operations, SQLite backed
│   │
│   ├── message/                 # Message persistence service
│   │   └── service.go           # Store user/assistant messages, tool calls
│   │
│   ├── permission/              # Tool permission system
│   │   └── service.go           # Async permission requests/approvals
│   │
│   ├── history/                 # File version history service
│   │   └── service.go           # Track file changes, restore versions
│   │
│   ├── lsp/                     # Language Server Protocol integration
│   │   ├── client.go            # LSP client wrapper (uses powernap)
│   │   └── manager.go           # Multi-LSP management
│   │
│   ├── mcp/                     # Model Context Protocol integration
│   │   ├── client.go            # MCP client implementation
│   │   └── manager.go           # MCP server lifecycle
│   │
│   ├── db/                      # Database layer
│   │   ├── connect.go           # SQLite connection setup
│   │   ├── migrations/          # Goose migrations
│   │   │   ├── 001_init.sql
│   │   │   ├── 002_messages.sql
│   │   │   ├── 003_files.sql
│   │   │   ├── 004_tool_calls.sql
│   │   │   └── 005_indexes.sql
│   │   └── queries/             # SQL queries (if using sqlc)
│   │
│   ├── pubsub/                  # Event publishing system
│   │   └── pubsub.go            # Generic pub/sub for service events
│   │
│   ├── fsext/                   # Filesystem utilities
│   │   ├── ls.go                # Directory listing with ignore patterns
│   │   ├── fileutil.go          # File operations, glob walking
│   │   └── ignore.go            # .gitignore/.crushignore parsing
│   │
│   ├── filepathext/             # Path manipulation utilities
│   │   └── filepath.go          # Smart path joining, resolution
│   │
│   ├── diff/                    # Diff generation
│   │   └── diff.go              # Unified diff creation for file changes
│   │
│   ├── csync/                   # Concurrency utilities
│   │   └── map.go               # Thread-safe map implementation
│   │
│   ├── event/                   # Application events
│   │   └── event.go             # Event type definitions
│   │
│   └── [other utilities]/       # Additional helper packages
│       ├── analytics/           # Usage analytics
│       ├── clipboard/           # Clipboard operations
│       ├── document/            # Document parsing (PDF, DOCX)
│       ├── telemetry/           # Telemetry collection
│       └── version/             # Version management
│
└── [generated/external]/        # Not in repo but important at runtime
    └── ~/.crush/                # User data directory
        ├── crush.db             # SQLite database
        ├── crush.json           # User config
        └── logs/                # Application logs
```

## Key Directory Purposes

### `internal/agent/`
**The brain of Crush.** Handles all AI interactions:
- Coordinator selects mode (coder/research/task) → loads appropriate prompt template
- SessionAgent executes turns, manages LLM streaming, handles tool calls
- Tool implementations provide capabilities (file operations, shell, search, etc.)

### `internal/config/`
**Configuration cascade:** Loads from `.crush.json` → environment vars → defaults
- Defines available tools, agents, LSP servers, MCP servers
- Manages provider credentials (Anthropic, OpenAI, etc.)
- Controls mode-specific behavior (research vs coder)

### `internal/tui/`
**User interface layer (Bubble Tea).** Completely decoupled from app logic via events:
- Pages: Different screens (chat, sessions, files, config)
- Dialogs: Modal interactions (permissions, model picker)
- Updates in response to events published by services

### `internal/app/`
**Central orchestrator:** Wires everything together
- Initializes all services (session, message, permission, history)
- Sets up event routing (services → TUI)
- Creates agent coordinator
- Manages LSP client lifecycle

### Services (`internal/session/`, `internal/message/`, etc.)
**Persistence and state management:**
- All backed by SQLite via `internal/db/`
- Publish events for UI updates (via `internal/pubsub/`)
- Provide CRUD operations for their domain

### `internal/fsext/`
**File filtering logic:** Where `.gitignore`, `.crushignore`, and common ignore patterns are enforced
- Used by ls/glob tools to filter file listings
- Different behavior in research mode (more lenient)

### `internal/lsp/`
**IDE-like features:** Diagnostics, go-to-definition, find-references
- Integrates via `powernap` library
- Supports multiple language servers (gopls, typescript-language-server, etc.)

### `internal/mcp/`
**External tool integration:** Model Context Protocol for extending capabilities
- Allows adding custom tools/resources without modifying Crush
- Configured in `.crush.json` under `mcp` section

## 1. Application Startup

```
┌─────────────────────────────────────────────────────────────────┐
│ main.go:13-24                                                   │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│ func main()                                                     │
│   ├─ Load .env files (autoload)                                 │
│   ├─ Start pprof server if CRUSH_PROFILE set (localhost:6060)   │
│   └─ cmd.Execute() ──────────────────────────────────────────┐  │
└──────────────────────────────────────────────────────────────┼──┘
                                                               │
                                                               ▼
┌─────────────────────────────────────────────────────────────────┐
│ internal/cmd/root.go:128-151                                    │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│ func Execute()                                                  │
│   ├─ Setup colored heartbit logo for version output             │
│   └─ fang.Execute(rootCmd) with context and signal handling ─┐  │
└──────────────────────────────────────────────────────────────┼──┘
                                                               │
                                                               ▼
```

## 2. Root Command Execution (Interactive Mode)

```
┌───────────────────────────────────────────────────────────────┐
│ internal/cmd/root.go:76-104                                   │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ rootCmd.RunE                                                  │
│   ├─ setupApp(cmd) ───────────────────────────────────────┐   │
│   ├─ event.AppInitialized()                               │   │
│   ├─ ui := tui.New(app)                                   │   │
│   ├─ program := tea.NewProgram(ui)                        │   │
│   ├─ go app.Subscribe(program)  ← Start event forwarding  │   │
│   └─ program.Run()              ← Start Bubble Tea TUI    │   │
└───────────────────────────────────────────────────────────┼───┘
                                                            │
                        ┌───────────────────────────────────┘
                        ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/cmd/root.go:155-197                                  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func setupApp(cmd) → *app.App                                 │
│   ├─ Parse flags: -d (debug), -y (yolo), -D (data-dir), -r (research) │
│   ├─ ResolveCwd()                                             │
│   ├─ config.Init(cwd, dataDir, debug, research) ──────────┐   │
│   ├─ Create .crush directory                              │   │
│   ├─ db.Connect(ctx, dataDir) ← SQLite + migrations       │   │
│   ├─ app.New(ctx, conn, cfg) ─────────────────────────────┼─┐ │
│   └─ event.Init() if metrics enabled                      │ │ │
└───────────────────────────────────────────────────────────┼─┼─┘
                                                            │ │
                        ┌───────────────────────────────────┘ │
                        ▼                                     │
┌───────────────────────────────────────────────────────────┐ │
│ internal/config/load.go                                   │ │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│ │
│ func Init(cwd, dataDir, debug)                            │ │
│   ├─ Find config file (.crush.json, crush.json, ~/.config)│ │
│   ├─ Load LSP configs                                     │ │
│   ├─ Load MCP configs                                     │ │
│   ├─ Load provider configs                                │ │
│   └─ Return *Config                                       │ │
└───────────────────────────────────────────────────────────┘ │
                                                              │
                        ┌─────────────────────────────────────┘
                        ▼
```

## 3. Application Initialization

```
┌───────────────────────────────────────────────────────────────┐
│ internal/app/app.go:53-97                                     │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func New(ctx, conn, cfg) → *App                               │
│   ├─ Create services:                                         │
│   │   ├─ session.Service    (SQLite backed)                   │
│   │   ├─ message.Service    (SQLite backed)                   │
│   │   ├─ history.Service    (File versioning)                 │
│   │   └─ permission.Service (Tool approvals)                  │
│   ├─ setupEvents() ───────────────────────────────────────┐   │
│   ├─ initLSPClients(ctx)                                  │   │
│   └─ InitCoderAgent(ctx) ─────────────────────────────────┼─┐ │
└───────────────────────────────────────────────────────────┼─┼─┘
                                                            │ │
             ┌──────────────────────────────────────────────┘ │
             ▼                                                │
┌──────────────────────────────────────────────────────────┐  │
│ internal/app/app.go:214-230                              │  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│  │
│ func setupEvents()                                       │  │
│   ├─ Create event channel (buffered, cap 100)            │  │
│   ├─ Subscribe to all service events:                    │  │
│   │   ├─ Sessions                                        │  │
│   │   ├─ Messages                                        │  │
│   │   ├─ Permissions                                     │  │
│   │   ├─ History                                         │  │
│   │   ├─ MCP                                             │  │
│   │   └─ LSP                                             │  │
│   └─ All events → app.events chan  → TUI via Subscribe() │  │
└──────────────────────────────────────────────────────────┘  │
                                                              │
             ┌────────────────────────────────────────────────┘
             ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/app/app.go:265-288                                   │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func InitCoderAgent(ctx)                                      │
│   └─ agent.NewCoordinator() ───────────────────────────────┐  │
└────────────────────────────────────────────────────────────┼──┘
                                                             │
             ┌───────────────────────────────────────────────┘
             ▼
```

## 4. Agent Coordinator Creation

```
┌───────────────────────────────────────────────────────────────┐
│ internal/agent/coordinator.go:71-108                          │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func NewCoordinator(cfg, sessions, messages, permissions...)  │
│   ├─ Load agent config (AgentCoder)                           │
│   ├─ coderPrompt() ← Load coder.md.tpl ─────────────────────┐ │
│   ├─ buildAgent(ctx, prompt, agentCfg) ─────────────────────┼─┤
│   └─ Store as currentAgent                                  │ │
└───────────────────────────────────────────────────────────┼─┼─┘
                                                            │ │
             ┌──────────────────────────────────────────────┘ │
             ▼                                                │
┌──────────────────────────────────────────────────────────┐  │
│ internal/agent/prompts.go:18-24                          │  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│  │
│ func coderPrompt(opts)                                   │  │
│   └─ Load embedded templates/coder.md.tpl ◄──────────────┼──┘
│      ↑                                                   │
│      │ THIS IS WHERE CODING-SPECIFIC PROMPT IS LOADED    │
│      │         workflow, editing_files, code_conventions │
│      └───────────────────────────────────────────────────┘
└──────────────────────────────────────────────────────────┘

             ┌─────────────────────────────────────────────────┐
             ▼                                                 │
┌───────────────────────────────────────────────────────────┐  │
│ internal/agent/coordinator.go:buildAgent() (lines 400+)   │  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│  │
│   ├─ Create Fantasy providers (Anthropic, OpenAI, etc.)   │  │
│   ├─ Select models (large + small)                        │  │
│   ├─ Build tool list: ────────────────────────────────────┼─┐│
│   │   ├─ NewViewTool       (read files)                   │ ││
│   │   ├─ NewEditTool       (edit files)                   │ ││
│   │   ├─ NewMultiEditTool  (multi-file edit)              │ ││
│   │   ├─ NewWriteTool      (create files)                 │ ││
│   │   ├─ NewGlobTool       (pattern matching)             │ ││
│   │   ├─ NewLsTool         (list directory)               │ ││
│   │   ├─ NewGrepTool       (search)                       │ ││
│   │   ├─ NewBashTool       (shell commands)               │ ││
│   │   ├─ NewFetchTool      (HTTP requests)                │ ││
│   │   ├─ NewDownloadTool   (file downloads)               │ ││
│   │   ├─ NewDiagnosticsTool (LSP diagnostics)             │ ││
│   │   ├─ NewReferencesTool  (find references)             │ ││
│   │   └─ MCP tools if configured                          │ ││
│   └─ NewSessionAgent(opts) ───────────────────────────────┼─┤│
└───────────────────────────────────────────────────────────┼─┼┘
                                                            │ │
                                                            │ │
```

## 5. TUI Initialization & Event Loop

```
┌───────────────────────────────────────────────────────────────┐
│ internal/tui/tui.go:87-105                                    │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func (a appModel) Init()                                      │
│   ├─ Initialize current page (chat page by default)           │
│   ├─ Initialize status bar                                    │
│   └─ Request terminal version if needed                       │
└───────────────────────────────────────────────────────────────┘
             │
             ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/tui/tui.go:107+                                      │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func (a *appModel) Update(msg tea.Msg)                        │
│   ├─ Handle window resize, keyboard, mouse events             │
│   ├─ Route to current page                                    │
│   ├─ Handle dialogs (model picker, permissions, etc.)         │
│   └─ Handle service events from app.Subscribe()               │
└───────────────────────────────────────────────────────────────┘
             │
             ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/app/app.go:290-320                                   │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func (app *App) Subscribe(program)                            │
│   ├─ Read app.events channel (populated by setupEvents())     │
│   └─ Forward all events to TUI via program.Send(msg)          │
│      (Sessions, Messages, Permissions, History, MCP, LSP)     │
└───────────────────────────────────────────────────────────────┘
```

## 6. User Submits Prompt

```
User types message and presses Enter
             │
             ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/tui/page/chat/chat.go (handleSubmit)                 │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│   ├─ Create session if needed                                 │
│   ├─ Create user message in DB                                │
│   └─ app.AgentCoordinator.Run(ctx, sessionID, prompt) ────┐   │
└───────────────────────────────────────────────────────────┼───┘
                                                            │
             ┌──────────────────────────────────────────────┘
             ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/agent/coordinator.go:111-145                         │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func (c *coordinator) Run(ctx, sessionID, prompt)             │
│   ├─ Get model config and max tokens                          │
│   ├─ Merge provider options (config + catwalk + model)        │
│   └─ c.currentAgent.Run(SessionAgentCall) ────────────────┐   │
└───────────────────────────────────────────────────────────┼───┘
                                                            │
             ┌──────────────────────────────────────────────┘
             ▼
```

## 7. Agent Execution

```
┌───────────────────────────────────────────────────────────────┐
│ internal/agent/agent.go:114-350+                              │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ func (a *sessionAgent) Run(ctx, call)                         │
│   ├─ Check if session busy → queue if yes                     │
│   ├─ Create fantasy.Agent with:                               │
│   │   ├─ System prompt (coder.md.tpl rendered)                │
│   │   └─ Tools (view, edit, bash, etc.)                       │
│   ├─ Load session history from DB                             │
│   ├─ Add user message to history                              │
│   ├─ Create assistant message placeholder in DB               │
│   └─ agent.Stream() ───────────────────────────────────────┐  │
│      ↑                                                     │  │
│      │ THIS SENDS REQUEST TO LLM PROVIDER                  │  │
│      └─────────────────────────────────────────────────────┘  │
│                                                               │
│   Stream callback handler:                                    │
│   ├─ StreamPartTextDelta → Update message content             │
│   ├─ StreamPartToolUse    → Execute tool ─────────────────┐   │
│   ├─ StreamPartTextEnd    → Finalize message              │   │
│   └─ StreamEnd            → Complete                      │   │
└───────────────────────────────────────────────────────────┼───┘
                                                            │
             ┌──────────────────────────────────────────────┘
             ▼
```

## 8. Tool Execution (Example: View Tool)

```
┌────────────────────────────────────────────────────────────────┐
│ LLM decides to use view tool                                   │
│ Returns: {"file_path": "main.go", "offset": 0, "limit": 100}   │
└────────────────────────────────────────────────────────────────┘
             │
             ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/agent/agent.go (handleToolUse callback)              │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│   ├─ Deserialize tool call parameters                         │
│   ├─ Check permissions ────────────────────────────────────┐  │
│   └─ Call tool.Execute(ctx, params) ───────────────────────┼─┐│
└────────────────────────────────────────────────────────────┼─┼┘
                                                             │ │
             ┌───────────────────────────────────────────────┘ │
             ▼                                                 │
┌───────────────────────────────────────────────────────────┐  │
│ internal/permission/service.go                            │  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│  │
│ func Request(permissionRequest)                           │  │
│   ├─ Check if tool in allowed_tools → auto-approve        │  │
│   ├─ Check if --yolo flag set → auto-approve              │  │
│   ├─ Publish permission request event ────────────────────┼─┐│
│   └─ Wait for approval or denial                          │ ││
└───────────────────────────────────────────────────────────┘ ││
             │                                                ││
             ▼                                                ││
┌───────────────────────────────────────────────────────────┐ ││
│ TUI receives permission event via app.Subscribe()         │ ││
│   ├─ Show permission dialog to user                       │ ││
│   ├─ User clicks Approve/Deny                             │ ││
│   └─ permission.Service.Approve()/Deny() ◄────────────────┼─┘│
└───────────────────────────────────────────────────────────┘  │
                                                               │
             ┌─────────────────────────────────────────────────┘
             ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/agent/tools/view.go:54-188                           │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ NewViewTool() → fantasy.AgentTool                             │
│   ├─ Validate file_path parameter                             │
│   ├─ Check file exists                                        │
│   ├─ Check file size (max 250KB)                              │
│   ├─ Check if image → reject ◄────────────────────────────┐   │
│   ├─ Read file content                                    │   │
│   ├─ Check UTF-8 validity → reject if not ◄───────────────┼─┐ │
│   ├─ Add line numbers                                     │ │ │
│   ├─ Notify LSP clients                                   │ │ │
│   ├─ Get diagnostics from LSP                             │ │ │
│   └─ Return formatted response to LLM                     │ │ │
│      ↑                                                    │ │ │
│      │ REJECTION POINTS FOR NON-CODE FILES                │ │ │
│      └────────────────────────────────────────────────────┘ │ │
└───────────────────────────────────────────────────────────┼─┼─┘
                                                            │ │
                    WHERE NON-TEXT FILES ARE BLOCKED ◄──────┘ │
                    WHERE BINARY FILES ARE BLOCKED ◄──────────┘
```

## 9. File Filtering During Discovery

```
User: "analyze all files in this directory"
             │
             ▼
LLM calls ls or glob tool
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│ internal/agent/tools/ls.go or glob.go                          │
│   └─ internal/fsext/ls.go:209+ (ListDirectory)                 │
│       or fileutil.go:79+ (GlobWithDoubleStar)                  │
└────────────────────────────────────────────────────────────────┘
             │
             ▼
┌───────────────────────────────────────────────────────────────┐
│ internal/fsext/ls.go:98-206                                   │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│ type directoryLister                                          │
│ func shouldIgnore(path, ignorePatterns)                       │
│   ├─ Check custom ignorePatterns first                        │
│   ├─ Check commonIgnorePatterns: ◄─────────────────────────┐  │
│   │   ├─ .git, .svn, .hg                                   │  │
│   │   ├─ .vscode, .idea                                    │  │
│   │   ├─ *.log ← IGNORED                                   │  │
│   │   ├─ *.tmp ← IGNORED                                   │  │
│   │   ├─ *.pyc, *.o, *.so, *.exe                           │  │
│   │   ├─ node_modules, target, dist, build                 │  │
│   │   └─ .crush                                            │  │
│   ├─ Check .gitignore in directory                         │  │
│   ├─ Check .crushignore in directory                       │  │
│   ├─ Check parent .gitignore/.crushignore                  │  │
│   └─ Check ~/.config/git/ignore                            │  │
│      ↑                                                     │  │
│      │ FILE FILTERING HAPPENS HERE                         │  │
│      └─────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
             │
             ▼
Only non-ignored files are returned to LLM
```

## 10. Response Streaming Back to User

```
┌────────────────────────────────────────────────────────────────┐
│ Tool results returned to LLM                                   │
│   ↓                                                            │
│ LLM generates response with tool results                       │
│   ↓                                                            │
│ Stream chunks received in agent.Stream() callback              │
│   ↓                                                            │
│ Update message in DB via message.Service                       │
│   ↓                                                            │
│ Publish message update event                                   │
│   ↓                                                            │
│ app.Subscribe() forwards to TUI                                │
│   ↓                                                            │
│ TUI receives message event                                     │
│   ↓                                                            │
│ Chat page updates message display                              │
│   ↓                                                            │
│ User sees response in real-time                                │
└────────────────────────────────────────────────────────────────┘
```

## Key Control Flow Paths

### Path 1: Where System Prompt is Loaded
```
main.go
  → internal/cmd/root.go:setupApp()
    → internal/app/app.go:New()
      → internal/app/app.go:InitCoderAgent()
        → internal/agent/coordinator.go:NewCoordinator()
          → internal/agent/prompts.go:coderPrompt()
            → internal/agent/templates/coder.md.tpl ← LOADED HERE
```

### Path 2: Where File Filtering Happens
```
User types "list files"
  → TUI submits to Agent
    → Agent calls ls/glob tool
      → internal/fsext/ls.go:ListDirectory()
        → internal/fsext/ls.go:directoryLister.shouldIgnore()
          → internal/fsext/ls.go:commonIgnorePatterns() ← *.log, *.tmp BLOCKED
          → .gitignore checked
          → .crushignore checked
```

### Path 3: Where Binary Files Are Rejected
```
Agent calls view tool on "document.pdf"
  → internal/agent/tools/view.go:NewViewTool()
    → Check file exists
    → Check file size
    → isImageFile() check ← rejects images
    → Read file
    → utf8.ValidString(content) ← rejects binary files
    → If not UTF-8 → return error "not valid UTF-8"
```

### Path 4: Event Flow (Real-time Updates)
```
Any service event (message created, permission requested, etc.)
  → Service publishes to channel
    → internal/app/app.go:setupEvents() subscriber picks it up
      → Forwards to app.events channel
        → internal/app/app.go:Subscribe() reads from channel
          → program.Send(msg) to TUI
            → internal/tui/tui.go:Update() receives event
              → Routes to appropriate page/dialog
                → UI updates
```

## Critical Files Summary

### Startup & Configuration
- `main.go` - Entry point
- `internal/cmd/root.go` - CLI command setup
- `internal/config/load.go` - Config loading cascade

### Core Application
- `internal/app/app.go` - Main application orchestration
- `internal/agent/coordinator.go` - Agent lifecycle management
- `internal/agent/agent.go` - Agent execution logic

### System Prompts (Coding-Specific Behavior)
- `internal/agent/prompts.go` - Prompt loading
- **`internal/agent/templates/coder.md.tpl`** ← MAIN CODING PROMPT

### File Filtering (Non-Code File Blocking)
- **`internal/fsext/ls.go:18-81`** ← commonIgnorePatterns (*.log, *.tmp, etc.)
- `internal/fsext/fileutil.go` - File walking with ignore checks

### Tool Implementations (Binary/Non-UTF8 Blocking)
- **`internal/agent/tools/view.go:150-162`** ← Image and binary rejection
- `internal/agent/tools/edit.go` - File editing
- `internal/agent/tools/bash.go` - Shell execution
- `internal/agent/tools/ls.go` - Directory listing
- `internal/agent/tools/glob.go` - Pattern matching

### TUI & Events
- `internal/tui/tui.go` - Main TUI logic
- `internal/app/app.go:setupEvents()` - Event subscription
- `internal/pubsub/` - Pub/Sub infrastructure

### Services (Persistence & State)
- `internal/session/service.go` - Session management
- `internal/message/service.go` - Message persistence
- `internal/permission/service.go` - Tool permission handling
- `internal/history/service.go` - File edit history

## Data Flow Diagram

```
┌─────────┐         ┌──────────┐         ┌─────────┐
│  User   │────────▶│   TUI    │────────▶│   App   │
│ (Input) │         │(Bubble   │         │Instance │
└─────────┘         │  Tea)    │         └─────────┘
                    └──────────┘              │
                         ▲                    │
                         │                    ▼
                         │              ┌──────────────┐
                         │              │Coordinator   │
                         │              │(Agent Mgmt)  │
                         │              └──────────────┘
                         │                    │
                         │                    ▼
                    ┌────────────────────────────────┐
                    │   SessionAgent                 │
                    │   ┌────────────────────────┐   │
                    │   │ Fantasy Agent          │   │
                    │   │ - System Prompt        │   │
                    │   │ - Tools                │   │
                    │   │ - LLM Provider         │   │
                    │   └────────────────────────┘   │
                    └────────────────────────────────┘
                              │         ▲
                              ▼         │
                    ┌─────────────────────────┐
                    │  LLM Provider           │
                    │  (Anthropic/OpenAI/etc) │
                    └─────────────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │ Tool Execution      │
                    │ - view (read)       │
                    │ - edit (modify)     │
                    │ - bash (execute)    │
                    │ - ls/glob (list)    │
                    │ - grep (search)     │
                    └─────────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │  File System        │
                    │  + Filters          │
                    │  + .gitignore       │
                    │  + .crushignore     │
                    └─────────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │  SQLite DB          │
                    │  - Sessions         │
                    │  - Messages         │
                    │  - Files (History)  │
                    └─────────────────────┘
```

## Modification Points for General-Purpose Mode

To make Crush work as a general-purpose research assistant, modify these control points:

1. **System Prompt Selection** (internal/agent/coordinator.go:96)
   - Add logic to choose between `coderPrompt()` and `researchPrompt()`
   - Create new `internal/agent/templates/research.md.tpl`

2. **File Filtering** (internal/fsext/ls.go:18-81)
   - Make `commonIgnorePatterns` configurable
   - Add config flag for research mode vs coder mode

3. **Binary File Handling** (internal/agent/tools/view.go:150-162)
   - Add PDF text extraction instead of rejection
   - Add Word document parsing
   - Make UTF-8 check optional with fallback encoding

4. **Configuration** (internal/config/config.go)
   - Add `Mode` field ("coder" | "research")
   - Add file type inclusion/exclusion lists
   - Add document processing options
