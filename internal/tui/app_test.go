package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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

func (f *fakeStatsRunner) TemplateShow(_ context.Context, name string) (*runway.Template, error) {
	f.calls = append(f.calls, "template "+name)
	return &runway.Template{Name: name, Category: "cache"}, nil
}

func (f *fakeStatsRunner) VolumeShow(_ context.Context, name string) (*runway.Volume, error) {
	f.calls = append(f.calls, "volume "+name)
	return &runway.Volume{Name: name, Mountpoint: "/data"}, nil
}

func (f *fakeStatsRunner) AppConfigSet(_ context.Context, app, assignment string) error {
	f.calls = append(f.calls, "app config set "+app+" "+assignment)
	return nil
}

func (f *fakeStatsRunner) AppConfigRm(_ context.Context, app, key string) error {
	f.calls = append(f.calls, "app config rm "+app+" "+key)
	return nil
}

func (f *fakeStatsRunner) ServiceConfigSet(_ context.Context, name, assignment string) error {
	f.calls = append(f.calls, "service config set "+name+" "+assignment)
	return nil
}

func (f *fakeStatsRunner) ServiceConfigRm(_ context.Context, name, key string) error {
	f.calls = append(f.calls, "service config rm "+name+" "+key)
	return nil
}

func (f *fakeStatsRunner) AppPlanSet(_ context.Context, app, plan string) error {
	f.calls = append(f.calls, "plan "+app+" "+plan)
	return nil
}

func (f *fakeStatsRunner) VolumeSet(_ context.Context, name string, sizeMiB int) error {
	f.calls = append(f.calls, fmt.Sprintf("volume set %s %d", name, sizeMiB))
	return nil
}

func (f *fakeStatsRunner) ServiceBackupCreate(_ context.Context, name string) error {
	f.calls = append(f.calls, "backup "+name)
	return nil
}

func (f *fakeStatsRunner) DomainVerify(_ context.Context, domain string) error {
	f.calls = append(f.calls, "verify "+domain)
	return nil
}

func (f *fakeStatsRunner) DomainList(context.Context) ([]runway.RootDomain, error) {
	return []runway.RootDomain{{Domain: "example.com"}}, nil
}

func (f *fakeStatsRunner) AppRestart(_ context.Context, name string) error {
	f.calls = append(f.calls, "restart app "+name)
	return nil
}

func (f *fakeStatsRunner) ServiceRestart(_ context.Context, name string) error {
	f.calls = append(f.calls, "restart service "+name)
	return nil
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
	result, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
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
	result, _ := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
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
	if v := m.View().Content; !strings.Contains(v, "Fetching templates") {
		t.Errorf("view = %q, want loading", v)
	}
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
	result, _ = m.Update(cmd())
	if v := result.(Model).View().Content; !strings.Contains(v, "valkey") {
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
		if _, ok := genericViews[v]; !ok {
			t.Errorf("view %d not registered", v)
		}
	}
}

func TestActionRequestCallsRunner(t *testing.T) {
	for kind, want := range map[msg.ActionKind]string{
		msg.ActionRestartApp:     "restart app x",
		msg.ActionRestartService: "restart service x",
		msg.ActionCreateBackup:   "backup x",
		msg.ActionVerifyDomain:   "verify x",
	} {
		fake := &fakeStatsRunner{}
		_, cmd := NewModel(fake).Update(msg.ActionRequestMsg{Kind: kind, Target: "x"})
		if cmd == nil {
			t.Fatalf("kind %d: expected a command", kind)
		}
		res, ok := cmd().(msg.ActionResultMsg)
		if !ok || res.Kind != kind || res.Target != "x" || res.Err != nil {
			t.Errorf("kind %d: cmd() = %#v", kind, res)
		}
		if len(fake.calls) != 1 || fake.calls[0] != want {
			t.Errorf("kind %d: calls = %v, want [%s]", kind, fake.calls, want)
		}
	}
}

func pressKeys(m tea.Model, keys ...string) tea.Model {
	for _, k := range keys {
		m, _ = m.Update(tea.KeyPressMsg{Code: []rune(k)[0], Text: k})
	}
	return m
}

