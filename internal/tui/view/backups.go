package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// The backup schema of "service backup ls -o json" is not documented, so
// backups are rendered generically: any list of objects (bare, or wrapped in
// "results"/"data") becomes a table whose columns are the objects' keys.
// Anything else is shown as pretty-printed JSON.

// backupColumnOrder lists keys shown first, in this order; other keys follow
// alphabetically.
var backupColumnOrder = []string{"name", "id", "created", "status", "size"}

func renderBackups(raw json.RawMessage) string {
	items, ok := backupItems(raw)
	if !ok {
		return prettyJSON(raw)
	}
	if len(items) == 0 {
		return "No backups yet. Backups aren't available on dev plans."
	}
	cols := backupColumns(items)
	rows := make([][]string, len(items))
	for i, it := range items {
		rows[i] = make([]string, len(cols))
		for j, c := range cols {
			rows[i][j] = formatBackupValue(it[c])
		}
	}
	return textTable(cols, rows)
}

// backupItems extracts the list of backup objects. ok is false if raw is not
// a recognised shape.
func backupItems(raw json.RawMessage) (items []map[string]any, ok bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, true
	}
	if decodeNumbers(trimmed, &items) == nil {
		return items, true
	}
	var wrapper map[string]json.RawMessage
	if decodeNumbers(trimmed, &wrapper) != nil {
		return nil, false
	}
	for _, key := range []string{"results", "data"} {
		if inner, found := wrapper[key]; found {
			return backupItems(inner)
		}
	}
	return nil, false
}

func decodeNumbers(data []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(v)
}

func backupColumns(items []map[string]any) []string {
	seen := map[string]bool{}
	for _, it := range items {
		for k := range it {
			seen[k] = true
		}
	}
	var cols []string
	for _, k := range backupColumnOrder {
		if seen[k] {
			cols = append(cols, k)
			delete(seen, k)
		}
	}
	rest := make([]string, 0, len(seen))
	for k := range seen {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	return append(cols, rest...)
}

func formatBackupValue(v any) string {
	switch v := v.(type) {
	case nil:
		return "-"
	case string:
		return orDash(v)
	case json.Number, bool:
		return fmt.Sprint(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(b)
	}
}

func prettyJSON(raw json.RawMessage) string {
	var b bytes.Buffer
	if json.Indent(&b, raw, "", "  ") != nil {
		return string(raw)
	}
	return b.String()
}
