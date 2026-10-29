package godiff

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestOutputSupportsDiffValuesAndPointers(t *testing.T) {
	value := Diff{Path: "value", Left: 1, Right: 2}
	mapChange := MapDiff{Path: "items[key]", Left: 1, Right: 2, Key: "key", ChangeType: ChangeTypeUpdated}
	sliceChange := SliceDiff{Path: "items", Right: 2, Index: 0, ChangeType: ChangeTypeAdded}
	structChange := StructDiff{Path: "user.Name", Left: "old", FieldName: "Name", ChangeType: ChangeTypeRemoved}
	pointers := &DiffResult{Diffs: []any{&value, &mapChange, &sliceChange, &structChange}}
	values := &DiffResult{Diffs: []any{value, mapChange, sliceChange, structChange}}

	if got, want := values.String(), pointers.String(); got != want {
		t.Errorf("value and pointer text differs:\n%s\n%s", got, want)
	}
	if got, want := values.ToJSON(), pointers.ToJSON(); got != want {
		t.Errorf("value and pointer JSON differs:\n%s\n%s", got, want)
	}
	if strings.Contains(values.ToJSON(), "unknown") {
		t.Fatal("documented diff value types must be recognized")
	}
}

func TestOutputHandlesNilDiffEntries(t *testing.T) {
	result := &DiffResult{Diffs: []any{nil, (*Diff)(nil), (*MapDiff)(nil), (*SliceDiff)(nil), (*StructDiff)(nil)}}
	if got := strings.Count(result.String(), "Unknown diff type"); got != result.Count() {
		t.Errorf("got %d unknown text entries, want %d", got, result.Count())
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var changes []map[string]any
	if err := json.Unmarshal(data, &changes); err != nil {
		t.Fatal(err)
	}
	if len(changes) != result.Count() {
		t.Fatalf("got %d JSON entries, want %d", len(changes), result.Count())
	}
	for _, change := range changes {
		if change["type"] != "unknown" || change["change"] != "UNKNOWN" {
			t.Errorf("unexpected nil entry: %v", change)
		}
	}
}

func TestStringPreservesStructLocation(t *testing.T) {
	for _, tc := range []struct {
		path, field, want string
	}{
		{"", "Name", "Name"},
		{"user.Name", "Name", "user.Name"},
		{"user.Name", "", "user.Name"},
	} {
		result := &DiffResult{}
		result.AddStructDiff(tc.path, tc.field, "old", "new", ChangeTypeUpdated)
		if want := "UPDATED " + tc.want + ": old -> new"; !strings.Contains(result.String(), want) {
			t.Errorf("String() = %q, want location %q", result.String(), want)
		}
	}
}

type scalarStringer string

func (s scalarStringer) String() string { return "string: " + string(s) }

type scalarFormatter int

func (n scalarFormatter) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "formatted: %d", int(n))
}

type scalarError bool

func (b scalarError) Error() string { return fmt.Sprintf("error: %t", bool(b)) }

func TestStringPreservesScalarFormatting(t *testing.T) {
	type namedString string
	type namedInt int
	type namedBool bool
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"empty string", ""},
		{"string", "hello"},
		{"string with special characters", "line\n\t\"text\"\x00"},
		{"false", false},
		{"true", true},
		{"zero", 0},
		{"positive int", 12345},
		{"negative int", -12345},
		{"maximum int", maxInt},
		{"minimum int", -maxInt - 1},
		{"named string", namedString("hello")},
		{"named int", namedInt(12345)},
		{"named bool", namedBool(true)},
		{"stringer", scalarStringer("hello")},
		{"formatter", scalarFormatter(12345)},
		{"error", scalarError(true)},
		{"typed nil", (*int)(nil)},
		{"typed nil stringer", (*scalarStringer)(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := &DiffResult{}
			result.AddDiff("value", tc.value, tc.value)
			formatted := fmt.Sprint(tc.value)
			want := "Found 1 differences:\nUPDATED value: " + formatted + " -> " + formatted + "\n"
			if got := result.String(); got != want {
				t.Errorf("String() = %q, want %q", got, want)
			}
		})
	}
}

type cyclicStringer map[string]any

func (cyclicStringer) String() string { return "custom string" }

type cyclicFormatter map[string]any

func (cyclicFormatter) Format(state fmt.State, _ rune) { _, _ = fmt.Fprint(state, "custom format") }

