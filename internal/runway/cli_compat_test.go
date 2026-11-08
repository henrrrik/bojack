//go:build compat

package runway

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// TestCLICompat checks the installed, logged-in runway CLI against our types.
// It needs real credentials, so it only runs with the compat build tag:
//
//	go test -tags compat -run TestCLICompat -v ./internal/runway/
//
// It calls every read-only Runner method (not AppOpen, which opens a browser)
// on the first running app and first service, then fails on command errors,
// decode errors, or fields our types expect that the CLI no longer sends.
// New fields the CLI added are only logged. Only field paths are printed,
// never values, so DSNs and credentials stay out of the output.
func TestCLICompat(t *testing.T) {
	if out, err := exec.Command("runway", "version").Output(); err == nil {
		t.Logf("runway CLI %s", strings.TrimSpace(string(out)))
	}
	rec := &recordingExecutor{inner: NewExecCommandExecutor("runway")}
	r := NewCLIRunner(rec)
	ctx := context.Background()

	app, svc := pickTargets(t, r)
	t.Logf("using app %q, service %q", app, svc)

	cases := []struct {
		name string
		run  func() (reflect.Type, error)
	}{
		{"app ls", call(func() ([]AppListEntry, error) { return r.AppList(ctx) })},
		{"app show", call(func() (*AppDetail, error) { return r.AppShow(ctx, app) })},
		{"app ps", call(func() ([]Process, error) { return r.AppProcesses(ctx, app) })},
		{"app stats", call(func() (Stats, error) { return r.AppStats(ctx, app, "24h") })},
		{"app config ls", call(func() (AppConfig, error) { return r.AppConfig(ctx, app) })},
		{"app domain ls", call(func() ([]Domain, error) { return r.AppDomains(ctx, app) })},
		{"app share ls", call(func() (*AppShareList, error) { return r.AppShares(ctx, app) })},
		{"app logs", call(func() ([]LogEntry, error) { return r.AppLogs(ctx, app, LogOptions{Limit: 20}) })},
		{"app logs (filtered)", call(func() ([]LogEntry, error) {
			return r.AppLogs(ctx, app, LogOptions{Limit: 20, Since: 24 * 3600e9, NoBuild: true, NoSystem: true})
		})},
		{"service ls", call(func() ([]Service, error) { return r.ServiceList(ctx) })},
		{"service show", call(func() (*Service, error) { return r.ServiceShow(ctx, svc) })},
		{"service stats", call(func() (Stats, error) { return r.ServiceStats(ctx, svc, "24h") })},
		{"service dsn", call(func() (ServiceDSN, error) { return r.ServiceDSN(ctx, svc) })},
		{"service config ls", call(func() (ServiceConfig, error) { return r.ServiceConfig(ctx, svc) })},
		{"service backup ls", call(func() (json.RawMessage, error) { return r.ServiceBackups(ctx, svc) })},
		{"volume ls", call(func() ([]Volume, error) { return r.VolumeList(ctx) })},
		{"domain ls", call(func() ([]RootDomain, error) { return r.DomainList(ctx) })},
		{"policy ls", call(func() ([]Policy, error) { return r.PolicyList(ctx) })},
		{"region ls", call(func() ([]Region, error) { return r.RegionList(ctx) })},
		{"plan ls", call(func() ([]Plan, error) { return r.PlanList(ctx) })},
		{"template ls", call(func() ([]Template, error) { return r.TemplateList(ctx) })},
		{"template show", call(func() (*Template, error) {
			ts, err := r.TemplateList(ctx)
			if err != nil || len(ts) == 0 {
				return nil, fmt.Errorf("need a template: %v", err)
			}
			return r.TemplateShow(ctx, ts[0].Name)
		})},
		{"volume show", call(func() (*Volume, error) {
			vs, err := r.VolumeList(ctx)
			if err != nil || len(vs) == 0 {
				return nil, fmt.Errorf("need a volume: %v", err)
			}
			return r.VolumeShow(ctx, vs[0].Name)
		})},
		{"registry ls", call(func() ([]Registry, error) { return r.RegistryList(ctx) })},
		{"key ls", call(func() ([]SSHKey, error) { return r.SSHKeyList(ctx) })},
		{"api-key ls", call(func() ([]APIKey, error) { return r.APIKeyList(ctx) })},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			typ, err := c.run()
			if err != nil {
				t.Fatalf("runway %s: %v", strings.Join(rec.lastArgs, " "), err)
			}
			checkSchema(t, typ, rec.lastOut)
		})
	}
}

// call wraps a Runner method so the case reports its result type.
func call[T any](fn func() (T, error)) func() (reflect.Type, error) {
	return func() (reflect.Type, error) {
		_, err := fn()
		return reflect.TypeOf((*T)(nil)).Elem(), err
	}
}

func checkSchema(t *testing.T, typ reflect.Type, out []byte) {
	t.Helper()
	var data any
	if err := json.Unmarshal(out, &data); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if data == nil || reflect.ValueOf(data).Kind() == reflect.Slice && reflect.ValueOf(data).Len() == 0 {
		t.Logf("empty result; nothing to compare")
		return
	}
	d := schemaDiff(typ, data)
	for _, p := range d.Missing {
		t.Errorf("field %s is expected by our types but missing from CLI output", p)
	}
	for _, p := range d.Unknown {
		t.Logf("new field %s (not used by runway-tui)", p)
	}
}

// pickTargets returns the first app with a runtime status (services also
// appear in "app ls" but have none) and the first service.
func pickTargets(t *testing.T, r *CLIRunner) (app, svc string) {
	t.Helper()
	ctx := context.Background()
	apps, err := r.AppList(ctx)
	if err != nil {
		t.Fatalf("app ls: %v (is the runway CLI installed and logged in?)", err)
	}
	for _, a := range apps {
		if a.Status.Status != "" {
			app = a.App.ID
			break
		}
	}
	services, err := r.ServiceList(ctx)
	if err != nil {
		t.Fatalf("service ls: %v", err)
	}
	if len(services) > 0 {
		svc = services[0].Name
	}
	if app == "" || svc == "" {
		t.Skip("need at least one running app and one service to check against")
	}
	return app, svc
}

// recordingExecutor keeps the arguments and output of the last command.
type recordingExecutor struct {
	inner    CommandExecutor
	lastArgs []string
	lastOut  []byte
}

func (e *recordingExecutor) Execute(ctx context.Context, args ...string) ([]byte, error) {
	out, err := e.inner.Execute(ctx, args...)
	e.lastArgs, e.lastOut = args, out
	return out, err
}
