package msg

import (
	"encoding/json"

	"github.com/henrrrik/runway-tui/internal/runway"
)

// ViewID identifies which view is active.
type ViewID int

const (
	ViewMainMenu ViewID = iota
	ViewAppList
	ViewAppDetail
	ViewServiceList
	ViewServiceDetail
	ViewVolumeList
	ViewDomainList
	ViewPolicyList
	ViewRegionList
	ViewPlanList
)

// NavigateMsg requests a view transition.
type NavigateMsg struct {
	View  ViewID
	Param string
}

// NavigateBackMsg requests returning to the previous view.
type NavigateBackMsg struct{}

// Result messages carry data from async fetches.

type AppListResultMsg struct {
	Apps []runway.AppListEntry
	Err  error
}

type AppDetailResultMsg struct {
	Detail *runway.AppDetail
	Err    error
}

type AppProcessesResultMsg struct {
	Processes []runway.Process
	Err       error
}

type AppStatsResultMsg struct {
	Stats *runway.AppStats
	Err   error
}

type AppConfigResultMsg struct {
	Config runway.AppConfig
	Err    error
}

type AppDomainsResultMsg struct {
	Domains []runway.Domain
	Err     error
}

type AppSharesResultMsg struct {
	Shares *runway.AppShareList
	Err    error
}

type AppLogsResultMsg struct {
	Logs []runway.LogEntry
	Err  error
}

type ServiceListResultMsg struct {
	Services []runway.Service
	Err      error
}

type ServiceDetailResultMsg struct {
	Service *runway.Service
	Err     error
}

type ServiceDSNResultMsg struct {
	DSN runway.ServiceDSN
	Err error
}

type ServiceConfigResultMsg struct {
	Config runway.ServiceConfig
	Err    error
}

type ServiceBackupsResultMsg struct {
	Backups json.RawMessage
	Err     error
}

type VolumeListResultMsg struct {
	Volumes []runway.Volume
	Err     error
}

type DomainListResultMsg struct {
	Domains []runway.RootDomain
	Err     error
}

type PolicyListResultMsg struct {
	Policies []runway.Policy
	Err      error
}

type RegionListResultMsg struct {
	Regions []runway.Region
	Err     error
}

type PlanListResultMsg struct {
	Plans []runway.Plan
	Err   error
}
