package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"interset/internal/config"
	"interset/internal/mcp"
	"interset/internal/platform"
	"interset/internal/pty"
	"interset/internal/registry"
	"interset/internal/session"
	"interset/internal/store"
	"interset/internal/ui"
)

type Model struct {
	width          int
	height         int
	mode           appMode
	showSidebar    bool
	focus          focusArea
	sidebarIndex   int
	providers      []registry.Provider
	sessions       *session.Manager
	mcpProfiles    []mcp.Profile
	activeProfile  string
	cfg            config.App
	ptyManager     pty.Manager
	runtimeEvents  chan tea.Msg
	spinner        spinner.Model
	startedAt      time.Time
	lastStatusNote string
}

func New() Model {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	providers := registry.DefaultProviders()
	profileSet := mcp.DefaultProfiles()
	activeProfile := cfg.DefaultMCPProfile
	if activeProfile == "" {
		activeProfile = profileSet[0].ID
	}

	spin := spinner.New()
	spin.Spinner = spinner.MiniDot
	spin.Style = ui.Theme().Accent

	model := Model{
		mode:           modeHome,
		showSidebar:    true,
		focus:          focusHome,
		providers:      providers,
		sessions:       session.NewManager(),
		mcpProfiles:    profileSet,
		activeProfile:  activeProfile,
		cfg:            cfg,
		ptyManager:     pty.NewManager(),
		runtimeEvents:  make(chan tea.Msg, 512),
		spinner:        spin,
		startedAt:      time.Now(),
		lastStatusNote: "home ready",
	}
	model.sidebarIndex = model.indexForProvider(cfg.DefaultProvider)

	return model
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		detectProvidersCmd(m.providers),
		restoreSnapshotCmd(m.cfg.RestoreOnStartup),
		waitForAsyncMsg(m.runtimeEvents),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, resizeRuntimeCmd(m.sessions.ActiveRuntime(), msg.Width, msg.Height)
	case tea.KeyMsg:
		return m.handleKey(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case ProviderDetectionFinishedMsg:
		m.providers = msg.Providers
		m.sidebarIndex = m.indexForProvider(m.selectedProvider().ID)
		m.lastStatusNote = "providers detected"
		return m, nil
	case RestoreCompletedMsg:
		cmds := []tea.Cmd{waitForAsyncMsg(m.runtimeEvents)}
		if msg.Err != nil {
			m.lastStatusNote = "restore failed"
			return m, tea.Batch(cmds...)
		}

		if len(msg.Snapshot.Tabs) == 0 {
			return m, tea.Batch(cmds...)
		}

		m.sessions.Restore(msg.Snapshot)
		m.mode = modeWorkspace
		m.focus = focusSession
		m.lastStatusNote = "tabs restored"

		for _, sessionEntry := range m.sessions.Sessions() {
			cmds = append(cmds, startRuntimeCmd(sessionEntry.ID, sessionEntry.RuntimeVersion, m.ptyManager, m.startSpec(sessionEntry), m.runtimeEvents))
		}

		return m, tea.Batch(cmds...)
	case SessionStartedMsg:
		current := m.sessions.SessionByID(msg.SessionID)
		if current == nil || current.RuntimeVersion != msg.RuntimeVersion {
			return m, tea.Batch(closeRuntimeCmd(msg.Runtime), waitForAsyncMsg(m.runtimeEvents))
		}

		m.sessions.AttachRuntime(msg.SessionID, msg.Runtime)
		if started := m.sessions.SessionByID(msg.SessionID); started != nil {
			m.setProviderStatus(started.ProviderID, registry.StatusBusy, "")
		}
		return m, tea.Batch(waitForAsyncMsg(m.runtimeEvents), saveSnapshotCmd(m.snapshot()))
	case SessionOutputReceivedMsg:
		current := m.sessions.SessionByID(msg.SessionID)
		if current == nil || current.RuntimeVersion != msg.RuntimeVersion {
			return m, waitForAsyncMsg(m.runtimeEvents)
		}

		m.sessions.AppendOutput(msg.SessionID, msg.Data)
		if outputSession := m.sessions.SessionByID(msg.SessionID); outputSession != nil {
			m.setProviderStatus(outputSession.ProviderID, registry.StatusBusy, "")
		}
		return m, waitForAsyncMsg(m.runtimeEvents)
	case SessionExitedMsg:
		current := m.sessions.SessionByID(msg.SessionID)
		if current == nil || current.RuntimeVersion != msg.RuntimeVersion {
			return m, waitForAsyncMsg(m.runtimeEvents)
		}

		m.sessions.MarkExited(msg.SessionID, msg.ExitCode)
		if exited := m.sessions.SessionByID(msg.SessionID); exited != nil {
			m.setProviderStatus(exited.ProviderID, registry.StatusIdle, "")
		}
		m.lastStatusNote = "session exited"
		return m, tea.Batch(waitForAsyncMsg(m.runtimeEvents), saveSnapshotCmd(m.snapshot()))
	case SessionFailedMsg:
		current := m.sessions.SessionByID(msg.SessionID)
		if current == nil || current.RuntimeVersion != msg.RuntimeVersion {
			return m, waitForAsyncMsg(m.runtimeEvents)
		}

		m.sessions.MarkFailed(msg.SessionID, msg.Err)
		if failed := m.sessions.SessionByID(msg.SessionID); failed != nil {
			m.setProviderStatus(failed.ProviderID, registry.StatusError, msg.Err)
		}
		m.lastStatusNote = "session failed"
		return m, tea.Batch(waitForAsyncMsg(m.runtimeEvents), saveSnapshotCmd(m.snapshot()))
	case SnapshotSavedMsg:
		if msg.Err != nil {
			m.lastStatusNote = "state save failed"
		}
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Starting Interset..."
	}

	active := m.sessions.Active()
	var output []string
	if active != nil {
		output = m.sessions.OutputTail(active.ID, maxLines(m.height))
	}

	return ui.RenderWorkspace(ui.WorkspaceProps{
		Width:          m.width,
		Height:         m.height,
		Mode:           modeLabel(m.mode),
		ShowSidebar:    m.showSidebar,
		Focus:          int(m.focus),
		Providers:      m.providers,
		SidebarIndex:   m.sidebarIndex,
		Tabs:           m.sessions.Tabs(),
		ActiveSession:  active,
		ActiveOutput:   output,
		ActiveProfile:  m.activeProfile,
		StatusNote:     strings.ToLower(m.lastStatusNote),
		SpinnerFrame:   m.spinner.View(),
		Uptime:         time.Since(m.startedAt).Round(time.Second),
	})
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == modeWorkspace && m.focus == focusSession {
		if handled, cmd := m.handleSessionInput(msg); handled {
			return m, cmd
		}
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "ctrl+b":
		m.showSidebar = !m.showSidebar
		if m.showSidebar {
			m.focus = focusSidebar
		} else if m.mode == modeHome {
			m.focus = focusHome
		} else {
			m.focus = focusSession
		}
		m.lastStatusNote = "sidebar toggled"
		return m, nil
	case "up", "k":
		m.moveSidebar(-1)
		m.lastStatusNote = "provider selected"
		return m, nil
	case "down", "j":
		m.moveSidebar(1)
		m.lastStatusNote = "provider selected"
		return m, nil
	case "enter":
		return m.openSelectedProvider()
	case "ctrl+t":
		return m.openShellTab()
	case "ctrl+w":
		return m.closeActiveTab()
	case "tab", "ctrl+tab", "l":
		if m.sessions.HasTabs() {
			m.sessions.Next()
			m.mode = modeWorkspace
			m.focus = focusTabs
			m.lastStatusNote = "next tab"
			return m, saveSnapshotCmd(m.snapshot())
		}
		return m, nil
	case "shift+tab", "ctrl+shift+tab", "h":
		if m.sessions.HasTabs() {
			m.sessions.Prev()
			m.mode = modeWorkspace
			m.focus = focusTabs
			m.lastStatusNote = "previous tab"
			return m, saveSnapshotCmd(m.snapshot())
		}
		return m, nil
	case "left":
		if m.mode == modeWorkspace && m.showSidebar {
			m.focus = focusSidebar
			m.lastStatusNote = "sidebar focus"
		}
		return m, nil
	case "right":
		if m.mode == modeWorkspace && m.sessions.HasTabs() {
			m.focus = focusSession
			m.lastStatusNote = "session focus"
		}
		return m, nil
	case "ctrl+r":
		return m.restartActiveSession()
	case "ctrl+p":
		m.lastStatusNote = "command palette planned"
		return m, nil
	}

	return m, nil
}

