package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IntersetHomeDir() (string, error) {
	if root := os.Getenv("INTERSET_HOME"); root != "" {
		return root, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "Interset"), nil
}

func EnsureIntersetHome() (string, error) {
	root, err := IntersetHomeDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}

	return root, nil
}

func ResolveDefaultShell(preferred string) []string {
	candidates := []string{}
	if preferred != "" {
		candidates = append(candidates, preferred)
	}

	if IsWindows() {
		candidates = append(candidates, "pwsh.exe", "powershell.exe", "cmd.exe")
	} else {
		candidates = append(candidates, "bash", "zsh", "sh")
	}

	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return []string{path}
		}
	}

	if IsWindows() {
		return []string{"cmd.exe"}
	}

	return []string{"sh"}
}
