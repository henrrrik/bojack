package view

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func TestTickRow(t *testing.T) {
	// 10 rows of 12.8 each for 0..128; row 0 is the top.
	for _, c := range []struct {
		v    float64
		want int
	}{{0, 9}, {12.7, 9}, {12.8, 8}, {64, 4}, {70, 4}, {127, 0}, {128, 0}} {
		if got := tickRow(c.v, 128, 10); got != c.want {
			t.Errorf("tickRow(%v) = %d, want %d", c.v, got, c.want)
		}
	}
}

// flatMetric is a metric with a constant value over an hour.
func flatMetric(v, limit float64) runway.StatsMetric {
	start := time.Date(2026, 11, 8, 10, 0, 0, 0, time.Local).UnixMilli()
	var data []runway.StatsPoint
	for i := range 13 {
		data = append(data, runway.StatsPoint{Timestamp: start + int64(i)*5*60*1000, Value: v})
	}
	return runway.StatsMetric{Title: "Memory", Unit: "MiB", Max: limit, Data: data}
}

func chartLines(t *testing.T, m runway.StatsMetric, w, h int) []string {
	t.Helper()
	out := renderChart(testDetailStyles(), m, w, h)
	lines := strings.Split(ansi.Strip(out), "\n")
	if len(lines) != h {
		t.Fatalf("chart has %d lines, want %d:\n%s", len(lines), h, ansi.Strip(out))
	}
	for i, l := range lines {
		if n := ansi.StringWidth(l); n > w {
			t.Errorf("line %d is %d wide, want at most %d: %q", i, n, w, l)
		}
	}
	return lines
}

func rowWithBraille(lines []string) int {
	for i, l := range lines {
		for _, r := range l {
			if r > 0x2800 && r <= 0x28FF {
				return i
			}
		}
	}
	return -1
}

func TestChartValueSitsOnItsTickRow(t *testing.T) {
	// 0..128 MiB in 10 plot rows: ticks 0, 32, 64, 96, 128; 70 is in the
	// row that also holds 64 (64–76.8).
	lines := chartLines(t, flatMetric(70, 128), 60, 12)
	row := rowWithBraille(lines)
	if row < 0 || !strings.HasPrefix(strings.TrimSpace(lines[row]), "64") {
		t.Errorf("the line for 70 should be on the row labelled 64, got row %d:\n%s", row, strings.Join(lines, "\n"))
	}
	if !strings.HasPrefix(strings.TrimSpace(lines[0]), "128") {
		t.Errorf("top row should be labelled with the limit 128:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.HasPrefix(strings.TrimSpace(lines[9]), "0") {
		t.Errorf("bottom plot row should be labelled 0:\n%s", strings.Join(lines, "\n"))
	}
}

func TestChartTimeLabels(t *testing.T) {
	lines := chartLines(t, flatMetric(10, 128), 60, 12)
	last := lines[len(lines)-1]
	if !strings.Contains(last, "10:00") || !strings.Contains(last, "11:00") {
		t.Errorf("time axis should show the start and end, got %q", last)
	}
	if !strings.Contains(lines[len(lines)-2], "└") {
		t.Errorf("second-last line should be the time axis, got %q", lines[len(lines)-2])
	}
}

func TestChartFractionalLimit(t *testing.T) {
	lines := chartLines(t, runway.StatsMetric{Max: 0.5, Data: flatMetric(0.1, 0).Data}, 60, 8)
	if !strings.HasPrefix(strings.TrimSpace(lines[0]), "0.5") {
		t.Errorf("a 0.5 core limit should be labelled 0.5, got:\n%s", strings.Join(lines, "\n"))
	}
}

func TestChartTooSmallOrEmpty(t *testing.T) {
	if renderChart(testDetailStyles(), runway.StatsMetric{}, 60, 12) != "" {
		t.Error("no data should render nothing")
	}
	if renderChart(testDetailStyles(), flatMetric(1, 2), 10, 3) != "" {
		t.Error("a tiny area should render nothing")
	}
}

func TestTimeLabelLayout(t *testing.T) {
	if got := timeLayout(6 * time.Hour); got != "15:04" {
		t.Errorf("6h layout = %q", got)
	}
	if got := timeLayout(7 * 24 * time.Hour); got != "Jan 2" {
		t.Errorf("7d layout = %q", got)
	}
}
