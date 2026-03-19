package session

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusIdle     Status = "idle"
	StatusStarting Status = "starting"
	StatusBusy     Status = "busy"
	StatusError    Status = "error"
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
}

type Tab struct {
	ID         string
	SessionID  string
	Title      string
	ProviderID string
	Active     bool
}

type Manager struct {
	tabs         []Tab
	sessions     map[string]*Session
	activeTabIdx int
	nextID       int
}

func NewManager() *Manager {
	return &Manager{
		sessions:     make(map[string]*Session),
		activeTabIdx: -1,
	}
}

func (m *Manager) New(providerID, title string) {
	m.nextID++
	tabID := fmt.Sprintf("tab-%d", m.nextID)
	sessionID := fmt.Sprintf("session-%d", m.nextID)

	for i := range m.tabs {
		m.tabs[i].Active = false
	}

	m.tabs = append(m.tabs, Tab{
		ID:         tabID,
		SessionID:  sessionID,
		Title:      title,
		ProviderID: providerID,
		Active:     true,
	})

	m.sessions[sessionID] = &Session{
		ID:             sessionID,
		TabID:          tabID,
		ProviderID:     providerID,
		Title:          title + " session",
		Cwd:            "~",
		Env:            map[string]string{},
		Status:         StatusIdle,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		MCPProfile:     "safe-default",
	}

	m.activeTabIdx = len(m.tabs) - 1
}

func (m *Manager) CloseActive() {
	if len(m.tabs) == 0 || m.activeTabIdx < 0 {
		return
	}

	active := m.tabs[m.activeTabIdx]
	delete(m.sessions, active.SessionID)
	m.tabs = append(m.tabs[:m.activeTabIdx], m.tabs[m.activeTabIdx+1:]...)

	if len(m.tabs) == 0 {
		m.activeTabIdx = -1
		return
	}

	if m.activeTabIdx >= len(m.tabs) {
		m.activeTabIdx = len(m.tabs) - 1
	}

	for i := range m.tabs {
		m.tabs[i].Active = i == m.activeTabIdx
	}
}

func (m *Manager) Next() {
	if len(m.tabs) == 0 {
		return
	}

	m.activeTabIdx = (m.activeTabIdx + 1) % len(m.tabs)
	for i := range m.tabs {
		m.tabs[i].Active = i == m.activeTabIdx
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
}

func (m *Manager) Active() *Session {
	if len(m.tabs) == 0 || m.activeTabIdx < 0 {
		return nil
	}
	active := m.tabs[m.activeTabIdx]
	return m.sessions[active.SessionID]
}

func (m *Manager) Tabs() []Tab {
	out := make([]Tab, len(m.tabs))
	copy(out, m.tabs)
	return out
}

func (m *Manager) HasTabs() bool {
	return len(m.tabs) > 0
}
