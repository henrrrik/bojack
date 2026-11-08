package view

import "github.com/charmbracelet/lipgloss"

// DetailStyles holds styles needed by detail views.
type DetailStyles struct {
	Title       lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	Error       lipgloss.Style
	HelpBar     lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	StatusUp    lipgloss.Style
	StatusDown  lipgloss.Style
	StatusWarn  lipgloss.Style
	Sparkline   lipgloss.Style
}
