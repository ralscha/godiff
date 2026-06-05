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
