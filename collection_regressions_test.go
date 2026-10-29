package godiff

import (
	"math"
	"reflect"
	"testing"
)

func TestCollectionChangesRespectIgnoredPaths(t *testing.T) {
	tests := []struct {
		name        string
		left, right any
		path        string
	}{
		{"map addition", map[string]int{}, map[string]int{"secret": 1}, "[secret]"},
		{"map removal", map[string]int{"secret": 1}, map[string]int{}, "[secret]"},
		{"slice addition", []int{1}, []int{1, 2}, "[1]"},
		{"slice removal", []int{1, 2}, []int{1}, "[1]"},
		{"slice replacement", []int{1, 2}, []int{3, 2}, "[0]"},
		{"dynamic slice replacement", []any{1, 2}, []any{3, 2}, "[0]"},
	}
	for _, tt := range tests {
		for _, unordered := range []bool{false, true} {
			t.Run(tt.name+map[bool]string{false: "/ordered", true: "/unordered"}[unordered], func(t *testing.T) {
				opts := []CompareOption{WithIgnoreFields(tt.path)}
				if unordered {
					opts = append(opts, WithIgnoreSliceOrder())
				}
				result, err := Compare(tt.left, tt.right, opts...)
				if err != nil || result.HasDifferences() {
					t.Fatalf("ignored collection entry produced a difference: result=%v err=%v", result, err)
				}
			})
		}
	}
}

func TestCollectionChangesRespectMaxDepth(t *testing.T) {
	tests := []struct {
		name        string
		left, right any
	}{
		{"map addition", map[string]int{}, map[string]int{"x": 1}},
		{"map removal", map[string]int{"x": 1}, map[string]int{}},
		{"slice addition", []int{}, []int{1}},
		{"slice removal", []int{1}, []int{}},
	}
	for _, tt := range tests {
		for _, unordered := range []bool{false, true} {
			t.Run(tt.name+map[bool]string{false: "/ordered", true: "/unordered"}[unordered], func(t *testing.T) {
				opts := []CompareOption{WithMaxDepth(1)}
				if unordered {
					opts = append(opts, WithIgnoreSliceOrder())
				}
				result, err := Compare(tt.left, tt.right, opts...)
				if err != nil || result.HasDifferences() {
					t.Fatalf("collection entry exceeded the depth limit: result=%v err=%v", result, err)
				}
			})
		}
	}
}

func TestUnorderedComparisonFindsCompleteMatching(t *testing.T) {
	withinOne := WithCustomComparators(map[reflect.Type]func(any, any, *CompareConfig) (bool, error){
		reflect.TypeFor[int](): func(left, right any, _ *CompareConfig) (bool, error) {
			return math.Abs(float64(left.(int)-right.(int))) <= 1, nil
		},
	})
	// Matching 2 with 1 first would leave 1 unmatched. A complete matching
	// exists: 2 with 3, and 1 with 1, regardless of the input order.
	for _, left := range [][]int{{2, 1}, {1, 2}} {
		for _, right := range [][]int{{1, 3}, {3, 1}} {
			result, err := Compare(left, right, WithIgnoreSliceOrder(), withinOne)
			if err != nil || result.HasDifferences() {
				t.Fatalf("matching %v with %v failed: result=%v err=%v", left, right, result, err)
			}
		}
	}
}

func TestUnorderedComparisonReassignsMultipleMatches(t *testing.T) {
	comparators := WithCustomComparators(map[reflect.Type]func(any, any, *CompareConfig) (bool, error){
		reflect.TypeFor[int](): func(left, right any, _ *CompareConfig) (bool, error) {
			switch left.(int) {
			case 0:
				return right == 3 || right == 4, nil
			case 1:
				return right == 4 || right == 5, nil
			case 2:
				return right == 3, nil
			default:
				return false, nil
			}
		},
	})
	for _, left := range [][]int{{0, 1, 2}, {0, 1, 2, 2}} {
		result, err := Compare(left, []int{3, 4, 5}, WithIgnoreSliceOrder(), comparators)
		if err != nil || result.Count() != len(left)-3 {
			t.Fatalf("expected a complete matching with only excess duplicates removed: result=%v err=%v", result, err)
		}
		if result.Count() == 1 {
			diff := result.Diffs[0].(*Diff)
			if diff.Left != 2 || diff.Right != nil {
				t.Fatalf("expected the unmatched duplicate to be removed: %+v", diff)
			}
		}
	}
}

func TestUnorderedIgnoredIndexDoesNotConsumeOtherValues(t *testing.T) {
	for _, values := range []struct{ left, right any }{
		{[]int{99, 1}, []int{1, 99}},
		{[]any{99, 1}, []any{1, 99}},
	} {
		result, err := Compare(values.left, values.right, WithIgnoreSliceOrder(), WithIgnoreFields("[0]"))
		if err != nil || result.Count() != 2 {
			t.Fatalf("expected removal of 1 and addition of 99 after filtering index zero: result=%v err=%v", result, err)
		}
	}
}

func BenchmarkUnorderedDuplicateCounts(b *testing.B) {
	left := make([]int, 1000)
	right := make([]int, 999)
	for i := range left {
		left[i] = i % 10
	}
	copy(right, left)
	option := WithIgnoreSliceOrder()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Compare(left, right, option); err != nil {
			b.Fatal(err)
		}
	}
}
