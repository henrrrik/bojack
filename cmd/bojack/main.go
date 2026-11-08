package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/bojack/internal/runway"
	"github.com/henrrrik/bojack/internal/tui"
)

const usage = `Usage: bojack [--version]

A terminal UI for Runway. Needs the runway CLI on your PATH, logged in.`

func main() {
	opts, err := parseArgs(os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println(usage)
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "bojack: %v\n%s\n", err, usage)
		os.Exit(2)
	}
	if opts.version {
		info, _ := debug.ReadBuildInfo()
		fmt.Println("bojack", resolveVersion(version, info))
		return
	}
	executor := runway.NewExecCommandExecutor("runway")
	runner := runway.NewCLIRunner(executor)
	model := tui.NewModel(runner)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
