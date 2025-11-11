# Crush - Agent Development Guide

This document provides essential information for agents working on the Crush codebase. Crush is a terminal-based AI assistant for developers, written in Go 1.25.0. See `PROGRAM_FLOW.md` for detailed architecture documentation.

## Quick Start

### Build & Run
```bash
# Build the binary
task build

# Run interactively
task run

# Run with profiling enabled
task dev

# Run specific tests
task test ./internal/agent/...

# Format code
task fmt

# Run linters
task lint
task lint:fix
```

### Key Commands
- `task build` - Compile the application
- `task test` - Run all tests
- `task lint` - Run golangci-lint
- `task install` - Build and install to $GOPATH/bin
- `task schema` - Generate JSON schema for configuration

## Project Structure

```
crush/
├── main.go                         # Entry point with profiling setup
├── go.mod / go.sum                 # Dependencies (Go 1.25.0)
├── Taskfile.yaml                   # Task runner configuration
├── .golangci.yml                   # Linter config (gofumpt, goimports, staticcheck)
│
├── internal/                       # All code is in internal/ (not importable)
│   ├── cmd/                        # CLI commands (root, run, logs, schema, update-providers)
│   ├── app/                        # Main orchestrator, services initialization
│   ├── agent/                      # AI agent system (coordinator, session agent, tools)
│   │   ├── templates/              # System prompts (coder.md.tpl, research.md.tpl, etc.)
│   │   └── tools/                  # Tool implementations (view, edit, bash, etc.)
│   ├── config/                     # Configuration loading and management
│   ├── tui/                        # Terminal UI (Bubble Tea framework)
│   │   ├── page/                   # UI pages (chat, sessions, files, config)
│   │   └── component/              # Reusable UI components
│   ├── db/                         # SQLite database layer
│   │   ├── migrations/             # Goose SQL migrations
│   │   └── *.sql.go                # Generated SQL queries (sqlc)
│   ├── session/                    # Session management service
│   ├── message/                    # Message persistence service
│   ├── permission/                 # Tool permission system
│   ├── history/                    # File edit history service
│   ├── lsp/                        # Language Server Protocol integration
│   ├── fsext/                      # Filesystem utilities (ignore patterns, listing)
│   ├── pubsub/                     # Event publishing system
│   └── [other packages]/           # Utilities for logging, env, version, etc.
```

## Essential Code Patterns

### 1. Package Organization
- All production code lives in `internal/` (not importable externally)
- Each package has a clear responsibility
- Tests use `*_test.go` naming convention
- Use `testify/require` for assertions in tests

**Example structure:**
```go
package agent

// Types and interfaces
type SessionAgent interface {
    Run(context.Context, SessionAgentCall) (*fantasy.AgentResult, error)
    SetTools([]fantasy.AgentTool)
}

// Main implementation
type sessionAgent struct {
    largeModel Model
    smallModel Model
    // ... other fields
}

// Constructor
func NewSessionAgent(opts SessionAgentOptions) SessionAgent {
    return &sessionAgent{...}
}

// Methods
func (a *sessionAgent) Run(ctx context.Context, call SessionAgentCall) (*fantasy.AgentResult, error) {
    // Implementation
}
```

### 2. Error Handling
- Use standard Go error returns
- Wrap errors with context: `fmt.Errorf("operation X failed: %w", err)`
- Log errors at appropriate levels: `slog.Error()`, `slog.Warn()`, `slog.Info()`

```go
if err != nil {
    slog.Error("Failed to initialize agent", "error", err)
    return nil, fmt.Errorf("agent initialization failed: %w", err)
}
```

### 3. Context Usage
- Always accept `context.Context` as first parameter in functions that do I/O
- Use context for cancellation and timeouts
- Never ignore context cancellation signals

```go
func (a *sessionAgent) Run(ctx context.Context, call SessionAgentCall) (*fantasy.AgentResult, error) {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Use context for LLM operations, tool calls, etc.
}
```

### 4. Service Layer Pattern
All services (Sessions, Messages, Permissions, History) follow this pattern:

