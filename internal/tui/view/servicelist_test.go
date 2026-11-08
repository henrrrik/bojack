package view

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func TestServiceListLoading(t *testing.T) {
	var m Viewer = NewServiceListModel(testDetailStyles(), 80, 24)
	v := m.View()
	if !strings.Contains(v, "Fetching services") {
		t.Error("loading view should say 'Fetching services'")
	}
}

func TestServiceListReceivesData(t *testing.T) {
	var m Viewer = NewServiceListModel(testDetailStyles(), 80, 24)
	services := []runway.Service{
		{Name: "my-db", Type: "postgres", Plan: "standard-1", HasBackup: true},
	}
	m, _ = m.Update(msg.ListResultMsg[runway.Service]{Items: services})
	v := m.View()
	if !strings.Contains(v, "my-db") {
		t.Errorf("view should contain service name, got %q", v)
	}
}

func TestServiceListSelectEmitsNavigate(t *testing.T) {
	var m Viewer = NewServiceListModel(testDetailStyles(), 80, 24)
	services := []runway.Service{
		{Name: "my-db", Type: "postgres", Plan: "s-1"},
	}
	m, _ = m.Update(msg.ListResultMsg[runway.Service]{Items: services})
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

func TestServiceListEmpty(t *testing.T) {
	var m Viewer = NewServiceListModel(testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.ListResultMsg[runway.Service]{Items: []runway.Service{}})
	v := m.View()
	if !strings.Contains(v, "No services found") {
		t.Error("should show empty message")
	}
}
