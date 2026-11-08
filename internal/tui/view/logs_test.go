package view

import (
	"testing"
	"time"
)

func TestNextLogSince(t *testing.T) {
	tests := []struct{ in, want time.Duration }{
		{0, time.Hour}, {time.Hour, 6 * time.Hour}, {6 * time.Hour, 24 * time.Hour},
		{24 * time.Hour, 7 * 24 * time.Hour}, {7 * 24 * time.Hour, 0}, {42 * time.Minute, 0},
	}
	for _, tt := range tests {
		if got := nextLogSince(tt.in); got != tt.want {
			t.Errorf("nextLogSince(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestLogSinceLabel(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{0, "all"}, {time.Hour, "1h"}, {24 * time.Hour, "24h"}, {7 * 24 * time.Hour, "7d"},
	}
	for _, tt := range tests {
		if got := logSinceLabel(tt.in); got != tt.want {
			t.Errorf("logSinceLabel(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNextLogLimit(t *testing.T) {
	tests := []struct{ in, want int }{{100, 500}, {500, 1000}, {1000, 100}, {7, 100}}
	for _, tt := range tests {
		if got := nextLogLimit(tt.in); got != tt.want {
			t.Errorf("nextLogLimit(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
