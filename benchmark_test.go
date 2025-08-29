package godiff

import (
	"testing"
	"time"
)

type BenchmarkStruct struct {
	ID       int
	Name     string
	Age      int
	Email    string
	Address  *Address
	Hobbies  []string
	Metadata map[string]any
	Active   bool
	Created  time.Time
}

func createLargeBenchmarkStruct(id int) BenchmarkStruct {
	return BenchmarkStruct{
		ID:    id,
		Name:  "User Name " + string(rune(id)),
		Age:   20 + (id % 50),
		Email: "user" + string(rune(id)) + "@example.com",
		Address: &Address{
			Street:  "123 Main St " + string(rune(id)),
			City:    "City " + string(rune(id%10)),
			Country: "Country",
		},
		Hobbies: []string{"hobby1", "hobby2", "hobby3"},
		Metadata: map[string]any{
			"role":        "user",
			"permissions": []string{"read", "write"},
			"settings":    map[string]bool{"notifications": true},
		},
		Active:  id%2 == 0,
		Created: time.Now(),
	}
}

func BenchmarkCompareBasicTypes(b *testing.B) {
	left := "hello"
	right := "world"

	for b.Loop() {
		_, _ = Compare(left, right)
	}
}

func BenchmarkCompareStructs(b *testing.B) {
	left := createLargeBenchmarkStruct(1)
	right := createLargeBenchmarkStruct(2)

	for b.Loop() {
		_, _ = Compare(left, right)
	}
}

func BenchmarkCompareSlices(b *testing.B) {
	left := make([]int, 100)
	right := make([]int, 100)

	for i := range left {
		left[i] = i
		right[i] = i + 1
	}

	for b.Loop() {
		_, _ = Compare(left, right)
	}
}

func BenchmarkCompareSlicesWithIgnoreOrder(b *testing.B) {
	type TestStruct struct {
		Items []int `diff:"ignoreOrder"`
	}

	left := TestStruct{Items: make([]int, 100)}
	right := TestStruct{Items: make([]int, 100)}

	for i := range left.Items {
		left.Items[i] = i
		right.Items[99-i] = i
	}

	for b.Loop() {
		_, _ = Compare(left, right)
	}
}

func BenchmarkCompareLargeStructSlices(b *testing.B) {
	left := make([]BenchmarkStruct, 50)
	right := make([]BenchmarkStruct, 50)

	for i := range left {
		left[i] = createLargeBenchmarkStruct(i)
		right[i] = createLargeBenchmarkStruct(i + 1)
	}

	for b.Loop() {
		_, _ = Compare(left, right)
	}
}

func BenchmarkCompareMaps(b *testing.B) {
	left := make(map[string]int)
	right := make(map[string]int)

	for i := range 100 {
		key := "key" + string(rune(i))
		left[key] = i
		right[key] = i + 1
	}

	for b.Loop() {
		_, _ = Compare(left, right)
	}
}

func BenchmarkCompareIdentical(b *testing.B) {
	data := createLargeBenchmarkStruct(1)

	for b.Loop() {
		_, _ = Compare(data, data)
	}
}

func BenchmarkStringGeneration(b *testing.B) {
	left := createLargeBenchmarkStruct(1)
	right := createLargeBenchmarkStruct(2)

	result, _ := Compare(left, right)

	for b.Loop() {
		_ = result.String()
	}
}

func BenchmarkJSONGeneration(b *testing.B) {
	left := createLargeBenchmarkStruct(1)
	right := createLargeBenchmarkStruct(2)

	result, _ := Compare(left, right)

	for b.Loop() {
		_ = result.ToJSON()
	}
}
