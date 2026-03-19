package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"interset/internal/registry"
	"interset/internal/session"
)

type WorkspaceProps struct {
	Width         int
	Height        int
	Mode          string
	ShowSidebar   bool
	Focus         int
	Providers     []registry.Provider
	SidebarIndex  int
	Tabs          []session.Tab
	ActiveSession *session.Session
	ActiveOutput  []string
	ActiveProfile string
	StatusNote    string
	SpinnerFrame  string
	Uptime        time.Duration
}

type layoutMode struct {
	compact        bool
	minimal        bool
	inlineSidebar  bool
	dedicatedPanel bool
	sidebarWidth   int
}

func RenderWorkspace(props WorkspaceProps) string {
	if props.Width < 44 || props.Height < 14 {
		return "Interset needs a slightly larger terminal window."
	}

	layout := resolveLayout(props.Width, props.ShowSidebar)
	statusHeight := 1
	bodyHeight := max(props.Height-statusHeight, 1)

	body := ""
	if layout.dedicatedPanel {
		body = renderSidebar(props, props.Width, bodyHeight, layout)
	} else if props.Mode == "home" {
		body = renderHome(props, layout, bodyHeight)
	} else {
		body = renderWorkspaceMode(props, layout, bodyHeight)
	}

	status := renderStatusBar(props, props.Width)
	return body + "\n" + status
}

func renderHome(props WorkspaceProps, layout layoutMode, height int) string {
	mainWidth := props.Width
	if layout.inlineSidebar {
		mainWidth -= layout.sidebarWidth
	}

	home := renderHomePanel(props, mainWidth, height, layout)
	if !layout.inlineSidebar {
		return home
	}

	sidebar := renderSidebar(props, layout.sidebarWidth, height, layout)
	return joinColumns(sidebar, home, layout.sidebarWidth, mainWidth, height)
}

func renderWorkspaceMode(props WorkspaceProps, layout layoutMode, height int) string {
	mainWidth := props.Width
	if layout.inlineSidebar {
		mainWidth -= layout.sidebarWidth
	}

	tabsHeight := 3
	contentHeight := max(height-tabsHeight, 1)
	tabs := renderTabs(props.Tabs, mainWidth, tabsHeight, layout)
	sessionView := renderSessionPanel(props, mainWidth, contentHeight, layout)
	main := tabs + "\n" + sessionView

	if !layout.inlineSidebar {
		return main
	}

	sidebar := renderSidebar(props, layout.sidebarWidth, height, layout)
	return joinColumns(sidebar, main, layout.sidebarWidth, mainWidth, height)
}

func renderHomePanel(props WorkspaceProps, width, height int, layout layoutMode) string {
	subtitle := "A premium multi-CLI workstation for AI and developer tooling"
	if layout.minimal {
		subtitle = "Multi-CLI workstation"
	}

	selected := selectedProviderLabel(props)
	lines := make([]string, 0, 20)
	lines = append(lines, renderHeroLines(width)...)
	lines = append(lines, "")
	lines = append(lines, centerText(fitText(subtitle, width), width))
	lines = append(lines, "")
	lines = append(lines, centerText("Selected provider: "+fitText(selected, max(width-20, 8)), width))
	lines = append(lines, "")
	lines = append(lines, centerText(renderActions(), width))
	lines = append(lines, "")
	lines = append(lines, renderProviderDeck(props, layout, width)...)

	return renderCenteredBlock(lines, width, height)
}

func renderSidebar(props WorkspaceProps, width, height int, layout layoutMode) string {
	title := "PROVIDERS"
	subtitle := "ready to launch"
	if layout.dedicatedPanel {
		title = "INTERSET PROVIDERS"
		subtitle = "pick one and press Enter"
	}

	lines := []string{
		title,
		subtitle,
		"",
	}

	for i, provider := range props.Providers {
		lines = append(lines, renderSidebarRow(provider, i == props.SidebarIndex, width-2, props.SpinnerFrame))
	}

	if !layout.minimal {
		lines = append(lines, "")
		lines = append(lines, fitText("up/down select  enter open  ctrl+b toggle", max(width-2, 12)))
	}

	return renderTopBlock(lines, width, height, true)
}

