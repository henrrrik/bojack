package runway

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type FakeExecutor struct {
	responses map[string][]byte
	lastArgs  []string
}

func (f *FakeExecutor) Execute(_ context.Context, args ...string) ([]byte, error) {
	f.lastArgs = args
	key := strings.Join(args, " ")
	resp, ok := f.responses[key]
	if !ok {
		return nil, fmt.Errorf("unexpected command: %s", key)
	}
	return resp, nil
}

func newFake(responses map[string][]byte) *FakeExecutor {
	return &FakeExecutor{responses: responses}
}

func TestAppList(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app ls -o json": []byte(appListJSON),
	})
	r := NewCLIRunner(fake)
	entries, err := r.AppList(context.Background())
	if err != nil {
		t.Fatalf("AppList: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].App.ID != "my-app" {
		t.Errorf("App.ID = %q", entries[0].App.ID)
	}
}

func TestAppShow(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app show -a my-app -o json": []byte(appDetailJSON),
	})
	r := NewCLIRunner(fake)
	detail, err := r.AppShow(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("AppShow: %v", err)
	}
	if detail.App.ID != "my-app" {
		t.Errorf("App.ID = %q", detail.App.ID)
	}
	if !detail.Settings.Routable {
		t.Error("Settings.Routable should be true")
	}
}

func TestAppProcesses(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app ps -a my-app -o json": []byte(processListJSON),
	})
	r := NewCLIRunner(fake)
	procs, err := r.AppProcesses(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("AppProcesses: %v", err)
	}
	if len(procs) != 1 {
		t.Fatalf("got %d, want 1", len(procs))
	}
	if procs[0].State != "up" {
		t.Errorf("State = %q", procs[0].State)
	}
}

func TestAppStats(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app stats -a my-app -i 168h -o json": []byte(appStatsJSON),
	})
	r := NewCLIRunner(fake)
	stats, err := r.AppStats(context.Background(), "my-app", "168h")
	if err != nil {
		t.Fatalf("AppStats: %v", err)
	}
	if v, ok := stats["cpu"].Latest(); !ok || v != 0.25 {
		t.Errorf("cpu latest = %v, %v; want 0.25, true", v, ok)
	}
}

func TestServiceStats(t *testing.T) {
	fake := newFake(map[string][]byte{
		"service stats my-db -i 1h -o json": []byte(appStatsJSON),
	})
	r := NewCLIRunner(fake)
	stats, err := r.ServiceStats(context.Background(), "my-db", "1h")
	if err != nil {
		t.Fatalf("ServiceStats: %v", err)
	}
	if v, ok := stats["cpu"].Latest(); !ok || v != 0.25 {
		t.Errorf("cpu latest = %v, %v; want 0.25, true", v, ok)
	}
}

func TestAppConfig(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app config ls -a my-app -o json": []byte(appConfigJSON),
	})
	r := NewCLIRunner(fake)
	config, err := r.AppConfig(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("AppConfig: %v", err)
	}
	if config["PORT"] != "8080" {
		t.Errorf("PORT = %q", config["PORT"])
	}
}

func TestAppDomains(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app domain ls -a my-app -o json": []byte(appDomainListJSON),
	})
	r := NewCLIRunner(fake)
	domains, err := r.AppDomains(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("AppDomains: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("got %d, want 1", len(domains))
	}
}

func TestAppShares(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app share ls -a my-app -o json": []byte(appShareListJSON),
	})
	r := NewCLIRunner(fake)
	shares, err := r.AppShares(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("AppShares: %v", err)
	}
	if len(shares.Data) != 2 {
		t.Fatalf("got %d, want 2", len(shares.Data))
	}
}

