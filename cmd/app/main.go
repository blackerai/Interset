package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"interset/internal/app"
)

func main() {
	program := tea.NewProgram(
		app.New(),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		log.Fatalf("interset exited with error: %v", err)
	}
}
