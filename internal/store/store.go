package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"interset/internal/platform"
)

type Snapshot struct {
	Tabs         []StoredTab `json:"tabs"`
	LastOpenAt   time.Time   `json:"last_open_at"`
	ActiveTabID  string      `json:"active_tab_id"`
	WindowWidth  int         `json:"window_width"`
	WindowHeight int         `json:"window_height"`
}

type StoredTab struct {
	ID             string            `json:"id"`
	SessionID      string            `json:"session_id"`
	Title          string            `json:"title"`
	ProviderID     string            `json:"provider_id"`
	Cwd            string            `json:"cwd"`
	MCPProfile     string            `json:"mcp_profile"`
	LaunchCommand  []string          `json:"launch_command"`
	RuntimeKind    string            `json:"runtime_kind"`
	Env            map[string]string `json:"env"`
	CreatedAt      time.Time         `json:"created_at"`
	LastActivityAt time.Time         `json:"last_activity_at"`
}

func Path() (string, error) {
	root, err := platform.EnsureIntersetHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "state.json"), nil
}

func Load() (Snapshot, error) {
	path, err := Path()
	if err != nil {
		return Snapshot{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Snapshot{}, nil
		}
		return Snapshot{}, err
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func Save(snapshot Snapshot) error {
	path, err := Path()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
