package godiff

import (
	"fmt"
	"math/big"
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

// WithAdditionalTypeHandlers adds handlers ahead of the defaults. Earlier
// handlers take precedence, so this can also override a default handler.
func WithAdditionalTypeHandlers(handlers ...TypeHandler) CompareOption {
	return func(c *CompareConfig) {
		combined := make([]TypeHandler, 0, len(handlers)+len(c.TypeHandlers))
		combined = append(combined, handlers...)
		c.TypeHandlers = append(combined, c.TypeHandlers...)
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
		if opt != nil {
			opt(config)
		}
	}

	if config.visitedPairs == nil {
		config.visitedPairs = make(map[visit]bool)
	}

	if config.ignoreFieldsSet == nil && len(config.IgnoreFields) > 0 {
		config.ignoreFieldsSet = make(map[string]bool, len(config.IgnoreFields))
		for _, field := range config.IgnoreFields {
			config.ignoreFieldsSet[field] = true
		}
	}
	if len(opts) > 0 {
		config.hasCustomTypeHandlers = containsCustomTypeHandler(config.TypeHandlers)
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

	leftVal := reflect.ValueOf(left)
	rightVal := reflect.ValueOf(right)

	if handleInvalidValues(path, left, right, leftVal, rightVal, result) {
		return nil
	}

	leftType := leftVal.Type()
	rightType := rightVal.Type()

	if leftType != rightType {
		// Special case: nil pointers of different types are considered equal
		if leftVal.Kind() == reflect.Pointer && rightVal.Kind() == reflect.Pointer &&
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
			if customComparator == nil {
				return fmt.Errorf("custom comparator for %s is nil", leftType)
			}
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
			if handler == nil || (config.hasCustomTypeHandlers && isNilTypeHandler(handler)) {
				continue
			}
			if handler.CanHandle(leftType) {
				return handler.Compare(left, right, path, result, config)
			}
		}
	}

	// Early exit for references that expose exactly the same value. This comes
	// after configured comparators and handlers so they remain authoritative.
	// Slice length matters because slices can share a backing array while
	// exposing different values.
	switch leftVal.Kind() {
	case reflect.Slice:
		if leftVal.Pointer() == rightVal.Pointer() && leftVal.Len() == rightVal.Len() {
			return nil
		}
	case reflect.Pointer, reflect.Map, reflect.Chan, reflect.Func:
		if leftVal.Pointer() == rightVal.Pointer() {
			return nil
		}
	}

	leftKind := leftVal.Kind()
	if leftKind == reflect.Map || leftKind == reflect.Pointer || leftKind == reflect.Slice {
		if pair, ok := referenceVisit(leftVal, rightVal); ok {
			if config.visitedPairs[pair] {
				return nil
			}
			config.visitedPairs[pair] = true
			defer delete(config.visitedPairs, pair)
		}
	}

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

func isNilTypeHandler(handler TypeHandler) bool {
	if handler == nil {
		return true
	}

	value := reflect.ValueOf(handler)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func containsCustomTypeHandler(handlers []TypeHandler) bool {
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		if isNilTypeHandler(handler) {
			return true
		}
		switch handler.(type) {
		case *TimeHandler, *InterfaceHandler, *FunctionHandler, *ChannelHandler:
		default:
			return true
		}
	}
	return false
}

func canSkipDeepEqualValues(config *CompareConfig) bool {
	return len(config.CustomComparators) == 0 && !config.hasCustomTypeHandlers
}

func comparisonDepthExceeded(config *CompareConfig) bool {
	return config.MaxDepth > 0 && config.currentDepth >= config.MaxDepth
}

func isPathIgnored(path string, config *CompareConfig) bool {
	if config.ignoreFieldsSet != nil {
		return config.ignoreFieldsSet[path]
	}
	return slices.Contains(config.IgnoreFields, path)
}

func canCompareDirectly(left, right any, config *CompareConfig) bool {
	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)
	if !leftValue.IsValid() || !rightValue.IsValid() {
		return true
	}

	leftType := leftValue.Type()
	if leftType != rightValue.Type() {
		return !config.CompareNumericValues ||
			!isNumericKind(leftValue.Kind()) || !isNumericKind(rightValue.Kind())
	}
	return canCompareTypeDirectly(leftType, config)
}

func canCompareTypeDirectly(typ reflect.Type, config *CompareConfig) bool {
	if _, exists := config.CustomComparators[typ]; exists {
		return false
	}

	if config.hasCustomTypeHandlers {
		for _, handler := range config.TypeHandlers {
			if handler != nil && !isNilTypeHandler(handler) && handler.CanHandle(typ) {
				return false
			}
		}
	} else {
		switch typ.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Struct:
			for _, handler := range config.TypeHandlers {
				if handler != nil && handler.CanHandle(typ) {
					return false
				}
			}
		}
	}

	switch typ.Kind() {
	case reflect.Array, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
		return false
	default:
		return true
	}
}

