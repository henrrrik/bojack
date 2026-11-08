package view

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// maxTableCellWidth caps column widths in textTable; longer cells are cut
// with "…".
const maxTableCellWidth = 40

// textTable renders a plain left-aligned table with a header row, for use
// inside detail tabs. Cells may be pre-styled: widths are measured without
// ANSI codes.
func textTable(header []string, rows [][]string) string {
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = lipgloss.Width(h)
	}
	for _, r := range rows {
		for i, c := range r {
			widths[i] = max(widths[i], min(lipgloss.Width(c), maxTableCellWidth))
		}
	}
	var b strings.Builder
	writeRow := func(cells []string) {
		for i, c := range cells {
			if i > 0 {
				b.WriteString("  ")
			}
			c = truncate(c, widths[i])
			b.WriteString(c + strings.Repeat(" ", widths[i]-lipgloss.Width(c)))
		}
		b.WriteString("\n")
	}
	writeRow(header)
	for _, r := range rows {
		writeRow(r)
	}
	return b.String()
}

// truncate shortens s to width visible characters. Only plain strings are
// cut; styled cells are expected to be short.
func truncate(s string, width int) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	r := []rune(s)
	return string(r[:width-1]) + "…"
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
