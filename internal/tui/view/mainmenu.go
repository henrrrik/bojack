package view

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/henrrrik/runway-tui/internal/msg"
)

type menuItem struct {
	title string
	desc  string
	view  msg.ViewID
}

var menuItems = []menuItem{
	{"Apps", "Manage applications", msg.ViewAppList},
	{"Services", "Manage services", msg.ViewServiceList},
	{"Volumes", "Manage volumes", msg.ViewVolumeList},
	{"Domains", "Manage root domains", msg.ViewDomainList},
	{"Policies", "Manage policies", msg.ViewPolicyList},
	{"Regions", "See available regions", msg.ViewRegionList},
	{"Plans", "See available plans", msg.ViewPlanList},
	{"Templates", "See available app templates", msg.ViewTemplateList},
	{"Registries", "See container registries", msg.ViewRegistryList},
	{"SSH Keys", "See deploy SSH keys", msg.ViewSSHKeyList},
	{"API Keys", "See API keys", msg.ViewAPIKeyList},
}

// MainMenuModel is the main menu view.
type MainMenuModel struct {
	cursor int
	width  int
	height int
}

// NewMainMenuModel creates a new main menu model.
func NewMainMenuModel() MainMenuModel {
	return MainMenuModel{}
}

func (m MainMenuModel) Update(teaMsg tea.Msg) (MainMenuModel, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = teaMsg.Width
		m.height = teaMsg.Height
	case tea.KeyPressMsg:
		switch teaMsg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "enter":
			item := menuItems[m.cursor]
			return m, func() tea.Msg {
				return msg.NavigateMsg{View: item.view}
			}
		}
	}
	return m, nil
}

func (m MainMenuModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1).
		PaddingLeft(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		PaddingLeft(2)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		PaddingLeft(4)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingLeft(1).
		MarginTop(1)

	s := titleStyle.Render("Runway TUI") + "\n\n"

	for i, item := range menuItems {
		if i == m.cursor {
			s += selectedStyle.Render("> "+item.title) + " " + descStyle.Render(item.desc) + "\n"
		} else {
			s += normalStyle.Render(item.title) + " " + descStyle.Render(item.desc) + "\n"
		}
	}

	s += "\n" + helpStyle.Render("enter: select | q: quit")
	return s
}
