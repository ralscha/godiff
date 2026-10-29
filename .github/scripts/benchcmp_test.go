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

func TestParseFileRetainsMemoryMetricsAfterThroughputAndCustomMetrics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bench.txt")
	content := "BenchmarkExample-8 100 100 ns/op 10 MB/s 4 comparisons/op 1000 B/op 100 allocs/op\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	results, err := parseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	result := results["BenchmarkExample-8"]
	if result.nsPerOp != 100 || result.bytesPerOp != 1000 || result.allocsPerOp != 100 {
		t.Fatalf("metrics lost after throughput/custom metrics: %+v", result)
	}
	if pctChange(10, result.bytesPerOp) <= 0.1 || pctChange(1, result.allocsPerOp) <= 0.1 {
		t.Fatal("memory regressions should exceed the default threshold")
	}
}

func TestParseBenchmarkLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want benchResult
		ok   bool
	}{
		{"time only", "BenchmarkExample-8 100 10 ns/op", benchResult{"BenchmarkExample-8", 10, 0, 0}, true},
		{"metric order", "BenchmarkExample-8 100 2 allocs/op 1e2 ns/op 20 B/op", benchResult{"BenchmarkExample-8", 100, 20, 2}, true},
		{"ordinary output", "PASS", benchResult{}, false},
		{"missing time", "BenchmarkExample-8 100 20 B/op", benchResult{}, false},
		{"invalid iteration count", "BenchmarkExample-8 invalid 10 ns/op", benchResult{}, false},
		{"missing unit", "BenchmarkExample-8 100 10 ns/op 20", benchResult{}, false},
		{"invalid memory metric", "BenchmarkExample-8 100 10 ns/op invalid B/op", benchResult{}, false},
		{"negative metric", "BenchmarkExample-8 100 -10 ns/op", benchResult{}, false},
		{"NaN metric", "BenchmarkExample-8 100 NaN ns/op", benchResult{}, false},
		{"infinite metric", "BenchmarkExample-8 100 +Inf ns/op", benchResult{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseBenchmarkLine(tt.line)
			if ok != tt.ok || (ok && got != tt.want) {
				t.Fatalf("parseBenchmarkLine = (%+v, %v), want (%+v, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}
