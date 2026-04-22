package view

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// VolumeListModel displays a table of volumes.
type VolumeListModel struct {
	table   table.Model
	spinner spinner.Model
	helpBar lipgloss.Style
	volumes []runway.Volume
	loading bool
	err     error
	width   int
	height  int
}

// NewVolumeListModel creates a new volume list model.
func NewVolumeListModel(helpBar lipgloss.Style, width, height int) VolumeListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return VolumeListModel{
		spinner: s,
		helpBar: helpBar,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m VolumeListModel) Update(teaMsg tea.Msg) (VolumeListModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.VolumeListResultMsg:
		m.loading = false
		if teaMsg.Err != nil {
			m.err = teaMsg.Err
			return m, nil
		}
		m.volumes = teaMsg.Volumes
		m.table = buildVolumeTable(teaMsg.Volumes, m.height)
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

func (m VolumeListModel) View() string {
	if m.loading {
		return m.spinner.View() + " Fetching volumes..."
	}
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) +
			"\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	}
	if len(m.volumes) == 0 {
		return "No volumes found.\n\n" + renderHelp(m.helpBar, "esc: back")
	}
	return m.table.View() + "\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
}

func buildVolumeTable(volumes []runway.Volume, height int) table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "App", Width: 20},
		{Title: "Mountpoint", Width: 28},
		{Title: "Size", Width: 14},
	}

	rows := make([]table.Row, len(volumes))
	for i, v := range volumes {
		rows[i] = table.Row{v.Name, v.App, v.Mountpoint, v.Size}
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
