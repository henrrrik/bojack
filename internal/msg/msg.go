package msg

import (
	"encoding/json"

	"github.com/henrrrik/bojack/internal/runway"
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
	ViewTemplateList
	ViewRegistryList
	ViewSSHKeyList
	ViewAPIKeyList
	ViewTemplateDetail
	ViewVolumeDetail
)

// NavigateMsg requests a view transition.
type NavigateMsg struct {
	View  ViewID
	Param string
}

// NavigateBackMsg requests returning to the previous view.
type NavigateBackMsg struct{}

// Result messages carry data from async fetches.

// Resource names the app, service or item a fetch result belongs to. Views
// drop results for another resource, which can arrive late after the user
// has moved on to a different one.
type Resource struct {
	Name string
}

// ResourceName returns the name the result was fetched for.
func (r Resource) ResourceName() string { return r.Name }

// ResourceResult is a fetch result for one named resource.
type ResourceResult interface {
	ResourceName() string
}

type AppDetailResultMsg struct {
	Resource
	Detail *runway.AppDetail
	Err    error
}

type AppProcessesResultMsg struct {
	Resource
	Processes []runway.Process
	Err       error
}

// AppStatsResultMsg carries app stats for the given interval.
type AppStatsResultMsg struct {
	Resource
	Stats    runway.Stats
	Interval string
	Err      error
}

// ServiceStatsResultMsg carries service stats for the given interval.
type ServiceStatsResultMsg struct {
	Resource
	Stats    runway.Stats
	Interval string
	Err      error
}

// StatsIntervalMsg asks the root model to refetch stats for the current
// detail view over a new interval.
type StatsIntervalMsg struct {
	Interval string
}

type AppConfigResultMsg struct {
	Resource
	Config runway.AppConfig
	Err    error
}

type AppDomainsResultMsg struct {
	Resource
	Domains []runway.Domain
	Err     error
}

type AppSharesResultMsg struct {
	Resource
	Shares *runway.AppShareList
	Err    error
}

// AppLogsResultMsg carries logs fetched with the given options.
type AppLogsResultMsg struct {
	Resource
	Logs    []runway.LogEntry
	Options runway.LogOptions
	Err     error
}

// LogOptionsMsg asks the root model to refetch logs for the current app
// detail view with new options.
type LogOptionsMsg struct {
	Options runway.LogOptions
}

type ServiceDetailResultMsg struct {
	Resource
	Service *runway.Service
	Err     error
}

type ServiceDSNResultMsg struct {
	Resource
	DSN runway.ServiceDSN
	Err error
}

type ServiceConfigResultMsg struct {
	Resource
	Config runway.ServiceConfig
	Err    error
}

type ServiceBackupsResultMsg struct {
	Resource
	Backups json.RawMessage
	Err     error
}

// ListResultMsg carries the result of a list fetch for a generic list view.
type ListResultMsg[T any] struct {
	Items []T
	Err   error
}

// ItemResultMsg carries one item fetched for an ItemModel[T].
type ItemResultMsg[T any] struct {
	Resource
	Item *T
	Err  error
}

// ActionKind identifies an action that changes something in Runway.
type ActionKind int

const (
	ActionRestartApp ActionKind = iota
	ActionRestartService
	ActionCreateBackup
	ActionVerifyDomain
	ActionSetAppConfig
	ActionRemoveAppConfig
	ActionSetServiceConfig
	ActionRemoveServiceConfig
	ActionSetAppPlan
	ActionResizeVolume
)

// ActionRequestMsg asks the root model to run an action on a target (an app,
// service, domain or volume name). Arg is the value the user typed for
// actions that take one, e.g. KEY=value or a plan name.
type ActionRequestMsg struct {
	Kind   ActionKind
	Target string
	Arg    string
}

// ActionResultMsg reports the outcome of an action.
type ActionResultMsg struct {
	Kind   ActionKind
	Target string
	Err    error
}

// FollowTickMsg is a periodic tick for an auto-refreshing ("following") view.
// ID identifies the follow session, so ticks from an old session are ignored.
type FollowTickMsg struct {
	ID int
}
