// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// GoExecutor implements Go execution using the system Go compiler
type GoExecutor struct {
	options ExecutorOptions
	mu      sync.Mutex
}

// NewGoExecutor creates a new Go executor
func NewGoExecutor(opts ExecutorOptions) *GoExecutor {
	return &GoExecutor{
		options: opts,
	}
}

// Execute runs Go code using the system Go compiler
func (g *GoExecutor) Execute(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	start := time.Now()

	// Add default timeout if context doesn't have one
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	// Use mutex to ensure thread safety
	g.mu.Lock()
	defer g.mu.Unlock()

	result := &ExecutionResult{
		Language: Go,
	}

	// Check if Go is available
	if !g.IsAvailable() {
		result.Error = "Go is not installed. Please install Go from https://golang.org/dl/ or install this package using your system's package manager"
		result.ExitCode = ExitCodeGoNotInstalled
		return result, nil
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "codezone-go-*")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create temp directory: %v", err)
		result.ExitCode = 1
		return result, nil
	}
	defer os.RemoveAll(tempDir)

	// Prepare the Go code
	goCode := g.prepareGoCode(code)

	// Write code to temporary file
	tempFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(tempFile, []byte(goCode), 0644); err != nil {
		result.Error = fmt.Sprintf("Failed to write temp file: %v", err)
		result.ExitCode = 1
		return result, nil
	}

	// Execute Go code
	cmd := exec.CommandContext(ctx, "go", "run", tempFile)

	// Hide the command prompt window on Windows
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}

	cmd.Dir = tempDir

	// Set up input if provided
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	// Capture output
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err = cmd.Run()

	// Process results
	result.Output = strings.TrimSpace(stdout.String())

	if err != nil {
		// Handle different types of errors
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "Execution timed out"
			result.ExitCode = 124
		} else {
			stderrText := strings.TrimSpace(stderr.String())
			if stderrText != "" {
				result.Error = g.cleanGoError(stderrText)
			} else {
				result.Error = err.Error()
			}
			if exitError, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitError.ExitCode()
			} else {
				result.ExitCode = 1
			}
		}
	}

	duration := time.Since(start)
	result.Duration = duration
	result.DurationString = formatDuration(duration)
	return result, nil
}

// prepareGoCode wraps user code in a proper Go program structure if needed
func (g *GoExecutor) prepareGoCode(code string) string {
	// Check if code already has package declaration
	if strings.Contains(code, "package ") {
		return code
	}

	// Check if code has a main function
	hasMain := strings.Contains(code, "func main(")

	if hasMain {
		// Code has main function but no package, add package main
		return fmt.Sprintf("package main\n\n%s", code)
	}

	// No main function, wrap the code in main
	return fmt.Sprintf(`package main

import "fmt"

func main() {
%s
}`, g.indentCode(code))
}

// indentCode adds proper indentation to user code
func (g *GoExecutor) indentCode(code string) string {
	lines := strings.Split(code, "\n")
	var indentedLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			indentedLines = append(indentedLines, "\t"+line)
		} else {
			indentedLines = append(indentedLines, line)
		}
	}
	return strings.Join(indentedLines, "\n")
}

// cleanGoError removes temporary file paths from Go error messages for cleaner output
func (g *GoExecutor) cleanGoError(errorText string) string {
	lines := strings.Split(errorText, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Remove temp file paths - replace with "main.go"
		if strings.Contains(line, "/tmp/codezone-go-") {
			// Find the pattern and replace it
			parts := strings.Split(line, "/")
			for i, part := range parts {
				if strings.HasPrefix(part, "codezone-go-") {
					// Replace everything up to and including this part with "main.go"
					newLine := "main.go"
					if i+1 < len(parts) {
						newLine += strings.Join(parts[i+1:], "/")
					}
					line = strings.Replace(line, strings.Join(parts[:i+2], "/"), newLine, 1)
					break
				}
			}
		}
		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, "\n")
}

// Language returns the language identifier
func (g *GoExecutor) Language() Language {
	return Go
}

// IsAvailable checks if Go compiler is available on the system
func (g *GoExecutor) IsAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

// Cleanup performs any necessary cleanup (no-op for Go executor)
func (g *GoExecutor) Cleanup() error {
	return nil
}
