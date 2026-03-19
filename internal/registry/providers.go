package registry

type Status string

const (
	StatusUnknown      Status = "unknown"
	StatusMissing      Status = "missing"
	StatusIdle         Status = "idle"
	StatusStarting     Status = "starting"
	StatusBusy         Status = "busy"
	StatusAuthRequired Status = "auth-required"
	StatusError        Status = "error"
)

type Provider struct {
	ID            string
	DisplayName   string
	Symbol        string
	DetectCommand []string
	LaunchCommand []string
	InstallHint   string
	AuthHint      string
	SupportsMCP   bool
	Status        Status
}

func DefaultProviders() []Provider {
	return []Provider{
		{
			ID:            "codex",
			DisplayName:   "Codex",
			Symbol:        "[C]",
			DetectCommand: []string{"codex", "--help"},
			LaunchCommand: []string{"codex"},
			InstallHint:   "Install via OpenAI tooling instructions when provider detection lands.",
			AuthHint:      "Requires an authenticated OpenAI session.",
			SupportsMCP:   true,
			Status:        StatusIdle,
		},
		{
			ID:            "gemini",
			DisplayName:   "Gemini CLI",
			Symbol:        "[G]",
			DetectCommand: []string{"gemini", "--help"},
			LaunchCommand: []string{"gemini"},
			InstallHint:   "Install Gemini CLI and make sure it is present in PATH.",
			AuthHint:      "Requires Google auth or API credentials.",
			SupportsMCP:   true,
			Status:        StatusStarting,
		},
		{
			ID:            "warp",
			DisplayName:   "Warp CLI",
			Symbol:        "[W]",
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
			DetectCommand: []string{"resend", "--help"},
			LaunchCommand: []string{"resend"},
			InstallHint:   "Install Resend CLI and export required credentials.",
			AuthHint:      "Requires API credentials.",
			SupportsMCP:   false,
			Status:        StatusAuthRequired,
		},
		{
			ID:            "linear",
			DisplayName:   "Linear CLI",
			Symbol:        "[L]",
			DetectCommand: []string{"linear", "--help"},
			LaunchCommand: []string{"linear"},
			InstallHint:   "Install Linear CLI from the official package source.",
			AuthHint:      "Requires workspace auth.",
			SupportsMCP:   false,
			Status:        StatusBusy,
		},
		{
			ID:            "crush",
			DisplayName:   "Crush",
			Symbol:        "[X]",
			DetectCommand: []string{"crush", "--help"},
			LaunchCommand: []string{"crush"},
			InstallHint:   "Install Charmbracelet Crush and expose it in PATH.",
			AuthHint:      "Provider-specific auth flow.",
			SupportsMCP:   true,
			Status:        StatusMissing,
		},
	}
}
