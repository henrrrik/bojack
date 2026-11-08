package view

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// renderTabBar renders tab names side by side, with the active one highlighted.
func renderTabBar(styles DetailStyles, names []string, active int) string {
	tabs := make([]string, len(names))
	for i, n := range names {
		tabs[i] = tabStyle(styles, i == active).Render(n)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func tabStyle(styles DetailStyles, active bool) lipgloss.Style {
	if active {
		return styles.TabActive
	}
	return styles.TabInactive
}

// clickedTab returns the tab under a left click, for a detail view that
// renders title, a blank line, then the tab bar. ok is false if the click
// is anywhere else.
func clickedTab(c tea.MouseClickMsg, styles DetailStyles, title string, names []string, active int) (int, bool) {
	top := lipgloss.Height(title) + 1
	if c.Button != tea.MouseLeft || c.Y < top || c.Y >= top+lipgloss.Height(renderTabBar(styles, names, active)) {
		return 0, false
	}
	left := 0
	for i, n := range names {
		right := left + lipgloss.Width(tabStyle(styles, i == active).Render(n))
		if c.X >= left && c.X < right {
			return i, true
		}
		left = right
	}
	return 0, false
}
