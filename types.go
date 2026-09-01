package godiff

import (
	"reflect"
)

// ChangeType represents the type of change detected
type ChangeType string

const (
	ChangeTypeAdded   ChangeType = "ADDED"
	ChangeTypeRemoved ChangeType = "REMOVED"
	ChangeTypeUpdated ChangeType = "UPDATED"
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

// visit identifies a pair of reference values currently being compared.
// Slice lengths are part of the identity because two slices can share a
// backing array while exposing different values.
type visit struct {
	typ               reflect.Type
	left, right       uintptr
	leftLen, rightLen int
}

// visitTracker keeps the common shallow case inline. A map is only needed for
// unusually deep reference graphs, avoiding a large map bucket on every
// comparison that happens to contain a pointer, map, or slice.
type visitTracker struct {
	inline [2]visit
	count  int
	deep   map[visit]struct{}
}

func (vt *visitTracker) enter(pair visit) bool {
	if vt.deep != nil {
		if _, seen := vt.deep[pair]; seen {
			return false
		}
		vt.deep[pair] = struct{}{}
		return true
	}

	for i := range vt.count {
		if vt.inline[i] == pair {
			return false
		}
	}
	if vt.count < len(vt.inline) {
		vt.inline[vt.count] = pair
		vt.count++
		return true
	}

	vt.deep = make(map[visit]struct{}, len(vt.inline)+1)
	for _, active := range vt.inline {
		vt.deep[active] = struct{}{}
	}
	vt.deep[pair] = struct{}{}
	return true
}

func (vt *visitTracker) leave(pair visit) {
	if vt.deep != nil {
		delete(vt.deep, pair)
		return
	}

	// Visits are entered and left by nested compareValues calls, so the active
	// inline entries form a stack.
	vt.count--
	vt.inline[vt.count] = visit{}
}

// AddDiff adds a basic Diff to the result
func (dr *DiffResult) AddDiff(path string, left, right any) {
	dr.Diffs = append(dr.Diffs, &Diff{Path: path, Left: left, Right: right})
}

// AddStructDiff adds a StructDiff to the result
func (dr *DiffResult) AddStructDiff(path, fieldName string, left, right any, changeType ChangeType) {
	dr.Diffs = append(dr.Diffs, &StructDiff{
		Path: path, Left: left, Right: right,
		FieldName:  fieldName,
		ChangeType: changeType,
	})
}

// AddSliceDiff adds a SliceDiff to the result
func (dr *DiffResult) AddSliceDiff(path string, index int, left, right any, changeType ChangeType) {
	dr.Diffs = append(dr.Diffs, &SliceDiff{
		Path: path, Left: left, Right: right,
		Index:      index,
		ChangeType: changeType,
	})
}

// AddMapDiff adds a MapDiff to the result
func (dr *DiffResult) AddMapDiff(path string, key, left, right any, changeType ChangeType) {
	dr.Diffs = append(dr.Diffs, &MapDiff{
		Path: path, Left: left, Right: right,
		Key:        key,
		ChangeType: changeType,
	})
}

// CompareConfig holds configuration options for the comparison.
// Note: CompareConfig is not thread-safe. Do not share a single config instance
// across multiple concurrent Compare calls.
type CompareConfig struct {
	// IgnoreFields is a list of field paths to ignore during comparison (e.g., "User.Password").
	IgnoreFields []string
	// IgnoreSliceOrder, if true, ignores element order when comparing slices.
	IgnoreSliceOrder bool
	// CompareNumericValues, if true, compares numeric values across different types.
	// For example, int(1) and int64(1) would be considered equal.
	// This applies to integer, floating-point, and complex types.
	CompareNumericValues bool
	// hasCustomTypeHandlers avoids repeatedly scanning the handler list on hot paths.
	hasCustomTypeHandlers bool
	// CustomComparators is a map of custom comparison functions for specific types.
	CustomComparators map[reflect.Type]func(left, right any, config *CompareConfig) (bool, error)
	// TypeHandlers is a list of handlers for comparing custom or complex types.
	TypeHandlers []TypeHandler
	// MaxDepth limits the recursion depth for comparison. 0 means unlimited.
	MaxDepth int
	// visitedPairs tracks active reference pairs for cycle detection (internal use only).
	visitedPairs *visitTracker
	// ignoreFieldsSet is a pre-computed set for O(1) lookup (internal use only)
	ignoreFieldsSet map[string]bool
	// currentDepth tracks the current recursion depth (internal use only)
	currentDepth int
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
		IgnoreSliceOrder: false,
		TypeHandlers:     DefaultTypeHandlers(),
	}
}
