package session

import (
	"fmt"
	"strings"
	"time"

	"interset/internal/pty"
	"interset/internal/store"
)

type Status string

const (
	StatusIdle     Status = "idle"
	StatusStarting Status = "starting"
	StatusBusy     Status = "busy"
	StatusError    Status = "error"
	StatusExited   Status = "exited"
)

type RuntimeKind string

const (
	RuntimeShell    RuntimeKind = "shell"
	RuntimeProvider RuntimeKind = "provider"
)

type Session struct {
	ID             string
	TabID          string
	ProviderID     string
	Title          string
	Cwd            string
	Env            map[string]string
	Status         Status
	CreatedAt      time.Time
	LastActivityAt time.Time
	MCPProfile     string
	Unread         bool
	Dirty          bool
	LaunchCommand  []string
	RuntimeKind    RuntimeKind
	RuntimeVersion int
	ExitCode       *int
	LastError      string
	Output         string
}

type Tab struct {
	ID         string
	SessionID  string
	Title      string
	ProviderID string
	Active     bool
}

type RuntimeState struct {
	Process pty.Session
}

type CreateOptions struct {
	ProviderID    string
	Title         string
	Cwd           string
	Env           map[string]string
	MCPProfile    string
	LaunchCommand []string
	RuntimeKind   RuntimeKind
}

type Manager struct {
	tabs         []Tab
	sessions     map[string]*Session
	runtimes     map[string]*RuntimeState
	activeTabIdx int
	nextID       int
	outputLimit  int
}

func NewManager() *Manager {
	return &Manager{
		sessions:     make(map[string]*Session),
		runtimes:     make(map[string]*RuntimeState),
		activeTabIdx: -1,
		outputLimit:  64000,
	}
}

func (m *Manager) Create(options CreateOptions) *Session {
	m.nextID++
	tabID := fmt.Sprintf("tab-%d", m.nextID)
	sessionID := fmt.Sprintf("session-%d", m.nextID)

	for i := range m.tabs {
		m.tabs[i].Active = false
	}

	tab := Tab{
		ID:         tabID,
		SessionID:  sessionID,
		Title:      options.Title,
		ProviderID: options.ProviderID,
		Active:     true,
	}
	m.tabs = append(m.tabs, tab)

	now := time.Now()
	entry := &Session{
		ID:             sessionID,
		TabID:          tabID,
		ProviderID:     options.ProviderID,
		Title:          options.Title,
		Cwd:            options.Cwd,
		Env:            copyEnv(options.Env),
		Status:         StatusStarting,
		CreatedAt:      now,
		LastActivityAt: now,
		MCPProfile:     options.MCPProfile,
		LaunchCommand:  append([]string{}, options.LaunchCommand...),
		RuntimeKind:    options.RuntimeKind,
		RuntimeVersion: 1,
	}
	m.sessions[sessionID] = entry
	m.activeTabIdx = len(m.tabs) - 1

	return entry
}

func (m *Manager) Restore(snapshot store.Snapshot) {
	m.tabs = nil
	m.sessions = make(map[string]*Session)
	m.runtimes = make(map[string]*RuntimeState)
	m.activeTabIdx = -1
	m.nextID = 0

	for _, stored := range snapshot.Tabs {
		m.nextID++
		sessionID := stored.SessionID
		if sessionID == "" {
			sessionID = stored.ID
		}

		tab := Tab{
			ID:         stored.ID,
			SessionID:  sessionID,
			Title:      stored.Title,
			ProviderID: stored.ProviderID,
			Active:     stored.ID == snapshot.ActiveTabID,
		}
		m.tabs = append(m.tabs, tab)

		entry := &Session{
			ID:             sessionID,
			TabID:          stored.ID,
			ProviderID:     stored.ProviderID,
			Title:          stored.Title,
			Cwd:            stored.Cwd,
			Env:            copyEnv(stored.Env),
			Status:         StatusStarting,
			CreatedAt:      stored.CreatedAt,
			LastActivityAt: stored.LastActivityAt,
			MCPProfile:     stored.MCPProfile,
			LaunchCommand:  append([]string{}, stored.LaunchCommand...),
			RuntimeKind:    RuntimeKind(stored.RuntimeKind),
			RuntimeVersion: 1,
		}
		m.sessions[sessionID] = entry
		if tab.Active {
			m.activeTabIdx = len(m.tabs) - 1
		}
	}

	if m.activeTabIdx < 0 && len(m.tabs) > 0 {
		m.activeTabIdx = len(m.tabs) - 1
		m.tabs[m.activeTabIdx].Active = true
	}
}

func (m *Manager) AttachRuntime(sessionID string, runtime pty.Session) {
	if session := m.sessions[sessionID]; session != nil {
		session.Status = StatusBusy
		session.LastError = ""
		session.ExitCode = nil
		session.LastActivityAt = time.Now()
	}
	m.runtimes[sessionID] = &RuntimeState{Process: runtime}
}

func (m *Manager) MarkStarted(sessionID string) {
	if session := m.sessions[sessionID]; session != nil {
		session.Status = StatusBusy
		session.LastError = ""
		session.LastActivityAt = time.Now()
	}
}

