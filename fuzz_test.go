package godiff

import (
	"reflect"
	"testing"
	"time"
)

func FuzzCompareStrings(f *testing.F) {
	f.Add("hello", "world")
	f.Add("", "")
	f.Add("test", "")
	f.Add("", "test")
	f.Add("same", "same")
	f.Add("unicode: 你好", "unicode: 世界")
	f.Add("special\nchars\ttab", "special\rchars\ttab")

	f.Fuzz(func(t *testing.T, left, right string) {
		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed with strings %q and %q: %v", left, right, err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		if left == right && len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for identical strings %q, got %d", left, len(result.Diffs))
		}

		if left != right && len(result.Diffs) != 1 {
			t.Errorf("Expected exactly 1 difference for different strings %q vs %q, got %d", left, right, len(result.Diffs))
		}

		_ = result.String()
	})
}

func FuzzCompareInts(f *testing.F) {
	f.Add(int64(0), int64(0))
	f.Add(int64(1), int64(-1))
	f.Add(int64(9223372036854775807), int64(-9223372036854775808))
	f.Add(int64(42), int64(42))

	f.Fuzz(func(t *testing.T, left, right int64) {
		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed with ints %d and %d: %v", left, right, err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		if left == right && len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for identical ints %d, got %d", left, len(result.Diffs))
		}

		if left != right && len(result.Diffs) != 1 {
			t.Errorf("Expected exactly 1 difference for different ints %d vs %d, got %d", left, right, len(result.Diffs))
		}

		_ = result.String()
	})
}

func FuzzCompareSlices(f *testing.F) {
	f.Add([]byte{1, 2, 3}, []byte{1, 2, 3})
	f.Add([]byte{1, 2, 3}, []byte{3, 2, 1})
	f.Add([]byte{}, []byte{})
	f.Add([]byte{1}, []byte{})
	f.Add([]byte{}, []byte{1})

	f.Fuzz(func(t *testing.T, left, right []byte) {
		leftInts := make([]int, len(left))
		rightInts := make([]int, len(right))

		for i, b := range left {
			leftInts[i] = int(b)
		}
		for i, b := range right {
			rightInts[i] = int(b)
		}

		result, err := Compare(leftInts, rightInts)
		if err != nil {
			t.Fatalf("Compare failed with slices %v and %v: %v", leftInts, rightInts, err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		if reflect.DeepEqual(leftInts, rightInts) && len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for identical slices %v, got %d", leftInts, len(result.Diffs))
		}

		_ = result.String()
	})
}

type FuzzStruct struct {
	ID     int
	Name   string
	Value  float64
	Active bool
}

func FuzzCompareStructs(f *testing.F) {
	f.Add(int32(1), "test", float32(3.14), true, int32(1), "test", float32(3.14), true)
	f.Add(int32(1), "test", float32(3.14), true, int32(2), "different", float32(2.71), false)
	f.Add(int32(0), "", float32(0.0), false, int32(0), "", float32(0.0), false)

	f.Fuzz(func(t *testing.T, leftID int32, leftName string, leftValue float32, leftActive bool,
		rightID int32, rightName string, rightValue float32, rightActive bool) {

		if len(leftName) > 1000 {
			leftName = leftName[:1000]
		}
		if len(rightName) > 1000 {
			rightName = rightName[:1000]
		}

		leftStruct := FuzzStruct{
			ID:     int(leftID),
			Name:   leftName,
			Value:  float64(leftValue),
			Active: leftActive,
		}

		rightStruct := FuzzStruct{
			ID:     int(rightID),
			Name:   rightName,
			Value:  float64(rightValue),
			Active: rightActive,
		}

		result, err := Compare(leftStruct, rightStruct)
		if err != nil {
			t.Fatalf("Compare failed with structs %+v and %+v: %v", leftStruct, rightStruct, err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		if reflect.DeepEqual(leftStruct, rightStruct) && len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for identical structs %+v, got %d", leftStruct, len(result.Diffs))
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("String() method panicked: %v", r)
			}
		}()
		_ = result.String()
	})
}

func FuzzCompareMaps(f *testing.F) {
	f.Add("key1", "value1", "key2", "value2", "key1", "value1", "key2", "value2")
	f.Add("key1", "value1", "", "", "key2", "value2", "", "")
	f.Add("", "", "", "", "", "", "", "")

	f.Fuzz(func(t *testing.T, leftKey1, leftValue1, leftKey2, leftValue2, rightKey1, rightValue1, rightKey2, rightValue2 string) {
		leftKey1 = limitString(leftKey1, 50)
		leftValue1 = limitString(leftValue1, 50)
		leftKey2 = limitString(leftKey2, 50)
		leftValue2 = limitString(leftValue2, 50)
		rightKey1 = limitString(rightKey1, 50)
		rightValue1 = limitString(rightValue1, 50)
		rightKey2 = limitString(rightKey2, 50)
		rightValue2 = limitString(rightValue2, 50)

		leftMap := make(map[string]string)
		rightMap := make(map[string]string)

		if leftKey1 != "" {
			leftMap[leftKey1] = leftValue1
		}
		if leftKey2 != "" && leftKey2 != leftKey1 {
			leftMap[leftKey2] = leftValue2
		}

		if rightKey1 != "" {
			rightMap[rightKey1] = rightValue1
		}
		if rightKey2 != "" && rightKey2 != rightKey1 {
			rightMap[rightKey2] = rightValue2
		}

		result, err := Compare(leftMap, rightMap)
		if err != nil {
			t.Fatalf("Compare failed with maps %v and %v: %v", leftMap, rightMap, err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		if reflect.DeepEqual(leftMap, rightMap) && len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for identical maps %v, got %d", leftMap, len(result.Diffs))
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("String() method panicked with maps %v and %v: %v", leftMap, rightMap, r)
			}
		}()
		_ = result.String()
	})
}

func limitString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

func FuzzCompareWithNils(f *testing.F) {
	f.Add("test", true)
	f.Add("", false)
	f.Add("value", false)

	f.Fuzz(func(t *testing.T, value string, firstIsNil bool) {
		var left, right any

		if firstIsNil {
			left = nil
			right = value
		} else {
			left = value
			right = nil
		}

		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed with nil comparison %v and %v: %v", left, right, err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		if left == nil && right == nil && len(result.Diffs) != 0 {
			t.Error("Expected no differences for nil vs nil")
		}

		if (left == nil) != (right == nil) && len(result.Diffs) != 1 {
			t.Errorf("Expected exactly 1 difference for nil vs non-nil, got %d", len(result.Diffs))
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("String() method panicked with nil comparison: %v", r)
			}
		}()
		_ = result.String()
	})
}

