package view

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
			b.WriteString(c + strings.Repeat(" ", max(0, widths[i]-lipgloss.Width(c))))
		}
		b.WriteString("\n")
	}
	writeRow(header)
	for _, r := range rows {
		writeRow(r)
	}
	return b.String()
}

// truncate shortens s to at most width terminal cells, ending in "…". It
// measures cells, not runes, so wide characters (CJK, emoji) and styled
// strings are cut safely.
func truncate(s string, width int) string {
	return ansi.Truncate(s, width, "…")
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
