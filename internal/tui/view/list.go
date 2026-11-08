package view

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
)

// Viewer is a view the root model can hold without knowing its concrete type.
type Viewer interface {
	Update(tea.Msg) (Viewer, tea.Cmd)
	View() string
}

// ListModel is a read-only table of T, filled by a msg.ListResultMsg[T].
type ListModel[T any] struct {
	noun    string // plural, e.g. "templates"
	columns []table.Column
	row     func(T) table.Row

	table   table.Model
	spinner spinner.Model
	helpBar lipgloss.Style
	items   []T
	loading bool
	err     error
	width   int
	height  int
}

// NewListModel creates a list view of T with the given columns and row builder.
func NewListModel[T any](noun string, columns []table.Column, row func(T) table.Row,
	helpBar lipgloss.Style, width, height int) ListModel[T] {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return ListModel[T]{
		noun: noun, columns: columns, row: row,
		spinner: s, helpBar: helpBar, loading: true, width: width, height: height,
	}
}

func (m ListModel[T]) Update(teaMsg tea.Msg) (Viewer, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = teaMsg.Width, teaMsg.Height
	case msg.ListResultMsg[T]:
		m.loading, m.err, m.items = false, teaMsg.Err, teaMsg.Items
		if m.err == nil {
			m.table = m.buildTable()
		}
		return m, nil
	}
	var cmd tea.Cmd
	if m.loading {
		m.spinner, cmd = m.spinner.Update(teaMsg)
	} else {
		m.table, cmd = m.table.Update(teaMsg)
	}
	return m, cmd
}

func (m ListModel[T]) View() string {
	switch {
	case m.loading:
		return m.spinner.View() + " Fetching " + m.noun + "..."
	case m.err != nil:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) +
			"\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	case len(m.items) == 0:
		return "No " + m.noun + " found.\n\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
	}
	return m.table.View() + "\n" + renderHelp(m.helpBar, "r: refresh", "esc: back")
}

func (m ListModel[T]) buildTable() table.Model {
	rows := make([]table.Row, len(m.items))
	for i, it := range m.items {
		rows[i] = m.row(it)
	}
	t := table.New(
		table.WithColumns(m.columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight(m.height, len(rows))),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Bold(true)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	t.SetStyles(s)
	return t
}

// FetchList returns a command that runs fetch and wraps the result for ListModel[T].
func FetchList[T any](ctx context.Context, fetch func(context.Context) ([]T, error)) tea.Cmd {
	return func() tea.Msg {
		items, err := fetch(ctx)
		return msg.ListResultMsg[T]{Items: items, Err: err}
	}
}
