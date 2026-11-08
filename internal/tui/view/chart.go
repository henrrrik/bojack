package view

import (
	"math"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// Data set names; ntcharts draws them in sorted order, so the series is
// drawn last, on top of the threshold lines.
const (
	chartWarnSet     = "1-warn"
	chartCriticalSet = "2-critical"
	chartSeriesSet   = "3-series"
)

// renderChart draws metric as a braille line chart, width by height cells
// including its axes. The Y axis runs from 0 to the plan limit (or the peak,
// if higher), so the line shows usage against the limit; warning and
// critical thresholds are drawn as horizontal lines. Returns "" if there is
// no data or too little room.
//
// ntcharts draws the plot; the axes are drawn here so that each Y label is
// on the row that contains its value (ntcharts labels the row whose top edge
// is the value, which makes lines look half a row off their labels).
func renderChart(styles DetailStyles, metric runway.StatsMetric, width, height int) string {
	if len(metric.Data) == 0 || height < 5 {
		return ""
	}
	pts := sortedPoints(metric.Data)
	start, end := pointTime(pts[0]), pointTime(pts[len(pts)-1])
	if !end.After(start) {
		end = start.Add(time.Minute)
	}
	yMax := chartYMax(metric, pts)
	rows := height - 2 // the time axis and its labels take two rows
	ticks := yTicks(yMax, rows)
	labelW := 0
	for _, t := range ticks {
		labelW = max(labelW, len(formatStatValue(t)))
	}
	plotW := width - labelW - 1
	if plotW < 10 {
		return ""
	}
	plot := plotSeries(styles, metric, pts, start, end, yMax, plotW, rows)
	axis := axisStyle(styles)

	labels := make([]string, rows)
	for _, t := range ticks {
		labels[tickRow(t, yMax, rows)] = formatStatValue(t)
	}
	var b strings.Builder
	for r, line := range strings.Split(plot, "\n")[:rows] {
		edge := "│"
		if labels[r] != "" {
			edge = "┤"
		}
		b.WriteString(axis.Render(padLeft(labels[r], labelW)+edge) + line + "\n")
	}
	b.WriteString(axis.Render(strings.Repeat(" ", labelW)+"└"+strings.Repeat("─", plotW)) + "\n")
	b.WriteString(axis.Render(strings.Repeat(" ", labelW+1) + timeAxisLabels(start, end, plotW)))
	return b.String()
}

// plotSeries draws the series and thresholds with ntcharts, without axes.
func plotSeries(styles DetailStyles, metric runway.StatsMetric, pts []runway.StatsPoint,
	start, end time.Time, yMax float64, w, h int) string {
	c := timeserieslinechart.New(w, h,
		timeserieslinechart.WithTimeRange(start, end),
		timeserieslinechart.WithYRange(0, yMax),
		timeserieslinechart.WithXYSteps(0, 0),
	)
	threshold := func(set string, v float64, style lipgloss.Style) {
		if v <= 0 || v > yMax {
			return
		}
		c.SetDataSetStyle(set, style)
		c.PushDataSet(set, timeserieslinechart.TimePoint{Time: start, Value: v})
		c.PushDataSet(set, timeserieslinechart.TimePoint{Time: end, Value: v})
	}
	threshold(chartWarnSet, metric.WarnAbove, styles.StatusWarn.Faint(true))
	threshold(chartCriticalSet, metric.CriticalAbove, styles.StatusDown.Faint(true))
	c.SetDataSetStyle(chartSeriesSet, styles.Sparkline)
	for _, p := range pts {
		c.PushDataSet(chartSeriesSet, timeserieslinechart.TimePoint{Time: pointTime(p), Value: p.Value})
	}
	c.DrawBrailleAll()
	return c.View()
}

func sortedPoints(data []runway.StatsPoint) []runway.StatsPoint {
	pts := append([]runway.StatsPoint(nil), data...)
	sort.Slice(pts, func(i, j int) bool { return pts[i].Timestamp < pts[j].Timestamp })
	return pts
}

func pointTime(p runway.StatsPoint) time.Time {
	return time.UnixMilli(p.Timestamp)
}

// chartYMax is the top of the Y axis: the plan limit, or the peak if the
// metric has no limit or exceeds it.
func chartYMax(metric runway.StatsMetric, pts []runway.StatsPoint) float64 {
	top := metric.Max
	for _, p := range pts {
		top = math.Max(top, p.Value)
	}
	if top <= 0 {
		return 1
	}
	return top
}

// yTicks returns evenly spaced Y labels from 0 to yMax: quarters when there
// are enough rows to keep them on separate rows, else 0, half and the top.
func yTicks(yMax float64, rows int) []float64 {
	n := 2
	if rows >= 8 {
		n = 4
	}
	ticks := make([]float64, n+1)
	for i := range ticks {
		ticks[i] = yMax * float64(i) / float64(n)
	}
	return ticks
}

// tickRow returns the plot row (0 at the top) whose value band contains v,
// with yMax in the top row.
func tickRow(v, yMax float64, rows int) int {
	band := int(math.Floor(v / yMax * float64(rows)))
	return min(rows-1, max(0, rows-1-band))
}

// timeAxisLabels spreads time labels over width columns: the start, the end,
// and as many in between as fit with some space.
func timeAxisLabels(start, end time.Time, width int) string {
	layout := timeLayout(end.Sub(start))
	labelW := len(start.Format(layout))
	n := max(1, min(4, width/(labelW+8))) // gaps between labels
	line := []rune(strings.Repeat(" ", width))
	for i := 0; i <= n; i++ {
		t := start.Add(time.Duration(float64(end.Sub(start)) * float64(i) / float64(n)))
		pos := (width - labelW) * i / n
		copy(line[pos:], []rune(t.Local().Format(layout)))
	}
	return string(line)
}

// timeLayout picks clock times for spans up to a day and a half, dates beyond.
func timeLayout(span time.Duration) string {
	if span > 36*time.Hour {
		return "Jan 2"
	}
	return "15:04"
}

func padLeft(s string, w int) string {
	return strings.Repeat(" ", max(0, w-len(s))) + s
}

// axisStyle is the label colour without the label's fixed width.
func axisStyle(styles DetailStyles) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.Label.GetForeground())
}
