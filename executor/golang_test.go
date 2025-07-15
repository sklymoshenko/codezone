package executor

import (
	"context"
	"testing"
	"time"
)

func TestGoExecutor_SimpleCode(t *testing.T) {
	if !isGoAvailable() {
		t.Skip("Go compiler not available, skipping test")
	}

	executor := NewGoExecutor(DefaultExecutorOptions())

	code := `fmt.Println("Hello, World!")`

	result, err := executor.Execute(context.Background(), code, "")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result.Error != "" {
		t.Fatalf("Execution error: %s", result.Error)
	}

	expected := "Hello, World!"
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

func TestGoExecutor_WithMain(t *testing.T) {
	if !isGoAvailable() {
		t.Skip("Go compiler not available, skipping test")
	}

	executor := NewGoExecutor(DefaultExecutorOptions())

	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello from main!")
}`

	result, err := executor.Execute(context.Background(), code, "")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result.Error != "" {
		t.Fatalf("Execution error: %s", result.Error)
	}

	expected := "Hello from main!"
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

func TestGoExecutor_WithInput(t *testing.T) {
	if !isGoAvailable() {
		t.Skip("Go compiler not available, skipping test")
	}

	executor := NewGoExecutor(DefaultExecutorOptions())

	code := `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		fmt.Printf("You said: %s", scanner.Text())
	}
}`

	result, err := executor.Execute(context.Background(), code, "Hello Input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result.Error != "" {
		t.Fatalf("Execution error: %s", result.Error)
	}

	expected := "You said: Hello Input"
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

func TestGoExecutor_CompileError(t *testing.T) {
	if !isGoAvailable() {
		t.Skip("Go compiler not available, skipping test")
	}

	executor := NewGoExecutor(DefaultExecutorOptions())

	// Use code that will definitely fail to compile
	code := `package main

func main() {
	undefinedFunction()
}`

	result, err := executor.Execute(context.Background(), code, "")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result.Error == "" {
		t.Fatal("Expected compile error, but got none")
	}

	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for compile error")
	}
}

func TestGoExecutor_Timeout(t *testing.T) {
	if !isGoAvailable() {
		t.Skip("Go compiler not available, skipping test")
	}

	executor := NewGoExecutor(DefaultExecutorOptions())

	code := `package main

import (
	"fmt"
	"time"
)

func main() {
	time.Sleep(1 * time.Second)
	fmt.Println("This should not print")
}`

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := executor.Execute(ctx, code, "")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result.Error == "" {
		t.Fatal("Expected timeout error, but got none")
	}

	// Accept both 124 (timeout) and 1 (kill signal) as valid timeout exit codes
	if result.ExitCode != 124 && result.ExitCode != 1 {
		t.Errorf("Expected exit code 124 or 1 for timeout, got %d", result.ExitCode)
	}
}

func TestGoExecutor_IsAvailable(t *testing.T) {
	executor := NewGoExecutor(DefaultExecutorOptions())

	// This test just checks that IsAvailable doesn't panic
	available := executor.IsAvailable()

	// The result depends on whether Go is installed on the test system
	t.Logf("Go available: %v", available)
}

func TestGoExecutor_Language(t *testing.T) {
	executor := NewGoExecutor(DefaultExecutorOptions())

	if executor.Language() != Go {
		t.Errorf("Expected language %s, got %s", Go, executor.Language())
	}
}

// Helper function to check if Go is available
func isGoAvailable() bool {
	executor := NewGoExecutor(DefaultExecutorOptions())
	return executor.IsAvailable()
}
