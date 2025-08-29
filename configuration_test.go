package godiff

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type TestStruct struct {
	ID      int
	Items   []int `diff:"ignoreOrder"`
	Names   []string
	Numbers []int
}

type IDStruct struct {
	ID   int `diff:"id"`
	Name string
}

func TestIgnoreOrderTag(t *testing.T) {
	leftStruct := TestStruct{
		ID:      1,
		Items:   []int{1, 2, 3},
		Names:   []string{"a", "b", "c"},
		Numbers: []int{10, 20, 30},
	}

	rightStruct := TestStruct{
		ID:      1,
		Items:   []int{3, 1, 2},
		Names:   []string{"c", "b", "a"},
		Numbers: []int{10, 20, 30, 40},
	}

	result, err := Compare(leftStruct, rightStruct)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	expectedDiffs := 3
	if len(result.Diffs) != expectedDiffs {
		t.Errorf("Expected %d differences, got %d: %s", expectedDiffs, len(result.Diffs), result.String())
	}

	foundNameDiff1 := false
	foundNameDiff2 := false
	foundNumberDiff := false

	for _, diff := range result.Diffs {
		switch d := diff.(type) {
		case *SliceDiff:
			if d.Path == "Names" && d.Index == 0 && d.Left == "a" && d.Right == "c" {
				foundNameDiff1 = true
			} else if d.Path == "Names" && d.Index == 2 && d.Left == "c" && d.Right == "a" {
				foundNameDiff2 = true
			} else if d.Path == "Numbers" && d.Index == 3 && d.Left == nil && d.Right == 40 {
				foundNumberDiff = true
			}
		}
	}

	if !foundNameDiff1 {
		t.Error("Missing name difference at index 0")
	}
	if !foundNameDiff2 {
		t.Error("Missing name difference at index 2")
	}
	if !foundNumberDiff {
		t.Error("Missing number addition difference")
	}
}

func TestIgnoreOrderWithDuplicates(t *testing.T) {
	type DuplicateStruct struct {
		Values []int `diff:"ignoreOrder"`
	}

	leftStruct := DuplicateStruct{
		Values: []int{1, 2, 2, 3},
	}

	rightStruct := DuplicateStruct{
		Values: []int{3, 2, 1, 2},
	}

	result, err := Compare(leftStruct, rightStruct)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences with duplicates and ignoreOrder, got %d: %s", len(result.Diffs), result.String())
		for i, diff := range result.Diffs {
			t.Errorf("Unexpected diff %d: %+v", i, diff)
		}
	}
}

func TestIgnoreOrderDifferentContent(t *testing.T) {
	type TestStruct struct {
		Items []int `diff:"ignoreOrder"`
	}

	leftStruct := TestStruct{
		Items: []int{1, 2, 3},
	}

	rightStruct := TestStruct{
		Items: []int{1, 2, 4},
	}

	result, err := Compare(leftStruct, rightStruct)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
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
				if d.Left == 3 {
					foundRemoved = true
				}
			} else if d.Left == nil && d.Right != nil {
				if d.Right == 4 {
					foundAdded = true
				}
			}
		}
	}

	if !foundRemoved {
		t.Error("Missing removal of value 3")
	}
	if !foundAdded {
		t.Error("Missing addition of value 4")
	}
}

