package godiff

import (
	"reflect"
	"testing"
)

type Person struct {
	Name    string
	Age     int
	Address *Address
	Hobbies []string
}

type Address struct {
	Street  string
	City    string
	Country string
}

func TestBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same strings", "hello", "hello", 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different strings", "hello", "world", 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != "hello" {
						t.Errorf("Expected left value 'hello', got %v", diff.Left)
					}
					if diff.Right != "world" {
						t.Errorf("Expected right value 'world', got %v", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for string comparison")
				}
			},
		},
		{
			"Same integers", 42, 42, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different integers", 42, 43, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != 42 {
						t.Errorf("Expected left value 42, got %v", diff.Left)
					}
					if diff.Right != 43 {
						t.Errorf("Expected right value 43, got %v", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for integer comparison")
				}
			},
		},
		{
			"Same booleans", true, true, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different booleans", true, false, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != true {
						t.Errorf("Expected left value true, got %v", diff.Left)
					}
					if diff.Right != false {
						t.Errorf("Expected right value false, got %v", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for boolean comparison")
				}
			},
		},
		{
			"String to int", "hello", 42, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != "hello" {
						t.Errorf("Expected left value 'hello', got %v", diff.Left)
					}
					if diff.Right != 42 {
						t.Errorf("Expected right value 42, got %v", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for type mismatch comparison")
				}
			},
		},
		{
			"Nil to value", nil, "test", 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != nil {
						t.Errorf("Expected left value nil, got %v", diff.Left)
					}
					if diff.Right != "test" {
						t.Errorf("Expected right value 'test', got %v", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for nil comparison")
				}
			},
		},
		{
			"Value to nil", "test", nil, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != "test" {
						t.Errorf("Expected left value 'test', got %v", diff.Left)
					}
					if diff.Right != nil {
						t.Errorf("Expected right value nil, got %v", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for nil comparison")
				}
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

func TestStructComparison(t *testing.T) {
	leftPerson := Person{
		Name: "Alice",
		Age:  30,
		Address: &Address{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
		},
		Hobbies: []string{"reading", "swimming"},
	}

	rightPerson := Person{
		Name: "Alice",
		Age:  31,
		Address: &Address{
			Street:  "123 Main St",
			City:    "Boston",
			Country: "USA",
		},
		Hobbies: []string{"reading", "hiking"},
	}

	result, err := Compare(leftPerson, rightPerson)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	expectedDiffs := 3
	if len(result.Diffs) != expectedDiffs {
		t.Errorf("Expected %d differences, got %d: %s", expectedDiffs, len(result.Diffs), result.String())
	}

	foundAgeDiff := false
	foundCityDiff := false
	foundHobbyDiff := false

	for _, diff := range result.Diffs {
		switch d := diff.(type) {
		case *StructDiff:
			if d.Path == "Age" && d.FieldName == "Age" {
				if d.Left == 30 && d.Right == 31 {
					foundAgeDiff = true
				}
			} else if d.Path == "Address.City" && d.FieldName == "City" {
				if d.Left == "New York" && d.Right == "Boston" {
					foundCityDiff = true
				}
			}
		case *SliceDiff:
			if d.Path == "Hobbies" && d.Index == 1 {
				if d.Left == "swimming" && d.Right == "hiking" {
					foundHobbyDiff = true
				}
			}
		}
	}

	if !foundAgeDiff {
		t.Error("Missing age difference")
	}
	if !foundCityDiff {
		t.Error("Missing city difference")
	}
	if !foundHobbyDiff {
		t.Error("Missing hobby difference")
	}
}

func TestPointerComparison(t *testing.T) {
	addr1 := &Address{Street: "123 Main St", City: "New York"}
	addr2 := &Address{Street: "456 Oak Ave", City: "Boston"}

	tests := []struct {
		name     string
		left     any
		right    any
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same pointer", addr1, addr1, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different pointers same content", addr1, &Address{Street: "123 Main St", City: "New York"}, 0,
			func(t *testing.T, result *DiffResult) {
			},
		},
		{
			"Different pointers different content", addr1, addr2, 2,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 2 {
					t.Fatalf("Expected 2 diffs, got %d", len(result.Diffs))
				}

				foundStreetDiff := false
				foundCityDiff := false

				for _, diff := range result.Diffs {
					if structDiff, ok := diff.(*StructDiff); ok {
						if structDiff.Path == "Street" && structDiff.FieldName == "Street" {
							if structDiff.Left == "123 Main St" && structDiff.Right == "456 Oak Ave" {
								foundStreetDiff = true
							}
						} else if structDiff.Path == "City" && structDiff.FieldName == "City" {
							if structDiff.Left == "New York" && structDiff.Right == "Boston" {
								foundCityDiff = true
							}
						}
					}
				}

				if !foundStreetDiff {
					t.Error("Missing street difference")
				}
				if !foundCityDiff {
					t.Error("Missing city difference")
				}
			},
		},
		{
			"Nil to pointer", nil, addr1, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != nil {
						t.Errorf("Expected left value nil, got %v", diff.Left)
					}
					if diff.Right != addr1 {
						t.Errorf("Expected right value %v, got %v", addr1, diff.Right)
					}
				} else {
					t.Error("Expected Diff type for nil pointer comparison")
				}
			},
		},
		{
			"Pointer to nil", addr1, nil, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != addr1 {
						t.Errorf("Expected left value %v, got %v", addr1, diff.Left)
					}
					if diff.Right != nil {
						t.Errorf("Expected right value nil, got %v", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for pointer to nil comparison")
				}
			},
		},
		{
			"Both nil", nil, nil, 0,
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

func TestNestedStructures(t *testing.T) {
	leftData := map[string]any{
		"users": []Person{
			{
				Name: "Alice",
				Age:  25,
				Address: &Address{
					Street:  "123 Main St",
					City:    "New York",
					Country: "USA",
				},
			},
		},
	}

	rightData := map[string]any{
		"users": []Person{
			{
				Name: "Alice",
				Age:  26,
				Address: &Address{
					Street:  "123 Main St",
					City:    "Boston",
					Country: "USA",
				},
			},
			{
				Name: "Bob",
				Age:  30,
				Address: &Address{
					Street:  "456 Oak Ave",
					City:    "Chicago",
					Country: "USA",
				},
			},
		},
	}

	result, err := Compare(leftData, rightData)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) < 3 {
		t.Errorf("Expected at least 3 differences, got %d: %s", len(result.Diffs), result.String())
	}

	foundAgeDiff := false
	foundCityDiff := false
	foundNewUserDiff := false

	for _, diff := range result.Diffs {
		switch d := diff.(type) {
		case *StructDiff:
			if d.Path == "[users][0].Age" && d.FieldName == "Age" {
				if d.Left == 25 && d.Right == 26 {
					foundAgeDiff = true
				}
			} else if d.Path == "[users][0].Address.City" && d.FieldName == "City" {
				if d.Left == "New York" && d.Right == "Boston" {
					foundCityDiff = true
				}
			}
		case *SliceDiff:
			if d.Path == "[users]" && d.ChangeType == "ADDED" && d.Index == 1 {
				if user, ok := d.Right.(Person); ok {
					if user.Name == "Bob" && user.Age == 30 {
						foundNewUserDiff = true
					}
				}
			}
		}
	}

	if !foundAgeDiff {
		t.Error("Missing age difference for Alice")
	}
	if !foundCityDiff {
		t.Error("Missing city difference for Alice")
	}
	if !foundNewUserDiff {
		t.Error("Missing new user (Bob) addition")
	}
}

