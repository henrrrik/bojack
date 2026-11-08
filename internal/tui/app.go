package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
	"github.com/henrrrik/runway-tui/internal/tui/view"
)

// Model is the root bubbletea model.
type Model struct {
	runner      runway.Runner
	keys        KeyMap
	styles      Styles
	width       int
	height      int
	currentView msg.ViewID
	viewStack   []msg.ViewID
	paramStack  []string
	// statsInterval is the stats interval shared by app and service detail
	// views; it persists across refresh and navigation.
	statsInterval string
	// logOptions are the app logs filters; they persist like statsInterval.
	logOptions runway.LogOptions

	mainMenu      view.MainMenuModel
	appDetail     view.AppDetailModel
	serviceDetail view.ServiceDetailModel
	// list is the current generic list view (see listViews).
	list view.Viewer
}

// listViews describes the generic list views.
var listViews = map[msg.ViewID]listEntry{
	msg.ViewAppList:      listView(view.NewAppListModel, runway.Runner.AppList),
	msg.ViewServiceList:  listView(view.NewServiceListModel, runway.Runner.ServiceList),
	msg.ViewVolumeList:   listView(view.NewVolumeListModel, runway.Runner.VolumeList),
	msg.ViewDomainList:   listView(view.NewDomainListModel, runway.Runner.DomainList),
	msg.ViewPolicyList:   listView(view.NewPolicyListModel, runway.Runner.PolicyList),
	msg.ViewRegionList:   listView(view.NewRegionListModel, runway.Runner.RegionList),
	msg.ViewPlanList:     listView(view.NewPlanListModel, runway.Runner.PlanList),
	msg.ViewTemplateList: listView(view.NewTemplateListModel, runway.Runner.TemplateList),
	msg.ViewRegistryList: listView(view.NewRegistryListModel, runway.Runner.RegistryList),
	msg.ViewSSHKeyList:   listView(view.NewSSHKeyListModel, runway.Runner.SSHKeyList),
	msg.ViewAPIKeyList:   listView(view.NewAPIKeyListModel, runway.Runner.APIKeyList),
}

// listEntry builds a list view and, separately, the command that (re)fills
// it, so a list can be refetched without being rebuilt.
type listEntry struct {
	build func(Model) view.Viewer
	fetch func(Model) tea.Cmd
}

// listView pairs a list view constructor with the Runner method that fills it.
func listView[T any](
	newModel func(view.DetailStyles, int, int) view.ListModel[T],
	fetch func(runway.Runner, context.Context) ([]T, error),
) listEntry {
	return listEntry{
		build: func(m Model) view.Viewer { return newModel(m.styles.DetailStyles(), m.width, m.height) },
		fetch: func(m Model) tea.Cmd {
			r := m.runner
			return view.FetchList(view.CmdContext(), func(ctx context.Context) ([]T, error) { return fetch(r, ctx) })
		},
	}
}

// actionRunners maps each action to the Runner method that performs it.
var actionRunners = map[msg.ActionKind]func(runway.Runner, context.Context, string) error{
	msg.ActionRestartApp:     runway.Runner.AppRestart,
	msg.ActionRestartService: runway.Runner.ServiceRestart,
	msg.ActionCreateBackup:   runway.Runner.ServiceBackupCreate,
	msg.ActionVerifyDomain:   runway.Runner.DomainVerify,
}

// NewModel creates the root model.
func NewModel(runner runway.Runner) Model {
	keys := DefaultKeyMap()
	styles := DefaultStyles()
	return Model{
		runner:        runner,
		keys:          keys,
		styles:        styles,
		currentView:   msg.ViewMainMenu,
		statsInterval: view.DefaultStatsInterval,
		logOptions:    view.DefaultLogOptions,
		mainMenu:      view.NewMainMenuModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
		return m, m.propagateSize()
	case msg.NavigateMsg:
		return m.handleNavigate(teaMsg)
	case msg.NavigateBackMsg:
		return m.handleBack()
	case msg.StatsIntervalMsg:
		return m.handleStatsInterval(teaMsg)
	case msg.LogOptionsMsg:
		return m.handleLogOptions(teaMsg)
	case msg.ActionRequestMsg:
		return m, m.actionCmd(teaMsg)
	case msg.ActionResultMsg:
		return m.handleActionResult(teaMsg)
	case tea.KeyMsg:
		return m.handleGlobalKey(teaMsg)
	}
	return m.updateCurrentView(teaMsg)
}

func (m Model) View() string {
	switch m.currentView {
	case msg.ViewMainMenu:
		return m.mainMenu.View()
	case msg.ViewAppDetail:
		return m.appDetail.View()
	case msg.ViewServiceDetail:
		return m.serviceDetail.View()
	default:
		return m.viewLeaf()
	}
}

// viewLeaf renders the current generic list view.
func (m Model) viewLeaf() string {
	if _, ok := listViews[m.currentView]; ok && m.list != nil {
		return m.list.View()
	}
	return m.mainMenu.View()
}

func (m Model) handleGlobalKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(k, m.keys.Quit):
		if m.currentView == msg.ViewMainMenu {
			return m, tea.Quit
		}
		return m.handleBack()
	case key.Matches(k, m.keys.Back):
		if m.currentView != msg.ViewMainMenu {
			return m.handleBack()
		}
	case key.Matches(k, m.keys.Refresh):
		return m.handleRefresh()
	case key.Matches(k, m.keys.Open):
		return m.handleOpen()
	}
	return m.updateCurrentView(k)
}

