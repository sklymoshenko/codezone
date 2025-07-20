package executor

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestJavaScriptExecutor_Execute(t *testing.T) {
	executor := NewTypeScriptExecutor(DefaultExecutorOptions())

	t.Run("should execute simple JavaScript code", func(t *testing.T) {
		code := `console.log("Hello, World!");`
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}

		if !strings.Contains(result.Output, "Hello, World!") {
			t.Fatalf("Expected output to contain 'Hello, World!', got %s", result.Output)
		}

		if result.Language != TypeScript {
			t.Fatalf("Expected language TypeScript, got %s", result.Language)
		}

		if result.Duration <= 0 {
			t.Fatalf("Expected positive duration, got %v", result.Duration)
		}

		if result.DurationString == "" {
			t.Fatalf("Expected non-empty duration string")
		}
	})

	t.Run("should handle console.error", func(t *testing.T) {
		code := `console.error("This is an error");`
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}

		if !strings.Contains(result.Error, "This is an error") {
			t.Fatalf("Expected error to contain 'This is an error', got %s", result.Error)
		}
	})

	t.Run("should handle console.warn", func(t *testing.T) {
		code := `console.warn("This is a warning");`
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !strings.Contains(result.Output, "This is a warning") {
			t.Fatalf("Expected output to contain 'This is a warning', got %s", result.Output)
		}
	})

	t.Run("should handle syntax errors", func(t *testing.T) {
		code := `const x = {;` // Invalid syntax
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if runtime.GOOS == "windows" {
			if result.ExitCode != 2 {
				t.Fatalf("Expected exit code 2, got %d", result.ExitCode)
			}
		} else {
			if result.ExitCode != 1 {
				t.Fatalf("Expected exit code 1, got %d", result.ExitCode)
			}
		}

		if result.Error == "" {
			t.Fatalf("Expected error message for syntax error")
		}
	})

	t.Run("should handle runtime errors", func(t *testing.T) {
		code := `throw new Error("Runtime error");`
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 1 {
			t.Fatalf("Expected exit code 1, got %d", result.ExitCode)
		}

		if !strings.Contains(result.Error, "Runtime error") {
			t.Fatalf("Expected error to contain 'Runtime error', got %s", result.Error)
		}
	})

	t.Run("should return expression results", func(t *testing.T) {
		code := `2 + 2`
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}

		if !strings.Contains(result.Output, "4") {
			t.Fatalf("Expected output to contain '4', got %s", result.Output)
		}
	})

	t.Run("should handle multiple console outputs", func(t *testing.T) {
		code := `
			console.log("First line");
			console.log("Second line");
			console.error("Error line");
		`
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !strings.Contains(result.Output, "First line") {
			t.Fatalf("Expected output to contain 'First line', got %s", result.Output)
		}

		if !strings.Contains(result.Output, "Second line") {
			t.Fatalf("Expected output to contain 'Second line', got %s", result.Output)
		}

		if !strings.Contains(result.Error, "Error line") {
			t.Fatalf("Expected error to contain 'Error line', got %s", result.Error)
		}
	})

	t.Run("should handle timeout", func(t *testing.T) {
		code := `while(true) { /* infinite loop */ }`
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		result, err := executor.Execute(ctx, code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 124 {
			t.Fatalf("Expected exit code 124 (timeout), got %d", result.ExitCode)
		}

		if !strings.Contains(result.Error, "timed out") {
			t.Fatalf("Expected error to contain 'timed out', got %s", result.Error)
		}
	})

	t.Run("should handle empty code", func(t *testing.T) {
		code := ``
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}
	})

	t.Run("should handle console.info", func(t *testing.T) {
		code := `console.info("Info message");`
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !strings.Contains(result.Output, "Info message") {
			t.Fatalf("Expected output to contain 'Info message', got %s", result.Output)
		}
	})

	t.Run("should handle console with multiple arguments", func(t *testing.T) {
		code := `console.log("Value:", 42, "Type:", typeof 42);`
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedTokens := []string{"Value:", "42", "Type:", "number"}
		for _, token := range expectedTokens {
			if !strings.Contains(result.Output, token) {
				t.Fatalf("Expected output to contain '%s', got %s", token, result.Output)
			}
		}
	})

	t.Run("should handle template literals", func(t *testing.T) {
		code := "function greet(name) {\n" +
			"  return `Hello, ${name}!`;\n" +
			"}\n" +
			"console.log(greet(\"World\"));"
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}

		if !strings.Contains(result.Output, "Hello, World!") {
			t.Fatalf("Expected output to contain 'Hello, World!', got %s", result.Output)
		}
	})

	t.Run("should handle template literals with expressions", func(t *testing.T) {
		code := "const name = \"Alice\";\n" +
			"const age = 25;\n" +
			"const message = `${name} is ${age} years old`;\n" +
			"console.log(message);"
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}

		if !strings.Contains(result.Output, "Alice is 25 years old") {
			t.Fatalf("Expected output to contain 'Alice is 25 years old', got %s", result.Output)
		}
	})

	t.Run("should handle nested template literals", func(t *testing.T) {
		code := "const greeting = \"Hello\";\n" +
			"const name = \"Bob\";\n" +
			"const message = `${greeting}, ${name}! Welcome!`;\n" +
			"console.log(message);"
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}

		if !strings.Contains(result.Output, "Hello, Bob! Welcome!") {
			t.Fatalf("Expected output to contain 'Hello, Bob! Welcome!', got %s", result.Output)
		}
	})

	t.Run("should handle arrow functions", func(t *testing.T) {
		code := "const add = (a, b) => a + b;\n" +
			"console.log(add(2, 3));"
		result, err := executor.Execute(context.Background(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}

		if !strings.Contains(result.Output, "5") {
			t.Fatalf("Expected output to contain '5', got %s", result.Output)
		}
	})
}

