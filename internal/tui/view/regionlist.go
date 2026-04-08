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

// RegionListModel displays a table of regions.
type RegionListModel struct {
	table   table.Model
	spinner spinner.Model
	helpBar lipgloss.Style
	regions []runway.Region
	loading bool
	err     error
	width   int
	height  int
}

// NewRegionListModel creates a new region list model.
func NewRegionListModel(helpBar lipgloss.Style, width, height int) RegionListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return RegionListModel{
		spinner: s,
		helpBar: helpBar,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m RegionListModel) Update(teaMsg tea.Msg) (RegionListModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.RegionListResultMsg:
		m.loading = false
		if teaMsg.Err != nil {
			m.err = teaMsg.Err
			return m, nil
		}
		m.regions = teaMsg.Regions
		m.table = buildRegionTable(teaMsg.Regions, m.height)
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

func (m RegionListModel) View() string {
	if m.loading {
		return m.spinner.View() + " Fetching regions..."
	}
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) +
			"\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	}
	if len(m.regions) == 0 {
		return "No regions found.\n\n" + renderHelp(m.helpBar, "esc: back")
	}
	return m.table.View() + "\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
}

func buildRegionTable(regions []runway.Region, height int) table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 16},
		{Title: "Name", Width: 16},
		{Title: "Country", Width: 16},
		{Title: "Continent", Width: 14},
		{Title: "Live", Width: 6},
	}

	rows := make([]table.Row, len(regions))
	for i, r := range regions {
		rows[i] = table.Row{
			r.ID, r.Name, r.Country, r.Continent, fmt.Sprintf("%v", r.LiveStatus),
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