func (m Model) handleSessionInput(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		text := string(msg.Runes)
		if text == "" {
			return false, nil
		}
		return true, m.writeToActive(text)
	case tea.KeySpace:
		return true, m.writeToActive(" ")
	case tea.KeyEnter:
		return true, m.writeToActive("\r\n")
	case tea.KeyBackspace:
		return true, m.writeToActive("\b")
	case tea.KeyTab:
		return true, m.writeToActive("\t")
	}

	return false, nil
}

func (m Model) openShellTab() (tea.Model, tea.Cmd) {
	command := platform.ResolveDefaultShell(m.cfg.DefaultShell)
	profileID, env := mcp.ResolveEnv(m.cfg, m.mcpProfiles, ".", m.activeProfile, nil, nil)

	entry := m.sessions.Create(session.CreateOptions{
		ProviderID:    "shell",
		Title:         "Shell",
		Cwd:           ".",
		Env:           env,
		MCPProfile:    profileID,
		LaunchCommand: command,
		RuntimeKind:   session.RuntimeShell,
	})

	m.mode = modeWorkspace
	m.focus = focusSession
	m.lastStatusNote = "shell tab opened"

	return m, tea.Batch(
		startRuntimeCmd(entry.ID, entry.RuntimeVersion, m.ptyManager, m.startSpec(entry), m.runtimeEvents),
		saveSnapshotCmd(m.snapshot()),
	)
}

