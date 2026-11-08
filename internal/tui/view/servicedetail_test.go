package view

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/bojack/internal/msg"
	"github.com/henrrrik/bojack/internal/runway"
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
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.activeTab != 1 {
		t.Errorf("tab = %d, want 1", m.activeTab)
	}
}

func TestServiceDetailDSNTab(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.ServiceDSNResultMsg{
		DSN: runway.ServiceDSN{"DATABASE_URL": "postgres://localhost/db"},
	})
	m.activeTab = 4
	v := m.View()
	if !strings.Contains(v, "DATABASE_URL") {
		t.Error("DSN tab should show DSN key")
	}
}

func TestServiceDetailVolumeSizeInMB(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.ServiceDetailResultMsg{
		Service: &runway.Service{
			Name:    "my-db",
			Volumes: []runway.ServiceVolume{{Name: "data", Size: "5120", Mountpoint: "/data"}},
		},
	})
	v := m.View()
	if !strings.Contains(v, "5.0 GB") {
		t.Errorf("volume size should be shown as 5.0 GB, got %q", v)
	}
	if strings.Contains(v, "bytes") {
		t.Errorf("volume size is in MB, not bytes, got %q", v)
	}
}

func TestServiceDetailStatsTab(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.ServiceStatsResultMsg{
		Stats: runway.Stats{
			"db_tps": {Title: "Transactions", Unit: "tps", Data: []runway.StatsPoint{{Value: 1}, {Value: 2}}},
		},
		Interval: DefaultStatsInterval,
	})
	m.activeTab = 1
	v := m.View()
	if !strings.Contains(v, "Stats") || !strings.Contains(v, "Transactions") || !strings.Contains(v, "2 tps") {
		t.Errorf("service stats tab should render metrics, got:\n%s", v)
	}
	if !intervalLineRE.MatchString(v) {
		t.Errorf("service stats tab should show interval, got:\n%s", v)
	}
}

func TestServiceStatsIntervalKey(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 80, 24)
	m.activeTab = 1
	m, cmd := m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	if m.StatsInterval() != "168h" || cmd == nil {
		t.Fatalf("interval = %q, cmd = %v; want 168h and a fetch command", m.StatsInterval(), cmd)
	}
	if got, ok := cmd().(msg.StatsIntervalMsg); !ok || got.Interval != "168h" {
		t.Errorf("cmd() = %#v, want StatsIntervalMsg{168h}", got)
	}
}
