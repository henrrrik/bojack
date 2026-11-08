package view

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func seriesMetric(limit float64, values ...float64) runway.StatsMetric {
	m := runway.StatsMetric{Max: limit}
	for i, v := range values {
		m.Data = append(m.Data, runway.StatsPoint{Timestamp: int64(i) * 60000, Value: v})
	}
	return m
}

func gauge(m runway.StatsMetric, width int) string {
	return ansi.Strip(renderGauge(testDetailStyles(), m, width))
}

func TestGaugeFillsToTheLatestValue(t *testing.T) {
	if got, want := gauge(seriesMetric(128, 64), 10), "━━━━━─────  50%"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := gauge(seriesMetric(100, 15), 10), "━╸────────  15%"; got != want {
		t.Errorf("half cells: got %q, want %q", got, want)
	}
}

func TestGaugeMarksThePeak(t *testing.T) {
	// peak 90% of 10 cells is cell 9; latest 20% fills 2 cells.
	if got, want := gauge(seriesMetric(100, 10, 90, 20), 10), "━━───────│  20%"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	// A peak inside the filled part is not marked.
	if got := gauge(seriesMetric(100, 10, 50), 10); strings.Contains(got, "│") {
		t.Errorf("latest is the peak, no marker expected: %q", got)
	}
}

func TestGaugeShowsTinyValues(t *testing.T) {
	// 0.0279 of 0.5 cores is under half a cell but not zero.
	if got, want := gauge(seriesMetric(0.5, 0.0279), 4), "╸───   6%"; got != want {
		t.Errorf("a small non-zero value should still show: got %q, want %q", got, want)
	}
	if got := gauge(seriesMetric(0.5, 0), 4); got != "────   0%" {
		t.Errorf("zero should be empty: %q", got)
	}
}

func TestGaugeOverTheLimit(t *testing.T) {
	if got, want := gauge(seriesMetric(100, 150), 4), "━━━━ 150%"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGaugeNeedsALimitAndData(t *testing.T) {
	if got := gauge(seriesMetric(0, 5), 10); got != "" {
		t.Errorf("no limit: got %q, want nothing", got)
	}
	if got := gauge(seriesMetric(100), 10); got != "" {
		t.Errorf("no data: got %q, want nothing", got)
	}
}

func TestOverviewShowsGauges(t *testing.T) {
	out := ansi.Strip(statsView{}.render(testDetailStyles(), twoMetrics(), 100, 30, "24h"))
	ram := firstLineWith(strings.Split(out, "\n"), "Memory")
	if !strings.Contains(ram, "━") || !strings.HasSuffix(strings.TrimRight(ram, " "), "55%") {
		t.Errorf("the overview should show a usage gauge per metric: %q", ram)
	}
}

func TestGaugeTinyPercentIsNotZero(t *testing.T) {
	if got, want := gauge(seriesMetric(10, 0.0228), 4), "╸───  <1%"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
