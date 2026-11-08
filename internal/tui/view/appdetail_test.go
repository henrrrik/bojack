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
		StatusWarn:  lipgloss.NewStyle(),
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

func statsModel(stats runway.AppStats) AppDetailModel {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	m, _ = m.Update(msg.AppStatsResultMsg{Stats: stats})
	m.activeTab = 2
	return m
}

func TestStatsTabShowsSparklines(t *testing.T) {
	m := statsModel(runway.AppStats{
		"cpu": {Title: "CPU", Unit: "cores", Max: 0.5, Data: []runway.StatsPoint{{Value: 0.1}, {Value: 0.2}, {Value: 0.3}}},
	})
	v := m.View()
	if !strings.ContainsAny(v, "▁▂▃▄▅▆▇█") {
		t.Error("stats tab should contain sparkline characters")
	}
}

func TestStatsTabShowsLatestMaxAndUnit(t *testing.T) {
	m := statsModel(runway.AppStats{
		"ram": {Title: "Memory", Unit: "MiB", Max: 128, Data: []runway.StatsPoint{{Value: 10}, {Value: 12.87109375}}},
	})
	v := m.View()
	if !strings.Contains(v, "Memory") {
		t.Errorf("stats tab should show metric title, got:\n%s", v)
	}
	if !strings.Contains(v, "12.9 / 128 MiB") {
		t.Errorf("stats tab should show latest / max unit, got:\n%s", v)
	}
}

func TestStatsTabEmptySeries(t *testing.T) {
	m := statsModel(runway.AppStats{
		"hdd_ephemeral_app": {Title: "HDD (ephemeral, app)", Unit: "MiB", Max: 2048},
	})
	v := m.View()
	if !strings.Contains(v, "HDD (ephemeral, app)") || !strings.Contains(v, "no data") {
		t.Errorf("empty metric should show title and 'no data', got:\n%s", v)
	}
}

func TestStatsTabOrdersMetrics(t *testing.T) {
	m := statsModel(runway.AppStats{
		"net_out": {Title: "Network (out)", Data: []runway.StatsPoint{{Value: 1}}},
		"cpu":     {Title: "CPU", Data: []runway.StatsPoint{{Value: 1}}},
		"ram":     {Title: "Memory", Data: []runway.StatsPoint{{Value: 1}}},
	})
	v := m.View()
	cpu, ram, net := strings.Index(v, "CPU"), strings.Index(v, "Memory"), strings.Index(v, "Network (out)")
	if cpu < 0 || ram < 0 || net < 0 || !(cpu < ram && ram < net) {
		t.Errorf("metrics out of order (cpu=%d ram=%d net=%d):\n%s", cpu, ram, net, v)
	}
}

func TestStatsTabNoStats(t *testing.T) {
	m := statsModel(nil)
	if !strings.Contains(m.View(), "No stats available.") {
		t.Error("nil stats should show 'No stats available.'")
	}
}

func TestStatLevel(t *testing.T) {
	metric := runway.StatsMetric{WarnAbove: 3.25, CriticalAbove: 4}
	tests := []struct {
		v    float64
		want statLevel
	}{
		{1, levelOK}, {3.25, levelOK}, {3.5, levelWarn}, {4, levelWarn}, {4.1, levelCritical},
	}
	for _, tt := range tests {
		if got := statLevelFor(metric, tt.v); got != tt.want {
			t.Errorf("statLevelFor(%v) = %v, want %v", tt.v, got, tt.want)
		}
	}
	if got := statLevelFor(runway.StatsMetric{}, 100); got != levelOK {
		t.Errorf("no thresholds should be levelOK, got %v", got)
	}
}

func TestFormatStatValue(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0"},
		{0.000457, "0.000457"},
		{0.0045, "0.0045"},
		{0.5, "0.5"},
		{12.87109375, "12.9"},
		{128, "128"},
		{1331.2, "1331"},
	}
	for _, tt := range tests {
		if got := formatStatValue(tt.v); got != tt.want {
			t.Errorf("formatStatValue(%v) = %q, want %q", tt.v, got, tt.want)
		}
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
