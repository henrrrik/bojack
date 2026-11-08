package runway

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const appListJSON = `[
  {
    "app": {
      "created": "2026-04-05T11:08:16Z",
      "id": "my-app",
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "owner": "alice",
      "updated": "2026-04-06T09:00:00Z",
      "type": "app",
      "plan": "standard-1",
      "region": "eu-north-1",
      "features": [
        {"name": "metrics", "value": true},
        {"name": "autoscale", "value": false}
      ],
      "meta": {"role": "owner"}
    },
    "status": {
      "status": "running",
      "deployment_status": "running",
      "build_status": "",
      "integrations": []
    },
    "domains": [
      {
        "app": "my-app",
        "created": "2026-04-05 11:08:16.123456 +0000 GMT",
        "domain": "my-app.runway.app",
        "path": "/",
        "targetPath": "",
        "owner": "alice",
        "creatorName": "alice",
        "updated": "2026-04-05 11:08:16.123456 +0000 GMT",
        "deletable": false,
        "root": {
          "domain": "runway.app",
          "owner": "system",
          "created": "2025-01-01T00:00:00Z",
          "updated": "2025-01-01T00:00:00Z",
          "verified": true,
          "lastcheck": "2026-04-05T00:00:00Z",
          "txtrecord": "",
          "lastcheckmsg": ""
        }
      }
    ]
  }
]`

const appDetailJSON = `{
  "app": {
    "created": "2026-04-05T11:08:16Z",
    "id": "my-app",
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "owner": "alice",
    "updated": "2026-04-06T09:00:00Z",
    "type": "app",
    "plan": "standard-1",
    "region": "eu-north-1",
    "features": [],
    "meta": {"role": "owner"}
  },
  "domains": [],
  "settings": {
    "lockdown": false,
    "routable": true
  }
}`

const processListJSON = `[
  {
    "release": "v42",
    "type": "web",
    "name": "web.1",
    "state": "up",
    "started": "2026-04-05T12:00:00Z",
    "integrations": []
  }
]`

const appStatsJSON = `{
  "cpu": {
    "title": "CPU",
    "unit": "cores",
    "max": 0.5,
    "warnAbove": 0.325,
    "criticalAbove": 0.4,
    "interval": {"tickMilliseconds": 3600000, "startTimestamp": 1794009439173, "endTimestamp": 1794095839173},
    "data": [
      {"ts": 1794009439174, "v": 0.1},
      {"ts": 1794013039174, "v": 0.25}
    ]
  },
  "ram": {
    "title": "Memory",
    "unit": "MiB",
    "max": 128,
    "warnAbove": 83.2,
    "criticalAbove": 102.4,
    "interval": {"tickMilliseconds": 3600000, "startTimestamp": 1794009439173, "endTimestamp": 1794095839173},
    "data": [{"ts": 1794009439174, "v": 12.83203125}]
  },
  "hdd_ephemeral_app": {
    "title": "HDD (ephemeral, app)",
    "unit": "MiB",
    "max": 2048,
    "warnAbove": 1331.2,
    "criticalAbove": 1638.4,
    "interval": {"tickMilliseconds": 3600000, "startTimestamp": 1794009439173, "endTimestamp": 1794095839173},
    "data": []
  },
  "net_in": {
    "title": "Network (in)",
    "unit": "MBit/s",
    "max": 5,
    "warnAbove": 3.25,
    "criticalAbove": 4,
    "interval": {"tickMilliseconds": 3600000, "startTimestamp": 1794009439173, "endTimestamp": 1794095839173},
    "data": [{"ts": 1794009439174, "v": 0.0045}]
  }
}`

const appConfigJSON = `{
  "DATABASE_URL": "postgres://localhost/mydb",
  "PORT": "8080"
}`

const appDomainListJSON = `[
  {
    "app": "my-app",
    "created": "2026-04-05 11:08:16.123456 +0000 GMT",
    "domain": "my-app.runway.app",
    "path": "/",
    "targetPath": "",
    "owner": "alice",
    "creatorName": "alice",
    "updated": "2026-04-05 11:08:16.123456 +0000 GMT",
    "deletable": false,
    "root": {
      "domain": "runway.app",
      "owner": "system",
      "created": "2025-01-01T00:00:00Z",
      "updated": "2025-01-01T00:00:00Z",
      "verified": true,
      "lastcheck": "2026-04-05T00:00:00Z",
      "txtrecord": "",
      "lastcheckmsg": ""
    }
  }
]`

