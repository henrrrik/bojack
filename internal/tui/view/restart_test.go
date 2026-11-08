package view

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
)

func key(k string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)} }

func TestRestartConfirmFlow(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	if !strings.Contains(m.View(), "R: restart") {
		t.Error("help should offer R: restart")
	}
	m, cmd := m.Update(key("R"))
	if cmd != nil {
		t.Fatal("R should only ask for confirmation")
	}
	if !strings.Contains(m.View(), "Restart app terry? (y/N)") {
		t.Errorf("should ask for confirmation, got:\n%s", m.View())
	}
	m, cmd = m.Update(key("y"))
	if cmd == nil {
		t.Fatal("y should request the restart")
	}
	if got, ok := cmd().(msg.RestartRequestMsg); !ok || got.Name != "terry" || got.Service {
		t.Errorf("cmd() = %#v, want app restart request for terry", cmd())
	}
	if !strings.Contains(m.View(), "Restarting terry") {
		t.Errorf("should show progress, got:\n%s", m.View())
	}
	m, _ = m.Update(msg.RestartResultMsg{Name: "terry"})
	if !strings.Contains(m.View(), "Restarted terry.") {
		t.Errorf("should show success, got:\n%s", m.View())
	}
}

func TestRestartAnyOtherKeyCancels(t *testing.T) {
	for _, k := range []string{"n", "N", "a", "x"} {
		m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
		m, _ = m.Update(key("R"))
		m, cmd := m.Update(key(k))
		if cmd != nil || strings.Contains(m.View(), "Restart app terry?") {
			t.Errorf("key %q should cancel without restarting", k)
		}
	}
}

func TestRestartTabKeyCancelsWithoutSwitching(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.activeTab != 0 || strings.Contains(m.View(), "(y/N)") {
		t.Errorf("tab while confirming should just cancel, tab = %d", m.activeTab)
	}
}

func TestRestartFailureShown(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	m, _ = m.Update(key("y"))
	m, _ = m.Update(msg.RestartResultMsg{Name: "terry", Err: errors.New("app does not exist")})
	if !strings.Contains(m.View(), "Restart failed: app does not exist") {
		t.Errorf("should show failure, got:\n%s", m.View())
	}
}

func TestRestartResultForOtherTargetIgnored(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	m, _ = m.Update(key("y"))
	m, _ = m.Update(msg.RestartResultMsg{Name: "other"})
	m, _ = m.Update(msg.RestartResultMsg{Name: "terry", Service: true})
	if !strings.Contains(m.View(), "Restarting terry") {
		t.Errorf("results for other targets should be ignored, got:\n%s", m.View())
	}
}

func TestRestartKeyIgnoredWhileRunning(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	m, _ = m.Update(key("y"))
	m, cmd := m.Update(key("R"))
	if cmd != nil || strings.Contains(m.View(), "(y/N)") {
		t.Error("R while a restart is running should not prompt again")
	}
}

func TestServiceRestartRequest(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	if !strings.Contains(m.View(), "Restart service my-db? (y/N)") {
		t.Errorf("should ask for confirmation, got:\n%s", m.View())
	}
	_, cmd := m.Update(key("y"))
	if got, ok := cmd().(msg.RestartRequestMsg); !ok || got.Name != "my-db" || !got.Service {
		t.Errorf("cmd() = %#v, want service restart request for my-db", cmd())
	}
}
