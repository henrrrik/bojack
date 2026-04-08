package view

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func testDetailStyles() DetailStyles {
	return DetailStyles{
		Title:       lipgloss.NewStyle(),
		TabActive:   lipgloss.NewStyle().Bold(true),
		TabInactive: lipgloss.NewStyle(),
		Error:       lipgloss.NewStyle(),
		HelpBar:     lipgloss.NewStyle(),
		Label:       lipgloss.NewStyle().Width(16),
		Value:       lipgloss.NewStyle(),
		StatusUp:    lipgloss.NewStyle(),
		StatusDown:  lipgloss.NewStyle(),
		Sparkline:   lipgloss.NewStyle(),
	}
}

func TestAppDetailShowsAppName(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	v := m.View()
	if !strings.Contains(v, "my-app") {
		t.Errorf("view should contain app name, got %q", v)
	}
}

func TestAppDetailInfoTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppDetailResultMsg{
		Detail: &runway.AppDetail{
			App:      runway.App{ID: "my-app", Plan: "standard-1", Region: "eu-north-1"},
			Settings: runway.AppSettings{Routable: true},
		},
	})
	v := m.View()
	if !strings.Contains(v, "standard-1") {
		t.Error("info tab should show plan")
	}
}

func TestAppDetailTabSwitching(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	if m.activeTab != 0 {
		t.Fatalf("initial tab = %d, want 0", m.activeTab)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.activeTab != 1 {
		t.Errorf("tab = %d, want 1", m.activeTab)
	}
	// Shift+tab goes back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.activeTab != 0 {
		t.Errorf("tab = %d, want 0", m.activeTab)
	}
	// Wraps around backwards
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.activeTab != len(appTabs)-1 {
		t.Errorf("tab = %d, want %d", m.activeTab, len(appTabs)-1)
	}
}

func TestAppDetailProcessesTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppProcessesResultMsg{
		Processes: []runway.Process{
			{Name: "web.1", Type: "web", State: "up", Release: "v42"},
		},
	})
	m.activeTab = 1
	v := m.View()
	if !strings.Contains(v, "web.1") {
		t.Error("processes tab should show process name")
	}
}

func TestAppDetailConfigTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppConfigResultMsg{
		Config: runway.AppConfig{"PORT": "8080"},
	})
	m.activeTab = 3
	v := m.View()
	if !strings.Contains(v, "PORT") {
		t.Error("config tab should show config key")
	}
}

func TestAppDetailLogsTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppLogsResultMsg{
		Logs: []runway.LogEntry{
			{Content: "shutting down server", Time: "2026-04-05T23:27:57Z", Type: "app"},
		},
	})
	m.activeTab = 6
	v := m.View()
	if !strings.Contains(v, "shutting down server") {
		t.Error("logs tab should show log content")
	}
	if !strings.Contains(v, "[app]") {
		t.Error("logs tab should show log type")
	}
}

func TestSparklineAscending(t *testing.T) {
	got := sparkline([]float64{0, 1, 2, 3, 4, 5, 6, 7}, 40)
	want := "▁▂▃▄▅▆▇█"
	if got != want {
		t.Errorf("sparkline = %q, want %q", got, want)
	}
}

func TestSparklineEmpty(t *testing.T) {
	got := sparkline(nil, 40)
	if got != "" {
		t.Errorf("sparkline = %q, want empty", got)
	}
}

func TestSparklineAllSame(t *testing.T) {
	got := sparkline([]float64{5, 5, 5, 5}, 40)
	// All same values should produce middle blocks
	for _, r := range got {
		if r != '▄' {
			t.Errorf("sparkline = %q, expected all ▄", got)
			break
		}
	}
}

func TestSparklineTruncatesWidth(t *testing.T) {
	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i)
	}
	got := sparkline(values, 20)
	// Should use last 20 values, so length = 20 runes
	runes := []rune(got)
	if len(runes) != 20 {
		t.Errorf("len = %d, want 20", len(runes))
	}
}

func TestStatsTabShowsSparklines(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppStatsResultMsg{
		Stats: &runway.AppStats{
			All: []runway.StatsEntry{
				{Values: runway.StatsValues{CPU: runway.StatsCPU{Load: 0.1}}},
				{Values: runway.StatsValues{CPU: runway.StatsCPU{Load: 0.5}}},
				{Values: runway.StatsValues{CPU: runway.StatsCPU{Load: 0.9}}},
			},
			Latest: runway.StatsValues{
				CPU: runway.StatsCPU{Load: 0.9},
				RAM: runway.StatsRAM{UsedBytes: 1024},
				HDD: runway.StatsHDD{UsedBytes: 2048},
				Net: runway.StatsNet{ReceivedBytes: 512, TransmittedBytes: 256},
			},
		},
	})
	m.activeTab = 2
	v := m.View()
	// Should contain sparkline block characters
	if !strings.ContainsAny(v, "▁▂▃▄▅▆▇█") {
		t.Error("stats tab should contain sparkline characters")
	}
	// Should still show the latest value
	if !strings.Contains(v, "0.90") {
		t.Error("stats tab should show latest CPU load")
	}
}

func TestAppDetailErrorTab(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppDetailResultMsg{Err: errTest})
	v := m.View()
	if !strings.Contains(v, "Error") {
		t.Error("should show error on info tab")
	}
}
