package main

// benchcmp.go compares benchmark results of the current code vs a baseline (e.g. last tag).
// It fails (non‑zero exit code) if any benchmark shows a regression greater than the
// configured thresholds (time, bytes, allocs). It expects input files produced by
// "go test -bench=. -benchmem -run=^$ -count=N ./...".
//
// Usage:
//   go run .github/scripts/benchcmp.go \
//     -base benchmark_base.txt -current benchmark_current.txt \
//     -time 0.10 -bytes 0.10 -allocs 0.10
//
// Threshold flags represent allowed relative increase (e.g. 0.10 == 10%).
// If a benchmark appears only in current, it's ignored (treated as new).
// If only in base, also ignored (removed benchmark).

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

type benchResult struct {
	name        string
	nsPerOp     float64
	bytesPerOp  float64
	allocsPerOp float64
}

// Benchmark output contains value/unit pairs. Throughput and custom metrics
// can appear between ns/op and the memory metrics, so parse each unit by name.
func parseBenchmarkLine(line string) (benchResult, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 || len(fields)%2 != 0 || !strings.HasPrefix(fields[0], "Benchmark") {
		return benchResult{}, false
	}
	if _, err := strconv.ParseUint(fields[1], 10, 64); err != nil {
		return benchResult{}, false
	}

	result := benchResult{name: fields[0]}
	hasTime := false
	for i := 2; i < len(fields); i += 2 {
		var metric *float64
		switch fields[i+1] {
		case "ns/op":
			metric = &result.nsPerOp
			hasTime = true
		case "B/op":
			metric = &result.bytesPerOp
		case "allocs/op":
			metric = &result.allocsPerOp
		default:
			continue
		}
		value, err := strconv.ParseFloat(fields[i], 64)
		if err != nil || value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			return benchResult{}, false
		}
		*metric = value
	}
	return result, hasTime
}

func parseFile(path string) (map[string]benchResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	samples := make(map[string][]benchResult)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		result, ok := parseBenchmarkLine(scanner.Text())
		if !ok {
			continue
		}
		samples[result.name] = append(samples[result.name], result)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(samples) == 0 {
		return nil, errors.New("no benchmarks parsed from " + path)
	}

	// A benchmark file commonly contains multiple runs (the workflow uses
	// -count=5). Use their medians instead of whichever noisy sample happened
	// to be printed last.
	results := make(map[string]benchResult, len(samples))
	for name, values := range samples {
		nsPerOp := make([]float64, len(values))
		bytesPerOp := make([]float64, len(values))
		allocsPerOp := make([]float64, len(values))
		for i, value := range values {
			nsPerOp[i] = value.nsPerOp
			bytesPerOp[i] = value.bytesPerOp
			allocsPerOp[i] = value.allocsPerOp
		}
		results[name] = benchResult{
			name:        name,
			nsPerOp:     median(nsPerOp),
			bytesPerOp:  median(bytesPerOp),
			allocsPerOp: median(allocsPerOp),
		}
	}
	return results, nil
}

func median(values []float64) float64 {
	slices.Sort(values)
	middle := len(values) / 2
	if len(values)%2 == 1 {
		return values[middle]
	}
	return (values[middle-1] + values[middle]) / 2
}

func pctChange(base, current float64) float64 {
	if base == 0 {
		if current == 0 {
			return 0
		}
		return 1.0
	}
	return (current - base) / base
}

func main() {
	var (
		baseFile     string
		currentFile  string
		timeThresh   float64
		bytesThresh  float64
		allocsThresh float64
	)
	flag.StringVar(&baseFile, "base", "", "baseline benchmark file (last tag)")
	flag.StringVar(&currentFile, "current", "", "current benchmark file (HEAD)")
	flag.Float64Var(&timeThresh, "time", 0.10, "allowed fractional increase in ns/op (e.g. 0.10 = 10%)")
	flag.Float64Var(&bytesThresh, "bytes", 0.10, "allowed fractional increase in B/op")
	flag.Float64Var(&allocsThresh, "allocs", 0.10, "allowed fractional increase in allocs/op")
	flag.Parse()

	if baseFile == "" || currentFile == "" {
		fmt.Fprintln(os.Stderr, "base and current files are required")
		os.Exit(2)
	}

	base, err := parseFile(baseFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse base:", err)
		os.Exit(2)
	}
	cur, err := parseFile(currentFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse current:", err)
		os.Exit(2)
	}

	var hadRegression bool
	fmt.Println("Benchmark regression report (thresholds: time", timeThresh, "bytes", bytesThresh, "allocs", allocsThresh, ")")
	fmt.Println("Name\tTime(ns/op) delta %\tBytes(B/op) delta %\tAllocs(allocs/op) delta %\tStatus")
	for name, baseRes := range base {
		curRes, ok := cur[name]
		if !ok {
			continue
		}
		timeDelta := pctChange(baseRes.nsPerOp, curRes.nsPerOp)
		bytesDelta := pctChange(baseRes.bytesPerOp, curRes.bytesPerOp)
		allocsDelta := pctChange(baseRes.allocsPerOp, curRes.allocsPerOp)
		status := "OK"
		if timeDelta > timeThresh || bytesDelta > bytesThresh || allocsDelta > allocsThresh {
			status = "REGRESSION"
			hadRegression = true
		}
		fmt.Printf("%s\t%.2f%%\t%.2f%%\t%.2f%%\t%s\n", name, timeDelta*100, bytesDelta*100, allocsDelta*100, status)
	}

	if hadRegression {
		fmt.Fprintln(os.Stderr, "benchmark regressions detected (exceeded thresholds)")
		os.Exit(1)
	}
}