func referenceVisit(leftVal, rightVal reflect.Value) (visit, bool) {
	if leftVal.Type() != rightVal.Type() {
		return visit{}, false
	}

	switch leftVal.Kind() {
	case reflect.Map, reflect.Pointer:
		if leftVal.IsNil() || rightVal.IsNil() {
			return visit{}, false
		}
		return visit{
			typ:   leftVal.Type(),
			left:  leftVal.Pointer(),
			right: rightVal.Pointer(),
		}, true
	case reflect.Slice:
		if leftVal.IsNil() || rightVal.IsNil() {
			return visit{}, false
		}
		return visit{
			typ:      leftVal.Type(),
			left:     leftVal.Pointer(),
			right:    rightVal.Pointer(),
			leftLen:  leftVal.Len(),
			rightLen: rightVal.Len(),
		}, true
	default:
		return visit{}, false
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

		fieldConfig := config
		if (field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array) &&
			hasDiffTag(diffTag, "ignoreOrder") {
			modifiedConfig := *config
			modifiedConfig.IgnoreSliceOrder = true
			fieldConfig = &modifiedConfig
		}

		leftField := leftVal.Field(i).Interface()
		rightField := rightVal.Field(i).Interface()
		if canSkipDeepEqualValues(fieldConfig) && reflect.DeepEqual(leftField, rightField) {
			continue
		}
		if comparisonDepthExceeded(fieldConfig) {
			continue
		}
		if canCompareTypeDirectly(field.Type, fieldConfig) {
			if !reflect.DeepEqual(leftField, rightField) {
				result.AddStructDiff(fieldPath, field.Name, leftField, rightField, ChangeTypeUpdated)
			}
			continue
		}

		start := len(result.Diffs)
		if err := compareValues(fieldPath, leftField, rightField, result, fieldConfig); err != nil {
			return err
		}
		normalizeStructFieldDiff(result, start, fieldPath, field.Name, leftField, rightField)
	}
	return nil
}

func normalizeStructFieldDiff(result *DiffResult, start int, path, fieldName string, left, right any) {
	if len(result.Diffs) == start+1 {
		if diff, ok := result.Diffs[start].(*Diff); ok && diff.Path == path {
			result.Diffs[start] = &StructDiff{
				Path: path, Left: left, Right: right,
				FieldName:  fieldName,
				ChangeType: ChangeTypeUpdated,
			}
		}
	}
}

// compareSlices compares two slices using appropriate algorithm based on configuration
func compareSlices(path string, leftVal, rightVal reflect.Value, result *DiffResult, config *CompareConfig) error {
	if config.IgnoreSliceOrder {
		return compareSlicesAdvanced(path, leftVal, rightVal, result, config)
	}

	leftLen := leftVal.Len()
	rightLen := rightVal.Len()
	maxLen := max(rightLen, leftLen)
	elementType := leftVal.Type().Elem()
	directElementType := canCompareTypeDirectly(elementType, config)
	dynamicElementType := elementType.Kind() == reflect.Interface

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
			if comparisonDepthExceeded(config) {
				continue
			}
			if directElementType || (dynamicElementType && canCompareDirectly(leftElem, rightElem, config)) {
				if len(config.IgnoreFields) > 0 {
					elementPath := path + "[" + itoa(i) + "]"
					if isPathIgnored(elementPath, config) {
						continue
					}
				}
				if !reflect.DeepEqual(leftElem, rightElem) {
					result.AddSliceDiff(path, i, leftElem, rightElem, ChangeTypeUpdated)
				}
				continue
			}

			elementPath := path + "[" + itoa(i) + "]"
			if isPathIgnored(elementPath, config) {
				continue
			}
			start := len(result.Diffs)
			if err := compareValues(elementPath, leftElem, rightElem, result, config); err != nil {
				return err
			}
			normalizeSliceElementDiff(result, start, path, elementPath, i, leftElem, rightElem)
		} else if hasLeftElem {
			// removed
			result.Diffs = append(result.Diffs, &SliceDiff{
				Path:       path,
				Left:       leftElem,
				Right:      nil,
				Index:      i,
				ChangeType: ChangeTypeRemoved,
			})
		} else if hasRightElem {
			// added
			result.Diffs = append(result.Diffs, &SliceDiff{
				Path:       path,
				Left:       nil,
				Right:      rightElem,
				Index:      i,
				ChangeType: ChangeTypeAdded,
			})
		}
	}
	return nil
}

