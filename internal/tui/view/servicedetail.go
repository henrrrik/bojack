package view

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

var serviceTabs = []string{"Info", "Stats", "Config", "Backups", "DSN"}

const (
	serviceInfoTab = iota
	serviceStatsTab
	serviceConfigTab
	serviceBackupsTab
	serviceDSNTab
)

// ServiceDetailModel shows service details with tabbed sections.
type ServiceDetailModel struct {
	name          string
	activeTab     int
	statsInterval string
	statsView     statsView
	actions       actionPrompt
	follow        follower
	revealSecrets bool
	styles        DetailStyles
	spinner       spinner.Model
	width         int
	height        int

	service *runway.Service
	stats   runway.Stats
	config  runway.ServiceConfig
	backups json.RawMessage
	dsn     runway.ServiceDSN

	loaded [5]bool
	errors [5]error
}

// NewServiceDetailModel creates a new service detail model.
func NewServiceDetailModel(name string, styles DetailStyles, width, height int) ServiceDetailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return ServiceDetailModel{
		name:          name,
		statsInterval: DefaultStatsInterval,
		actions:       newActionPrompt(restartServiceAction, createBackupAction),
		styles:        styles,
		spinner:       s,
		width:         width,
		height:        height,
	}
}

// WithStatsInterval sets the interval the stats tab is showing.
func (m ServiceDetailModel) WithStatsInterval(interval string) ServiceDetailModel {
	m.statsInterval = interval
	return m
}

// StatsInterval returns the interval the stats tab is showing.
func (m ServiceDetailModel) StatsInterval() string {
	return m.statsInterval
}

func (m ServiceDetailModel) Update(teaMsg tea.Msg) (ServiceDetailModel, tea.Cmd) {
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
func (m ServiceDetailModel) applyResult(teaMsg tea.Msg) ServiceDetailModel {
	switch teaMsg := teaMsg.(type) {
	case msg.ServiceDetailResultMsg:
		m.loaded[serviceInfoTab], m.service, m.errors[serviceInfoTab] = true, teaMsg.Service, teaMsg.Err
	case msg.ServiceStatsResultMsg:
		m = m.applyStats(teaMsg.Stats, teaMsg.Interval, teaMsg.Err)
	case msg.ServiceConfigResultMsg:
		m.loaded[serviceConfigTab], m.config, m.errors[serviceConfigTab] = true, teaMsg.Config, teaMsg.Err
	case msg.ServiceBackupsResultMsg:
		m.loaded[serviceBackupsTab], m.backups, m.errors[serviceBackupsTab] = true, teaMsg.Backups, teaMsg.Err
	case msg.ServiceDSNResultMsg:
		m.loaded[serviceDSNTab], m.dsn, m.errors[serviceDSNTab] = true, teaMsg.DSN, teaMsg.Err
	}
	return m
}

// onFollowTick refetches stats if that tab is showing, and schedules the next tick.
func (m ServiceDetailModel) onFollowTick(t msg.FollowTickMsg) tea.Cmd {
	if !m.follow.accepts(t) {
		return nil
	}
	if m.activeTab == serviceStatsTab {
		return m.follow.onTick(requestStatsInterval(m.statsInterval))
	}
	return m.follow.onTick(nil)
}

// applyStats stores a stats result, ignoring stale results fetched for a
// previous interval.
func (m ServiceDetailModel) applyStats(stats runway.Stats, interval string, err error) ServiceDetailModel {
	if interval != m.statsInterval {
		return m
	}
	m.loaded[serviceStatsTab] = true
	m.stats = stats
	m.errors[serviceStatsTab] = err
	return m
}

func (m ServiceDetailModel) handleKey(k tea.KeyPressMsg) (ServiceDetailModel, tea.Cmd) {
	if p, cmd, ok := m.actions.handleKey(k, m.name, m.tabActions()...); ok {
		m.actions = p
		return m, cmd
	}
	switch k.String() {
	case "tab":
		m.activeTab = (m.activeTab + 1) % len(serviceTabs)
	case "shift+tab":
		m.activeTab = (m.activeTab - 1 + len(serviceTabs)) % len(serviceTabs)
	default:
		return m.tabKey(k.String())
	}
	return m, nil
}

// tabActions are the actions offered on the active tab only.
func (m ServiceDetailModel) tabActions() []actionSpec {
	if m.activeTab == serviceConfigTab {
		return serviceConfigActions
	}
	return nil
}

// ClaimsKey claims every key while an action is asking for input or
// confirmation.
func (m ServiceDetailModel) ClaimsKey(string) bool {
	return m.actions.claimsKeys()
}

// tabKey handles keys specific to the active tab.
func (m ServiceDetailModel) tabKey(k string) (ServiceDetailModel, tea.Cmd) {
	switch m.activeTab {
	case serviceStatsTab:
		return m.statsKey(k)
	case serviceConfigTab, serviceDSNTab:
		if k == revealKey {
			m.revealSecrets = !m.revealSecrets
		}
	}
	return m, nil
}

func (m ServiceDetailModel) statsKey(k string) (ServiceDetailModel, tea.Cmd) {
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
		m.loaded[serviceStatsTab] = false
		cmd = requestStatsInterval(m.statsInterval)
	}
	return m, cmd
}