```go
// Define service interface
type Service interface {
    Create(ctx context.Context, data Entity) error
    Get(ctx context.Context, id string) (Entity, error)
    Update(ctx context.Context, id string, data Entity) error
    Delete(ctx context.Context, id string) error
}

// Implementation with database queries
type serviceImpl struct {
    querier db.Querier
    // other dependencies
}

// Each method calls database and publishes events
func (s *serviceImpl) Create(ctx context.Context, data Entity) error {
    // Validate
    // Query database
    // Publish event
    return nil
}
```

### 5. Tool Implementation Pattern
Tools are implementations of `fantasy.AgentTool`:

```go
// Define parameters struct with JSON tags
type ToolParams struct {
    FilePath string `json:"file_path" description:"Path to file"`
    Offset   int    `json:"offset,omitempty" description:"Start line (0-based)"`
}

// Create tool struct
type myTool struct {
    workingDir string
    permissions permission.Service
}

// Implement Execute method
func (t *myTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
    var p ToolParams
    if err := json.Unmarshal(params, &p); err != nil {
        return "", fmt.Errorf("invalid params: %w", err)
    }
    
    // Request permission
    if err := t.permissions.Request(ctx, ...); err != nil {
        return "", err
    }
    
    // Perform operation
    result := doWork(p)
    return result, nil
}

// Return tool via NewXxxTool function
func NewViewTool(...) fantasy.AgentTool {
    return fantasy.AgentTool{
        Name: "view",
        Description: "Read file contents",
        ParamsSchema: schema, // JSON schema
        Execute: viewTool{...}.Execute,
    }
}
```

### 6. Testing Patterns
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Use `t.Parallel()` for parallel execution where possible
- Use `t.TempDir()` for filesystem tests

```go
func TestFunctionality(t *testing.T) {
    t.Parallel()
    
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"success case", "valid input", false},
        {"error case", "invalid input", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := SomeFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

### 7. Configuration Management
Configuration is loaded from multiple sources in priority order:
1. `.crush.json` (local project)
2. `crush.json` (project root fallback)
3. `$HOME/.config/crush/crush.json` (user global)
4. Environment variables
5. Built-in defaults

All configuration paths go through `internal/config/load.go`.

### 8. Event-Driven Architecture
- Services publish events via `internal/pubsub/`
- TUI subscribes and updates in response to events
- Events flow: Service → pubsub → App.events → TUI
- Never call TUI directly from services

```go
// In service
pubsub.Publish(SessionCreatedEvent{SessionID: sessionID})

// In TUI Update method
case msg SessionCreatedEvent:
    // Update UI
