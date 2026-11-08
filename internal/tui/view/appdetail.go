package view

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

var appTabs = []string{"Info", "Processes", "Stats", "Config", "Domains", "Sharing", "Logs"}

const (
	appInfoTab   = 0
	appStatsTab  = 2
	appConfigTab = 3
	appLogsTab   = 6
)

// AppDetailModel shows app details with tabbed sections.
type AppDetailModel struct {
	appName       string
	activeTab     int
	statsInterval string
	statsView     statsView
	actions       actionPrompt
	follow        follower
	logOptions    runway.LogOptions
	revealSecrets bool
	styles        DetailStyles
	spinner       spinner.Model
	width         int
	height        int

	detail    *runway.AppDetail
	processes []runway.Process
	stats     runway.Stats
	config    runway.AppConfig
	domains   []runway.Domain
	shares    *runway.AppShareList
	logs      []runway.LogEntry

	loaded [7]bool
	errors [7]error
}

// NewAppDetailModel creates a new app detail model.
func NewAppDetailModel(appName string, styles DetailStyles, width, height int) AppDetailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return AppDetailModel{
		appName:       appName,
		statsInterval: DefaultStatsInterval,
		actions:       newActionPrompt(restartAppAction),
		logOptions:    DefaultLogOptions,
		styles:        styles,
		spinner:       s,
		width:         width,
		height:        height,
	}
}

// WithStatsInterval sets the interval the stats tab is showing.
func (m AppDetailModel) WithStatsInterval(interval string) AppDetailModel {
	m.statsInterval = interval
	return m
}

// StatsInterval returns the interval the stats tab is showing.
func (m AppDetailModel) StatsInterval() string {
	return m.statsInterval
}

// WithLogOptions sets the options the logs tab is showing.
func (m AppDetailModel) WithLogOptions(opts runway.LogOptions) AppDetailModel {
	m.logOptions = opts
	return m
}

// LogOptions returns the options the logs tab is showing.
func (m AppDetailModel) LogOptions() runway.LogOptions {
	return m.logOptions
}

func (m AppDetailModel) Update(teaMsg tea.Msg) (AppDetailModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.ActionResultMsg:
		m.actions = m.actions.apply(teaMsg)
	case tea.PasteMsg:
		m.actions = m.actions.paste(teaMsg)
	case msg.FollowTickMsg:
		return m, m.onFollowTick(teaMsg)
	case tea.KeyPressMsg:
		return m.handleKey(teaMsg)
	case tea.MouseClickMsg:
		return m.handleClick(teaMsg), nil
	default:
		m = m.applyResult(teaMsg)
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(teaMsg)
	return m, cmd
}

// applyResult stores the data from a fetch result message, if it is one.
func (m AppDetailModel) applyResult(teaMsg tea.Msg) AppDetailModel {
	switch teaMsg := teaMsg.(type) {
	case msg.AppDetailResultMsg:
		m.loaded[0], m.detail, m.errors[0] = true, teaMsg.Detail, teaMsg.Err
	case msg.AppProcessesResultMsg:
		m.loaded[1], m.processes, m.errors[1] = true, teaMsg.Processes, teaMsg.Err
	case msg.AppStatsResultMsg:
		m = m.applyStats(teaMsg.Stats, teaMsg.Interval, teaMsg.Err)
	case msg.AppConfigResultMsg:
		m.loaded[appConfigTab], m.config, m.errors[appConfigTab] = true, teaMsg.Config, teaMsg.Err
	case msg.AppDomainsResultMsg:
		m.loaded[4], m.domains, m.errors[4] = true, teaMsg.Domains, teaMsg.Err
	case msg.AppSharesResultMsg:
		m.loaded[5], m.shares, m.errors[5] = true, teaMsg.Shares, teaMsg.Err
	case msg.AppLogsResultMsg:
		m = m.applyLogs(teaMsg)
	}
	return m
}

// applyStats stores a stats result, ignoring stale results fetched for a
// previous interval.
func (m AppDetailModel) applyStats(stats runway.Stats, interval string, err error) AppDetailModel {
	if interval != m.statsInterval {
		return m
	}
	m.loaded[appStatsTab] = true
	m.stats = stats
	m.errors[appStatsTab] = err
	return m
}

// isLiveTab reports whether the active tab can auto-refresh.
func (m AppDetailModel) isLiveTab() bool {
	return m.activeTab == appStatsTab || m.activeTab == appLogsTab
}