const appShareListJSON = `{
  "data": [
    {"user": "alice", "role": "owner"},
    {"user": "bob", "role": "developer"}
  ]
}`

const serviceListJSON = `[
  {
    "name": "my-db",
    "type": "postgres",
    "plan": "standard-1",
    "has_backup": true,
    "endpoints": null,
    "credentials": null,
    "volumes": [
      {"name": "data", "size": "1073741824", "mountpoint": "/var/lib/postgresql"}
    ]
  }
]`

const serviceShowJSON = `{
  "name": "my-db",
  "type": "postgres",
  "plan": "standard-1",
  "has_backup": true,
  "endpoints": [
    {"URL": "postgres://host:5432/db", "Description": "Primary"}
  ],
  "credentials": {
    "database": "mydb",
    "hostname": "db.internal",
    "password": "secret123",
    "port": "5432",
    "username": "admin"
  },
  "volumes": [
    {"name": "data", "size": "1073741824", "mountpoint": "/var/lib/postgresql"}
  ]
}`

const serviceDSNJSON = `{
  "DATABASE_URL": "postgres://admin:secret123@db.internal:5432/mydb"
}`

const volumeListJSON = `[
  {
    "name": "data",
    "app": "my-db",
    "mountpoint": "/var/lib/postgresql",
    "size": 5120,
    "size_extra": 0,
    "replicas": [],
    "is_replica": false,
    "attachable": false,
    "detachable": false,
    "deletable": false,
    "resizable": true
  }
]`

const domainListJSON = `[
  {
    "domain": "example.com",
    "owner": "alice",
    "created": "2026-01-01T00:00:00Z",
    "updated": "2026-01-01T00:00:00Z",
    "verified": true,
    "lastcheck": "2026-04-05T00:00:00Z",
    "txtrecord": "runway-verify=abc123",
    "lastcheckmsg": "OK"
  }
]`

const policyListJSON = `[
  {"name": "allow-deploy", "app": "my-app", "from": "user:bob"}
]`

const regionListJSON = `[
  {
    "id": "eu-north-1",
    "name": "Stockholm",
    "live_status": true,
    "country_code": "SE",
    "country": "Sweden",
    "continent": "Europe",
    "dns": {
      "a_record": "1.2.3.4",
      "cname_record": "eu-north-1.runway.app",
      "base_domain": "runway.app"
    }
  }
]`

const planListJSON = `[
  {
    "plan": "standard-1",
    "sku": "std1",
    "cpu": 1,
    "ram_mb": 512,
    "hdd_mb": 1024,
    "bandwidth_mbits": 100,
    "price_hour": 0.01,
    "description": "Standard 1x",
    "autoscaling": false,
    "persistency": true,
    "storage_upgrades": true,
    "permutation": "default"
  }
]`

func unmarshalAppListEntry(t *testing.T) AppListEntry {
	t.Helper()
	var entries []AppListEntry
	if err := json.Unmarshal([]byte(appListJSON), &entries); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	return entries[0]
}

func TestUnmarshalAppListApp(t *testing.T) {
	e := unmarshalAppListEntry(t)
	if e.App.ID != "my-app" {
		t.Errorf("App.ID = %q, want %q", e.App.ID, "my-app")
	}
	if e.App.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("App.UUID = %q", e.App.UUID)
	}
	if e.App.Owner != "alice" {
		t.Errorf("App.Owner = %q", e.App.Owner)
	}
	if e.App.Type != "app" {
		t.Errorf("App.Type = %q", e.App.Type)
	}
	if e.App.Plan != "standard-1" {
		t.Errorf("App.Plan = %q", e.App.Plan)
	}
	if e.App.Region != "eu-north-1" {
		t.Errorf("App.Region = %q", e.App.Region)
	}
	if e.App.Meta.Role != "owner" {
		t.Errorf("Meta.Role = %q", e.App.Meta.Role)
	}
}

