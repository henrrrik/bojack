package view

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func TestServiceDetailShowsName(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 80, 24)
	v := m.View()
	if !strings.Contains(v, "my-db") {
		t.Errorf("view should contain service name, got %q", v)
	}
}

func TestServiceDetailInfoTab(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.ServiceDetailResultMsg{
		Service: &runway.Service{
			Name: "my-db",
			Type: "postgres",
			Plan: "standard-1",
			Credentials: &runway.ServiceCredentials{
				Username: "admin",
				Password: "secret",
			},
		},
	})
	v := m.View()
	if !strings.Contains(v, "postgres") {
		t.Error("info tab should show service type")
	}
	if strings.Contains(v, "secret") {
		t.Error("info tab should not show password in plaintext")
	}
}

func TestServiceDetailTabSwitching(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 80, 24)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.activeTab != 1 {
		t.Errorf("tab = %d, want 1", m.activeTab)
	}
}

func TestServiceDetailDSNTab(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.ServiceDSNResultMsg{
		DSN: runway.ServiceDSN{"DATABASE_URL": "postgres://localhost/db"},
	})
	m.activeTab = 3
	v := m.View()
	if !strings.Contains(v, "DATABASE_URL") {
		t.Error("DSN tab should show DSN key")
	}
}
