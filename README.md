# Interset

Interset is a Bubble Tea workstation for AI and developer CLIs. The product goal is not to emulate `tmux`, but to give developers a curated multi-session browser for tools like Codex, Gemini CLI, Crush, Linear CLI, Resend CLI, and future providers.

## Product shape

The initial shell is designed around four persistent regions:

- left provider sidebar
- top browser-style tabs
- center live session panel
- bottom status bar

The shell is intentionally opinionated:

- tabs represent real sessions, not panes
- provider integration stays modular behind a registry
- PTY concerns remain isolated from rendering
- MCP uses one canonical internal schema plus adapters per provider
- persistence restores tabs, settings, profiles, and recent context without pretending processes survive app restarts

## Why Bubble Tea fits

Bubble Tea is a strong fit for this product because it gives us:

- a deterministic message loop for keyboard-heavy workflows
- composable submodels for tabs, sidebar, terminal viewport, and status
- explicit command boundaries for side effects like provider detection, session spawning, restore, and persistence
- clear control over render performance and focus

This matters because the workstation needs to juggle UI state, PTY IO, provider health, and restore flows without becoming an event-driven mess.

## Proposed architecture

```text
cmd/app              # bootstraps the Bubble Tea program
internal/app         # root model, global update loop, message orchestration
internal/ui          # theme, layout, sidebar, tabs, session panel, status bar
internal/session     # tab/session models, lifecycle, restore-ready state
internal/pty         # PTY abstraction and process bridge
internal/registry    # provider definitions, detection metadata, install/auth hints
internal/mcp         # canonical MCP schema, profiles, provider adapters
internal/config      # app config and workspace overrides
internal/store       # local persistence for tabs, recents, settings, profiles
internal/platform    # OS-specific decisions and process helpers
```

### Module responsibilities

`internal/app`

- owns top-level navigation and focus
- routes cross-cutting messages
- batches commands for restore, provider detection, and status changes
- prevents PTY details from leaking into layout code

`internal/ui`

- renders the workstation chrome
- keeps view concerns separate from mutable process state
- centralizes theming and spacing rules

`internal/session`

- models tab state and live-session identity
- tracks active tab, unread activity, restart intent, and restore snapshots
- owns session metadata, not terminal drawing

`internal/pty`

- launches interactive processes
- translates PTY output into typed messages
- handles resize and input
- hides OS-specific behavior behind an interface

`internal/registry`

- stores provider metadata
- maps detect command, launch command, install hints, auth hints, and MCP support
- supports future providers without reshaping app state

`internal/mcp`

- defines the canonical schema for profiles
- exports provider-specific config fragments later
- supports global profiles plus tab/workspace overrides

`internal/store`

- persists serializable state only
- restores tabs, provider preferences, MCP profile references, recents, and window preferences

`internal/platform`

- isolates Windows, macOS, and Linux differences
- will matter heavily once PTY and auth flows land

## Core data model

The app should treat tabs and sessions as separate but tightly linked concepts.

```go
type Session struct {
    ID             string
    TabID          string
    ProviderID     string
    Title          string
    Cwd            string
    Env            map[string]string
    Status         SessionStatus
    CreatedAt      time.Time
    LastActivityAt time.Time
    MCPProfile     string
    Unread         bool
    Dirty          bool
}
```

Key design decision: restoring the app should restore session intent and tab context, but not promise that old PTY handles still exist after a cold restart.

## Bubble Tea message flow

The main message categories for v1 should be:

- `tea.WindowSizeMsg`
- global key input
- `ProviderDetectionFinishedMsg`
- `ProviderStatusChangedMsg`
- `SessionCreatedMsg`
- `SessionOutputMsg`
- `SessionExitedMsg`
- `SessionErrorMsg`
- `TabSelectedMsg`
- `SidebarToggledMsg`
- `RestoreFinishedMsg`
- `MCPProfileChangedMsg`

Command ownership should stay explicit:

- provider detection lives behind commands started by `app.Init()`
- PTY launch lives behind session commands
- persistence write-back is triggered by state changes but kept outside reducers
- UI components never reach into process management directly

## MVP roadmap

### Phase 1: product shell

- Bubble Tea app boot
- premium dark theme
- sidebar, tabs, center panel, status bar
- keybindings for navigation and tab actions

### Phase 2: navigation and state

- focus management
- provider selection
- new/close/switch tab flows
- tab rename
- sidebar loader/state badges

### Phase 3: real sessions

- PTY bridge
- provider launch commands
- resize + input forwarding
- output streaming into the active session view
- background sessions stay alive across tab switches

### Phase 4: persistence

- tab/session snapshots
- settings and recents
- provider defaults
- MCP profile persistence
- restore on startup

### Phase 5: provider registry

- detect installed CLIs
- install/auth hints
- provider health snapshots
- future provider onboarding path

### Phase 6: canonical MCP layer

- internal schema
- reusable profiles
- workspace and tab overrides
- provider adapters/exporters

### Phase 7: polish

- command palette
- quick actions
- tab search
- startup restore UX
- performance tuning and stability hardening

## Real technical risks

- PTY behavior differs substantially across Windows, macOS, and Linux
- terminal rendering and keyboard forwarding can fight Bubble Tea focus rules if boundaries are unclear
- some CLIs may expect a fully native shell environment or login-shell semantics
- provider-specific MCP export formats will diverge over time
- restoring tabs is easy; restoring true process continuity across app restarts is not realistic for v1

## Tradeoffs

- Use a canonical MCP schema internally to avoid provider lock-in
- Delay deep provider integration until registry, launch, and health are solid
- Prefer one active terminal viewport over multiplexed panes for clarity
- Prioritize restoreable intent over magical process resurrection

## Distribution

Runtime should remain a native Go binary. Future npm distribution can wrap that binary for easier install:

1. publish platform binaries from CI
2. ship an npm package that downloads the right binary for the host OS/arch
3. expose a thin `npx interset` or global install path
4. keep runtime logic in Go, not Node

That gives easy developer onboarding without pulling the core app into Electron or a Node runtime.
