package view

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

type thing struct{ name string }

func newThingList() Viewer {
	return NewListModel("things", []table.Column{{Title: "Name", Width: 10}},
		func(t thing) table.Row { return table.Row{t.name} }, lipgloss.NewStyle(), 80, 24)
}

func update(v Viewer, m any) Viewer {
	v, _ = v.Update(m)
	return v
}

func TestListLoading(t *testing.T) {
	if v := newThingList().View(); !strings.Contains(v, "Fetching things...") {
		t.Errorf("got %q", v)
	}
}

func TestListRendersRows(t *testing.T) {
	v := update(newThingList(), msg.ListResultMsg[thing]{Items: []thing{{"alpha"}, {"beta"}}})
	out := v.View()
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") || !strings.Contains(out, "r: refresh") {
		t.Errorf("got %q", out)
	}
}

func TestListEmpty(t *testing.T) {
	v := update(newThingList(), msg.ListResultMsg[thing]{})
	if out := v.View(); !strings.Contains(out, "No things found.") {
		t.Errorf("got %q", out)
	}
}

func TestListError(t *testing.T) {
	v := update(newThingList(), msg.ListResultMsg[thing]{Err: errors.New("boom")})
	if out := v.View(); !strings.Contains(out, "Error: boom") {
		t.Errorf("got %q", out)
	}
}

func TestListIgnoresOtherResultTypes(t *testing.T) {
	v := update(newThingList(), msg.ListResultMsg[runway.Template]{Items: []runway.Template{{Name: "x"}}})
	if out := v.View(); !strings.Contains(out, "Fetching things...") {
		t.Errorf("result for another list type should be ignored, got %q", out)
	}
}

func TestTemplateRow(t *testing.T) {
	row := templateRow(runway.Template{Name: "valkey", Category: "cache", Latest: "0.2.0", MinPlan: "free-launch", Description: "KV store"})
	want := table.Row{"valkey", "cache", "0.2.0", "free-launch", "KV store"}
	if strings.Join(row, "|") != strings.Join(want, "|") {
		t.Errorf("row = %v, want %v", row, want)
	}
}

func TestRegistryRow(t *testing.T) {
	row := registryRow(runway.Registry{Hostname: "https://r.example.com/org", Organization: "org", Username: "robot", Managed: true})
	want := table.Row{"https://r.example.com/org", "org", "robot", "yes"}
	if strings.Join(row, "|") != strings.Join(want, "|") {
		t.Errorf("row = %v, want %v", row, want)
	}
}

func TestSSHKeyRow(t *testing.T) {
	row := sshKeyRow(runway.SSHKey{
		ID:      "laptop",
		Created: "2026-04-03T18:00:13Z",
		Public:  "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPYs8cCE74KXqXajjjp+FLgaQVmH6NvBu6HO93INnLG8",
	})
	want := table.Row{"laptop", "ssh-ed25519", "SHA256:r+gxGXVllVUsUaraxwnBUoSxkln6pT3fsFka+nNg+6g", "2026-04-03"}
	if strings.Join(row, "|") != strings.Join(want, "|") {
		t.Errorf("row = %v, want %v", row, want)
	}
}

func TestSSHFingerprintInvalid(t *testing.T) {
	for _, in := range []string{"", "ssh-ed25519", "ssh-ed25519 !!notbase64!!"} {
		if got := sshFingerprint(in); got != "-" {
			t.Errorf("sshFingerprint(%q) = %q, want -", in, got)
		}
	}
}

func TestAPIKeyRow(t *testing.T) {
	row := apiKeyRow(runway.APIKey{Name: "ci", Status: "ACTIVE", Created: "2026-09-01T19:29:29.98926476Z",
		Expires: "2027-09-01T19:29:29Z", LastUsed: ""})
	want := table.Row{"ci", "ACTIVE", "2026-09-01", "2027-09-01", "-"}
	if strings.Join(row, "|") != strings.Join(want, "|") {
		t.Errorf("row = %v, want %v", row, want)
	}
}

func selectableThings() Viewer {
	m := NewListModel("things", []table.Column{{Title: "Name", Width: 10}},
		func(t thing) table.Row { return table.Row{t.name} }, lipgloss.NewStyle(), 80, 24).
		WithSelect(func(t thing) tea.Msg { return msg.NavigateMsg{View: msg.ViewAppDetail, Param: t.name} })
	return update(m, msg.ListResultMsg[thing]{Items: []thing{{"alpha"}, {"beta"}}})
}

func TestListSelectUsesCursorItem(t *testing.T) {
	v := update(selectableThings(), tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command on enter")
	}
	if nav, ok := cmd().(msg.NavigateMsg); !ok || nav.Param != "beta" {
		t.Errorf("cmd() = %#v, want navigate to beta", cmd())
	}
}

func TestListWithoutSelectIgnoresEnter(t *testing.T) {
	v := update(newThingList(), msg.ListResultMsg[thing]{Items: []thing{{"alpha"}}})
	if _, cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Errorf("enter without a select handler should do nothing, got %#v", cmd())
	}
	if strings.Contains(v.View(), "enter: details") {
		t.Error("help should not offer enter without a select handler")
	}
}

func TestListSelectHelp(t *testing.T) {
	if !strings.Contains(selectableThings().View(), "enter: details") {
		t.Error("help should offer enter when selectable")
	}
}
