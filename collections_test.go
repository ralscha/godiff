package godiff

import (
	"reflect"
	"testing"
)

func TestSliceComparison(t *testing.T) {
	tests := []struct {
		name     string
		left     []int
		right    []int
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same slices", []int{1, 2, 3}, []int{1, 2, 3}, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different elements", []int{1, 2, 3}, []int{1, 4, 3}, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if sliceDiff, ok := result.Diffs[0].(*SliceDiff); ok {
					if sliceDiff.Index != 1 {
						t.Errorf("Expected diff at index 1, got %d", sliceDiff.Index)
					}
					if sliceDiff.Left != 2 {
						t.Errorf("Expected left value 2, got %v", sliceDiff.Left)
					}
					if sliceDiff.Right != 4 {
						t.Errorf("Expected right value 4, got %v", sliceDiff.Right)
					}
					if sliceDiff.ChangeType != "UPDATED" {
						t.Errorf("Expected UPDATED change type, got %s", sliceDiff.ChangeType)
					}
				} else {
					t.Error("Expected SliceDiff type")
				}
			},
		},
		{
			"Different length", []int{1, 2}, []int{1, 2, 3}, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if sliceDiff, ok := result.Diffs[0].(*SliceDiff); ok {
					if sliceDiff.Index != 2 {
						t.Errorf("Expected diff at index 2, got %d", sliceDiff.Index)
					}
					if sliceDiff.Left != nil {
						t.Errorf("Expected left value nil, got %v", sliceDiff.Left)
					}
					if sliceDiff.Right != 3 {
						t.Errorf("Expected right value 3, got %v", sliceDiff.Right)
					}
					if sliceDiff.ChangeType != "ADDED" {
						t.Errorf("Expected ADDED change type, got %s", sliceDiff.ChangeType)
					}
				} else {
					t.Error("Expected SliceDiff type")
				}
			},
		},
		{
			"Completely different", []int{1, 2}, []int{3, 4}, 2,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 2 {
					t.Fatalf("Expected 2 diffs, got %d", len(result.Diffs))
				}

				foundIndex0 := false
				foundIndex1 := false

				for _, diff := range result.Diffs {
					if sliceDiff, ok := diff.(*SliceDiff); ok {
						switch sliceDiff.Index {
						case 0:
							if sliceDiff.Left == 1 && sliceDiff.Right == 3 {
								foundIndex0 = true
							}
						case 1:
							if sliceDiff.Left == 2 && sliceDiff.Right == 4 {
								foundIndex1 = true
							}
						}
					}
				}

				if !foundIndex0 {
					t.Error("Missing diff at index 0")
				}
				if !foundIndex1 {
					t.Error("Missing diff at index 1")
				}
			},
		},
		{
			"Empty slices", []int{}, []int{}, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d: %s", tt.expected, len(result.Diffs), result.String())
			}

			tt.validate(t, result)
		})
	}
}

func TestMapComparison(t *testing.T) {
	leftMap := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	rightMap := map[string]int{
		"a": 1,
		"b": 4,
		"d": 5,
	}

	result, err := Compare(leftMap, rightMap)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	expectedDiffs := 3
	if len(result.Diffs) != expectedDiffs {
		t.Errorf("Expected %d differences, got %d: %s", expectedDiffs, len(result.Diffs), result.String())
	}

	foundValueChange := false
	foundKeyRemoved := false
	foundKeyAdded := false

	for _, diff := range result.Diffs {
		switch d := diff.(type) {
		case *MapDiff:
			switch d.ChangeType {
			case "UPDATED":
				if d.Key == "b" && d.Left == 2 && d.Right == 4 {
					foundValueChange = true
				}
			case "REMOVED":
				if d.Key == "c" && d.Left == 3 && d.Right == nil {
					foundKeyRemoved = true
				}
			case "ADDED":
				if d.Key == "d" && d.Left == nil && d.Right == 5 {
					foundKeyAdded = true
				}
			}
		}
	}

	if !foundValueChange {
		t.Error("Missing value change for key 'b'")
	}
	if !foundKeyRemoved {
		t.Error("Missing removal of key 'c'")
	}
	if !foundKeyAdded {
		t.Error("Missing addition of key 'd'")
	}
}

