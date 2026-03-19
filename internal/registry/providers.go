package registry

import "os/exec"

type Status string

const (
	StatusUnknown      Status = "unknown"
	StatusMissing      Status = "missing"
	StatusIdle         Status = "idle"
	StatusStarting     Status = "starting"
	StatusBusy         Status = "busy"
	StatusAuthRequired Status = "auth-required"
	StatusError        Status = "error"
	StatusExited       Status = "exited"
)

type Provider struct {
	ID            string
	DisplayName   string
	Symbol        string
	Executable    string
	DetectCommand []string
	LaunchCommand []string
	InstallHint   string
	AuthHint      string
	SupportsMCP   bool
	Status        Status
	DetectedPath  string
	LastError     string
}

func DefaultProviders() []Provider {
	return []Provider{
		{
			ID:            "codex",
			DisplayName:   "Codex",
			Symbol:        "[C]",
			Executable:    "codex",
			DetectCommand: []string{"codex", "--help"},
			LaunchCommand: []string{"codex"},
			InstallHint:   "Install via OpenAI tooling instructions when provider detection lands.",
			AuthHint:      "Requires an authenticated OpenAI session.",
			SupportsMCP:   true,
			Status:        StatusUnknown,
		},
		{
			ID:            "gemini",
			DisplayName:   "Gemini CLI",
			Symbol:        "[G]",
			Executable:    "gemini",
			DetectCommand: []string{"gemini", "--help"},
			LaunchCommand: []string{"gemini"},
			InstallHint:   "Install Gemini CLI and make sure it is present in PATH.",
			AuthHint:      "Requires Google auth or API credentials.",
			SupportsMCP:   true,
			Status:        StatusUnknown,
		},
		{
			ID:            "warp",
			DisplayName:   "Warp CLI",
			Symbol:        "[W]",
			Executable:    "warp",
			DetectCommand: []string{"warp", "--help"},
			LaunchCommand: []string{"warp"},
			InstallHint:   "Warp CLI availability depends on platform and local installation.",
			AuthHint:      "Provider-specific auth flow.",
			SupportsMCP:   false,
			Status:        StatusUnknown,
		},
		{
			ID:            "resend",
			DisplayName:   "Resend CLI",
			Symbol:        "[R]",
			Executable:    "resend",
			DetectCommand: []string{"resend", "--help"},
			LaunchCommand: []string{"resend"},
			InstallHint:   "Install Resend CLI and export required credentials.",
			AuthHint:      "Requires API credentials.",
			SupportsMCP:   false,
			Status:        StatusUnknown,
		},
		{
			ID:            "linear",
			DisplayName:   "Linear CLI",
			Symbol:        "[L]",
			Executable:    "linear",
			DetectCommand: []string{"linear", "--help"},
			LaunchCommand: []string{"linear"},
			InstallHint:   "Install Linear CLI from the official package source.",
			AuthHint:      "Requires workspace auth.",
			SupportsMCP:   false,
			Status:        StatusUnknown,
		},
		{
			ID:            "crush",
			DisplayName:   "Crush",
			Symbol:        "[X]",
			Executable:    "crush",
			DetectCommand: []string{"crush", "--help"},
			LaunchCommand: []string{"crush"},
			InstallHint:   "Install Charmbracelet Crush and expose it in PATH.",
			AuthHint:      "Provider-specific auth flow.",
			SupportsMCP:   true,
			Status:        StatusUnknown,
		},
	}
}

func DetectProviders(providers []Provider) []Provider {
	out := make([]Provider, len(providers))
	for i, provider := range providers {
		out[i] = DetectProvider(provider)
	}
	return out
}

func DetectProvider(provider Provider) Provider {
	target := provider.Executable
	if target == "" && len(provider.LaunchCommand) > 0 {
		target = provider.LaunchCommand[0]
	}

	if target == "" {
		provider.Status = StatusError
		provider.LastError = "missing executable metadata"
		return provider
	}

	path, err := exec.LookPath(target)
	if err != nil {
		provider.Status = StatusMissing
		provider.DetectedPath = ""
		provider.LastError = err.Error()
		return provider
	}

	provider.Status = StatusIdle
	provider.DetectedPath = path
	provider.LastError = ""
	return provider
}
