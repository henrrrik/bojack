package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name    string
		ldflags string
		info    *debug.BuildInfo
		want    string
	}{
		{"set at link time", "v0.1.0", &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, "v0.1.0"},
		{"go install @version", "", &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}}, "v0.1.0"},
		{"local build", "", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "dev"},
		{"no build info", "", nil, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.ldflags, tt.info); got != tt.want {
				t.Errorf("resolveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		args        []string
		wantVersion bool
		wantErr     bool
	}{
		{nil, false, false},
		{[]string{"--version"}, true, false},
		{[]string{"-version"}, true, false},
		{[]string{"--bogus"}, false, true},
	}
	for _, tt := range tests {
		got, err := parseArgs(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseArgs(%v) err = %v, wantErr %v", tt.args, err, tt.wantErr)
		}
		if err == nil && got.version != tt.wantVersion {
			t.Errorf("parseArgs(%v).version = %v, want %v", tt.args, got.version, tt.wantVersion)
		}
	}
}
