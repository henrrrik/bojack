package view

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/runway"
)

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
