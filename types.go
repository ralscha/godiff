package godiff

import (
	"reflect"
)

// ChangeType represents the type of change detected
type ChangeType string

const (
	ChangeTypeAdded      ChangeType = "ADDED"
	ChangeTypeRemoved    ChangeType = "REMOVED"
	ChangeTypeUpdated    ChangeType = "UPDATED"
	ChangeTypeIDMismatch ChangeType = "ID_MISMATCH"
)

// Diff represents a single difference between two values
type Diff struct {
	Path  string // JSON path to the differing field
	Left  any    // Left value (nil if added)
	Right any    // Right value (nil if removed)
}

// MapDiff represents a difference in a map
type MapDiff struct {
	Diff
	Key        any        // The map key that changed
	ChangeType ChangeType // Type of change: ADDED, REMOVED, UPDATED
}

// SliceDiff represents a difference in a slice
type SliceDiff struct {
	Diff
	Index      int        // The slice index that changed
	ChangeType ChangeType // Type of change: ADDED, REMOVED, UPDATED
}

// StructDiff represents a difference in a struct
type StructDiff struct {
	Diff
	FieldName  string     // The struct field name that changed
	ChangeType ChangeType // Type of change: ADDED, REMOVED, UPDATED
}

// DiffResult contains all differences found between two values
type DiffResult struct {
	Diffs []any // Can hold Diff, MapDiff, SliceDiff, or StructDiff
}

// CompareConfig holds configuration options for the comparison
type CompareConfig struct {
	// IgnoreFields is a list of field paths to ignore during comparison (e.g., "User.Password").
	IgnoreFields []string
	// IDFieldNames is a list of field names to use as unique identifiers for matching structs.
	// This is only used if a struct does not have a `diff:"id"` tag.
	// By default, this is empty.
	IDFieldNames []string
	// IgnoreSliceOrder, if true, ignores element order when comparing slices.
	IgnoreSliceOrder bool
	// CustomComparators is a map of custom comparison functions for specific types.
	CustomComparators map[reflect.Type]func(left, right any, config *CompareConfig) (bool, error)
	// TypeHandlers is a list of handlers for comparing custom or complex types.
	TypeHandlers []TypeHandler
	// visitedPairs tracks visited pointer pairs for cycle detection (internal use only)
	visitedPairs map[[2]uintptr]bool
}

// TypeHandler defines an interface for handling specific types during comparison
type TypeHandler interface {
	CanHandle(typ reflect.Type) bool
	Compare(left, right any, path string, result *DiffResult, config *CompareConfig) error
}

// DefaultCompareConfig returns the default configuration
func DefaultCompareConfig() *CompareConfig {
	return &CompareConfig{
		IgnoreFields:     []string{},
		IDFieldNames:     []string{},
		IgnoreSliceOrder: false,
		TypeHandlers:     DefaultTypeHandlers(),
		visitedPairs:     make(map[[2]uintptr]bool),
	}
}
