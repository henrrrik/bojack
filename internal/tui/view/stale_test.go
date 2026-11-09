package view

import (
	"strings"
	"testing"

	"github.com/henrrrik/bojack/internal/msg"
	"github.com/henrrrik/bojack/internal/runway"
)

// Results carry the name they were fetched for; a view for another resource
// (opened after the fetch started) must ignore them.

func TestAppDetailIgnoresAnotherAppsResults(t *testing.T) {
	m := NewAppDetailModel("app-b", testDetailStyles(), 100, 30)
	m, _ = m.Update(msg.AppConfigResultMsg{Resource: msg.Resource{Name: "app-a"}, Config: runway.AppConfig{"FROM_A": "1"}})
	if m.loaded[appConfigTab] || len(m.config) != 0 {
		t.Errorf("app-b took app-a's config: %v", m.config)
	}
	m, _ = m.Update(msg.AppConfigResultMsg{Resource: msg.Resource{Name: "app-b"}, Config: runway.AppConfig{"FROM_B": "1"}})
	if !m.loaded[appConfigTab] || m.config["FROM_B"] != "1" {
		t.Errorf("app-b should take its own config, got %v", m.config)
	}
}

func TestServiceDetailIgnoresAnotherServicesResults(t *testing.T) {
	m := NewServiceDetailModel("db-b", testDetailStyles(), 100, 30)
	m, _ = m.Update(msg.ServiceBackupsResultMsg{Resource: msg.Resource{Name: "db-a"}, Backups: []byte(`[{"id":"from-a"}]`)})
	m.activeTab = serviceBackupsTab
	if strings.Contains(m.View(), "from-a") {
		t.Errorf("db-b shows db-a's backups:\n%s", m.View())
	}
}

func TestVolumeDetailIgnoresAnotherVolume(t *testing.T) {
	var v Viewer = NewVolumeDetailModel("vol-b", testDetailStyles(), 100, 30)
	v, _ = v.Update(msg.ItemResultMsg[runway.Volume]{Resource: msg.Resource{Name: "vol-a"},
		Item: &runway.Volume{Name: "vol-a", SizeMB: 1024, Resizable: true}})
	if out := v.View(); strings.Contains(out, "vol-a") || !strings.Contains(out, "Fetching volume") {
		t.Errorf("vol-b took vol-a's data:\n%s", out)
	}
}
