package view

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/bojack/internal/msg"
)

// actionSpec describes an action a view offers. Its texts are format strings
// given the target name and, for actions with input, the checked value as
// shown to the user (use %[1]s and %[2]s).
type actionSpec struct {
	kind     msg.ActionKind
	key      string // shift+letter, since lowercase keys are taken
	help     string
	question string // e.g. "Restart app %s?"
	running  string
	done     string
	failed   string
	// input, if set, is the label of a value the user types before
	// confirming.
	input string
	// check validates the typed value and returns how to show it in the
	// texts, e.g. only the key of KEY=value.
	check func(string) (string, error)
	// echo, if set, renders the value while it is typed, e.g. masked.
	echo func(string) string
}

var (
	restartAppAction = actionSpec{kind: msg.ActionRestartApp, key: "R", help: "R: restart",
		question: "Restart app %s?", running: "Restarting %s…", done: "Restarted %s.", failed: "Restart failed"}
	restartServiceAction = actionSpec{kind: msg.ActionRestartService, key: "R", help: "R: restart",
		question: "Restart service %s?", running: "Restarting %s…", done: "Restarted %s.", failed: "Restart failed"}
	createBackupAction = actionSpec{kind: msg.ActionCreateBackup, key: "B", help: "B: backup",
		question: "Create a backup of %s?", running: "Requesting backup of %s…", done: "Backup of %s requested.",
		failed: "Backup failed"}
	verifyDomainAction = actionSpec{kind: msg.ActionVerifyDomain, key: "V", help: "V: verify",
		question: "Re-check verification of %s?", running: "Checking %s…", done: "Verification check for %s done.",
		failed: "Verification failed"}
)

type actionState int

const (
	actionIdle actionState = iota
	actionInput
	actionConfirming
	actionRunning
	actionDone
	actionFailed
)

// actionPrompt is the (input-then-)confirm-then-run flow for a view's
// actions. Only "y" confirms; any other key cancels. One action runs at a
// time.
type actionPrompt struct {
	specs    []actionSpec
	active   actionSpec
	target   string
	input    string // the value being typed, or the raw value once checked
	shown    string // the checked value as shown in texts
	inputErr error
	state    actionState
	err      error
}

func newActionPrompt(specs ...actionSpec) actionPrompt {
	return actionPrompt{specs: specs}
}

// claimsKeys reports whether the prompt needs every key: while typing a
// value or confirming.
func (p actionPrompt) claimsKeys() bool {
	return p.state == actionInput || p.state == actionConfirming
}

// handleKey reports whether it consumed the key. While typing or confirming
// it consumes every key. target is what the action would apply to; ""
// (e.g. an empty list) disables starting one. extra are actions offered in
// addition to the prompt's own, e.g. for the active tab.
func (p actionPrompt) handleKey(km tea.KeyPressMsg, target string, extra ...actionSpec) (actionPrompt, tea.Cmd, bool) {
	switch p.state {
	case actionInput:
		return p.inputKey(km), nil, true
	case actionConfirming:
		if km.String() != "y" {
			p.state = actionIdle
			return p, nil, true
		}
		p.state = actionRunning
		req := msg.ActionRequestMsg{Kind: p.active.kind, Target: p.target, Arg: p.input}
		return p, func() tea.Msg { return req }, true
	case actionRunning:
		return p, nil, false
	}
	if target == "" {
		return p, nil, false
	}
	for _, s := range append(p.specs[:len(p.specs):len(p.specs)], extra...) {
		if s.key == km.String() {
			return p.start(s, target), nil, true
		}
	}
	return p, nil, false
}

func (p actionPrompt) start(s actionSpec, target string) actionPrompt {
	p.active, p.target, p.input, p.shown, p.inputErr = s, target, "", "", nil
	p.state = actionConfirming
	if s.input != "" {
		p.state = actionInput
	}
	return p
}

// inputKey edits the typed value; enter checks it and moves on to confirming.
func (p actionPrompt) inputKey(km tea.KeyPressMsg) actionPrompt {
	switch km.Code {
	case tea.KeyEscape:
		p.state = actionIdle
	case tea.KeyEnter:
		return p.submit()
	case tea.KeyBackspace:
		if r := []rune(p.input); len(r) > 0 {
			p.input = string(r[:len(r)-1])
		}
	default:
		p.input += km.Text
	}
	return p
}

// paste appends pasted text to the value being typed, if any.
func (p actionPrompt) paste(pm tea.PasteMsg) actionPrompt {
	if p.state == actionInput {
		p.input += pm.Content
	}
	return p
}

func (p actionPrompt) submit() actionPrompt {
	shown, err := p.input, error(nil)
	if p.active.check != nil {
		shown, err = p.active.check(p.input)
	}
	if err != nil {
		p.inputErr = err
		return p
	}
	p.shown, p.inputErr, p.state = shown, nil, actionConfirming
	return p
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

// help lists the action keys, including extra ones.
func (p actionPrompt) help(extra ...actionSpec) []string {
	var h []string
	for _, s := range append(p.specs[:len(p.specs):len(p.specs)], extra...) {
		h = append(h, s.help)
	}
	return h
}

// text formats one of the active action's texts.
func (p actionPrompt) text(format string) string {
	if p.active.input == "" {
		return fmt.Sprintf(format, p.target)
	}
	return fmt.Sprintf(format, p.target, p.shown)
}

// status renders the prompt line, or "" when there is nothing to show.
func (p actionPrompt) status(styles DetailStyles) string {
	switch p.state {
	case actionInput:
		return p.inputStatus(styles)
	case actionConfirming:
		return styles.StatusWarn.Render(p.text(p.active.question) + " (y/N)")
	case actionRunning:
		return styles.Value.Render(p.text(p.active.running))
	case actionDone:
		return styles.StatusUp.Render(p.text(p.active.done))
	case actionFailed:
		return styles.Error.Render(p.active.failed + ": " + p.err.Error())
	}
	return ""
}

func (p actionPrompt) inputStatus(styles DetailStyles) string {
	typed := p.input
	if p.active.echo != nil {
		typed = p.active.echo(typed)
	}
	s := styles.StatusWarn.Render(p.active.input+": ") + styles.Value.Render(typed+"█") +
		"  " + styles.HelpBar.Render("enter: continue | esc: cancel")
	if p.inputErr != nil {
		s += "\n" + styles.Error.Render(p.inputErr.Error())
	}
	return s
}

// withStatus appends the status line, if any, to a view body.
func (p actionPrompt) withStatus(body string, styles DetailStyles) string {
	if s := p.status(styles); s != "" {
		return body + "\n" + s
	}
	return body
}
