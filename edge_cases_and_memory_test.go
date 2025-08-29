package godiff

import (
	"testing"
)

func TestEdgeCases(t *testing.T) {
	t.Run("Compare nil slices", func(t *testing.T) {
		var leftSlice, rightSlice []int
		result, err := Compare(leftSlice, rightSlice)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for nil slices, got %d", len(result.Diffs))

			for i, diff := range result.Diffs {
				t.Errorf("Unexpected diff %d: %+v", i, diff)
			}
		}
	})

	t.Run("Compare empty slice with nil slice", func(t *testing.T) {
		leftSlice := []int{}
		var rightSlice []int
		result, err := Compare(leftSlice, rightSlice)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}

		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for empty vs nil slice, got %d", len(result.Diffs))

			for i, diff := range result.Diffs {
				t.Errorf("Unexpected diff %d: %+v", i, diff)
			}
		}
	})

	t.Run("Compare functions with nil", func(t *testing.T) {
		var f1, f2 func()
		result, err := Compare(f1, f2)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for nil functions, got %d", len(result.Diffs))

			for i, diff := range result.Diffs {
				t.Errorf("Unexpected diff %d: %+v", i, diff)
			}
		}

		f1 = func() {}
		result, err = Compare(f1, f2)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference for function vs nil, got %d", len(result.Diffs))

			for i, diff := range result.Diffs {
				t.Errorf("Diff %d: %+v", i, diff)
			}
		} else {

			if diff, ok := result.Diffs[0].(*Diff); ok {
				if diff.Left == nil || diff.Right == nil {
					t.Errorf("Expected both left and right values to be non-nil for function diff, got %+v -> %+v", diff.Left, diff.Right)
				}
			} else {
				t.Errorf("Expected Diff type, got %T", result.Diffs[0])
			}
		}
	})

	t.Run("HasDifferences and Count methods", func(t *testing.T) {
		result1, _ := Compare("same", "same")
		if result1.HasDifferences() {
			t.Error("Expected no differences")
		}
		if result1.Count() != 0 {
			t.Errorf("Expected count 0, got %d", result1.Count())
		}

		result2, _ := Compare("left", "right")
		if !result2.HasDifferences() {
			t.Error("Expected differences")
		}
		if result2.Count() != 1 {
			t.Errorf("Expected count 1, got %d", result2.Count())
		}
	})

	t.Run("Large path names", func(t *testing.T) {
		type DeepNested struct {
			Level1 struct {
				Level2 struct {
					Level3 struct {
						Level4 struct {
							Level5 struct {
								Value string
							}
						}
					}
				}
			}
		}

		left := DeepNested{}
		left.Level1.Level2.Level3.Level4.Level5.Value = "left"

		right := DeepNested{}
		right.Level1.Level2.Level3.Level4.Level5.Value = "right"

		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference for deeply nested struct, got %d", len(result.Diffs))
		}

		expectedPath := "Level1.Level2.Level3.Level4.Level5.Value"
		if diff, ok := result.Diffs[0].(*StructDiff); ok {
			if diff.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, diff.Path)
			}
		} else {
			t.Error("Expected a StructDiff")
		}
	})

	t.Run("Test abs function", func(t *testing.T) {
		if abs(5) != 5 {
			t.Errorf("abs(5) should be 5, got %d", abs(5))
		}
		if abs(-5) != 5 {
			t.Errorf("abs(-5) should be 5, got %d", abs(-5))
		}
		if abs(0) != 0 {
			t.Errorf("abs(0) should be 0, got %d", abs(0))
		}
	})
}

