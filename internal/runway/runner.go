package runway

import (
	"context"
	"encoding/json"
	"os/exec"
)

// CommandExecutor runs a CLI command and returns its stdout.
type CommandExecutor interface {
	Execute(ctx context.Context, args ...string) ([]byte, error)
}

// Runner provides typed access to Runway CLI commands.
type Runner interface {
	AppList(ctx context.Context) ([]AppListEntry, error)
	AppShow(ctx context.Context, appName string) (*AppDetail, error)
	AppProcesses(ctx context.Context, appName string) ([]Process, error)
	AppStats(ctx context.Context, appName, interval string) (Stats, error)
	AppConfig(ctx context.Context, appName string) (AppConfig, error)
	AppDomains(ctx context.Context, appName string) ([]Domain, error)
	AppShares(ctx context.Context, appName string) (*AppShareList, error)
	AppLogs(ctx context.Context, appName string) ([]LogEntry, error)
	AppOpen(ctx context.Context, appName string) error
	ServiceList(ctx context.Context) ([]Service, error)
	ServiceShow(ctx context.Context, name string) (*Service, error)
	ServiceDSN(ctx context.Context, name string) (ServiceDSN, error)
	ServiceStats(ctx context.Context, name, interval string) (Stats, error)
	ServiceConfig(ctx context.Context, name string) (ServiceConfig, error)
	ServiceBackups(ctx context.Context, name string) (json.RawMessage, error)
	VolumeList(ctx context.Context) ([]Volume, error)
	DomainList(ctx context.Context) ([]RootDomain, error)
	PolicyList(ctx context.Context) ([]Policy, error)
	RegionList(ctx context.Context) ([]Region, error)
	PlanList(ctx context.Context) ([]Plan, error)
}

// ExecCommandExecutor implements CommandExecutor using os/exec.
type ExecCommandExecutor struct {
	binary string
}

// NewExecCommandExecutor creates an executor that calls the given binary.
func NewExecCommandExecutor(binary string) *ExecCommandExecutor {
	return &ExecCommandExecutor{binary: binary}
}

func (e *ExecCommandExecutor) Execute(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, e.binary, args...).Output()
}

// CLIRunner implements Runner by shelling out to the runway CLI.
type CLIRunner struct {
	exec CommandExecutor
}

// NewCLIRunner creates a Runner backed by the given CommandExecutor.
func NewCLIRunner(exec CommandExecutor) *CLIRunner {
	return &CLIRunner{exec: exec}
}

func (r *CLIRunner) AppList(ctx context.Context) ([]AppListEntry, error) {
	return runJSON[[]AppListEntry](ctx, r.exec, "app", "ls", "-o", "json")
}

func (r *CLIRunner) AppShow(ctx context.Context, appName string) (*AppDetail, error) {
	return runJSONPtr[AppDetail](ctx, r.exec, "app", "show", "-a", appName, "-o", "json")
}

func (r *CLIRunner) AppProcesses(ctx context.Context, appName string) ([]Process, error) {
	return runJSON[[]Process](ctx, r.exec, "app", "ps", "-a", appName, "-o", "json")
}

// AppStats fetches app metrics over interval (a Go duration such as "24h").
func (r *CLIRunner) AppStats(ctx context.Context, appName, interval string) (Stats, error) {
	return runJSON[Stats](ctx, r.exec, "app", "stats", "-a", appName, "-i", interval, "-o", "json")
}

func (r *CLIRunner) AppConfig(ctx context.Context, appName string) (AppConfig, error) {
	return runJSON[AppConfig](ctx, r.exec, "app", "config", "ls", "-a", appName, "-o", "json")
}

func (r *CLIRunner) AppDomains(ctx context.Context, appName string) ([]Domain, error) {
	return runJSON[[]Domain](ctx, r.exec, "app", "domain", "ls", "-a", appName, "-o", "json")
}

func (r *CLIRunner) AppShares(ctx context.Context, appName string) (*AppShareList, error) {
	return runJSONPtr[AppShareList](ctx, r.exec, "app", "share", "ls", "-a", appName, "-o", "json")
}

func (r *CLIRunner) AppLogs(ctx context.Context, appName string) ([]LogEntry, error) {
	return runJSON[[]LogEntry](ctx, r.exec, "app", "logs", "-a", appName, "--limit", "100", "-o", "json")
}

func (r *CLIRunner) AppOpen(ctx context.Context, appName string) error {
	_, err := r.exec.Execute(ctx, "app", "open", "-a", appName)
	return err
}

func (r *CLIRunner) ServiceList(ctx context.Context) ([]Service, error) {
	return runJSON[[]Service](ctx, r.exec, "service", "ls", "-o", "json")
}

func (r *CLIRunner) ServiceShow(ctx context.Context, name string) (*Service, error) {
	return runJSONPtr[Service](ctx, r.exec, "service", "show", name, "-o", "json")
}

func (r *CLIRunner) ServiceDSN(ctx context.Context, name string) (ServiceDSN, error) {
	return runJSON[ServiceDSN](ctx, r.exec, "service", "dsn", name, "-o", "json")
}

// ServiceStats fetches service metrics over interval (a Go duration such as "24h").
func (r *CLIRunner) ServiceStats(ctx context.Context, name, interval string) (Stats, error) {
	return runJSON[Stats](ctx, r.exec, "service", "stats", name, "-i", interval, "-o", "json")
}

func (r *CLIRunner) ServiceConfig(ctx context.Context, name string) (ServiceConfig, error) {
	return runJSON[ServiceConfig](ctx, r.exec, "service", "config", "ls", name, "-o", "json")
}

func (r *CLIRunner) ServiceBackups(ctx context.Context, name string) (json.RawMessage, error) {
	out, err := r.exec.Execute(ctx, "service", "backup", "ls", name, "-o", "json")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}

func (r *CLIRunner) VolumeList(ctx context.Context) ([]Volume, error) {
	return runJSON[[]Volume](ctx, r.exec, "volume", "ls", "-o", "json")
}

func (r *CLIRunner) DomainList(ctx context.Context) ([]RootDomain, error) {
	return runJSON[[]RootDomain](ctx, r.exec, "domain", "ls", "-o", "json")
}

func (r *CLIRunner) PolicyList(ctx context.Context) ([]Policy, error) {
	return runJSON[[]Policy](ctx, r.exec, "policy", "ls", "-o", "json")
}

func (r *CLIRunner) RegionList(ctx context.Context) ([]Region, error) {
	return runJSON[[]Region](ctx, r.exec, "region", "ls", "-o", "json")
}

func (r *CLIRunner) PlanList(ctx context.Context) ([]Plan, error) {
	return runJSON[[]Plan](ctx, r.exec, "plan", "ls", "-o", "json")
}

func runJSON[T any](ctx context.Context, exec CommandExecutor, args ...string) (T, error) {
	var result T
	out, err := exec.Execute(ctx, args...)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(out, &result)
	return result, err
}

func runJSONPtr[T any](ctx context.Context, exec CommandExecutor, args ...string) (*T, error) {
	out, err := exec.Execute(ctx, args...)
	if err != nil {
		return nil, err
	}
	var result T
	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
