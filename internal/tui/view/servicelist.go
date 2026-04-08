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

// ServiceListModel displays a table of services.
type ServiceListModel struct {
	table    table.Model
	spinner  spinner.Model
	helpBar  lipgloss.Style
	services []runway.Service
	loading  bool
	err      error
	width    int
	height   int
}

// NewServiceListModel creates a new service list model.
func NewServiceListModel(helpBar lipgloss.Style, width, height int) ServiceListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return ServiceListModel{
		spinner: s,
		helpBar: helpBar,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m ServiceListModel) Update(teaMsg tea.Msg) (ServiceListModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.ServiceListResultMsg:
		m.loading = false
		if teaMsg.Err != nil {
			m.err = teaMsg.Err
			return m, nil
		}
		m.services = teaMsg.Services
		m.table = buildServiceTable(teaMsg.Services, m.width, m.height)
		return m, nil
	case tea.KeyMsg:
		if teaMsg.String() == "enter" && len(m.services) > 0 {
			row := m.table.SelectedRow()
			if row != nil {
				return m, func() tea.Msg {
					return msg.NavigateMsg{View: msg.ViewServiceDetail, Param: row[0]}
				}
			}
		}
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

func (m ServiceListModel) View() string {
	if m.loading {
		return m.spinner.View() + " Fetching services..."
	}
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) +
			"\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	}
	if len(m.services) == 0 {
		return "No services found.\n\n" + renderHelp(m.helpBar, "esc: back")
	}
	return m.table.View() + "\n" + renderHelp(m.helpBar, "enter: details", "r: refresh", "esc: back")
}

func buildServiceTable(services []runway.Service, width, height int) table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Type", Width: 12},
		{Title: "Plan", Width: 14},
		{Title: "Backup", Width: 8},
	}

	rows := make([]table.Row, len(services))
	for i, s := range services {
		rows[i] = table.Row{
			s.Name,
			s.Type,
			s.Plan,
			fmt.Sprintf("%v", s.HasBackup),
		}
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