func TestMapWithDifferentValueTypes(t *testing.T) {
	leftMap := map[string]any{
		"a": 1,
		"b": "hello",
	}
	rightMap := map[string]any{
		"a": 1,
		"b": 123,
	}

	result, err := Compare(leftMap, rightMap)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	expectedDiffs := 1
	if len(result.Diffs) != expectedDiffs {
		t.Errorf("Expected %d differences, got %d: %s", expectedDiffs, len(result.Diffs), result.String())
	}

	if len(result.Diffs) > 0 {
		if mapDiff, ok := result.Diffs[0].(*MapDiff); ok {
			if mapDiff.Key != "b" {
				t.Errorf("Expected diff for key 'b', got %v", mapDiff.Key)
			}
			if mapDiff.Left != "hello" {
				t.Errorf("Expected left value 'hello', got %v", mapDiff.Left)
			}
			if mapDiff.Right != 123 {
				t.Errorf("Expected right value 123, got %v", mapDiff.Right)
			}
			if mapDiff.ChangeType != "UPDATED" {
				t.Errorf("Expected UPDATED change type, got %s", mapDiff.ChangeType)
			}
		} else {
			t.Error("Expected MapDiff type for map value change")
		}
	}
}

func TestNonComparableSliceElements(t *testing.T) {
	type NonComparable struct {
		Slice []int
	}
	leftSlice := []NonComparable{{Slice: []int{1}}, {Slice: []int{2}}}
	rightSlice := []NonComparable{{Slice: []int{1}}, {Slice: []int{3}}}

	config := DefaultCompareConfig()
	config.IgnoreSliceOrder = true

	result, err := CompareWithConfig(leftSlice, rightSlice, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}

	expectedDiffs := 2
	if len(result.Diffs) != expectedDiffs {
		t.Errorf("Expected %d differences, got %d: %s", expectedDiffs, len(result.Diffs), result.String())
	}

	foundRemoved := false
	foundAdded := false

	for _, diff := range result.Diffs {
		if d, ok := diff.(*Diff); ok {
			if d.Left != nil && d.Right == nil {
				if removedElem, ok := d.Left.(NonComparable); ok {
					if len(removedElem.Slice) == 1 && removedElem.Slice[0] == 2 {
						foundRemoved = true
					}
				}
			} else if d.Left == nil && d.Right != nil {
				if addedElem, ok := d.Right.(NonComparable); ok {
					if len(addedElem.Slice) == 1 && addedElem.Slice[0] == 3 {
						foundAdded = true
					}
				}
			}
		}
	}

	if !foundRemoved {
		t.Error("Missing removal of element with Slice [2]")
	}
	if !foundAdded {
		t.Error("Missing addition of element with Slice [3]")
	}
}

func TestCompareSlicesByValue2(t *testing.T) {
	type SimpleStruct struct {
		Name  string
		Value int
	}

	left := []SimpleStruct{
		{Name: "A", Value: 1},
		{Name: "B", Value: 2},
	}
	right := []SimpleStruct{
		{Name: "B", Value: 2},
		{Name: "A", Value: 1},
	}

	config := DefaultCompareConfig()
	config.IgnoreSliceOrder = true
	result, err := CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}
	if result.HasDifferences() {
		t.Errorf("Expected no differences when order is ignored, but got: %s", result.String())
	}

	right = []SimpleStruct{
		{Name: "B", Value: 3},
		{Name: "A", Value: 1},
	}
	result, err = CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}
	if !result.HasDifferences() {
		t.Errorf("Expected differences, but got none")
	}
	if result.Count() != 2 {
		t.Errorf("Expected 2 diffs, got %d: %s", result.Count(), result.String())
	}

	left = []SimpleStruct{
		{Name: "A", Value: 1},
		{Name: "B", Value: 2},
	}
	right = []SimpleStruct{
		{Name: "B", Value: 2},
		{Name: "C", Value: 3},
	}
	result, err = CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}
	if result.Count() != 2 {
		t.Errorf("Expected 2 diffs, got %d: %s", result.Count(), result.String())
	}
}

