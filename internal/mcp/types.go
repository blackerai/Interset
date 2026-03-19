package mcp

import (
	"path/filepath"
	"strings"

	"interset/internal/config"
)

type Server struct {
	ID       string
	Enabled  bool
	ReadOnly bool
	Command  []string
	Env      map[string]string
}

type Profile struct {
	ID                    string
	DisplayName           string
	Servers               []Server
	Env                   map[string]string
	WorkspaceScopes       []string
	ReadOnly              bool
	ConfirmDestructiveOps bool
}

func DefaultProfiles() []Profile {
	return []Profile{
		{
			ID:                    "safe-default",
			DisplayName:           "Safe Default",
			ReadOnly:              true,
			ConfirmDestructiveOps: true,
		},
		{
			ID:                    "web-dev",
			DisplayName:           "Web Dev",
			ConfirmDestructiveOps: true,
			Env: map[string]string{
				"INTERSET_PROFILE": "web-dev",
			},
		},
		{
			ID:                    "backend",
			DisplayName:           "Backend",
			ConfirmDestructiveOps: true,
			Env: map[string]string{
				"INTERSET_PROFILE": "backend",
			},
		},
		{
			ID:                    "productivity",
			DisplayName:           "Productivity",
			ConfirmDestructiveOps: false,
			Env: map[string]string{
				"INTERSET_PROFILE": "productivity",
			},
		},
		{
			ID:                    "power",
			DisplayName:           "Power",
			ConfirmDestructiveOps: false,
			Env: map[string]string{
				"INTERSET_PROFILE": "power",
			},
		},
	}
}

func ResolveProfileID(cfg config.App, cwd string, tabProfileID string) string {
	profileID := cfg.DefaultMCPProfile
	bestMatchLen := -1

	for workspace, candidate := range cfg.WorkspaceProfiles {
		if workspace == "" {
			continue
		}
		cleanWorkspace := filepath.Clean(workspace)
		cleanCwd := filepath.Clean(cwd)
		if strings.HasPrefix(cleanCwd, cleanWorkspace) && len(cleanWorkspace) > bestMatchLen {
			bestMatchLen = len(cleanWorkspace)
			profileID = candidate
		}
	}

	if tabProfileID != "" {
		profileID = tabProfileID
	}

	if profileID == "" {
		return "safe-default"
	}
	return profileID
}

func ResolveProfile(profiles []Profile, id string) Profile {
	for _, profile := range profiles {
		if profile.ID == id {
			return profile
		}
	}
	for _, profile := range profiles {
		if profile.ID == "safe-default" {
			return profile
		}
	}
	return Profile{ID: id, DisplayName: id}
}

func ResolveEnv(cfg config.App, profiles []Profile, cwd string, tabProfileID string, tabEnv map[string]string, providerEnv map[string]string) (string, map[string]string) {
	profileID := ResolveProfileID(cfg, cwd, tabProfileID)
	profile := ResolveProfile(profiles, profileID)

	out := map[string]string{}
	for key, value := range profile.Env {
		out[key] = value
	}
	for key, value := range tabEnv {
		out[key] = value
	}
	for key, value := range providerEnv {
		out[key] = value
	}

	out["INTERSET_MCP_PROFILE"] = profile.ID

	return profile.ID, out
}