```

## Important Files & Conventions

### Critical Paths
- **System Prompts**: `internal/agent/templates/coder.md.tpl`, `research.md.tpl`
- **Tool Implementations**: `internal/agent/tools/*.go`
- **Configuration Loading**: `internal/config/load.go`
- **File Filtering**: `internal/fsext/ls.go` (ignore patterns)
- **Tool Permissions**: `internal/permission/service.go`
- **Keymappings**: `internal/tui/keys.go` and component-specific `keys.go` files

### Naming Conventions
- **Functions**: `CamelCase` for exported, `camelCase` for unexported
- **Constants**: `UPPER_CASE` for package-level constants
- **Interfaces**: `CamelCase`, usually end with "Service" or singular noun
- **Packages**: `lowercase`, avoid underscores
- **Test files**: `name_test.go` in same package

### Code Style
- Run `task fmt` before committing (uses gofumpt + goimports)
- Max line length: follow gofumpt defaults (~120 chars)
- Comments above exported symbols (GoDocs)
- Use blank lines to separate logical sections
- Prefer `github.com/charmbracelet/x/` utilities over stdlib when available

```go
// GetSessionMessages retrieves all messages for a session.
// Returns messages in chronological order.
func (s *Service) GetSessionMessages(ctx context.Context, sessionID string) ([]Message, error) {
    // Implementation
}
```

## Keymappings

Crush keymappings are organized in a distributed pattern where each UI component maintains its own `keys.go` file. This allows component-specific shortcuts while maintaining global keybindings.

### Global Keybindings

**File**: `internal/tui/keys.go`

Global shortcuts available throughout the TUI:

```go
type KeyMap struct {
    Quit     key.Binding  // ctrl+c → Exit application
    Help     key.Binding  // ctrl+g → Show help
    Commands key.Binding  // ctrl+p → Open commands dialog
    Suspend  key.Binding  // ctrl+z → Suspend application
    Sessions key.Binding  // ctrl+s → Open sessions dialog
}
```

| Shortcut | Action | File |
|----------|--------|------|
| `ctrl+c` | Quit | `internal/tui/keys.go` |
| `ctrl+g` | Help | `internal/tui/keys.go` |
| **`ctrl+p`** | **Commands** | `internal/tui/keys.go` |
| `ctrl+z` | Suspend | `internal/tui/keys.go` |
| `ctrl+s` | Sessions | `internal/tui/keys.go` |

### Component-Specific Keybindings

Each major UI component defines its own keybindings in a local `keys.go` file:

**Chat Page** (`internal/tui/page/chat/keys.go`):
- `ctrl+n` - New session
- `ctrl+f` - Add attachment
- `esc`/`alt+esc` - Cancel
- `tab` - Change focus
- `ctrl+d` - Toggle details

**Chat Editor** (`internal/tui/components/chat/editor/keys.go`):
- `/` - Add file
- `enter` - Send message
- `ctrl+o` - Open in external editor
- `shift+enter`/`ctrl+j` - Insert newline

**Completions** (`internal/tui/components/completions/keys.go`):
- `ctrl+p` - Insert previous completion
- `ctrl+n` - Insert next completion
- `up`/`down` - Navigate completions
- `enter` - Select completion
- `esc` - Close completions

**Dialog Components** (multiple files):
- `Commands Dialog` (`internal/tui/components/dialogs/commands/keys.go`)
- `Models Dialog` (`internal/tui/components/dialogs/models/keys.go`)
- `Sessions Dialog` (`internal/tui/components/dialogs/sessions/keys.go`)
- `Permissions Dialog` (`internal/tui/components/dialogs/permissions/keys.go`)

### Adding or Modifying Keybindings

**Pattern**: Each component uses the Charm `bubbles/key` package:

```go
// File: internal/tui/components/example/keys.go
package example

import "github.com/charmbracelet/bubbles/v2/key"

type KeyMap struct {
    MyAction key.Binding
}

func DefaultKeyMap() KeyMap {
    return KeyMap{
        MyAction: key.NewBinding(
            key.WithKeys("ctrl+x"),           // Actual key(s)
            key.WithHelp("ctrl+x", "action"), // Help text
        ),
    }
}
```

**To add a new keybinding**:
1. Identify the component: Is it global or component-specific?
2. Add to the appropriate `keys.go` file
3. Update the `KeyMap` struct with the new binding
4. Use `key.WithKeys()` for the actual keys
5. Use `key.WithHelp()` for display in help text
6. Update the component's `Update()` method to handle the new key
7. Test with `task run`

### Keymapping Architecture

Keybindings use the `github.com/charmbracelet/bubbles/v2/key` package:

- Each `key.Binding` can have multiple trigger keys: `key.WithKeys("ctrl+p", "cmd+p")`
- Help text is separate from the actual key: `key.WithHelp("ctrl+p", "commands")`
- Components access keybindings via `key.Matches(msg, keymap.Action)`
- Multiple keys can map to the same action

**Example handling in Update method**:

```go
func (c *component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, c.keymap.MyAction):
            // Handle action
            return c, someCmd()
        }
    }
    return c, nil
}
```

## Linting Rules

Active linters in `.golangci.yml`:
- `bodyclose` - Detect unclosed HTTP response bodies
- `goprintffuncname` - Validate printf function naming
- `misspell` - Detect spelling errors
- `noctx` - Detect functions missing context parameter
- `staticcheck` - General Go correctness checks
- `whitespace` - Validate whitespace usage

Disabled linters (by design):
- `errcheck` - Intentional error handling varies
- `unused` - False positives with generated code
- `ineffassign` - Reasonable for this codebase

Run `task lint:fix` to auto-fix most issues.

## Database & Migrations

- **Engine**: SQLite (embedded in binary)
- **Connection**: `internal/db/connect.go`
- **Migrations**: `internal/db/migrations/*.sql` (Goose)
- **Query Generation**: sqlc (generates `*.sql.go` files)
- **Tables**: Sessions, Messages, Files, ToolCalls

When adding database features:
1. Create new migration in `internal/db/migrations/NNN_*.sql`
2. Run migrations via `internal/db/connect.go`
3. Update sqlc queries if needed
4. Regenerate with sqlc

## Dependencies

Key external dependencies:
- `charm.land/fantasy` - LLM provider abstraction
- `github.com/charmbracelet/bubbletea/v2` - TUI framework
- `github.com/charmbracelet/lipgloss/v2` - Terminal styling
- `github.com/spf13/cobra` - CLI framework
- `github.com/modelcontextprotocol/go-sdk` - MCP support
- `github.com/ncruces/go-sqlite3` - SQLite driver

Run `go mod tidy` to clean up dependencies.

## Common Tasks

### Adding a New Tool
1. Create `internal/agent/tools/newtool.go`
2. Implement `fantasy.AgentTool` interface
3. Add parameters struct with JSON tags and descriptions
4. Implement permission checking if needed
5. Register in `internal/agent/coordinator.go:buildAgent()`
6. Add tests in `newtool_test.go`
7. Document in appropriate system prompt template

### Adding Configuration Option
1. Add field to `internal/config/Config` struct
2. Update JSON loading in `internal/config/load.go`
3. Update schema generation if user-facing
4. Add defaults if needed
5. Add tests in `internal/config/*_test.go`

### Modifying System Prompt
1. Edit `internal/agent/templates/coder.md.tpl` or other template
2. Test with `task run -- -c <project_dir>`
3. Verify agent behavior is as expected
4. No code changes needed (templates are embedded)

### Debugging
Enable debug mode with:
```bash
task dev  # Run with profiling enabled
crush -d  # Run with debug logging
```

Profiling endpoints available at `http://localhost:6060`:
- `/debug/pprof/profile` - CPU profile
- `/debug/pprof/heap` - Memory profile
- `/debug/pprof/allocs` - Allocation profile

View profiles:
```bash
task profile:cpu    # 10s CPU profile
task profile:heap   # Memory profile
task profile:allocs # Allocations profile
```

### Testing
```bash
# All tests
task test

# Specific package
task test ./internal/agent/...

# With coverage
task test -- -cover ./...

# Match pattern
task test -- -run TestName ./...

# Verbose output
task test -- -v ./...
```

## Gotchas & Important Notes

1. **File Filtering**: The `view` tool rejects non-UTF8 files and images. This is intentional for the coder mode. Research mode may need different behavior.

2. **Context Cancellation**: Always respect context cancellation. Don't ignore `ctx.Done()`.

3. **Permissions**: Tool execution requires permission unless `--yolo` flag is set or tool is in `allowed_tools` config.

4. **Event Loop**: Never block in event handlers. Keep event handling quick; move heavy work to goroutines.

5. **Database**: All database calls must use context and can be cancelled. Always close connections properly.

6. **Concurrency**: Use `internal/csync/` thread-safe maps for shared state. Prefer channels for goroutine coordination.

7. **Logging**: Use `slog` package with structured fields, never `fmt.Println` for logging.

8. **LSP Integration**: LSP clients are initialized lazily and cached in `app.LSPClients`. Multiple LSP servers can be active.

## Version Management

Version is set at build time:
```bash
-ldflags="-X github.com/mbsulliv/crush/internal/version.Version=<VERSION>"
```

Current version accessed via:
```go
import "github.com/mbsulliv/crush/internal/version"
println(version.Version)
```

## Build & Release

GoReleaser configuration in `.goreleaser.yml`:
- Builds for Linux, macOS, Windows, FreeBSD, OpenBSD, NetBSD
- CGO disabled for portability
- Nightly builds supported
- Shell completion and man pages generated

Release tagged versions with semver via:
```bash
task release  # Interactive semver selection
```

## Documentation

- **Architecture**: See `PROGRAM_FLOW.md` for detailed flow diagrams
- **README.md**: User-facing documentation
- **This file (CRUSH.md)**: Agent development guide
- **Inline comments**: Above exported symbols and for complex logic
- **Type definitions**: Document struct fields with comments

## Contributing Tips

1. **Keep commits focused**: One logical change per commit
2. **Run linters**: `task lint:fix` before committing
3. **Add tests**: All new functionality should have tests
4. **Update docs**: If behavior changes, update relevant documentation
5. **Test edge cases**: Handle errors and invalid inputs gracefully
6. **Consider performance**: Profile before and after for hot paths
7. **Review PROGRAM_FLOW.md**: Understand architecture before major changes

---

# Crush Agent System & Tools Reference

## Agent System Overview

Crush orchestrates AI agents using a coordinator pattern. The system is built around three distinct agent personality types that provide different capabilities and restrictions based on their purpose.

### Architecture Components

- **Coordinator** (`internal/agent/coordinator.go`): Main orchestrator managing agent execution, tool selection, and session management
- **Session Agent** (`internal/agent/agent.go`): Individual agent execution context with model configuration
- **Agent Configuration** (`internal/config/config.go`): Central definitions for agent types and tool access
- **Tool Registry** (`internal/agent/tools/*.go`): 15 tools available to agents

### Agent Selection Logic

Agents are selected based on operating mode:

```go
// From internal/agent/coordinator.go:94-104
switch cfg.Options.Mode {
case config.AgentResearch:
    agentName = config.AgentCoder // Use coder agent config for now
    promptFn = researchPrompt
case config.AgentTask:
    agentName = config.AgentTask
    promptFn = taskPrompt
default:
    agentName = config.AgentCoder
    promptFn = coderPrompt
}
```

---

## Three Agent Types

### 1. Coder Agent (`AgentCoder`)

**Configuration** (`internal/config/config.go:530-537`):
```go
AgentCoder: {
    ID:           AgentCoder,
    Name:         "Coder",
    Description:  "An agent that helps with executing coding tasks.",
    Model:        SelectedModelTypeLarge,
    ContextPaths: c.Options.ContextPaths,
    AllowedTools: allowedTools,  // All 15 tools
}
```

**Capabilities**:
- Full access to all 15 tools
- Can read, modify, and execute code
- Can run bash commands and tests
- Can download files and modify filesystem
- System prompt: `internal/agent/templates/coder.md.tpl`
- Default agent type

**Use Cases**:
- Implementing features
- Fixing bugs
- Running tests
- Code refactoring
- File operations

### 2. Task Agent (`AgentTask`)

**Configuration** (`internal/config/config.go:539-548`):
```go
AgentTask: {
    ID:           AgentCoder,  // Note: Uses AgentCoder for ID
    Name:         "Task",
    Description:  "An agent that helps with searching for context and finding implementation details.",
    Model:        SelectedModelTypeLarge,
    ContextPaths: c.Options.ContextPaths,
    AllowedTools: resolveReadOnlyTools(allowedTools),  // Read-only only
    AllowedMCP:   map[string][]string{},  // MCPs disabled
}
```

**Capabilities**:
- Read-only tools only: `glob`, `grep`, `ls`, `view`, `sourcegraph`, `lsp_references`, `lsp_diagnostics`, `fetch`, `agent`, `journal`
- Cannot modify files or execute bash
- Cannot access MCPs by default
- Safe for untrusted tasks
- System prompt: `internal/agent/templates/task.md.tpl`

**Tool Filtering** (`internal/config/config.go:508-512`):
```go
func resolveReadOnlyTools(tools []string) []string {
    readOnlyTools := []string{"glob", "grep", "ls", "sourcegraph", "view"}
    return filterSlice(tools, readOnlyTools, true)
}
```

**Use Cases**:
- Searching for code patterns
- Finding where functions are used
- Gathering context for implementation
- Reading and analyzing files
- Safe sub-agent delegation

### 3. Research Agent (`AgentResearch`)

**Configuration** (`internal/config/config.go:95-96`):
```go
case config.AgentResearch:
    agentName = config.AgentCoder  // Currently uses Coder config
    promptFn = researchPrompt
```

**Capabilities**:
- Currently inherits Coder agent capabilities (all 15 tools)
- Specialized system prompt for document analysis
- System prompt: `internal/agent/templates/research.md.tpl`
- MANDATORY journal documentation requirement

**Special Behavior**:
- Must use `journal` tool to document findings
- Research prompt includes citation and source tracking rules
- Specialized for analyzing multiple documents

**System Prompt Rules** (from `research.md.tpl`):
- MANDATORY CITATIONS: Every claim must include source attribution
- ALWAYS READ BEFORE ANALYZING: Read documents fully first
- STRUCTURED OUTPUT: Clear organization with proper citations
- ALWAYS JOURNAL RESEARCH: Save findings to JOURNAL.md after each task

**Use Cases**:
- Analyzing research papers or documentation
- Document comparison and synthesis
- Creating searchable research records
- Cross-referencing multiple sources

---

## Agent Delegation Pattern

The **Agent Tool** enables hierarchical agent architecture where a primary agent can delegate tasks to a Task agent.

### How Agent Tool Works

**Location**: `internal/agent/agent_tool.go:28-109`

**Execution Flow**:
1. Primary agent (Coder) calls `agent` tool with a task prompt
2. Creates new task session with Task agent configuration
3. Task agent executes with read-only tools only
4. Returns results without ability to modify system
5. Primary agent receives results and continues execution

**Code Flow** (`agent_tool.go:64-79`):
```go
agentToolSessionID := c.sessions.CreateAgentToolSessionID(agentMessageID, call.ID)
session, err := c.sessions.CreateTaskSession(ctx, agentToolSessionID, sessionID, "New Agent Session")
// ... execute agent with restricted tools
result, err := agent.Run(ctx, SessionAgentCall{
    SessionID:        session.ID,
    Prompt:           params.Prompt,
    MaxOutputTokens:  maxTokens,
    ProviderOptions:  getProviderOptions(model, providerCfg),
    // ... other config
})
```

### Use Cases for Agent Tool

**Example 1: Search Delegation**
```
Coder Agent (primary): "I need to find where the foo() function is called"
                            ↓
                       Uses agent tool
                            ↓
Task Agent (sub): Searches with glob/grep/view
                       Returns results
                            ↓
Coder Agent: Receives findings, uses edit to modify calling code
```

**Example 2: Context Gathering**
```
Coder Agent: "I need context about the database layer"
                    ↓
              Uses agent tool
                    ↓
Task Agent: Uses lsp_references to find all DB calls
           Uses view to read relevant files
           Returns comprehensive context
                    ↓
Coder Agent: Uses context for implementation
```

---

## Complete Tool Reference (15 Tools)

All tools are defined in `internal/agent/tools/` and described in `internal/agent/tools/*.md`.

### Tool Availability Matrix

| Tool | Coder | Task | Research | Type |
|------|-------|------|----------|------|
| view | ✅ | ✅ | ✅ | Read-only |
| glob | ✅ | ✅ | ✅ | Read-only |
| grep | ✅ | ✅ | ✅ | Read-only |
| ls | ✅ | ✅ | ✅ | Read-only |
| fetch | ✅ | ✅ | ✅ | Read-only |
| sourcegraph | ✅ | ✅ | ✅ | Read-only |
| lsp_references | ✅ | ✅ | ✅ | Read-only |
| lsp_diagnostics | ✅ | ✅ | ✅ | Read-only |
| agent | ✅ | ✅ | ✅ | Delegation |
| journal | ✅ | ✅ | ✅ (mandatory) | Documentation |
| edit | ✅ | ❌ | ✅ | Modification |
| multiedit | ✅ | ❌ | ✅ | Modification |
| write | ✅ | ❌ | ✅ | Modification |
| download | ✅ | ❌ | ✅ | Network |
| bash | ✅ | ❌ | ✅ | Execution |

### Tool Categories

**Read-Only Tools** (10 tools): Can search and read but cannot modify
- `glob` - Fast file pattern matching
- `grep` - Content search with regex
- `ls` - Directory tree browsing
- `view` - File reading with line numbers
- `sourcegraph` - Public repository search
- `lsp_references` - Symbol reference lookup
- `lsp_diagnostics` - Code diagnostics
- `fetch` - URL content fetching
- `agent` - Sub-agent delegation
- `journal` - Documentation (read-only for writing)

**File Modification Tools** (4 tools): Modify filesystem
- `edit` - Targeted text replacement (exact match required)
- `multiedit` - Multiple edits in one operation
- `write` - Create or overwrite entire files
- `download` - Download files from URLs

**Execution Tools** (3 tools): Execute system operations
- `bash` - Execute shell commands
- `download` - Download files from network
- `edit`/`multiedit`/`write` - Modify files

### Tool Implementation Pattern

All tools follow a consistent pattern defined in CRUSH.md section 5:

**Location**: `internal/agent/tools/*.go`

**Required Components**:
1. Parameters struct with JSON tags and descriptions
2. Tool registration via `fantasy.NewAgentTool()`
3. Description loaded from embedded markdown file
4. Permission checking if accessing files outside working directory
5. Error handling and validation

**Example - View Tool**:
- File: `internal/agent/tools/view.go`
- Description: `view.md` (embedded)
- Parameters: `file_path`, `offset`, `limit`
- Permissions: Checks if file outside working directory
- Returns: File contents with line numbers and metadata

### Tool Configuration

**All Available Tools** (`internal/config/config.go:480-497`):
```go
func allToolNames() []string {
    return []string{
        "agent",
        "bash",
        "download",
        "edit",
        "multiedit",
        "lsp_diagnostics",
        "lsp_references",
        "fetch",
        "glob",
        "grep",
        "journal",
        "ls",
        "sourcegraph",
        "view",
        "write",
    }
}
```

**Tool Filtering** (`internal/config/config.go:500-512`):
```go
// Resolve allowed tools (exclude disabled tools)
allowedTools := resolveAllowedTools(allToolNames(), c.Options.DisabledTools)

// Task agent gets only read-only tools
func resolveReadOnlyTools(tools []string) []string {
    readOnlyTools := []string{"glob", "grep", "ls", "sourcegraph", "view"}
    return filterSlice(tools, readOnlyTools, true)
}
```

---

## Tool Execution Flow

### Permission & Execution Model

All tool execution goes through the permission system:

1. **Permission Request**: Tool checks if operation requires permission
2. **User Prompt**: If needed, prompts user to approve
3. **Execution**: Tool executes if approved
4. **Recording**: Successful operations recorded for history

**Permission Sources** (`internal/permission/service.go`):
- Files outside working directory require permission
- All tools require permission unless `--yolo` flag is set
- Can configure `allowed_tools` in config to auto-approve

### Tool Session Context

Tools have access to session context:

```go
// Getting session info from context
sessionID := tools.GetSessionFromContext(ctx)
messageID := tools.GetMessageFromContext(ctx)
```

This allows:
- Tracking which session invoked the tool
- Creating nested sessions for agent tool
- Recording tool calls in history

---

## System Prompts

System prompts define agent personality and behavior rules.

### Prompt Templates

**Coder Prompt** (`internal/agent/templates/coder.md.tpl`)
- Critical rules for autonomous coding
- Communication style (minimal, under 4 lines)
- Workflow instructions (read before editing)
- Code reference format

**Task Prompt** (`internal/agent/templates/task.md.tpl`)
- Rules for search and context gathering
- Emphasis on conciseness
- Direct answers without elaboration
- Absolute paths for file references

**Research Prompt** (`internal/agent/templates/research.md.tpl`)
- Mandatory citations for every statement
- Read documents fully before analyzing
- Structured output format
- Journal documentation requirement

**Agent Tool Prompt** (`internal/agent/templates/agent_tool.md`)
- Usage guidelines for agent tool
- Tool availability (glob, grep, ls, view)
- Limitations (no bash, edit, write)
- Stateless execution model

---

## Agent Configuration

### Configuration Loading

**Locations** (priority order):
1. `.crush.json` (local project)
2. `crush.json` (project root fallback)
3. `$HOME/.config/crush/crush.json` (user global)
4. Environment variables
5. Built-in defaults

**Configuration Setup** (`internal/config/config.go:526-551`):
```go
func (c *Config) SetupAgents() {
    allowedTools := resolveAllowedTools(allToolNames(), c.Options.DisabledTools)
    
    agents := map[string]Agent{
        AgentCoder: { ... },
        AgentTask: { ... },
    }
    c.Agents = agents
}
```

### Configuration Fields

**Agent Struct** (`internal/config/config.go:262-283`):
```go
type Agent struct {
    ID          string             // Agent identifier
    Name        string             // Display name
    Description string             // Human-readable description
    Disabled    bool               // Whether agent is disabled
    Model       SelectedModelType  // large or small model
    AllowedTools []string          // List of available tools
    AllowedMCP  map[string][]string // MCP server access
    ContextPaths []string          // Files/directories for context
}
```

---

## Important Implementation Details

### Agent Initialization

**Coordinator Creation** (`internal/agent/coordinator.go:71-127`):
1. Determines agent type based on mode
2. Loads appropriate system prompt template
3. Builds agent configuration
4. Creates session agent instance
5. Sets up tools and models

### Tool Building

**Tool Setup** (`internal/agent/coordinator.go:341-410`):
```go
func (c *coordinator) buildTools(ctx context.Context, agent config.Agent) ([]fantasy.AgentTool, error) {
    var allTools []fantasy.AgentTool
    
    // Add tools based on agent configuration
    // Filter by AllowedTools list
    // Check DisabledTools
    // Sort tools
    
    return filteredTools, nil
}
```

### Session Management

**Session Creation** (`internal/agent/coordinator.go:63-64`):
- Primary sessions: User-initiated agent runs
- Task sessions: Sub-agent delegated work
- Session tracking: Cost, messages, context
- Nested sessions: Agent tool creates child sessions

---

## Best Practices for Agent Development

### When Adding New Tools

1. **Create tool file**: `internal/agent/tools/newtool.go`
2. **Add description**: `internal/agent/tools/newtool.md`
3. **Implement interface**: `fantasy.AgentTool` with Execute method
4. **Register in coordinator**: Add to `buildTools()` method
5. **Add tests**: `internal/agent/tools/newtool_test.go`
6. **Consider safety**: Add permission checks if needed
7. **Document in prompts**: Update system prompt templates if needed

### When Modifying Agent Behavior

1. **Edit system prompt**: `internal/agent/templates/*.md.tpl`
2. **Test behavior**: Run `task run` and test with real scenarios
3. **Verify tools**: Ensure tool availability matches agent type
4. **Update configuration**: If changing allowed tools, update `SetupAgents()`
5. **No code changes needed**: Prompts are embedded and reloaded

### When Debugging Agent Issues

1. **Enable debug logging**: `crush -d` or set `debug: true` in config
2. **Check system prompt**: Verify correct prompt is loaded
3. **Review tool availability**: Confirm agent has required tools
4. **Check permissions**: Verify tool permissions not blocking execution
5. **Examine logs**: `.crush/logs/crush.log` contains detailed execution logs
6. **Test with `-c` flag**: Run against specific project directory

---

## References

- **Detailed Architecture**: See `PROGRAM_FLOW.md`
- **User Documentation**: See `README.md`
- **System Prompts**: `internal/agent/templates/*.md.tpl`
- **Tool Documentation**: `internal/agent/tools/*.md`
- **Configuration Reference**: `internal/config/config.go`
- **Coordinator Implementation**: `internal/agent/coordinator.go`