func TestUnmarshalAppListStatus(t *testing.T) {
	e := unmarshalAppListEntry(t)
	if e.Status.Status != "running" {
		t.Errorf("Status.Status = %q, want %q", e.Status.Status, "running")
	}
	if e.Status.DeploymentStatus != "running" {
		t.Errorf("Status.DeploymentStatus = %q", e.Status.DeploymentStatus)
	}
}

func TestUnmarshalAppListFeatures(t *testing.T) {
	e := unmarshalAppListEntry(t)
	if len(e.App.Features) != 2 {
		t.Fatalf("got %d features, want 2", len(e.App.Features))
	}
	if e.App.Features[0].Name != "metrics" || !e.App.Features[0].Value {
		t.Errorf("feature[0] = %+v", e.App.Features[0])
	}
}

func TestUnmarshalAppListDomains(t *testing.T) {
	e := unmarshalAppListEntry(t)
	if len(e.Domains) != 1 {
		t.Fatalf("got %d domains, want 1", len(e.Domains))
	}
	d := e.Domains[0]
	if d.Domain != "my-app.runway.app" {
		t.Errorf("Domain = %q", d.Domain)
	}
	if d.Deletable {
		t.Error("Deletable should be false")
	}
	if !d.Root.Verified {
		t.Error("Root.Verified should be true")
	}
}

func TestUnmarshalAppDetail(t *testing.T) {
	var detail AppDetail
	if err := json.Unmarshal([]byte(appDetailJSON), &detail); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if detail.App.ID != "my-app" {
		t.Errorf("App.ID = %q", detail.App.ID)
	}
	if detail.Settings.Lockdown {
		t.Error("Settings.Lockdown should be false")
	}
	if !detail.Settings.Routable {
		t.Error("Settings.Routable should be true")
	}
}

func TestUnmarshalProcessList(t *testing.T) {
	var procs []Process
	if err := json.Unmarshal([]byte(processListJSON), &procs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(procs) != 1 {
		t.Fatalf("got %d, want 1", len(procs))
	}
	p := procs[0]
	if p.Release != "v42" {
		t.Errorf("Release = %q", p.Release)
	}
	if p.Type != "web" {
		t.Errorf("Type = %q", p.Type)
	}
	if p.Name != "web.1" {
		t.Errorf("Name = %q", p.Name)
	}
	if p.State != "up" {
		t.Errorf("State = %q", p.State)
	}
}

func TestUnmarshalAppStats(t *testing.T) {
	var stats Stats
	if err := json.Unmarshal([]byte(appStatsJSON), &stats); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	cpu, ok := stats["cpu"]
	if !ok {
		t.Fatal("missing cpu metric")
	}
	meta := cpu
	meta.Data = nil
	want := StatsMetric{Title: "CPU", Unit: "cores", Max: 0.5, WarnAbove: 0.325, CriticalAbove: 0.4}
	if !reflect.DeepEqual(meta, want) {
		t.Errorf("metric = %+v, want %+v", meta, want)
	}
	if len(cpu.Data) != 2 {
		t.Fatalf("got %d points, want 2", len(cpu.Data))
	}
	if cpu.Data[0].Timestamp != 1794009439174 || cpu.Data[0].Value != 0.1 {
		t.Errorf("Data[0] = %+v", cpu.Data[0])
	}
}

func TestStatsMetricLatest(t *testing.T) {
	m := StatsMetric{Data: []StatsPoint{{Value: 1}, {Value: 2}}}
	if v, ok := m.Latest(); !ok || v != 2 {
		t.Errorf("Latest = %v, %v; want 2, true", v, ok)
	}
	if _, ok := (StatsMetric{}).Latest(); ok {
		t.Error("Latest on empty metric should return ok=false")
	}
}

func TestStatsMetricValues(t *testing.T) {
	m := StatsMetric{Data: []StatsPoint{{Value: 1}, {Value: 2}, {Value: 3}}}
	got := m.Values()
	want := []float64{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("Values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Values = %v, want %v", got, want)
		}
	}
}

