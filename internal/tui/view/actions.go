package view

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
)

// actionSpec describes an action a view offers. Texts take the target name.
type actionSpec struct {
	kind     msg.ActionKind
	key      string // shift+letter, since lowercase keys are taken
	help     string
	question string // e.g. "Restart app %s?"
	running  string
	done     string
	failed   string
}

var (
	restartAppAction = actionSpec{msg.ActionRestartApp, "R", "R: restart",
		"Restart app %s?", "Restarting %s…", "Restarted %s.", "Restart failed"}
	restartServiceAction = actionSpec{msg.ActionRestartService, "R", "R: restart",
		"Restart service %s?", "Restarting %s…", "Restarted %s.", "Restart failed"}
	createBackupAction = actionSpec{msg.ActionCreateBackup, "B", "B: backup",
		"Create a backup of %s?", "Requesting backup of %s…", "Backup of %s requested.", "Backup failed"}
	verifyDomainAction = actionSpec{msg.ActionVerifyDomain, "V", "V: verify",
		"Re-check verification of %s?", "Checking %s…", "Verification check for %s done.", "Verification failed"}
)

type actionState int

const (
	actionIdle actionState = iota
	actionConfirming
	actionRunning
	actionDone
	actionFailed
)

// actionPrompt is the confirm-then-run flow for a view's actions. Only "y"
// confirms; any other key cancels. One action runs at a time.
type actionPrompt struct {
	specs  []actionSpec
	active actionSpec
	target string
	state  actionState
	err    error
}

func newActionPrompt(specs ...actionSpec) actionPrompt {
	return actionPrompt{specs: specs}
}

// handleKey reports whether it consumed the key. While confirming it
// consumes every key. target is what the action would apply to; "" (e.g. an
// empty list) disables starting one.
func (p actionPrompt) handleKey(k, target string) (actionPrompt, tea.Cmd, bool) {
	if p.state == actionConfirming {
		if k != "y" {
			p.state = actionIdle
			return p, nil, true
		}
		p.state = actionRunning
		req := msg.ActionRequestMsg{Kind: p.active.kind, Target: p.target}
		return p, func() tea.Msg { return req }, true
	}
	if p.state == actionRunning || target == "" {
		return p, nil, false
	}
	for _, s := range p.specs {
		if s.key == k {
			p.active, p.target, p.state = s, target, actionConfirming
			return p, nil, true
		}
	}
	return p, nil, false
}

// apply records the result of the running action.
func (p actionPrompt) apply(r msg.ActionResultMsg) actionPrompt {
	if p.state != actionRunning || r.Kind != p.active.kind || r.Target != p.target {
		return p
	}
	p.state, p.err = actionDone, r.Err
	if r.Err != nil {
		p.state = actionFailed
	}
	return p
}

// help lists the action keys.
func (p actionPrompt) help() []string {
	h := make([]string, len(p.specs))
	for i, s := range p.specs {
		h[i] = s.help
	}
	return h
}

// status renders the prompt line, or "" when there is nothing to show.
func (p actionPrompt) status(styles DetailStyles) string {
	switch p.state {
	case actionConfirming:
		return styles.StatusWarn.Render(fmt.Sprintf(p.active.question, p.target) + " (y/N)")
	case actionRunning:
		return styles.Value.Render(fmt.Sprintf(p.active.running, p.target))
	case actionDone:
		return styles.StatusUp.Render(fmt.Sprintf(p.active.done, p.target))
	case actionFailed:
		return styles.Error.Render(p.active.failed + ": " + p.err.Error())
	}
	return ""
}

// withStatus appends the status line, if any, to a view body.
func (p actionPrompt) withStatus(body string, styles DetailStyles) string {
	if s := p.status(styles); s != "" {
		return body + "\n" + s
	}
	return body
}
