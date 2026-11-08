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

var appTabs = []string{"Info", "Processes", "Stats", "Config", "Domains", "Sharing", "Logs"}

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
	logs      []runway.LogEntry

	loaded [7]bool
	errors [7]error
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
	case msg.AppLogsResultMsg:
		m.loaded[6] = true
		m.logs = teaMsg.Logs
		m.errors[6] = teaMsg.Err
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

func (m AppDetailModel) renderStatsTab() string {
	if m.stats == nil {
		return "No stats available."
	}
	l := m.stats.Latest
	graphWidth := m.width - 4
	if graphWidth > 60 {
		graphWidth = 60
	}
	var b strings.Builder
	row := func(label, value string, series []float64) {
		b.WriteString(m.styles.Label.Render(label) + m.styles.Value.Render(value) + "\n")
		if graph := sparkline(series, graphWidth); graph != "" {
			b.WriteString("  " + m.styles.Sparkline.Render(graph) + "\n")
		}
		b.WriteString("\n")
	}
	all := m.stats.All
	row("CPU Load:", fmt.Sprintf("%.2f", l.CPU.Load),
		extractSeries(all, func(v runway.StatsValues) float64 { return v.CPU.Load }))
	row("RAM Used:", formatBytes(l.RAM.UsedBytes),
		extractSeries(all, func(v runway.StatsValues) float64 { return v.RAM.UsedBytes }))
	row("HDD Used:", formatBytes(l.HDD.UsedBytes),
		extractSeries(all, func(v runway.StatsValues) float64 { return v.HDD.UsedBytes }))
	row("Net Rx:", formatBytes(l.Net.ReceivedBytes),
		extractSeries(all, func(v runway.StatsValues) float64 { return v.Net.ReceivedBytes }))
	row("Net Tx:", formatBytes(l.Net.TransmittedBytes),
		extractSeries(all, func(v runway.StatsValues) float64 { return v.Net.TransmittedBytes }))
	return b.String()
}

func extractSeries(entries []runway.StatsEntry, fn func(runway.StatsValues) float64) []float64 {
	values := make([]float64, len(entries))
	for i, e := range entries {
		values[i] = fn(e.Values)
	}
	return values
}

var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func sparkline(values []float64, width int) string {
	if len(values) == 0 {
		return ""
	}
	if len(values) > width {
		values = values[len(values)-width:]
	}
	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	var sb strings.Builder
	for _, v := range values {
		idx := 3
		if max > min {
			idx = int((v - min) / (max - min) * 7)
			if idx > 7 {
				idx = 7
			}
		}
		sb.WriteRune(sparkBlocks[idx])
	}
	return sb.String()
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
	if len(m.logs) == 0 {
		return "No log entries."
	}
	var b strings.Builder
	for _, entry := range m.logs {
		fmt.Fprintf(&b, "  %s  [%s]  %s\n", entry.Time, entry.Type, entry.Content)
	}
	return b.String()
}

func formatBytes(b float64) string {
	const (
		kb = 1024.0
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", b/gb)
	case b >= mb:
		return fmt.Sprintf("%.1f MB", b/mb)
	case b >= kb:
		return fmt.Sprintf("%.1f KB", b/kb)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}
