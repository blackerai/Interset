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
	s := Theme()

	if props.Width < 44 || props.Height < 14 {
		return s.App.Render("Interset needs a slightly larger terminal window.")
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
	return s.App.Width(props.Width).Height(props.Height).Render(lipgloss.JoinVertical(lipgloss.Left, body, status))
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
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, home)
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
	main := lipgloss.JoinVertical(lipgloss.Left, tabs, sessionView)

	if !layout.inlineSidebar {
		return main
	}

	sidebar := renderSidebar(props, layout.sidebarWidth, height, layout)
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
}

func renderHomePanel(props WorkspaceProps, width, height int, layout layoutMode) string {
	s := Theme()

	subtitle := "A premium multi-CLI workstation for AI and developer tooling"
	if layout.minimal {
		subtitle = "Multi-CLI workstation"
	}

	selected := selectedProviderLabel(props)
	heroLines := renderHeroLines(width - 4)
	actions := []string{
		"Enter open provider",
		"Ctrl+T new session",
		"Ctrl+B toggle sidebar",
		"Q quit",
	}

	lines := make([]string, 0, 16)
	lines = append(lines, heroLines...)
	lines = append(lines, "")
	lines = append(lines, renderCenteredStyledLine(s.Muted, fitText(subtitle, max(width-8, 20)), width-4))
	lines = append(lines, "")
	lines = append(lines, renderCenteredStyledLine(s.Accent, "Selected provider: "+fitText(selected, max(width-26, 8)), width-4))
	lines = append(lines, "")
	lines = append(lines, renderCenteredStyledLine(s.Text, renderPills(actions), width-4))
	lines = append(lines, "")
	lines = append(lines, renderProviderDeck(props, layout, width-8)...)

	return s.HomeBox.Width(width).Height(height).Render(renderCenteredBlock(lines, width-4, height-2))
}

func renderSidebar(props WorkspaceProps, width, height int, layout layoutMode) string {
	s := Theme()

	title := "PROVIDERS"
	subtitle := "ready to launch"
	if layout.dedicatedPanel {
		title = "INTERSET PROVIDERS"
		subtitle = "pick one and press Enter"
	}

	lines := []string{
		s.SidebarTitle.Render(title),
		s.Muted.Render(subtitle),
		"",
	}

	for i, provider := range props.Providers {
		lines = append(lines, renderSidebarRow(provider, i == props.SidebarIndex, width-4, props.SpinnerFrame))
	}

	if !layout.minimal {
		lines = append(lines, "")
		lines = append(lines, s.Muted.Render(fitText("up/down select   enter open   ctrl+b toggle", max(width-4, 12))))
	}

	content := strings.Join(lines, "\n")
	return s.SidebarBox.Width(width).Height(height).Render(lipgloss.Place(width-2, height-2, lipgloss.Left, lipgloss.Top, content))
}

func renderSidebarRow(provider registry.Provider, active bool, width int, spinnerFrame string) string {
	s := Theme()

	status := providerStatusText(provider, spinnerFrame)
	leftWidth := max(width-len(status)-3, 8)
	label := fitText(provider.DisplayName, leftWidth)
	row := padRight(provider.Symbol+" "+label, leftWidth+4) + status

	if active {
		return s.SidebarItemActive.Width(width).Render(fitText(row, width))
	}
	return s.SidebarItem.Width(width).Render(fitText(row, width))
}

func renderTabs(tabs []session.Tab, width, height int, layout layoutMode) string {
	s := Theme()

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
				pieces = append(pieces, s.TabActive.Render(label))
			} else {
				pieces = append(pieces, s.Tab.Render(label))
			}
		}
	}
	pieces = append(pieces, s.TabGhost.Render("+"))

	row := fitText(strings.Join(pieces, "  "), max(width-4, 12))
	help := "ctrl+t new  ctrl+w close  tab next  shift+tab prev"
	if layout.compact {
		help = "ctrl+t new  ctrl+w close"
	}
	lines := []string{
		s.Muted.Render(row),
		s.Muted.Render(fitText(help, max(width-4, 12))),
	}

	return s.TabsBox.Width(width).Height(height).Render(renderFixedLines(lines, width-2, height-2))
}

