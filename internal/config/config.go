package config

type App struct {
	Theme             string            `yaml:"theme" toml:"theme"`
	DefaultProvider   string            `yaml:"default_provider" toml:"default_provider"`
	DefaultMCPProfile string            `yaml:"default_mcp_profile" toml:"default_mcp_profile"`
	WorkspaceProfiles map[string]string `yaml:"workspace_profiles" toml:"workspace_profiles"`
}

func Default() App {
	return App{
		Theme:             "interset-dark",
		DefaultProvider:   "codex",
		DefaultMCPProfile: "safe-default",
		WorkspaceProfiles: map[string]string{},
	}
}