func normalizeSliceElementDiff(result *DiffResult, start int, path, elementPath string, index int, left, right any) {
	if len(result.Diffs) == start+1 {
		if diff, ok := result.Diffs[start].(*Diff); ok && diff.Path == elementPath {
			result.Diffs[start] = &SliceDiff{
				Path: path, Left: left, Right: right,
				Index:      index,
				ChangeType: ChangeTypeUpdated,
			}
		}
	}
}

// compareSlicesAdvanced compares slices as multisets. The optional config keeps
// the helper convenient for direct internal tests while production calls pass
// the active comparison configuration.
func compareSlicesAdvanced(path string, leftVal, rightVal reflect.Value, result *DiffResult, configs ...*CompareConfig) error {
	config := DefaultCompareConfig()
	if len(configs) > 0 && configs[0] != nil {
		config = configs[0]
	}

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

	if canCountSliceValues(leftVal.Type().Elem(), config) {
		compareSlicesByCount(path, leftVal, rightVal, result)
		return nil
	}

	return compareSlicesSemantically(path, leftVal, rightVal, result, config)
}

func canCountSliceValues(elemType reflect.Type, config *CompareConfig) bool {
	if config.MaxDepth > 0 && config.currentDepth >= config.MaxDepth {
		return false
	}

	switch elemType.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.String:
	default:
		return false
	}

	if config.CustomComparators != nil {
		if _, exists := config.CustomComparators[elemType]; exists {
			return false
		}
	}
	if config.hasCustomTypeHandlers {
		for _, handler := range config.TypeHandlers {
			if handler != nil && !isNilTypeHandler(handler) && handler.CanHandle(elemType) {
				return false
			}
		}
	}
	return true
}

func compareSlicesByCount(path string, leftVal, rightVal reflect.Value, result *DiffResult) {
	leftCounts := make(map[any]int, leftVal.Len())
	for i := range leftVal.Len() {
		leftCounts[leftVal.Index(i).Interface()]++
	}

	rightRemaining := make(map[any]int, rightVal.Len())
	for i := range rightVal.Len() {
		rightRemaining[rightVal.Index(i).Interface()]++
	}
	if len(leftCounts) == len(rightRemaining) {
		equal := true
		for elem, count := range leftCounts {
			if rightRemaining[elem] != count {
				equal = false
				break
			}
		}
		if equal {
			return
		}
	}

	for i := range leftVal.Len() {
		elem := leftVal.Index(i).Interface()
		if rightRemaining[elem] > 0 {
			rightRemaining[elem]--
			continue
		}
		result.AddDiff(path, elem, nil)
	}

	for i := range rightVal.Len() {
		elem := rightVal.Index(i).Interface()
		if rightRemaining[elem] > 0 {
			rightRemaining[elem]--
			result.AddDiff(path, nil, elem)
		}
	}
}

