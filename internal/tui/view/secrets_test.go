package view

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	m, _ := dsnModel().Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	v := m.View()
	if !strings.Contains(v, "postgres://app:topsecret@db:5432/app") || !strings.Contains(v, "v: hide") {
		t.Errorf("v should reveal secrets, got:\n%s", v)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
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
	m, _ = m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	m.activeTab = serviceDSNTab
	if strings.Contains(m.View(), "topsecret") {
		t.Error("v outside the DSN tab should not reveal secrets")
	}
}

func TestIsSecretKey(t *testing.T) {
	tests := map[string]bool{
		"SECRET_KEY_BASE": true, "STRIPE_SECRET_KEY": true, "API_TOKEN": true, "DB_PASSWORD": true,
		"DB_PASS": true, "SMTP_PWD": true, "AUTH_TOKEN": true, "JWT_SALT": true, "PRIVATE_KEY": true,
		"AWS_ACCESS_KEY_ID": true, "GITHUB_CREDENTIALS": true, "apiKey": true, "session-secret": true,
		"TS_AUTHKEY": true, "OPENAI_APIKEY": true, "GHTOKEN": true, "OAUTH_CLIENTSECRET": true, "AUTH_SECRET": true,
		"PORT": false, "DATABASE_URL": false, "KEYBOARD_LAYOUT": false, "AUTH_URL": false,
		"LOG_LEVEL": false, "PASSENGER_COUNT": false, "": false,
	}
	for key, want := range tests {
		if got := isSecretKey(key); got != want {
			t.Errorf("isSecretKey(%q) = %v, want %v", key, got, want)
		}
	}
}

func TestMaskConfigValue(t *testing.T) {
	tests := []struct{ key, value, want string }{
		{"API_TOKEN", "abc123", secretMask},
		{"API_TOKEN", "", ""},
		{"DATABASE_URL", "postgres://u:pw@h/db", "postgres://u:********@h/db"},
		{"PORT", "8080", "8080"},
	}
	for _, tt := range tests {
		if got := maskConfigValue(tt.key, tt.value); got != tt.want {
			t.Errorf("maskConfigValue(%q, %q) = %q, want %q", tt.key, tt.value, got, tt.want)
		}
	}
}

var testConfig = map[string]string{
	"API_TOKEN":    "tok_live_123",
	"DATABASE_URL": "postgres://app:dbpw@db:5432/app",
	"PORT":         "8080",
}

func vKey() tea.KeyPressMsg { return tea.KeyPressMsg{Code: 'v', Text: "v"} }

func assertMaskedConfig(t *testing.T, v string) {
	t.Helper()
	if strings.Contains(v, "tok_live_123") || strings.Contains(v, "dbpw") {
		t.Errorf("config tab should mask secrets, got:\n%s", v)
	}
	if !strings.Contains(v, "8080") || !strings.Contains(v, "postgres://app:********@db:5432/app") || !strings.Contains(v, "v: reveal") {
		t.Errorf("config tab should show non-secret values and reveal hint, got:\n%s", v)
	}
}

func TestAppConfigTabMasksAndReveals(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.AppConfigResultMsg{Config: testConfig})
	m.activeTab = appConfigTab
	assertMaskedConfig(t, m.View())
	m, _ = m.Update(vKey())
	if v := m.View(); !strings.Contains(v, "tok_live_123") || !strings.Contains(v, "dbpw") || !strings.Contains(v, "v: hide") {
		t.Errorf("v should reveal config secrets, got:\n%s", v)
	}
}

func TestAppRevealKeyIgnoredOutsideConfigTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.AppConfigResultMsg{Config: testConfig})
	m, _ = m.Update(vKey()) // on Info tab
	m.activeTab = appConfigTab
	assertMaskedConfig(t, m.View())
}

func TestServiceConfigTabMasksAndReveals(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.ServiceConfigResultMsg{Config: testConfig})
	m.activeTab = serviceConfigTab
	assertMaskedConfig(t, m.View())
	m, _ = m.Update(vKey())
	if v := m.View(); !strings.Contains(v, "tok_live_123") {
		t.Errorf("v should reveal service config secrets, got:\n%s", v)
	}
}
