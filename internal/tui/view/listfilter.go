package view

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

const filterKey = "/"

// KeyClaimer is implemented by views that sometimes need keys the root model
// would otherwise handle globally (q, esc, backspace, r, o), e.g. while the
// user types into a text field.
type KeyClaimer interface {
	ClaimsKey(key string) bool
}

// ClaimsKey claims every key while a filter is being typed or an action asks
// for confirmation, and esc (to clear it) while a filter is applied.
func (m ListModel[T]) ClaimsKey(k string) bool {
	return m.filtering || m.actions.claimsKeys() || (m.filter != "" && k == "esc")
}

// startFilter enters filter-typing mode.
func (m ListModel[T]) startFilter() ListModel[T] {
	m.filtering = true
	return m
}

// filterKeyInput edits the filter while typing. ok is false for keys it
// leaves to the table, such as arrows to move the selection.
func (m ListModel[T]) filterKeyInput(k tea.KeyPressMsg) (ListModel[T], bool) {
	switch k.Code {
	case tea.KeyEnter:
		m.filtering = false
	case tea.KeyEscape:
		return m.setFilter(""), true
	case tea.KeyBackspace:
		if r := []rune(m.filter); len(r) > 0 {
			return m.setFilter(string(r[:len(r)-1])), true
		}
	default:
		if k.Text == "" {
			return m, false
		}
		return m.setFilter(m.filter + k.Text), true
	}
	return m, true
}

// setFilter applies a filter and rebuilds the visible rows. An empty filter
// also ends typing.
func (m ListModel[T]) setFilter(f string) ListModel[T] {
	m.filter = f
	if f == "" {
		m.filtering = false
	}
	return m.refilter()
}

// refilter recomputes the rows matching the filter: case-insensitive, in any
// rendered column.
func (m ListModel[T]) refilter() ListModel[T] {
	m.visible = m.visible[:0:0]
	q := strings.ToLower(m.filter)
	for _, it := range m.items {
		if q == "" || rowMatches(m.row(it), q) {
			m.visible = append(m.visible, it)
		}
	}
	m.table = m.buildTable()
	return m
}

func rowMatches(cells []string, lowerQuery string) bool {
	for _, c := range cells {
		if strings.Contains(strings.ToLower(c), lowerQuery) {
			return true
		}
	}
	return false
}

// filterStatus renders the filter line, or "" if there is no filter.
func (m ListModel[T]) filterStatus() string {
	switch {
	case m.filtering:
		return m.styles.Value.Render("Filter: " + m.filter + "█")
	case m.filter != "":
		return m.styles.Value.Render("Filter: " + m.filter)
	}
	return ""
}

// listHelp lists the keys available in the current state.
func (m ListModel[T]) listHelp() []string {
	if m.filtering {
		return []string{"enter: apply", "esc: clear"}
	}
	help := append(m.actions.help(), filterKey+": filter", "r: refresh")
	if m.onSelect != nil {
		help = append([]string{"enter: details"}, help...)
	}
	if m.filter != "" {
		return append(help, "esc: clear filter")
	}
	return append(help, "esc: back")
}
