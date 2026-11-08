package runway

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type sdInner struct {
	A string `json:"a"`
	B int    `json:"b"`
}

type sdOuter struct {
	Name  string             `json:"name"`
	Inner sdInner            `json:"inner"`
	List  []sdInner          `json:"list"`
	Ptr   *sdInner           `json:"ptr"`
	Map   map[string]sdInner `json:"map"`
	Skip  string             `json:"-"`
	Plain map[string]string  `json:"plain"`
}

func diffOf(t *testing.T, typ reflect.Type, raw string) (missing, unknown string) {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatal(err)
	}
	d := schemaDiff(typ, v)
	sort.Strings(d.Missing)
	sort.Strings(d.Unknown)
	return strings.Join(d.Missing, ","), strings.Join(d.Unknown, ",")
}

func TestSchemaDiffComplete(t *testing.T) {
	raw := `{"name":"x","inner":{"a":"1","b":2},"list":[{"a":"1","b":2}],"ptr":{"a":"1","b":2},
		"map":{"k":{"a":"1","b":2}},"plain":{"anything":"goes"}}`
	if missing, unknown := diffOf(t, reflect.TypeOf(sdOuter{}), raw); missing != "" || unknown != "" {
		t.Errorf("missing=%q unknown=%q, want none", missing, unknown)
	}
}

func TestSchemaDiffMissingAndUnknown(t *testing.T) {
	raw := `{"name":"x","inner":{"a":"1","extra":true},"list":[{"a":"1"},{"b":2}],"map":{"k":{"a":"1"}},"new":1}`
	missing, unknown := diffOf(t, reflect.TypeOf(sdOuter{}), raw)
	// list[].b is seen in the second item, so it is not missing.
	if want := ".inner.b,.map{}.b,.plain,.ptr"; missing != want {
		t.Errorf("missing = %q, want %q", missing, want)
	}
	if want := ".inner.extra,.new"; unknown != want {
		t.Errorf("unknown = %q, want %q", unknown, want)
	}
}

func TestSchemaDiffNullAndEmptyAreNotDescended(t *testing.T) {
	// A null/empty value means the key exists but there is nothing to check below it.
	raw := `{"name":"x","inner":{"a":"1","b":2},"list":[],"ptr":null,"map":{},"plain":null}`
	if missing, unknown := diffOf(t, reflect.TypeOf(sdOuter{}), raw); missing != "" || unknown != "" {
		t.Errorf("missing=%q unknown=%q, want none", missing, unknown)
	}
}

func TestSchemaDiffTopLevelSlice(t *testing.T) {
	missing, _ := diffOf(t, reflect.TypeOf([]sdInner{}), `[{"a":"1"}]`)
	if missing != "[].b" {
		t.Errorf("missing = %q, want [].b", missing)
	}
}

// The stats regression this guards against: the old {all, latest} payload
// decoded into Stats without error but every metric field was missing.
func TestSchemaDiffCatchesOldStatsFormat(t *testing.T) {
	raw := `{"all":[{"ts":1,"val":{}}],"latest":{"cpu":{"load":0.1}}}`
	if missing, _ := diffOf(t, reflect.TypeOf(Stats{}), raw); !strings.Contains(missing, "{}.title") {
		t.Errorf("old stats format should report missing metric fields, got %q", missing)
	}
}
