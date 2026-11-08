package view

import (
	"regexp"
	"sort"
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

// secretKeyWords are whole words that mark a config var as secret. Keys are
// split into words on non-alphanumerics.
var secretKeyWords = map[string]bool{"PASS": true, "PWD": true, "PRIVATE": true, "SALT": true, "CREDENTIAL": true, "CREDENTIALS": true}

// secretKeyParts mark a word as secret wherever they appear in it, catching
// compounds like CLIENTSECRET or GHTOKEN.
var secretKeyParts = []string{"SECRET", "TOKEN", "PASSWORD", "PASSWD"}

var nonAlnumRE = regexp.MustCompile(`[^A-Za-z0-9]+`)

// isSecretKey reports whether a config var name looks like it holds a secret.
// It errs towards masking: any word ending in KEY counts (AUTHKEY, APIKEY),
// even harmless ones like MONKEY.
func isSecretKey(key string) bool {
	for _, w := range nonAlnumRE.Split(strings.ToUpper(key), -1) {
		if secretKeyWords[w] || strings.HasSuffix(w, "KEY") || containsAny(w, secretKeyParts) {
			return true
		}
	}
	return false
}

func containsAny(s string, parts []string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// maskConfigValue hides a config value entirely if its key looks secret, and
// otherwise masks any password embedded in it (e.g. in DATABASE_URL).
func maskConfigValue(key, value string) string {
	if value != "" && isSecretKey(key) {
		return secretMask
	}
	return maskSecrets(value)
}

// renderConfig renders config vars sorted by key, masking secrets unless reveal.
func renderConfig(styles DetailStyles, config map[string]string, reveal bool) string {
	if len(config) == 0 {
		return "No config vars set."
	}
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		v := config[k]
		if !reveal {
			v = maskConfigValue(k, v)
		}
		b.WriteString(styles.Label.Render(k+":") + " " + styles.Value.Render(v) + "\n")
	}
	return b.String()
}

// revealKey toggles showing secrets on the Config and DSN tabs.
const revealKey = "v"

func revealHelp(revealed bool) string {
	if revealed {
		return revealKey + ": hide"
	}
	return revealKey + ": reveal"
}