func TestCompareWithConfig(t *testing.T) {
	type ConfigTest struct {
		A string
		B int
		C bool
	}

	leftStruct := ConfigTest{A: "hello", B: 42, C: true}
	rightStruct := ConfigTest{A: "world", B: 43, C: false}

	t.Run("Ignore one field", func(t *testing.T) {
		config := DefaultCompareConfig()
		config.IgnoreFields = []string{"B"}
		result, err := CompareWithConfig(leftStruct, rightStruct, config)
		if err != nil {
			t.Fatalf("CompareWithConfig failed: %v", err)
		}
		if len(result.Diffs) != 2 {
			t.Errorf("Expected 2 differences, got %d: %s", len(result.Diffs), result.String())
		}

		foundA := false
		foundC := false

		for _, diff := range result.Diffs {
			if structDiff, ok := diff.(*StructDiff); ok {
				if structDiff.Path == "A" && structDiff.Left == "hello" && structDiff.Right == "world" {
					foundA = true
				} else if structDiff.Path == "C" && structDiff.Left == true && structDiff.Right == false {
					foundC = true
				}
			}
		}

		if !foundA {
			t.Error("Missing difference in field A")
		}
		if !foundC {
			t.Error("Missing difference in field C")
		}
	})

	t.Run("Ignore multiple fields", func(t *testing.T) {
		config := DefaultCompareConfig()
		config.IgnoreFields = []string{"A", "C"}
		result, err := CompareWithConfig(leftStruct, rightStruct, config)
		if err != nil {
			t.Fatalf("CompareWithConfig failed: %v", err)
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())
		}

		if len(result.Diffs) > 0 {
			if structDiff, ok := result.Diffs[0].(*StructDiff); ok {
				if structDiff.Path != "B" {
					t.Errorf("Expected difference in field B, but got %s", structDiff.Path)
				}
				if structDiff.Left != 42 || structDiff.Right != 43 {
					t.Errorf("Expected B field values 42 -> 43, got %v -> %v", structDiff.Left, structDiff.Right)
				}
			} else {
				t.Error("Expected a StructDiff")
			}
		}
	})
}

func TestIgnoreTag(t *testing.T) {
	type IgnoreTest struct {
		A string
		B int `diff:"ignore"`
	}

	leftStruct := IgnoreTest{A: "hello", B: 42}
	rightStruct := IgnoreTest{A: "world", B: 43}

	result, err := Compare(leftStruct, rightStruct)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())
	}
	if diff, ok := result.Diffs[0].(*StructDiff); ok {
		if diff.Path != "A" {
			t.Errorf("Expected difference in field A, but got %s", diff.Path)
		}
	} else {
		t.Error("Expected a StructDiff")
	}
}

func TestCustomComparator(t *testing.T) {
	type CustomType struct {
		Value string
	}

	customComparator := func(left, right any, config *CompareConfig) (bool, error) {
		leftVal, ok1 := left.(CustomType)
		rightVal, ok2 := right.(CustomType)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("custom comparator received unexpected types")
		}

		if len(leftVal.Value) < 3 || len(rightVal.Value) < 3 {
			return leftVal.Value == rightVal.Value, nil
		}
		return leftVal.Value[:3] == rightVal.Value[:3], nil
	}

	config := DefaultCompareConfig()
	config.CustomComparators = map[reflect.Type]func(left, right any, config *CompareConfig) (bool, error){
		reflect.TypeOf(CustomType{}): customComparator,
	}

	t.Run("Custom comparator finds no difference when first 3 chars match", func(t *testing.T) {
		leftVal := CustomType{Value: "hello"}
		rightVal := CustomType{Value: "help me"}
		result, err := CompareWithConfig(leftVal, rightVal, config)
		if err != nil {
			t.Fatalf("CompareWithConfig failed: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences, got %d: %s", len(result.Diffs), result.String())

			for i, diff := range result.Diffs {
				t.Errorf("Unexpected diff %d: %+v", i, diff)
			}
		}
	})

	t.Run("Custom comparator finds difference when first 3 chars differ", func(t *testing.T) {
		leftVal := CustomType{Value: "hello"}
		rightVal := CustomType{Value: "world"}
		result, err := CompareWithConfig(leftVal, rightVal, config)
		if err != nil {
			t.Fatalf("CompareWithConfig failed: %v", err)
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())

			for i, diff := range result.Diffs {
				t.Errorf("Diff %d: %+v", i, diff)
			}
		} else {

			if diff, ok := result.Diffs[0].(*Diff); ok {
				if diff.Left != leftVal || diff.Right != rightVal {
					t.Errorf("Expected diff values %+v -> %+v, got %+v -> %+v", leftVal, rightVal, diff.Left, diff.Right)
				}
			} else {
				t.Errorf("Expected Diff type, got %T", result.Diffs[0])
			}
		}
	})

	t.Run("Custom comparator handles short strings", func(t *testing.T) {
		leftVal := CustomType{Value: "hi"}
		rightVal := CustomType{Value: "hi"}
		result, err := CompareWithConfig(leftVal, rightVal, config)
		if err != nil {
			t.Fatalf("CompareWithConfig failed: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for short equal strings, got %d: %s", len(result.Diffs), result.String())

			for i, diff := range result.Diffs {
				t.Errorf("Unexpected diff %d: %+v", i, diff)
			}
		}
	})
}

