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
	"os"
	"regexp"
	"strconv"
	"strings"
)

type benchResult struct {
	name        string
	nsPerOp     float64
	bytesPerOp  float64
	allocsPerOp float64
}

var benchLineRE = regexp.MustCompile(`^(Benchmark[^\s]+)\s+([0-9]+)\s+([0-9]+) ns/op(?:\s+([0-9]+) B/op\s+([0-9]+) allocs/op)?`)

func parseFile(path string) (map[string]benchResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	results := make(map[string]benchResult)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		m := benchLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		ns, _ := strconv.ParseFloat(m[3], 64)
		var bytesVal, allocsVal float64
		if m[4] != "" {
			bytesVal, _ = strconv.ParseFloat(m[4], 64)
		}
		if m[5] != "" {
			allocsVal, _ = strconv.ParseFloat(m[5], 64)
		}
		results[name] = benchResult{name: name, nsPerOp: ns, bytesPerOp: bytesVal, allocsPerOp: allocsVal}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no benchmarks parsed from " + path)
	}
	return results, nil
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
	fmt.Println("Name\tTime(ns/op)Δ%\tBytes(B/op)Δ%\tAllocs(allocs/op)Δ%\tStatus")
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
