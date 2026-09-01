package godiff

import (
	"encoding/json/v2"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAliasedSlicesWithDifferentLengths(t *testing.T) {
	backing := []int{1, 2}
	left := backing[:1]
	right := backing[:2]

	result, err := Compare(left, right)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if result.Count() != 1 {
		t.Fatalf("expected one addition, got %s", result)
	}
	diff, ok := result.Diffs[0].(*SliceDiff)
	if !ok || diff.Index != 1 || diff.ChangeType != ChangeTypeAdded || diff.Right != 2 {
		t.Fatalf("unexpected diff: %#v", result.Diffs[0])
	}
}

func TestReferenceCyclesInMapsAndSlices(t *testing.T) {
	t.Run("maps", func(t *testing.T) {
		left := map[string]any{"value": "left"}
		right := map[string]any{"value": "right"}
		left["self"] = left
		right["self"] = right

		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if result.Count() != 1 {
			t.Fatalf("expected the non-cyclic difference, got %s", result)
		}
		if diff, ok := result.Diffs[0].(*MapDiff); !ok || diff.Path != "[value]" {
			t.Fatalf("unexpected diff: %#v", result.Diffs[0])
		}
	})

	t.Run("slices", func(t *testing.T) {
		left := make([]any, 2)
		right := make([]any, 2)
		left[0], right[0] = "left", "right"
		left[1], right[1] = left, right

		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if result.Count() != 1 {
			t.Fatalf("expected the non-cyclic difference, got %s", result)
		}
		if diff, ok := result.Diffs[0].(*SliceDiff); !ok || diff.Index != 0 {
			t.Fatalf("unexpected diff: %#v", result.Diffs[0])
		}
	})
}

func TestDeepReferenceCycle(t *testing.T) {
	type node struct {
		Value int
		Next  *node
	}

	left := &node{Value: 1}
	left.Next = &node{Value: 2}
	left.Next.Next = &node{Value: 3}
	left.Next.Next.Next = left

	right := &node{Value: 1}
	right.Next = &node{Value: 20}
	right.Next.Next = &node{Value: 3}
	right.Next.Next.Next = right

	result, err := Compare(left, right)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if result.Count() != 1 {
		t.Fatalf("expected the non-cyclic difference, got %s", result)
	}
	if diff, ok := result.Diffs[0].(*StructDiff); !ok || diff.Path != "Next.Value" {
		t.Fatalf("unexpected diff: %#v", result.Diffs[0])
	}
}

func TestMapWithNonReflexiveKeyDoesNotPanic(t *testing.T) {
	left := map[float64]string{math.NaN(): "value"}
	right := map[float64]string{math.NaN(): "value"}

	result, err := Compare(left, right)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if result.Count() != 2 {
		t.Fatalf("NaN keys cannot match and should be reported as remove/add, got %s", result)
	}
}

func TestMapsSupportStructKeysWithUnexportedFields(t *testing.T) {
	now := time.Now()
	left := map[time.Time]int{now: 1}
	right := map[time.Time]int{now: 2}

	result, err := Compare(left, right)
	if err != nil || result.Count() != 1 {
		t.Fatalf("struct map keys should compare safely: result=%v err=%v", result, err)
	}
}

func TestNumericComparisonIsExact(t *testing.T) {
	tests := []struct {
		name  string
		left  any
		right any
		equal bool
	}{
		{"exact large integer", int64(1 << 53), float64(1 << 53), true},
		{"rounded signed integer", int64(1<<53 + 1), float64(1 << 53), false},
		{"rounded unsigned integer", uint64(math.MaxUint64), float64(math.MaxUint64), false},
		{"fractional float", int64(1), 1.5, false},
		{"infinite float", int64(0), math.Inf(1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right, WithCompareNumericValues())
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}
			if got := !result.HasDifferences(); got != tt.equal {
				t.Fatalf("equal = %v, want %v; %s", got, tt.equal, result)
			}
		})
	}
}