func renderSessionPanel(props WorkspaceProps, width, height int, layout layoutMode) string {
	s := Theme()

	if props.ActiveSession == nil {
		lines := []string{
			s.Title.Render("No live session yet"),
			s.Muted.Render("Open a provider to start."),
		}
		return s.SessionBox.Width(width).Height(height).Render(renderFixedLines(lines, width-4, height-2))
	}

	active := props.ActiveSession
	header := fmt.Sprintf("%s  ~  profile:%s", strings.ToUpper(active.ProviderID), active.MCPProfile)
	if active.Cwd != "" {
		header = fmt.Sprintf("%s  %s  profile:%s", strings.ToUpper(active.ProviderID), active.Cwd, active.MCPProfile)
	}

	lines := []string{
		s.Title.Render(fitText(active.Title, max(width-6, 12))),
		s.Muted.Render(fitText(header, max(width-6, 12))),
		"",
		s.Accent.Render("$ session viewport"),
		"",
		s.Text.Render(fitText("This pane is now reserved for the real PTY session viewport.", max(width-6, 12))),
		s.Text.Render(fitText("The shell no longer hardcodes a provider on startup.", max(width-6, 12))),
		"",
		s.Muted.Render("Session actions"),
		s.Text.Render("Enter from home or sidebar opens a provider"),
		s.Text.Render("Ctrl+T creates a new session from the selected provider"),
		s.Text.Render("Ctrl+W closes the current tab"),
	}

	if !layout.minimal {
		lines = append(lines, s.Text.Render("Ctrl+B toggles the provider surface"))
		lines = append(lines, s.Text.Render("Q exits Interset"))
	}

	return s.SessionBox.Width(width).Height(height).Render(renderFixedLines(lines, width-4, height-2))
}

func renderStatusBar(props WorkspaceProps, width int) string {
	s := Theme()

	provider := selectedProviderLabel(props)
	status := "ready"
	if props.ActiveSession != nil {
		provider = props.ActiveSession.ProviderID
		status = string(props.ActiveSession.Status)
	}

	left := fmt.Sprintf(" %s | %s | %s | %s ", props.Mode, provider, props.ActiveProfile, status)
	right := fmt.Sprintf(" %s | %s ", fitText(props.StatusNote, 28), props.Uptime)
	gap := max(width-lipgloss.Width(stripANSI(left))-lipgloss.Width(stripANSI(right))-2, 1)

	return s.StatusBar.Width(width).Render(left + strings.Repeat(" ", gap) + right)
}

func renderProviderDeck(props WorkspaceProps, layout layoutMode, width int) []string {
	if width < 20 {
		return []string{}
	}

	s := Theme()
	lines := []string{renderCenteredStyledLine(s.Muted, "Providers", width)}
	for i, provider := range props.Providers {
		text := fitText(provider.DisplayName, max(width-10, 8))
		style := s.Pill
		if i == props.SidebarIndex {
			style = s.PillActive
		}
		lines = append(lines, renderCenteredStyledLine(style, text, width))
		if layout.minimal && i >= 2 {
			break
		}
	}

	return lines
}

func renderHeroLines(width int) []string {
	s := Theme()

	hero := []string{
		" ___       _                       _   ",
		"|_ _|_ __ | |_ ___ _ __ ___  ___ | |_ ",
		" | || '_ \\| __/ _ \\ '__/ __|/ _ \\| __|",
		" | || | | | ||  __/ |  \\__ \\  __/| |_ ",
		"|___|_| |_|\\__\\___|_|  |___/\\___| \\__|",
	}

	if width < 54 {
		return []string{renderCenteredStyledLine(s.BrandPrimary, "INTERSET", width)}
	}

	lines := make([]string, 0, len(hero))
	for i, line := range hero {
		style := s.BrandPrimary
		if i == 1 || i == 3 {
			style = s.BrandSecondary
		}
		lines = append(lines, renderCenteredStyledLine(style, line, width))
	}

	return lines
}

func renderPills(items []string) string {
	s := Theme()

	rendered := make([]string, 0, len(items))
	for i, item := range items {
		style := s.Pill
		if i == 0 {
			style = s.PillActive
		}
		rendered = append(rendered, style.Render(item))
	}

	return strings.Join(rendered, "  ")
}

func resolveLayout(width int, showSidebar bool) layoutMode {
	mode := layoutMode{
		sidebarWidth: 26,
	}

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

func renderFixedLines(lines []string, width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}

	out := make([]string, 0, height)
	for _, line := range lines {
		if len(out) == height {
			break
		}
		raw := stripANSI(line)
		text := fitText(raw, width)
		padding := max(width-lipgloss.Width(text), 0)
		text += strings.Repeat(" ", padding)
		if line != raw {
			text = strings.Replace(line, raw, text, 1)
		}
		out = append(out, text)
	}

	for len(out) < height {
		out = append(out, strings.Repeat(" ", width))
	}

	return strings.Join(out, "\n")
}

func fitText(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(stripANSI(value))
	if len(runes) <= width {
		return string(runes)
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

func stripANSI(value string) string {
	var out []rune
	inANSI := false

	for _, r := range value {
		if r == '\x1b' {
			inANSI = true
			continue
		}
		if inANSI {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inANSI = false
			}
			continue
		}
		out = append(out, r)
	}

	return string(out)
}

func renderCenteredStyledLine(style lipgloss.Style, text string, width int) string {
	return style.Width(width).Align(lipgloss.Center).Render(fitText(text, width))
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
		out = append(out, line)
	}
	for len(out) < height {
		out = append(out, strings.Repeat(" ", width))
	}

	return strings.Join(out, "\n")
}

func padRight(value string, width int) string {
	gap := width - len([]rune(value))
	if gap <= 0 {
		return value
	}
	return value + strings.Repeat(" ", gap)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
