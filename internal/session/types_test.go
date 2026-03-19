package session

import (
	"testing"
	"time"

	"interset/internal/store"
)

func TestRestoreUsesStoredSessionID(t *testing.T) {
	manager := NewManager()
	now := time.Now()

	manager.Restore(store.Snapshot{
		ActiveTabID: "tab-9",
		Tabs: []store.StoredTab{
			{
				ID:             "tab-9",
				SessionID:      "session-9",
				Title:          "Shell",
				ProviderID:     "shell",
				Cwd:            ".",
				MCPProfile:     "safe-default",
				LaunchCommand:  []string{"cmd.exe"},
				RuntimeKind:    "shell",
				CreatedAt:      now,
				LastActivityAt: now,
			},
		},
	})

	active := manager.Active()
	if active == nil {
		t.Fatal("expected restored active session")
	}

	if active.ID != "session-9" {
		t.Fatalf("expected restored session id %q, got %q", "session-9", active.ID)
	}

	if manager.SessionByID("session-9") == nil {
		t.Fatal("expected session lookup by stored session id to succeed")
	}
}

func TestRestartActiveAdvancesRuntimeVersion(t *testing.T) {
	manager := NewManager()
	session := manager.Create(CreateOptions{
		ProviderID:    "shell",
		Title:         "Shell",
		Cwd:           ".",
		MCPProfile:    "safe-default",
		LaunchCommand: []string{"cmd.exe"},
		RuntimeKind:   RuntimeShell,
	})

	session.Output = "hello"
	session.LastError = "old"
	session.ExitCode = new(int)
	*session.ExitCode = 1

	restarted := manager.RestartActive()
	if restarted == nil {
		t.Fatal("expected active session to restart")
	}

	if restarted.RuntimeVersion != 2 {
		t.Fatalf("expected runtime version 2, got %d", restarted.RuntimeVersion)
	}

	if restarted.Output != "" {
		t.Fatalf("expected output to be cleared, got %q", restarted.Output)
	}

	if restarted.LastError != "" {
		t.Fatalf("expected error to be cleared, got %q", restarted.LastError)
	}

	if restarted.ExitCode != nil {
		t.Fatal("expected exit code to be cleared")
	}
}