func (m Model) openSelectedProvider() (tea.Model, tea.Cmd) {
	provider := m.selectedProvider()
	if provider.ID == "" {
		return m, nil
	}

	if provider.Status == registry.StatusMissing {
		m.lastStatusNote = "provider missing"
		return m, nil
	}

	command := provider.LaunchCommand
	if provider.DetectedPath != "" {
		command = append([]string{provider.DetectedPath}, provider.LaunchCommand[1:]...)
	}
	profileID, env := mcp.ResolveEnv(m.cfg, m.mcpProfiles, ".", m.activeProfile, nil, nil)

	entry := m.sessions.Create(session.CreateOptions{
		ProviderID:    provider.ID,
		Title:         provider.DisplayName,
		Cwd:           ".",
		Env:           env,
		MCPProfile:    profileID,
		LaunchCommand: command,
		RuntimeKind:   session.RuntimeProvider,
	})

	m.mode = modeWorkspace
	m.focus = focusSession
	m.setProviderStatus(provider.ID, registry.StatusStarting, "")
	m.lastStatusNote = fmt.Sprintf("opened %s", strings.ToLower(provider.DisplayName))

	return m, tea.Batch(
		startRuntimeCmd(entry.ID, entry.RuntimeVersion, m.ptyManager, m.startSpec(entry), m.runtimeEvents),
		saveSnapshotCmd(m.snapshot()),
	)
}

func (m Model) closeActiveTab() (tea.Model, tea.Cmd) {
	activeRuntime := m.sessions.ActiveRuntime()
	closed := m.sessions.CloseActive()
	if closed == nil {
		return m, nil
	}

	m.syncMode()
	if closed.ProviderID != "shell" {
		m.setProviderStatus(closed.ProviderID, registry.StatusIdle, "")
	}
	m.lastStatusNote = "tab closed"

	return m, tea.Batch(
		closeRuntimeCmd(activeRuntime),
		saveSnapshotCmd(m.snapshot()),
	)
}

