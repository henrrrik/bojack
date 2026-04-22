package view

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// DomainListModel displays a table of root domains.
type DomainListModel struct {
	table   table.Model
	spinner spinner.Model
	helpBar lipgloss.Style
	domains []runway.RootDomain
	loading bool
	err     error
	width   int
	height  int
}

// NewDomainListModel creates a new domain list model.
func NewDomainListModel(helpBar lipgloss.Style, width, height int) DomainListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return DomainListModel{
		spinner: s,
		helpBar: helpBar,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m DomainListModel) Update(teaMsg tea.Msg) (DomainListModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.DomainListResultMsg:
		m.loading = false
		if teaMsg.Err != nil {
			m.err = teaMsg.Err
			return m, nil
		}
		m.domains = teaMsg.Domains
		m.table = buildDomainTable(teaMsg.Domains, m.height)
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

func (m DomainListModel) View() string {
	if m.loading {
		return m.spinner.View() + " Fetching domains..."
	}
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) +
			"\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	}
	if len(m.domains) == 0 {
		return "No domains found.\n\n" + renderHelp(m.helpBar, "esc: back")
	}
	return m.table.View() + "\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
}

func buildDomainTable(domains []runway.RootDomain, height int) table.Model {
	columns := []table.Column{
		{Title: "Domain", Width: 30},
		{Title: "Owner", Width: 20},
		{Title: "Verified", Width: 10},
	}

	rows := make([]table.Row, len(domains))
	for i, d := range domains {
		rows[i] = table.Row{d.Domain, d.Owner, fmt.Sprintf("%v", d.Verified)}
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