func TestCompareSlicesByValue(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			name:     "non-comparable element type (slices)",
			left:     [][]int{{1}, {2}},
			right:    [][]int{{2}, {1}},
			expected: 0,
			validate: func(t *testing.T, result *DiffResult) {

			},
		},
		{
			name:     "non-comparable element type (maps)",
			left:     []map[string]int{{"a": 1}, {"b": 2}},
			right:    []map[string]int{{"b": 2}, {"a": 1}},
			expected: 0,
			validate: func(t *testing.T, result *DiffResult) {

			},
		},
		{
			name:     "large slices with ignore order",
			left:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			right:    []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			expected: 0,
			validate: func(t *testing.T, result *DiffResult) {

			},
		},
		{
			name:     "large slices with differences",
			left:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			right:    []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expected: 20,
			validate: func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 20 {
					t.Fatalf("Expected 20 diffs, got %d", len(result.Diffs))
				}

				removedCount := 0
				addedCount := 0

				for i, diff := range result.Diffs {
					t.Logf("Diff %d: %+v", i, diff)
					if sliceDiff, ok := diff.(*SliceDiff); ok {
						t.Logf("SliceDiff %d - ChangeType: %s", i, sliceDiff.ChangeType)
						switch sliceDiff.ChangeType {
						case "REMOVED":
							removedCount++
						case "ADDED":
							addedCount++
						}
					} else if d, ok := diff.(*Diff); ok {
						t.Logf("Diff %d - Left: %v, Right: %v", i, d.Left, d.Right)
						if d.Left != nil && d.Right == nil {
							removedCount++
						} else if d.Left == nil && d.Right != nil {
							addedCount++
						}
					}
				}

				if removedCount != 10 {
					t.Errorf("Expected 10 REMOVED diffs, got %d", removedCount)
				}
				if addedCount != 10 {
					t.Errorf("Expected 10 ADDED diffs, got %d", addedCount)
				}
			},
		},
		{
			name:     "slices with duplicate elements",
			left:     []string{"a", "b", "a", "c"},
			right:    []string{"a", "c", "b"},
			expected: 1,
			validate: func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}

				if d, ok := result.Diffs[0].(*Diff); ok {
					if d.Left != "a" {
						t.Errorf("Expected removed value 'a', got %v", d.Left)
					}
					if d.Right != nil {
						t.Errorf("Expected Right value nil for removal, got %v", d.Right)
					}
				} else {
					t.Error("Expected Diff type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CompareConfig{
				IgnoreSliceOrder: true,
			}
			result, err := CompareWithConfig(tt.left, tt.right, config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Diffs) != tt.expected {
				t.Errorf("expected %d differences, got %d", tt.expected, len(result.Diffs))
			}

			tt.validate(t, result)
		})
	}
}

func TestMapDiffChangeTypes(t *testing.T) {
	left := map[string]int{
		"keep":   1,
		"modify": 2,
		"remove": 3,
	}

	right := map[string]int{
		"keep":   1,
		"modify": 20,
		"add":    4,
	}

	result, err := Compare(left, right)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDiffs := 3
	if len(result.Diffs) != expectedDiffs {
		t.Errorf("expected %d differences, got %d", expectedDiffs, len(result.Diffs))
	}

	var hasModify, hasRemove, hasAdd bool
	for _, diff := range result.Diffs {
		if mapDiff, ok := diff.(*MapDiff); ok {
			switch mapDiff.ChangeType {
			case "UPDATED":
				hasModify = true
			case "REMOVED":
				hasRemove = true
			case "ADDED":
				hasAdd = true
			}
		}
	}

	if !hasModify {
		t.Error("expected to find UPDATED change")
	}
	if !hasRemove {
		t.Error("expected to find REMOVED change")
	}
	if !hasAdd {
		t.Error("expected to find ADDED change")
	}
}

func TestCompareSlicesElementByElement(t *testing.T) {
	tests := []struct {
		name     string
		left     []int
		right    []int
		expected int
	}{
		{
			name:     "identical slices",
			left:     []int{1, 2, 3},
			right:    []int{1, 2, 3},
			expected: 0,
		},
		{
			name:     "different elements same length",
			left:     []int{1, 2, 3},
			right:    []int{1, 4, 3},
			expected: 1,
		},
		{
			name:     "left longer than right",
			left:     []int{1, 2, 3, 4, 5},
			right:    []int{1, 2, 3},
			expected: 2,
		},
		{
			name:     "right longer than left",
			left:     []int{1, 2},
			right:    []int{1, 2, 3, 4, 5},
			expected: 3,
		},
		{
			name:     "completely different slices",
			left:     []int{1, 2, 3},
			right:    []int{4, 5, 6},
			expected: 3,
		},
		{
			name:     "empty left slice",
			left:     []int{},
			right:    []int{1, 2, 3},
			expected: 3,
		},
		{
			name:     "empty right slice",
			left:     []int{1, 2, 3},
			right:    []int{},
			expected: 3,
		},
		{
			name:     "both empty slices",
			left:     []int{},
			right:    []int{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &DiffResult{}

			leftVal := reflect.ValueOf(tt.left)
			rightVal := reflect.ValueOf(tt.right)

			err := compareSlicesElementByElement("test", leftVal, rightVal, result)
			if err != nil {
				t.Errorf("compareSlicesElementByElement returned error: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d", tt.expected, len(result.Diffs))
			}
		})
	}
}

