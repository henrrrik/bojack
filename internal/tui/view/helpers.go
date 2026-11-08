package view

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/henrrrik/bojack/internal/msg"
	"github.com/henrrrik/bojack/internal/runway"
)

// CmdContext returns a background context for CLI commands.
// Each command manages its own timeout via the CLI tool.
func CmdContext() context.Context {
	return context.Background()
}

// FetchAppDetail returns a batch command that fetches all app detail data,
// with stats over interval and logs selected by logOpts.
func FetchAppDetail(ctx context.Context, r runway.Runner, name, interval string, logOpts runway.LogOptions) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			d, err := r.AppShow(ctx, name)
			return msg.AppDetailResultMsg{Detail: d, Err: err}
		},
		func() tea.Msg {
			p, err := r.AppProcesses(ctx, name)
			return msg.AppProcessesResultMsg{Processes: p, Err: err}
		},
		FetchAppStats(ctx, r, name, interval),
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
		FetchAppLogs(ctx, r, name, logOpts),
	)
}

// FetchServiceDetail returns a batch command for service detail data,
// with stats over interval.
func FetchServiceDetail(ctx context.Context, r runway.Runner, name, interval string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			s, err := r.ServiceShow(ctx, name)
			return msg.ServiceDetailResultMsg{Service: s, Err: err}
		},
		FetchServiceStats(ctx, r, name, interval),
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

// FetchAppStats returns a command that fetches app stats over interval.
func FetchAppStats(ctx context.Context, r runway.Runner, name, interval string) tea.Cmd {
	return func() tea.Msg {
		s, err := r.AppStats(ctx, name, interval)
		return msg.AppStatsResultMsg{Stats: s, Interval: interval, Err: err}
	}
}

// FetchAppLogs returns a command that fetches app logs with opts.
func FetchAppLogs(ctx context.Context, r runway.Runner, name string, opts runway.LogOptions) tea.Cmd {
	return func() tea.Msg {
		l, err := r.AppLogs(ctx, name, opts)
		return msg.AppLogsResultMsg{Logs: l, Options: opts, Err: err}
	}
}

// FetchServiceStats returns a command that fetches service stats over interval.
func FetchServiceStats(ctx context.Context, r runway.Runner, name, interval string) tea.Cmd {
	return func() tea.Msg {
		s, err := r.ServiceStats(ctx, name, interval)
		return msg.ServiceStatsResultMsg{Stats: s, Interval: interval, Err: err}
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

// formatAppStatus renders e.g. "running (deployment: deploying, build: -)".
func formatAppStatus(s runway.AppStatus) string {
	return fmt.Sprintf("%s (deployment: %s, build: %s)",
		orDash(s.Status), orDash(s.DeploymentStatus), orDash(s.BuildStatus))
}

func formatIntegrations(items []json.RawMessage) string {
	if len(items) == 0 {
		return "none"
	}
	names := make([]string, len(items))
	for i, it := range items {
		names[i] = describeJSON(it)
	}
	return strings.Join(names, ", ")
}

// describeJSON summarises a JSON value of unknown schema: a string as is, an
// object by its "name" or "id", anything else as compact JSON.
func describeJSON(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var obj map[string]any
	if json.Unmarshal(raw, &obj) == nil {
		for _, k := range []string{"name", "id"} {
			if v, ok := obj[k].(string); ok && v != "" {
				return v
			}
		}
	}
	var b bytes.Buffer
	if json.Compact(&b, raw) != nil {
		return string(raw)
	}
	return b.String()
}
