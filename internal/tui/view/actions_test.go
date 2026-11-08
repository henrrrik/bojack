package view

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func key(k string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: []rune(k)[0], Text: k} }

func requestOf(t *testing.T, cmd tea.Cmd) msg.ActionRequestMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected an action request")
	}
	req, ok := cmd().(msg.ActionRequestMsg)
	if !ok {
		t.Fatalf("cmd() = %#v, want ActionRequestMsg", cmd())
	}
	return req
}

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
	if req := requestOf(t, cmd); req != (msg.ActionRequestMsg{Kind: msg.ActionRestartApp, Target: "terry"}) {
		t.Errorf("request = %+v", req)
	}
	if !strings.Contains(m.View(), "Restarting terry") {
		t.Errorf("should show progress, got:\n%s", m.View())
	}
	m, _ = m.Update(msg.ActionResultMsg{Kind: msg.ActionRestartApp, Target: "terry"})
	if !strings.Contains(m.View(), "Restarted terry.") {
		t.Errorf("should show success, got:\n%s", m.View())
	}
}

func TestActionAnyOtherKeyCancels(t *testing.T) {
	for _, k := range []string{"n", "N", "a", "x"} {
		m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
		m, _ = m.Update(key("R"))
		m, cmd := m.Update(key(k))
		if cmd != nil || strings.Contains(m.View(), "(y/N)") {
			t.Errorf("key %q should cancel without acting", k)
		}
	}
}

func TestActionTabKeyCancelsWithoutSwitching(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.activeTab != 0 || strings.Contains(m.View(), "(y/N)") {
		t.Errorf("tab while confirming should just cancel, tab = %d", m.activeTab)
	}
}

func TestActionFailureShown(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	m, _ = m.Update(key("y"))
	m, _ = m.Update(msg.ActionResultMsg{Kind: msg.ActionRestartApp, Target: "terry", Err: errors.New("app does not exist")})
	if !strings.Contains(m.View(), "Restart failed: app does not exist") {
		t.Errorf("should show failure, got:\n%s", m.View())
	}
}

func TestActionResultForOtherTargetIgnored(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	m, _ = m.Update(key("y"))
	m, _ = m.Update(msg.ActionResultMsg{Kind: msg.ActionRestartApp, Target: "other"})
	m, _ = m.Update(msg.ActionResultMsg{Kind: msg.ActionRestartService, Target: "terry"})
	if !strings.Contains(m.View(), "Restarting terry") {
		t.Errorf("results for other actions should be ignored, got:\n%s", m.View())
	}
}

func TestActionKeysIgnoredWhileRunning(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("R"))
	m, _ = m.Update(key("y"))
	m, cmd := m.Update(key("B"))
	if cmd != nil || strings.Contains(m.View(), "(y/N)") {
		t.Error("another action while one is running should not prompt")
	}
}

func TestServiceRestartAndBackup(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	v := m.View()
	if !strings.Contains(v, "R: restart") || !strings.Contains(v, "B: backup") {
		t.Errorf("help should offer restart and backup, got:\n%s", v)
	}
	m, _ = m.Update(key("R"))
	if !strings.Contains(m.View(), "Restart service my-db? (y/N)") {
		t.Errorf("got:\n%s", m.View())
	}
	_, cmd := m.Update(key("y"))
	if req := requestOf(t, cmd); req != (msg.ActionRequestMsg{Kind: msg.ActionRestartService, Target: "my-db"}) {
		t.Errorf("request = %+v", req)
	}

	m = NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(key("B"))
	if !strings.Contains(m.View(), "Create a backup of my-db? (y/N)") {
		t.Errorf("got:\n%s", m.View())
	}
	m, cmd = m.Update(key("y"))
	if req := requestOf(t, cmd); req != (msg.ActionRequestMsg{Kind: msg.ActionCreateBackup, Target: "my-db"}) {
		t.Errorf("request = %+v", req)
	}
	m, _ = m.Update(msg.ActionResultMsg{Kind: msg.ActionCreateBackup, Target: "my-db"})
	if !strings.Contains(m.View(), "Backup of my-db requested.") {
		t.Errorf("got:\n%s", m.View())
	}
}

func domainList() Viewer {
	var v Viewer = NewDomainListModel(testDetailStyles(), 120, 24)
	v, _ = v.Update(msg.ListResultMsg[runway.RootDomain]{Items: []runway.RootDomain{{Domain: "a.com"}, {Domain: "b.com"}}})
	return v
}

func TestDomainListVerifySelected(t *testing.T) {
	v := domainList()
	if !strings.Contains(v.View(), "V: verify") {
		t.Errorf("help should offer V: verify, got:\n%s", v.View())
	}
	v, _ = v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	v, cmd := v.Update(key("V"))
	if cmd != nil || !strings.Contains(v.View(), "Re-check verification of b.com? (y/N)") {
		t.Fatalf("V should confirm for the selected domain, got:\n%s", v.View())
	}
	v, cmd = v.Update(key("y"))
	if req := requestOf(t, cmd); req != (msg.ActionRequestMsg{Kind: msg.ActionVerifyDomain, Target: "b.com"}) {
		t.Errorf("request = %+v", req)
	}
	v, _ = v.Update(msg.ActionResultMsg{Kind: msg.ActionVerifyDomain, Target: "b.com"})
	if !strings.Contains(v.View(), "Verification check for b.com done.") {
		t.Errorf("got:\n%s", v.View())
	}
}

func TestListWithoutActionsIgnoresActionKeys(t *testing.T) {
	v := update(newThingList(), msg.ListResultMsg[thing]{Items: []thing{{"alpha"}}})
	if _, cmd := v.Update(key("V")); cmd != nil || strings.Contains(v.View(), "(y/N)") {
		t.Error("a list without actions should ignore action keys")
	}
}
