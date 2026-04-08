package view

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// AppListModel displays a table of apps.
type AppListModel struct {
	table   table.Model
	spinner spinner.Model
	helpBar lipgloss.Style
	apps    []runway.AppListEntry
	loading bool
	err     error
	width   int
	height  int
}

// NewAppListModel creates a new app list model.
func NewAppListModel(helpBar lipgloss.Style, width, height int) AppListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return AppListModel{
		spinner: s,
		helpBar: helpBar,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m AppListModel) Update(teaMsg tea.Msg) (AppListModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.AppListResultMsg:
		m.loading = false
		if teaMsg.Err != nil {
			m.err = teaMsg.Err
			return m, nil
		}
		m.apps = teaMsg.Apps
		m.table = buildAppTable(teaMsg.Apps, m.width, m.height)
		return m, nil
	case tea.KeyMsg:
		if teaMsg.String() == "enter" && len(m.apps) > 0 {
			row := m.table.SelectedRow()
			if row != nil {
				return m, func() tea.Msg {
					return msg.NavigateMsg{View: msg.ViewAppDetail, Param: row[0]}
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

func (m AppListModel) View() string {
	if m.loading {
		return m.spinner.View() + " Fetching apps..."
	}
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) +
			"\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	}
	if len(m.apps) == 0 {
		return "No apps found.\n\n" + renderHelp(m.helpBar, "esc: back")
	}
	return m.table.View() + "\n" + renderHelp(m.helpBar, "enter: details", "r: refresh", "esc: back")
}

func buildAppTable(apps []runway.AppListEntry, width, height int) table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Type", Width: 8},
		{Title: "Plan", Width: 14},
		{Title: "Region", Width: 14},
		{Title: "Role", Width: 10},
	}

	rows := make([]table.Row, len(apps))
	for i, a := range apps {
		rows[i] = table.Row{
			a.App.ID,
			a.App.Type,
			a.App.Plan,
			a.App.Region,
			a.App.Meta.Role,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight(height, len(rows))),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}

func tableHeight(termHeight, numRows int) int {
	h := termHeight - 4
	minH := numRows + 1
	if h < minH {
		h = minH
	}
	return h
}