func TestJavaScriptExecutor_Language(t *testing.T) {
	executor := NewTypeScriptExecutor(DefaultExecutorOptions())

	if executor.Language() != TypeScript {
		t.Fatalf("Expected language TypeScript, got %s", executor.Language())
	}
}

func TestJavaScriptExecutor_IsAvailable(t *testing.T) {
	executor := NewTypeScriptExecutor(DefaultExecutorOptions())

	if !executor.IsAvailable() {
		t.Fatalf("Expected executor to be available")
	}
}

func TestJavaScriptExecutor_Cleanup(t *testing.T) {
	executor := NewTypeScriptExecutor(DefaultExecutorOptions())

	err := executor.Cleanup()
	if err != nil {
		t.Fatalf("Expected no error from cleanup, got %v", err)
	}
}

func TestJavaScriptExecutor_ThreadSafety(t *testing.T) {
	executor := NewTypeScriptExecutor(DefaultExecutorOptions())

	// Run multiple executions concurrently to test thread safety
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			code := `console.log("Execution " + ` + string(rune('0'+id)) + `);`
			result, err := executor.Execute(context.Background(), code, "")

			if err != nil {
				t.Errorf("Execution %d failed: %v", id, err)
			}

			if result.ExitCode != 0 {
				t.Errorf("Execution %d had non-zero exit code: %d", id, result.ExitCode)
			}

			done <- true
		}(i)
	}

	// Wait for all executions to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestJavaScriptExecutor_ContextHandling(t *testing.T) {
	executor := NewTypeScriptExecutor(DefaultExecutorOptions())

	t.Run("should use default timeout when context is nil", func(t *testing.T) {
		code := `console.log("test");`
		result, err := executor.Execute(context.TODO(), code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
		}
	})

	t.Run("should respect context cancellation", func(t *testing.T) {
		code := "let i = 0;\n" +
			"while(i < 1000000) {\n" +
			"  i++;\n" +
			"}\n" +
			"console.log(\"Done\");"
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel immediately
		cancel()

		result, err := executor.Execute(ctx, code, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should timeout due to cancelled context
		if result.ExitCode != 124 {
			t.Fatalf("Expected timeout exit code 124, got %d", result.ExitCode)
		}
	})
}
