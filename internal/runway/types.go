package runway

import (
	"encoding/json"
	"sort"
	"strings"
)

// App represents a Runway application.
type App struct {
	Created  string       `json:"created"`
	ID       string       `json:"id"`
	UUID     string       `json:"uuid"`
	Owner    string       `json:"owner"`
	Updated  string       `json:"updated"`
	Type     string       `json:"type"`
	Plan     string       `json:"plan"`
	Region   string       `json:"region"`
	Features []AppFeature `json:"features"`
	Meta     AppMeta      `json:"meta"`
}

type AppFeature struct {
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type AppMeta struct {
	Role string `json:"role"`
}

// AppListEntry is one item from "app ls -o json".
type AppListEntry struct {
	App     App       `json:"app"`
	Status  AppStatus `json:"status"`
	Domains []Domain  `json:"domains"`
}

// AppStatus is the runtime status of an app. Fields are empty for services.
type AppStatus struct {
	Status           string `json:"status"`
	DeploymentStatus string `json:"deployment_status"`
	BuildStatus      string `json:"build_status"`
}

// AppDetail is the response from "app show -o json".
type AppDetail struct {
	App      App         `json:"app"`
	Domains  []Domain    `json:"domains"`
	Settings AppSettings `json:"settings"`
}

type AppSettings struct {
	Lockdown bool `json:"lockdown"`
	Routable bool `json:"routable"`
}

// Domain represents an app domain.
type Domain struct {
	App         string     `json:"app"`
	Created     string     `json:"created"`
	Domain      string     `json:"domain"`
	Path        string     `json:"path"`
	TargetPath  string     `json:"targetPath"`
	Owner       string     `json:"owner"`
	CreatorName string     `json:"creatorName"`
	Updated     string     `json:"updated"`
	Deletable   bool       `json:"deletable"`
	Root        DomainRoot `json:"root"`
}

type DomainRoot struct {
	Domain       string `json:"domain"`
	Owner        string `json:"owner"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
	Verified     bool   `json:"verified"`
	LastCheck    string `json:"lastcheck"`
	TXTRecord    string `json:"txtrecord"`
	LastCheckMsg string `json:"lastcheckmsg"`
}

// Process represents a running app process.
type Process struct {
	Release      string            `json:"release"`
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	State        string            `json:"state"`
	Started      string            `json:"started"`
	Integrations []json.RawMessage `json:"integrations"`
}

// LogEntry represents a single log line.
type LogEntry struct {
	Content string `json:"content"`
	Time    string `json:"time"`
	Type    string `json:"type"`
}

// Stats is the response from "app stats -o json" and "service stats -o json":
// a set of metric series keyed by metric name (e.g. "cpu", "ram",
// "hdd_ephemeral_app", "net_in", "db_tps"). The set of keys varies by app and
// service type, so it is modelled as a map.
type Stats map[string]StatsMetric

// StatsMetric is a single metric time series. Values are already in Unit.
type StatsMetric struct {
	Title         string       `json:"title"`
	Unit          string       `json:"unit"`
	Max           float64      `json:"max"`
	WarnAbove     float64      `json:"warnAbove"`
	CriticalAbove float64      `json:"criticalAbove"`
	Data          []StatsPoint `json:"data"`
}

type StatsPoint struct {
	Timestamp int64   `json:"ts"` // milliseconds since epoch
	Value     float64 `json:"v"`
}

// Latest returns the most recent value, or ok=false if there is no data.
func (m StatsMetric) Latest() (value float64, ok bool) {
	if len(m.Data) == 0 {
		return 0, false
	}
	return m.Data[len(m.Data)-1].Value, true
}

// Values returns the series values in order.
func (m StatsMetric) Values() []float64 {
	values := make([]float64, len(m.Data))
	for i, p := range m.Data {
		values[i] = p.Value
	}
	return values
}

// OrderedKeys returns the metric keys in display order: CPU, RAM, disks,
// network in/out, then anything else alphabetically.
func (s Stats) OrderedKeys() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ri, rj := statsKeyRank(keys[i]), statsKeyRank(keys[j])
		if ri != rj {
			return ri < rj
		}
		return keys[i] < keys[j]
	})
	return keys
}

func statsKeyRank(key string) int {
	switch {
	case key == "cpu":
		return 0
	case key == "ram":
		return 1
	case strings.HasPrefix(key, "hdd"):
		return 2
	case key == "net_in":
		return 3
	case key == "net_out":
		return 4
	default:
		return 5
	}
}

// AppConfig is a map of environment variable key-value pairs.
type AppConfig map[string]string

// AppShareList is the response from "app share ls -o json".
type AppShareList struct {
	Data []AppShare `json:"data"`
}

type AppShare struct {
	User string `json:"user"`
	Role string `json:"role"`
}

// Service represents a Runway managed service.
type Service struct {
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Plan        string              `json:"plan"`
	HasBackup   bool                `json:"has_backup"`
	Endpoints   []ServiceEndpoint   `json:"endpoints"`
	Credentials *ServiceCredentials `json:"credentials"`
	Volumes     []ServiceVolume     `json:"volumes"`
}

type ServiceEndpoint struct {
	URL         string `json:"URL"`
	Description string `json:"Description"`
}

type ServiceCredentials struct {
	Database string `json:"database"`
	Hostname string `json:"hostname"`
	Password string `json:"password"`
	Port     string `json:"port"`
	Username string `json:"username"`
}

type ServiceVolume struct {
	Name       string `json:"name"`
	Size       string `json:"size"` // in MB, sent as a string
	Mountpoint string `json:"mountpoint"`
}

// ServiceDSN is a map of DSN key-value pairs (key varies by service type).
type ServiceDSN map[string]string

// ServiceConfig is a nullable map of config key-value pairs.
type ServiceConfig map[string]string

// Volume represents a Runway volume.
type Volume struct {
	Name       string `json:"name"`
	App        string `json:"app"`
	Mountpoint string `json:"mountpoint"`
	SizeMB     int    `json:"size"`
	Attachable bool   `json:"attachable"`
	Detachable bool   `json:"detachable"`
	Deletable  bool   `json:"deletable"`
	Resizable  bool   `json:"resizable"`
}

// RootDomain represents a verified root domain.
type RootDomain struct {
	Domain       string `json:"domain"`
	Owner        string `json:"owner"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
	Verified     bool   `json:"verified"`
	LastCheck    string `json:"lastcheck"`
	TXTRecord    string `json:"txtrecord"`
	LastCheckMsg string `json:"lastcheckmsg"`
}

// Policy represents an access policy.
type Policy struct {
	Name string `json:"name"`
	App  string `json:"app"`
	From string `json:"from"`
}

// Region represents an available deployment region.
type Region struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	LiveStatus  bool      `json:"live_status"`
	CountryCode string    `json:"country_code"`
	Country     string    `json:"country"`
	Continent   string    `json:"continent"`
	DNS         RegionDNS `json:"dns"`
}

type RegionDNS struct {
	ARecord     string `json:"a_record"`
	CNAMERecord string `json:"cname_record"`
	BaseDomain  string `json:"base_domain"`
}

// Plan represents an available hosting plan.
type Plan struct {
	Plan            string  `json:"plan"`
	SKU             string  `json:"sku"`
	CPU             float64 `json:"cpu"`
	RAMMB           float64 `json:"ram_mb"`
	HDDMB           float64 `json:"hdd_mb"`
	BandwidthMbits  float64 `json:"bandwidth_mbits"`
	PriceHour       float64 `json:"price_hour"`
	Description     string  `json:"description"`
	Autoscaling     bool    `json:"autoscaling"`
	Persistency     bool    `json:"persistency"`
	StorageUpgrades bool    `json:"storage_upgrades"`
	Permutation     string  `json:"permutation"`
}