func TestServiceList(t *testing.T) {
	fake := newFake(map[string][]byte{
		"service ls -o json": []byte(serviceListJSON),
	})
	r := NewCLIRunner(fake)
	services, err := r.ServiceList(context.Background())
	if err != nil {
		t.Fatalf("ServiceList: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("got %d, want 1", len(services))
	}
	if services[0].Name != "my-db" {
		t.Errorf("Name = %q", services[0].Name)
	}
}

func TestServiceShow(t *testing.T) {
	fake := newFake(map[string][]byte{
		"service show my-db -o json": []byte(serviceShowJSON),
	})
	r := NewCLIRunner(fake)
	s, err := r.ServiceShow(context.Background(), "my-db")
	if err != nil {
		t.Fatalf("ServiceShow: %v", err)
	}
	if s.Credentials == nil {
		t.Fatal("Credentials should not be nil")
	}
	if s.Credentials.Username != "admin" {
		t.Errorf("Username = %q", s.Credentials.Username)
	}
}

func TestServiceDSN(t *testing.T) {
	fake := newFake(map[string][]byte{
		"service dsn my-db -o json": []byte(serviceDSNJSON),
	})
	r := NewCLIRunner(fake)
	dsn, err := r.ServiceDSN(context.Background(), "my-db")
	if err != nil {
		t.Fatalf("ServiceDSN: %v", err)
	}
	if dsn["DATABASE_URL"] == "" {
		t.Error("DATABASE_URL should not be empty")
	}
}

func TestServiceConfig(t *testing.T) {
	fake := newFake(map[string][]byte{
		"service config ls my-db -o json": []byte(`{"max_connections": "100"}`),
	})
	r := NewCLIRunner(fake)
	config, err := r.ServiceConfig(context.Background(), "my-db")
	if err != nil {
		t.Fatalf("ServiceConfig: %v", err)
	}
	if config["max_connections"] != "100" {
		t.Errorf("max_connections = %q", config["max_connections"])
	}
}

func TestServiceConfigNull(t *testing.T) {
	fake := newFake(map[string][]byte{
		"service config ls my-db -o json": []byte(`null`),
	})
	r := NewCLIRunner(fake)
	config, err := r.ServiceConfig(context.Background(), "my-db")
	if err != nil {
		t.Fatalf("ServiceConfig: %v", err)
	}
	if config != nil {
		t.Errorf("expected nil, got %v", config)
	}
}

func TestServiceBackups(t *testing.T) {
	fake := newFake(map[string][]byte{
		"service backup ls my-db -o json": []byte(`[]`),
	})
	r := NewCLIRunner(fake)
	raw, err := r.ServiceBackups(context.Background(), "my-db")
	if err != nil {
		t.Fatalf("ServiceBackups: %v", err)
	}
	if string(raw) != "[]" {
		t.Errorf("got %q", string(raw))
	}
}

func TestVolumeList(t *testing.T) {
	fake := newFake(map[string][]byte{
		"volume ls -o json": []byte(volumeListJSON),
	})
	r := NewCLIRunner(fake)
	volumes, err := r.VolumeList(context.Background())
	if err != nil {
		t.Fatalf("VolumeList: %v", err)
	}
	if len(volumes) != 1 {
		t.Fatalf("got %d, want 1", len(volumes))
	}
}

func TestDomainList(t *testing.T) {
	fake := newFake(map[string][]byte{
		"domain ls -o json": []byte(domainListJSON),
	})
	r := NewCLIRunner(fake)
	domains, err := r.DomainList(context.Background())
	if err != nil {
		t.Fatalf("DomainList: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("got %d, want 1", len(domains))
	}
}

func TestPolicyList(t *testing.T) {
	fake := newFake(map[string][]byte{
		"policy ls -o json": []byte(policyListJSON),
	})
	r := NewCLIRunner(fake)
	policies, err := r.PolicyList(context.Background())
	if err != nil {
		t.Fatalf("PolicyList: %v", err)
	}
	if len(policies) != 1 {
		t.Fatalf("got %d, want 1", len(policies))
	}
}

func TestRegionList(t *testing.T) {
	fake := newFake(map[string][]byte{
		"region ls -o json": []byte(regionListJSON),
	})
	r := NewCLIRunner(fake)
	regions, err := r.RegionList(context.Background())
	if err != nil {
		t.Fatalf("RegionList: %v", err)
	}
	if len(regions) != 1 {
		t.Fatalf("got %d, want 1", len(regions))
	}
}

func TestPlanList(t *testing.T) {
	fake := newFake(map[string][]byte{
		"plan ls -o json": []byte(planListJSON),
	})
	r := NewCLIRunner(fake)
	plans, err := r.PlanList(context.Background())
	if err != nil {
		t.Fatalf("PlanList: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("got %d, want 1", len(plans))
	}
}

func TestAppLogs(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app logs -a my-app --limit 100 -o json": []byte(appLogsJSON),
	})
	r := NewCLIRunner(fake)
	logs, err := r.AppLogs(context.Background(), "my-app", LogOptions{})
	if err != nil {
		t.Fatalf("AppLogs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("got %d, want 2", len(logs))
	}
	if logs[0].Content != "2026/04/05 23:27:57 shutting down server" {
		t.Errorf("Content = %q", logs[0].Content)
	}
}

func TestAppLogsWithOptions(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app logs -a my-app --no-build --no-system --start 2026-11-07T22:30:00 --limit 500 -o json": []byte(appLogsJSON),
	})
	now := time.Date(2026, 11, 8, 0, 30, 0, 0, time.UTC)
	r := NewCLIRunner(fake).WithClock(func() time.Time { return now })
	_, err := r.AppLogs(context.Background(), "my-app", LogOptions{
		Since: 2 * time.Hour, Limit: 500, NoBuild: true, NoRuntime: true, NoSystem: true,
	})
	if err != nil {
		t.Fatalf("AppLogs: %v", err)
	}
}