func (m ServiceDetailModel) View() string {
	var b strings.Builder
	b.WriteString(m.title() + "\n\n")
	b.WriteString(m.renderTabs() + "\n\n")
	if m.follow.on && m.activeTab == serviceStatsTab {
		b.WriteString(m.follow.banner(m.styles) + "\n\n")
	}
	b.WriteString(m.actions.withStatus(m.renderActiveTab(), m.styles))
	help := append([]string{"tab: switch tab"}, m.actions.help(m.tabActions()...)...)
	help = append(help, "esc: back")
	switch m.activeTab {
	case serviceStatsTab:
		help = append([]string{"↑/↓: metric", "i: interval", m.follow.help()}, help...)
	case serviceConfigTab, serviceDSNTab:
		help = append([]string{revealHelp(m.revealSecrets)}, help...)
	}
	b.WriteString("\n" + renderHelp(m.styles.HelpBar, help...))
	return b.String()
}

// bodyHeight is the height left for the active tab.
func (m ServiceDetailModel) bodyHeight() int {
	return detailBodyHeight(m.height, m.follow.on && m.activeTab == serviceStatsTab, m.actions.status(m.styles))
}

func (m ServiceDetailModel) renderTabs() string {
	return renderTabBar(m.styles, serviceTabs, m.activeTab)
}

func (m ServiceDetailModel) renderActiveTab() string {
	if !m.loaded[m.activeTab] {
		return m.spinner.View() + " Loading..."
	}
	if m.errors[m.activeTab] != nil {
		return m.styles.Error.Render("Error: " + m.errors[m.activeTab].Error())
	}
	switch m.activeTab {
	case serviceInfoTab:
		return m.renderInfoTab()
	case serviceStatsTab:
		return m.statsView.render(m.styles, m.stats, m.width, m.bodyHeight(), m.statsInterval)
	case serviceConfigTab:
		return m.renderConfigTab()
	case serviceBackupsTab:
		return m.renderBackupsTab()
	case serviceDSNTab:
		return m.renderDSNTab()
	}
	return ""
}

func (m ServiceDetailModel) renderInfoTab() string {
	if m.service == nil {
		return "No data"
	}
	s := m.service
	var b strings.Builder
	row := func(label, value string) {
		b.WriteString(m.styles.Label.Render(label) + m.styles.Value.Render(value) + "\n")
	}
	row("Name:", s.Name)
	row("Type:", s.Type)
	row("Plan:", s.Plan)
	row("Backup:", fmt.Sprintf("%v", s.HasBackup))
	if s.Credentials != nil {
		row("Host:", s.Credentials.Hostname)
		row("Port:", s.Credentials.Port)
		row("Database:", s.Credentials.Database)
		row("Username:", s.Credentials.Username)
		row("Password:", "********")
	}
	if len(s.Endpoints) > 0 {
		for _, ep := range s.Endpoints {
			row("Endpoint:", ep.URL+" ("+ep.Description+")")
		}
	}
	if len(s.Volumes) > 0 {
		for _, v := range s.Volumes {
			row("Volume:", fmt.Sprintf("%s at %s (%s)", v.Name, v.Mountpoint, formatMB(v.Size)))
		}
	}
	return b.String()
}

func (m ServiceDetailModel) renderConfigTab() string {
	return renderConfig(m.styles, m.config, m.revealSecrets)
}

func (m ServiceDetailModel) renderBackupsTab() string {
	return renderBackups(m.backups)
}

func (m ServiceDetailModel) renderDSNTab() string {
	if len(m.dsn) == 0 {
		return "No DSN available."
	}
	keys := make([]string, 0, len(m.dsn))
	for k := range m.dsn {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		v := m.dsn[k]
		if !m.revealSecrets {
			v = maskSecrets(v)
		}
		b.WriteString(m.styles.Label.Render(k+":") + " " + m.styles.Value.Render(v) + "\n")
	}
	return b.String()
}

func (m ServiceDetailModel) title() string {
	return m.styles.Title.Render("Service: " + m.name)
}

// handleClick switches to a clicked tab, unless an action is waiting for
// input or confirmation.
func (m ServiceDetailModel) handleClick(c tea.MouseClickMsg) ServiceDetailModel {
	if m.actions.claimsKeys() {
		return m
	}
	if i, ok := clickedTab(c, m.styles, m.title(), serviceTabs, m.activeTab); ok {
		m.activeTab = i
	}
	return m
}