func TestTimeHandler(t *testing.T) {
	leftTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	rightTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	result, err := Compare(leftTime, rightTime)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences for equal times, got %d: %s", len(result.Diffs), result.String())

		for i, diff := range result.Diffs {
			t.Errorf("Unexpected diff %d: %+v", i, diff)
		}
	}

	rightTime = time.Date(2023, 1, 1, 12, 0, 1, 0, time.UTC)
	result, err = Compare(leftTime, rightTime)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference for different times, got %d: %s", len(result.Diffs), result.String())

		for i, diff := range result.Diffs {
			t.Errorf("Diff %d: %+v", i, diff)
		}
	} else {

		if diff, ok := result.Diffs[0].(*Diff); ok {
			if diff.Left != leftTime || diff.Right != rightTime {
				t.Errorf("Expected diff values %+v -> %+v, got %+v -> %+v", leftTime, rightTime, diff.Left, diff.Right)
			}
		} else {
			t.Errorf("Expected Diff type, got %T", result.Diffs[0])
		}
	}
}

func TestInterfaceHandler(t *testing.T) {
	var leftInterface, rightInterface any
	leftInterface = "hello"
	rightInterface = "hello"

	result, err := Compare(leftInterface, rightInterface)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences for equal interface values, got %d: %s", len(result.Diffs), result.String())

		for i, diff := range result.Diffs {
			t.Errorf("Unexpected diff %d: %+v", i, diff)
		}
	}

	rightInterface = "world"
	result, err = Compare(leftInterface, rightInterface)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference for different interface values, got %d: %s", len(result.Diffs), result.String())

		for i, diff := range result.Diffs {
			t.Errorf("Diff %d: %+v", i, diff)
		}
	} else {

		if diff, ok := result.Diffs[0].(*Diff); ok {
			if diff.Left != "hello" || diff.Right != "world" {
				t.Errorf("Expected diff values %+v -> %+v, got %+v -> %+v", "hello", "world", diff.Left, diff.Right)
			}
		} else {
			t.Errorf("Expected Diff type, got %T", result.Diffs[0])
		}
	}
}

func TestFunctionHandler(t *testing.T) {
	func1 := func() {}
	func2 := func() {}

	result, err := Compare(func1, func1)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences for same function, got %d: %s", len(result.Diffs), result.String())

		for i, diff := range result.Diffs {
			t.Errorf("Unexpected diff %d: %+v", i, diff)
		}
	}

	result, err = Compare(func1, func2)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}
	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference for different functions, got %d: %s", len(result.Diffs), result.String())

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
}

func TestDefaultTypeHandlers(t *testing.T) {
	leftTime := time.Now()
	rightTime := leftTime.Add(time.Second)

	config := DefaultCompareConfig()
	config.TypeHandlers = DefaultTypeHandlers()

	result, err := CompareWithConfig(leftTime, rightTime, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}
	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference for different times, got %d: %s", len(result.Diffs), result.String())

		for i, diff := range result.Diffs {
			t.Errorf("Diff %d: %+v", i, diff)
		}
	} else {

		if diff, ok := result.Diffs[0].(*Diff); ok {
			if diff.Left != leftTime || diff.Right != rightTime {
				t.Errorf("Expected diff values %+v -> %+v, got %+v -> %+v", leftTime, rightTime, diff.Left, diff.Right)
			}
		} else {
			t.Errorf("Expected Diff type, got %T", result.Diffs[0])
		}
	}
}

