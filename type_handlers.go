package godiff

import (
	"fmt"
	"reflect"
	"time"
)

// TimeHandler handles time.Time comparisons
type TimeHandler struct{}

func (h *TimeHandler) CanHandle(typ reflect.Type) bool {
	return typ == reflect.TypeOf(time.Time{})
}

func (h *TimeHandler) Compare(left, right any, path string, result *DiffResult, config *CompareConfig) error {
	leftTime, ok1 := left.(time.Time)
	rightTime, ok2 := right.(time.Time)

	if !ok1 || !ok2 {
		return fmt.Errorf("TimeHandler received non-time values: left=%T, right=%T", left, right)
	}

	if !leftTime.Equal(rightTime) {
		result.Diffs = append(result.Diffs, &Diff{
			Path:  path,
			Left:  leftTime,
			Right: rightTime,
		})
	}
	return nil
}

// InterfaceHandler handles any types by comparing their underlying values
type InterfaceHandler struct{}

func (h *InterfaceHandler) CanHandle(typ reflect.Type) bool {
	return typ.Kind() == reflect.Interface
}

func (h *InterfaceHandler) Compare(left, right any, path string, result *DiffResult, config *CompareConfig) error {
	leftVal := reflect.ValueOf(left)
	rightVal := reflect.ValueOf(right)

	leftIsNil := !leftVal.IsValid() || (leftVal.Kind() == reflect.Interface && leftVal.IsNil())
	rightIsNil := !rightVal.IsValid() || (rightVal.Kind() == reflect.Interface && rightVal.IsNil())

	if leftIsNil && rightIsNil {
		return nil
	}

	if leftIsNil {
		result.Diffs = append(result.Diffs, &Diff{
			Path:  path,
			Left:  nil,
			Right: right,
		})
		return nil
	}

	if rightIsNil {
		result.Diffs = append(result.Diffs, &Diff{
			Path:  path,
			Left:  left,
			Right: nil,
		})
		return nil
	}

	return compareValues(path, leftVal.Elem().Interface(), rightVal.Elem().Interface(), result, config)
}

// FunctionHandler handles function types (compares by reference or ignores based on config)
type FunctionHandler struct{}

func (h *FunctionHandler) CanHandle(typ reflect.Type) bool {
	return typ.Kind() == reflect.Func
}

func (h *FunctionHandler) Compare(left, right any, path string, result *DiffResult, config *CompareConfig) error {
	// Functions compare by pointer identity; nil vs non-nil counts as diff.
	leftVal := reflect.ValueOf(left)
	rightVal := reflect.ValueOf(right)

	if !leftVal.IsValid() && !rightVal.IsValid() {
		return nil
	}

	if !leftVal.IsValid() || !rightVal.IsValid() {
		result.Diffs = append(result.Diffs, &Diff{Path: path, Left: left, Right: right})
		return nil
	}

	if leftVal.IsNil() && rightVal.IsNil() {
		return nil
	}

	if leftVal.IsNil() || rightVal.IsNil() {
		result.Diffs = append(result.Diffs, &Diff{Path: path, Left: left, Right: right})
		return nil
	}

	if leftVal.Pointer() != rightVal.Pointer() {
		result.Diffs = append(result.Diffs, &Diff{Path: path, Left: left, Right: right})
	}
	return nil
}

// ChannelHandler handles channel types
type ChannelHandler struct{}

func (h *ChannelHandler) CanHandle(typ reflect.Type) bool {
	return typ.Kind() == reflect.Chan
}

func (h *ChannelHandler) Compare(left, right any, path string, result *DiffResult, config *CompareConfig) error {
	if left != right {
		result.Diffs = append(result.Diffs, &Diff{Path: path, Left: left, Right: right})
	}
	return nil
}

// DefaultTypeHandlers returns a slice of default type handlers
func DefaultTypeHandlers() []TypeHandler {
	return []TypeHandler{
		&TimeHandler{},
		&InterfaceHandler{},
		&FunctionHandler{},
		&ChannelHandler{},
	}
}
