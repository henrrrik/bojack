package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
	"github.com/henrrrik/runway-tui/internal/tui/view"
)

// fakeStatsRunner implements runway.Runner; only the stats methods are
// usable; any other call panics on the nil embedded interface.
type fakeStatsRunner struct {
	runway.Runner
	calls []string
}

func (f *fakeStatsRunner) AppStats(_ context.Context, name, interval string) (runway.Stats, error) {
	f.calls = append(f.calls, "app "+name+" "+interval)
	return runway.Stats{}, nil
}

func (f *fakeStatsRunner) AppLogs(_ context.Context, name string, opts runway.LogOptions) ([]runway.LogEntry, error) {
	f.calls = append(f.calls, fmt.Sprintf("logs %s %+v", name, opts))
	return nil, nil
}

func (f *fakeStatsRunner) TemplateList(context.Context) ([]runway.Template, error) {
	f.calls = append(f.calls, "templates")
	return []runway.Template{{Name: "valkey"}}, nil
}

func (f *fakeStatsRunner) ServiceStats(_ context.Context, name, interval string) (runway.Stats, error) {
	f.calls = append(f.calls, "service "+name+" "+interval)
	return runway.Stats{}, nil
}

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

func TestStatsIntervalMsgFetchesAppStats(t *testing.T) {
	fake := &fakeStatsRunner{}
	m := NewModel(fake)
	result, _ := m.handleNavigate(msg.NavigateMsg{View: msg.ViewAppDetail, Param: "my-app"})
	result, cmd := result.(Model).Update(msg.StatsIntervalMsg{Interval: "168h"})
	if cmd == nil {
		t.Fatal("expected a fetch command")
	}
	got, ok := cmd().(msg.AppStatsResultMsg)
	if !ok || got.Interval != "168h" {
		t.Errorf("cmd() = %#v, want AppStatsResultMsg for 168h", got)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "app my-app 168h" {
		t.Errorf("calls = %v", fake.calls)
	}
	if result.(Model).statsInterval != "168h" {
		t.Errorf("statsInterval = %q, want 168h", result.(Model).statsInterval)
	}
}

func TestStatsIntervalMsgFetchesServiceStats(t *testing.T) {
	fake := &fakeStatsRunner{}
	m := NewModel(fake)
	result, _ := m.handleNavigate(msg.NavigateMsg{View: msg.ViewServiceDetail, Param: "my-db"})
	_, cmd := result.(Model).Update(msg.StatsIntervalMsg{Interval: "1h"})
	if cmd == nil {
		t.Fatal("expected a fetch command")
	}
	if got, ok := cmd().(msg.ServiceStatsResultMsg); !ok || got.Interval != "1h" {
		t.Errorf("cmd() = %#v, want ServiceStatsResultMsg for 1h", got)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "service my-db 1h" {
		t.Errorf("calls = %v", fake.calls)
	}
}

func TestStatsIntervalPersistsAcrossNavigation(t *testing.T) {
	m := NewModel(&fakeStatsRunner{})
	if m.statsInterval != view.DefaultStatsInterval {
		t.Errorf("default statsInterval = %q", m.statsInterval)
	}
	result, _ := m.handleNavigate(msg.NavigateMsg{View: msg.ViewAppDetail, Param: "my-app"})
	result, _ = result.(Model).Update(msg.StatsIntervalMsg{Interval: "720h"})
	result, _ = result.(Model).handleNavigate(msg.NavigateMsg{View: msg.ViewServiceDetail, Param: "my-db"})
	if got := result.(Model).serviceDetail.StatsInterval(); got != "720h" {
		t.Errorf("service detail interval = %q, want 720h", got)
	}
}

func TestLogOptionsMsgFetchesLogs(t *testing.T) {
	fake := &fakeStatsRunner{}
	m := NewModel(fake)
	result, _ := m.handleNavigate(msg.NavigateMsg{View: msg.ViewAppDetail, Param: "my-app"})
	opts := runway.LogOptions{Limit: 500, NoSystem: true}
	result, cmd := result.(Model).Update(msg.LogOptionsMsg{Options: opts})
	if cmd == nil {
		t.Fatal("expected a fetch command")
	}
	if got, ok := cmd().(msg.AppLogsResultMsg); !ok || got.Options != opts {
		t.Errorf("cmd() = %#v, want AppLogsResultMsg for %+v", got, opts)
	}
	if want := fmt.Sprintf("logs my-app %+v", opts); len(fake.calls) != 1 || fake.calls[0] != want {
		t.Errorf("calls = %v, want [%s]", fake.calls, want)
	}
	// persists to the next app detail view
	result, _ = result.(Model).handleNavigate(msg.NavigateMsg{View: msg.ViewAppDetail, Param: "other"})
	if got := result.(Model).appDetail.LogOptions(); got != opts {
		t.Errorf("next app detail options = %+v, want %+v", got, opts)
	}
}

func TestNavigateToListViewFetchesAndRenders(t *testing.T) {
	fake := &fakeStatsRunner{}
	result, cmd := NewModel(fake).handleNavigate(msg.NavigateMsg{View: msg.ViewTemplateList})
	m := result.(Model)
	if v := m.View(); !strings.Contains(v, "Fetching templates") {
		t.Errorf("view = %q, want loading", v)
	}
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
	result, _ = m.Update(cmd())
	if v := result.(Model).View(); !strings.Contains(v, "valkey") {
		t.Errorf("view = %q, want template row", v)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "templates" {
		t.Errorf("calls = %v", fake.calls)
	}
}

func TestAllListViewsRegistered(t *testing.T) {
	for _, v := range []msg.ViewID{
		msg.ViewVolumeList, msg.ViewDomainList, msg.ViewPolicyList, msg.ViewRegionList, msg.ViewPlanList,
		msg.ViewTemplateList, msg.ViewRegistryList, msg.ViewSSHKeyList, msg.ViewAPIKeyList,
	} {
		if _, ok := listViews[v]; !ok {
			t.Errorf("view %d not registered", v)
		}
	}
}
