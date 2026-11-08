package view

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// clickOn returns a left click on the first cell of the first occurrence of
// label in the rendered view.
func clickOn(t *testing.T, rendered, label string) tea.MouseClickMsg {
	t.Helper()
	for y, line := range strings.Split(ansi.Strip(rendered), "\n") {
		if i := strings.Index(line, label); i >= 0 {
			return tea.MouseClickMsg{X: len([]rune(line[:i])), Y: y, Button: tea.MouseLeft}
		}
	}
	t.Fatalf("%q not found in:\n%s", label, ansi.Strip(rendered))
	return tea.MouseClickMsg{}
}

func TestClickingAppTabSwitchesToIt(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 120, 24)
	m, _ = m.Update(clickOn(t, m.View(), "Config"))
	if m.activeTab != appConfigTab {
		t.Errorf("activeTab = %d, want Config (%d)", m.activeTab, appConfigTab)
	}
	m, _ = m.Update(clickOn(t, m.View(), "Logs"))
	if m.activeTab != appLogsTab {
		t.Errorf("activeTab = %d, want Logs (%d)", m.activeTab, appLogsTab)
	}
}

func TestClickingServiceTabSwitchesToIt(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(clickOn(t, m.View(), "DSN"))
	if m.activeTab != serviceDSNTab {
		t.Errorf("activeTab = %d, want DSN (%d)", m.activeTab, serviceDSNTab)
	}
}

func TestClicksOffTheTabsAreIgnored(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 120, 24)
	c := clickOn(t, m.View(), "Config")
	for name, click := range map[string]tea.MouseClickMsg{
		"title row":    {X: c.X, Y: 0, Button: tea.MouseLeft},
		"past the end": {X: 119, Y: c.Y, Button: tea.MouseLeft},
		"right button": {X: c.X, Y: c.Y, Button: tea.MouseRight},
	} {
		got, _ := m.Update(click)
		if got.activeTab != appInfoTab {
			t.Errorf("%s: activeTab = %d, want it unchanged", name, got.activeTab)
		}
	}
}

func TestClickingTabsIgnoredWhileConfirming(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"}) // restart asks to confirm
	m, _ = m.Update(clickOn(t, m.View(), "DSN"))
	if m.activeTab != serviceInfoTab {
		t.Errorf("a click should not switch tabs while an action asks to confirm, got tab %d", m.activeTab)
	}
}
