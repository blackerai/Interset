package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"interset/internal/platform"
)

type App struct {
	Theme             string
	DefaultProvider   string
	DefaultMCPProfile string
	DefaultShell      string
	RestoreOnStartup  bool
	WorkspaceProfiles map[string]string
}

func Default() App {
	return App{
		Theme:             "interset-dark",
		DefaultProvider:   "codex",
		DefaultMCPProfile: "safe-default",
		DefaultShell:      "",
		RestoreOnStartup:  true,
		WorkspaceProfiles: map[string]string{},
	}
}

func Path() (string, error) {
	root, err := platform.EnsureIntersetHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "config.toml"), nil
}

func Load() (App, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if writeErr := Save(cfg); writeErr != nil {
				return cfg, writeErr
			}
			return cfg, nil
		}
		return cfg, err
	}

	currentSection := ""
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		key, value, ok := splitAssignment(line)
		if !ok {
			continue
		}

		value = trimQuoted(value)
		if currentSection == "workspace_profiles" {
			cfg.WorkspaceProfiles[key] = value
			continue
		}

		switch key {
		case "theme":
			cfg.Theme = value
		case "default_provider":
			cfg.DefaultProvider = value
		case "default_mcp_profile":
			cfg.DefaultMCPProfile = value
		case "default_shell":
			cfg.DefaultShell = value
		case "restore_on_startup":
			cfg.RestoreOnStartup = strings.EqualFold(value, "true")
		}
	}

	return cfg, nil
}

func Save(cfg App) error {
	path, err := Path()
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("theme = \"" + cfg.Theme + "\"\n")
	b.WriteString("default_provider = \"" + cfg.DefaultProvider + "\"\n")
	b.WriteString("default_mcp_profile = \"" + cfg.DefaultMCPProfile + "\"\n")
	b.WriteString("default_shell = \"" + cfg.DefaultShell + "\"\n")
	if cfg.RestoreOnStartup {
		b.WriteString("restore_on_startup = true\n")
	} else {
		b.WriteString("restore_on_startup = false\n")
	}

	if len(cfg.WorkspaceProfiles) > 0 {
		b.WriteString("\n[workspace_profiles]\n")
		for key, value := range cfg.WorkspaceProfiles {
			b.WriteString(key + " = \"" + value + "\"\n")
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func splitAssignment(line string) (string, string, bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

func trimQuoted(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "\"")
	value = strings.TrimSuffix(value, "\"")
	return value
}
