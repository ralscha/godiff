package godiff

import (
	"math"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestComplexNumbers(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same complex64", complex64(1 + 2i), complex64(1 + 2i), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different complex64", complex64(1 + 2i), complex64(3 + 4i), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != complex64(1+2i) {
						t.Errorf("Expected left value %v, got %v", complex64(1+2i), diff.Left)
					}
					if diff.Right != complex64(3+4i) {
						t.Errorf("Expected right value %v, got %v", complex64(3+4i), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for complex number comparison")
				}
			},
		},
		{
			"Same complex128", 1 + 2i, 1 + 2i, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different complex128", 1 + 2i, 3 + 4i, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != 1+2i {
						t.Errorf("Expected left value %v, got %v", 1+2i, diff.Left)
					}
					if diff.Right != 3+4i {
						t.Errorf("Expected right value %v, got %v", 3+4i, diff.Right)
					}
				} else {
					t.Error("Expected Diff type for complex number comparison")
				}
			},
		},
		{
			"Complex64 to complex128", complex64(1 + 2i), 1 + 2i, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != complex64(1+2i) {
						t.Errorf("Expected left value %v, got %v", complex64(1+2i), diff.Left)
					}
					if diff.Right != 1+2i {
						t.Errorf("Expected right value %v, got %v", 1+2i, diff.Right)
					}
				} else {
					t.Error("Expected Diff type for type mismatch comparison")
				}
			},
		},
		{
			"Zero complex", complex(0, 0), complex(0, 0), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d: %s", tt.expected, len(result.Diffs), result.String())
			}

			tt.validate(t, result)
		})
	}
}

func TestNumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same uint", uint(42), uint(42), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different uint", uint(42), uint(43), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != uint(42) {
						t.Errorf("Expected left value %v, got %v", uint(42), diff.Left)
					}
					if diff.Right != uint(43) {
						t.Errorf("Expected right value %v, got %v", uint(43), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for uint comparison")
				}
			},
		},
		{
			"Same uint8", uint8(255), uint8(255), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different uint8", uint8(255), uint8(254), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != uint8(255) {
						t.Errorf("Expected left value %v, got %v", uint8(255), diff.Left)
					}
					if diff.Right != uint8(254) {
						t.Errorf("Expected right value %v, got %v", uint8(254), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for uint8 comparison")
				}
			},
		},
		{
			"Same uint16", uint16(65535), uint16(65535), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different uint16", uint16(65535), uint16(65534), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != uint16(65535) {
						t.Errorf("Expected left value %v, got %v", uint16(65535), diff.Left)
					}
					if diff.Right != uint16(65534) {
						t.Errorf("Expected right value %v, got %v", uint16(65534), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for uint16 comparison")
				}
			},
		},
		{
			"Same uint32", uint32(4294967295), uint32(4294967295), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different uint32", uint32(4294967295), uint32(4294967294), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != uint32(4294967295) {
						t.Errorf("Expected left value %v, got %v", uint32(4294967295), diff.Left)
					}
					if diff.Right != uint32(4294967294) {
						t.Errorf("Expected right value %v, got %v", uint32(4294967294), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for uint32 comparison")
				}
			},
		},
		{
			"Same uint64", uint64(18446744073709551615), uint64(18446744073709551615), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different uint64", uint64(18446744073709551615), uint64(18446744073709551614), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != uint64(18446744073709551615) {
						t.Errorf("Expected left value %v, got %v", uint64(18446744073709551615), diff.Left)
					}
					if diff.Right != uint64(18446744073709551614) {
						t.Errorf("Expected right value %v, got %v", uint64(18446744073709551614), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for uint64 comparison")
				}
			},
		},

		{
			"Same int8", int8(-128), int8(-128), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different int8", int8(-128), int8(127), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != int8(-128) {
						t.Errorf("Expected left value %v, got %v", int8(-128), diff.Left)
					}
					if diff.Right != int8(127) {
						t.Errorf("Expected right value %v, got %v", int8(127), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for int8 comparison")
				}
			},
		},
		{
			"Same int16", int16(-32768), int16(-32768), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different int16", int16(-32768), int16(32767), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != int16(-32768) {
						t.Errorf("Expected left value %v, got %v", int16(-32768), diff.Left)
					}
					if diff.Right != int16(32767) {
						t.Errorf("Expected right value %v, got %v", int16(32767), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for int16 comparison")
				}
			},
		},
		{
			"Same int32", int32(-2147483648), int32(-2147483648), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different int32", int32(-2147483648), int32(2147483647), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != int32(-2147483648) {
						t.Errorf("Expected left value %v, got %v", int32(-2147483648), diff.Left)
					}
					if diff.Right != int32(2147483647) {
						t.Errorf("Expected right value %v, got %v", int32(2147483647), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for int32 comparison")
				}
			},
		},

		{
			"Same float32", float32(3.14159), float32(3.14159), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different float32", float32(3.14159), float32(2.71828), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != float32(3.14159) {
						t.Errorf("Expected left value %v, got %v", float32(3.14159), diff.Left)
					}
					if diff.Right != float32(2.71828) {
						t.Errorf("Expected right value %v, got %v", float32(2.71828), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for float32 comparison")
				}
			},
		},
		{
			"Same float64", 3.141592653589793, 3.141592653589793, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different float64", 3.141592653589793, 2.718281828459045, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != 3.141592653589793 {
						t.Errorf("Expected left value %v, got %v", 3.141592653589793, diff.Left)
					}
					if diff.Right != 2.718281828459045 {
						t.Errorf("Expected right value %v, got %v", 2.718281828459045, diff.Right)
					}
				} else {
					t.Error("Expected Diff type for float64 comparison")
				}
			},
		},
		{
			"Float32 NaN", float32(math.NaN()), float32(math.NaN()), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}

					if _, ok := diff.Left.(float32); !ok {
						t.Errorf("Expected left value to be float32, got %T", diff.Left)
					}
					if _, ok := diff.Right.(float32); !ok {
						t.Errorf("Expected right value to be float32, got %T", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for float32 NaN comparison")
				}
			},
		},
		{
			"Float64 NaN", math.NaN(), math.NaN(), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}

					if _, ok := diff.Left.(float64); !ok {
						t.Errorf("Expected left value to be float64, got %T", diff.Left)
					}
					if _, ok := diff.Right.(float64); !ok {
						t.Errorf("Expected right value to be float64, got %T", diff.Right)
					}
				} else {
					t.Error("Expected Diff type for float64 NaN comparison")
				}
			},
		},
		{
			"int vs uint", 42, uint(42), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != 42 {
						t.Errorf("Expected left value %v, got %v", 42, diff.Left)
					}
					if diff.Right != uint(42) {
						t.Errorf("Expected right value %v, got %v", uint(42), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for int vs uint comparison")
				}
			},
		},
		{
			"int32 vs int64", int32(42), int64(42), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != int32(42) {
						t.Errorf("Expected left value %v, got %v", int32(42), diff.Left)
					}
					if diff.Right != int64(42) {
						t.Errorf("Expected right value %v, got %v", int64(42), diff.Right)
					}
				} else {
					t.Error("Expected Diff type for int32 vs int64 comparison")
				}
			},
		},
		{
			"float32 vs float64", float32(3.14), 3.14, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != float32(3.14) {
						t.Errorf("Expected left value %v, got %v", float32(3.14), diff.Left)
					}
					if diff.Right != 3.14 {
						t.Errorf("Expected right value %v, got %v", 3.14, diff.Right)
					}
				} else {
					t.Error("Expected Diff type for float32 vs float64 comparison")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d: %s", tt.expected, len(result.Diffs), result.String())
			}

			tt.validate(t, result)
		})
	}
}

