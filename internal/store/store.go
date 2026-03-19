package store

import "time"

type Snapshot struct {
	Tabs         []StoredTab
	LastOpenAt   time.Time
	ActiveTabID  string
	WindowWidth  int
	WindowHeight int
}

type StoredTab struct {
	ID         string
	Title      string
	ProviderID string
	Cwd        string
	MCPProfile string
}
