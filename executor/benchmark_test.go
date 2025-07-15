package executor

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// BenchmarkJavaScriptExecutor_SimpleExecution benchmarks basic code execution
func BenchmarkJavaScriptExecutor_SimpleExecution(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `console.log("Hello, World!");`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_MathOperations benchmarks mathematical computations
func BenchmarkJavaScriptExecutor_MathOperations(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		let result = 0;
		for (let i = 0; i < 1000; i++) {
			result += Math.sin(i) * Math.cos(i);
		}
		console.log(result);
	`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_StringOperations benchmarks string manipulation
func BenchmarkJavaScriptExecutor_StringOperations(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		let str = "Hello";
		for (let i = 0; i < 100; i++) {
			str += " World " + i;
		}
		console.log(str.length);
	`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_ArrayOperations benchmarks array manipulations
func BenchmarkJavaScriptExecutor_ArrayOperations(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		let arr = [];
		for (let i = 0; i < 1000; i++) {
			arr.push(i);
		}
		let sum = arr.reduce((a, b) => a + b, 0);
		console.log(sum);
	`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_ObjectOperations benchmarks object manipulations
func BenchmarkJavaScriptExecutor_ObjectOperations(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		let obj = {};
		for (let i = 0; i < 100; i++) {
			obj["key" + i] = i * 2;
		}
		let keys = Object.keys(obj);
		console.log(keys.length);
	`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_ConsoleOutput benchmarks console logging overhead
func BenchmarkJavaScriptExecutor_ConsoleOutput(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		for (let i = 0; i < 50; i++) {
			console.log("Log message", i);
		}
	`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_ErrorHandling benchmarks error scenarios
func BenchmarkJavaScriptExecutor_ErrorHandling(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		try {
			throw new Error("Test error");
		} catch (e) {
			console.error(e.message);
		}
	`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_CodeSizes benchmarks different code sizes
func BenchmarkJavaScriptExecutor_CodeSizes(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	ctx := context.Background()

	sizes := []struct {
		name string
		code string
	}{
		{
			name: "SmallCode",
			code: `console.log("small");`,
		},
		{
			name: "MediumCode",
			code: strings.Repeat(`console.log("line"); `, 50),
		},
		{
			name: "LargeCode",
			code: strings.Repeat(`console.log("line " + Math.random()); `, 200),
		},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := executor.Execute(ctx, size.code, "")
				if err != nil {
					b.Fatalf("Execution failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkJavaScriptExecutor_MemoryUsage benchmarks memory allocation patterns
func BenchmarkJavaScriptExecutor_MemoryUsage(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		// Create and garbage collect objects
		for (let i = 0; i < 100; i++) {
			let data = new Array(100).fill(i);
			data.map(x => x * 2);
		}
		console.log("Memory test complete");
	`
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_Parallel benchmarks concurrent executions
func BenchmarkJavaScriptExecutor_Parallel(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `console.log("Parallel execution", Math.random());`
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := executor.Execute(ctx, code, "")
			if err != nil {
				b.Fatalf("Execution failed: %v", err)
			}
		}
	})
}

// BenchmarkJavaScriptExecutor_ContextSetup benchmarks isolate creation overhead
func BenchmarkJavaScriptExecutor_ContextSetup(b *testing.B) {
	opts := DefaultExecutorOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor := NewJavaScriptExecutor(opts)
		code := `console.log("setup test");`
		ctx := context.Background()

		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}

		err = executor.Cleanup()
		if err != nil {
			b.Fatalf("Cleanup failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_ComplexAlgorithm benchmarks computational algorithms
func BenchmarkJavaScriptExecutor_ComplexAlgorithm(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		// Fibonacci sequence calculation
		function fibonacci(n) {
			if (n <= 1) return n;
			return fibonacci(n - 1) + fibonacci(n - 2);
		}
		
		let result = fibonacci(20);
		console.log("Fibonacci result:", result);
	`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJavaScriptExecutor_JSONOperations benchmarks JSON parsing/stringifying
func BenchmarkJavaScriptExecutor_JSONOperations(b *testing.B) {
	executor := NewJavaScriptExecutor(DefaultExecutorOptions())
	code := `
		let data = {
			users: [],
			metadata: { version: 1, timestamp: Date.now() }
		};
		
		for (let i = 0; i < 100; i++) {
			data.users.push({
				id: i,
				name: "User " + i,
				active: i % 2 === 0
			});
		}
		
		let json = JSON.stringify(data);
		let parsed = JSON.parse(json);
		console.log("Users count:", parsed.users.length);
	`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, code, "")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkFormatDuration_Performance benchmarks duration formatting performance
func BenchmarkFormatDuration_Performance(b *testing.B) {
	durations := []time.Duration{
		100 * time.Nanosecond,
		1500 * time.Nanosecond,
		1500 * time.Microsecond,
		1500 * time.Millisecond,
		2 * time.Second,
	}

	for _, d := range durations {
		b.Run(fmt.Sprintf("Duration_%v", d), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = formatDuration(d)
			}
		})
	}
}

// BenchmarkExecutionManager benchmarks the execution manager layer
func BenchmarkExecutionManager(b *testing.B) {
	manager := NewExecutionManager(DefaultExecutorOptions())
	defer manager.Cleanup()

	config := ExecutionConfig{
		Code:     `console.log("Manager test", Math.random());`,
		Language: JavaScript,
		Timeout:  0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := manager.Execute(config)
		if err != nil {
			b.Fatalf("Manager execution failed: %v", err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("Execution failed with exit code %d", result.ExitCode)
		}
	}
}

// BenchmarkExecutionManager_Parallel benchmarks concurrent manager usage
func BenchmarkExecutionManager_Parallel(b *testing.B) {
	manager := NewExecutionManager(DefaultExecutorOptions())
	defer manager.Cleanup()

	config := ExecutionConfig{
		Code:     `console.log("Parallel manager test", Math.random());`,
		Language: JavaScript,
		Timeout:  0,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := manager.Execute(config)
			if err != nil {
				b.Fatalf("Manager execution failed: %v", err)
			}
			if result.ExitCode != 0 {
				b.Fatalf("Execution failed with exit code %d", result.ExitCode)
			}
		}
	})
}
