package view

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func TestVolumeListShowsSize(t *testing.T) {
	m := NewVolumeListModel(lipgloss.NewStyle(), 120, 24)
	vols := []runway.Volume{{Name: "data", App: "my-db", Mountpoint: "/data", SizeMB: 5120}}
	m, _ = m.Update(msg.VolumeListResultMsg{Volumes: vols})
	v := m.View()
	if !strings.Contains(v, "5.0 GB") {
		t.Errorf("view should show size as 5.0 GB, got %q", v)
	}
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