// onFollowTick refetches the active live tab, keeping its current data on
// screen until the new data arrives, and schedules the next tick.
func (m AppDetailModel) onFollowTick(t msg.FollowTickMsg) tea.Cmd {
	if !m.follow.accepts(t) {
		return nil
	}
	switch m.activeTab {
	case appStatsTab:
		return m.follow.onTick(requestStatsInterval(m.statsInterval))
	case appLogsTab:
		return m.follow.onTick(requestLogOptions(m.logOptions))
	}
	return m.follow.onTick(nil)
}

// applyLogs stores a logs result, ignoring stale results fetched with
// previous options.
func (m AppDetailModel) applyLogs(r msg.AppLogsResultMsg) AppDetailModel {
	if r.Options != m.logOptions {
		return m
	}
	m.loaded[appLogsTab] = true
	m.logs = r.Logs
	m.errors[appLogsTab] = r.Err
	return m
}

func (m AppDetailModel) handleKey(k tea.KeyPressMsg) (AppDetailModel, tea.Cmd) {
	if p, cmd, ok := m.actions.handleKey(k, m.appName, m.tabActions()...); ok {
		m.actions = p
		return m, cmd
	}
	switch k.String() {
	case "tab":
		m.activeTab = (m.activeTab + 1) % len(appTabs)
	case "shift+tab":
		m.activeTab = (m.activeTab - 1 + len(appTabs)) % len(appTabs)
	default:
		return m.tabKey(k.String())
	}
	return m, nil
}

// tabActions are the actions offered on the active tab only.
func (m AppDetailModel) tabActions() []actionSpec {
	switch m.activeTab {
	case appInfoTab:
		return []actionSpec{appPlanAction}
	case appConfigTab:
		return appConfigActions
	}
	return nil
}

// ClaimsKey claims every key while an action is asking for input or
// confirmation.
func (m AppDetailModel) ClaimsKey(string) bool {
	return m.actions.claimsKeys()
}

// tabKey handles keys specific to the active tab.
func (m AppDetailModel) tabKey(k string) (AppDetailModel, tea.Cmd) {
	switch m.activeTab {
	case appStatsTab:
		return m.statsKey(k)
	case appLogsTab:
		return m.logsKey(k)
	case appConfigTab:
		if k == revealKey {
			m.revealSecrets = !m.revealSecrets
		}
	}
	return m, nil
}

func (m AppDetailModel) statsKey(k string) (AppDetailModel, tea.Cmd) {
	if s, ok := m.statsView.key(k, len(m.stats)); ok {
		m.statsView = s
		return m, nil
	}
	var cmd tea.Cmd
	switch k {
	case followKey:
		m.follow, cmd = m.follow.toggle()
	case statsIntervalKey:
		m.statsInterval = NextStatsInterval(m.statsInterval)
		m.loaded[appStatsTab] = false
		cmd = requestStatsInterval(m.statsInterval)
	}
	return m, cmd
}

func (m AppDetailModel) logsKey(k string) (AppDetailModel, tea.Cmd) {
	if k == followKey {
		var cmd tea.Cmd
		m.follow, cmd = m.follow.toggle()
		return m, cmd
	}
	if opts, ok := updateLogOptions(m.logOptions, k); ok {
		m.logOptions = opts
		m.loaded[appLogsTab] = false
		return m, requestLogOptions(opts)
	}
	return m, nil
}

func (m AppDetailModel) View() string {
	var b strings.Builder
	b.WriteString(m.title() + "\n\n")
	b.WriteString(m.renderTabs() + "\n\n")
	if m.follow.on && m.isLiveTab() {
		b.WriteString(m.follow.banner(m.styles) + "\n\n")
	}
	b.WriteString(m.actions.withStatus(m.renderActiveTab(), m.styles))
	help := append([]string{"tab: switch tab", "o: open"}, m.actions.help(m.tabActions()...)...)
	help = append(help, "esc: back")
	if m.isLiveTab() {
		help = append([]string{m.follow.help()}, help...)
	}
	switch m.activeTab {
	case appStatsTab:
		help = append([]string{"↑/↓: metric", "i: interval"}, help...)
	case appConfigTab:
		help = append([]string{revealHelp(m.revealSecrets)}, help...)
	case appLogsTab:
		help = append([]string{"a/b/s: toggle app/build/system", "t: range", "l: limit"}, help...)
	}
	b.WriteString("\n" + renderHelp(m.styles.HelpBar, help...))
	return b.String()
}

// bodyHeight is the height left for the active tab.
func (m AppDetailModel) bodyHeight() int {
	return detailBodyHeight(m.height, m.follow.on && m.isLiveTab(), m.actions.status(m.styles))
}

func (m AppDetailModel) renderTabs() string {
	return renderTabBar(m.styles, appTabs, m.activeTab)
}