func (m *Manager) AppendOutput(sessionID string, data string) {
	session := m.sessions[sessionID]
	if session == nil || data == "" {
		return
	}

	session.Output += data
	if len(session.Output) > m.outputLimit {
		session.Output = session.Output[len(session.Output)-m.outputLimit:]
	}
	session.Status = StatusBusy
	session.LastActivityAt = time.Now()
	session.Dirty = true
	if active := m.Active(); active != nil && active.ID != sessionID {
		session.Unread = true
	}
}

func (m *Manager) MarkExited(sessionID string, exitCode int) {
	session := m.sessions[sessionID]
	if session == nil {
		return
	}

	session.Status = StatusExited
	session.LastActivityAt = time.Now()
	session.ExitCode = &exitCode
	delete(m.runtimes, sessionID)
}

func (m *Manager) MarkFailed(sessionID string, err string) {
	session := m.sessions[sessionID]
	if session == nil {
		return
	}

	session.Status = StatusError
	session.LastActivityAt = time.Now()
	session.LastError = err
	delete(m.runtimes, sessionID)
}

func (m *Manager) RestartActive() *Session {
	active := m.Active()
	if active == nil {
		return nil
	}

	active.Status = StatusStarting
	active.LastError = ""
	active.ExitCode = nil
	active.Output = ""
	active.RuntimeVersion++
	return active
}

func (m *Manager) CloseActive() *Session {
	if len(m.tabs) == 0 || m.activeTabIdx < 0 {
		return nil
	}

	activeTab := m.tabs[m.activeTabIdx]
	activeSession := m.sessions[activeTab.SessionID]
	delete(m.sessions, activeTab.SessionID)
	delete(m.runtimes, activeTab.SessionID)
	m.tabs = append(m.tabs[:m.activeTabIdx], m.tabs[m.activeTabIdx+1:]...)

	if len(m.tabs) == 0 {
		m.activeTabIdx = -1
		return activeSession
	}

	if m.activeTabIdx >= len(m.tabs) {
		m.activeTabIdx = len(m.tabs) - 1
	}

	for i := range m.tabs {
		m.tabs[i].Active = i == m.activeTabIdx
	}

	return activeSession
}

func (m *Manager) Next() {
	if len(m.tabs) == 0 {
		return
	}

	m.activeTabIdx = (m.activeTabIdx + 1) % len(m.tabs)
	for i := range m.tabs {
		m.tabs[i].Active = i == m.activeTabIdx
	}
	if active := m.Active(); active != nil {
		active.Unread = false
	}
}

func (m *Manager) Prev() {
	if len(m.tabs) == 0 {
		return
	}

	m.activeTabIdx = (m.activeTabIdx - 1 + len(m.tabs)) % len(m.tabs)
	for i := range m.tabs {
		m.tabs[i].Active = i == m.activeTabIdx
	}
	if active := m.Active(); active != nil {
		active.Unread = false
	}
}

func (m *Manager) Active() *Session {
	if len(m.tabs) == 0 || m.activeTabIdx < 0 {
		return nil
	}
	active := m.tabs[m.activeTabIdx]
	return m.sessions[active.SessionID]
}

func (m *Manager) ActiveRuntime() pty.Session {
	active := m.Active()
	if active == nil {
		return nil
	}
	runtime := m.runtimes[active.ID]
	if runtime == nil {
		return nil
	}
	return runtime.Process
}

func (m *Manager) WriteActive(data string) error {
	runtime := m.ActiveRuntime()
	if runtime == nil || data == "" {
		return nil
	}
	_, err := runtime.Write([]byte(data))
	return err
}

func (m *Manager) CloseSessionRuntime(sessionID string) error {
	runtime := m.runtimes[sessionID]
	if runtime == nil || runtime.Process == nil {
		return nil
	}
	delete(m.runtimes, sessionID)
	return runtime.Process.Close()
}

func (m *Manager) SessionByID(sessionID string) *Session {
	return m.sessions[sessionID]
}

func (m *Manager) Tabs() []Tab {
	out := make([]Tab, len(m.tabs))
	copy(out, m.tabs)
	return out
}

func (m *Manager) Sessions() []*Session {
	out := make([]*Session, 0, len(m.tabs))
	for _, tab := range m.tabs {
		if session := m.sessions[tab.SessionID]; session != nil {
			out = append(out, session)
		}
	}
	return out
}

func (m *Manager) HasTabs() bool {
	return len(m.tabs) > 0
}

func (m *Manager) ActiveTabID() string {
	if len(m.tabs) == 0 || m.activeTabIdx < 0 {
		return ""
	}
	return m.tabs[m.activeTabIdx].ID
}

func (m *Manager) OutputTail(sessionID string, lines int) []string {
	session := m.sessions[sessionID]
	if session == nil || session.Output == "" {
		return nil
	}

	all := strings.Split(strings.ReplaceAll(session.Output, "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(all))
	for _, line := range all {
		clean := strings.TrimRight(line, "\r")
		if clean == "" && len(filtered) == 0 {
			continue
		}
		filtered = append(filtered, clean)
	}
	if len(filtered) <= lines {
		return filtered
	}
	return filtered[len(filtered)-lines:]
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
