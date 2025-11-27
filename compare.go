package godiff

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// CompareOption is a function that modifies a CompareConfig
type CompareOption func(*CompareConfig)

// WithIgnoreFields sets the fields to ignore during comparison
func WithIgnoreFields(fields ...string) CompareOption {
	return func(c *CompareConfig) {
		c.IgnoreFields = fields
	}
}

// WithIgnoreSliceOrder enables ignoring slice element order during comparison
func WithIgnoreSliceOrder() CompareOption {
	return func(c *CompareConfig) {
		c.IgnoreSliceOrder = true
	}
}

// WithCompareNumericValues enables comparing numeric values across different types
func WithCompareNumericValues() CompareOption {
	return func(c *CompareConfig) {
		c.CompareNumericValues = true
	}
}

// WithCustomComparators sets custom comparison functions for specific types
func WithCustomComparators(comparators map[reflect.Type]func(left, right any, config *CompareConfig) (bool, error)) CompareOption {
	return func(c *CompareConfig) {
		c.CustomComparators = comparators
	}
}

// WithTypeHandlers sets the type handlers for comparing custom or complex types
func WithTypeHandlers(handlers []TypeHandler) CompareOption {
	return func(c *CompareConfig) {
		c.TypeHandlers = handlers
	}
}

// WithMaxDepth sets the maximum recursion depth for comparison (0 means unlimited)
func WithMaxDepth(depth int) CompareOption {
	return func(c *CompareConfig) {
		c.MaxDepth = depth
	}
}

