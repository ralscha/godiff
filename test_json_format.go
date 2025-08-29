package main

import (
	"fmt"
	"godiff"
)

func main() {
	left := map[string]any{"name": "Alice", "age": 30}
	right := map[string]any{"name": "Bob", "age": 30}

	result, err := godiff.Compare(left, right)
	if err != nil {
		panic(err)
	}

	fmt.Println("JSON output:")
	fmt.Println(result.ToJSON())
}
