package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

var appTabs = []string{"Info", "Processes", "Stats", "Config", "Domains", "Sharing", "Logs"}

const (
	appStatsTab  = 2
	appConfigTab = 3
	appLogsTab   = 6
)

// AppDetailModel shows app details with tabbed sections.
type AppDetailModel struct {
	appName       string
	activeTab     int
	statsInterval string
	restart       restartPrompt
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
		restart:       newRestartPrompt(appName, false),
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
	case msg.RestartResultMsg:
		m.restart = m.restart.apply(teaMsg)
	case tea.KeyMsg:
		return m.handleKey(teaMsg)
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

func (m AppDetailModel) handleKey(k tea.KeyMsg) (AppDetailModel, tea.Cmd) {
	if p, cmd, ok := m.restart.handleKey(k.String()); ok {
		m.restart = p
		return m, cmd
	}
	if m.activeTab == appLogsTab {
		if opts, ok := updateLogOptions(m.logOptions, k.String()); ok {
			m.logOptions = opts
			m.loaded[appLogsTab] = false
			return m, requestLogOptions(opts)
		}
	}
	switch k.String() {
	case "tab":
		m.activeTab = (m.activeTab + 1) % len(appTabs)
	case "shift+tab":
		m.activeTab = (m.activeTab - 1 + len(appTabs)) % len(appTabs)
	case revealKey:
		if m.activeTab == appConfigTab {
			m.revealSecrets = !m.revealSecrets
		}
	case statsIntervalKey:
		if m.activeTab == appStatsTab {
			m.statsInterval = NextStatsInterval(m.statsInterval)
			m.loaded[appStatsTab] = false
			return m, requestStatsInterval(m.statsInterval)
		}
	}
	return m, nil
}

func (m AppDetailModel) View() string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("App: "+m.appName) + "\n\n")
	b.WriteString(m.renderTabs() + "\n\n")
	b.WriteString(m.restart.withStatus(m.renderActiveTab(), m.styles))
	help := []string{"tab: switch tab", "o: open", "R: restart", "esc: back"}
	switch m.activeTab {
	case appStatsTab:
		help = append([]string{"i: interval"}, help...)
	case appConfigTab:
		help = append([]string{revealHelp(m.revealSecrets)}, help...)
	case appLogsTab:
		help = append([]string{"a/b/s: toggle app/build/system", "t: range", "l: limit"}, help...)
	}
	b.WriteString("\n" + renderHelp(m.styles.HelpBar, help...))
	return b.String()
}

func (m AppDetailModel) renderTabs() string {
	var tabs []string
	for i, t := range appTabs {
		if i == m.activeTab {
			tabs = append(tabs, m.styles.TabActive.Render(t))
		} else {
			tabs = append(tabs, m.styles.TabInactive.Render(t))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
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
		return renderStats(m.styles, m.stats, m.width, m.statsInterval)
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
	var b strings.Builder
	for _, p := range m.processes {
		state := m.styles.StatusUp.Render(p.State)
		if p.State != "up" {
			state = m.styles.StatusDown.Render(p.State)
		}
		fmt.Fprintf(&b, "  %s  %s  %s  %s  %s\n", p.Name, p.Type, p.Release, state, p.Started)
	}
	return b.String()
}

func (m AppDetailModel) renderConfigTab() string {
	return renderConfig(m.styles, m.config, m.revealSecrets)
}

func (m AppDetailModel) renderDomainsTab() string {
	if len(m.domains) == 0 {
		return "No domains configured."
	}
	var b strings.Builder
	for _, d := range m.domains {
		fmt.Fprintf(&b, "  %s  path=%s  deletable=%v\n", d.Domain, d.Path, d.Deletable)
	}
	return b.String()
}

func (m AppDetailModel) renderSharingTab() string {
	if m.shares == nil || len(m.shares.Data) == 0 {
		return "No shares configured."
	}
	var b strings.Builder
	for _, s := range m.shares.Data {
		fmt.Fprintf(&b, "  %s  (%s)\n", s.User, s.Role)
	}
	return b.String()
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
