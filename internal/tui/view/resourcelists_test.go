package view

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func assertRow(t *testing.T, got, want table.Row) {
	t.Helper()
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("row = %q, want %q", got, want)
	}
}

func TestVolumeListShowsSize(t *testing.T) {
	var v Viewer = NewVolumeListModel(lipgloss.NewStyle(), 120, 24)
	v, _ = v.Update(msg.ListResultMsg[runway.Volume]{
		Items: []runway.Volume{{Name: "data", App: "my-db", Mountpoint: "/data", SizeMB: 5120}},
	})
	if out := v.View(); !strings.Contains(out, "5.0 GB") {
		t.Errorf("view should show size as 5.0 GB, got %q", out)
	}
}

func TestVolumeRow(t *testing.T) {
	assertRow(t, volumeRow(runway.Volume{Name: "data", App: "my-db", Mountpoint: "/data", SizeMB: 512}),
		table.Row{"data", "my-db", "/data", "512 MB", "-", "-"})
	assertRow(t, volumeRow(runway.Volume{Name: "data", SizeMB: 5120, SizeExtraMB: 1024,
		Replicas: []json.RawMessage{json.RawMessage(`"r1"`), json.RawMessage(`"r2"`)}}),
		table.Row{"data", "", "", "5.0 GB", "1.0 GB", "2"})
	assertRow(t, volumeRow(runway.Volume{Name: "data-r1", SizeMB: 512, IsReplica: true}),
		table.Row{"data-r1", "", "", "512 MB", "-", "replica"})
}

func TestDomainRow(t *testing.T) {
	assertRow(t, domainRow(runway.RootDomain{Domain: "example.com", Owner: "me", Verified: true}),
		table.Row{"example.com", "me", "true"})
}

func TestPolicyRow(t *testing.T) {
	assertRow(t, policyRow(runway.Policy{Name: "office", App: "my-app", From: "10.0.0.0/8"}),
		table.Row{"office", "my-app", "10.0.0.0/8"})
}

func TestRegionRow(t *testing.T) {
	assertRow(t, regionRow(runway.Region{ID: "eu-1", Name: "Frankfurt", Country: "Germany", Continent: "Europe", LiveStatus: false}),
		table.Row{"eu-1", "Frankfurt", "Germany", "Europe", "false"})
}

func TestPlanRow(t *testing.T) {
	assertRow(t, planRow(runway.Plan{Plan: "start-m", CPU: 0.5, RAMMB: 512, HDDMB: 2048, BandwidthMbits: 10, PriceHour: 0.0123, Description: "small"}),
		table.Row{"start-m", "0.5", "512", "2048", "10", "0.0123", "small"})
}

func TestFormatMB(t *testing.T) {
	cases := map[string]string{
		"512":  "512 MB",
		"1024": "1.0 GB",
		"5120": "5.0 GB",
		"":     "-",
		"abc":  "abc",
	}
	for in, want := range cases {
		if got := formatMB(in); got != want {
			t.Errorf("formatMB(%q) = %q, want %q", in, got, want)
		}
	}
}