func TestArrayTypes(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same int array", [3]int{1, 2, 3}, [3]int{1, 2, 3}, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different int array", [3]int{1, 2, 3}, [3]int{1, 2, 4}, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if sliceDiff, ok := result.Diffs[0].(*SliceDiff); ok {
					if sliceDiff.Index != 2 {
						t.Errorf("Expected diff at index 2, got %d", sliceDiff.Index)
					}
					if sliceDiff.Left != 3 {
						t.Errorf("Expected left value 3, got %v", sliceDiff.Left)
					}
					if sliceDiff.Right != 4 {
						t.Errorf("Expected right value 4, got %v", sliceDiff.Right)
					}
					if sliceDiff.ChangeType != "UPDATED" {
						t.Errorf("Expected UPDATED change type, got %s", sliceDiff.ChangeType)
					}
				} else {
					t.Error("Expected SliceDiff type for array comparison")
				}
			},
		},
		{
			"Same string array", [2]string{"hello", "world"}, [2]string{"hello", "world"}, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different string array", [2]string{"hello", "world"}, [2]string{"hello", "universe"}, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if sliceDiff, ok := result.Diffs[0].(*SliceDiff); ok {
					if sliceDiff.Index != 1 {
						t.Errorf("Expected diff at index 1, got %d", sliceDiff.Index)
					}
					if sliceDiff.Left != "world" {
						t.Errorf("Expected left value 'world', got %v", sliceDiff.Left)
					}
					if sliceDiff.Right != "universe" {
						t.Errorf("Expected right value 'universe', got %v", sliceDiff.Right)
					}
					if sliceDiff.ChangeType != "UPDATED" {
						t.Errorf("Expected UPDATED change type, got %s", sliceDiff.ChangeType)
					}
				} else {
					t.Error("Expected SliceDiff type for array comparison")
				}
			},
		},
		{
			"Different size arrays", [2]int{1, 2}, [3]int{1, 2, 3}, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left == nil || diff.Right == nil {
						t.Errorf("Expected both left and right values to be non-nil for type mismatch")
					}
				} else {
					t.Error("Expected Diff type for type mismatch comparison")
				}
			},
		},
		{
			"Empty arrays", [0]int{}, [0]int{}, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d: %s", tt.expected, len(result.Diffs), result.String())
			}

			tt.validate(t, result)
		})
	}
}

func TestArrayVsSlice(t *testing.T) {
	array := [3]int{1, 2, 3}
	slice := []int{1, 2, 3}

	result, err := Compare(array, slice)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) != 1 {
		t.Errorf("Expected 1 difference between array and slice, got %d: %s", len(result.Diffs), result.String())
	}

	if len(result.Diffs) > 0 {
		if diff, ok := result.Diffs[0].(*Diff); ok {
			if diff.Path != "" {
				t.Errorf("Expected empty path for root diff, got %s", diff.Path)
			}
			if diff.Left == nil || diff.Right == nil {
				t.Errorf("Expected both left and right values to be non-nil for type mismatch")
			}
		} else {
			t.Errorf("Expected Diff type for array vs slice comparison, got %T", result.Diffs[0])
		}
	}
}

func TestUintptrType(t *testing.T) {
	var x, y = 42, 42
	ptr1 := uintptr(unsafe.Pointer(&x))
	ptr2 := uintptr(unsafe.Pointer(&y))

	tests := []struct {
		name     string
		left     uintptr
		right    uintptr
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same uintptr", ptr1, ptr1, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different uintptr", ptr1, ptr2, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left != ptr1 {
						t.Errorf("Expected left value %v, got %v", ptr1, diff.Left)
					}
					if diff.Right != ptr2 {
						t.Errorf("Expected right value %v, got %v", ptr2, diff.Right)
					}
				} else {
					t.Error("Expected Diff type for uintptr comparison")
				}
			},
		},
		{
			"Zero uintptr", uintptr(0), uintptr(0), 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d: %s", tt.expected, len(result.Diffs), result.String())
			}

			tt.validate(t, result)
		})
	}
}

func TestUnsafePointer(t *testing.T) {
	var x, y = 42, 42
	ptr1 := unsafe.Pointer(&x)
	ptr2 := unsafe.Pointer(&y)

	tests := []struct {
		name     string
		left     unsafe.Pointer
		right    unsafe.Pointer
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same unsafe.Pointer", ptr1, ptr1, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different unsafe.Pointer", ptr1, ptr2, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}
					if diff.Left == nil || diff.Right == nil {
						t.Errorf("Expected both left and right values to be non-nil for pointer comparison")
					}
				} else {
					t.Error("Expected Diff type for unsafe.Pointer comparison")
				}
			},
		},
		{
			"Nil unsafe.Pointer", nil, nil, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Nil vs non-nil", nil, ptr1, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}

					if diff.Left == nil && diff.Right == nil {
						t.Errorf("Expected one nil and one non-nil value, got both nil")
					}
				} else {
					t.Error("Expected Diff type for nil vs non-nil pointer comparison")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d: %s", tt.expected, len(result.Diffs), result.String())
			}

			tt.validate(t, result)
		})
	}
}

