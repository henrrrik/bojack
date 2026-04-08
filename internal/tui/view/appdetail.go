package view

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

var appTabs = []string{"Info", "Processes", "Stats", "Config", "Domains", "Sharing"}

// AppDetailModel shows app details with tabbed sections.
type AppDetailModel struct {
	appName   string
	activeTab int
	styles    DetailStyles
	spinner   spinner.Model
	width     int
	height    int

	detail    *runway.AppDetail
	processes []runway.Process
	stats     *runway.AppStats
	config    runway.AppConfig
	domains   []runway.Domain
	shares    *runway.AppShareList

	loaded [6]bool
	errors [6]error
}

// NewAppDetailModel creates a new app detail model.
func NewAppDetailModel(appName string, styles DetailStyles, width, height int) AppDetailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return AppDetailModel{
		appName: appName,
		styles:  styles,
		spinner: s,
		width:   width,
		height:  height,
	}
}

func (m AppDetailModel) Update(teaMsg tea.Msg) (AppDetailModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.AppDetailResultMsg:
		m.loaded[0] = true
		m.detail = teaMsg.Detail
		m.errors[0] = teaMsg.Err
	case msg.AppProcessesResultMsg:
		m.loaded[1] = true
		m.processes = teaMsg.Processes
		m.errors[1] = teaMsg.Err
	case msg.AppStatsResultMsg:
		m.loaded[2] = true
		m.stats = teaMsg.Stats
		m.errors[2] = teaMsg.Err
	case msg.AppConfigResultMsg:
		m.loaded[3] = true
		m.config = teaMsg.Config
		m.errors[3] = teaMsg.Err
	case msg.AppDomainsResultMsg:
		m.loaded[4] = true
		m.domains = teaMsg.Domains
		m.errors[4] = teaMsg.Err
	case msg.AppSharesResultMsg:
		m.loaded[5] = true
		m.shares = teaMsg.Shares
		m.errors[5] = teaMsg.Err
	case tea.KeyMsg:
		return m.handleKey(teaMsg)
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(teaMsg)
	return m, cmd
}

func (m AppDetailModel) handleKey(k tea.KeyMsg) (AppDetailModel, tea.Cmd) {
	switch k.String() {
	case "tab":
		m.activeTab = (m.activeTab + 1) % len(appTabs)
	case "shift+tab":
		m.activeTab = (m.activeTab - 1 + len(appTabs)) % len(appTabs)
	}
	return m, nil
}

func (m AppDetailModel) View() string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("App: "+m.appName) + "\n\n")
	b.WriteString(m.renderTabs() + "\n\n")
	b.WriteString(m.renderActiveTab())
	b.WriteString("\n" + renderHelp(m.styles.HelpBar, "tab: switch tab", "o: open", "esc: back"))
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
	case 2:
		return m.renderStatsTab()
	case 3:
		return m.renderConfigTab()
	case 4:
		return m.renderDomainsTab()
	case 5:
		return m.renderSharingTab()
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
		b.WriteString(fmt.Sprintf("  %s  %s  %s  %s  %s\n", p.Name, p.Type, p.Release, state, p.Started))
	}
	return b.String()
}

func (m AppDetailModel) renderStatsTab() string {
	if m.stats == nil {
		return "No stats available."
	}
	l := m.stats.Latest
	var b strings.Builder
	row := func(label, value string) {
		b.WriteString(m.styles.Label.Render(label) + m.styles.Value.Render(value) + "\n")
	}
	row("CPU Load:", fmt.Sprintf("%.2f", l.CPU.Load))
	row("RAM Used:", formatBytes(l.RAM.UsedBytes))
	row("HDD Used:", formatBytes(l.HDD.UsedBytes))
	row("Net Rx:", formatBytes(l.Net.ReceivedBytes))
	row("Net Tx:", formatBytes(l.Net.TransmittedBytes))
	return b.String()
}

func (m AppDetailModel) renderConfigTab() string {
	if len(m.config) == 0 {
		return "No config vars set."
	}
	keys := make([]string, 0, len(m.config))
	for k := range m.config {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(m.styles.Label.Render(k+":") + " " + m.styles.Value.Render(m.config[k]) + "\n")
	}
	return b.String()
}

func (m AppDetailModel) renderDomainsTab() string {
	if len(m.domains) == 0 {
		return "No domains configured."
	}
	var b strings.Builder
	for _, d := range m.domains {
		b.WriteString(fmt.Sprintf("  %s  path=%s  deletable=%v\n", d.Domain, d.Path, d.Deletable))
	}
	return b.String()
}

func (m AppDetailModel) renderSharingTab() string {
	if m.shares == nil || len(m.shares.Data) == 0 {
		return "No shares configured."
	}
	var b strings.Builder
	for _, s := range m.shares.Data {
		b.WriteString(fmt.Sprintf("  %s  (%s)\n", s.User, s.Role))
	}
	return b.String()
}

func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
