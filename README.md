# GoDiff

[![CI](https://github.com/ralscha/godiff/actions/workflows/ci.yml/badge.svg)](https://github.com/ralscha/godiff/actions/workflows/ci.yml)

Structural diff library for Go. Compares structs, slices, maps, and primitives recursively.

## Requirements

Go 1.27.3 or newer.

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
| `WithTypeHandlers(handlers)` | Replace all type handlers; defaults handle `time.Time`, interfaces, functions, and channels |
| `WithAdditionalTypeHandlers(handlers...)` | Add higher-priority handlers while retaining the defaults |

Options apply recursively, including inside structs, maps, ordered slices, and
unordered slices. Numeric comparison is exact: large integers are not treated as
equal to rounded floating-point values.

Ignored paths also apply to added and removed entries. For example,
`WithIgnoreFields("Metadata[updatedAt]", "Items[0]")` skips that map entry and
slice index. When ignoring slice order, ignored indexes refer to the original
positions in each input, before matching. Unordered comparisons preserve
duplicate counts and find matching pairs using the configured comparators;
comparators should be deterministic because matching can call them repeatedly.

`WithMaxDepth(1)` compares root scalar values but skips fields and collection
entries. Each pointer dereference also consumes a depth level. Use zero for
unlimited depth; reference cycles are detected automatically.

Only exported struct fields are compared. Functions are compared by code
pointer, which cannot distinguish all closures with different captured values.
Use a custom comparator when your application needs different function semantics.

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
go run ./cmd/demo
```

## JSON

`result.ToJSON()` returns the formatted change array shown above using
`encoding/json/v2`. `DiffResult` implements both `json.Marshaler` and the more
efficient `json.MarshalerTo`, so it can be embedded directly in another JSON
document with `json.Marshal(result)`.

Use `WriteJSON` to write the same formatted array to an `io.Writer` and receive
serialization or write errors directly:

```go
if err := result.WriteJSON(os.Stdout); err != nil {
    return err
}
```

Streaming output can be partial when an error occurs. `ToJSON` keeps its
string-only API and returns an error object inside an array if serialization
fails. Text output marks cyclic values instead of recursing indefinitely.

## Benchmarks

Run the benchmark suite with allocation statistics:

```bash
go test ./... -bench . -benchmem -run ^$
```

## License

MIT