func TestOptionsApplyInsideContainers(t *testing.T) {
	caseInsensitive := map[reflect.Type]func(any, any, *CompareConfig) (bool, error){
		reflect.TypeFor[string](): func(left, right any, _ *CompareConfig) (bool, error) {
			return strings.EqualFold(left.(string), right.(string)), nil
		},
	}

	t.Run("numeric values in ordered slices", func(t *testing.T) {
		result, err := Compare([]any{int64(42)}, []any{float64(42)}, WithCompareNumericValues())
		if err != nil || result.HasDifferences() {
			t.Fatalf("configured numeric comparison was not applied: result=%v err=%v", result, err)
		}
	})

	t.Run("custom comparator in struct and slice", func(t *testing.T) {
		type container struct{ Value string }
		for _, values := range [][2]any{
			{container{Value: "VALUE"}, container{Value: "value"}},
			{[]string{"VALUE"}, []string{"value"}},
			{map[string]string{"value": "VALUE"}, map[string]string{"value": "value"}},
		} {
			result, err := Compare(values[0], values[1], WithCustomComparators(caseInsensitive))
			if err != nil || result.HasDifferences() {
				t.Fatalf("custom comparator was not applied: result=%v err=%v", result, err)
			}
		}
	})

	t.Run("unordered comparison uses configured semantics", func(t *testing.T) {
		result, err := Compare(
			[]string{"ONE", "TWO"},
			[]string{"two", "one"},
			WithIgnoreSliceOrder(),
			WithCustomComparators(caseInsensitive),
		)
		if err != nil || result.HasDifferences() {
			t.Fatalf("unordered comparison ignored the comparator: result=%v err=%v", result, err)
		}
	})

	t.Run("unordered comparison honors ignored fields", func(t *testing.T) {
		type item struct {
			ID    int
			Audit string `diff:"ignore"`
		}
		left := []item{{ID: 1, Audit: "old-a"}, {ID: 2, Audit: "old-b"}}
		right := []item{{ID: 2, Audit: "new-b"}, {ID: 1, Audit: "new-a"}}
		result, err := Compare(left, right, WithIgnoreSliceOrder())
		if err != nil || result.HasDifferences() {
			t.Fatalf("unordered comparison ignored struct tags: result=%v err=%v", result, err)
		}
	})
}

func TestComparisonConfigurationIsDefensive(t *testing.T) {
	result, err := Compare(1, 2, nil, WithTypeHandlers([]TypeHandler{nil}))
	if err != nil || result.Count() != 1 {
		t.Fatalf("nil options and handlers should be ignored: result=%v err=%v", result, err)
	}

	_, err = Compare(1, 2, WithCustomComparators(map[reflect.Type]func(any, any, *CompareConfig) (bool, error){
		reflect.TypeFor[int](): nil,
	}))
	if err == nil || !strings.Contains(err.Error(), "is nil") {
		t.Fatalf("expected a descriptive nil comparator error, got %v", err)
	}
}

func TestAdditionalTypeHandlersRetainDefaults(t *testing.T) {
	result, err := Compare(
		SpecialString("VALUE"),
		SpecialString("value"),
		WithAdditionalTypeHandlers(&SpecialStringHandler{}),
	)
	if err != nil || result.HasDifferences() {
		t.Fatalf("additional handler was not applied: result=%v err=%v", result, err)
	}

	leftTime := time.Now()
	result, err = Compare(leftTime, leftTime.Add(time.Second), WithAdditionalTypeHandlers(&SpecialStringHandler{}))
	if err != nil || result.Count() != 1 {
		t.Fatalf("default time handler was not retained: result=%v err=%v", result, err)
	}
}

