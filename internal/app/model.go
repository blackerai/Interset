package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"interset/internal/mcp"
	"interset/internal/registry"
	"interset/internal/session"
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
	spinner        spinner.Model
	startedAt      time.Time
	lastStatusNote string
}

func New() Model {
	spin := spinner.New()
	spin.Spinner = spinner.MiniDot
	spin.Style = ui.Theme().Accent

	providers := registry.DefaultProviders()
	profileSet := mcp.DefaultProfiles()

	return Model{
		mode:           modeHome,
		showSidebar:    true,
		focus:          focusHome,
		providers:      providers,
		sessions:       session.NewManager(),
		mcpProfiles:    profileSet,
		activeProfile:  profileSet[0].ID,
		spinner:        spin,
		startedAt:      time.Now(),
		lastStatusNote: "home ready",
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
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
		case "up", "k":
			m.moveSidebar(-1)
			m.lastStatusNote = "provider selected"
		case "down", "j":
			m.moveSidebar(1)
			m.lastStatusNote = "provider selected"
		case "enter":
			m.openSelectedProvider()
		case "ctrl+t":
			m.openSelectedProvider()
		case "ctrl+w":
			m.sessions.CloseActive()
			m.syncMode()
			m.lastStatusNote = "tab closed"
		case "tab", "ctrl+tab", "l":
			if m.sessions.HasTabs() {
				m.sessions.Next()
				m.mode = modeWorkspace
				m.focus = focusTabs
				m.lastStatusNote = "next tab"
			}
		case "shift+tab", "ctrl+shift+tab", "h":
			if m.sessions.HasTabs() {
				m.sessions.Prev()
				m.mode = modeWorkspace
				m.focus = focusTabs
				m.lastStatusNote = "previous tab"
			}
		case "left":
			if m.mode == modeWorkspace && m.showSidebar {
				m.focus = focusSidebar
				m.lastStatusNote = "sidebar focus"
			}
		case "right":
			if m.mode == modeWorkspace && m.sessions.HasTabs() {
				m.focus = focusSession
				m.lastStatusNote = "session focus"
			}
		case "ctrl+r":
			if active := m.sessions.Active(); active != nil {
				active.Status = session.StatusStarting
				active.LastActivityAt = time.Now()
				m.lastStatusNote = "restart requested"
			}
		case "ctrl+p":
			m.lastStatusNote = "command palette planned"
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Starting Interset..."
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
		ActiveSession:  m.sessions.Active(),
		ActiveProfile:  m.activeProfile,
		StatusNote:     strings.ToLower(m.lastStatusNote),
		SpinnerFrame:   m.spinner.View(),
		Uptime:         time.Since(m.startedAt).Round(time.Second),
	})
}

func (m *Model) moveSidebar(delta int) {
	if len(m.providers) == 0 {
		return
	}

	m.sidebarIndex = (m.sidebarIndex + delta + len(m.providers)) % len(m.providers)
}

func (m *Model) openSelectedProvider() {
	provider := m.selectedProvider()
	if provider.ID == "" {
		return
	}

	m.sessions.New(provider.ID, provider.DisplayName)
	m.mode = modeWorkspace
	m.focus = focusSession
	m.lastStatusNote = fmt.Sprintf("opened %s", strings.ToLower(provider.DisplayName))
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

func modeLabel(mode appMode) string {
	if mode == modeWorkspace {
		return "workspace"
	}
	return "home"
}
