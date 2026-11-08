package tui

import (
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
	appList       view.AppListModel
	appDetail     view.AppDetailModel
	serviceList   view.ServiceListModel
	serviceDetail view.ServiceDetailModel
	volumeList    view.VolumeListModel
	domainList    view.DomainListModel
	policyList    view.PolicyListModel
	regionList    view.RegionListModel
	planList      view.PlanListModel
	// list is the current generic list view (see listViews).
	list view.Viewer
}

// listViews builds the generic list views and the command that fills them.
var listViews = map[msg.ViewID]func(m Model) (view.Viewer, tea.Cmd){
	msg.ViewTemplateList: func(m Model) (view.Viewer, tea.Cmd) {
		return view.NewTemplateListModel(m.styles.HelpBar, m.width, m.height),
			view.FetchList(view.CmdContext(), m.runner.TemplateList)
	},
	msg.ViewRegistryList: func(m Model) (view.Viewer, tea.Cmd) {
		return view.NewRegistryListModel(m.styles.HelpBar, m.width, m.height),
			view.FetchList(view.CmdContext(), m.runner.RegistryList)
	},
	msg.ViewSSHKeyList: func(m Model) (view.Viewer, tea.Cmd) {
		return view.NewSSHKeyListModel(m.styles.HelpBar, m.width, m.height),
			view.FetchList(view.CmdContext(), m.runner.SSHKeyList)
	},
	msg.ViewAPIKeyList: func(m Model) (view.Viewer, tea.Cmd) {
		return view.NewAPIKeyListModel(m.styles.HelpBar, m.width, m.height),
			view.FetchList(view.CmdContext(), m.runner.APIKeyList)
	},
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
	case tea.KeyMsg:
		return m.handleGlobalKey(teaMsg)
	}
	return m.updateCurrentView(teaMsg)
}

func (m Model) View() string {
	switch m.currentView {
	case msg.ViewMainMenu:
		return m.mainMenu.View()
	case msg.ViewAppList:
		return m.appList.View()
	case msg.ViewAppDetail:
		return m.appDetail.View()
	case msg.ViewServiceList:
		return m.serviceList.View()
	case msg.ViewServiceDetail:
		return m.serviceDetail.View()
	default:
		return m.viewLeaf()
	}
}

// viewLeaf renders the simple list views.
func (m Model) viewLeaf() string {
	if _, ok := listViews[m.currentView]; ok && m.list != nil {
		return m.list.View()
	}
	switch m.currentView {
	case msg.ViewVolumeList:
		return m.volumeList.View()
	case msg.ViewDomainList:
		return m.domainList.View()
	case msg.ViewPolicyList:
		return m.policyList.View()
	case msg.ViewRegionList:
		return m.regionList.View()
	case msg.ViewPlanList:
		return m.planList.View()
	default:
		return m.mainMenu.View()
	}
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
	if build, ok := listViews[v]; ok {
		var cmd tea.Cmd
		m.list, cmd = build(m)
		return m, cmd
	}
	ctx := view.CmdContext()
	switch v {
	case msg.ViewAppList:
		m.appList = view.NewAppListModel(m.styles.HelpBar, m.width, m.height)
		return m, view.FetchAppList(ctx, m.runner)
	case msg.ViewAppDetail:
		m.appDetail = view.NewAppDetailModel(param, m.styles.DetailStyles(), m.width, m.height).
			WithStatsInterval(m.statsInterval).
			WithLogOptions(m.logOptions)
		return m, view.FetchAppDetail(ctx, m.runner, param, m.statsInterval, m.logOptions)
	case msg.ViewServiceList:
		m.serviceList = view.NewServiceListModel(m.styles.HelpBar, m.width, m.height)
		return m, view.FetchServiceList(ctx, m.runner)
	case msg.ViewServiceDetail:
		m.serviceDetail = view.NewServiceDetailModel(param, m.styles.DetailStyles(), m.width, m.height).
			WithStatsInterval(m.statsInterval)
		return m, view.FetchServiceDetail(ctx, m.runner, param, m.statsInterval)
	}
	return m.initLeafView(v)
}

// initLeafView creates a simple list view and returns the command that fills it.
func (m Model) initLeafView(v msg.ViewID) (Model, tea.Cmd) {
	ctx := view.CmdContext()
	switch v {
	case msg.ViewVolumeList:
		m.volumeList = view.NewVolumeListModel(m.styles.HelpBar, m.width, m.height)
		return m, view.FetchVolumeList(ctx, m.runner)
	case msg.ViewDomainList:
		m.domainList = view.NewDomainListModel(m.styles.HelpBar, m.width, m.height)
		return m, view.FetchDomainList(ctx, m.runner)
	case msg.ViewPolicyList:
		m.policyList = view.NewPolicyListModel(m.styles.HelpBar, m.width, m.height)
		return m, view.FetchPolicyList(ctx, m.runner)
	case msg.ViewRegionList:
		m.regionList = view.NewRegionListModel(m.styles.HelpBar, m.width, m.height)
		return m, view.FetchRegionList(ctx, m.runner)
	case msg.ViewPlanList:
		m.planList = view.NewPlanListModel(m.styles.HelpBar, m.width, m.height)
		return m, view.FetchPlanList(ctx, m.runner)
	}
	return m, nil
}

func (m Model) updateCurrentView(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentView {
	case msg.ViewMainMenu:
		m.mainMenu, cmd = m.mainMenu.Update(teaMsg)
	case msg.ViewAppList:
		m.appList, cmd = m.appList.Update(teaMsg)
	case msg.ViewAppDetail:
		m.appDetail, cmd = m.appDetail.Update(teaMsg)
	case msg.ViewServiceList:
		m.serviceList, cmd = m.serviceList.Update(teaMsg)
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
		return m, cmd
	}
	switch m.currentView {
	case msg.ViewVolumeList:
		m.volumeList, cmd = m.volumeList.Update(teaMsg)
	case msg.ViewDomainList:
		m.domainList, cmd = m.domainList.Update(teaMsg)
	case msg.ViewPolicyList:
		m.policyList, cmd = m.policyList.Update(teaMsg)
	case msg.ViewRegionList:
		m.regionList, cmd = m.regionList.Update(teaMsg)
	case msg.ViewPlanList:
		m.planList, cmd = m.planList.Update(teaMsg)
	}
	return m, cmd
}

func (m Model) propagateSize() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{Width: m.width, Height: m.height}
	}
}
