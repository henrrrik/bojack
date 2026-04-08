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

// PlanListModel displays a table of plans.
type PlanListModel struct {
	table   table.Model
	spinner spinner.Model
	helpBar lipgloss.Style
	plans   []runway.Plan
	loading bool
	err     error
	width   int
	height  int
}

// NewPlanListModel creates a new plan list model.
func NewPlanListModel(helpBar lipgloss.Style, width, height int) PlanListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return PlanListModel{
		spinner: s,
		helpBar: helpBar,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m PlanListModel) Update(teaMsg tea.Msg) (PlanListModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.PlanListResultMsg:
		m.loading = false
		if teaMsg.Err != nil {
			m.err = teaMsg.Err
			return m, nil
		}
		m.plans = teaMsg.Plans
		m.table = buildPlanTable(teaMsg.Plans, m.height)
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

func (m PlanListModel) View() string {
	if m.loading {
		return m.spinner.View() + " Fetching plans..."
	}
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) +
			"\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	}
	if len(m.plans) == 0 {
		return "No plans found.\n\n" + renderHelp(m.helpBar, "esc: back")
	}
	return m.table.View() + "\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
}

func buildPlanTable(plans []runway.Plan, height int) table.Model {
	columns := []table.Column{
		{Title: "Plan", Width: 14},
		{Title: "CPU", Width: 6},
		{Title: "RAM MB", Width: 8},
		{Title: "HDD MB", Width: 8},
		{Title: "BW Mbps", Width: 8},
		{Title: "Price/hr", Width: 10},
		{Title: "Description", Width: 20},
	}

	rows := make([]table.Row, len(plans))
	for i, p := range plans {
		rows[i] = table.Row{
			p.Plan,
			fmt.Sprintf("%d", p.CPU),
			fmt.Sprintf("%d", p.RAMMB),
			fmt.Sprintf("%d", p.HDDMB),
			fmt.Sprintf("%d", p.BandwidthMbits),
			fmt.Sprintf("%.4f", p.PriceHour),
			p.Description,
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