func TestMemoryEfficiency(t *testing.T) {
	t.Run("Large slice comparison doesn't panic", func(t *testing.T) {
		left := make([]int, 1000)
		right := make([]int, 1000)

		for i := range left {
			left[i] = i
			right[i] = i + 1
		}

		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if len(result.Diffs) != 1000 {
			t.Errorf("Expected 1000 differences, got %d", len(result.Diffs))

			for i, diff := range result.Diffs {
				t.Errorf("Diff %d: %+v", i, diff)
			}
		} else {

			for i := 0; i < 5 && i < len(result.Diffs); i++ {
				if sliceDiff, ok := result.Diffs[i].(*SliceDiff); ok {
					if sliceDiff.Left != i || sliceDiff.Right != i+1 {
						t.Errorf("Expected diff %d: %d -> %d, got %+v -> %+v", i, i, i+1, sliceDiff.Left, sliceDiff.Right)
					}
				} else {
					t.Errorf("Expected SliceDiff type for diff %d, got %T", i, result.Diffs[i])
				}
			}
		}
	})

	t.Run("Complex nested structure", func(t *testing.T) {
		type Complex struct {
			Maps   map[string]map[string][]string
			Slices []map[string][]int
		}

		left := Complex{
			Maps: map[string]map[string][]string{
				"outer1": {
					"inner1": {"a", "b", "c"},
					"inner2": {"d", "e", "f"},
				},
			},
			Slices: []map[string][]int{
				{"key1": {1, 2, 3}},
				{"key2": {4, 5, 6}},
			},
		}

		right := Complex{
			Maps: map[string]map[string][]string{
				"outer1": {
					"inner1": {"a", "b", "modified"},
					"inner2": {"d", "e", "f"},
				},
			},
			Slices: []map[string][]int{
				{"key1": {1, 2, 3}},
				{"key2": {4, 5, 7}},
			},
		}

		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if len(result.Diffs) != 2 {
			t.Errorf("Expected 2 differences, got %d: %s", len(result.Diffs), result.String())

			for i, diff := range result.Diffs {
				t.Errorf("Diff %d: %+v", i, diff)
			}
		} else {
			if len(result.Diffs) != 2 {
				t.Errorf("Expected exactly 2 differences, got %d", len(result.Diffs))
			}

			for i, diff := range result.Diffs {
				if diff == nil {
					t.Errorf("Diff %d is nil", i)
					continue
				}

				switch d := diff.(type) {
				case *SliceDiff:
					if d.Path == "" {
						t.Errorf("SliceDiff %d has empty path", i)
					}
				case *MapDiff:
					if d.Key == "" {
						t.Errorf("MapDiff %d has empty key", i)
					}
				case *StructDiff:
					if d.Path == "" {
						t.Errorf("StructDiff %d has empty path", i)
					}
				case *Diff:
					if d.Left == nil && d.Right == nil {
						t.Errorf("Diff %d has both left and right as nil", i)
					}
				default:
					t.Errorf("Diff %d has unexpected type %T", i, diff)
				}
			}
		}
	})
}

func TestVisitedPairsInitialization(t *testing.T) {
	config := &CompareConfig{}

	type Node struct {
		Value int
		Next  *Node
	}

	left := &Node{Value: 1}
	left.Next = left

	right := &Node{Value: 1}
	right.Next = right

	result, err := CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestComparePointersEdgeCases(t *testing.T) {
	type Node struct {
		Value int
		Next  *Node
	}

	tests := []struct {
		name     string
		left     any
		right    any
		expected int
	}{
		{
			name:     "same pointer address",
			left:     func() *int { x := 42; return &x }(),
			right:    func() *int { x := 42; return &x }(),
			expected: 0,
		},
		{
			name:     "circular reference handling",
			left:     func() *Node { n := &Node{Value: 1}; n.Next = n; return n }(),
			right:    func() *Node { n := &Node{Value: 1}; n.Next = n; return n }(),
			expected: 0,
		},
		{
			name:     "nil vs non-nil pointer",
			left:     (*int)(nil),
			right:    func() *int { x := 42; return &x }(),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Diffs) != tt.expected {
				t.Errorf("expected %d differences, got %d", tt.expected, len(result.Diffs))

				for i, diff := range result.Diffs {
					t.Errorf("Diff %d: %+v", i, diff)
				}
			} else if tt.expected > 0 {
				if len(result.Diffs) > 0 {
					if diff, ok := result.Diffs[0].(*Diff); ok {
						if diff.Left == nil && diff.Right == nil {
							t.Errorf("Expected meaningful diff values, got both nil")
						}
					} else {
						t.Errorf("Expected Diff type, got %T", result.Diffs[0])
					}
				}
			}
		})
	}
}
