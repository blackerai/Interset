package ui

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	Background    lipgloss.Color
	Panel         lipgloss.Color
	Elevated      lipgloss.Color
	Text          lipgloss.Color
	Muted         lipgloss.Color
	Blue          lipgloss.Color
	BlueSoft      lipgloss.Color
	Orange        lipgloss.Color
	OrangeSoft    lipgloss.Color
	Success       lipgloss.Color
	Warning       lipgloss.Color
	Error         lipgloss.Color
	Divider       lipgloss.Color
	SidebarActive lipgloss.Color
}

type Styles struct {
	Palette           Palette
	App               lipgloss.Style
	BrandPrimary      lipgloss.Style
	BrandSecondary    lipgloss.Style
	Title             lipgloss.Style
	Text              lipgloss.Style
	Muted             lipgloss.Style
	Accent            lipgloss.Style
	AccentWarm        lipgloss.Style
	Success           lipgloss.Style
	Warning           lipgloss.Style
	Error             lipgloss.Style
	SidebarBox        lipgloss.Style
	SidebarTitle      lipgloss.Style
	SidebarItem       lipgloss.Style
	SidebarItemActive lipgloss.Style
	TabsBox           lipgloss.Style
	Tab               lipgloss.Style
	TabActive         lipgloss.Style
	TabGhost          lipgloss.Style
	SessionBox        lipgloss.Style
	HomeBox           lipgloss.Style
	StatusBar         lipgloss.Style
	Pill              lipgloss.Style
	PillActive        lipgloss.Style
	Center            lipgloss.Style
}

var theme Styles

func Theme() Styles {
	if theme.Palette.Background != "" {
		return theme
	}

	palette := Palette{
		Background:    lipgloss.Color("#111417"),
		Panel:         lipgloss.Color("#171B1F"),
		Elevated:      lipgloss.Color("#1B222A"),
		Text:          lipgloss.Color("#E6EDF3"),
		Muted:         lipgloss.Color("#93A1AF"),
		Blue:          lipgloss.Color("#4CC2FF"),
		BlueSoft:      lipgloss.Color("#8AD9FF"),
		Orange:        lipgloss.Color("#FFB347"),
		OrangeSoft:    lipgloss.Color("#FFD08A"),
		Success:       lipgloss.Color("#57D38C"),
		Warning:       lipgloss.Color("#FFB347"),
		Error:         lipgloss.Color("#FF6B6B"),
		Divider:       lipgloss.Color("#22303B"),
		SidebarActive: lipgloss.Color("#162A3D"),
	}

	theme = Styles{
		Palette: palette,
		App: lipgloss.NewStyle().
			Background(palette.Background).
			Foreground(palette.Text),
		BrandPrimary: lipgloss.NewStyle().
			Foreground(palette.BlueSoft).
			Bold(true),
		BrandSecondary: lipgloss.NewStyle().
			Foreground(palette.OrangeSoft).
			Bold(true),
		Title: lipgloss.NewStyle().
			Foreground(palette.Text).
			Bold(true),
		Text: lipgloss.NewStyle().
			Foreground(palette.Text),
		Muted: lipgloss.NewStyle().
			Foreground(palette.Muted),
		Accent: lipgloss.NewStyle().
			Foreground(palette.BlueSoft).
			Bold(true),
		AccentWarm: lipgloss.NewStyle().
			Foreground(palette.OrangeSoft).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(palette.Success),
		Warning: lipgloss.NewStyle().
			Foreground(palette.Warning),
		Error: lipgloss.NewStyle().
			Foreground(palette.Error),
		SidebarBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderLeft(false).
			BorderTop(false).
			BorderBottom(false).
			BorderForeground(palette.Divider).
			Padding(1, 1),
		SidebarTitle: lipgloss.NewStyle().
			Foreground(palette.BlueSoft).
			Bold(true),
		SidebarItem: lipgloss.NewStyle().
			Foreground(palette.Text).
			Padding(0, 1),
		SidebarItemActive: lipgloss.NewStyle().
			Foreground(palette.BlueSoft).
			Bold(true).
			Padding(0, 1),
		TabsBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false).
			BorderForeground(palette.Divider).
			Padding(0, 1),
		Tab: lipgloss.NewStyle().
			Foreground(palette.Muted),
		TabActive: lipgloss.NewStyle().
			Foreground(palette.BlueSoft).
			Bold(true),
		TabGhost: lipgloss.NewStyle().
			Foreground(palette.Muted),
		SessionBox: lipgloss.NewStyle().
			Padding(1, 2),
		HomeBox: lipgloss.NewStyle().
			Padding(1, 2),
		StatusBar: lipgloss.NewStyle().
			Background(palette.Elevated).
			Foreground(palette.Text).
			Padding(0, 1),
		Pill: lipgloss.NewStyle().
			Foreground(palette.Muted),
		PillActive: lipgloss.NewStyle().
			Foreground(palette.BlueSoft).
			Bold(true),
		Center: lipgloss.NewStyle().
			Align(lipgloss.Center),
	}

	return theme
}
