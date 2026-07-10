package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFileAcceptsDecimalBenchmarkValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bench.txt")
	content := "BenchmarkCompareBasicTypes-8  2234391  512.2 ns/op  296 B/op  6 allocs/op\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write benchmark fixture: %v", err)
	}

	results, err := parseFile(path)
	if err != nil {
		t.Fatalf("parseFile failed: %v", err)
	}

	result, ok := results["BenchmarkCompareBasicTypes-8"]
	if !ok {
		t.Fatalf("expected benchmark result to be parsed")
	}
	if result.nsPerOp != 512.2 || result.bytesPerOp != 296 || result.allocsPerOp != 6 {
		t.Fatalf("unexpected parsed result: %+v", result)
	}
}

func TestParseFileUsesMedianOfRepeatedRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bench.txt")
	content := "" +
		"BenchmarkCompareStructs-8  100  100 ns/op  10 B/op  1 allocs/op\n" +
		"BenchmarkCompareStructs-8  100  500 ns/op  50 B/op  5 allocs/op\n" +
		"BenchmarkCompareStructs-8  100  101 ns/op  11 B/op  2 allocs/op\n" +
		"BenchmarkCompareStructs-8  100  103 ns/op  13 B/op  4 allocs/op\n" +
		"BenchmarkCompareStructs-8  100  102 ns/op  12 B/op  3 allocs/op\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write benchmark fixture: %v", err)
	}

	results, err := parseFile(path)
	if err != nil {
		t.Fatalf("parseFile failed: %v", err)
	}

	result := results["BenchmarkCompareStructs-8"]
	if result.nsPerOp != 102 || result.bytesPerOp != 12 || result.allocsPerOp != 3 {
		t.Fatalf("expected medians, got: %+v", result)
	}
}