func renderSidebarRow(provider registry.Provider, active bool, width int, spinnerFrame string) string {
	status := providerStatusText(provider, spinnerFrame)
	prefix := "  "
	if active {
		prefix = "> "
	}
	leftWidth := max(width-len(status)-len(prefix)-1, 8)
	label := fitText(provider.DisplayName, leftWidth)
	return fitText(prefix+padRight(provider.Symbol+" "+label, leftWidth+2)+" "+status, width)
}

func renderTabs(tabs []session.Tab, width, height int, layout layoutMode) string {
	pieces := []string{"tabs"}
	if len(tabs) == 0 {
		pieces = append(pieces, "no sessions")
	} else {
		labelWidth := 18
		if layout.compact {
			labelWidth = 12
		}
		for _, tab := range tabs {
			label := fitText(tab.Title, labelWidth)
			if tab.Active {
				pieces = append(pieces, "["+label+"]")
			} else {
				pieces = append(pieces, label)
			}
		}
	}
	pieces = append(pieces, "+")

	row := fitText(strings.Join(pieces, "  "), max(width, 12))
	help := "ctrl+t new  ctrl+w close  tab next  shift+tab prev"
	if layout.compact {
		help = "ctrl+t new  ctrl+w close"
	}

	lines := []string{
		row,
		fitText(help, max(width, 12)),
		strings.Repeat("-", min(width, 80)),
	}
	return renderTopBlock(lines, width, height, false)
}

func renderSessionPanel(props WorkspaceProps, width, height int, layout layoutMode) string {
	lines := make([]string, 0, 14)
	if props.ActiveSession == nil {
		lines = append(lines, "No live session yet")
		lines = append(lines, "Open a provider to start.")
		return renderTopBlock(lines, width, height, false)
	}

	active := props.ActiveSession
	header := fmt.Sprintf("%s  ~  profile:%s", strings.ToUpper(active.ProviderID), active.MCPProfile)
	if active.Cwd != "" {
		header = fmt.Sprintf("%s  %s  profile:%s", strings.ToUpper(active.ProviderID), active.Cwd, active.MCPProfile)
	}

	lines = append(lines, fitText(active.Title, width))
	lines = append(lines, fitText(header, width))
	lines = append(lines, fitText("status: "+string(active.Status), width))
	if len(active.LaunchCommand) > 0 {
		lines = append(lines, fitText("command: "+strings.Join(active.LaunchCommand, " "), width))
	}
	if active.LastError != "" {
		lines = append(lines, fitText("error: "+active.LastError, width))
	}
	lines = append(lines, "")

	if len(props.ActiveOutput) == 0 {
		lines = append(lines, "$ waiting for session output")
		lines = append(lines, "")
		lines = append(lines, fitText("Type directly in this pane to send input to the active session.", width))
		lines = append(lines, fitText("Ctrl+R restarts the current runtime and Ctrl+W closes the tab.", width))
		if !layout.minimal {
			lines = append(lines, fitText("Providers open from the home screen or sidebar and stay alive per tab.", width))
		}
		return renderTopBlock(lines, width, height, false)
	}

	lines = append(lines, "$ live session")
	lines = append(lines, "")
	for _, line := range props.ActiveOutput {
		lines = append(lines, fitText(line, width))
	}

	return renderTopBlock(lines, width, height, false)
}

