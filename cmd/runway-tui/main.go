package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/runway"
	"github.com/henrrrik/runway-tui/internal/tui"
)

func main() {
	executor := runway.NewExecCommandExecutor("runway")
	runner := runway.NewCLIRunner(executor)
	model := tui.NewModel(runner)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