func TestStringHandlesCyclicValues(t *testing.T) {
	cyclicMap := map[string]any{}
	cyclicMap["self"] = cyclicMap
	cyclicSlice := []any{nil}
	cyclicSlice[0] = cyclicSlice
	for _, value := range []any{
		cyclicMap,
		cyclicSlice,
		&cyclicMap,
		struct{ Value any }{cyclicMap},
		map[string]any{"nested": cyclicSlice},
		reflect.ValueOf(cyclicMap),
		reflect.ValueOf(cyclicSlice),
	} {
		result, err := Compare(nil, value)
		if err != nil {
			t.Fatal(err)
		}
		if got := result.String(); !strings.Contains(got, "<cycle:") {
			t.Errorf("expected cycle marker for %T, got %s", value, got)
		}
	}
	for _, key := range []any{cyclicMap, reflect.ValueOf(cyclicMap)} {
		result := &DiffResult{}
		result.AddMapDiff("key", key, nil, 1, ChangeTypeAdded)
		if got := result.ToJSON(); !strings.Contains(got, "<cycle:") {
			t.Errorf("expected cycle marker for manually supplied cyclic map key: %s", got)
		}
	}
}

func TestStringRetainsCustomFormattingAndSharedValues(t *testing.T) {
	stringer := cyclicStringer{}
	stringer["self"] = stringer
	formatter := cyclicFormatter{}
	formatter["self"] = formatter
	type node struct{ Next *node }
	cycle := &node{}
	cycle.Next = cycle
	shared := map[string]int{"value": 1}
	cyclicMap := map[string]any{}
	cyclicMap["self"] = cyclicMap
	var boxed any = cyclicMap
	for _, value := range []any{
		stringer, formatter, cycle, &boxed, []any{shared, shared},
		reflect.Value{}, reflect.ValueOf(shared), reflect.ValueOf(stringer),
		[]any{reflect.ValueOf(cyclicMap)}, reflect.ValueOf(reflect.ValueOf(cyclicMap)),
	} {
		result := &DiffResult{}
		result.AddDiff("value", nil, value)
		if got, want := result.String(), "<nil> -> "+fmt.Sprint(value); !strings.Contains(got, want) {
			t.Errorf("formatting changed for %T: got %s, want %s", value, got, want)
		}
	}
}

type failingJSONWriter struct{ err error }

func (w failingJSONWriter) Write([]byte) (int, error) { return 0, w.err }

func TestWriteJSON(t *testing.T) {
	result := &DiffResult{}
	result.AddStructDiff("user.Name", "Name", "old", "new", ChangeTypeUpdated)
	result.AddMapDiff("[key]", "key", nil, map[string]int{"b": 2, "a": 1}, ChangeTypeAdded)
	for _, input := range []*DiffResult{result, {}, nil} {
		var output bytes.Buffer
		if err := input.WriteJSON(&output); err != nil {
			t.Fatal(err)
		}
		if got, want := output.String(), input.ToJSON(); got != want {
			t.Errorf("WriteJSON = %q, ToJSON = %q", got, want)
		}
	}

	t.Run("writer errors", func(t *testing.T) {
		want := errors.New("write failed")
		if err := result.WriteJSON(failingJSONWriter{want}); !errors.Is(err, want) {
			t.Errorf("WriteJSON error = %v, want %v", err, want)
		}
	})
	t.Run("serialization errors", func(t *testing.T) {
		invalid := &DiffResult{}
		invalid.AddDiff("callback", func() {}, nil)
		var output bytes.Buffer
		if err := invalid.WriteJSON(&output); err == nil {
			t.Fatal("WriteJSON must return serialization errors")
		}
	})
}

func TestInterfaceHandlerConcreteValues(t *testing.T) {
	handler := &InterfaceHandler{}
	for _, tc := range []struct {
		name        string
		left, right any
		count       int
	}{
		{"strings equal", "same", "same", 0},
		{"strings different", "left", "right", 1},
		{"numbers", 1, 2, 1},
		{"slices", []int{1}, []int{2}, 1},
		{"typed nil pointers", (*any)(nil), (*any)(nil), 0},
		{"typed nil and value", (*any)(nil), "value", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := &DiffResult{}
			if err := handler.Compare(tc.left, tc.right, "field", result, DefaultCompareConfig()); err != nil {
				t.Fatal(err)
			}
			if got := result.Count(); got != tc.count {
				t.Errorf("got %d differences, want %d", got, tc.count)
			}
		})
	}
}
