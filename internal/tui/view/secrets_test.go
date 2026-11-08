package view

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func TestMaskSecrets(t *testing.T) {
	tests := []struct{ in, want string }{
		{"postgres://app:s3cr%40t@db-rw.svc:5432/app?sslmode=require", "postgres://app:********@db-rw.svc:5432/app?sslmode=require"},
		{"redis://:hunter2@cache:6379/0", "redis://:********@cache:6379/0"},
		{"mysql://user:p:a@ss@host/db", "mysql://user:********@host/db"},
		{"redis://cache:6379", "redis://cache:6379"},
		{"redis://user@cache:6379", "redis://user@cache:6379"},
		{"host=db user=app password=secret dbname=app", "host=db user=app password=******** dbname=app"},
		{"mongodb://db/app?authSource=admin&password=secret&tls=true", "mongodb://db/app?authSource=admin&password=********&tls=true"},
		{"just-a-value", "just-a-value"},
	}
	for _, tt := range tests {
		if got := maskSecrets(tt.in); got != tt.want {
			t.Errorf("maskSecrets(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func dsnModel() ServiceDetailModel {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.ServiceDSNResultMsg{DSN: runway.ServiceDSN{
		"DATABASE_URL": "postgres://app:topsecret@db:5432/app",
		"ALT_URL":      "postgres://app:topsecret@db-ro:5432/app",
	}})
	m.activeTab = serviceDSNTab
	return m
}

func TestDSNTabMasksByDefault(t *testing.T) {
	v := dsnModel().View()
	if strings.Contains(v, "topsecret") {
		t.Errorf("DSN tab should mask passwords by default, got:\n%s", v)
	}
	if !strings.Contains(v, "postgres://app:********@db:5432/app") || !strings.Contains(v, "v: reveal") {
		t.Errorf("DSN tab should show masked DSN and reveal hint, got:\n%s", v)
	}
}

func TestDSNTabRevealToggle(t *testing.T) {
	m, _ := dsnModel().Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	v := m.View()
	if !strings.Contains(v, "postgres://app:topsecret@db:5432/app") || !strings.Contains(v, "v: hide") {
		t.Errorf("v should reveal secrets, got:\n%s", v)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if strings.Contains(m.View(), "topsecret") {
		t.Error("second v should hide secrets again")
	}
}

func TestDSNTabSortedKeys(t *testing.T) {
	v := dsnModel().View()
	if a, d := strings.Index(v, "ALT_URL"), strings.Index(v, "DATABASE_URL"); a < 0 || d < 0 || a > d {
		t.Errorf("DSN keys should be sorted, got:\n%s", v)
	}
}

func TestRevealKeyIgnoredOutsideDSNTab(t *testing.T) {
	m := dsnModel()
	m.activeTab = serviceInfoTab
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	m.activeTab = serviceDSNTab
	if strings.Contains(m.View(), "topsecret") {
		t.Error("v outside the DSN tab should not reveal secrets")
	}
}
