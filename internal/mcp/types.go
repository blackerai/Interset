package mcp

type Server struct {
	ID       string
	Enabled  bool
	ReadOnly bool
	Command  []string
	Env      map[string]string
}

type Profile struct {
	ID                     string
	DisplayName            string
	Servers                []Server
	Env                    map[string]string
	WorkspaceScopes        []string
	ReadOnly               bool
	ConfirmDestructiveOps  bool
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
		},
		{
			ID:                    "backend",
			DisplayName:           "Backend",
			ConfirmDestructiveOps: true,
		},
		{
			ID:                    "productivity",
			DisplayName:           "Productivity",
			ConfirmDestructiveOps: false,
		},
		{
			ID:                    "power",
			DisplayName:           "Power",
			ConfirmDestructiveOps: false,
		},
	}
}