// The CLI's --no-runtime also drops system logs, so the runner filters
// runtime ("app") entries itself instead of passing the flag.
func TestAppLogsNoRuntimeFiltersClientSide(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app logs -a my-app --limit 100 -o json": []byte(`[
			{"content": "runtime line", "time": "2026-11-01T00:00:00Z", "type": "app"},
			{"content": "probe failed", "time": "2026-11-01T00:00:01Z", "type": "system"},
			{"content": "building", "time": "2026-11-01T00:00:02Z", "type": "build"}
		]`),
	})
	logs, err := NewCLIRunner(fake).AppLogs(context.Background(), "my-app", LogOptions{NoRuntime: true})
	if err != nil {
		t.Fatalf("AppLogs: %v", err)
	}
	if len(logs) != 2 || logs[0].Type != "system" || logs[1].Type != "build" {
		t.Errorf("logs = %+v, want only system and build entries", logs)
	}
}

func listRunner() *CLIRunner {
	return NewCLIRunner(newFake(map[string][]byte{
		"template ls -o json": []byte(templateListJSON),
		"registry ls -o json": []byte(registryListJSON),
		"key ls -o json":      []byte(keyListJSON),
		"api-key ls -o json":  []byte(apiKeyListJSON),
	}))
}

func TestTemplateList(t *testing.T) {
	if got, err := listRunner().TemplateList(context.Background()); err != nil || len(got) != 1 || got[0].Name != "valkey" {
		t.Errorf("TemplateList = %+v, %v", got, err)
	}
}

func TestRegistryList(t *testing.T) {
	if got, err := listRunner().RegistryList(context.Background()); err != nil || len(got) != 1 || got[0].Organization != "org" {
		t.Errorf("RegistryList = %+v, %v", got, err)
	}
}

func TestSSHKeyList(t *testing.T) {
	if got, err := listRunner().SSHKeyList(context.Background()); err != nil || len(got) != 1 || got[0].ID != "laptop" {
		t.Errorf("SSHKeyList = %+v, %v", got, err)
	}
}

func TestAPIKeyList(t *testing.T) {
	if got, err := listRunner().APIKeyList(context.Background()); err != nil || len(got) != 2 || got[1].Status != "REVOKED" {
		t.Errorf("APIKeyList = %+v, %v", got, err)
	}
}

func TestAppRestart(t *testing.T) {
	fake := newFake(map[string][]byte{"app restart -a my-app -y": nil})
	if err := NewCLIRunner(fake).AppRestart(context.Background(), "my-app"); err != nil {
		t.Fatalf("AppRestart: %v", err)
	}
}

func TestServiceRestart(t *testing.T) {
	fake := newFake(map[string][]byte{"service restart my-db -y": nil})
	if err := NewCLIRunner(fake).ServiceRestart(context.Background(), "my-db"); err != nil {
		t.Fatalf("ServiceRestart: %v", err)
	}
}

func TestServiceBackupCreate(t *testing.T) {
	fake := newFake(map[string][]byte{"service backup create my-db -y": nil})
	if err := NewCLIRunner(fake).ServiceBackupCreate(context.Background(), "my-db"); err != nil {
		t.Fatalf("ServiceBackupCreate: %v", err)
	}
}

func TestDomainVerify(t *testing.T) {
	fake := newFake(map[string][]byte{"domain verify example.com -y": nil})
	if err := NewCLIRunner(fake).DomainVerify(context.Background(), "example.com"); err != nil {
		t.Fatalf("DomainVerify: %v", err)
	}
}

func TestExecErrorIncludesStderr(t *testing.T) {
	e := NewExecCommandExecutor("sh")
	_, err := e.Execute(context.Background(), "-c", "echo 'app not found' >&2; exit 1")
	if err == nil || !strings.Contains(err.Error(), "app not found") {
		t.Errorf("error = %v, want it to include stderr", err)
	}
}

// Real failure output from runway CLI 1.49.0.
const (
	boxedErrorStdout = "ERROR running runway app restart -a nope -y: \n" +
		"╭──────\n" +
		"│app does not exist, or permissions are missing\n" +
		"╰──────\n"
	infoLogStderr   = `level=info msg="non-interactive: Restart \"nope\" with the latest configuration? ==> Yes"` + "\n"
	jsonErrorStdout = `{
  "action": "running runway app show -a nope -o json",
  "error": "You do not have permission to perform this action.",
  "status": "error"
}`
)

func TestCLIErrorMessage(t *testing.T) {
	tests := []struct{ name, stdout, stderr, want string }{
		{"json", jsonErrorStdout, "", "You do not have permission to perform this action."},
		{"boxed", boxedErrorStdout, infoLogStderr, "app does not exist, or permissions are missing"},
		{"plain stderr", "", "boom\n", "boom"},
		{"nothing useful", "", infoLogStderr, ""},
	}
	for _, tt := range tests {
		if got := cliErrorMessage([]byte(tt.stdout), []byte(tt.stderr)); got != tt.want {
			t.Errorf("%s: cliErrorMessage = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestExecErrorUsesCLIMessage(t *testing.T) {
	script := "printf '%s' '" + jsonErrorStdout + "'; exit 1"
	_, err := NewExecCommandExecutor("sh").Execute(context.Background(), "-c", script)
	if err == nil || err.Error() != "You do not have permission to perform this action." {
		t.Fatalf("err = %v", err)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Error("error should still wrap the *exec.ExitError")
	}
}

func TestExecSuccessReturnsStdout(t *testing.T) {
	out, err := NewExecCommandExecutor("sh").Execute(context.Background(), "-c", "echo hi; echo noise >&2")
	if err != nil || strings.TrimSpace(string(out)) != "hi" {
		t.Errorf("out = %q, err = %v", out, err)
	}
}

func TestExecutorError(t *testing.T) {
	fake := &FakeExecutor{responses: map[string][]byte{}}
	r := NewCLIRunner(fake)
	_, err := r.AppList(context.Background())
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestAppOpenArgs(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app open -a my-app": {},
	})
	r := NewCLIRunner(fake)
	err := r.AppOpen(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("AppOpen: %v", err)
	}
	expected := "app open -a my-app"
	got := strings.Join(fake.lastArgs, " ")
	if got != expected {
		t.Errorf("args = %q, want %q", got, expected)
	}
}

func TestTemplateShow(t *testing.T) {
	fake := newFake(map[string][]byte{"template show valkey -o json": []byte(
		`{"name":"valkey","latest":"0.2.0","versions":["0.1.1","0.2.0"],"displayName":"Valkey","minPlan":"free-launch"}`)})
	got, err := NewCLIRunner(fake).TemplateShow(context.Background(), "valkey")
	if err != nil || got.Name != "valkey" || len(got.Versions) != 2 || got.MinPlan != "free-launch" {
		t.Errorf("TemplateShow = %+v, %v", got, err)
	}
}

func TestVolumeShow(t *testing.T) {
	fake := newFake(map[string][]byte{"volume show data-1 -o json": []byte(
		`{"name":"data-1","app":"db","mountpoint":"/data","size":5120,"size_extra":0,"replicas":[],"resizable":true}`)})
	got, err := NewCLIRunner(fake).VolumeShow(context.Background(), "data-1")
	if err != nil || got.Name != "data-1" || got.SizeMB != 5120 || !got.Resizable {
		t.Errorf("VolumeShow = %+v, %v", got, err)
	}
}

func TestInputActions(t *testing.T) {
	fake := newFake(map[string][]byte{
		"app config set -a terry FOO=bar -y":  nil,
		"app config rm -a terry FOO -y":       nil,
		"service config set db FOO=bar -y":    nil,
		"service config rm db FOO -y":         nil,
		"app plan set -a terry -p starter -y": nil,
		"volume set data-1 6144 -y":           nil,
	})
	r, ctx := NewCLIRunner(fake), context.Background()
	for name, err := range map[string]error{
		"AppConfigSet":     r.AppConfigSet(ctx, "terry", "FOO=bar"),
		"AppConfigRm":      r.AppConfigRm(ctx, "terry", "FOO"),
		"ServiceConfigSet": r.ServiceConfigSet(ctx, "db", "FOO=bar"),
		"ServiceConfigRm":  r.ServiceConfigRm(ctx, "db", "FOO"),
		"AppPlanSet":       r.AppPlanSet(ctx, "terry", "starter"),
		"VolumeSet":        r.VolumeSet(ctx, "data-1", 6144),
	} {
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
