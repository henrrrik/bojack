package view

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

func typeKeys(v Viewer, keys ...string) Viewer {
	for _, k := range keys {
		var m tea.KeyPressMsg
		switch k {
		case "enter":
			m = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "esc":
			m = tea.KeyPressMsg{Code: tea.KeyEsc}
		case "backspace":
			m = tea.KeyPressMsg{Code: tea.KeyBackspace}
		default:
			m = key(k)
		}
		v, _ = v.Update(m)
	}
	return v
}

func pets() Viewer {
	v := NewListModel("pets", []table.Column{{Title: "Name", Width: 10}, {Title: "Kind", Width: 10}},
		func(p [2]string) table.Row { return table.Row{p[0], p[1]} }, testDetailStyles(), 80, 24).
		WithSelect(func(p [2]string) tea.Msg { return msg.NavigateMsg{Param: p[0]} })
	var vv Viewer = v
	return update(vv, msg.ListResultMsg[[2]string]{Items: [][2]string{{"Rex", "dog"}, {"Tom", "cat"}, {"Bella", "Dog"}}})
}

func TestFilterTypingNarrowsRows(t *testing.T) {
	v := typeKeys(pets(), "/", "d", "o", "g")
	out := v.View()
	if !strings.Contains(out, "Rex") || !strings.Contains(out, "Bella") || strings.Contains(out, "Tom") {
		t.Errorf("filter 'dog' should match any column, case-insensitively, got:\n%s", out)
	}
	if !strings.Contains(out, "Filter: dog") {
		t.Errorf("should show the filter being typed, got:\n%s", out)
	}
}

func TestFilterBackspaceAndApply(t *testing.T) {
	v := typeKeys(pets(), "/", "t", "x", "backspace", "enter")
	out := v.View()
	if !strings.Contains(out, "Tom") || strings.Contains(out, "Rex") {
		t.Errorf("filter should be 't' after backspace, got:\n%s", out)
	}
	if !strings.Contains(out, "Filter: t") || !strings.Contains(out, "esc: clear filter") {
		t.Errorf("applied filter should be shown with how to clear it, got:\n%s", out)
	}
}

func TestFilterEscClears(t *testing.T) {
	for _, keys := range [][]string{{"/", "t", "esc"}, {"/", "t", "enter", "esc"}} {
		out := typeKeys(pets(), keys...).View()
		if !strings.Contains(out, "Rex") || !strings.Contains(out, "Tom") || strings.Contains(out, "Filter:") {
			t.Errorf("%v: esc should clear the filter, got:\n%s", keys, out)
		}
	}
}

func TestFilterSelectUsesVisibleRows(t *testing.T) {
	v := typeKeys(pets(), "/", "b", "e", "enter")
	_, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should select")
	}
	if nav := cmd().(msg.NavigateMsg); nav.Param != "Bella" {
		t.Errorf("selected %q, want Bella (first visible row)", nav.Param)
	}
}

func TestFilterNoMatches(t *testing.T) {
	out := typeKeys(pets(), "/", "z", "z").View()
	if !strings.Contains(out, `No pets match "zz".`) {
		t.Errorf("got:\n%s", out)
	}
	if _, cmd := typeKeys(pets(), "/", "z", "enter").Update(tea.KeyPressMsg{Code: tea.KeyEnter}); cmd != nil {
		t.Error("enter with no visible rows should do nothing")
	}
}

func TestFilterSurvivesRefresh(t *testing.T) {
	v := typeKeys(pets(), "/", "c", "a", "t", "enter")
	v = update(v, msg.ListResultMsg[[2]string]{Items: [][2]string{{"Rex", "dog"}, {"Tom", "cat"}, {"Kit", "cat"}}})
	out := v.View()
	if !strings.Contains(out, "Kit") || strings.Contains(out, "Rex") || !strings.Contains(out, "Filter: cat") {
		t.Errorf("a refresh should keep the filter, got:\n%s", out)
	}
}

func TestFilterClaimsKeys(t *testing.T) {
	claims := func(v Viewer, k string) bool { return v.(KeyClaimer).ClaimsKey(k) }
	v := pets()
	for _, k := range []string{"q", "esc", "r", "backspace"} {
		if claims(v, k) {
			t.Errorf("without a filter the list should not claim %q", k)
		}
	}
	typing := typeKeys(v, "/")
	for _, k := range []string{"q", "esc", "r", "o", "backspace"} {
		if !claims(typing, k) {
			t.Errorf("while typing the list should claim %q", k)
		}
	}
	applied := typeKeys(v, "/", "t", "enter")
	if !claims(applied, "esc") || claims(applied, "q") || claims(applied, "backspace") {
		t.Error("with an applied filter the list should claim only esc")
	}
}

func TestFilterActionTargetsVisibleRow(t *testing.T) {
	v := typeKeys(domainList(), "/", "b", "enter")
	v, _ = v.Update(key("V"))
	if !strings.Contains(v.View(), "Re-check verification of b.com?") {
		t.Errorf("action should target the filtered row, got:\n%s", v.View())
	}
}

func TestFilterHelp(t *testing.T) {
	if !strings.Contains(pets().View(), "/: filter") {
		t.Error("help should offer /: filter")
	}
	if out := typeKeys(pets(), "/").View(); !strings.Contains(out, "enter: apply") || !strings.Contains(out, "esc: clear") {
		t.Errorf("help while typing should explain enter/esc, got:\n%s", out)
	}
}

var _ = runway.RootDomain{}