// Compare compares two values of any type and returns the differences.
// Optional configuration can be provided via CompareOption functions.
func Compare(left, right any, opts ...CompareOption) (*DiffResult, error) {
	config := DefaultCompareConfig()

	for _, opt := range opts {
		opt(config)
	}

	if config.visitedPairs == nil {
		config.visitedPairs = make(map[[2]uintptr]bool)
	}

	if config.ignoreFieldsSet == nil && len(config.IgnoreFields) > 0 {
		config.ignoreFieldsSet = make(map[string]bool, len(config.IgnoreFields))
		for _, field := range config.IgnoreFields {
			config.ignoreFieldsSet[field] = true
		}
	}
	config.currentDepth = 0
	result := &DiffResult{}
	err := compareValues("", left, right, result, config)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// handleInvalidValues checks if either value is invalid and records a diff if needed
// Returns true if handled (one or both values invalid), false if both are valid
func handleInvalidValues(path string, left, right any, leftVal, rightVal reflect.Value, result *DiffResult) bool {
	if !leftVal.IsValid() && !rightVal.IsValid() {
		return true // both invalid, no diff
	}

	if !leftVal.IsValid() {
		result.AddDiff(path, nil, right)
		return true
	}

	if !rightVal.IsValid() {
		result.AddDiff(path, left, nil)
		return true
	}

	return false // both valid, not handled
}

// compareValues recursively compares two values and records differences
func compareValues(path string, left, right any, result *DiffResult, config *CompareConfig) error {
	if config.MaxDepth > 0 {
		if config.currentDepth >= config.MaxDepth {
			return nil
		}
		config.currentDepth++
		defer func() { config.currentDepth-- }()
	}

	if config.ignoreFieldsSet != nil {
		if config.ignoreFieldsSet[path] {
			return nil
		}
	} else if slices.Contains(config.IgnoreFields, path) {
		return nil
	}

	// Early exit: identical reference types (ptr/map/slice/chan/func) share same pointer
	if left != nil && right != nil {
		lv := reflect.ValueOf(left)
		rv := reflect.ValueOf(right)
		if lv.IsValid() && rv.IsValid() && lv.Type() == rv.Type() {
			switch lv.Kind() {
			case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
				if lv.Pointer() == rv.Pointer() {
					return nil
				}
			}
		}
	}

	leftVal := reflect.ValueOf(left)
	rightVal := reflect.ValueOf(right)

	if handleInvalidValues(path, left, right, leftVal, rightVal, result) {
		return nil
	}

	leftType := leftVal.Type()
	rightType := rightVal.Type()

	if leftType != rightType {
		// Special case: nil pointers of different types are considered equal
		if leftVal.Kind() == reflect.Ptr && rightVal.Kind() == reflect.Ptr &&
			leftVal.IsNil() && rightVal.IsNil() {
			return nil
		}
		// Check if both are numeric types and config allows cross-type numeric comparison
		if config.CompareNumericValues && isNumericKind(leftVal.Kind()) && isNumericKind(rightVal.Kind()) {
			if numericValuesEqual(leftVal, rightVal) {
				return nil
			}
			result.Diffs = append(result.Diffs, &Diff{
				Path:  path,
				Left:  left,
				Right: right,
			})
			return nil
		}
		result.Diffs = append(result.Diffs, &Diff{
			Path:  path,
			Left:  left,
			Right: right,
		})
		return nil
	}

	if config.CustomComparators != nil {
		if customComparator, exists := config.CustomComparators[leftType]; exists {
			equal, err := customComparator(left, right, config)
			if err != nil {
				return err
			}
			if !equal {
				result.Diffs = append(result.Diffs, &Diff{
					Path:  path,
					Left:  left,
					Right: right,
				})
			}
			return nil
		}
	}

	if config.TypeHandlers != nil {
		for _, handler := range config.TypeHandlers {
			if handler.CanHandle(leftType) {
				return handler.Compare(left, right, path, result, config)
			}
		}
	}

	leftKind := leftVal.Kind()
	switch leftKind {
	case reflect.Struct:
		return compareStructs(path, leftVal, rightVal, result, config)
	case reflect.Slice, reflect.Array:
		return compareSlices(path, leftVal, rightVal, result, config)
	case reflect.Map:
		return compareMaps(path, leftVal, rightVal, result, config)
	case reflect.Pointer:
		return comparePointers(path, leftVal, rightVal, result, config)
	default:
		if leftVal.Type().Comparable() {
			if left != right {
				result.Diffs = append(result.Diffs, &Diff{Path: path, Left: left, Right: right})
			}
			return nil
		}
		if !reflect.DeepEqual(left, right) {
			result.Diffs = append(result.Diffs, &Diff{Path: path, Left: left, Right: right})
		}
		return nil
	}
}

// isFieldIgnored checks if a field should be ignored based on IgnoreFields configuration
// It checks multiple patterns:
// 1. Simple field name (e.g., "Meta")
// 2. Full path (e.g., "User.Meta" or "Address.City")
// 3. Type-qualified field name (e.g., "MyStruct.Meta")
func isFieldIgnored(fieldPath string, fieldName string, structType reflect.Type, config *CompareConfig) bool {
	if len(config.IgnoreFields) == 0 {
		return false
	}

	if config.ignoreFieldsSet != nil {
		if config.ignoreFieldsSet[fieldPath] {
			return true
		}

		if config.ignoreFieldsSet[fieldName] {
			return true
		}

		structTypeName := structType.Name()
		if structTypeName != "" {
			typeQualifiedName := structTypeName + "." + fieldName
			if config.ignoreFieldsSet[typeQualifiedName] {
				return true
			}
		}
		return false
	}

	// Fall back to slice search
	if slices.Contains(config.IgnoreFields, fieldPath) {
		return true
	}

	if slices.Contains(config.IgnoreFields, fieldName) {
		return true
	}

	structTypeName := structType.Name()
	if structTypeName != "" {
		typeQualifiedName := structTypeName + "." + fieldName
		if slices.Contains(config.IgnoreFields, typeQualifiedName) {
			return true
		}
	}

	return false
}

// compareStructs compares two structs field by field
func compareStructs(path string, leftVal, rightVal reflect.Value, result *DiffResult, config *CompareConfig) error {
	typ := leftVal.Type()
	numFields := leftVal.NumField()

	for i := range numFields {
		field := typ.Field(i)
		// Skip unexported fields to avoid calling Interface() on values we can't access from
		// another package (this prevents panics for types like time.Time).
		if !field.IsExported() {
			continue
		}

		var fieldPath string
		if path == "" {
			fieldPath = field.Name
		} else {
			fieldPath = path + "." + field.Name
		}

		diffTag := field.Tag.Get("diff")
		if isFieldIgnored(fieldPath, field.Name, typ, config) || hasDiffTag(diffTag, "ignore") {
			continue
		}

		leftField := leftVal.Field(i)
		rightField := rightVal.Field(i)
		leftFieldInterface := leftField.Interface()
		rightFieldInterface := rightField.Interface()

		if field.Type.Kind() == reflect.Slice {
			modifiedConfig := config

			if hasDiffTag(diffTag, "ignoreOrder") {
				modifiedConfig = &CompareConfig{
					IgnoreFields:      config.IgnoreFields,
					IgnoreSliceOrder:  true,
					CustomComparators: config.CustomComparators,
					TypeHandlers:      config.TypeHandlers,
					visitedPairs:      config.visitedPairs,
					ignoreFieldsSet:   config.ignoreFieldsSet,
					MaxDepth:          config.MaxDepth,
					currentDepth:      config.currentDepth,
				}
			}

			err := compareSlices(fieldPath, leftField, rightField, result, modifiedConfig)
			if err != nil {
				return err
			}
		} else {
			if !reflect.DeepEqual(leftFieldInterface, rightFieldInterface) {
				leftKind := leftField.Kind()
				if leftKind == reflect.Pointer || leftKind == reflect.Struct ||
					leftKind == reflect.Map || leftKind == reflect.Interface {
					err := compareValues(fieldPath, leftFieldInterface, rightFieldInterface, result, config)
					if err != nil {
						return err
					}
				} else {
					result.Diffs = append(result.Diffs, &StructDiff{
						Diff: Diff{
							Path:  fieldPath,
							Left:  leftFieldInterface,
							Right: rightFieldInterface,
						},
						FieldName:  field.Name,
						ChangeType: ChangeTypeUpdated,
					})
				}
			}
		}
	}
	return nil
}

// compareSlices compares two slices using appropriate algorithm based on configuration
func compareSlices(path string, leftVal, rightVal reflect.Value, result *DiffResult, config *CompareConfig) error {
	if config.IgnoreSliceOrder {
		return compareSlicesAdvanced(path, leftVal, rightVal, result)
	}

	leftLen := leftVal.Len()
	rightLen := rightVal.Len()
	maxLen := max(rightLen, leftLen)

	for i := range maxLen {
		var leftElem, rightElem any
		var hasLeftElem, hasRightElem bool

		if i < leftLen {
			leftElem = leftVal.Index(i).Interface()
			hasLeftElem = true
		}
		if i < rightLen {
			rightElem = rightVal.Index(i).Interface()
			hasRightElem = true
		}

		if hasLeftElem && hasRightElem {
			leftElemVal := reflect.ValueOf(leftElem)
			if isBasicKind(leftElemVal.Kind()) && !reflect.DeepEqual(leftElem, rightElem) {
				result.Diffs = append(result.Diffs, &SliceDiff{
					Diff: Diff{
						Path:  path,
						Left:  leftElem,
						Right: rightElem,
					},
					Index:      i,
					ChangeType: ChangeTypeUpdated,
				})
			} else {
				elementPath := path + "[" + itoa(i) + "]"
				err := compareValues(elementPath, leftElem, rightElem, result, config)
				if err != nil {
					return err
				}
			}
		} else if hasLeftElem {
			// removed
			result.Diffs = append(result.Diffs, &SliceDiff{
				Diff: Diff{
					Path:  path,
					Left:  leftElem,
					Right: nil,
				},
				Index:      i,
				ChangeType: ChangeTypeRemoved,
			})
		} else if hasRightElem {
			// added
			result.Diffs = append(result.Diffs, &SliceDiff{
				Diff: Diff{
					Path:  path,
					Left:  nil,
					Right: rightElem,
				},
				Index:      i,
				ChangeType: ChangeTypeAdded,
			})
		}
	}
	return nil
}

// compareSlicesAdvanced compares slices using ID-based matching or value-based matching
func compareSlicesAdvanced(path string, leftVal, rightVal reflect.Value, result *DiffResult) error {

	if !leftVal.IsValid() && !rightVal.IsValid() {
		return nil
	}

	if !leftVal.IsValid() {
		if rightVal.IsValid() {
			result.Diffs = append(result.Diffs, &Diff{
				Path:  path,
				Left:  nil,
				Right: rightVal.Interface(),
			})
		}
		return nil
	}

	if !rightVal.IsValid() {
		result.Diffs = append(result.Diffs, &Diff{
			Path:  path,
			Left:  leftVal.Interface(),
			Right: nil,
		})
		return nil
	}

	if leftVal.Type() != rightVal.Type() {
		result.Diffs = append(result.Diffs, &Diff{
			Path:  path,
			Left:  leftVal.Interface(),
			Right: rightVal.Interface(),
		})
		return nil
	}

	return compareSlicesByValue(path, leftVal, rightVal, result)
}

// compareSlicesByValue compares slices using value-based matching (similar to the original ignoreOrder)
func compareSlicesByValue(path string, leftVal, rightVal reflect.Value, result *DiffResult) error {
	elemType := leftVal.Type().Elem()
	if !elemType.Comparable() {
		return compareSlicesWithDeepEqual(path, leftVal, rightVal, result)
	}

	leftLen := leftVal.Len()
	rightLen := rightVal.Len()

	// For very small slices, use simple comparison to avoid overhead of maps
	if leftLen <= 5 && rightLen <= 5 {
		return compareSlicesSimple(path, leftVal, rightVal, result)
	}

	leftCounts := make(map[any]int, leftLen)
	rightCounts := make(map[any]int, rightLen)

	for i := 0; i < leftLen; i++ {
		elem := leftVal.Index(i).Interface()
		leftCounts[elem]++
	}

	for i := 0; i < rightLen; i++ {
		elem := rightVal.Index(i).Interface()
		rightCounts[elem]++
	}

	maxDiffs := leftLen + rightLen
	if cap(result.Diffs) < len(result.Diffs)+maxDiffs {
		result.Diffs = slices.Grow(result.Diffs, maxDiffs)
	}

	// removed
	for elem, leftCount := range leftCounts {
		rightCount := rightCounts[elem]
		if leftCount > rightCount {
			for j := 0; j < leftCount-rightCount; j++ {
				result.Diffs = append(result.Diffs, &Diff{
					Path:  path,
					Left:  elem,
					Right: nil,
				})
			}
		}
	}

	// added
	for elem, rightCount := range rightCounts {
		leftCount := leftCounts[elem]
		if rightCount > leftCount {
			for j := 0; j < rightCount-leftCount; j++ {
				result.Diffs = append(result.Diffs, &Diff{
					Path:  path,
					Left:  nil,
					Right: elem,
				})
			}
		}
	}

	return nil
}

// compareSlicesUnordered provides unified comparison for slices ignoring order
// Uses DeepEqual for matching elements
func compareSlicesUnordered(path string, leftVal, rightVal reflect.Value, result *DiffResult) error {
	leftLen := leftVal.Len()
	rightLen := rightVal.Len()

	rightMatched := make([]bool, rightLen)

	for i := range leftLen {
		leftElem := leftVal.Index(i).Interface()
		found := false

		for j := range rightLen {
			if !rightMatched[j] {
				rightElem := rightVal.Index(j).Interface()
				if reflect.DeepEqual(leftElem, rightElem) {
					rightMatched[j] = true
					found = true
					break
				}
			}
		}

		if !found {
			result.Diffs = append(result.Diffs, &Diff{
				Path:  path,
				Left:  leftElem,
				Right: nil,
			})
		}
	}

	// Find unmatched right elements
	for j := range rightLen {
		if !rightMatched[j] {
			rightElem := rightVal.Index(j).Interface()
			result.Diffs = append(result.Diffs, &Diff{
				Path:  path,
				Left:  nil,
				Right: rightElem,
			})
		}
	}

	return nil
}

// compareSlicesSimple provides optimized comparison for small slices
func compareSlicesSimple(path string, leftVal, rightVal reflect.Value, result *DiffResult) error {
	return compareSlicesUnordered(path, leftVal, rightVal, result)
}

// compareSlicesWithDeepEqual compares slices using DeepEqual for non-comparable types, ignoring order
func compareSlicesWithDeepEqual(path string, leftVal, rightVal reflect.Value, result *DiffResult) error {
	return compareSlicesUnordered(path, leftVal, rightVal, result)
}

// isBasicKind returns true if the kind is a basic comparable type (numeric, bool, or string)
func isBasicKind(k reflect.Kind) bool {
	return k <= reflect.Complex128 || k == reflect.String
}

// isNumericKind returns true if the kind is a numeric type (int, uint, float, complex)
func isNumericKind(k reflect.Kind) bool {
	return (k >= reflect.Int && k <= reflect.Float64) || k == reflect.Complex64 || k == reflect.Complex128
}

// isIntegerKind returns true if the kind is an integer type
func isIntegerKind(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Uintptr
}

// isFloatKind returns true if the kind is a floating-point type
func isFloatKind(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}

// numericValuesEqual compares two numeric values across different types
func numericValuesEqual(leftVal, rightVal reflect.Value) bool {
	leftKind := leftVal.Kind()
	rightKind := rightVal.Kind()

	// Both are signed integers
	if isSignedIntKind(leftKind) && isSignedIntKind(rightKind) {
		return leftVal.Int() == rightVal.Int()
	}

	// Both are unsigned integers
	if isUnsignedIntKind(leftKind) && isUnsignedIntKind(rightKind) {
		return leftVal.Uint() == rightVal.Uint()
	}

	// Both are floats
	if isFloatKind(leftKind) && isFloatKind(rightKind) {
		return leftVal.Float() == rightVal.Float()
	}

	// Mixed signed/unsigned integers - need careful comparison
	if isSignedIntKind(leftKind) && isUnsignedIntKind(rightKind) {
		leftInt := leftVal.Int()
		rightUint := rightVal.Uint()
		if leftInt < 0 {
			return false
		}
		return uint64(leftInt) == rightUint
	}
	if isUnsignedIntKind(leftKind) && isSignedIntKind(rightKind) {
		leftUint := leftVal.Uint()
		rightInt := rightVal.Int()
		if rightInt < 0 {
			return false
		}
		return leftUint == uint64(rightInt)
	}

	// Integer and float comparison
	if isIntegerKind(leftKind) && isFloatKind(rightKind) {
		var leftFloat float64
		if isSignedIntKind(leftKind) {
			leftFloat = float64(leftVal.Int())
		} else {
			leftFloat = float64(leftVal.Uint())
		}
		return leftFloat == rightVal.Float()
	}
	if isFloatKind(leftKind) && isIntegerKind(rightKind) {
		var rightFloat float64
		if isSignedIntKind(rightKind) {
			rightFloat = float64(rightVal.Int())
		} else {
			rightFloat = float64(rightVal.Uint())
		}
		return leftVal.Float() == rightFloat
	}

	// Complex numbers
	if (leftKind == reflect.Complex64 || leftKind == reflect.Complex128) &&
		(rightKind == reflect.Complex64 || rightKind == reflect.Complex128) {
		return leftVal.Complex() == rightVal.Complex()
	}

	return false
}

// isSignedIntKind returns true if the kind is a signed integer
func isSignedIntKind(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

// isUnsignedIntKind returns true if the kind is an unsigned integer
func isUnsignedIntKind(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uintptr
}

// compareMaps compares two maps key by keywithout fmt
func itoa(i int) string {
	return strconv.Itoa(i)
}

// compareMaps compares two maps key by key
func compareMaps(path string, leftVal, rightVal reflect.Value, result *DiffResult, config *CompareConfig) error {
	for _, key := range leftVal.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		elementPath := path + "[" + keyStr + "]"

		rightMapVal := rightVal.MapIndex(key)
		leftMapVal := leftVal.MapIndex(key)
		if !rightMapVal.IsValid() {
			// Key removed
			result.Diffs = append(result.Diffs, &MapDiff{
				Diff: Diff{
					Path:  elementPath,
					Left:  leftMapVal.Interface(),
					Right: nil,
				},
				Key:        key.Interface(),
				ChangeType: ChangeTypeRemoved,
			})
			continue
		}

		leftInterface := leftMapVal.Interface()
		rightInterface := rightMapVal.Interface()

		leftValReflect := reflect.ValueOf(leftInterface)
		rightValReflect := reflect.ValueOf(rightInterface)

		// Check for type mismatch with potential numeric comparison
		if leftValReflect.Type() != rightValReflect.Type() {
			if config.CompareNumericValues && isNumericKind(leftValReflect.Kind()) && isNumericKind(rightValReflect.Kind()) {
				if !numericValuesEqual(leftValReflect, rightValReflect) {
					result.Diffs = append(result.Diffs, &MapDiff{
						Diff: Diff{
							Path:  elementPath,
							Left:  leftInterface,
							Right: rightInterface,
						},
						Key:        key.Interface(),
						ChangeType: ChangeTypeUpdated,
					})
				}
			} else {
				result.Diffs = append(result.Diffs, &MapDiff{
					Diff: Diff{
						Path:  elementPath,
						Left:  leftInterface,
						Right: rightInterface,
					},
					Key:        key.Interface(),
					ChangeType: ChangeTypeUpdated,
				})
			}
			continue
		}

		if isBasicKind(leftValReflect.Kind()) {
			if !reflect.DeepEqual(leftInterface, rightInterface) {
				result.Diffs = append(result.Diffs, &MapDiff{
					Diff: Diff{
						Path:  elementPath,
						Left:  leftInterface,
						Right: rightInterface,
					},
					Key:        key.Interface(),
					ChangeType: ChangeTypeUpdated,
				})
			}
		} else {
			tempResult := &DiffResult{}
			err := compareValues(elementPath, leftInterface, rightInterface, tempResult, config)
			if err != nil {
				return err
			}

			if len(tempResult.Diffs) > 0 {
				result.Diffs = append(result.Diffs, tempResult.Diffs...)
			}
		}
	}

	// added
	for _, key := range rightVal.MapKeys() {
		if !leftVal.MapIndex(key).IsValid() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			elementPath := path + "[" + keyStr + "]"

			result.Diffs = append(result.Diffs, &MapDiff{
				Diff: Diff{
					Path:  elementPath,
					Left:  nil,
					Right: rightVal.MapIndex(key).Interface(),
				},
				Key:        key.Interface(),
				ChangeType: ChangeTypeAdded,
			})
		}
	}

	return nil
}