func TestChannelTypes(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	var nilCh chan int

	tests := []struct {
		name     string
		left     any
		right    any
		expected int
		validate func(t *testing.T, result *DiffResult)
	}{
		{
			"Same channel", ch1, ch1, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Different channels", ch1, ch2, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}

					if diff.Left == nil || diff.Right == nil {
						t.Errorf("Expected both left and right values to be non-nil for channel comparison")
					}
				} else {
					t.Error("Expected Diff type for channel comparison")
				}
			},
		},
		{
			"Nil channels", nilCh, nilCh, 0,
			func(t *testing.T, result *DiffResult) {

			},
		},
		{
			"Nil vs non-nil channel", nilCh, ch1, 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}

					if diff.Left == nil && diff.Right == nil {
						t.Errorf("Expected one nil and one non-nil value, got both nil")
					}
				} else {
					t.Error("Expected Diff type for nil vs non-nil channel comparison")
				}
			},
		},
		{
			"Different channel types", make(chan int), make(chan string), 1,
			func(t *testing.T, result *DiffResult) {
				if len(result.Diffs) != 1 {
					t.Fatalf("Expected 1 diff, got %d", len(result.Diffs))
				}
				if diff, ok := result.Diffs[0].(*Diff); ok {
					if diff.Path != "" {
						t.Errorf("Expected empty path for root diff, got %s", diff.Path)
					}

					if diff.Left == nil || diff.Right == nil {
						t.Errorf("Expected both left and right values to be non-nil for type mismatch")
					}
				} else {
					t.Error("Expected Diff type for type mismatch comparison")
				}
			},
		},
	}

	defer close(ch1)
	defer close(ch2)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d: %s", tt.expected, len(result.Diffs), result.String())
			}

			tt.validate(t, result)
		})
	}
}

func TestZeroValues(t *testing.T) {
	type TestStruct struct {
		IntVal    int
		StringVal string
		BoolVal   bool
		SliceVal  []int
		MapVal    map[string]int
		PtrVal    *int
	}

	zero1 := TestStruct{}
	zero2 := TestStruct{}

	result, err := Compare(zero1, zero2)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) != 0 {
		t.Errorf("Expected no differences between zero values, got %d: %s", len(result.Diffs), result.String())
	}

	nonZero := TestStruct{
		IntVal:    42,
		StringVal: "test",
		BoolVal:   true,
		SliceVal:  []int{1, 2, 3},
		MapVal:    map[string]int{"key": 1},
		PtrVal:    &[]int{42}[0],
	}

	result, err = Compare(zero1, nonZero)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(result.Diffs) == 0 {
		t.Error("Expected differences between zero and non-zero values")
	}
}

func TestInterfaceHandler2(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected int
	}{
		{
			name:     "both nil interfaces",
			left:     (any)(nil),
			right:    (any)(nil),
			expected: 0,
		},
		{
			name:     "left nil interface",
			left:     (any)(nil),
			right:    "hello",
			expected: 1,
		},
		{
			name:     "right nil interface",
			left:     "hello",
			right:    (any)(nil),
			expected: 1,
		},
		{
			name:     "different interface values",
			left:     any("hello"),
			right:    any(42),
			expected: 1,
		},
		{
			name:     "same interface values",
			left:     any("hello"),
			right:    any("hello"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right, WithTypeHandlers(DefaultTypeHandlers()))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Diffs) != tt.expected {
				t.Errorf("expected %d differences, got %d", tt.expected, len(result.Diffs))
			}
		})
	}
}

func TestStructWithSliceTags(t *testing.T) {
	type StructWithIgnoreOrder struct {
		Name  string
		Items []int `diff:"ignoreOrder"`
	}

	t.Run("ignoreOrder tag", func(t *testing.T) {
		left := StructWithIgnoreOrder{
			Name:  "test",
			Items: []int{1, 2, 3},
		}

		right := StructWithIgnoreOrder{
			Name:  "test",
			Items: []int{3, 2, 1},
		}

		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Diffs) != 0 {
			t.Errorf("expected no differences, got %d", len(result.Diffs))
		}
	})

}

func TestHasExactDiffTagEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		diffTag  string
		tagValue string
		expected bool
	}{
		{
			name:     "empty tag",
			diffTag:  "",
			tagValue: "ignore",
			expected: false,
		},
		{
			name:     "single matching tag",
			diffTag:  "ignore",
			tagValue: "ignore",
			expected: true,
		},
		{
			name:     "multiple tags with match",
			diffTag:  "ignore,ignoreOrder",
			tagValue: "ignoreOrder",
			expected: true,
		},
		{
			name:     "multiple tags without match",
			diffTag:  "ignore,ignoreOrder",
			tagValue: "notfound",
			expected: false,
		},
		{
			name:     "tag with spaces",
			diffTag:  "ignore , ignoreOrder",
			tagValue: "ignore",
			expected: true,
		},
		{
			name:     "partial match should not work",
			diffTag:  "ignoreOrder",
			tagValue: "ignore",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasDiffTag(tt.diffTag, tt.tagValue)
			if result != tt.expected {
				t.Errorf("hasDiffTag(%q, %q) = %v, expected %v", tt.diffTag, tt.tagValue, result, tt.expected)
			}
		})
	}
}

func TestInterfaceHandlerDirectUsage(t *testing.T) {
	handler := &InterfaceHandler{}

	var nilInterface any = nil
	var stringInterface any = "hello"
	var stringInterface2 any = "world"
	var intInterface any = 42

	tests := []struct {
		name         string
		left         *any
		right        *any
		expectedDiff int
	}{
		{
			name:         "both nil interface pointers",
			left:         &nilInterface,
			right:        &nilInterface,
			expectedDiff: 0,
		},
		{
			name:         "left nil interface, right has value",
			left:         &nilInterface,
			right:        &stringInterface,
			expectedDiff: 1,
		},
		{
			name:         "right nil interface, left has value",
			left:         &stringInterface,
			right:        &nilInterface,
			expectedDiff: 1,
		},
		{
			name:         "both valid interfaces with same values",
			left:         &stringInterface,
			right:        &stringInterface,
			expectedDiff: 0,
		},
		{
			name:         "both valid interfaces with different values",
			left:         &stringInterface,
			right:        &stringInterface2,
			expectedDiff: 1,
		},
		{
			name:         "different types in interfaces",
			left:         &stringInterface,
			right:        &intInterface,
			expectedDiff: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &DiffResult{}
			config := DefaultCompareConfig()

			err := handler.Compare(tt.left, tt.right, "test", result, config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Diffs) != tt.expectedDiff {
				t.Errorf("expected %d differences, got %d", tt.expectedDiff, len(result.Diffs))
			}
		})
	}
}

func TestComparePointersAdditionalEdgeCases(t *testing.T) {
	type TestStruct struct {
		Value int
	}

	tests := []struct {
		name     string
		left     any
		right    any
		expected int
	}{
		{
			name:     "nil pointers of same type",
			left:     (*int)(nil),
			right:    (*int)(nil),
			expected: 0,
		},
		{
			name:     "nil pointer vs non-nil pointer",
			left:     (*int)(nil),
			right:    func() *int { x := 42; return &x }(),
			expected: 1,
		},
		{
			name:     "non-nil pointer vs nil pointer",
			left:     func() *int { x := 42; return &x }(),
			right:    (*int)(nil),
			expected: 1,
		},
		{
			name:     "pointers to same value",
			left:     func() *int { x := 42; return &x }(),
			right:    func() *int { x := 42; return &x }(),
			expected: 0,
		},
		{
			name:     "pointers to different values",
			left:     func() *int { x := 42; return &x }(),
			right:    func() *int { x := 43; return &x }(),
			expected: 1,
		},
		{
			name:     "circular reference detection",
			left:     func() *TestStruct { s := &TestStruct{Value: 1}; return s }(),
			right:    func() *TestStruct { s := &TestStruct{Value: 1}; return s }(),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &DiffResult{}
			leftVal := reflect.ValueOf(tt.left)
			rightVal := reflect.ValueOf(tt.right)

			config := DefaultCompareConfig()
			err := comparePointers("test", leftVal, rightVal, result, config)
			if err != nil {
				t.Fatalf("comparePointers failed: %v", err)
			}

			if len(result.Diffs) != tt.expected {
				t.Errorf("Expected %d differences, got %d", tt.expected, len(result.Diffs))
			}
		})
	}
}

