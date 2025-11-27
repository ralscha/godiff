# GoDiff

[![CI](https://github.com/ralscha/godiff/actions/workflows/ci.yml/badge.svg)](https://github.com/ralscha/godiff/actions/workflows/ci.yml)

Structural diff library for Go. Compares structs, slices, maps, and primitives recursively.

## Quick Start

```bash
go get github.com/ralscha/godiff
```

```go
person1 := Person{Name: "Alice", Age: 30}
person2 := Person{Name: "Bob", Age: 30}

result, err := godiff.Compare(person1, person2)
if err != nil {
    panic(err)
}

fmt.Println("Has Differences:", result.HasDifferences())
// Output: Has Differences: true

fmt.Println("Total differences:", result.Count())
// Output: Total differences: 1

for _, diff := range result.Diffs {
    switch d := diff.(type) {
    case *godiff.MapDiff:
        // For map diffs
    case *godiff.SliceDiff:
        // For slice diffs
    case *godiff.StructDiff:
        fmt.Println("StructDiff - Field:", d.FieldName, "ChangeType:", d.ChangeType, "Left:", d.Left, "Right:", d.Right)
    case *godiff.Diff:
        // For primitive diffs
    }
}
// Output: StructDiff - Field: Name ChangeType: updated Left: Alice Right: Bob

fmt.Println("JSON Output:", result.ToJSON())
// Output:
// JSON Output: [
//   {
//     "type": "struct",
//     "path": "",
//     "leftValue": "Alice",
//     "rightValue": "Bob",
//     "fieldName": "Name",
//     "change": "UPDATED"
//   }
// ]


fmt.Println(result.String())
// Output:
// Found 1 differences:
// UPDATED Name: Alice -> Bob
```

## Configuration

### Options

```go
result, err := godiff.Compare(left, right,
    godiff.WithIgnoreFields("Password", "User.Meta"),
    godiff.WithIgnoreSliceOrder(),
    godiff.WithCompareNumericValues(),
    godiff.WithMaxDepth(10),
)
```

| Option | Description |
|--------|-------------|
| `WithIgnoreFields(fields...)` | Skip specific fields by name or path |
| `WithIgnoreSliceOrder()` | Compare slices without regard to element order |
| `WithCompareNumericValues()` | Compare numeric values across different types |
| `WithMaxDepth(n)` | Limit recursion depth (0 = unlimited) |
| `WithCustomComparators(map)` | Custom comparison functions for specific types |
| `WithTypeHandlers(handlers)` | Custom handlers for complex types |

### Struct Tags

```go
type Product struct {
    Name   string
    Tags   []string `diff:"ignoreOrder"` // Compare ignoring order
    Secret string   `diff:"ignore"`      // Skip this field
}
```

## Demo

See the demo application for more examples:

```bash
go run cmd/demo/main.go
```

## License

MIT
