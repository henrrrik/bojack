package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/runway-tui/internal/runway"
	"github.com/henrrrik/runway-tui/internal/tui"
)

func main() {
	executor := runway.NewExecCommandExecutor("runway")
	runner := runway.NewCLIRunner(executor)
	model := tui.NewModel(runner)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
