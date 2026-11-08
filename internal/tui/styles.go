package tui

import (
	"charm.land/lipgloss/v2"
	"github.com/henrrrik/runway-tui/internal/tui/view"
)

// Styles holds shared lipgloss styles.
type Styles struct {
	Title       lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	TabGap      lipgloss.Style
	Error       lipgloss.Style
	HelpBar     lipgloss.Style
	StatusUp    lipgloss.Style
	StatusDown  lipgloss.Style
	StatusWarn  lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	Sparkline   lipgloss.Style
}

// DefaultStyles returns the default style set.
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			PaddingLeft(1),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("205")).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 2),
		TabGap: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("238")),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			PaddingLeft(1),
		HelpBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(1),
		StatusUp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")),
		StatusDown: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
		StatusWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(16),
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")),
		Sparkline: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")),
	}
}

// DetailStyles converts Styles to the view.DetailStyles used by detail views.
func (s Styles) DetailStyles() view.DetailStyles {
	return view.DetailStyles{
		Title:       s.Title,
		TabActive:   s.TabActive,
		TabInactive: s.TabInactive,
		Error:       s.Error,
		HelpBar:     s.HelpBar,
		Label:       s.Label,
		Value:       s.Value,
		StatusUp:    s.StatusUp,
		StatusDown:  s.StatusDown,
		StatusWarn:  s.StatusWarn,
		Sparkline:   s.Sparkline,
	}
}
