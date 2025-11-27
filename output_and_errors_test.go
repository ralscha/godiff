package godiff

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type SimpleStruct struct {
	ID   int
	Name string
}

type DeepStruct struct {
	Level1 struct {
		Level2 struct {
			Level3 struct {
				Level4 struct {
					Level5 struct {
						Level6 struct {
							Level7 struct {
								Level8 struct {
									Level9 struct {
										Level10 struct {
											Value string
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestJSONOutput2(t *testing.T) {
	leftData := map[string]any{
		"name":  "Alice",
		"age":   30,
		"items": []int{1, 2, 3},
	}

	rightData := map[string]any{
		"name":  "Bob",
		"age":   25,
		"items": []int{1, 2, 4},
	}

	result, err := Compare(leftData, rightData)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	jsonStr := result.ToJSON()

	var parsed []any
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	if len(parsed) == 0 {
		t.Error("JSON output should contain changes")
	}

	resultNoDiff, err := Compare("same", "same")
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	jsonStr = resultNoDiff.ToJSON()

	var parsedNoDiff []any
	err = json.Unmarshal([]byte(jsonStr), &parsedNoDiff)
	if err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	if len(parsedNoDiff) != 0 {
		t.Error("JSON output should be empty array for no differences")
	}
	if !strings.Contains(jsonStr, "[]") {
		t.Error("JSON output should be empty array for no differences")
	}
}

func TestCustomComparatorErrors(t *testing.T) {
	type ErrorType struct {
		Value string
	}

	errorComparator := func(left, right any, config *CompareConfig) (bool, error) {
		return false, errors.New("custom comparator error")
	}

	comparators := map[reflect.Type]func(left, right any, config *CompareConfig) (bool, error){
		reflect.TypeOf(ErrorType{}): errorComparator,
	}

	left := ErrorType{Value: "test1"}
	right := ErrorType{Value: "test2"}

	result, err := Compare(left, right, WithCustomComparators(comparators))
	if err == nil {
		t.Error("Expected error from custom comparator, got nil")
	}
	if result != nil {
		t.Error("Expected nil result when custom comparator returns error")
	}
}

func TestCircularReferences(t *testing.T) {
	type Node struct {
		Name string
		Next *Node
	}

	t.Run("simple circular reference", func(t *testing.T) {

		node1 := &Node{Name: "first"}
		node2 := &Node{Name: "second"}
		node1.Next = node2
		node2.Next = node1

		node3 := &Node{Name: "first"}
		node4 := &Node{Name: "second"}
		node3.Next = node4
		node4.Next = node3

		result, err := Compare(node1, node3)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if result == nil {
			t.Fatal("Result is nil")
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected 0 diffs, got %d", len(result.Diffs))
		}
	})

	t.Run("circular reference with differences", func(t *testing.T) {

		node1 := &Node{Name: "first"}
		node2 := &Node{Name: "second"}
		node1.Next = node2
		node2.Next = node1

		node3 := &Node{Name: "first"}
		node4 := &Node{Name: "different"}
		node3.Next = node4
		node4.Next = node3

		result, err := Compare(node1, node3)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if result == nil {
			t.Fatal("Result is nil")
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 diff, got %d", len(result.Diffs))
		}
		if len(result.Diffs) > 0 {
			diff, ok := result.Diffs[0].(StructDiff)
			if !ok {
				if diffPtr, ok := result.Diffs[0].(*StructDiff); ok {
					diff = *diffPtr
				} else {
					t.Fatalf("Expected StructDiff type, got %T", result.Diffs[0])
				}
			}
			if diff.Path != "Next.Name" {
				t.Errorf("Expected path 'Next.Name', got '%s'", diff.Path)
			}
			if diff.Left != "second" {
				t.Errorf("Expected left value 'second', got '%v'", diff.Left)
			}
			if diff.Right != "different" {
				t.Errorf("Expected right value 'different', got '%v'", diff.Right)
			}
		}
	})

	t.Run("self-referencing structure", func(t *testing.T) {

		node1 := &Node{Name: "self"}
		node1.Next = node1

		node2 := &Node{Name: "self"}
		node2.Next = node2

		result, err := Compare(node1, node2)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}
		if result == nil {
			t.Fatal("Result is nil")
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected 0 diffs, got %d", len(result.Diffs))
		}
	})
}

type SpecialString string

type SpecialStringHandler struct{}

func (h *SpecialStringHandler) CanHandle(typ reflect.Type) bool {
	return typ == reflect.TypeOf(SpecialString(""))
}

func (h *SpecialStringHandler) Compare(left, right any, path string, result *DiffResult, config *CompareConfig) error {
	leftVal, ok1 := left.(SpecialString)
	rightVal, ok2 := right.(SpecialString)

	if !ok1 || !ok2 {
		return errors.New("SpecialStringHandler received unexpected types")
	}

	if !strings.EqualFold(string(leftVal), string(rightVal)) {
		result.Diffs = append(result.Diffs, &Diff{
			Path:  path,
			Left:  leftVal,
			Right: rightVal,
		})
	}
	return nil
}

func TestTypeHandlerInterface(t *testing.T) {

	handlers := append(DefaultTypeHandlers(), &SpecialStringHandler{})

	left := SpecialString("Hello")
	right := SpecialString("hello")

	result, err := Compare(left, right, WithTypeHandlers(handlers))
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences with custom handler, got %d", len(result.Diffs))
	}

	right = "world"
	result, err = Compare(left, right, WithTypeHandlers(handlers))
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference with custom handler, got %d", len(result.Diffs))
	}
}

type ErrorType struct {
	Value string
}

type ErrorHandler struct{}

func (h *ErrorHandler) CanHandle(typ reflect.Type) bool {
	return typ == reflect.TypeOf(ErrorType{})
}

func (h *ErrorHandler) Compare(left, right any, path string, result *DiffResult, config *CompareConfig) error {
	return errors.New("type handler error")
}

func TestTypeHandlerError(t *testing.T) {

	handlers := append(DefaultTypeHandlers(), &ErrorHandler{})

	left := ErrorType{Value: "test1"}
	right := ErrorType{Value: "test2"}

	result, err := Compare(left, right, WithTypeHandlers(handlers))
	if err == nil {
		t.Error("Expected error from type handler, got nil")
	}
	if result != nil {
		t.Error("Expected nil result when type handler returns error")
	}
}

func TestEmptyContainers(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
	}{
		{"Empty vs nil slice", []int{}, []int(nil), 0},
		{"Empty vs nil map", map[string]int{}, map[string]int(nil), 0},
		{"Empty slice vs non-empty", []int{}, []int{1}, 1},
		{"Empty map vs non-empty", map[string]int{}, map[string]int{"key": 1}, 1},
		{"Empty arrays", [0]int{}, [0]int{}, 0},
		{"Empty string slice", []string{}, []string{}, 0},
		{"Empty nested map", map[string]map[string]int{}, map[string]map[string]int{}, 0},
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
		})
	}
}

func TestTypeConversionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
	}{
		{"Different interface types", any("string"), any(123), 1},
		{"Interface vs concrete", any("hello"), "hello", 0},
		{"Nil interface vs nil", any(nil), nil, 0},
		{"Nil interface vs value", any(nil), "value", 1},
		{"Different pointer types", (*int)(nil), (*string)(nil), 0},
		{"Typed nil vs untyped nil", (*int)(nil), nil, 1},
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
		})
	}
}

func TestInvalidDiffTags(t *testing.T) {
	type TestStruct struct {
		Field1 string `diff:"invalid"`
		Field2 int    `diff:"ignore,invalid"`
		Field3 bool   `diff:""`
		Field4 string `diff:"ignoreOrder,invalid"`
	}

	left := TestStruct{
		Field1: "value1",
		Field2: 42,
		Field3: true,
		Field4: "value4",
	}

	right := TestStruct{
		Field1: "value1_changed",
		Field2: 43,
		Field3: false,
		Field4: "value4_changed",
	}

	result, err := Compare(left, right)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	expectedDiffs := 3
	if len(result.Diffs) != expectedDiffs {
		t.Errorf("Expected %d differences, got %d: %s", expectedDiffs, len(result.Diffs), result.String())
	}
}

func TestStringOutputEdgeCases(t *testing.T) {

	left := DeepStruct{}
	left.Level1.Level2.Level3.Level4.Level5.Level6.Level7.Level8.Level9.Level10.Value = "left"

	right := DeepStruct{}
	right.Level1.Level2.Level3.Level4.Level5.Level6.Level7.Level8.Level9.Level10.Value = "right"

	result, err := Compare(left, right)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	output := result.String()
	if len(output) == 0 {
		t.Error("String output should not be empty")
	}

	expectedPath := "Level1.Level2.Level3.Level4.Level5.Level6.Level7.Level8.Level9.Level10.Value"
	if !strings.Contains(output, expectedPath) {
		t.Errorf("String output should contain path %s", expectedPath)
	}
}

func TestStringOutput(t *testing.T) {
	tests := []struct {
		name  string
		left  any
		right any
	}{
		{
			name:  "complex nested structure",
			left:  map[string][]SimpleStruct{"items": {{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}},
			right: map[string][]SimpleStruct{"items": {{ID: 1, Name: "A-modified"}, {ID: 3, Name: "C"}}},
		},
		{
			name:  "slice differences",
			left:  []string{"a", "b", "c"},
			right: []string{"a", "d", "c"},
		},
		{
			name:  "map differences",
			left:  map[string]string{"key1": "value1", "key2": "value2"},
			right: map[string]string{"key1": "value1-modified", "key3": "value3"},
		},
		{
			name:  "empty diff result",
			left:  "same",
			right: "same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := result.String()
			if len(output) == 0 {
				t.Error("expected non-empty string output")
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	tests := []struct {
		name  string
		left  any
		right any
	}{
		{
			name:  "valid JSON output with differences",
			left:  SimpleStruct{ID: 1, Name: "A"},
			right: SimpleStruct{ID: 1, Name: "B"},
		},
		{
			name:  "complex structure JSON",
			left:  map[string]any{"nested": []int{1, 2, 3}},
			right: map[string]any{"nested": []int{1, 2, 4}},
		},
		{
			name:  "no differences JSON",
			left:  "same",
			right: "same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			jsonOutput := result.ToJSON()
			if len(jsonOutput) == 0 {
				t.Error("expected non-empty JSON output")
			}

			var parsed []any
			err = json.Unmarshal([]byte(jsonOutput), &parsed)
			if err != nil {
				t.Fatalf("Generated JSON is not a valid array: %v", err)
			}
		})
	}
}

func TestStringOutputAdditionalEdgeCases(t *testing.T) {
	t.Run("empty diff result", func(t *testing.T) {
		result := &DiffResult{}
		output := result.String()
		if output != "No differences found" {
			t.Errorf("Expected 'No differences found', got '%s'", output)
		}
	})

	t.Run("unknown diff type", func(t *testing.T) {
		result := &DiffResult{
			Diffs: []any{"unknown type"},
		}
		output := result.String()
		if !strings.Contains(output, "Unknown diff type") {
			t.Errorf("Expected 'Unknown diff type' in output, got '%s'", output)
		}
	})

	t.Run("struct diff with empty field name", func(t *testing.T) {
		result := &DiffResult{
			Diffs: []any{
				&StructDiff{
					Diff: Diff{
						Path:  "test",
						Left:  "left",
						Right: "right",
					},
					FieldName:  "",
					ChangeType: ChangeTypeUpdated,
				},
			},
		}
		output := result.String()
		if !strings.Contains(output, "UPDATED") {
			t.Errorf("Expected 'UPDATED' in output, got '%s'", output)
		}
	})

	t.Run("map diff with complex key", func(t *testing.T) {
		result := &DiffResult{
			Diffs: []any{
				&MapDiff{
					Diff: Diff{
						Path:  "test",
						Left:  "left",
						Right: "right",
					},
					Key:        map[string]int{"complex": 1},
					ChangeType: ChangeTypeUpdated,
				},
			},
		}
		output := result.String()
		if !strings.Contains(output, "UPDATED test") {
			t.Errorf("Expected 'UPDATED test' in output, got '%s'", output)
		}
	})
}

func TestToJSONEdgeCases(t *testing.T) {
	t.Run("empty diff result", func(t *testing.T) {
		result := &DiffResult{}
		jsonOutput := result.ToJSON()
		if jsonOutput != "[]" {
			t.Errorf("Expected empty array '[]', got '%s'", jsonOutput)
		}
	})

	t.Run("unknown diff type", func(t *testing.T) {
		result := &DiffResult{
			Diffs: []any{"unknown type"},
		}
		jsonOutput := result.ToJSON()
		if !strings.Contains(jsonOutput, "unknown") {
			t.Errorf("Expected 'unknown' in JSON output, got '%s'", jsonOutput)
		}
	})

	t.Run("JSON marshaling error simulation", func(t *testing.T) {
		result := &DiffResult{
			Diffs: []any{
				&Diff{
					Path:  "test",
					Left:  func() {},
					Right: "right",
				},
			},
		}
		jsonOutput := result.ToJSON()
		if !strings.Contains(jsonOutput, "error") {
			t.Errorf("Expected error in JSON output, got '%s'", jsonOutput)
		}
	})
}
