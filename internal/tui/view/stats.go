package view

import (
	"math"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/henrrrik/bojack/internal/msg"
	"github.com/henrrrik/bojack/internal/runway"
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

// statsView renders a stats tab: one summary line per metric, with the
// selected metric charted below them. A pinned metric is charted above the
// selected one, to compare the two over the same time range.
type statsView struct {
	cursor int    // selected metric, in OrderedKeys order
	pinned string // key of the pinned metric, or ""
}

// Chart size limits, in cells.
const (
	minChartHeight = 6
	maxChartHeight = 18
	maxChartWidth  = 120
	overviewGauge  = 24
)

const (
	// pinKey pins the selected metric, or unpins it if it is pinned.
	pinKey = "p"
	// pinMarker marks the pinned metric in the overview.
	pinMarker = "• "
)

// key moves the selection among the metrics with the given keys, or pins
// one; ok reports whether k was a stats key.
func (s statsView) key(k string, keys []string) (statsView, bool) {
	n := len(keys)
	switch k {
	case "up", "k":
		s.cursor = max(0, min(s.cursor, n-1)-1)
	case "down", "j":
		s.cursor = max(0, min(n-1, s.cursor+1))
	case pinKey:
		s = s.togglePin(keys)
	default:
		return s, false
	}
	return s, true
}

func (s statsView) togglePin(keys []string) statsView {
	if len(keys) == 0 {
		return s
	}
	sel := keys[min(s.cursor, len(keys)-1)]
	if s.pinned == sel {
		s.pinned = ""
	} else {
		s.pinned = sel
	}
	return s
}

// help is the stats tab's key help.
func (s statsView) help() string {
	if s.pinned != "" {
		return "p: pin/unpin"
	}
	return "p: pin to compare"
}

// render draws the tab in width by height cells.
func (s statsView) render(styles DetailStyles, stats runway.Stats, width, height int, interval string) string {
	var b strings.Builder
	b.WriteString(styles.Label.Render("Interval:") + styles.Value.Render(statsIntervalLabel(interval)) + "\n\n")
	if len(stats) == 0 {
		b.WriteString("No stats available.\n")
		return b.String()
	}
	keys := stats.OrderedKeys()
	sel := min(s.cursor, len(keys)-1)
	b.WriteString(renderOverview(styles, stats, keys, sel, s.pinned, width))
	charted, chartH := s.chartLayout(stats, keys[sel], height-2-len(keys)-1)
	if chartH < minChartHeight {
		return b.String()
	}
	metrics := make([]runway.StatsMetric, len(charted))
	for i, k := range charted {
		metrics[i] = stats[k]
	}
	for i, chart := range renderCharts(styles, metrics, min(width-2, maxChartWidth), chartH) {
		b.WriteString("\n" + styles.Value.Bold(true).Render(chartTitle(charted[i], metrics[i])) + "\n" + chart)
	}
	return b.String()
}

// chartLayout picks the metrics to chart in room lines, each chart with a
// title line, and the height of each chart: the pinned and the selected
// metric if both fit, else only the selected one.
func (s statsView) chartLayout(stats runway.Stats, selected string, room int) ([]string, int) {
	if _, ok := stats[s.pinned]; ok && s.pinned != selected {
		if h := min(maxChartHeight, room/2-1); h >= minChartHeight {
			return []string{s.pinned, selected}, h
		}
	}
	return []string{selected}, min(maxChartHeight, room-1)
}

func chartTitle(key string, metric runway.StatsMetric) string {
	title := metricTitle(key, metric)
	if metric.Unit != "" {
		title += " (" + metric.Unit + ")"
	}
	return title
}

// renderOverview renders one line per metric: a selection (or pin) marker,
// the title, the latest value against the limit, and a usage gauge.
func renderOverview(styles DetailStyles, stats runway.Stats, keys []string, sel int, pinned string, width int) string {
	titleW := statsLabelWidth(stats, keys, 0)
	values := make([]string, len(keys))
	valueW := 0
	for i, k := range keys {
		values[i] = renderStatValue(styles, stats[k])
		valueW = max(valueW, lipgloss.Width(values[i]))
	}
	gaugeW := min(overviewGauge, width-2-titleW-valueW-2-5) // 5: the percentage
	var b strings.Builder
	for i, k := range keys {
		metric := stats[k]
		marker, title := "  ", styles.Label.UnsetWidth().Render(padRight(metricTitle(k, metric), titleW))
		switch {
		case i == sel:
			marker = styles.Title.UnsetPaddingLeft().Render("› ")
			title = styles.Value.Bold(true).Render(padRight(metricTitle(k, metric), titleW))
		case k == pinned:
			marker = styles.Title.UnsetPaddingLeft().Render(pinMarker)
		}
		line := marker + title + padRight(values[i], valueW)
		if gauge := renderGauge(styles, metric, gaugeW); gauge != "" {
			line += "  " + gauge
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// padRight pads s, which may contain styling, to w cells.
func padRight(s string, w int) string {
	return s + strings.Repeat(" ", max(0, w-lipgloss.Width(s)))
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
	return levelStyle(styles, statLevelFor(metric, latest)).Render(text)
}

func levelStyle(styles DetailStyles, level statLevel) lipgloss.Style {
	switch level {
	case levelCritical:
		return styles.StatusDown
	case levelWarn:
		return styles.StatusWarn
	default:
		return styles.Value
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

// detailBodyHeight is the height left for a detail view's active tab, given
// the terminal height: the title, tab bar and help take 6 lines, the follow
// banner 2 more, and the action status line its own height.
func detailBodyHeight(height int, following bool, status string) int {
	h := height - 6
	if following {
		h -= 2
	}
	if status != "" {
		h -= lipgloss.Height(status)
	}
	return h
}
