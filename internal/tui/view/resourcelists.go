package view

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// NewAppListModel lists apps; enter opens the app's detail view. Services
// also appear in "app ls" (type "db") and open the service detail view.
func NewAppListModel(helpBar lipgloss.Style, width, height int) ListModel[runway.AppListEntry] {
	return NewListModel("apps", []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Type", Width: 8},
		{Title: "Plan", Width: 14},
		{Title: "Region", Width: 14},
		{Title: "Role", Width: 10},
		{Title: "Status", Width: 10},
	}, appRow, helpBar, width, height).WithSelect(func(a runway.AppListEntry) tea.Msg {
		target := msg.ViewAppDetail
		if a.App.Type == "db" {
			target = msg.ViewServiceDetail
		}
		return msg.NavigateMsg{View: target, Param: a.App.ID}
	})
}

func appRow(a runway.AppListEntry) table.Row {
	return table.Row{a.App.ID, a.App.Type, a.App.Plan, a.App.Region, a.App.Meta.Role, orDash(a.Status.Status)}
}

// NewServiceListModel lists services; enter opens the service detail view.
func NewServiceListModel(helpBar lipgloss.Style, width, height int) ListModel[runway.Service] {
	return NewListModel("services", []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Type", Width: 12},
		{Title: "Plan", Width: 14},
		{Title: "Backup", Width: 8},
	}, serviceRow, helpBar, width, height).WithSelect(func(s runway.Service) tea.Msg {
		return msg.NavigateMsg{View: msg.ViewServiceDetail, Param: s.Name}
	})
}

func serviceRow(s runway.Service) table.Row {
	return table.Row{s.Name, s.Type, s.Plan, fmt.Sprintf("%v", s.HasBackup)}
}

// NewVolumeListModel lists volumes.
func NewVolumeListModel(helpBar lipgloss.Style, width, height int) ListModel[runway.Volume] {
	return NewListModel("volumes", []table.Column{
		{Title: "Name", Width: 20},
		{Title: "App", Width: 20},
		{Title: "Mountpoint", Width: 28},
		{Title: "Size", Width: 14},
	}, volumeRow, helpBar, width, height)
}

func volumeRow(v runway.Volume) table.Row {
	return table.Row{v.Name, v.App, v.Mountpoint, formatMB(strconv.Itoa(v.SizeMB))}
}

// NewDomainListModel lists root domains.
func NewDomainListModel(helpBar lipgloss.Style, width, height int) ListModel[runway.RootDomain] {
	return NewListModel("domains", []table.Column{
		{Title: "Domain", Width: 30},
		{Title: "Owner", Width: 20},
		{Title: "Verified", Width: 10},
	}, domainRow, helpBar, width, height)
}

func domainRow(d runway.RootDomain) table.Row {
	return table.Row{d.Domain, d.Owner, fmt.Sprintf("%v", d.Verified)}
}

// NewPolicyListModel lists access policies.
func NewPolicyListModel(helpBar lipgloss.Style, width, height int) ListModel[runway.Policy] {
	return NewListModel("policies", []table.Column{
		{Title: "Name", Width: 24},
		{Title: "App", Width: 20},
		{Title: "From", Width: 24},
	}, policyRow, helpBar, width, height)
}

func policyRow(p runway.Policy) table.Row {
	return table.Row{p.Name, p.App, p.From}
}

// NewRegionListModel lists regions.
func NewRegionListModel(helpBar lipgloss.Style, width, height int) ListModel[runway.Region] {
	return NewListModel("regions", []table.Column{
		{Title: "ID", Width: 16},
		{Title: "Name", Width: 16},
		{Title: "Country", Width: 16},
		{Title: "Continent", Width: 14},
		{Title: "Live", Width: 6},
	}, regionRow, helpBar, width, height)
}

func regionRow(r runway.Region) table.Row {
	return table.Row{r.ID, r.Name, r.Country, r.Continent, fmt.Sprintf("%v", r.LiveStatus)}
}

// NewPlanListModel lists plans.
func NewPlanListModel(helpBar lipgloss.Style, width, height int) ListModel[runway.Plan] {
	return NewListModel("plans", []table.Column{
		{Title: "Plan", Width: 14},
		{Title: "CPU", Width: 6},
		{Title: "RAM MB", Width: 8},
		{Title: "HDD MB", Width: 8},
		{Title: "BW Mbps", Width: 8},
		{Title: "Price/hr", Width: 10},
		{Title: "Description", Width: 20},
	}, planRow, helpBar, width, height)
}

func planRow(p runway.Plan) table.Row {
	return table.Row{
		p.Plan,
		fmt.Sprintf("%g", p.CPU),
		fmt.Sprintf("%g", p.RAMMB),
		fmt.Sprintf("%g", p.HDDMB),
		fmt.Sprintf("%g", p.BandwidthMbits),
		fmt.Sprintf("%.4f", p.PriceHour),
		p.Description,
	}
}