func compareSlicesSemantically(path string, leftVal, rightVal reflect.Value, result *DiffResult, config *CompareConfig) error {
	rightMatched := make([]bool, rightVal.Len())

	for i := range leftVal.Len() {
		leftElem := leftVal.Index(i).Interface()
		matched := false
		for j := range rightVal.Len() {
			if rightMatched[j] {
				continue
			}
			rightElem := rightVal.Index(j).Interface()
			candidateResult := &DiffResult{}
			elementPath := path + "[" + itoa(i) + "]"
			if err := compareValues(elementPath, leftElem, rightElem, candidateResult, config); err != nil {
				return err
			}
			if !candidateResult.HasDifferences() {
				rightMatched[j] = true
				matched = true
				break
			}
		}
		if !matched {
			result.AddDiff(path, leftElem, nil)
		}
	}

	for j := range rightVal.Len() {
		if !rightMatched[j] {
			result.AddDiff(path, nil, rightVal.Index(j).Interface())
		}
	}
	return nil
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
		return integerValueEqualsFloat(leftVal, rightVal.Float())
	}
	if isFloatKind(leftKind) && isIntegerKind(rightKind) {
		return integerValueEqualsFloat(rightVal, leftVal.Float())
	}

	// Complex numbers
	if (leftKind == reflect.Complex64 || leftKind == reflect.Complex128) &&
		(rightKind == reflect.Complex64 || rightKind == reflect.Complex128) {
		return leftVal.Complex() == rightVal.Complex()
	}

	return false
}

func integerValueEqualsFloat(integer reflect.Value, float float64) bool {
	floatValue := new(big.Rat)
	if floatValue.SetFloat64(float) == nil || !floatValue.IsInt() {
		return false
	}

	integerValue := new(big.Int)
	if isSignedIntKind(integer.Kind()) {
		integerValue.SetInt64(integer.Int())
	} else {
		integerValue.SetUint64(integer.Uint())
	}
	return floatValue.Num().Cmp(integerValue) == 0
}

// isSignedIntKind returns true if the kind is a signed integer
func isSignedIntKind(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

// isUnsignedIntKind returns true if the kind is an unsigned integer
func isUnsignedIntKind(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uintptr
}

// itoa formats a slice index without pulling formatting into hot paths.
func itoa(i int) string {
	return strconv.Itoa(i)
}

// compareMaps compares two maps key by key
func compareMaps(path string, leftVal, rightVal reflect.Value, result *DiffResult, config *CompareConfig) error {
	valueType := leftVal.Type().Elem()
	directValueType := canCompareTypeDirectly(valueType, config)
	dynamicValueType := valueType.Kind() == reflect.Interface
	leftIterator := leftVal.MapRange()
	for leftIterator.Next() {
		key := leftIterator.Key()
		leftMapVal := leftIterator.Value()
		keyStr := fmt.Sprintf("%v", key.Interface())
		elementPath := path + "[" + keyStr + "]"

		rightMapVal := rightVal.MapIndex(key)
		if !rightMapVal.IsValid() {
			// Key removed
			result.AddMapDiff(elementPath, key.Interface(), leftMapVal.Interface(), nil, ChangeTypeRemoved)
			continue
		}

		leftInterface := leftMapVal.Interface()
		rightInterface := rightMapVal.Interface()
		if comparisonDepthExceeded(config) || isPathIgnored(elementPath, config) {
			continue
		}
		if directValueType || (dynamicValueType && canCompareDirectly(leftInterface, rightInterface, config)) {
			if !reflect.DeepEqual(leftInterface, rightInterface) {
				result.AddMapDiff(elementPath, key.Interface(), leftInterface, rightInterface, ChangeTypeUpdated)
			}
			continue
		}

		start := len(result.Diffs)
		if err := compareValues(elementPath, leftInterface, rightInterface, result, config); err != nil {
			return err
		}
		normalizeMapValueDiff(result, start, elementPath, key.Interface(), leftInterface, rightInterface)
	}

	// added
	rightIterator := rightVal.MapRange()
	for rightIterator.Next() {
		key := rightIterator.Key()
		if !leftVal.MapIndex(key).IsValid() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			elementPath := path + "[" + keyStr + "]"
			result.AddMapDiff(elementPath, key.Interface(), nil, rightIterator.Value().Interface(), ChangeTypeAdded)
		}
	}

	return nil
}

func normalizeMapValueDiff(result *DiffResult, start int, path string, key, left, right any) {
	if len(result.Diffs) == start+1 {
		if diff, ok := result.Diffs[start].(*Diff); ok && diff.Path == path {
			result.Diffs[start] = &MapDiff{
				Path: path, Left: left, Right: right,
				Key:        key,
				ChangeType: ChangeTypeUpdated,
			}
		}
	}
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

	return compareValues(path, leftVal.Elem().Interface(), rightVal.Elem().Interface(), result, config)
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