func TestNonComparableMapValues(t *testing.T) {
	left := map[string][]int{
		"list1": {1, 2, 3},
		"list2": {4, 5, 6},
	}

	right := map[string][]int{
		"list1": {1, 2, 4},
		"list3": {7, 8, 9},
	}

	result, err := Compare(left, right)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Diffs) == 0 {
		t.Error("expected differences in map with slice values")
	}

	for i, diff := range result.Diffs {
		t.Logf("Diff %d: %+v", i, diff)
		if mapDiff, ok := diff.(*MapDiff); ok {
			t.Logf("MapDiff %d - Key: %s, ChangeType: %s, Left: %v, Right: %v", i, mapDiff.Key, mapDiff.ChangeType, mapDiff.Left, mapDiff.Right)
		}
	}

	foundModified := false
	foundAdded := false
	foundRemoved := false

	for _, diff := range result.Diffs {
		if mapDiff, ok := diff.(*MapDiff); ok {
			if mapDiff.Key == "list2" && mapDiff.ChangeType == "REMOVED" {
				if leftSlice, ok := mapDiff.Left.([]int); ok {
					if len(leftSlice) == 3 && leftSlice[0] == 4 && leftSlice[1] == 5 && leftSlice[2] == 6 {
						foundRemoved = true
					}
				}
			} else if mapDiff.Key == "list3" && mapDiff.ChangeType == "ADDED" {
				if mapDiff.Left == nil {
					if rightSlice, ok := mapDiff.Right.([]int); ok {
						if len(rightSlice) == 3 && rightSlice[0] == 7 && rightSlice[1] == 8 && rightSlice[2] == 9 {
							foundAdded = true
						}
					}
				}
			}
		} else if sliceDiff, ok := diff.(*SliceDiff); ok {
			if sliceDiff.ChangeType == "UPDATED" && sliceDiff.Index == 2 {
				if sliceDiff.Left == 3 && sliceDiff.Right == 4 {
					foundModified = true
				}
			}
		}
	}

	if !foundModified {
		t.Error("Missing modification of slice element at index 2 (3 -> 4)")
	}
	if !foundRemoved {
		t.Error("Missing removal of key 'list2'")
	}
	if !foundAdded {
		t.Error("Missing addition of key 'list3'")
	}
}

func TestCompareSlicesAdvanced(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
	}{
		{
			name:     "nil slices",
			left:     []int(nil),
			right:    []int(nil),
			expected: 0,
		},
		{
			name:     "nil left slice, valid right",
			left:     []int(nil),
			right:    []int{1, 2, 3},
			expected: 3,
		},
		{
			name:     "valid left slice, nil right",
			left:     []int{1, 2, 3},
			right:    []int(nil),
			expected: 3,
		},
		{
			name:     "different slice types",
			left:     []int{1, 2, 3},
			right:    []string{"1", "2", "3"},
			expected: 1,
		},
		{
			name:     "empty slices",
			left:     []int{},
			right:    []int{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &DiffResult{}
			leftVal := reflect.ValueOf(tt.left)
			rightVal := reflect.ValueOf(tt.right)

			err := compareSlicesAdvanced("test", leftVal, rightVal, result, DefaultCompareConfig())
			if err != nil {
				t.Fatalf("compareSlicesAdvanced failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d", tt.expected, len(result.Diffs))
			}
		})
	}
}

func TestCompareMapsEdgeCases(t *testing.T) {
	t.Run("nil maps", func(t *testing.T) {
		result := &DiffResult{}
		leftVal := reflect.ValueOf(map[string]int(nil))
		rightVal := reflect.ValueOf(map[string]int(nil))

		err := compareMaps("test", leftVal, rightVal, result, DefaultCompareConfig())
		if err != nil {
			t.Fatalf("compareMaps failed: %v", err)
		}

		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for nil maps, got %d", len(result.Diffs))
		}
	})

	t.Run("empty maps", func(t *testing.T) {
		result := &DiffResult{}
		leftVal := reflect.ValueOf(map[string]int{})
		rightVal := reflect.ValueOf(map[string]int{})

		err := compareMaps("test", leftVal, rightVal, result, DefaultCompareConfig())
		if err != nil {
			t.Fatalf("compareMaps failed: %v", err)
		}

		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for empty maps, got %d", len(result.Diffs))
		}
	})

	t.Run("map with complex keys", func(t *testing.T) {
		result := &DiffResult{}
		left := map[[2]int]string{{1, 2}: "value1"}
		right := map[[2]int]string{{1, 2}: "value2"}

		leftVal := reflect.ValueOf(left)
		rightVal := reflect.ValueOf(right)

		err := compareMaps("test", leftVal, rightVal, result, DefaultCompareConfig())
		if err != nil {
			t.Fatalf("compareMaps failed: %v", err)
		}

		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference for map with complex keys, got %d", len(result.Diffs))
		}
	})
}
