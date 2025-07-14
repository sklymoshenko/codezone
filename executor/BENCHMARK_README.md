# JavaScript Executor Benchmarks

This directory contains benchmarks for the JavaScript executor system, focusing on V8 execution performance.

**⚠️ Warning: These benchmarks may crash due to V8go threading issues. Use timeout protection in CI/CD.**

## Quick Start

```bash
# Run all benchmarks (may crash due to V8go threading issues)
go test -bench=. -benchmem ./executor/

# Run with timeout protection
timeout 30s go test -bench=. -benchmem ./executor/

# Run individual benchmarks to avoid crashes
go test -bench=BenchmarkJavaScript_SimpleCode -benchmem ./executor/

# Save baseline results for comparison
go test -bench=. -benchmem ./executor/ > benchmarks_baseline.txt
```

## Benchmark Categories

### JavaScript Execution Performance

Tests various JavaScript execution scenarios that matter for real-world usage:

- **BenchmarkJavaScript_SimpleCode** - Basic arithmetic and variable assignment
- **BenchmarkJavaScript_MathOperations** - Mathematical computations
- **BenchmarkJavaScript_StringManipulation** - String operations and concatenation
- **BenchmarkJavaScript_ArrayOperations** - Array creation and manipulation
- **BenchmarkJavaScript_ObjectOperations** - Object creation and property access
- **BenchmarkJavaScript_ConsoleOutput** - Console logging overhead
- **BenchmarkJavaScript_ErrorHandling** - Exception handling performance
- **BenchmarkJavaScript_EmptyCode** - Minimal execution overhead
- **BenchmarkJavaScript_ParallelExecution** - Concurrent execution scaling

## Interpreting Results

### Performance Metrics

```
BenchmarkName-PROCS    ITERATIONS    NANOSECONDS/OP    BYTES/OP    ALLOCS/OP
```

- **ITERATIONS**: How many times the benchmark ran (higher = more reliable)
- **NANOSECONDS/OP**: Time per operation (lower = faster)
- **BYTES/OP**: Memory allocated per operation (lower = more efficient)
- **ALLOCS/OP**: Number of allocations per operation (lower = less GC pressure)

### Regression Detection

**🔴 Performance Regression Indicators:**

- Time per operation increases by >20%
- Memory usage increases by >50%
- Allocation count increases significantly
- Iterations drop dramatically (indicates slower execution)

**🟡 Watch For:**

- Gradual increases over multiple commits
- Large variations in run-to-run results
- Platform-specific performance differences

**🟢 Good Performance:**

- Consistent timing across runs
- Low memory allocation
- High iteration counts

### Example Analysis

**Baseline:**

```
BenchmarkJavaScript_SimpleCode-24    1000    1.8 ms/op    8.5 KB/op    120 allocs/op
```

**After Changes:**

```
BenchmarkJavaScript_SimpleCode-24     750    2.4 ms/op   12.8 KB/op    180 allocs/op
```

**Analysis:**

- ❌ **33% slower** (1.8 → 2.4 ms/op)
- ❌ **51% more memory** (8.5 → 12.8 KB/op)
- ❌ **50% more allocations** (120 → 180 allocs/op)
- **Verdict**: Clear performance regression requiring investigation

## Troubleshooting

### V8 Crashes

If you encounter V8 isolate disposal crashes:

```
Fatal error in v8::Isolate::Dispose()
Disposing the isolate that is entered by a thread
```

**Solutions:**

1. Use timeout protection: `timeout 30s go test -bench=. -benchmem ./executor/`
2. Run benchmarks with `-p 1` to disable parallelism: `go test -bench=. -p 1`
3. Run individual benchmarks separately to avoid crashes

### Inconsistent Results

- Run with `-count=N` for multiple iterations: `go test -bench=. -count=5`
- Ensure system is idle during benchmarking
- Use consistent hardware and Go versions
- Consider CPU throttling and background processes

### Memory Benchmarks

For detailed memory profiling:

```bash
go test -bench=BenchmarkJavaScript_SimpleCode -benchmem -memprofile=mem.prof ./executor/
go tool pprof mem.prof
```

## Performance Targets

| Component           | Target Time | Target Memory | Notes                  |
| ------------------- | ----------- | ------------- | ---------------------- |
| Simple JS Execution | < 10 ms/op  | < 1 MB/op     | V8 overhead acceptable |
| Math Operations     | < 15 ms/op  | < 1.5 MB/op   | Complex calculations   |
| String Manipulation | < 8 ms/op   | < 800 KB/op   | String processing      |
| Console Output      | < 5 ms/op   | < 500 KB/op   | Logging overhead       |

These targets help identify when optimizations are needed or when regressions occur.