func TestTimeHandlerEdgeCases(t *testing.T) {
	handler := &TimeHandler{}

	t.Run("invalid time values", func(t *testing.T) {
		result := &DiffResult{}
		err := handler.Compare("not a time", time.Now(), "test", result, DefaultCompareConfig())
		if err == nil {
			t.Error("Expected error for non-time values")
		}
	})

	t.Run("one valid one invalid time", func(t *testing.T) {
		result := &DiffResult{}
		err := handler.Compare(time.Now(), "not a time", "test", result, DefaultCompareConfig())
		if err == nil {
			t.Error("Expected error for mixed time/non-time values")
		}
	})
}

func TestInterfaceHandlerEdgeCases(t *testing.T) {
	handler := &InterfaceHandler{}

	t.Run("nil interface values", func(t *testing.T) {
		result := &DiffResult{}
		err := handler.Compare(nil, nil, "test", result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences, got %d", len(result.Diffs))
		}
	})

	t.Run("one nil interface", func(t *testing.T) {
		result := &DiffResult{}
		err := handler.Compare(nil, "value", "test", result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference, got %d", len(result.Diffs))
		}
	})

}

func TestFunctionHandlerEdgeCases(t *testing.T) {
	handler := &FunctionHandler{}

	t.Run("nil functions", func(t *testing.T) {
		result := &DiffResult{}
		var f1, f2 func()
		err := handler.Compare(f1, f2, "test", result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for nil functions, got %d", len(result.Diffs))
		}
	})

	t.Run("one nil function", func(t *testing.T) {
		result := &DiffResult{}
		f1 := func() {}
		var f2 func()
		err := handler.Compare(f1, f2, "test", result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference for function vs nil, got %d", len(result.Diffs))
		}
	})

	t.Run("different functions", func(t *testing.T) {
		result := &DiffResult{}
		f1 := func() {}
		f2 := func() {}
		err := handler.Compare(f1, f2, "test", result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference for different functions, got %d", len(result.Diffs))
		}
	})
}

func TestChannelHandlerEdgeCases(t *testing.T) {
	handler := &ChannelHandler{}

	t.Run("same channel", func(t *testing.T) {
		result := &DiffResult{}
		ch := make(chan int)
		err := handler.Compare(ch, ch, "test", result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for same channel, got %d", len(result.Diffs))
		}
	})

	t.Run("different channels", func(t *testing.T) {
		result := &DiffResult{}
		ch1 := make(chan int)
		ch2 := make(chan int)
		err := handler.Compare(ch1, ch2, "test", result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference for different channels, got %d", len(result.Diffs))
		}
	})

	t.Run("nil channels", func(t *testing.T) {
		result := &DiffResult{}
		var ch1, ch2 chan int
		err := handler.Compare(ch1, ch2, "test", result, DefaultCompareConfig())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences for nil channels, got %d", len(result.Diffs))
		}
	})
}