func TestNoDifferences(t *testing.T) {
	person := Person{
		Name: "Test",
		Age:  30,
		Address: &Address{
			Street:  "Test St",
			City:    "Test City",
			Country: "Test Country",
		},
		Hobbies: []string{"test1", "test2"},
	}

	result, err := Compare(person, person)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences, got %d: %s", len(result.Diffs), result.String())
	}

	if result.String() != "No differences found" {
		t.Errorf("Expected 'No differences found', got: %s", result.String())
	}
}

type MyStruct struct {
	ID    int
	Value string
	Meta  string
}

type AnotherStruct struct {
	Name string
	Meta string
	Info string
}

func TestIgnoreFieldsSimpleName(t *testing.T) {
	left := MyStruct{ID: 1, Value: "old", Meta: "ignored"}
	right := MyStruct{ID: 1, Value: "new", Meta: "also ignored"}

	config := DefaultCompareConfig()
	config.IgnoreFields = []string{"Meta"}

	result, err := CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}

	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())
	}

	if len(result.Diffs) > 0 {
		if structDiff, ok := result.Diffs[0].(*StructDiff); ok {
			if structDiff.Path != "Value" {
				t.Errorf("Expected difference in field Value, but got %s", structDiff.Path)
			}
		}
	}
}

func TestIgnoreFieldsFullPath(t *testing.T) {
	left := MyStruct{ID: 1, Value: "old", Meta: "ignored"}
	right := MyStruct{ID: 1, Value: "new", Meta: "also ignored"}

	config := DefaultCompareConfig()
	config.IgnoreFields = []string{"MyStruct.Meta"}

	result, err := CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}

	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())
	}

	if len(result.Diffs) > 0 {
		if structDiff, ok := result.Diffs[0].(*StructDiff); ok {
			if structDiff.Path != "Value" {
				t.Errorf("Expected difference in field Value, but got %s", structDiff.Path)
			}
		}
	}
}

