# GoDiff - Structural Diff Library for Go

[![CI](https://github.com/ralscha/godiff/actions/workflows/ci.yml/badge.svg)](https://github.com/ralscha/godiff/actions/workflows/ci.yml)

GoDiff is a Go library for computing structural differences between any two values. It provides detailed, recursive comparison of primitives, structs, pointers, slices, and maps, returning a structured result that describes all changes.

## Installation

To use GoDiff in your project, install it using `go get`:
```bash
go get github.com/ralscha/godiff
```

## Quick Start

### Basic Comparison

```go
package main

import (
    "fmt"
    "github.com/ralscha/godiff"
)

func main() {
    leftValue := map[string]any{"name": "Alice", "age": 30}
    rightValue := map[string]any{"name": "Bob", "age": 30}
    
    result, err := godiff.Compare(leftValue, rightValue)
    if err != nil {
        panic(err)
    }
    
    // The String() method provides a human-readable summary
    fmt.Println(result.String())
	// Output:
	/*
	Found 1 differences:
	UPDATED [name]: Alice -> Bob
	 */
}
```

### Using Configuration

```go
// Define two structs to compare
left := MyStruct{ID: 1, Value: "old", Meta: "ignored"}
right := MyStruct{ID: 1, Value: "new", Meta: "also ignored"}

// Configure the comparison to ignore the "Meta" field
config := godiff.DefaultCompareConfig()
config.IgnoreFields = []string{"Meta"}

result, err := godiff.CompareWithConfig(left, right, config)
if err != nil {
    // Handle error
}
fmt.Println(result.String())
// Output:
/*
   Found 1 differences:
   UPDATED Value: old -> new
*/

fmt.Println(result.ToJSON())
// Output:
/*
[
  {
    "type": "struct",
    "path": "",
    "leftValue": "old",
    "rightValue": "new",
    "fieldName": "Value",
    "change": "UPDATED"
  }
]
*/
```

## Features

- **Deep Comparison**: Recursively compares any Go values, including nested structs, slices, and maps.
- **Flexible Configuration**: Customize comparison behavior with `CompareConfig`.
- **Struct Tags**: Use tags like `diff:"ignore"`, `diff:"id"`, and `diff:"ignoreOrder"` to control comparison at the field level.
- **Advanced Slice Handling**: Compare slices with order sensitivity or ignore order (`ignoreOrder`)
- **Built-in Type Handlers**: Special handling for `time.Time`, functions, channels, and interface types.
- **Detailed Output**: Get structured diffs with specialized types (`MapDiff`, `SliceDiff`, `StructDiff`) that provide rich context.
- **Extensible**: Add support for custom types using `TypeHandler` or provide custom comparison logic with `CustomComparators`.

## Public API

### Types

#### ChangeType
Represents the type of change detected.
- `ChangeTypeAdded`: Value was added.
- `ChangeTypeRemoved`: Value was removed.
- `ChangeTypeUpdated`: Value was updated.
- `ChangeTypeIDMismatch`: The IDs of two structs are different.

#### Diff Types
GoDiff returns a `DiffResult` containing a slice of diff objects. The specific type of object depends on where the difference was found.

- **`Diff`**: A generic difference for basic types.
  ```go
  type Diff struct {
      Path  string // JSON-like path to the differing field
      Left  any    // Left value (nil if added)
      Right any    // Right value (nil if removed)
  }
  ```
- **`StructDiff`**: A difference found within a struct field.
  ```go

  type StructDiff struct {
      Diff
      FieldName  string     // The struct field name that changed
      ChangeType ChangeType // Type of change: ADDED, REMOVED, UPDATED
  }
  ```
- **`SliceDiff`**: A difference found within a slice or array.
  ```go
  type SliceDiff struct {
      Diff
      Index      int        // The slice/array index that changed
      ChangeType ChangeType // Type of change: ADDED, REMOVED, UPDATED
  }
  ```
- **`MapDiff`**: A difference found within a map.
  ```go
  type MapDiff struct {
      Diff
      Key        any        // The map key that changed
      ChangeType ChangeType // Type of change: ADDED, REMOVED, UPDATED
  }
  ```

#### DiffResult
Contains all differences found.
```go
type DiffResult struct {
    Diffs []any // Holds a collection of Diff, StructDiff, SliceDiff, or MapDiff
}
```
`DiffResult` has several helpful methods:
- `String() string`: Returns a human-readable summary of differences.
- `ToJSON() string`: Returns a JSON representation of the diffs.
- `HasDifferences() bool`: Returns `true` if any diffs were found.
- `Count() int`: Returns the total number of differences.

#### CompareConfig
Configuration options for comparison.
```go
type CompareConfig struct {
    // IgnoreFields is a list of field paths to ignore during comparison (e.g., "User.Password").
    IgnoreFields          []string
    // IDFieldNames is a list of field names to use as unique identifiers for matching structs (e.g., "ID", "UUID").
    // This is only used if a struct does not have a `diff:"id"` tag. By default, this is empty.
    IDFieldNames          []string
    // IgnoreSliceOrder, if true, ignores element order when comparing slices.
    IgnoreSliceOrder      bool
    // CustomComparators is a map of custom comparison functions for specific types.
    CustomComparators     map[reflect.Type]func(left, right any, config *CompareConfig) (bool, error)
    // TypeHandlers is a list of handlers for comparing custom or complex types.
    TypeHandlers          []TypeHandler
}
```

### Functions

#### Compare
```go
func Compare(left, right any) (*DiffResult, error)
```
Compares two values using default configuration.

#### CompareWithConfig
```go
func CompareWithConfig(left, right any, config *CompareConfig) (*DiffResult, error)
```
Compares two values with custom configuration.

#### DefaultCompareConfig
```go
func DefaultCompareConfig() *CompareConfig
```
Returns the default configuration with sensible defaults and built-in `TypeHandler` implementations for `time.Time`, `any`, `func`, and `chan`.

## Usage Examples

### Using Struct Tags

Struct tags provide fine-grained control over the comparison logic directly in your type definitions.

```go
type Product struct {
    SKU      string   `diff:"id"`            // Use as a unique ID for matching
    Name     string
    Tags     []string `diff:"ignoreOrder"`   // Ignore order when comparing this slice
    Secret   string   `diff:"ignore"`        // Ignore this field completely
}
```

#### Available Struct Tags:
- `diff:"id"`: Marks a field as the unique identifier for a struct. When both structs have ID fields:
  - If IDs match: GoDiff compares the remaining fields normally
  - If IDs don't match: Returns a single `StructDiff` with `ChangeTypeIDMismatch` and stops comparing other fields
- `diff:"ignore"`: Skips a field entirely during comparison.
- `diff:"ignoreOrder"`: For slice fields, compares elements without regard to their order. Ensures both slices have the same elements with the same frequencies.

#### ID Matching Example:
```go
type User struct {
    ID   int    `diff:"id"`
    Name string
    Age  int
}

left := User{ID: 1, Name: "Alice", Age: 25}
right := User{ID: 2, Name: "Alice", Age: 25} // Different ID

result, _ := godiff.Compare(left, right)
// Result: Single StructDiff with ChangeTypeIDMismatch
// Other fields (Name, Age) are not compared when IDs differ
```

### Advanced Configuration

```go
config := godiff.DefaultCompareConfig()

// Use struct fields named "ID", "UUID", or "SKU" to match structs
// if they don't have a `diff:"id"` tag.
config.IDFieldNames = []string{"ID", "UUID", "SKU"}

// Ignore fields using multiple matching patterns:
config.IgnoreFields = []string{
    "Password",           // Simple field name - matches any field named "Password" 
    "User.Password",      // Full path - matches Password field in User struct
    "MyStruct.Metadata", // Type-qualified - matches Metadata field in MyStruct type
    "Config.DB.Host",    // Nested path - matches deeply nested fields
}

// Ignores slice element order
config.IgnoreSliceOrder = true

result, err := godiff.CompareWithConfig(leftData, rightData, config)
```

#### Field Ignore Patterns

GoDiff supports flexible field matching patterns:

1. **Simple Field Name**: `"Password"` - Ignores any field named Password in any struct
2. **Full Path**: `"User.Address.Street"` - Ignores specific field at exact path
3. **Type-Qualified**: `"MyStruct.CreatedAt"` - Ignores CreatedAt field only in MyStruct type
4. **Nested Paths**: `"Config.Database.Connection.Timeout"` - Handles deep nesting

### Built-in Type Handlers

GoDiff includes specialized handlers for complex Go types:

#### Time Handler
- Compares `time.Time` values using `time.Time.Equal()` method
- Handles timezone differences correctly
- Example: `time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)` vs `time.Date(2023, 1, 1, 12, 0, 1, 0, time.UTC)`

#### Interface Handler  
- Safely handles `interface{}` and `any` types
- Compares underlying values after type assertion
- Handles nil interface values correctly

#### Function Handler
- Compares functions by pointer identity
- Useful for configuration structs containing callback functions
- nil vs non-nil function differences are detected

#### Channel Handler
- Compares channels by reference
- Detects when different channel instances are used


### Advanced Slice Comparison

#### Order-Sensitive Comparison (Default)
```go
left := []string{"a", "b", "c"}
right := []string{"a", "c", "b"}
// Results in differences at index 1 and 2
```

#### Order-Insensitive Comparison  
```go
type Config struct {
    Tags []string `diff:"ignoreOrder"`
}

left := Config{Tags: []string{"admin", "user", "moderator"}}
right := Config{Tags: []string{"user", "admin", "moderator"}}
// No differences - same elements, different order
```

## Interpreting Results

The `String()` method on `DiffResult` gives a readable summary. Each line indicates the path, the change, and the values.

Example output from the demo:
```
UPDATED Age.Age: 30 -> 31
UPDATED Email.Email: alice@example.com -> alice.right@example.com
UPDATED Address.City.City: New York -> Boston
UPDATED Tags[1]: user -> moderator
UPDATED Metadata[lastLogin][lastLogin]: 2023-01-01 -> 2024-01-01
ADDED Metadata[status][status]: active
```

For programmatic access, you can iterate through `result.Diffs` and use a type switch to handle each diff type.

```go
for _, d := range result.Diffs {
    switch diff := d.(type) {
    case *godiff.StructDiff:
        fmt.Printf("Struct field '%s' at path '%s' was %s", diff.FieldName, diff.Path, diff.ChangeType)
    case *godiff.SliceDiff:
        // ... handle slice diffs
    case *godiff.MapDiff:
        // ... handle map diffs
    }
}
```

### Working with Results

The `DiffResult` provides several methods for working with comparison results:

```go
result, err := godiff.Compare(left, right)
if err != nil {
    log.Fatal(err)
}

// Check if there are any differences
if result.HasDifferences() {
    fmt.Printf("Found %d differences\n", result.Count())
    
    // Human-readable summary
    fmt.Println(result.String())
    
    // Structured JSON output for APIs or storage
    jsonOutput := result.ToJSON()
    fmt.Println(jsonOutput)
} else {
    fmt.Println("Objects are identical")
}
```

## Demo

You can find a demo application in `cmd/demo/main.go` that showcases the features of the library.

Run the demo:
```bash
go run cmd/demo/main.go
```

## License

MIT - See LICENSE file for details.
