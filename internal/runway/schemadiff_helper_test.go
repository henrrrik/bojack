package runway

import (
	"reflect"
	"sort"
	"strings"
)

// schemaDiffResult lists JSON field paths that a Go type expects but the data
// never contained (Missing), and keys the data had that the type does not
// declare (Unknown). Paths use "." for fields, "[]" for slice elements and
// "{}" for map values, e.g. "[].app.meta.role" or "{}.title".
type schemaDiffResult struct {
	Missing []string
	Unknown []string
}

// schemaDiff compares decoded JSON (as produced by json.Unmarshal into any)
// with the JSON shape of typ. A field counts as present if any object at its
// path has the key, even with a null or empty value; such values are not
// descended into, since there is nothing below them to check.
func schemaDiff(typ reflect.Type, data any) schemaDiffResult {
	w := schemaWalker{expected: map[string]bool{}, present: map[string]bool{}, unknown: map[string]bool{}}
	w.walk(typ, data, "")
	var d schemaDiffResult
	for p := range w.expected {
		if !w.present[p] {
			d.Missing = append(d.Missing, p)
		}
	}
	for p := range w.unknown {
		d.Unknown = append(d.Unknown, p)
	}
	sort.Strings(d.Missing)
	sort.Strings(d.Unknown)
	return d
}

type schemaWalker struct {
	expected, present, unknown map[string]bool
}

func (w schemaWalker) walk(typ reflect.Type, data any, path string) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if data == nil {
		return
	}
	switch typ.Kind() {
	case reflect.Struct:
		w.walkStruct(typ, data, path)
	case reflect.Slice, reflect.Array:
		items, _ := data.([]any)
		for _, it := range items {
			w.walk(typ.Elem(), it, path+"[]")
		}
	case reflect.Map:
		obj, _ := data.(map[string]any)
		for _, v := range obj {
			w.walk(typ.Elem(), v, path+"{}")
		}
	}
}

// walkStruct records typ's fields as expected at path. If data is not an
// object (e.g. the CLI changed a field's shape), every field is missing.
func (w schemaWalker) walkStruct(typ reflect.Type, data any, path string) {
	obj, isObj := data.(map[string]any)
	known := map[string]bool{}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		name := jsonFieldName(f)
		if name == "" {
			continue
		}
		p := path + "." + name
		known[name] = true
		w.expected[p] = true
		if v, ok := obj[name]; ok {
			w.present[p] = true
			w.walk(f.Type, v, p)
		}
	}
	if !isObj {
		return
	}
	for k := range obj {
		if !known[k] {
			w.unknown[path+"."+k] = true
		}
	}
}

// jsonFieldName returns the JSON key encoding/json uses for f, or "" if the
// field is not encoded.
func jsonFieldName(f reflect.StructField) string {
	if !f.IsExported() {
		return ""
	}
	tag := f.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	if name, _, _ := strings.Cut(tag, ","); name != "" {
		return name
	}
	return f.Name
}
