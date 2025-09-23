package main

import (
	"fmt"
	"godiff"
	"reflect"
	"time"
)

type User struct {
	Name     string
	Age      int
	Email    string
	Address  *Address
	Tags     []string
	Metadata map[string]any
}

type Address struct {
	Street  string
	City    string
	Country string
}

func main() {
	// Example 1: Basic types
	fmt.Println("1. Comparing basic types:")
	result1, _ := godiff.Compare("hello", "world")
	fmt.Println(result1.String())
	fmt.Println()

	// Example 2: Struct comparison
	fmt.Println("2. Comparing structs:")
	leftUser := User{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
		Address: &Address{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
		},
		Tags: []string{"admin", "user"},
		Metadata: map[string]any{
			"role":      "admin",
			"lastLogin": "2023-01-01",
		},
	}

	rightUser := User{
		Name:  "Alice",
		Age:   31,
		Email: "alice.right@example.com",
		Address: &Address{
			Street:  "123 Main St",
			City:    "Boston",
			Country: "USA",
		},
		Tags: []string{"admin", "moderator"},
		Metadata: map[string]any{
			"role":      "admin",
			"lastLogin": "2024-01-01",
			"status":    "active",
		},
	}

	result2, _ := godiff.Compare(leftUser, rightUser)
	fmt.Println(result2.String())
	fmt.Println()

	// Example 3: Slice comparison
	fmt.Println("3. Comparing slices:")
	leftSlice := []int{1, 2, 3}
	rightSlice := []int{1, 4, 5}
	result3, _ := godiff.Compare(leftSlice, rightSlice)
	fmt.Println(result3.String())
	fmt.Println()

	// Example 4: Map comparison
	fmt.Println("4. Comparing maps:")
	leftMap := map[string]int{"a": 1, "b": 2, "c": 3}
	rightMap := map[string]int{"a": 1, "b": 4, "d": 5}
	result4, _ := godiff.Compare(leftMap, rightMap)
	fmt.Println(result4.String())
	fmt.Println()

	// Example 5: No differences
	fmt.Println("5. Comparing identical values:")
	result5, _ := godiff.Compare("same", "same")
	fmt.Println(result5.String())
	fmt.Println()

	// Example 6: Slice with ignoreOrder tag
	fmt.Println("6. Comparing slices with ignoreOrder tag:")
	type Product struct {
		Name  string
		Sizes []string `diff:"ignoreOrder"`
	}
	leftProduct := Product{Name: "Shirt", Sizes: []string{"S", "M", "L"}}
	rightProduct := Product{Name: "Shirt", Sizes: []string{"L", "S", "M"}}
	result6, _ := godiff.Compare(leftProduct, rightProduct)
	fmt.Println(result6.String())
	fmt.Println()

	// Example 7: Struct with ignored field
	fmt.Println("7. Comparing structs with an ignored field:")
	type UserWithIgnore struct {
		ID    int
		Email string `diff:"ignore"`
		Role  string
	}
	leftUserWithIgnore := UserWithIgnore{ID: 1, Email: "left@example.com", Role: "user"}
	rightUserWithIgnore := UserWithIgnore{ID: 1, Email: "right@example.com", Role: "admin"}
	result7, _ := godiff.Compare(leftUserWithIgnore, rightUserWithIgnore)
	fmt.Println(result7.String())
	fmt.Println()

	// Example 8: Type changed
	fmt.Println("8. Comparing different types (TYPE_CHANGED):")
	result8, _ := godiff.Compare(42, "42")
	fmt.Println(result8.String())
	fmt.Println()

	// Example 9: Time comparison with TimeHandler
	fmt.Println("9. Comparing time values:")
	time1 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	time2 := time.Date(2023, 1, 1, 12, 0, 1, 0, time.UTC) // 1-second difference
	result10, _ := godiff.Compare(time1, time2)
	fmt.Println(result10.String())
	fmt.Println()

	// Example 10: Interface comparison with InterfaceHandler
	fmt.Println("10. Comparing interface values:")
	var iface1, iface2 any
	iface1 = "hello"
	iface2 = "world"
	result11, _ := godiff.Compare(iface1, iface2)
	fmt.Println(result11.String())
	fmt.Println()

	// Example 11: Function comparison with FunctionHandler
	fmt.Println("11. Comparing function references:")
	func1 := func() { fmt.Println("Function 1") }
	func2 := func() { fmt.Println("Function 2") }
	result12, _ := godiff.Compare(func1, func2)
	fmt.Println(result12.String())
	fmt.Println()

	// Example 12: Nil vs value comparisons
	fmt.Println("12. Comparing nil vs values:")
	result13a, _ := godiff.Compare(nil, "value")
	result13b, _ := godiff.Compare("value", nil)
	fmt.Println("nil -> value:", result13a.String())
	fmt.Println("value -> nil:", result13b.String())
	fmt.Println()

	// Example 13: Complex nested structure
	fmt.Println("13. Comparing complex nested structures:")
	type ComplexStruct struct {
		NestedMap    map[string][]map[string]int
		NestedSlice  []map[string][]string
		PointerField *User
	}

	leftComplex := ComplexStruct{
		NestedMap: map[string][]map[string]int{
			"outer": {
				{"inner1": 1, "inner2": 2},
				{"inner3": 3, "inner4": 4},
			},
		},
		NestedSlice: []map[string][]string{
			{"key1": {"a", "b", "c"}},
			{"key2": {"d", "e", "f"}},
		},
		PointerField: &leftUser,
	}

	rightComplex := ComplexStruct{
		NestedMap: map[string][]map[string]int{
			"outer": {
				{"inner1": 1, "inner2": 5},
				{"inner3": 3, "inner4": 4},
			},
		},
		NestedSlice: []map[string][]string{
			{"key1": {"a", "b", "c"}},
			{"key2": {"d", "e", "modified"}},
		},
		PointerField: &rightUser,
	}

	result14, _ := godiff.Compare(leftComplex, rightComplex)
	fmt.Println(result14.String())
	fmt.Println()

	// Example 14: Custom comparator example
	fmt.Println("14. Using custom comparator:")
	type CustomType struct {
		Value string
	}

	customComparator := func(left, right any, config *godiff.CompareConfig) (bool, error) {
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

	customConfig := godiff.DefaultCompareConfig()
	customConfig.CustomComparators = map[reflect.Type]func(left, right any, config *godiff.CompareConfig) (bool, error){
		reflect.TypeOf(CustomType{}): customComparator,
	}

	leftCustom := CustomType{Value: "hello world"}
	rightCustom := CustomType{Value: "help me please"}
	result15, _ := godiff.CompareWithConfig(leftCustom, rightCustom, customConfig)
	fmt.Println(result15.String())
	fmt.Println()

	// Example 15: Channel comparison with ChannelHandler
	fmt.Println("15. Comparing channels:")
	ch1 := make(chan int, 5)
	ch2 := make(chan int, 5)
	result16a, _ := godiff.Compare(ch1, ch2)
	result16b, _ := godiff.Compare(ch1, ch1)
	fmt.Println("Different channels:", result16a.String())
	fmt.Println("Same channel:", result16b.String())
	fmt.Println()

	// Example 16: Output methods demonstration
	fmt.Println("16. Output methods demonstration:")
	result17, _ := godiff.Compare("hello", "world")

	fmt.Println("String output:")
	fmt.Println(result17.String())
	fmt.Println()

	fmt.Println("HasDifferences():", result17.HasDifferences())
	fmt.Println()

	fmt.Println("Count():", result17.Count())
	fmt.Println()

	fmt.Println("ToJSON() output:")
	fmt.Println(result17.ToJSON())
	fmt.Println()

}
