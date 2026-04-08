package runway

import "encoding/json"

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
	App     App      `json:"app"`
	Domains []Domain `json:"domains"`
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

// AppStats is the response from "app stats -o json".
type AppStats struct {
	All    []StatsEntry `json:"all"`
	Latest StatsValues  `json:"latest"`
}

type StatsEntry struct {
	Timestamp int64       `json:"ts"`
	Values    StatsValues `json:"val"`
}

type StatsValues struct {
	RAM StatsRAM `json:"ram"`
	CPU StatsCPU `json:"cpu"`
	HDD StatsHDD `json:"hdd"`
	Net StatsNet `json:"net"`
}

type StatsRAM struct {
	UsedBytes int64 `json:"used_bytes"`
}

type StatsCPU struct {
	Load float64 `json:"load"`
}

type StatsHDD struct {
	UsedBytes int64 `json:"used_bytes"`
}

type StatsNet struct {
	ReceivedBytes    int64 `json:"received_bytes"`
	TransmittedBytes int64 `json:"transmitted_bytes"`
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
	Size       string `json:"size"`
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
	Size       string `json:"size"`
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
	CPU             int     `json:"cpu"`
	RAMMB           int     `json:"ram_mb"`
	HDDMB           int     `json:"hdd_mb"`
	BandwidthMbits  int     `json:"bandwidth_mbits"`
	PriceHour       float64 `json:"price_hour"`
	Description     string  `json:"description"`
	Autoscaling     bool    `json:"autoscaling"`
	Persistency     bool    `json:"persistency"`
	StorageUpgrades bool    `json:"storage_upgrades"`
	Permutation     string  `json:"permutation"`
}
