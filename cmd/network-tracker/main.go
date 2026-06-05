package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eskylake/network-tracker/internal/app"
	"github.com/eskylake/network-tracker/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(app.New(cfg), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "network-tracker failed: %v\n", err)
		os.Exit(1)
	}
}