func TestStatsOrderedKeys(t *testing.T) {
	stats := Stats{
		"net_out": {}, "zzz_custom": {}, "ram": {}, "hdd_ephemeral_worker": {},
		"net_in": {}, "cpu": {}, "hdd_ephemeral_app": {},
	}
	got := stats.OrderedKeys()
	want := []string{"cpu", "ram", "hdd_ephemeral_app", "hdd_ephemeral_worker", "net_in", "net_out", "zzz_custom"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("OrderedKeys = %v, want %v", got, want)
	}
}

func TestUnmarshalAppConfig(t *testing.T) {
	var config AppConfig
	if err := json.Unmarshal([]byte(appConfigJSON), &config); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if config["DATABASE_URL"] != "postgres://localhost/mydb" {
		t.Errorf("DATABASE_URL = %q", config["DATABASE_URL"])
	}
	if config["PORT"] != "8080" {
		t.Errorf("PORT = %q", config["PORT"])
	}
}

func TestUnmarshalAppDomainList(t *testing.T) {
	var domains []Domain
	if err := json.Unmarshal([]byte(appDomainListJSON), &domains); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("got %d, want 1", len(domains))
	}
	if domains[0].Domain != "my-app.runway.app" {
		t.Errorf("Domain = %q", domains[0].Domain)
	}
}

func TestUnmarshalAppShareList(t *testing.T) {
	var shares AppShareList
	if err := json.Unmarshal([]byte(appShareListJSON), &shares); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(shares.Data) != 2 {
		t.Fatalf("got %d, want 2", len(shares.Data))
	}
	if shares.Data[0].User != "alice" || shares.Data[0].Role != "owner" {
		t.Errorf("shares[0] = %+v", shares.Data[0])
	}
	if shares.Data[1].User != "bob" || shares.Data[1].Role != "developer" {
		t.Errorf("shares[1] = %+v", shares.Data[1])
	}
}

func TestUnmarshalServiceList(t *testing.T) {
	var services []Service
	if err := json.Unmarshal([]byte(serviceListJSON), &services); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("got %d, want 1", len(services))
	}
	s := services[0]
	if s.Name != "my-db" {
		t.Errorf("Name = %q", s.Name)
	}
	if s.Type != "postgres" {
		t.Errorf("Type = %q", s.Type)
	}
	if !s.HasBackup {
		t.Error("HasBackup should be true")
	}
	if s.Endpoints != nil {
		t.Errorf("Endpoints should be nil, got %v", s.Endpoints)
	}
	if s.Credentials != nil {
		t.Errorf("Credentials should be nil, got %v", s.Credentials)
	}
	if len(s.Volumes) != 1 {
		t.Fatalf("got %d volumes, want 1", len(s.Volumes))
	}
	if s.Volumes[0].Name != "data" {
		t.Errorf("Volume.Name = %q", s.Volumes[0].Name)
	}
}