// comparePointers compares two pointers by dereferencing them
func comparePointers(path string, leftVal, rightVal reflect.Value, result *DiffResult, config *CompareConfig) error {
	if leftVal.IsNil() && rightVal.IsNil() {
		return nil
	}

	if leftVal.IsNil() {
		return compareValues(path, nil, rightVal.Elem().Interface(), result, config)
	}

	if rightVal.IsNil() {
		return compareValues(path, leftVal.Elem().Interface(), nil, result, config)
	}

	leftPtr := leftVal.Pointer()
	rightPtr := rightVal.Pointer()
	pairKey := [2]uintptr{leftPtr, rightPtr}

	if config.visitedPairs[pairKey] {
		return nil
	}

	config.visitedPairs[pairKey] = true
	err := compareValues(path, leftVal.Elem().Interface(), rightVal.Elem().Interface(), result, config)
	delete(config.visitedPairs, pairKey)

	return err
}

// hasDiffTag checks if the diff tag contains an exact match for the given tag value
func hasDiffTag(diffTag, tagValue string) bool {
	if diffTag == "" {
		return false
	}
	tags := strings.SplitSeq(diffTag, ",")
	for tag := range tags {
		if strings.TrimSpace(tag) == tagValue {
			return true
		}
	}
	return false
}
