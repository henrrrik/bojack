package view

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/henrrrik/bojack/internal/msg"
)

func TestRenderBackupsEmpty(t *testing.T) {
	for _, raw := range []string{``, `null`, `[]`, `{"count":0,"results":[]}`, `{"data":null}`} {
		if got := renderBackups(json.RawMessage(raw)); got != "No backups yet. Backups aren't available on dev plans." {
			t.Errorf("renderBackups(%q) = %q, want no backups", raw, got)
		}
	}
}

func TestRenderBackupsTable(t *testing.T) {
	raw := `[
		{"size": 1048576, "status": "done", "created": "2026-11-01T02:00:00Z", "name": "daily-1", "zeta": true},
		{"size": 2097152, "status": "running", "created": "2026-11-02T02:00:00Z", "name": "daily-2", "zeta": false}
	]`
	got := renderBackups(json.RawMessage(raw))
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want header + 2 rows, got:\n%s", got)
	}
	header := strings.Fields(lines[0])
	want := []string{"name", "created", "status", "size", "zeta"}
	if strings.Join(header, " ") != strings.Join(want, " ") {
		t.Errorf("header = %v, want %v", header, want)
	}
	if f := strings.Fields(lines[1]); strings.Join(f, " ") != "daily-1 2026-11-01T02:00:00Z done 1048576 true" {
		t.Errorf("row 1 = %v", f)
	}
}

func TestRenderBackupsUnwrapsResults(t *testing.T) {
	for _, raw := range []string{
		`{"count":1,"total":1,"results":[{"id":"b1","created":"2026-11-01"}]}`,
		`{"status":"ok","data":[{"id":"b1","created":"2026-11-01"}]}`,
	} {
		got := renderBackups(json.RawMessage(raw))
		if !strings.Contains(got, "b1") || !strings.Contains(got, "id") || strings.Contains(got, "count") {
			t.Errorf("renderBackups(%s) =\n%s", raw, got)
		}
	}
}

func TestRenderBackupsNestedAndMissingValues(t *testing.T) {
	raw := `[{"id":"b1","meta":{"k":"v"}},{"id":"b2"}]`
	got := renderBackups(json.RawMessage(raw))
	if !strings.Contains(got, `{"k":"v"}`) {
		t.Errorf("nested values should render as compact JSON, got:\n%s", got)
	}
	if !strings.Contains(got, "b2") || !strings.Contains(got, "-") {
		t.Errorf("missing values should render as -, got:\n%s", got)
	}
}

func TestRenderBackupsUnknownShapeFallsBackToJSON(t *testing.T) {
	raw := `{"something":"else"}`
	got := renderBackups(json.RawMessage(raw))
	if !strings.Contains(got, `"something": "else"`) {
		t.Errorf("unknown shapes should be pretty-printed, got:\n%s", got)
	}
}

func TestServiceDetailBackupsTabUsesTable(t *testing.T) {
	m := NewServiceDetailModel("my-db", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.ServiceBackupsResultMsg{Backups: json.RawMessage(`{"results":[{"id":"b1"}]}`)})
	m.activeTab = serviceBackupsTab
	v := m.View()
	if !strings.Contains(v, "b1") || strings.Contains(v, "results") {
		t.Errorf("backups tab should render a table, got:\n%s", v)
	}
}
