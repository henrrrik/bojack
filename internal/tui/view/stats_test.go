package view

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/henrrrik/bojack/internal/runway"
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
	s, ok := statsView{}.key("down", twoMetrics().OrderedKeys())
	if !ok || s.cursor != 1 {
		t.Fatalf("down should move to the second metric, got %+v", s)
	}
	out := ansi.Strip(s.render(testDetailStyles(), twoMetrics(), 100, 30, "24h"))
	if !strings.Contains(out, "128┤") || strings.Contains(out, "0.5┤") {
		t.Errorf("the chart should follow the cursor:\n%s", out)
	}
	if s, _ = s.key("down", twoMetrics().OrderedKeys()); s.cursor != 1 {
		t.Error("down on the last metric should stay")
	}
	if s, _ = s.key("k", twoMetrics().OrderedKeys()); s.cursor != 0 {
		t.Error("k should move up")
	}
	if _, ok = s.key("x", twoMetrics().OrderedKeys()); ok {
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

func TestStatsPinComparesTwoCharts(t *testing.T) {
	keys := twoMetrics().OrderedKeys()
	s, ok := statsView{}.key(pinKey, keys)
	if !ok || s.pinned != "cpu" {
		t.Fatalf("p should pin the selected metric, got %+v", s)
	}
	s, _ = s.key("down", keys)
	out := ansi.Strip(s.render(testDetailStyles(), twoMetrics(), 100, 40, "24h"))
	cpu, ram := strings.Index(out, "0.5┤"), strings.Index(out, "128┤")
	if cpu < 0 || ram < 0 || cpu > ram {
		t.Errorf("the pinned chart should be drawn above the selected one:\n%s", out)
	}
	if !strings.HasPrefix(firstLineWith(strings.Split(out, "\n"), "CPU"), pinMarker) {
		t.Errorf("the pinned metric should be marked in the overview:\n%s", out)
	}
	if n := strings.Count(out, "\n") + 1; n > 40 {
		t.Errorf("stats use %d lines, want at most 40", n)
	}
}

func TestStatsPinAgainUnpins(t *testing.T) {
	keys := twoMetrics().OrderedKeys()
	s, _ := statsView{}.key(pinKey, keys)
	s, _ = s.key(pinKey, keys)
	if s.pinned != "" {
		t.Errorf("p on the pinned metric should unpin it, got %q", s.pinned)
	}
	s, _ = s.key(pinKey, keys)
	s, _ = s.key("down", keys)
	s, _ = s.key(pinKey, keys)
	if s.pinned != "ram" {
		t.Errorf("p on another metric should pin that one instead, got %q", s.pinned)
	}
}

func TestStatsPinnedAndSelectedSameDrawsOneChart(t *testing.T) {
	s, _ := statsView{}.key(pinKey, twoMetrics().OrderedKeys())
	out := ansi.Strip(s.render(testDetailStyles(), twoMetrics(), 100, 40, "24h"))
	if n := strings.Count(out, "└"); n != 1 {
		t.Errorf("got %d charts, want 1:\n%s", n, out)
	}
}

func TestStatsPinWithoutRoomForTwoShowsSelected(t *testing.T) {
	keys := twoMetrics().OrderedKeys()
	s, _ := statsView{}.key(pinKey, keys)
	s, _ = s.key("down", keys)
	out := ansi.Strip(s.render(testDetailStyles(), twoMetrics(), 100, 16, "24h"))
	if strings.Count(out, "└") != 1 || !strings.Contains(out, "128┤") {
		t.Errorf("with room for one chart, only the selected one should be drawn:\n%s", out)
	}
}

func TestStatsPinnedMetricGoneIsIgnored(t *testing.T) {
	s := statsView{pinned: "gone", cursor: 1}
	out := ansi.Strip(s.render(testDetailStyles(), twoMetrics(), 100, 40, "24h"))
	if strings.Count(out, "└") != 1 {
		t.Errorf("a pinned metric missing from the stats should not be charted:\n%s", out)
	}
}

func TestStackedChartsShareTheirColumns(t *testing.T) {
	big := flatMetric(700, 1024)
	small := flatMetric(0.1, 0.5)
	small.Data = small.Data[3:] // starts later than big
	out := ansi.Strip(strings.Join(renderCharts(testDetailStyles(), []runway.StatsMetric{small, big}, 80, 10), "\n"))
	var axes []int
	for _, l := range strings.Split(out, "\n") {
		if i := strings.Index(l, "└"); i >= 0 {
			axes = append(axes, i)
		}
	}
	if len(axes) != 2 || axes[0] != axes[1] {
		t.Errorf("both charts should start their plot in the same column, got %v:\n%s", axes, out)
	}
	labels := strings.Split(out, "\n")
	if strings.Count(out, "10:00") != 2 {
		t.Errorf("both charts should span the same time range (from 10:00):\n%s", strings.Join(labels, "\n"))
	}
}

func TestStatsKeysBeforeDataArrives(t *testing.T) {
	s := statsView{}
	for _, k := range []string{"down", "j", "p", "up"} {
		s, _ = s.key(k, nil)
	}
	if s.cursor < 0 {
		t.Fatalf("cursor = %d after keys on empty stats, want >= 0", s.cursor)
	}
	keys := twoMetrics().OrderedKeys()
	_ = s.render(testDetailStyles(), twoMetrics(), 100, 30, "24h")
	if s, _ = s.key("p", keys); s.pinned != keys[0] {
		t.Errorf("pinned = %q, want the first metric %q", s.pinned, keys[0])
	}
}
