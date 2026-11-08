package view

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// DefaultLogOptions shows all log types over the full retention period.
var DefaultLogOptions = runway.LogOptions{Limit: runway.DefaultLogLimit}

// logSinceOptions are the time ranges the logs tab cycles through; 0 means
// the CLI's full retention period (two weeks).
var logSinceOptions = []time.Duration{0, time.Hour, 6 * time.Hour, 24 * time.Hour, 7 * 24 * time.Hour}

var logLimitOptions = []int{100, 500, 1000}

func nextLogSince(d time.Duration) time.Duration {
	for i, o := range logSinceOptions {
		if o == d {
			return logSinceOptions[(i+1)%len(logSinceOptions)]
		}
	}
	return logSinceOptions[0]
}

func logSinceLabel(d time.Duration) string {
	switch {
	case d == 0:
		return "all"
	case d%(24*time.Hour) == 0 && d >= 48*time.Hour:
		return strconv.Itoa(int(d/(24*time.Hour))) + "d"
	default:
		return strconv.Itoa(int(d/time.Hour)) + "h"
	}
}

func nextLogLimit(n int) int {
	for i, o := range logLimitOptions {
		if o == n {
			return logLimitOptions[(i+1)%len(logLimitOptions)]
		}
	}
	return logLimitOptions[0]
}

// updateLogOptions applies a logs-tab key to opts. ok is false if the key
// is not a logs key.
func updateLogOptions(opts runway.LogOptions, key string) (runway.LogOptions, bool) {
	switch key {
	case "a":
		opts.NoRuntime = !opts.NoRuntime
	case "b":
		opts.NoBuild = !opts.NoBuild
	case "s":
		opts.NoSystem = !opts.NoSystem
	case "t":
		opts.Since = nextLogSince(opts.Since)
	case "l":
		opts.Limit = nextLogLimit(opts.Limit)
	default:
		return opts, false
	}
	return opts, true
}

// requestLogOptions returns a command asking the root model to refetch logs.
func requestLogOptions(opts runway.LogOptions) tea.Cmd {
	return func() tea.Msg { return msg.LogOptionsMsg{Options: opts} }
}

func checkbox(on bool, label string) string {
	if on {
		return "[x] " + label
	}
	return "[ ] " + label
}

// logOptionsSummary renders the logs tab header, e.g.
// "[x] app  [x] build  [ ] system   Range: 24h   Limit: 500".
func logOptionsSummary(o runway.LogOptions) string {
	return checkbox(!o.NoRuntime, "app") + "  " + checkbox(!o.NoBuild, "build") + "  " +
		checkbox(!o.NoSystem, "system") + "   Range: " + logSinceLabel(o.Since) +
		"   Limit: " + strconv.Itoa(o.Limit)
}
