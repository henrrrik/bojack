package view

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// PolicyListModel displays a table of policies.
type PolicyListModel struct {
	table    table.Model
	spinner  spinner.Model
	helpBar  lipgloss.Style
	policies []runway.Policy
	loading  bool
	err      error
	width    int
	height   int
}

// NewPolicyListModel creates a new policy list model.
func NewPolicyListModel(helpBar lipgloss.Style, width, height int) PolicyListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return PolicyListModel{
		spinner: s,
		helpBar: helpBar,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m PolicyListModel) Update(teaMsg tea.Msg) (PolicyListModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.PolicyListResultMsg:
		m.loading = false
		if teaMsg.Err != nil {
			m.err = teaMsg.Err
			return m, nil
		}
		m.policies = teaMsg.Policies
		m.table = buildPolicyTable(teaMsg.Policies, m.height)
		return m, nil
	}
	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(teaMsg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(teaMsg)
	return m, cmd
}

func (m PolicyListModel) View() string {
	if m.loading {
		return m.spinner.View() + " Fetching policies..."
	}
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) +
			"\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	}
	if len(m.policies) == 0 {
		return "No policies found.\n\n" + renderHelp(m.helpBar, "esc: back")
	}
	return m.table.View() + "\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
}

func buildPolicyTable(policies []runway.Policy, height int) table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 24},
		{Title: "App", Width: 20},
		{Title: "From", Width: 24},
	}

	rows := make([]table.Row, len(policies))
	for i, p := range policies {
		rows[i] = table.Row{p.Name, p.App, p.From}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight(height, len(rows))),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Bold(true)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	t.SetStyles(s)
	return t
}
