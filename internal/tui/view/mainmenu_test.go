package view

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/bojack/internal/msg"
)

func TestMainMenuRender(t *testing.T) {
	m := NewMainMenuModel()
	v := m.View()
	if !strings.Contains(v, "Bojack - Runway Manager") {
		t.Error("View should contain title")
	}
	if !strings.Contains(v, "Apps") {
		t.Error("View should contain Apps")
	}
	if !strings.Contains(v, "Services") {
		t.Error("View should contain Services")
	}
}

func TestMainMenuNavigation(t *testing.T) {
	m := NewMainMenuModel()
	if m.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.cursor)
	}

	// Move down
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}

	// Move up
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}

	// Can't go above 0
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestMainMenuSelect(t *testing.T) {
	m := NewMainMenuModel()
	// First item is Apps -> ViewAppList
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on enter, got nil")
	}
	result := cmd()
	nav, ok := result.(msg.NavigateMsg)
	if !ok {
		t.Fatalf("expected NavigateMsg, got %T", result)
	}
	if nav.View != msg.ViewAppList {
		t.Errorf("View = %d, want ViewAppList", nav.View)
	}
}

func TestMainMenuHasAccountLists(t *testing.T) {
	v := NewMainMenuModel().View()
	for _, want := range []string{"Templates", "Registries", "SSH Keys", "API Keys"} {
		if !strings.Contains(v, want) {
			t.Errorf("menu should contain %q", want)
		}
	}
}
