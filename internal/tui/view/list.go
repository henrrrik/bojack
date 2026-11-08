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
	// onSelect, if set, builds the message sent when enter is pressed on an item.
	onSelect func(T) tea.Msg

	table   table.Model
	spinner spinner.Model
	styles  DetailStyles
	// actions run on the selected item; actionTarget names it.
	actions      actionPrompt
	actionTarget func(T) string
	items        []T
	loading      bool
	err          error
	width        int
	height       int
}

// NewListModel creates a list view of T with the given columns and row builder.
func NewListModel[T any](noun string, columns []table.Column, row func(T) table.Row,
	styles DetailStyles, width, height int) ListModel[T] {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return ListModel[T]{
		noun: noun, columns: columns, row: row,
		spinner: s, styles: styles, loading: true, width: width, height: height,
	}
}

// WithSelect makes enter on an item send the message built by onSelect.
func (m ListModel[T]) WithSelect(onSelect func(T) tea.Msg) ListModel[T] {
	m.onSelect = onSelect
	return m
}

// WithActions offers actions on the selected item, which target names.
func (m ListModel[T]) WithActions(target func(T) string, specs ...actionSpec) ListModel[T] {
	m.actions, m.actionTarget = newActionPrompt(specs...), target
	return m
}

func (m ListModel[T]) selectedTarget() string {
	if item, ok := m.selected(); ok && m.actionTarget != nil {
		return m.actionTarget(item)
	}
	return ""
}

// handleKey runs actions and selection. ok is false if the key is left for
// the table (navigation).
func (m ListModel[T]) handleKey(k string) (ListModel[T], tea.Cmd, bool) {
	if p, cmd, ok := m.actions.handleKey(k, m.selectedTarget()); ok {
		m.actions = p
		return m, cmd, true
	}
	if k == "enter" && m.onSelect != nil {
		if item, ok := m.selected(); ok {
			return m, func() tea.Msg { return m.onSelect(item) }, true
		}
	}
	return m, nil, false
}

// selected returns the item under the table cursor.
func (m ListModel[T]) selected() (T, bool) {
	var zero T
	i := m.table.Cursor()
	if m.loading || i < 0 || i >= len(m.items) {
		return zero, false
	}
	return m.items[i], true
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
	case msg.ActionResultMsg:
		m.actions = m.actions.apply(teaMsg)
		return m, nil
	case tea.KeyMsg:
		if next, cmd, ok := m.handleKey(teaMsg.String()); ok {
			return next, cmd
		}
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
			"\n\n" + renderHelp(m.styles.HelpBar, "r: refresh", "esc: back")
	case len(m.items) == 0:
		return "No " + m.noun + " found.\n\n" + renderHelp(m.styles.HelpBar, "r: refresh", "esc: back")
	}
	help := append(m.actions.help(), "r: refresh", "esc: back")
	if m.onSelect != nil {
		help = append([]string{"enter: details"}, help...)
	}
	return m.actions.withStatus(m.table.View(), m.styles) + "\n" + renderHelp(m.styles.HelpBar, help...)
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

// tableHeight sizes a table to the terminal, but at least tall enough for all rows.
func tableHeight(termHeight, numRows int) int {
	h := termHeight - 4
	minH := numRows + 1
	if h < minH {
		h = minH
	}
	return h
}
