package view

import (
	"regexp"
	"strings"
)

const secretMask = "********"

// passwordParamRE matches password key/value pairs in libpq-style DSNs
// ("password=x") and query strings ("&password=x").
var passwordParamRE = regexp.MustCompile(`(?i)\b(password|passwd|pwd)=([^\s&;]+)`)

// maskSecrets hides passwords in a connection string: the password part of a
// URL's userinfo and any password= parameter. Everything else is preserved
// byte for byte.
func maskSecrets(s string) string {
	return passwordParamRE.ReplaceAllString(maskURLPassword(s), "${1}="+secretMask)
}

func maskURLPassword(s string) string {
	scheme := strings.Index(s, "://")
	if scheme < 0 {
		return s
	}
	start := scheme + len("://")
	end := len(s)
	if i := strings.IndexAny(s[start:], "/?#"); i >= 0 {
		end = start + i
	}
	authority := s[start:end]
	at := strings.LastIndex(authority, "@")
	if at < 0 {
		return s
	}
	colon := strings.Index(authority[:at], ":")
	if colon < 0 {
		return s
	}
	return s[:start+colon+1] + secretMask + s[start+at:]
}

// revealKey toggles showing secrets on the DSN tab.
const revealKey = "v"

func revealHelp(revealed bool) string {
	if revealed {
		return revealKey + ": hide"
	}
	return revealKey + ": reveal"
}