func (m AppDetailModel) renderActiveTab() string {
	if !m.loaded[m.activeTab] {
		return m.spinner.View() + " Loading..."
	}
	if m.errors[m.activeTab] != nil {
		return m.styles.Error.Render("Error: " + m.errors[m.activeTab].Error())
	}
	switch m.activeTab {
	case 0:
		return m.renderInfoTab()
	case 1:
		return m.renderProcessesTab()
	case appStatsTab:
		return m.statsView.render(m.styles, m.stats, m.width, m.bodyHeight(), m.statsInterval)
	case appConfigTab:
		return m.renderConfigTab()
	case 4:
		return m.renderDomainsTab()
	case 5:
		return m.renderSharingTab()
	case 6:
		return m.renderLogsTab()
	}
	return ""
}

func (m AppDetailModel) renderInfoTab() string {
	if m.detail == nil {
		return "No data"
	}
	a := m.detail.App
	var b strings.Builder
	row := func(label, value string) {
		b.WriteString(m.styles.Label.Render(label) + m.styles.Value.Render(value) + "\n")
	}
	row("ID:", a.ID)
	row("UUID:", a.UUID)
	row("Type:", a.Type)
	row("Owner:", a.Owner)
	row("Plan:", a.Plan)
	row("Region:", a.Region)
	row("Role:", a.Meta.Role)
	row("Source:", orDash(a.Source))
	row("Status:", formatAppStatus(m.detail.Status))
	row("Integrations:", formatIntegrations(m.detail.Status.Integrations))
	row("Created:", a.Created)
	row("Updated:", a.Updated)
	row("Lockdown:", fmt.Sprintf("%v", m.detail.Settings.Lockdown))
	row("Routable:", fmt.Sprintf("%v", m.detail.Settings.Routable))
	if len(a.Features) > 0 {
		var feats []string
		for _, f := range a.Features {
			feats = append(feats, fmt.Sprintf("%s=%v", f.Name, f.Value))
		}
		row("Features:", strings.Join(feats, ", "))
	}
	return b.String()
}

func (m AppDetailModel) renderProcessesTab() string {
	if len(m.processes) == 0 {
		return "No processes running."
	}
	rows := make([][]string, len(m.processes))
	for i, p := range m.processes {
		state := m.styles.StatusUp.Render(p.State)
		if p.State != "up" {
			state = m.styles.StatusDown.Render(p.State)
		}
		rows[i] = []string{p.Name, p.Type, p.Release, state, p.Started}
	}
	return textTable([]string{"Name", "Type", "Release", "State", "Started"}, rows)
}

func (m AppDetailModel) renderConfigTab() string {
	return renderConfig(m.styles, m.config, m.revealSecrets)
}

func (m AppDetailModel) renderDomainsTab() string {
	if len(m.domains) == 0 {
		return "No domains configured."
	}
	rows := make([][]string, len(m.domains))
	for i, d := range m.domains {
		// Runway's default domain (e.g. *.pqapp.dev) has no root domain record.
		root, verified := "-", "-"
		if d.Root.Domain != "" {
			root, verified = d.Root.Domain, yesNo(d.Root.Verified)
		}
		rows[i] = []string{d.Domain, orDash(d.Path), orDash(d.TargetPath), root, verified, yesNo(d.Deletable)}
	}
	return textTable([]string{"Domain", "Path", "Target", "Root", "Verified", "Removable"}, rows)
}

func (m AppDetailModel) renderSharingTab() string {
	if m.shares == nil || len(m.shares.Data) == 0 {
		return "No shares configured."
	}
	rows := make([][]string, len(m.shares.Data))
	for i, sh := range m.shares.Data {
		rows[i] = []string{sh.User, sh.Role}
	}
	return textTable([]string{"User", "Role"}, rows)
}

func (m AppDetailModel) renderLogsTab() string {
	var b strings.Builder
	b.WriteString(m.styles.Value.Render(logOptionsSummary(m.logOptions)) + "\n\n")
	if len(m.logs) == 0 {
		b.WriteString("No log entries.\n")
		return b.String()
	}
	for _, entry := range m.logs {
		fmt.Fprintf(&b, "  %s  [%s]  %s\n", entry.Time, entry.Type, entry.Content)
	}
	return b.String()
}

func (m AppDetailModel) title() string {
	return m.styles.Title.Render("App: " + m.appName)
}

// handleClick switches to a clicked tab, unless an action is waiting for
// input or confirmation.
func (m AppDetailModel) handleClick(c tea.MouseClickMsg) AppDetailModel {
	if m.actions.claimsKeys() {
		return m
	}
	if i, ok := clickedTab(c, m.styles, m.title(), appTabs, m.activeTab); ok {
		m.activeTab = i
	}
	return m
}
