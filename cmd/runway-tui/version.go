package main

import (
	"flag"
	"io"
	"runtime/debug"
)

// version is set at link time by release builds:
// -ldflags "-X main.version=v0.1.0".
var version = ""

// resolveVersion prefers the link-time version, then the module version that
// `go install …@vX` records, and falls back to "dev" for local builds.
func resolveVersion(ldflags string, info *debug.BuildInfo) string {
	if ldflags != "" {
		return ldflags
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

type options struct {
	version bool
}

// parseArgs parses the command-line flags (without the program name).
func parseArgs(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("runway-tui", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&o.version, "version", false, "print the version and exit")
	return o, fs.Parse(args)
}
