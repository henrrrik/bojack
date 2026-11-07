package view

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func TestAppListLoading(t *testing.T) {
	m := NewAppListModel(lipgloss.NewStyle(), 80, 24)
	v := m.View()
	if !strings.Contains(v, "Fetching apps") {
		t.Errorf("loading view should say 'Fetching apps', got %q", v)
	}
}

func TestAppListReceivesData(t *testing.T) {
	m := NewAppListModel(lipgloss.NewStyle(), 80, 24)
	apps := []runway.AppListEntry{
		{
			App: runway.App{
				ID:     "test-app",
				Type:   "app",
				Plan:   "standard-1",
				Region: "eu-north-1",
				Meta:   runway.AppMeta{Role: "owner"},
			},
		},
	}
	m, _ = m.Update(msg.AppListResultMsg{Apps: apps})
	if m.loading {
		t.Error("should not be loading after result")
	}
	v := m.View()
	if !strings.Contains(v, "test-app") {
		t.Errorf("view should contain app name, got %q", v)
	}
}

func TestAppListError(t *testing.T) {
	m := NewAppListModel(lipgloss.NewStyle(), 80, 24)
	m, _ = m.Update(msg.AppListResultMsg{Err: errTest})
	v := m.View()
	if !strings.Contains(v, "Error") {
		t.Errorf("view should show error, got %q", v)
	}
}

func TestAppListEmpty(t *testing.T) {
	m := NewAppListModel(lipgloss.NewStyle(), 80, 24)
	m, _ = m.Update(msg.AppListResultMsg{Apps: []runway.AppListEntry{}})
	v := m.View()
	if !strings.Contains(v, "No apps found") {
		t.Errorf("view should show empty message, got %q", v)
	}
}

func TestAppListSelectEmitsNavigate(t *testing.T) {
	m := NewAppListModel(lipgloss.NewStyle(), 80, 24)
	apps := []runway.AppListEntry{
		{App: runway.App{ID: "my-app", Type: "app", Plan: "s-1", Region: "eu", Meta: runway.AppMeta{Role: "owner"}}},
	}
	m, _ = m.Update(msg.AppListResultMsg{Apps: apps})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on enter")
	}
	result := cmd()
	nav, ok := result.(msg.NavigateMsg)
	if !ok {
		t.Fatalf("expected NavigateMsg, got %T", result)
	}
	if nav.View != msg.ViewAppDetail {
		t.Errorf("View = %d, want ViewAppDetail", nav.View)
	}
	if nav.Param != "my-app" {
		t.Errorf("Param = %q, want %q", nav.Param, "my-app")
	}
}

func TestAppListSelectDbNavigatesToServiceDetail(t *testing.T) {
	m := NewAppListModel(lipgloss.NewStyle(), 80, 24)
	apps := []runway.AppListEntry{
		{App: runway.App{ID: "my-db", Type: "db", Plan: "s-1", Region: "eu", Meta: runway.AppMeta{Role: "owner"}}},
	}
	m, _ = m.Update(msg.AppListResultMsg{Apps: apps})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on enter")
	}
	result := cmd()
	nav, ok := result.(msg.NavigateMsg)
	if !ok {
		t.Fatalf("expected NavigateMsg, got %T", result)
	}
	if nav.View != msg.ViewServiceDetail {
		t.Errorf("View = %d, want ViewServiceDetail", nav.View)
	}
	if nav.Param != "my-db" {
		t.Errorf("Param = %q, want %q", nav.Param, "my-db")
	}
}

var errTest = &testError{}

type testError struct{}

func (e *testError) Error() string { return "test error" }

func TestAppListShowsStatus(t *testing.T) {
	m := NewAppListModel(lipgloss.NewStyle(), 120, 24)
	apps := []runway.AppListEntry{
		{App: runway.App{ID: "up-app", Type: "app"}, Status: runway.AppStatus{Status: "running"}},
		{App: runway.App{ID: "db-app", Type: "db"}},
	}
	m, _ = m.Update(msg.AppListResultMsg{Apps: apps})
	v := m.View()
	if !strings.Contains(v, "Status") {
		t.Errorf("view should have a Status column, got %q", v)
	}
	if !strings.Contains(v, "running") {
		t.Errorf("view should show app status, got %q", v)
	}
}