func TestIgnoreFieldsSelectiveByType(t *testing.T) {
	type NestedStruct struct {
		Left  MyStruct
		Right AnotherStruct
	}

	left := NestedStruct{
		Left:  MyStruct{ID: 1, Value: "old", Meta: "ignored"},
		Right: AnotherStruct{Name: "test", Meta: "should see", Info: "info"},
	}
	right := NestedStruct{
		Left:  MyStruct{ID: 1, Value: "new", Meta: "also ignored"},
		Right: AnotherStruct{Name: "test", Meta: "should see change", Info: "info"},
	}

	config := DefaultCompareConfig()
	config.IgnoreFields = []string{"MyStruct.Meta"}

	result, err := CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}

	if len(result.Diffs) != 2 {
		t.Errorf("Expected 2 differences, got %d: %s", len(result.Diffs), result.String())
	}

	foundValueDiff := false
	foundAnotherMetaDiff := false
	for _, diff := range result.Diffs {
		if structDiff, ok := diff.(*StructDiff); ok {
			if structDiff.Path == "Left.Value" {
				foundValueDiff = true
			}
			if structDiff.Path == "Right.Meta" {
				foundAnotherMetaDiff = true
			}
		}
	}

	if !foundValueDiff {
		t.Error("Expected to find difference in Left.Value")
	}
	if !foundAnotherMetaDiff {
		t.Error("Expected to find difference in Right.Meta (should not be ignored)")
	}
}

func TestIgnoreFieldsMultiplePatterns(t *testing.T) {
	type TestStruct struct {
		ID    int
		Name  string
		Value string
		Meta  string
	}

	left := TestStruct{ID: 1, Name: "old", Value: "old", Meta: "ignored"}
	right := TestStruct{ID: 2, Name: "new", Value: "new", Meta: "also ignored"}

	config := DefaultCompareConfig()
	config.IgnoreFields = []string{"TestStruct.Meta", "Name"}

	result, err := CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}

	if len(result.Diffs) != 2 {
		t.Errorf("Expected 2 differences, got %d: %s", len(result.Diffs), result.String())
	}

	foundIDDiff := false
	foundValueDiff := false
	for _, diff := range result.Diffs {
		if structDiff, ok := diff.(*StructDiff); ok {
			switch structDiff.Path {
			case "ID":
				foundIDDiff = true
			case "Value":
				foundValueDiff = true
			}
		}
	}

	if !foundIDDiff {
		t.Error("Expected to find difference in ID")
	}
	if !foundValueDiff {
		t.Error("Expected to find difference in Value")
	}
}

