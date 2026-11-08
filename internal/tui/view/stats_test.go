package view

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// intervalLineRE matches the stats header showing the default interval.
var intervalLineRE = regexp.MustCompile(`Interval:\s+24h`)

func TestNextStatsInterval(t *testing.T) {
	tests := []struct{ in, want string }{
		{"1h", "6h"}, {"6h", "24h"}, {"24h", "168h"}, {"168h", "720h"}, {"720h", "1h"}, {"bogus", "1h"},
	}
	for _, tt := range tests {
		if got := NextStatsInterval(tt.in); got != tt.want {
			t.Errorf("NextStatsInterval(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStatsIntervalLabel(t *testing.T) {
	tests := []struct{ in, want string }{
		{"1h", "1h"}, {"24h", "24h"}, {"168h", "7d"}, {"720h", "30d"}, {"bogus", "bogus"},
	}
	for _, tt := range tests {
		if got := statsIntervalLabel(tt.in); got != tt.want {
			t.Errorf("statsIntervalLabel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func twoMetrics() runway.Stats {
	cpu := flatMetric(0.1, 0.5)
	cpu.Title, cpu.Unit = "CPU", "cores"
	return runway.Stats{"cpu": cpu, "ram": flatMetric(70, 128)}
}

func TestStatsOverviewAndChart(t *testing.T) {
	out := ansi.Strip(statsView{}.render(testDetailStyles(), twoMetrics(), 100, 30, "24h"))
	lines := strings.Split(out, "\n")
	cpuLine, ramLine := firstLineWith(lines, "CPU"), firstLineWith(lines, "Memory")
	if !strings.HasPrefix(cpuLine, "›") || strings.HasPrefix(ramLine, "›") {
		t.Errorf("the first metric should be selected:\n%s", out)
	}
	if !strings.Contains(cpuLine, "0.1 / 0.5 cores") || !strings.Contains(ramLine, "70 / 128 MiB") {
		t.Errorf("each metric should have a one-line summary:\n%s", out)
	}
	if !strings.Contains(out, "└") || !strings.Contains(out, "0.5┤") {
		t.Errorf("the selected metric should be charted:\n%s", out)
	}
	if n := len(lines); n > 30 {
		t.Errorf("stats use %d lines, want at most 30", n)
	}
}

func TestStatsCursorSelectsChart(t *testing.T) {
	s, ok := statsView{}.key("down", 2)
	if !ok || s.cursor != 1 {
		t.Fatalf("down should move to the second metric, got %+v", s)
	}
	out := ansi.Strip(s.render(testDetailStyles(), twoMetrics(), 100, 30, "24h"))
	if !strings.Contains(out, "128┤") || strings.Contains(out, "0.5┤") {
		t.Errorf("the chart should follow the cursor:\n%s", out)
	}
	if s, _ = s.key("down", 2); s.cursor != 1 {
		t.Error("down on the last metric should stay")
	}
	if s, _ = s.key("k", 2); s.cursor != 0 {
		t.Error("k should move up")
	}
	if _, ok = s.key("x", 2); ok {
		t.Error("other keys are not stats keys")
	}
}

func TestStatsWithoutRoomSkipsChart(t *testing.T) {
	out := ansi.Strip(statsView{}.render(testDetailStyles(), twoMetrics(), 100, 8, "24h"))
	if strings.Contains(out, "└") || !strings.Contains(out, "Memory") {
		t.Errorf("a short terminal should show only the summaries:\n%s", out)
	}
}

func TestStatsCursorClampedWhenMetricsShrink(t *testing.T) {
	out := ansi.Strip(statsView{cursor: 5}.render(testDetailStyles(), twoMetrics(), 100, 30, "24h"))
	if !strings.Contains(out, "128┤") {
		t.Errorf("an out-of-range cursor should select the last metric:\n%s", out)
	}
}

func firstLineWith(lines []string, sub string) string {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return l
		}
	}
	return ""
}