func TestCompareNumericValues(t *testing.T) {
	tests := []struct {
		name         string
		left         any
		right        any
		expectedDiff int
	}{
		// Same type comparisons (should work without CompareNumericValues)
		{"int vs int same", int(42), int(42), 0},
		{"int vs int different", int(42), int(43), 1},

		// Cross-type signed integer comparisons
		{"int vs int32 same value", int(42), int32(42), 0},
		{"int vs int64 same value", int(42), int64(42), 0},
		{"int8 vs int16 same value", int8(42), int16(42), 0},
		{"int16 vs int32 same value", int16(42), int32(42), 0},
		{"int32 vs int64 same value", int32(42), int64(42), 0},
		{"int vs int64 different value", int(42), int64(43), 1},

		// Cross-type unsigned integer comparisons
		{"uint vs uint32 same value", uint(42), uint32(42), 0},
		{"uint vs uint64 same value", uint(42), uint64(42), 0},
		{"uint8 vs uint16 same value", uint8(42), uint16(42), 0},
		{"uint16 vs uint32 same value", uint16(42), uint32(42), 0},
		{"uint32 vs uint64 same value", uint32(42), uint64(42), 0},
		{"uint vs uint64 different value", uint(42), uint64(43), 1},

		// Mixed signed/unsigned integer comparisons
		{"int vs uint same positive value", int(42), uint(42), 0},
		{"int32 vs uint32 same value", int32(42), uint32(42), 0},
		{"int64 vs uint64 same value", int64(42), uint64(42), 0},
		{"negative int vs uint", int(-1), uint(1), 1},
		{"int vs uint different value", int(42), uint(43), 1},

		// Float comparisons
		{"float32 vs float64 same value", float32(3.5), float64(3.5), 0},
		{"float32 vs float64 different value", float32(3.5), float64(3.6), 1},

		// Integer vs float comparisons
		{"int vs float64 same value", int(42), float64(42.0), 0},
		{"int vs float64 different value", int(42), float64(42.5), 1},
		{"int64 vs float64 same value", int64(100), float64(100.0), 0},
		{"uint vs float64 same value", uint(42), float64(42.0), 0},

		// Complex number comparisons
		{"complex64 vs complex128 same value", complex64(1 + 2i), complex128(1 + 2i), 0},
		{"complex64 vs complex128 different value", complex64(1 + 2i), complex128(1 + 3i), 1},

		// Edge cases
		{"zero int vs zero uint", int(0), uint(0), 0},
		{"zero int vs zero float64", int(0), float64(0.0), 0},
		{"max int8 vs int64", int8(127), int64(127), 0},
		{"max uint8 vs uint64", uint8(255), uint64(255), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.left, tt.right, WithCompareNumericValues())
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expectedDiff {
				t.Errorf("Expected %d differences, got %d: %s", tt.expectedDiff, len(result.Diffs), result.String())
			}
		})
	}
}

func TestCompareNumericValuesDisabled(t *testing.T) {
	// When CompareNumericValues is false (default), different numeric types should be considered different
	tests := []struct {
		name         string
		left         any
		right        any
		expectedDiff int
	}{
		{"int vs int64 same value", int(42), int64(42), 1},
		{"int vs uint same value", int(42), uint(42), 1},
		{"float32 vs float64 same value", float32(3.5), float64(3.5), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// CompareNumericValues is false by default when using Compare without options

			result, err := Compare(tt.left, tt.right)
			if err != nil {
				t.Fatalf("Compare failed: %v", err)
			}

			if len(result.Diffs) != tt.expectedDiff {
				t.Errorf("Expected %d differences, got %d: %s", tt.expectedDiff, len(result.Diffs), result.String())
			}
		})
	}
}

func TestCompareNumericValuesInMaps(t *testing.T) {
	t.Run("maps with different numeric types but same values", func(t *testing.T) {
		left := map[string]any{
			"count":  int(42),
			"price":  float32(19.99),
			"amount": int64(100),
		}
		right := map[string]any{
			"count":  int64(42),
			"price":  float64(19.99),
			"amount": uint(100),
		}

		result, err := Compare(left, right, WithCompareNumericValues())
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}

		// float32(19.99) and float64(19.99) may have precision differences
		// so we check that at least count and amount are equal
		// The exact number of diffs depends on float precision
		if len(result.Diffs) > 1 {
			t.Errorf("Expected at most 1 difference (due to float precision), got %d: %s", len(result.Diffs), result.String())
		}
	})

	t.Run("maps with different numeric types without CompareNumericValues", func(t *testing.T) {
		left := map[string]any{
			"count": int(42),
		}
		right := map[string]any{
			"count": int64(42),
		}

		// CompareNumericValues is false by default when using Compare without options

		result, err := Compare(left, right)
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}

		if len(result.Diffs) != 1 {
			t.Errorf("Expected 1 difference, got %d: %s", len(result.Diffs), result.String())
		}
	})
}

func TestCompareNumericValuesInStructs(t *testing.T) {
	type Container struct {
		Value any
	}

	t.Run("structs with different numeric types but same values", func(t *testing.T) {
		left := Container{Value: int(42)}
		right := Container{Value: int64(42)}

		result, err := Compare(left, right, WithCompareNumericValues())
		if err != nil {
			t.Fatalf("Compare failed: %v", err)
		}

		if len(result.Diffs) != 0 {
			t.Errorf("Expected no differences, got %d: %s", len(result.Diffs), result.String())
		}
	})
}
