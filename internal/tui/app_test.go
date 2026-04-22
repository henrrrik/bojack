package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
)

func TestInitialViewIsMainMenu(t *testing.T) {
	m := NewModel(nil)
	if m.currentView != msg.ViewMainMenu {
		t.Errorf("currentView = %d, want ViewMainMenu", m.currentView)
	}
}

func TestNavigatePushesStack(t *testing.T) {
	m := NewModel(nil)
	nav := msg.NavigateMsg{View: msg.ViewAppList, Param: ""}
	result, _ := m.handleNavigate(nav)
	rm := result.(Model)
	if rm.currentView != msg.ViewAppList {
		t.Errorf("currentView = %d, want ViewAppList", rm.currentView)
	}
	if len(rm.viewStack) != 1 {
		t.Fatalf("viewStack len = %d, want 1", len(rm.viewStack))
	}
	if rm.viewStack[0] != msg.ViewMainMenu {
		t.Errorf("viewStack[0] = %d, want ViewMainMenu", rm.viewStack[0])
	}
}

func TestBackPopsStack(t *testing.T) {
	m := NewModel(nil)
	m.currentView = msg.ViewAppList
	m.viewStack = []msg.ViewID{msg.ViewMainMenu}
	m.paramStack = []string{""}
	result, _ := m.handleBack()
	rm := result.(Model)
	if rm.currentView != msg.ViewMainMenu {
		t.Errorf("currentView = %d, want ViewMainMenu", rm.currentView)
	}
	if len(rm.viewStack) != 0 {
		t.Errorf("viewStack len = %d, want 0", len(rm.viewStack))
	}
}

func TestBackOnEmptyStackIsNoop(t *testing.T) {
	m := NewModel(nil)
	result, _ := m.handleBack()
	rm := result.(Model)
	if rm.currentView != msg.ViewMainMenu {
		t.Errorf("currentView = %d, want ViewMainMenu", rm.currentView)
	}
}

func TestQuitFromMainMenu(t *testing.T) {
	m := NewModel(nil)
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = result
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
}

func TestQuitFromSubViewGoesBack(t *testing.T) {
	m := NewModel(nil)
	m.currentView = msg.ViewAppList
	m.viewStack = []msg.ViewID{msg.ViewMainMenu}
	m.paramStack = []string{""}
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	rm := result.(Model)
	if rm.currentView != msg.ViewMainMenu {
		t.Errorf("currentView = %d, want ViewMainMenu", rm.currentView)
	}
}

func TestWindowSizeUpdatesModel(t *testing.T) {
	m := NewModel(nil)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	rm := result.(Model)
	if rm.width != 120 {
		t.Errorf("width = %d, want 120", rm.width)
	}
	if rm.height != 40 {
		t.Errorf("height = %d, want 40", rm.height)
	}
}

func TestNestedNavigation(t *testing.T) {
	m := NewModel(nil)
	// Navigate: main menu -> app list -> app detail
	r1, _ := m.handleNavigate(msg.NavigateMsg{View: msg.ViewAppList})
	m1 := r1.(Model)

	r2, _ := m1.handleNavigate(msg.NavigateMsg{View: msg.ViewAppDetail, Param: "my-app"})
	m2 := r2.(Model)

	if m2.currentView != msg.ViewAppDetail {
		t.Errorf("currentView = %d, want ViewAppDetail", m2.currentView)
	}
	if len(m2.viewStack) != 2 {
		t.Fatalf("viewStack len = %d, want 2", len(m2.viewStack))
	}

	// Back to app list
	r3, _ := m2.handleBack()
	m3 := r3.(Model)
	if m3.currentView != msg.ViewAppList {
		t.Errorf("currentView = %d, want ViewAppList", m3.currentView)
	}

	// Back to main menu
	r4, _ := m3.handleBack()
	m4 := r4.(Model)
	if m4.currentView != msg.ViewMainMenu {
		t.Errorf("currentView = %d, want ViewMainMenu", m4.currentView)
	}
}
