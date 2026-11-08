package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
)

// restartKey starts a restart; it is shift+r because r is refresh.
const restartKey = "R"

type restartState int

const (
	restartIdle restartState = iota
	restartConfirming
	restartRunning
	restartDone
	restartFailed
)

// restartPrompt is the confirm-then-restart flow shared by the app and
// service detail views. Only "y" confirms; any other key cancels.
type restartPrompt struct {
	name    string
	service bool
	state   restartState
	err     error
}

func newRestartPrompt(name string, service bool) restartPrompt {
	return restartPrompt{name: name, service: service}
}

// handleKey reports whether it consumed the key. While confirming it
// consumes every key, so nothing else happens until the prompt is answered.
func (p restartPrompt) handleKey(k string) (restartPrompt, tea.Cmd, bool) {
	switch {
	case p.state == restartConfirming && k == "y":
		p.state = restartRunning
		req := msg.RestartRequestMsg{Name: p.name, Service: p.service}
		return p, func() tea.Msg { return req }, true
	case p.state == restartConfirming:
		p.state = restartIdle
		return p, nil, true
	case k == restartKey && p.state != restartRunning:
		p.state = restartConfirming
		return p, nil, true
	}
	return p, nil, false
}

// apply records the result of a restart of this prompt's target.
func (p restartPrompt) apply(r msg.RestartResultMsg) restartPrompt {
	if r.Name != p.name || r.Service != p.service || p.state != restartRunning {
		return p
	}
	p.state, p.err = restartDone, r.Err
	if r.Err != nil {
		p.state = restartFailed
	}
	return p
}

// status renders the prompt line, or "" when there is nothing to show.
func (p restartPrompt) status(styles DetailStyles) string {
	kind := "app"
	if p.service {
		kind = "service"
	}
	switch p.state {
	case restartConfirming:
		return styles.StatusWarn.Render("Restart " + kind + " " + p.name + "? (y/N)")
	case restartRunning:
		return styles.Value.Render("Restarting " + p.name + "…")
	case restartDone:
		return styles.StatusUp.Render("Restarted " + p.name + ".")
	case restartFailed:
		return styles.Error.Render("Restart failed: " + p.err.Error())
	}
	return ""
}

// withStatus appends the restart status line, if any, to a view body.
func (p restartPrompt) withStatus(body string, styles DetailStyles) string {
	if s := p.status(styles); s != "" {
		return body + "\n" + s
	}
	return body
}