func TestActionSuccessRefetchesDetail(t *testing.T) {
	result, _ := NewModel(&fakeStatsRunner{}).handleNavigate(msg.NavigateMsg{View: msg.ViewAppDetail, Param: "terry"})
	result = pressKeys(result, "R", "y")
	result, cmd := result.Update(msg.ActionResultMsg{Kind: msg.ActionRestartApp, Target: "terry"})
	if cmd == nil {
		t.Error("a successful action should refetch the view")
	}
	if !strings.Contains(result.View().Content, "Restarted terry.") {
		t.Errorf("result should reach the view, got:\n%s", result.View().Content)
	}
	if _, cmd = result.Update(msg.ActionResultMsg{Kind: msg.ActionRestartApp, Target: "terry", Err: errors.New("x")}); cmd != nil {
		t.Error("a failed action should not refetch")
	}
}

func TestActionSuccessRefetchesList(t *testing.T) {
	fake := &fakeStatsRunner{}
	result, cmd := NewModel(fake).handleNavigate(msg.NavigateMsg{View: msg.ViewDomainList})
	result, _ = result.Update(cmd())
	result = pressKeys(result, "V", "y")
	result, cmd = result.Update(msg.ActionResultMsg{Kind: msg.ActionVerifyDomain, Target: "example.com"})
	if cmd == nil {
		t.Fatal("a successful action should refetch the list")
	}
	if _, ok := cmd().(msg.ListResultMsg[runway.RootDomain]); !ok {
		t.Errorf("refetch should reload domains, got %#v", cmd())
	}
	if !strings.Contains(result.View().Content, "Verification check for example.com done.") {
		t.Errorf("result should reach the list, got:\n%s", result.View().Content)
	}
}

func TestFilterKeysReachListBeforeGlobalKeys(t *testing.T) {
	fake := &fakeStatsRunner{}
	result, cmd := NewModel(fake).handleNavigate(msg.NavigateMsg{View: msg.ViewTemplateList})
	result, _ = result.Update(cmd())
	result = pressKeys(result, "/", "q", "r")
	m := result.(Model)
	if m.currentView != msg.ViewTemplateList || !strings.Contains(m.View().Content, "Filter: qr") {
		t.Fatalf("q and r should type into the filter, view=%d:\n%s", m.currentView, m.View().Content)
	}
	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m = result.(Model); m.currentView != msg.ViewTemplateList || strings.Contains(m.View().Content, "Filter:") {
		t.Fatalf("first esc should clear the filter and stay, view=%d", m.currentView)
	}
	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if result.(Model).currentView != msg.ViewMainMenu {
		t.Error("second esc should go back")
	}
}

func TestTemplateDetailFromListAndBack(t *testing.T) {
	fake := &fakeStatsRunner{}
	result, cmd := NewModel(fake).handleNavigate(msg.NavigateMsg{View: msg.ViewTemplateList})
	result, _ = result.Update(cmd())
	result, cmd = result.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	result, cmd = result.Update(cmd()) // NavigateMsg to the detail view
	if m := result.(Model); m.currentView != msg.ViewTemplateDetail || m.currentParam() != "valkey" {
		t.Fatalf("enter should open the template detail, view=%d param=%q", m.currentView, m.currentParam())
	}
	result, _ = result.Update(cmd())
	if v := result.View().Content; !strings.Contains(v, "Template: valkey") || !strings.Contains(v, "cache") {
		t.Errorf("detail view should show the template, got:\n%s", v)
	}
	if strings.Join(fake.calls, ",") != "templates,template valkey" {
		t.Errorf("calls = %v", fake.calls)
	}
	result, _ = result.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m := result.(Model); m.currentView != msg.ViewTemplateList || !strings.Contains(m.View().Content, "valkey") ||
		strings.Contains(m.View().Content, "Template: valkey") {
		t.Errorf("esc should return to the loaded list, view=%d:\n%s", m.currentView, m.View().Content)
	}
}

func TestVolumeDetailRefresh(t *testing.T) {
	fake := &fakeStatsRunner{}
	result, cmd := NewModel(fake).handleNavigate(msg.NavigateMsg{View: msg.ViewVolumeDetail, Param: "data-1"})
	result, _ = result.Update(cmd())
	result, cmd = result.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if cmd == nil {
		t.Fatal("r should refetch the volume")
	}
	result, _ = result.Update(cmd())
	if v := result.View().Content; !strings.Contains(v, "Volume: data-1") || !strings.Contains(v, "/data") {
		t.Errorf("view = %q", v)
	}
	if strings.Join(fake.calls, ",") != "volume data-1,volume data-1" {
		t.Errorf("calls = %v", fake.calls)
	}
}

