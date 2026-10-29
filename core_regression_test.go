package godiff

import (
	"errors"
	"math"
	"math/big"
	"reflect"
	"testing"
)

func TestSharedReferencesHonorNestedComparators(t *testing.T) {
	value := 42
	for name, shared := range map[string]any{
		"pointer": &value,
		"slice":   []int{value},
		"map":     map[string]int{"value": value},
	} {
		t.Run(name, func(t *testing.T) {
			for _, comparatorError := range []error{nil, errors.New("comparison failed")} {
				calls := 0
				result, err := Compare(shared, shared, WithCustomComparators(map[reflect.Type]func(any, any, *CompareConfig) (bool, error){
					reflect.TypeFor[int](): func(_, _ any, _ *CompareConfig) (bool, error) {
						calls++
						return false, comparatorError
					},
				}))
				if calls != 1 || !errors.Is(err, comparatorError) {
					t.Fatalf("calls = %d, error = %v; want one call and %v", calls, err, comparatorError)
				}
				if err == nil && result.Count() != 1 {
					t.Fatalf("nested comparator difference was lost: %v", result)
				}
			}
		})
	}
}

type unequalIntHandler struct{}

func (*unequalIntHandler) CanHandle(typ reflect.Type) bool {
	return typ == reflect.TypeFor[int]()
}

func (*unequalIntHandler) Compare(left, right any, path string, result *DiffResult, _ *CompareConfig) error {
	result.AddDiff(path, left, right)
	return nil
}

func TestSharedReferencesHonorNestedHandlers(t *testing.T) {
	type node struct {
		Value int
		Next  *node
	}
	shared := &node{Value: 42}
	shared.Next = shared
	result, err := Compare(shared, shared, WithAdditionalTypeHandlers(&unequalIntHandler{}))
	if err != nil || result.Count() != 1 {
		t.Fatalf("shared cycle should visit the nested handler once: result=%v err=%v", result, err)
	}
}

func TestFunctionsCompareConsistentlyWithoutHandlers(t *testing.T) {
	fn := func() {}
	for name, values := range map[string][2]any{
		"root":  {fn, fn},
		"slice": {[]func(){fn}, []func(){fn}},
		"map":   {map[string]func(){"fn": fn}, map[string]func(){"fn": fn}},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := Compare(values[0], values[1], WithTypeHandlers(nil))
			if err != nil || result.HasDifferences() {
				t.Fatalf("function identity should agree at every level: result=%v err=%v", result, err)
			}
		})
	}
}

func TestTypedNilPointersCompareConsistently(t *testing.T) {
	var left *int
	var right *string
	for name, values := range map[string][2]any{
		"root":  {left, right},
		"slice": {[]any{left}, []any{right}},
		"array": {[1]any{left}, [1]any{right}},
		"map":   {map[string]any{"value": left}, map[string]any{"value": right}},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := Compare(values[0], values[1])
			if err != nil || result.HasDifferences() {
				t.Fatalf("nil pointers should compare equally at every level: result=%v err=%v", result, err)
			}
		})
	}
}

func TestIntegerFloatComparisonBoundaries(t *testing.T) {
	integers := []any{
		int64(math.MinInt64), int64(math.MinInt64 + 1), int64(-1), int64(0), int64(1),
		int64(1 << 53), int64(1<<53 + 1), int64(math.MaxInt64),
		uint64(0), uint64(1 << 63), uint64(math.MaxUint64),
	}
	floats := []float64{
		math.Inf(-1), -0x1p63, math.Nextafter(-0x1p63, math.Inf(-1)),
		math.Nextafter(-0x1p63, 0), -1, math.Copysign(0, -1), 0, 0.5, 1,
		math.SmallestNonzeroFloat64, 0x1p53, math.Nextafter(0x1p63, 0),
		0x1p63, math.Nextafter(0x1p64, 0), 0x1p64, math.Inf(1), math.NaN(),
	}
	for _, integer := range integers {
		for _, floating := range floats {
			assertIntegerFloatComparison(t, integer, floating)
		}
	}
}

func assertIntegerFloatComparison(t *testing.T, integer any, floating float64) {
	t.Helper()
	var integerRat big.Rat
	value := reflect.ValueOf(integer)
	if value.Kind() == reflect.Int64 {
		integerRat.SetInt64(value.Int())
	} else {
		integerRat.SetUint64(value.Uint())
	}
	floatRat := new(big.Rat).SetFloat64(floating)
	equal := floatRat != nil && integerRat.Cmp(floatRat) == 0
	for _, pair := range [][2]any{{integer, floating}, {floating, integer}} {
		result, err := Compare(pair[0], pair[1], WithCompareNumericValues())
		if err != nil || result.HasDifferences() == equal {
			t.Fatalf("Compare(%v (%T), %v (%T)): result=%v err=%v, want equal=%v", pair[0], pair[0], pair[1], pair[1], result, err, equal)
		}
	}
}

func FuzzCompareIntegerFloat(f *testing.F) {
	for _, floating := range []float64{0, -1, 0.5, 0x1p53, -0x1p63, 0x1p63, 0x1p64, math.Inf(1), math.NaN()} {
		f.Add(uint64(math.MaxUint64), math.Float64bits(floating))
	}
	f.Fuzz(func(t *testing.T, integer, floatBits uint64) {
		for _, value := range []any{integer, int64(integer)} {
			assertIntegerFloatComparison(t, value, math.Float64frombits(floatBits))
		}
		// Exercise equality and rounding boundaries as well as arbitrary pairs.
		assertIntegerFloatComparison(t, integer, float64(integer))
		assertIntegerFloatComparison(t, int64(integer), float64(int64(integer)))
	})
}

func BenchmarkCompareIntegerFloat(b *testing.B) {
	for b.Loop() {
		_, _ = Compare(int64(1<<53), float64(1<<53), WithCompareNumericValues())
	}
}
