package view

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// NewAppListModel lists apps; enter opens the app's detail view. Services
// also appear in "app ls" (type "db") and open the service detail view.
func NewAppListModel(styles DetailStyles, width, height int) ListModel[runway.AppListEntry] {
	return NewListModel("apps", []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Type", Width: 8},
		{Title: "Plan", Width: 14},
		{Title: "Region", Width: 14},
		{Title: "Role", Width: 10},
		{Title: "Status", Width: 10},
	}, appRow, styles, width, height).WithSelect(func(a runway.AppListEntry) tea.Msg {
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
func NewServiceListModel(styles DetailStyles, width, height int) ListModel[runway.Service] {
	return NewListModel("services", []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Type", Width: 12},
		{Title: "Plan", Width: 14},
		{Title: "Backup", Width: 8},
	}, serviceRow, styles, width, height).WithSelect(func(s runway.Service) tea.Msg {
		return msg.NavigateMsg{View: msg.ViewServiceDetail, Param: s.Name}
	})
}

func serviceRow(s runway.Service) table.Row {
	return table.Row{s.Name, s.Type, s.Plan, fmt.Sprintf("%v", s.HasBackup)}
}

// NewVolumeListModel lists volumes.
func NewVolumeListModel(styles DetailStyles, width, height int) ListModel[runway.Volume] {
	return NewListModel("volumes", []table.Column{
		{Title: "Name", Width: 30},
		{Title: "App", Width: 24},
		{Title: "Mountpoint", Width: 28},
		{Title: "Size", Width: 10},
		{Title: "Extra", Width: 10},
		{Title: "Replicas", Width: 8},
	}, volumeRow, styles, width, height)
}

func volumeRow(v runway.Volume) table.Row {
	extra := "-"
	if v.SizeExtraMB > 0 {
		extra = formatMB(strconv.Itoa(v.SizeExtraMB))
	}
	replicas := "-"
	switch {
	case v.IsReplica:
		replicas = "replica"
	case len(v.Replicas) > 0:
		replicas = strconv.Itoa(len(v.Replicas))
	}
	return table.Row{v.Name, v.App, v.Mountpoint, formatMB(strconv.Itoa(v.SizeMB)), extra, replicas}
}

// NewDomainListModel lists root domains.
func NewDomainListModel(styles DetailStyles, width, height int) ListModel[runway.RootDomain] {
	return NewListModel("domains", []table.Column{
		{Title: "Domain", Width: 30},
		{Title: "Owner", Width: 20},
		{Title: "Verified", Width: 10},
		{Title: "Last Check", Width: 10},
		{Title: "Message", Width: 30},
	}, domainRow, styles, width, height).
		WithActions(func(d runway.RootDomain) string { return d.Domain }, verifyDomainAction)
}

func domainRow(d runway.RootDomain) table.Row {
	return table.Row{d.Domain, d.Owner, fmt.Sprintf("%v", d.Verified), formatDate(d.LastCheck), orDash(d.LastCheckMsg)}
}

// NewPolicyListModel lists access policies.
func NewPolicyListModel(styles DetailStyles, width, height int) ListModel[runway.Policy] {
	return NewListModel("policies", []table.Column{
		{Title: "Name", Width: 24},
		{Title: "App", Width: 20},
		{Title: "From", Width: 24},
	}, policyRow, styles, width, height)
}

func policyRow(p runway.Policy) table.Row {
	return table.Row{p.Name, p.App, p.From}
}

// NewRegionListModel lists regions.
func NewRegionListModel(styles DetailStyles, width, height int) ListModel[runway.Region] {
	return NewListModel("regions", []table.Column{
		{Title: "ID", Width: 16},
		{Title: "Name", Width: 16},
		{Title: "Country", Width: 16},
		{Title: "Continent", Width: 14},
		{Title: "Live", Width: 6},
	}, regionRow, styles, width, height)
}

func regionRow(r runway.Region) table.Row {
	return table.Row{r.ID, r.Name, r.Country, r.Continent, fmt.Sprintf("%v", r.LiveStatus)}
}

// NewPlanListModel lists plans.
func NewPlanListModel(styles DetailStyles, width, height int) ListModel[runway.Plan] {
	return NewListModel("plans", []table.Column{
		{Title: "Plan", Width: 14},
		{Title: "CPU", Width: 6},
		{Title: "RAM MB", Width: 8},
		{Title: "HDD MB", Width: 8},
		{Title: "BW Mbps", Width: 8},
		{Title: "Price/hr", Width: 10},
		{Title: "Description", Width: 20},
	}, planRow, styles, width, height)
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