func (m Model) handleRefresh() (tea.Model, tea.Cmd) {
	param := m.currentParam()
	return m.initView(m.currentView, param)
}

func (m Model) handleOpen() (tea.Model, tea.Cmd) {
	if m.currentView != msg.ViewAppDetail {
		return m.updateCurrentView(tea.KeyMsg{})
	}
	param := m.currentParam()
	runner := m.runner
	return m, func() tea.Msg {
		_ = runner.AppOpen(view.CmdContext(), param)
		return nil
	}
}

// handleStatsInterval records the new interval and refetches stats for the
// current detail view.
func (m Model) handleStatsInterval(iv msg.StatsIntervalMsg) (tea.Model, tea.Cmd) {
	m.statsInterval = iv.Interval
	ctx, param := view.CmdContext(), m.currentParam()
	switch m.currentView {
	case msg.ViewAppDetail:
		m.appDetail = m.appDetail.WithStatsInterval(iv.Interval)
		return m, view.FetchAppStats(ctx, m.runner, param, iv.Interval)
	case msg.ViewServiceDetail:
		m.serviceDetail = m.serviceDetail.WithStatsInterval(iv.Interval)
		return m, view.FetchServiceStats(ctx, m.runner, param, iv.Interval)
	}
	return m, nil
}

// handleLogOptions records new log options and refetches the current app's logs.
func (m Model) handleLogOptions(lo msg.LogOptionsMsg) (tea.Model, tea.Cmd) {
	m.logOptions = lo.Options
	if m.currentView != msg.ViewAppDetail {
		return m, nil
	}
	m.appDetail = m.appDetail.WithLogOptions(lo.Options)
	return m, view.FetchAppLogs(view.CmdContext(), m.runner, m.currentParam(), lo.Options)
}

// actionCmd runs an action and reports the result.
func (m Model) actionCmd(req msg.ActionRequestMsg) tea.Cmd {
	r, run := m.runner, actionRunners[req.Kind]
	return func() tea.Msg {
		return msg.ActionResultMsg{Kind: req.Kind, Target: req.Target, Err: run(r, view.CmdContext(), req.Target)}
	}
}

// handleActionResult passes the result to the current view and, after a
// success, refetches the view's data so it reflects the change.
func (m Model) handleActionResult(res msg.ActionResultMsg) (tea.Model, tea.Cmd) {
	next, cmd := m.updateCurrentView(res)
	m = next.(Model)
	if res.Err != nil {
		return m, cmd
	}
	return m, tea.Batch(cmd, m.refetchCurrent())
}

// refetchCurrent reloads the current view's data without rebuilding it.
func (m Model) refetchCurrent() tea.Cmd {
	ctx, param := view.CmdContext(), m.currentParam()
	switch m.currentView {
	case msg.ViewAppDetail:
		return view.FetchAppDetail(ctx, m.runner, param, m.statsInterval, m.logOptions)
	case msg.ViewServiceDetail:
		return view.FetchServiceDetail(ctx, m.runner, param, m.statsInterval)
	}
	if e, ok := listViews[m.currentView]; ok {
		return e.fetch(m)
	}
	return nil
}

func (m Model) currentParam() string {
	if len(m.paramStack) > 0 {
		return m.paramStack[len(m.paramStack)-1]
	}
	return ""
}

func (m Model) handleNavigate(nav msg.NavigateMsg) (tea.Model, tea.Cmd) {
	m.viewStack = append(m.viewStack, m.currentView)
	m.paramStack = append(m.paramStack, nav.Param)
	m.currentView = nav.View
	m, cmd := m.initView(nav.View, nav.Param)
	return m, cmd
}

func (m Model) handleBack() (tea.Model, tea.Cmd) {
	if len(m.viewStack) == 0 {
		return m, nil
	}
	n := len(m.viewStack) - 1
	m.currentView = m.viewStack[n]
	m.viewStack = m.viewStack[:n]
	m.paramStack = m.paramStack[:n]
	return m, nil
}

func (m Model) initView(v msg.ViewID, param string) (Model, tea.Cmd) {
	if e, ok := listViews[v]; ok {
		m.list = e.build(m)
		return m, e.fetch(m)
	}
	ctx := view.CmdContext()
	switch v {
	case msg.ViewAppDetail:
		m.appDetail = view.NewAppDetailModel(param, m.styles.DetailStyles(), m.width, m.height).
			WithStatsInterval(m.statsInterval).
			WithLogOptions(m.logOptions)
		return m, view.FetchAppDetail(ctx, m.runner, param, m.statsInterval, m.logOptions)
	case msg.ViewServiceDetail:
		m.serviceDetail = view.NewServiceDetailModel(param, m.styles.DetailStyles(), m.width, m.height).
			WithStatsInterval(m.statsInterval)
		return m, view.FetchServiceDetail(ctx, m.runner, param, m.statsInterval)
	}
	return m, nil
}

func (m Model) updateCurrentView(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentView {
	case msg.ViewMainMenu:
		m.mainMenu, cmd = m.mainMenu.Update(teaMsg)
	case msg.ViewAppDetail:
		m.appDetail, cmd = m.appDetail.Update(teaMsg)
	case msg.ViewServiceDetail:
		m.serviceDetail, cmd = m.serviceDetail.Update(teaMsg)
	default:
		return m.updateLeafView(teaMsg)
	}
	return m, cmd
}

func (m Model) updateLeafView(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if _, ok := listViews[m.currentView]; ok && m.list != nil {
		m.list, cmd = m.list.Update(teaMsg)
	}
	return m, cmd
}

func (m Model) propagateSize() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{Width: m.width, Height: m.height}
	}
}
