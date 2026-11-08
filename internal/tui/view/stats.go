package view

import (
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"strconv"
	"strings"

	"github.com/henrrrik/runway-tui/internal/runway"
)

// statsInterval pairs a display label with the duration passed to the CLI's
// --interval flag (which takes Go durations, so "7d" must be sent as "168h").
type statsInterval struct {
	label string
	arg   string
}

var statsIntervals = []statsInterval{
	{"1h", "1h"},
	{"6h", "6h"},
	{"24h", "24h"},
	{"7d", "168h"},
	{"30d", "720h"},
}

// DefaultStatsInterval matches the CLI's default --interval.
const DefaultStatsInterval = "24h"

// NextStatsInterval returns the interval after arg, wrapping around.
// Unknown values restart at the shortest interval.
func NextStatsInterval(arg string) string {
	for i, iv := range statsIntervals {
		if iv.arg == arg {
			return statsIntervals[(i+1)%len(statsIntervals)].arg
		}
	}
	return statsIntervals[0].arg
}

// statsIntervalKey is the key that cycles the stats interval.
const statsIntervalKey = "i"

// requestStatsInterval returns a command asking the root model to refetch
// stats for interval.
func requestStatsInterval(interval string) tea.Cmd {
	return func() tea.Msg { return msg.StatsIntervalMsg{Interval: interval} }
}

func statsIntervalLabel(arg string) string {
	for _, iv := range statsIntervals {
		if iv.arg == arg {
			return iv.label
		}
	}
	return arg
}

// renderStats renders a stats tab: the interval header followed by one row
// per metric with its latest value and a sparkline.
func renderStats(styles DetailStyles, stats runway.Stats, width int, interval string) string {
	var b strings.Builder
	b.WriteString(styles.Label.Render("Interval:") + styles.Value.Render(statsIntervalLabel(interval)) + "\n\n")
	if len(stats) == 0 {
		b.WriteString("No stats available.\n")
		return b.String()
	}
	graphWidth := width - 4
	if graphWidth > 60 {
		graphWidth = 60
	}
	keys := stats.OrderedKeys()
	label := styles.Label.Width(statsLabelWidth(stats, keys, styles.Label.GetWidth()))
	for _, key := range keys {
		metric := stats[key]
		b.WriteString(label.Render(metricTitle(key, metric)+":") + renderStatValue(styles, metric) + "\n")
		if graph := sparkline(metric.Values(), graphWidth); graph != "" {
			b.WriteString("  " + styles.Sparkline.Render(graph) + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func metricTitle(key string, metric runway.StatsMetric) string {
	if metric.Title == "" {
		return key
	}
	return metric.Title
}

// statsLabelWidth widens the label column to fit the longest metric title.
func statsLabelWidth(stats runway.Stats, keys []string, min int) int {
	w := min
	for _, k := range keys {
		if n := len(metricTitle(k, stats[k])) + 2; n > w {
			w = n
		}
	}
	return w
}

func renderStatValue(styles DetailStyles, metric runway.StatsMetric) string {
	latest, ok := metric.Latest()
	if !ok {
		return styles.Value.Render("no data")
	}
	text := formatStatValue(latest)
	if metric.Max > 0 {
		text += " / " + formatStatValue(metric.Max)
	}
	if metric.Unit != "" {
		text += " " + metric.Unit
	}
	switch statLevelFor(metric, latest) {
	case levelCritical:
		return styles.StatusDown.Render(text)
	case levelWarn:
		return styles.StatusWarn.Render(text)
	default:
		return styles.Value.Render(text)
	}
}

type statLevel int

const (
	levelOK statLevel = iota
	levelWarn
	levelCritical
)

// statLevelFor classifies v against the metric's thresholds. A zero
// threshold means it is not set.
func statLevelFor(metric runway.StatsMetric, v float64) statLevel {
	switch {
	case metric.CriticalAbove > 0 && v > metric.CriticalAbove:
		return levelCritical
	case metric.WarnAbove > 0 && v > metric.WarnAbove:
		return levelWarn
	default:
		return levelOK
	}
}

// formatStatValue formats a metric value compactly: whole numbers at 100+,
// one decimal at 1+, and three significant digits below that.
func formatStatValue(v float64) string {
	switch {
	case v == 0:
		return "0"
	case math.Abs(v) >= 100:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case math.Abs(v) >= 1:
		return strconv.FormatFloat(math.Round(v*10)/10, 'f', -1, 64)
	default:
		return strconv.FormatFloat(v, 'g', 3, 64)
	}
}

var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func sparkline(values []float64, width int) string {
	if len(values) == 0 {
		return ""
	}
	values = downsample(values, width)
	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	var sb strings.Builder
	for _, v := range values {
		idx := 3
		if max > min {
			idx = int((v - min) / (max - min) * 7)
			if idx > 7 {
				idx = 7
			}
		}
		sb.WriteRune(sparkBlocks[idx])
	}
	return sb.String()
}

// downsample reduces values to at most width points by splitting the series
// into width buckets and keeping each bucket's maximum, so the whole interval
// is shown and short spikes stay visible.
func downsample(values []float64, width int) []float64 {
	if width <= 0 || len(values) <= width {
		return values
	}
	out := make([]float64, width)
	for i := range out {
		lo, hi := i*len(values)/width, (i+1)*len(values)/width
		maxV := values[lo]
		for _, v := range values[lo+1 : hi] {
			maxV = math.Max(maxV, v)
		}
		out[i] = maxV
	}
	return out
}
