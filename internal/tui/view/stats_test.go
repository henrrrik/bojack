package view

import (
	"regexp"
	"testing"
)

// intervalLineRE matches the stats header showing the default interval.
var intervalLineRE = regexp.MustCompile(`Interval:\s+24h`)

func TestNextStatsInterval(t *testing.T) {
	tests := []struct{ in, want string }{
		{"1h", "6h"}, {"6h", "24h"}, {"24h", "168h"}, {"168h", "720h"}, {"720h", "1h"}, {"bogus", "1h"},
	}
	for _, tt := range tests {
		if got := NextStatsInterval(tt.in); got != tt.want {
			t.Errorf("NextStatsInterval(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStatsIntervalLabel(t *testing.T) {
	tests := []struct{ in, want string }{
		{"1h", "1h"}, {"24h", "24h"}, {"168h", "7d"}, {"720h", "30d"}, {"bogus", "bogus"},
	}
	for _, tt := range tests {
		if got := statsIntervalLabel(tt.in); got != tt.want {
			t.Errorf("statsIntervalLabel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