func TestMaxDepthAppliesToCollectionFields(t *testing.T) {
	type container struct{ Values []int }
	result, err := Compare(
		container{Values: []int{1}},
		container{Values: []int{2}},
		WithMaxDepth(1),
	)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if result.HasDifferences() {
		t.Fatalf("collection field exceeded the configured depth: %s", result)
	}

	for _, option := range []CompareOption{nil, WithIgnoreSliceOrder()} {
		options := []CompareOption{WithMaxDepth(1)}
		if option != nil {
			options = append(options, option)
		}
		result, err = Compare([]int{1}, []int{2}, options...)
		if err != nil || result.HasDifferences() {
			t.Fatalf("root slice elements exceeded the configured depth: result=%v err=%v", result, err)
		}
	}
}

func TestCustomComparatorRemainsAuthoritativeForSharedReferences(t *testing.T) {
	shared := []int{1}
	result, err := Compare(shared, shared, WithCustomComparators(map[reflect.Type]func(any, any, *CompareConfig) (bool, error){
		reflect.TypeFor[[]int](): func(_, _ any, _ *CompareConfig) (bool, error) { return false, nil },
	}))
	if err != nil || result.Count() != 1 {
		t.Fatalf("custom comparator should run for a shared reference: result=%v err=%v", result, err)
	}
}

func TestJSONRetainsZeroIndexAndSupportsMarshalJSON(t *testing.T) {
	result, err := Compare([]int{1}, []int{2})
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	for name, data := range map[string][]byte{
		"ToJSON":      []byte(result.ToJSON()),
		"MarshalJSON": mustCallMarshalJSON(t, result),
		"MarshalerTo": mustMarshalJSON(t, result),
	} {
		t.Run(name, func(t *testing.T) {
			var changes []map[string]any
			if err := json.Unmarshal(data, &changes); err != nil {
				t.Fatalf("invalid JSON: %v\n%s", err, data)
			}
			index, exists := changes[0]["index"]
			if !exists || index != float64(0) {
				t.Fatalf("index zero was omitted: %s", data)
			}
		})
	}
}

func TestJSONRetainsEmptyMapKey(t *testing.T) {
	result, err := Compare(map[string]int{"": 1}, map[string]int{"": 2})
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	var changes []map[string]any
	if err := json.Unmarshal([]byte(result.ToJSON()), &changes); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	key, exists := changes[0]["key"]
	if !exists || key != "" {
		t.Fatalf("empty map key was omitted: %s", result.ToJSON())
	}
}

func TestJSONRetainsNonNilEmptyValues(t *testing.T) {
	result := &DiffResult{Diffs: []any{
		&Diff{Path: "value", Left: "", Right: []int{}},
	}}

	var changes []map[string]any
	if err := json.Unmarshal([]byte(result.ToJSON()), &changes); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if left, exists := changes[0]["leftValue"]; !exists || left != "" {
		t.Fatalf("empty string was omitted: %s", result.ToJSON())
	}
	if right, exists := changes[0]["rightValue"]; !exists {
		t.Fatalf("empty slice was omitted: %s", result.ToJSON())
	} else if values, ok := right.([]any); !ok || len(values) != 0 {
		t.Fatalf("empty slice changed shape: %#v", right)
	}
}

func TestJSONV2RejectsInvalidUTF8(t *testing.T) {
	result := &DiffResult{Diffs: []any{
		&Diff{Path: "value", Left: string([]byte{0xff}), Right: "valid"},
	}}
	if _, err := json.Marshal(result); err == nil {
		t.Fatal("encoding/json/v2 should reject invalid UTF-8")
	}
}

func mustCallMarshalJSON(t *testing.T, result *DiffResult) []byte {
	t.Helper()
	data, err := result.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	return data
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	return data
}

func TestNilDiffResultMethods(t *testing.T) {
	var result *DiffResult
	if result.HasDifferences() || result.Count() != 0 || result.String() != "No differences found" || result.ToJSON() != "[]" {
		t.Fatal("nil DiffResult should behave like an empty result")
	}
}
