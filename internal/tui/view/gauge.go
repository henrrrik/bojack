package view

import (
	"fmt"
	"math"
	"strings"

	"github.com/henrrrik/runway-tui/internal/runway"
)

// Gauge glyphs: a heavy line fills up to the latest value (in half cells),
// a light line is the headroom, and a tick marks the interval's peak.
const (
	gaugeFull  = "━"
	gaugeHalf  = "╸"
	gaugeEmpty = "─"
	gaugePeak  = "│"
)

// renderGauge draws the metric's latest value against its limit in width
// cells, coloured by its thresholds, with the peak over the interval marked
// and the percentage of the limit after it. Metrics without a limit or data
// get "".
func renderGauge(styles DetailStyles, metric runway.StatsMetric, width int) string {
	latest, ok := metric.Latest()
	if !ok || metric.Max <= 0 || width < 1 {
		return ""
	}
	halves, peakCell := gaugeCells(latest, peakValue(metric.Values()), metric.Max, width)
	fill := strings.Repeat(gaugeFull, halves/2)
	if halves%2 == 1 {
		fill += gaugeHalf
	}
	var rest strings.Builder
	for c := (halves + 1) / 2; c < width; c++ {
		if c == peakCell {
			rest.WriteString(gaugePeak)
		} else {
			rest.WriteString(gaugeEmpty)
		}
	}
	pct := " " + padLeft(percentLabel(latest/metric.Max*100), 4)
	return levelStyle(styles, statLevelFor(metric, latest)).Render(fill) +
		axisStyle(styles).Faint(true).Render(rest.String()) + styles.Value.Render(pct)
}

// gaugeCells returns how many half cells of width the latest value fills
// (at least one if it is above zero), and the cell holding the peak, or -1
// if the peak is not above the latest value.
func gaugeCells(latest, peak, limit float64, width int) (halves, peakCell int) {
	frac := func(v float64) float64 { return math.Min(1, math.Max(0, v/limit)) }
	halves = int(math.Round(frac(latest) * float64(2*width)))
	if halves == 0 && latest > 0 {
		halves = 1
	}
	peakCell = -1
	if peak > latest {
		peakCell = min(width-1, int(frac(peak)*float64(width)))
	}
	return halves, peakCell
}

func peakValue(values []float64) float64 {
	peak := math.Inf(-1)
	for _, v := range values {
		peak = math.Max(peak, v)
	}
	return peak
}

// percentLabel rounds p to a whole percent, showing "<1%" rather than "0%"
// for small values above zero.
func percentLabel(p float64) string {
	if p > 0 && p < 0.5 {
		return "<1%"
	}
	return fmt.Sprintf("%.0f%%", p)
}
