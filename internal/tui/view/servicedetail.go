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

var serviceTabs = []string{"Info", "Config", "Backups", "DSN"}

// ServiceDetailModel shows service details with tabbed sections.
type ServiceDetailModel struct {
	name      string
	activeTab int
	styles    DetailStyles
	spinner   spinner.Model
	width     int
	height    int

	service *runway.Service
	config  runway.ServiceConfig
	backups string
	dsn     runway.ServiceDSN

	loaded [4]bool
	errors [4]error
}

// NewServiceDetailModel creates a new service detail model.
func NewServiceDetailModel(name string, styles DetailStyles, width, height int) ServiceDetailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return ServiceDetailModel{
		name:    name,
		styles:  styles,
		spinner: s,
		width:   width,
		height:  height,
	}
}

func (m ServiceDetailModel) Update(teaMsg tea.Msg) (ServiceDetailModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case msg.ServiceDetailResultMsg:
		m.loaded[0] = true
		m.service = teaMsg.Service
		m.errors[0] = teaMsg.Err
	case msg.ServiceConfigResultMsg:
		m.loaded[1] = true
		m.config = teaMsg.Config
		m.errors[1] = teaMsg.Err
	case msg.ServiceBackupsResultMsg:
		m.loaded[2] = true
		m.backups = string(teaMsg.Backups)
		m.errors[2] = teaMsg.Err
	case msg.ServiceDSNResultMsg:
		m.loaded[3] = true
		m.dsn = teaMsg.DSN
		m.errors[3] = teaMsg.Err
	case tea.KeyMsg:
		return m.handleKey(teaMsg)
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(teaMsg)
	return m, cmd
}

func (m ServiceDetailModel) handleKey(k tea.KeyMsg) (ServiceDetailModel, tea.Cmd) {
	switch k.String() {
	case "tab":
		m.activeTab = (m.activeTab + 1) % len(serviceTabs)
	case "shift+tab":
		m.activeTab = (m.activeTab - 1 + len(serviceTabs)) % len(serviceTabs)
	}
	return m, nil
}

func (m ServiceDetailModel) View() string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("Service: "+m.name) + "\n\n")
	b.WriteString(m.renderTabs() + "\n\n")
	b.WriteString(m.renderActiveTab())
	b.WriteString("\n" + renderHelp(m.styles.HelpBar, "tab: switch tab", "esc: back"))
	return b.String()
}

func (m ServiceDetailModel) renderTabs() string {
	var tabs []string
	for i, t := range serviceTabs {
		if i == m.activeTab {
			tabs = append(tabs, m.styles.TabActive.Render(t))
		} else {
			tabs = append(tabs, m.styles.TabInactive.Render(t))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m ServiceDetailModel) renderActiveTab() string {
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
		return m.renderConfigTab()
	case 2:
		return m.renderBackupsTab()
	case 3:
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
			row("Volume:", fmt.Sprintf("%s at %s (%s bytes)", v.Name, v.Mountpoint, v.Size))
		}
	}
	return b.String()
}

func (m ServiceDetailModel) renderConfigTab() string {
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

func (m ServiceDetailModel) renderBackupsTab() string {
	if m.backups == "" || m.backups == "[]" || m.backups == "null" {
		return "No backups found."
	}
	return m.backups
}

func (m ServiceDetailModel) renderDSNTab() string {
	if len(m.dsn) == 0 {
		return "No DSN available."
	}
	var b strings.Builder
	for k, v := range m.dsn {
		b.WriteString(m.styles.Label.Render(k+":") + " " + m.styles.Value.Render(v) + "\n")
	}
	return b.String()
}
