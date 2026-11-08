package view

import (
	"strings"
	"testing"

	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// tableLines returns the tab body's lines between the tab bar and the help line, split into fields.
func tabFields(t *testing.T, v, header string) [][]string {
	t.Helper()
	lines := strings.Split(v, "\n")
	for i, l := range lines {
		if strings.Join(strings.Fields(l), " ") == header {
			var rows [][]string
			for _, r := range lines[i+1:] {
				if f := strings.Fields(r); len(f) > 0 && !strings.Contains(r, "esc: back") {
					rows = append(rows, f)
				}
			}
			return rows
		}
	}
	t.Fatalf("header %q not found in:\n%s", header, v)
	return nil
}

func TestProcessesTable(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.AppProcessesResultMsg{Processes: []runway.Process{
		{Name: "terry-cmd-6fddfd6cd-p2zrl", Type: "cmd", Release: "v7", State: "up", Started: "2026-10-13T19:33:12UTC"},
		{Name: "terry-cmd-x", Type: "cmd", Release: "v6", State: "crashed", Started: "2026-10-12T10:00:00UTC"},
	}})
	m.activeTab = 1
	rows := tabFields(t, m.View(), "Name Type Release State Started")
	if len(rows) != 2 || strings.Join(rows[0], " ") != "terry-cmd-6fddfd6cd-p2zrl cmd v7 up 2026-10-13T19:33:12UTC" ||
		rows[1][3] != "crashed" {
		t.Errorf("rows = %q", rows)
	}
}

func TestDomainsTable(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.AppDomainsResultMsg{Domains: []runway.Domain{
		{Domain: "terry.pqapp.dev", Path: "/", TargetPath: "/"},
		{Domain: "www.example.com", Path: "/api", TargetPath: "/", Deletable: true,
			Root: runway.DomainRoot{Domain: "example.com", Verified: true}},
	}})
	m.activeTab = 4
	rows := tabFields(t, m.View(), "Domain Path Target Root Verified Removable")
	want := []string{"terry.pqapp.dev / / - - no", "www.example.com /api / example.com yes yes"}
	if len(rows) != 2 || strings.Join(rows[0], " ") != want[0] || strings.Join(rows[1], " ") != want[1] {
		t.Errorf("rows = %q, want %q", rows, want)
	}
}

func TestSharingTable(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.AppSharesResultMsg{Shares: &runway.AppShareList{Data: []runway.AppShare{
		{User: "henrrrik", Role: "owner"}, {User: "alice", Role: "developer"},
	}}})
	m.activeTab = 5
	rows := tabFields(t, m.View(), "User Role")
	if len(rows) != 2 || strings.Join(rows[1], " ") != "alice developer" {
		t.Errorf("rows = %q", rows)
	}
}

func TestDetailTablesEmptyStates(t *testing.T) {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 24)
	m, _ = m.Update(msg.AppProcessesResultMsg{})
	m, _ = m.Update(msg.AppDomainsResultMsg{})
	m, _ = m.Update(msg.AppSharesResultMsg{})
	for tab, want := range map[int]string{1: "No processes running.", 4: "No domains configured.", 5: "No shares configured."} {
		m.activeTab = tab
		if !strings.Contains(m.View(), want) {
			t.Errorf("tab %d should say %q", tab, want)
		}
	}
}

func TestTextTableKeepsStyledCellsAligned(t *testing.T) {
	out := textTable([]string{"A", "B"}, [][]string{{"\x1b[32mup\x1b[0m", "x"}, {"crashed", "y"}})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// "x" and "y" must start in the same visible column despite the ANSI codes.
	if strings.Index(stripANSI(lines[1]), "x") != strings.Index(lines[2], "y") {
		t.Errorf("misaligned:\n%s", out)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			in = true
		case in && r == 'm':
			in = false
		case !in:
			b.WriteRune(r)
		}
	}
	return b.String()
}
