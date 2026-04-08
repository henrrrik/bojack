package view

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func testDetailStyles() DetailStyles {
	return DetailStyles{
		Title:       lipgloss.NewStyle(),
		TabActive:   lipgloss.NewStyle().Bold(true),
		TabInactive: lipgloss.NewStyle(),
		Error:       lipgloss.NewStyle(),
		HelpBar:     lipgloss.NewStyle(),
		Label:       lipgloss.NewStyle().Width(16),
		Value:       lipgloss.NewStyle(),
		StatusUp:    lipgloss.NewStyle(),
		StatusDown:  lipgloss.NewStyle(),
	}
}

func TestAppDetailShowsAppName(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	v := m.View()
	if !strings.Contains(v, "my-app") {
		t.Errorf("view should contain app name, got %q", v)
	}
}

func TestAppDetailInfoTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppDetailResultMsg{
		Detail: &runway.AppDetail{
			App:      runway.App{ID: "my-app", Plan: "standard-1", Region: "eu-north-1"},
			Settings: runway.AppSettings{Routable: true},
		},
	})
	v := m.View()
	if !strings.Contains(v, "standard-1") {
		t.Error("info tab should show plan")
	}
}

func TestAppDetailTabSwitching(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	if m.activeTab != 0 {
		t.Fatalf("initial tab = %d, want 0", m.activeTab)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.activeTab != 1 {
		t.Errorf("tab = %d, want 1", m.activeTab)
	}
	// Shift+tab goes back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.activeTab != 0 {
		t.Errorf("tab = %d, want 0", m.activeTab)
	}
	// Wraps around backwards
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.activeTab != len(appTabs)-1 {
		t.Errorf("tab = %d, want %d", m.activeTab, len(appTabs)-1)
	}
}

func TestAppDetailProcessesTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppProcessesResultMsg{
		Processes: []runway.Process{
			{Name: "web.1", Type: "web", State: "up", Release: "v42"},
		},
	})
	m.activeTab = 1
	v := m.View()
	if !strings.Contains(v, "web.1") {
		t.Error("processes tab should show process name")
	}
}

func TestAppDetailConfigTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppConfigResultMsg{
		Config: runway.AppConfig{"PORT": "8080"},
	})
	m.activeTab = 3
	v := m.View()
	if !strings.Contains(v, "PORT") {
		t.Error("config tab should show config key")
	}
}

func TestAppDetailErrorTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppDetailResultMsg{Err: errTest})
	v := m.View()
	if !strings.Contains(v, "Error") {
		t.Error("should show error on info tab")
	}
}