func TestIgnoreFieldsNestedStructs(t *testing.T) {
	type Inner struct {
		Meta string
		Data string
	}
	type Outer struct {
		Name  string
		Inner Inner
	}

	left := Outer{
		Name:  "test",
		Inner: Inner{Meta: "ignored", Data: "old"},
	}
	right := Outer{
		Name:  "test",
		Inner: Inner{Meta: "also ignored", Data: "new"},
	}

	config := DefaultCompareConfig()
	config.IgnoreFields = []string{"Inner.Meta"}

	result, err := CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}

	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())
	}

	if len(result.Diffs) > 0 {
		if structDiff, ok := result.Diffs[0].(*StructDiff); ok {
			if structDiff.Path != "Inner.Data" {
				t.Errorf("Expected difference in field Inner.Data, but got %s", structDiff.Path)
			}
		}
	}
}

func TestUserExampleScenario(t *testing.T) {
	left := MyStruct{ID: 1, Value: "old", Meta: "ignored"}
	right := MyStruct{ID: 1, Value: "new", Meta: "also ignored"}

	config := DefaultCompareConfig()
	config.IgnoreFields = []string{"Meta"}

	result, err := CompareWithConfig(left, right, config)
	if err != nil {
		t.Fatalf("CompareWithConfig failed: %v", err)
	}

	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())
	}

	if len(result.Diffs) > 0 {
		if structDiff, ok := result.Diffs[0].(*StructDiff); ok {
			if structDiff.Path != "Value" {
				t.Errorf("Expected difference in field Value, but got %s", structDiff.Path)
			}
			if structDiff.Left != "old" || structDiff.Right != "new" {
				t.Errorf("Expected Value change from 'old' to 'new', got %v -> %v", structDiff.Left, structDiff.Right)
			}
		} else {
			t.Error("Expected a StructDiff")
		}
	}
}

func TestCompareValuesEdgeCases(t *testing.T) {
	t.Run("ignore field path", func(t *testing.T) {
		result := &DiffResult{}
		config := DefaultCompareConfig()
		config.IgnoreFields = []string{"test"}
		err := compareValues("test", "left", "right", result, config)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for ignored field, got %d", len(result.Diffs))
		}
	})

	t.Run("identical reference types", func(t *testing.T) {
		result := &DiffResult{}
		slice := []int{1, 2, 3}
		err := compareValues("test", slice, slice, result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for identical slices, got %d", len(result.Diffs))
		}
	})

	t.Run("custom comparator", func(t *testing.T) {
		result := &DiffResult{}
		config := DefaultCompareConfig()
		config.CustomComparators = map[reflect.Type]func(left, right any, config *CompareConfig) (bool, error){
			reflect.TypeOf(""): func(left, right any, config *CompareConfig) (bool, error) {
				return left == right, nil
			},
		}
		err := compareValues("test", "same", "same", result, config)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences with custom comparator, got %d", len(result.Diffs))
		}
	})
}

func TestCompareStructsEdgeCases(t *testing.T) {

	t.Run("struct with unexported fields", func(t *testing.T) {
		type StructWithUnexported struct {
			Exported   string
			unexported int
		}

		left := StructWithUnexported{Exported: "test", unexported: 1}
		right := StructWithUnexported{Exported: "test", unexported: 2}

		result := &DiffResult{}
		leftVal := reflect.ValueOf(left)
		rightVal := reflect.ValueOf(right)

		err := compareStructs("test", leftVal, rightVal, result, DefaultCompareConfig())
		if err != nil {
			t.Fatalf("compareStructs failed: %v", err)
		}

		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for struct with unexported fields, got %d", len(result.Diffs))
		}
	})

	t.Run("struct with diff tag ignore", func(t *testing.T) {
		type StructWithIgnore struct {
			Field1 string `diff:"ignore"`
			Field2 int
		}

		left := StructWithIgnore{Field1: "left", Field2: 1}
		right := StructWithIgnore{Field1: "right", Field2: 1}

		result := &DiffResult{}
		leftVal := reflect.ValueOf(left)
		rightVal := reflect.ValueOf(right)

		err := compareStructs("test", leftVal, rightVal, result, DefaultCompareConfig())
		if err != nil {
			t.Fatalf("compareStructs failed: %v", err)
		}

		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences with ignored field, got %d", len(result.Diffs))
		}
	})
}