func renderStatusBar(props WorkspaceProps, width int) string {
	provider := selectedProviderLabel(props)
	profile := props.ActiveProfile
	status := "ready"
	if props.ActiveSession != nil {
		provider = props.ActiveSession.ProviderID
		profile = props.ActiveSession.MCPProfile
		status = string(props.ActiveSession.Status)
	}

	left := fmt.Sprintf(" %s | %s | %s | %s ", props.Mode, provider, profile, status)
	right := fmt.Sprintf(" %s | %s ", fitText(props.StatusNote, 28), props.Uptime)
	gap := max(width-len(left)-len(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

func renderProviderDeck(props WorkspaceProps, layout layoutMode, width int) []string {
	if width < 20 {
		return []string{}
	}

	lines := []string{centerText("Providers", width)}
	for i, provider := range props.Providers {
		text := provider.DisplayName
		if i == props.SidebarIndex {
			text = "> " + text
		}
		lines = append(lines, centerText(fitText(text, width), width))
		if layout.minimal && i >= 2 {
			break
		}
	}
	return lines
}

func renderHeroLines(width int) []string {
	hero := []string{
		" ___       _                       _   ",
		"|_ _|_ __ | |_ ___ _ __ ___  ___ | |_ ",
		" | || '_ \\| __/ _ \\ '__/ __|/ _ \\| __|",
		" | || | | | ||  __/ |  \\__ \\  __/| |_ ",
		"|___|_| |_|\\__\\___|_|  |___/\\___| \\__|",
	}

	if width < 54 {
		return []string{centerText("INTERSET", width)}
	}

	lines := make([]string, 0, len(hero))
	for _, line := range hero {
		lines = append(lines, centerText(line, width))
	}
	return lines
}

func renderActions() string {
	return "Enter open provider  Ctrl+T new session  Ctrl+B toggle sidebar  Q quit"
}

func resolveLayout(width int, showSidebar bool) layoutMode {
	mode := layoutMode{sidebarWidth: 26}
	switch {
	case width >= 120:
		mode.inlineSidebar = showSidebar
	case width >= 96:
		mode.compact = true
		mode.sidebarWidth = 20
		mode.inlineSidebar = showSidebar
	default:
		mode.compact = true
		mode.minimal = true
		mode.dedicatedPanel = showSidebar
	}
	return mode
}

func selectedProviderLabel(props WorkspaceProps) string {
	if len(props.Providers) == 0 {
		return "none"
	}
	index := props.SidebarIndex
	if index < 0 || index >= len(props.Providers) {
		index = 0
	}
	return props.Providers[index].DisplayName
}

func providerStatusText(provider registry.Provider, spinnerFrame string) string {
	switch provider.Status {
	case registry.StatusIdle:
		return "idle"
	case registry.StatusStarting:
		return spinnerFrame + " start"
	case registry.StatusBusy:
		return spinnerFrame + " busy"
	case registry.StatusExited:
		return "exit"
	case registry.StatusAuthRequired:
		return "auth"
	case registry.StatusError:
		return "error"
	case registry.StatusMissing:
		return "missing"
	default:
		return "detect"
	}
}

func renderCenteredBlock(lines []string, width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}

	if len(lines) > height {
		lines = lines[:height]
	}

	topPad := max((height-len(lines))/2, 0)
	out := make([]string, 0, height)
	for i := 0; i < topPad; i++ {
		out = append(out, strings.Repeat(" ", width))
	}
	for _, line := range lines {
		out = append(out, padRight(fitText(line, width), width))
	}
	for len(out) < height {
		out = append(out, strings.Repeat(" ", width))
	}
	return strings.Join(out, "\n")
}

func renderTopBlock(lines []string, width, height int, divider bool) string {
	if width < 1 || height < 1 {
		return ""
	}

	out := make([]string, 0, height)
	for _, line := range lines {
		if len(out) == height {
			break
		}
		text := fitText(line, width)
		if divider {
			text = padRight(text, max(width-1, 1)) + verticalDivider()
		} else {
			text = padRight(text, width)
		}
		out = append(out, text)
	}

	fill := strings.Repeat(" ", width)
	if divider {
		fill = strings.Repeat(" ", max(width-1, 1)) + verticalDivider()
	}
	for len(out) < height {
		out = append(out, fill)
	}

	return strings.Join(out, "\n")
}

func joinColumns(left, right string, leftWidth, rightWidth, height int) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	out := make([]string, 0, height)

	for i := 0; i < height; i++ {
		l := strings.Repeat(" ", leftWidth)
		r := strings.Repeat(" ", rightWidth)
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		out = append(out, l+r)
	}

	return strings.Join(out, "\n")
}

func centerText(text string, width int) string {
	trimmed := fitText(text, width)
	visibleWidth := lipgloss.Width(trimmed)
	if visibleWidth >= width {
		return trimmed
	}
	left := (width - visibleWidth) / 2
	right := width - visibleWidth - left
	return strings.Repeat(" ", left) + trimmed + strings.Repeat(" ", right)
}

func verticalDivider() string {
	return "|"
}

func fitText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

func padRight(value string, width int) string {
	gap := width - len([]rune(value))
	if gap <= 0 {
		return value
	}
	return value + strings.Repeat(" ", gap)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
