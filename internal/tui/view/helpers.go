package view

import (
	"context"
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// CmdContext returns a background context for CLI commands.
// Each command manages its own timeout via the CLI tool.
func CmdContext() context.Context {
	return context.Background()
}

// FetchAppList returns a command that fetches the app list.
func FetchAppList(ctx context.Context, r runway.Runner) tea.Cmd {
	return func() tea.Msg {
		apps, err := r.AppList(ctx)
		return msg.AppListResultMsg{Apps: apps, Err: err}
	}
}

// FetchAppDetail returns a batch command that fetches all app detail data.
func FetchAppDetail(ctx context.Context, r runway.Runner, name string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			d, err := r.AppShow(ctx, name)
			return msg.AppDetailResultMsg{Detail: d, Err: err}
		},
		func() tea.Msg {
			p, err := r.AppProcesses(ctx, name)
			return msg.AppProcessesResultMsg{Processes: p, Err: err}
		},
		func() tea.Msg {
			s, err := r.AppStats(ctx, name)
			return msg.AppStatsResultMsg{Stats: s, Err: err}
		},
		func() tea.Msg {
			c, err := r.AppConfig(ctx, name)
			return msg.AppConfigResultMsg{Config: c, Err: err}
		},
		func() tea.Msg {
			d, err := r.AppDomains(ctx, name)
			return msg.AppDomainsResultMsg{Domains: d, Err: err}
		},
		func() tea.Msg {
			s, err := r.AppShares(ctx, name)
			return msg.AppSharesResultMsg{Shares: s, Err: err}
		},
		func() tea.Msg {
			l, err := r.AppLogs(ctx, name)
			return msg.AppLogsResultMsg{Logs: l, Err: err}
		},
	)
}

// FetchServiceList returns a command that fetches the service list.
func FetchServiceList(ctx context.Context, r runway.Runner) tea.Cmd {
	return func() tea.Msg {
		services, err := r.ServiceList(ctx)
		return msg.ServiceListResultMsg{Services: services, Err: err}
	}
}

// FetchServiceDetail returns a batch command for service detail data.
func FetchServiceDetail(ctx context.Context, r runway.Runner, name string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			s, err := r.ServiceShow(ctx, name)
			return msg.ServiceDetailResultMsg{Service: s, Err: err}
		},
		func() tea.Msg {
			d, err := r.ServiceDSN(ctx, name)
			return msg.ServiceDSNResultMsg{DSN: d, Err: err}
		},
		func() tea.Msg {
			c, err := r.ServiceConfig(ctx, name)
			return msg.ServiceConfigResultMsg{Config: c, Err: err}
		},
		func() tea.Msg {
			b, err := r.ServiceBackups(ctx, name)
			return msg.ServiceBackupsResultMsg{Backups: b, Err: err}
		},
	)
}

func FetchVolumeList(ctx context.Context, r runway.Runner) tea.Cmd {
	return func() tea.Msg {
		v, err := r.VolumeList(ctx)
		return msg.VolumeListResultMsg{Volumes: v, Err: err}
	}
}

func FetchDomainList(ctx context.Context, r runway.Runner) tea.Cmd {
	return func() tea.Msg {
		d, err := r.DomainList(ctx)
		return msg.DomainListResultMsg{Domains: d, Err: err}
	}
}

func FetchPolicyList(ctx context.Context, r runway.Runner) tea.Cmd {
	return func() tea.Msg {
		p, err := r.PolicyList(ctx)
		return msg.PolicyListResultMsg{Policies: p, Err: err}
	}
}

func FetchRegionList(ctx context.Context, r runway.Runner) tea.Cmd {
	return func() tea.Msg {
		r2, err := r.RegionList(ctx)
		return msg.RegionListResultMsg{Regions: r2, Err: err}
	}
}

func FetchPlanList(ctx context.Context, r runway.Runner) tea.Cmd {
	return func() tea.Msg {
		p, err := r.PlanList(ctx)
		return msg.PlanListResultMsg{Plans: p, Err: err}
	}
}

func renderHelp(style lipgloss.Style, items ...string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += " | "
		}
		result += item
	}
	return style.Render(result)
}

// formatMB renders a size given in megabytes, e.g. "5120" -> "5.0 GB".
// Unparseable input is returned unchanged; empty input becomes "-".
func formatMB(s string) string {
	if s == "" {
		return "-"
	}
	mb, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	if mb < 1024 {
		return fmt.Sprintf("%.0f MB", mb)
	}
	return fmt.Sprintf("%.1f GB", mb/1024)
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