func FuzzCompareWithOptions(f *testing.F) {
	f.Add("left", "right", true)
	f.Add("same", "same", false)

	f.Fuzz(func(t *testing.T, leftStr, rightStr string, ignoreSliceOrder bool) {
		if len(leftStr) > 100 {
			leftStr = leftStr[:100]
		}
		if len(rightStr) > 100 {
			rightStr = rightStr[:100]
		}

		var result *DiffResult
		var err error
		if ignoreSliceOrder {
			result, err = Compare(leftStr, rightStr, WithIgnoreSliceOrder())
		} else {
			result, err = Compare(leftStr, rightStr)
		}
		if err != nil {
			t.Fatalf("Compare failed with strings %q and %q: %v", leftStr, rightStr, err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		if leftStr == rightStr && len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for identical strings %q with any config, got %d", leftStr, len(result.Diffs))
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("String() method panicked with config comparison: %v", r)
			}
		}()
		_ = result.String()
	})
}

func FuzzCompareTime(f *testing.F) {
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	f.Add(int64(0), int64(0))
	f.Add(int64(3600), int64(7200))
	f.Add(int64(-3600), int64(3600))

	f.Fuzz(func(t *testing.T, leftSeconds, rightSeconds int64) {
		if leftSeconds > 1e10 {
			leftSeconds = 1e10
		}
		if leftSeconds < -1e10 {
			leftSeconds = -1e10
		}
		if rightSeconds > 1e10 {
			rightSeconds = 1e10
		}
		if rightSeconds < -1e10 {
			rightSeconds = -1e10
		}

		leftTime := baseTime.Add(time.Duration(leftSeconds) * time.Second)
		rightTime := baseTime.Add(time.Duration(rightSeconds) * time.Second)

		result, err := Compare(leftTime, rightTime)
		if err != nil {
			t.Fatalf("Compare failed with times %v and %v: %v", leftTime, rightTime, err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		if leftTime.Equal(rightTime) && len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for identical times %v, got %d", leftTime, len(result.Diffs))
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("String() method panicked with time comparison: %v", r)
			}
		}()
		_ = result.String()
	})
}