func (m Model) restartActiveSession() (tea.Model, tea.Cmd) {
	activeRuntime := m.sessions.ActiveRuntime()
	active := m.sessions.RestartActive()
	if active == nil {
		return m, nil
	}

	if active.ProviderID != "shell" {
		m.setProviderStatus(active.ProviderID, registry.StatusStarting, "")
	}
	m.lastStatusNote = "restart requested"

	return m, tea.Batch(
		closeRuntimeCmd(activeRuntime),
		startRuntimeCmd(active.ID, active.RuntimeVersion, m.ptyManager, m.startSpec(active), m.runtimeEvents),
		saveSnapshotCmd(m.snapshot()),
	)
}

func (m Model) writeToActive(data string) tea.Cmd {
	return func() tea.Msg {
		active := m.sessions.Active()
		if active == nil {
			return nil
		}
		if err := m.sessions.WriteActive(data); err != nil {
			return SessionFailedMsg{
				SessionID:      active.ID,
				RuntimeVersion: active.RuntimeVersion,
				Err:            err.Error(),
			}
		}
		return nil
	}
}

func (m *Model) moveSidebar(delta int) {
	if len(m.providers) == 0 {
		return
	}

	m.sidebarIndex = (m.sidebarIndex + delta + len(m.providers)) % len(m.providers)
}

func (m *Model) syncMode() {
	if m.sessions.HasTabs() {
		m.mode = modeWorkspace
		if m.showSidebar {
			m.focus = focusSidebar
		} else {
			m.focus = focusSession
		}
		return
	}

	m.mode = modeHome
	m.focus = focusHome
	m.lastStatusNote = "home ready"
}

func (m Model) selectedProvider() registry.Provider {
	if len(m.providers) == 0 {
		return registry.Provider{}
	}

	index := m.sidebarIndex
	if index < 0 || index >= len(m.providers) {
		index = 0
	}
	return m.providers[index]
}

func (m Model) startSpec(entry *session.Session) pty.StartSpec {
	return pty.StartSpec{
		Command: entry.LaunchCommand,
		Cwd:     entry.Cwd,
		Env:     entry.Env,
	}
}

func (m Model) snapshot() store.Snapshot {
	snapshot := store.Snapshot{
		LastOpenAt:   time.Now(),
		ActiveTabID:  m.sessions.ActiveTabID(),
		WindowWidth:  m.width,
		WindowHeight: m.height,
	}

	for _, tab := range m.sessions.Tabs() {
		sessionEntry := m.sessions.SessionByID(tab.SessionID)
		if sessionEntry == nil {
			continue
		}

		snapshot.Tabs = append(snapshot.Tabs, store.StoredTab{
			ID:             tab.ID,
			SessionID:      sessionEntry.ID,
			Title:          tab.Title,
			ProviderID:     sessionEntry.ProviderID,
			Cwd:            sessionEntry.Cwd,
			MCPProfile:     sessionEntry.MCPProfile,
			LaunchCommand:  append([]string{}, sessionEntry.LaunchCommand...),
			RuntimeKind:    string(sessionEntry.RuntimeKind),
			Env:            copyEnv(sessionEntry.Env),
			CreatedAt:      sessionEntry.CreatedAt,
			LastActivityAt: sessionEntry.LastActivityAt,
		})
	}

	return snapshot
}

func (m *Model) setProviderStatus(providerID string, status registry.Status, err string) {
	for i := range m.providers {
		if m.providers[i].ID != providerID {
			continue
		}
		m.providers[i].Status = status
		m.providers[i].LastError = err
		return
	}
}

func (m Model) indexForProvider(providerID string) int {
	for i, provider := range m.providers {
		if provider.ID == providerID {
			return i
		}
	}
	return 0
}

func modeLabel(mode appMode) string {
	if mode == modeWorkspace {
		return "workspace"
	}
	return "home"
}

func maxLines(height int) int {
	if height < 20 {
		return 8
	}
	return height - 10
}

func copyEnv(source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}

	out := make(map[string]string, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