func TestInputActionRequestsCallRunner(t *testing.T) {
	for _, c := range []struct {
		req  msg.ActionRequestMsg
		want string
	}{
		{msg.ActionRequestMsg{Kind: msg.ActionSetAppConfig, Target: "x", Arg: "A=1"}, "app config set x A=1"},
		{msg.ActionRequestMsg{Kind: msg.ActionRemoveAppConfig, Target: "x", Arg: "A"}, "app config rm x A"},
		{msg.ActionRequestMsg{Kind: msg.ActionSetServiceConfig, Target: "x", Arg: "A=1"}, "service config set x A=1"},
		{msg.ActionRequestMsg{Kind: msg.ActionRemoveServiceConfig, Target: "x", Arg: "A"}, "service config rm x A"},
		{msg.ActionRequestMsg{Kind: msg.ActionSetAppPlan, Target: "x", Arg: "starter"}, "plan x starter"},
		{msg.ActionRequestMsg{Kind: msg.ActionResizeVolume, Target: "x", Arg: "6144"}, "volume set x 6144"},
	} {
		fake := &fakeStatsRunner{}
		_, cmd := NewModel(fake).Update(c.req)
		res, ok := cmd().(msg.ActionResultMsg)
		if !ok || res.Kind != c.req.Kind || res.Target != "x" || res.Err != nil {
			t.Errorf("%+v: cmd() = %#v", c.req, res)
		}
		if len(fake.calls) != 1 || fake.calls[0] != c.want {
			t.Errorf("%+v: calls = %v, want [%s]", c.req, fake.calls, c.want)
		}
	}
}

func TestResizeRejectsNonNumericArg(t *testing.T) {
	fake := &fakeStatsRunner{}
	_, cmd := NewModel(fake).Update(msg.ActionRequestMsg{Kind: msg.ActionResizeVolume, Target: "x", Arg: "big"})
	if res := cmd().(msg.ActionResultMsg); res.Err == nil || len(fake.calls) != 0 {
		t.Errorf("a non-numeric size should fail without calling the CLI, res=%+v calls=%v", res, fake.calls)
	}
}

func TestDetailInputGetsGlobalKeys(t *testing.T) {
	result, _ := NewModel(&fakeStatsRunner{}).handleNavigate(msg.NavigateMsg{View: msg.ViewAppDetail, Param: "terry"})
	result = pressKeys(result, "P", "q", "r", "o")
	m := result.(Model)
	if m.currentView != msg.ViewAppDetail || !strings.Contains(m.View().Content, "qro") {
		t.Fatalf("q, r and o should type into the input, view=%d:\n%s", m.currentView, m.View().Content)
	}
	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m = result.(Model); m.currentView != msg.ViewAppDetail || strings.Contains(m.View().Content, "qro") {
		t.Fatalf("esc should cancel the input and stay, view=%d", m.currentView)
	}
	result, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if result.(Model).currentView != msg.ViewMainMenu {
		t.Error("second esc should go back")
	}
}

func TestVolumeResizeRefetches(t *testing.T) {
	fake := &fakeStatsRunner{}
	result, cmd := NewModel(fake).handleNavigate(msg.NavigateMsg{View: msg.ViewVolumeDetail, Param: "data-1"})
	result, _ = result.Update(msg.ItemResultMsg[runway.Volume]{Item: &runway.Volume{Name: "data-1", SizeMB: 100, Resizable: true}})
	_ = cmd
	result = pressKeys(result, "S", "200")
	result, _ = result.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	result, cmd = result.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	result, cmd = result.Update(cmd())
	result, cmd = result.Update(cmd()) // ActionResultMsg
	if cmd == nil {
		t.Fatal("a successful resize should refetch the volume")
	}
	if _, ok := cmd().(msg.ItemResultMsg[runway.Volume]); !ok {
		t.Errorf("refetch should reload the volume")
	}
	if !strings.Contains(result.View().Content, "Resized data-1") {
		t.Errorf("view:\n%s", result.View().Content)
	}
	if strings.Join(fake.calls, ",") != "volume set data-1 200,volume data-1" {
		t.Errorf("calls = %v", fake.calls)
	}
}

func TestViewIsFullScreen(t *testing.T) {
	if !NewModel(nil).View().AltScreen {
		t.Error("the TUI should use the alternate screen")
	}
}
