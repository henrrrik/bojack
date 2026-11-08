package view

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/runway-tui/internal/msg"
)

// field is one label/value row of an item view.
type field struct {
	label, value string
}

// ItemModel shows one T, fetched by name, as label/value rows. It is filled
// by a msg.ItemResultMsg[T].
type ItemModel[T any] struct {
	noun   string // singular, e.g. "template"
	name   string
	fields func(T) []field
	// itemActions, if set, lists the actions offered on the loaded item.
	itemActions func(T) []actionSpec
	actions     actionPrompt

	item    *T
	err     error
	loading bool
	spinner spinner.Model
	styles  DetailStyles
	width   int
	height  int
}

// NewItemModel creates a view of the item called name; fields lists what to show.
func NewItemModel[T any](noun, name string, fields func(T) []field,
	styles DetailStyles, width, height int) ItemModel[T] {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return ItemModel[T]{
		noun: noun, name: name, fields: fields,
		loading: true, spinner: s, styles: styles, width: width, height: height,
	}
}

// WithActions offers the actions itemActions returns for the loaded item.
func (m ItemModel[T]) WithActions(itemActions func(T) []actionSpec) ItemModel[T] {
	m.itemActions = itemActions
	return m
}

// currentActions are the actions offered on the loaded item.
func (m ItemModel[T]) currentActions() []actionSpec {
	if m.item == nil || m.itemActions == nil || m.err != nil {
		return nil
	}
	return m.itemActions(*m.item)
}

// ClaimsKey claims every key while an action is asking for input or
// confirmation.
func (m ItemModel[T]) ClaimsKey(string) bool {
	return m.actions.claimsKeys()
}

func (m ItemModel[T]) Update(teaMsg tea.Msg) (Viewer, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = teaMsg.Width, teaMsg.Height
	case msg.ItemResultMsg[T]:
		m.loading, m.item, m.err = false, teaMsg.Item, teaMsg.Err
		return m, nil
	case msg.ActionResultMsg:
		m.actions = m.actions.apply(teaMsg)
		return m, nil
	case tea.PasteMsg:
		m.actions = m.actions.paste(teaMsg)
		return m, nil
	case tea.KeyPressMsg:
		if p, cmd, ok := m.actions.handleKey(teaMsg, m.name, m.currentActions()...); ok {
			m.actions = p
			return m, cmd
		}
	}
	var cmd tea.Cmd
	if m.loading {
		m.spinner, cmd = m.spinner.Update(teaMsg)
	}
	return m, cmd
}

func (m ItemModel[T]) View() string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render(capitalize(m.noun)+": "+m.name) + "\n\n")
	switch {
	case m.loading:
		b.WriteString(m.spinner.View() + " Fetching " + m.noun + "...\n")
	case m.err != nil:
		b.WriteString(m.styles.Error.Render("Error: "+m.err.Error()) + "\n")
	case m.item != nil:
		b.WriteString(m.actions.withStatus(m.renderFields(m.fields(*m.item)), m.styles))
	}
	help := append([]string{"r: refresh"}, m.actions.help(m.currentActions()...)...)
	b.WriteString("\n" + renderHelp(m.styles.HelpBar, append(help, "esc: back")...))
	return b.String()
}

// renderFields renders label/value rows. A multi-line value starts on the
// line after its label.
func (m ItemModel[T]) renderFields(fields []field) string {
	var b strings.Builder
	for _, f := range fields {
		v := strings.TrimRight(f.value, "\n")
		if strings.Contains(v, "\n") {
			b.WriteString(m.styles.Label.Render(f.label) + "\n" + m.styles.Value.Render(v) + "\n")
			continue
		}
		b.WriteString(m.styles.Label.Render(f.label) + m.styles.Value.Render(v) + "\n")
	}
	return b.String()
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FetchItem returns a command that runs fetch and wraps the result for ItemModel[T].
func FetchItem[T any](ctx context.Context, fetch func(context.Context) (*T, error)) tea.Cmd {
	return func() tea.Msg {
		item, err := fetch(ctx)
		return msg.ItemResultMsg[T]{Item: item, Err: err}
	}
}