func TestUnmarshalServiceShow(t *testing.T) {
	var s Service
	if err := json.Unmarshal([]byte(serviceShowJSON), &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if s.Credentials == nil {
		t.Fatal("Credentials should not be nil")
	}
	if s.Credentials.Username != "admin" {
		t.Errorf("Username = %q", s.Credentials.Username)
	}
	if s.Credentials.Password != "secret123" {
		t.Errorf("Password = %q", s.Credentials.Password)
	}
	if len(s.Endpoints) != 1 {
		t.Fatalf("got %d endpoints, want 1", len(s.Endpoints))
	}
	if s.Endpoints[0].Description != "Primary" {
		t.Errorf("Description = %q", s.Endpoints[0].Description)
	}
}

func TestUnmarshalServiceDSN(t *testing.T) {
	var dsn ServiceDSN
	if err := json.Unmarshal([]byte(serviceDSNJSON), &dsn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dsn["DATABASE_URL"] != "postgres://admin:secret123@db.internal:5432/mydb" {
		t.Errorf("DATABASE_URL = %q", dsn["DATABASE_URL"])
	}
}

func TestUnmarshalVolumeList(t *testing.T) {
	var volumes []Volume
	if err := json.Unmarshal([]byte(volumeListJSON), &volumes); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(volumes) != 1 {
		t.Fatalf("got %d, want 1", len(volumes))
	}
	v := volumes[0]
	if v.Name != "data" {
		t.Errorf("Name = %q", v.Name)
	}
	if v.App != "my-db" {
		t.Errorf("App = %q", v.App)
	}
	if v.SizeMB != 5120 {
		t.Errorf("SizeMB = %d, want 5120", v.SizeMB)
	}
	if !v.Resizable {
		t.Error("Resizable should be true")
	}
	if v.Deletable {
		t.Error("Deletable should be false")
	}
}

func TestUnmarshalDomainList(t *testing.T) {
	var domains []RootDomain
	if err := json.Unmarshal([]byte(domainListJSON), &domains); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("got %d, want 1", len(domains))
	}
	d := domains[0]
	if d.Domain != "example.com" {
		t.Errorf("Domain = %q", d.Domain)
	}
	if !d.Verified {
		t.Error("Verified should be true")
	}
	if d.TXTRecord != "runway-verify=abc123" {
		t.Errorf("TXTRecord = %q", d.TXTRecord)
	}
}

func TestUnmarshalPolicyList(t *testing.T) {
	var policies []Policy
	if err := json.Unmarshal([]byte(policyListJSON), &policies); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(policies) != 1 {
		t.Fatalf("got %d, want 1", len(policies))
	}
	p := policies[0]
	if p.Name != "allow-deploy" {
		t.Errorf("Name = %q", p.Name)
	}
	if p.App != "my-app" {
		t.Errorf("App = %q", p.App)
	}
	if p.From != "user:bob" {
		t.Errorf("From = %q", p.From)
	}
}

func TestUnmarshalRegionList(t *testing.T) {
	var regions []Region
	if err := json.Unmarshal([]byte(regionListJSON), &regions); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(regions) != 1 {
		t.Fatalf("got %d, want 1", len(regions))
	}
	r := regions[0]
	if r.ID != "eu-north-1" {
		t.Errorf("ID = %q", r.ID)
	}
	if r.Name != "Stockholm" {
		t.Errorf("Name = %q", r.Name)
	}
	if !r.LiveStatus {
		t.Error("LiveStatus should be true")
	}
	if r.CountryCode != "SE" {
		t.Errorf("CountryCode = %q", r.CountryCode)
	}
	if r.DNS.ARecord != "1.2.3.4" {
		t.Errorf("DNS.ARecord = %q", r.DNS.ARecord)
	}
}

const appLogsJSON = `[
  {
    "content": "2026/04/05 23:27:57 shutting down server",
    "time": "2026-04-05T23:27:57Z",
    "type": "app"
  },
  {
    "content": "2026/04/05 23:27:56 request handled",
    "time": "2026-04-05T23:27:56Z",
    "type": "app"
  }
]`

func TestUnmarshalLogEntries(t *testing.T) {
	var logs []LogEntry
	if err := json.Unmarshal([]byte(appLogsJSON), &logs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("got %d, want 2", len(logs))
	}
	if logs[0].Content != "2026/04/05 23:27:57 shutting down server" {
		t.Errorf("Content = %q", logs[0].Content)
	}
	if logs[0].Time != "2026-04-05T23:27:57Z" {
		t.Errorf("Time = %q", logs[0].Time)
	}
	if logs[0].Type != "app" {
		t.Errorf("Type = %q", logs[0].Type)
	}
}

func TestUnmarshalPlanList(t *testing.T) {
	var plans []Plan
	if err := json.Unmarshal([]byte(planListJSON), &plans); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("got %d, want 1", len(plans))
	}
	p := plans[0]
	if p.Plan != "standard-1" {
		t.Errorf("Plan = %q", p.Plan)
	}
	if p.CPU != 1 {
		t.Errorf("CPU = %f", p.CPU)
	}
	if p.RAMMB != 512 {
		t.Errorf("RAMMB = %f", p.RAMMB)
	}
	if p.PriceHour != 0.01 {
		t.Errorf("PriceHour = %f", p.PriceHour)
	}
	if !p.Persistency {
		t.Error("Persistency should be true")
	}
}
