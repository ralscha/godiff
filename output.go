package godiff

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

var (
	_ json.Marshaler   = (*DiffResult)(nil)
	_ json.MarshalerTo = (*DiffResult)(nil)
)

// String returns a human-readable representation of the diff result
func (dr *DiffResult) String() string {
	if dr == nil || len(dr.Diffs) == 0 {
		return "No differences found"
	}

	var sb strings.Builder
	sb.Grow(30 + len(dr.Diffs)*90)

	sb.WriteString("Found ")
	sb.WriteString(strconv.Itoa(len(dr.Diffs)))
	sb.WriteString(" differences:\n")

	for _, diff := range dr.Diffs {
		d := describeDiff(diff)
		if d.kind == "unknown" {
			sb.WriteString("? Unknown diff type\n")
			continue
		}
		sb.WriteString(string(d.change))
		sb.WriteString(" ")
		if d.kind == "struct" && d.Path == "" {
			sb.WriteString(d.fieldName)
		} else {
			sb.WriteString(d.Path)
		}
		if d.kind == "slice" {
			sb.WriteString("[")
			sb.WriteString(strconv.Itoa(d.index))
			sb.WriteString("]")
		}
		sb.WriteString(": ")
		switch d.change {
		case ChangeTypeAdded:
			writeDiffValue(&sb, d.Right)
		case ChangeTypeRemoved:
			writeDiffValue(&sb, d.Left)
		default:
			writeDiffValue(&sb, d.Left)
			sb.WriteString(" -> ")
			writeDiffValue(&sb, d.Right)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func writeDiffValue(out *strings.Builder, value any) {
	if hasFormattedValueCycle(value) {
		fmt.Fprintf(out, "<cycle: %T>", value)
		return
	}
	fmt.Fprint(out, value)
}

func hasFormattedValueCycle(value any) bool {
	// fmt unwraps reflect.Value once when it is the top-level argument. Nested
	// reflect.Values instead use their String method like other Stringers.
	reflected, ok := value.(reflect.Value)
	if !ok {
		reflected = reflect.ValueOf(value)
	}
	var active visitTracker
	return hasFormattingCycle(reflected, &active, 0)
}

// fmt handles nested pointers as addresses, but recursively formats maps and
// slices without cycle detection. Check the values it will traverse before
// handing them over. User-defined formatting methods retain their behavior.
func hasFormattingCycle(value reflect.Value, active *visitTracker, depth int) bool {
	if !value.IsValid() {
		return false
	}
	if value.CanInterface() && value.Type().NumMethod() > 0 {
		switch value.Interface().(type) {
		case fmt.Formatter, fmt.Stringer, error:
			return false
		}
	}
	switch value.Kind() {
	case reflect.Interface:
		return hasFormattingCycle(value.Elem(), active, depth+1)
	case reflect.Pointer:
		if depth == 0 && !value.IsNil() {
			switch value.Elem().Kind() {
			case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct:
				return hasFormattingCycle(value.Elem(), active, depth+1)
			}
		}
	case reflect.Map, reflect.Slice:
		if value.IsNil() || value.Len() == 0 {
			return false
		}
		pair := visit{typ: value.Type(), left: value.Pointer(), leftLen: value.Len()}
		if !active.enter(pair) {
			return true
		}
		defer active.leave(pair)
		if value.Kind() == reflect.Map {
			iter := value.MapRange()
			for iter.Next() {
				if hasFormattingCycle(iter.Key(), active, depth+1) || hasFormattingCycle(iter.Value(), active, depth+1) {
					return true
				}
			}
			return false
		}
		fallthrough
	case reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if hasFormattingCycle(value.Index(i), active, depth+1) {
				return true
			}
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if hasFormattingCycle(value.Field(i), active, depth+1) {
				return true
			}
		}
	}
	return false
}

// diffDetails gives both renderers a common view of the supported value and
// pointer forms. A typed nil pointer is treated like any other unknown entry.
type diffDetails struct {
	Diff
	kind      string
	change    ChangeType
	key       any
	index     int
	fieldName string
}

func describeDiff(diff any) diffDetails {
	switch d := diff.(type) {
	case Diff:
		return diffDetails{Diff: d, kind: "value", change: ChangeTypeUpdated}
	case MapDiff:
		return diffDetails{Diff: d.Diff, kind: "map", change: d.ChangeType, key: d.Key}
	case SliceDiff:
		return diffDetails{Diff: d.Diff, kind: "slice", change: d.ChangeType, index: d.Index}
	case StructDiff:
		return diffDetails{Diff: d.Diff, kind: "struct", change: d.ChangeType, fieldName: d.FieldName}
	case *Diff:
		if d != nil {
			return describeDiff(*d)
		}
	case *MapDiff:
		if d != nil {
			return describeDiff(*d)
		}
	case *SliceDiff:
		if d != nil {
			return describeDiff(*d)
		}
	case *StructDiff:
		if d != nil {
			return describeDiff(*d)
		}
	}
	return diffDetails{Diff: Diff{Path: "unknown"}, kind: "unknown", change: "UNKNOWN"}
}

// HasDifferences returns true if there are any differences
func (dr *DiffResult) HasDifferences() bool {
	return dr != nil && len(dr.Diffs) > 0
}

// Count returns the number of differences
func (dr *DiffResult) Count() int {
	if dr == nil {
		return 0
	}
	return len(dr.Diffs)
}

type jsonChange struct {
	Type      string  `json:"type"`
	Path      string  `json:"path"`
	Left      any     `json:"leftValue,omitzero"`
	Right     any     `json:"rightValue,omitzero"`
	Key       *string `json:"key,omitzero"`
	Index     *int    `json:"index,omitzero"`
	FieldName string  `json:"fieldName,omitempty"`
	Change    string  `json:"change"`
}

// MarshalJSON implements json.Marshaler using the same change-oriented schema
// as ToJSON. In particular, slice index zero and empty map keys are retained.
func (dr *DiffResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(dr.jsonChanges(), json.Deterministic(true))
}

// MarshalJSONTo implements json.MarshalerTo without allocating an intermediate
// JSON buffer. MarshalEncode preserves any options configured by the caller.
func (dr *DiffResult) MarshalJSONTo(out *jsontext.Encoder) error {
	return json.MarshalEncode(out, dr.jsonChanges())
}

func (dr *DiffResult) jsonChanges() []jsonChange {
	count := 0
	if dr != nil {
		count = len(dr.Diffs)
	}
	changes := make([]jsonChange, 0, count)
	if dr == nil {
		return changes
	}

	for _, diff := range dr.Diffs {
		d := describeDiff(diff)
		jc := jsonChange{Type: d.kind, Path: d.Path, Left: d.Left, Right: d.Right, Change: string(d.change)}
		switch d.kind {
		case "map":
			var key string
			if hasFormattedValueCycle(d.key) {
				key = fmt.Sprintf("<cycle: %T>", d.key)
			} else {
				key = fmt.Sprint(d.key)
			}
			jc.Key = &key
		case "slice":
			index := d.index
			jc.Index = &index
		case "struct":
			jc.FieldName = d.fieldName
			if d.fieldName != "" {
				if d.Path == d.fieldName {
					jc.Path = ""
				} else {
					jc.Path = strings.TrimSuffix(d.Path, "."+d.fieldName)
				}
			}
		}
		changes = append(changes, jc)
	}
	return changes
}

// WriteJSON writes the formatted change array to out, returning serialization
// and writer errors. Like other streaming encoders, it may write partial output
// before returning an error. The output matches ToJSON on success.
func (dr *DiffResult) WriteJSON(out io.Writer) error {
	return json.MarshalWrite(out, dr.jsonChanges(), json.Deterministic(true),
		jsontext.Multiline(true), jsontext.WithIndent("  "))
}

// ToJSON returns a JSON representation of the diff result
func (dr *DiffResult) ToJSON() string {
	if dr == nil || len(dr.Diffs) == 0 {
		return `[]`
	}

	jsonBytes, err := json.Marshal(
		dr,
		json.Deterministic(true),
		jsontext.Multiline(true),
		jsontext.WithIndent("  "),
	)
	if err != nil {
		errorJSON, _ := json.Marshal(
			[]map[string]string{{"error": "Failed to marshal JSON: " + err.Error()}},
			json.Deterministic(true),
		)
		return string(errorJSON)
	}

	return string(jsonBytes)
}

// String returns a human-readable representation of the ChangeType
func (ct ChangeType) String() string {
	switch ct {
	case ChangeTypeAdded:
		return "added"
	case ChangeTypeRemoved:
		return "removed"
	case ChangeTypeUpdated:
		return "updated"
	default:
		return string(ct)
	}
}