func TestCompareWithConfigNilConfig(t *testing.T) {
	result, err := CompareWithConfig("hello", "world", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Diffs) != 1 {
		t.Errorf("expected 1 difference, got %d", len(result.Diffs))
	}

	if len(result.Diffs) > 0 {
		if diff, ok := result.Diffs[0].(*Diff); ok {
			if diff.Left != "hello" {
				t.Errorf("Expected left value 'hello', got %v", diff.Left)
			}
			if diff.Right != "world" {
				t.Errorf("Expected right value 'world', got %v", diff.Right)
			}
		} else {
			t.Errorf("Expected Diff type, got %T", result.Diffs[0])
		}
	}
}

func TestIsFieldIgnoredEdgeCases(t *testing.T) {
	type TestStruct struct {
		Field string
	}

	config := &CompareConfig{
		IgnoreFields: []string{"Field", "TestStruct.Field", "OtherField"},
	}
	typ := reflect.TypeOf(TestStruct{})

	tests := []struct {
		name      string
		fieldPath string
		fieldName string
		expected  bool
	}{
		{"exact field path match", "Field", "Field", true},
		{"type-qualified match", "TestStruct.Field", "Field", true},
		{"simple field name match", "OtherField", "OtherField", true},
		{"no match", "NonExistent", "NonExistent", false},
		{"empty ignore fields", "Field", "Field", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := config
			if tt.name == "empty ignore fields" {
				testConfig = &CompareConfig{IgnoreFields: []string{}}
			}

			ignored := isFieldIgnored(tt.fieldPath, tt.fieldName, typ, testConfig)
			if ignored != tt.expected {
				t.Errorf("Expected %v, got %v for %s", tt.expected, ignored, tt.fieldPath)
			}
		})
	}
}

func TestGetObjectIDAdditionalEdgeCases(t *testing.T) {
	type StructWithPrivateID struct {
		id   int
		Name string
	}

	type StructWithZeroID struct {
		ID   int `diff:"id"`
		Name string
	}

	tests := []struct {
		name     string
		obj      any
		config   *CompareConfig
		expected any
		hasID    bool
	}{
		{
			name:     "nil object",
			obj:      nil,
			config:   DefaultCompareConfig(),
			expected: nil,
			hasID:    false,
		},
		{
			name:     "private ID field",
			obj:      StructWithPrivateID{id: 123, Name: "test"},
			config:   &CompareConfig{IDFieldNames: []string{"id"}},
			expected: nil,
			hasID:    false,
		},
		{
			name:     "zero ID value",
			obj:      StructWithZeroID{ID: 0, Name: "test"},
			config:   DefaultCompareConfig(),
			expected: nil,
			hasID:    false,
		},
		{
			name:     "valid ID from config",
			obj:      StructWithZeroID{ID: 123, Name: "test"},
			config:   &CompareConfig{IDFieldNames: []string{"ID"}},
			expected: 123,
			hasID:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, hasID := getObjectID(tt.obj, tt.config)
			if hasID != tt.hasID {
				t.Errorf("Expected hasID=%v, got %v", tt.hasID, hasID)
			}
			if !reflect.DeepEqual(id, tt.expected) {
				t.Errorf("Expected id=%v, got %v", tt.expected, id)
			}
		})
	}
}

func TestIDComparison(t *testing.T) {
	type StructWithID struct {
		ID   int `diff:"id"`
		Name string
	}

	left := StructWithID{ID: 1, Name: "Alice"}
	right := StructWithID{ID: 2, Name: "Alice"}

	result, err := Compare(left, right)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) != 1 {
		t.Fatalf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())
	}

	diff, ok := result.Diffs[0].(*StructDiff)
	if !ok {
		t.Fatalf("Expected a StructDiff, got %T", result.Diffs[0])
	}

	if diff.ChangeType != ChangeTypeIDMismatch {
		t.Errorf("Expected change type %s, got %s", ChangeTypeIDMismatch, diff.ChangeType)
	}
}
