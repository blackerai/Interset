package mcp

import (
	"path/filepath"
	"testing"

	"interset/internal/config"
)

func TestResolveEnvMergeOrder(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultMCPProfile = "safe-default"
	cfg.WorkspaceProfiles = map[string]string{
		filepath.Clean(`C:\work`): "backend",
	}

	profiles := []Profile{
		{
			ID:  "backend",
			Env: map[string]string{"SHARED": "profile", "PROFILE_ONLY": "yes"},
		},
		{
			ID:  "web-dev",
			Env: map[string]string{"SHARED": "web"},
		},
	}

	profileID, env := ResolveEnv(
		cfg,
		profiles,
		filepath.Clean(`C:\work\repo`),
		"web-dev",
		map[string]string{"SHARED": "tab", "TAB_ONLY": "yes"},
		map[string]string{"SHARED": "provider", "PROVIDER_ONLY": "yes"},
	)

	if profileID != "web-dev" {
		t.Fatalf("expected tab override profile, got %q", profileID)
	}

	if env["INTERSET_MCP_PROFILE"] != "web-dev" {
		t.Fatalf("expected INTERSET_MCP_PROFILE to be set, got %q", env["INTERSET_MCP_PROFILE"])
	}

	if env["SHARED"] != "provider" {
		t.Fatalf("expected provider env to win, got %q", env["SHARED"])
	}

	if env["TAB_ONLY"] != "yes" {
		t.Fatalf("expected tab env to survive merge, got %q", env["TAB_ONLY"])
	}

	if env["PROVIDER_ONLY"] != "yes" {
		t.Fatalf("expected provider env to survive merge, got %q", env["PROVIDER_ONLY"])
	}
}
