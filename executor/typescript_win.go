//go:build windows

// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.
// This file uses Otto (MIT licensed by Robert Krimen)

package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/robertkrimen/otto"
)

// TypeScriptExecutor implements JavaScript execution
type TypeScriptExecutor struct {
	options ExecutorOptions
	mu      sync.Mutex // Protect Otto operations
}

// NewTypeScriptExecutor creates a new TypeScript executor
func NewTypeScriptExecutor(opts ExecutorOptions) *TypeScriptExecutor {
	return &TypeScriptExecutor{
		options: opts,
	}
}

// Execute runs JavaScript or TypeScript code
func (js *TypeScriptExecutor) Execute(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	start := time.Now()

	// Add a default timeout if context doesn't have one
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	// Use mutex to ensure thread safety
	js.mu.Lock()
	defer js.mu.Unlock()

	result := &ExecutionResult{
		Language: TypeScript,
	}

	// --- Always transpile with esbuild TypeScript loader ---
	transpileResult := api.Transform(code, api.TransformOptions{
		Loader:       api.LoaderTS,
		Format:       api.FormatDefault,
		Sourcemap:    api.SourceMapNone,
		Target:       api.ESNext,
		MinifySyntax: false,
	})
	if len(transpileResult.Errors) > 0 {
		tsErrors := make([]string, len(transpileResult.Errors))
		for i, err := range transpileResult.Errors {
			tsErrors[i] = err.Text
		}
		result.Error = "TypeScript transpile error:\n" + strings.Join(tsErrors, "\n")
		result.ExitCode = 2
		result.Duration = time.Since(start)
		result.DurationString = formatDuration(result.Duration)
		return result, nil
	}
	code = string(transpileResult.Code)
	// --- End transpile ---

	// First, try Node.js if available
	if js.isNodeAvailable() {
		nodeResult := js.executeWithNode(ctx, code)
		nodeResult.Duration = time.Since(start)
		nodeResult.DurationString = formatDuration(nodeResult.Duration)
		return nodeResult, nil
	}

	ottoResult := js.executeWithOtto(ctx, code)

	// Check if Otto failed due to unsupported features
	if ottoResult.ExitCode != 0 && js.isOttoUnsupportedFeatureError(ottoResult.Error) {
		// Otto failed due to unsupported features, but Node.js isn't available
		result.Error = "Otto failed to execute the code."
		result.ExitCode = ExitCodeNodeNotAvailable
		result.Duration = time.Since(start)
		result.DurationString = formatDuration(result.Duration)
		return result, nil
	}

	// Return Otto's result (either success or other error)
	ottoResult.Duration = time.Since(start)
	ottoResult.DurationString = formatDuration(ottoResult.Duration)
	return ottoResult, nil
}

// executeWithOtto runs code using Otto
func (js *TypeScriptExecutor) executeWithOtto(ctx context.Context, code string) *ExecutionResult {
	result := &ExecutionResult{
		Language: TypeScript,
	}

	// Create a new Otto VM for each execution
	vm := otto.New()

	// Set up console logging
	outputs := make([]string, 0, 10)
	errors := make([]string, 0, 5)
	if err := js.setupConsole(vm, &outputs, &errors); err != nil {
		result.Error = fmt.Sprintf("Failed to setup console: %v", err)
		result.ExitCode = 1
		return result
	}

	// Execute with timeout using a channel
	done := make(chan struct{})
	var execErr error
	var value otto.Value

	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				execErr = fmt.Errorf("panic during execution: %v", r)
			}
		}()

		value, execErr = vm.Run(code)
	}()

	// Wait for execution or timeout
	select {
	case <-done:
		// Execution completed
		if execErr != nil {
			result.Error = execErr.Error()
			result.ExitCode = 1
		} else {
			// Include the final expression result if it's meaningful
			if !value.IsUndefined() && !value.IsNull() {
				if str, err := value.ToString(); err == nil {
					outputs = append(outputs, str)
				}
			}
		}

		result.Output = strings.Join(outputs, "\n")
		if len(errors) > 0 {
			if result.Error != "" {
				result.Error += "\n" + strings.Join(errors, "\n")
			} else {
				result.Error = strings.Join(errors, "\n")
			}
		}

	case <-ctx.Done():
		// Timeout occurred
		result.Error = "Execution timed out"
		result.ExitCode = 124
	}

	return result
}

// executeWithNode runs code using Node.js
func (js *TypeScriptExecutor) executeWithNode(ctx context.Context, code string) *ExecutionResult {
	result := &ExecutionResult{
		Language: TypeScript,
	}

	// Check if Node.js is available
	if !js.isNodeAvailable() {
		result.Error = "Node.js is not available for fallback execution"
		result.ExitCode = ExitCodeNodeNotAvailable
		return result
	}

	// Create a temporary file with the code
	tempFile, err := createTempFile(code)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create temp file: %v", err)
		result.ExitCode = 158
		return result
	}
	defer tempFile.Close()

	// Execute with Node.js
	cmd := exec.CommandContext(ctx, "node", tempFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "Execution timed out"
			result.ExitCode = 124
		} else {
			result.Error = string(output)
			result.ExitCode = 1
		}
	} else {
		result.Output = string(output)
		result.ExitCode = 0
	}

	return result
}

// isOttoUnsupportedFeatureError determines if Otto failed due to unsupported ES2016+ features
func (js *TypeScriptExecutor) isOttoUnsupportedFeatureError(errorMsg string) bool {
	// Check for common Otto limitations
	fallbackPatterns := []string{
		"arrow function",
		"=>",
		"template literal",
		"`",
		"const",
		"let",
		"class",
		"import",
		"export",
		"async",
		"await",
		"spread",
		"destructuring",
		"Unexpected reserved word",
		"Unexpected token ILLEGAL",
	}

	for _, pattern := range fallbackPatterns {
		if strings.Contains(strings.ToLower(errorMsg), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// isNodeAvailable checks if Node.js is installed and available
func (js *TypeScriptExecutor) isNodeAvailable() bool {
	cmd := exec.Command("node", "--version")
	return cmd.Run() == nil
}

// createTempFile creates a temporary file with the given code
func createTempFile(code string) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "codezone-*.js")
	if err != nil {
		return nil, err
	}

	_, err = tempFile.WriteString(code)
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, err
	}

	return tempFile, nil
}

// setupConsole sets up console.log, console.error, etc.
func (js *TypeScriptExecutor) setupConsole(vm *otto.Otto, outputs *[]string, errors *[]string) error {
	// Create console object
	console, err := vm.Object("console = {}")
	if err != nil {
		return fmt.Errorf("failed to create console object: %w", err)
	}

	// console.log
	logFn, err := vm.ToValue(func(call otto.FunctionCall) otto.Value {
		args := make([]string, len(call.ArgumentList))
		for i, arg := range call.ArgumentList {
			if str, err := arg.ToString(); err == nil {
				args[i] = str
			} else {
				args[i] = arg.String()
			}
		}
		*outputs = append(*outputs, strings.Join(args, " "))
		return otto.UndefinedValue()
	})
	if err != nil {
		return fmt.Errorf("failed to create log function: %w", err)
	}
	console.Set("log", logFn)

	// console.error
	errorFn, err := vm.ToValue(func(call otto.FunctionCall) otto.Value {
		args := make([]string, len(call.ArgumentList))
		for i, arg := range call.ArgumentList {
			if str, err := arg.ToString(); err == nil {
				args[i] = str
			} else {
				args[i] = arg.String()
			}
		}
		*errors = append(*errors, strings.Join(args, " "))
		return otto.UndefinedValue()
	})
	if err != nil {
		return fmt.Errorf("failed to create error function: %w", err)
	}
	console.Set("error", errorFn)

	// console.warn (treat as output)
	warnFn, err := vm.ToValue(func(call otto.FunctionCall) otto.Value {
		args := make([]string, len(call.ArgumentList))
		for i, arg := range call.ArgumentList {
			if str, err := arg.ToString(); err == nil {
				args[i] = str
			} else {
				args[i] = arg.String()
			}
		}
		*outputs = append(*outputs, strings.Join(args, " "))
		return otto.UndefinedValue()
	})
	if err != nil {
		return fmt.Errorf("failed to create warn function: %w", err)
	}
	console.Set("warn", warnFn)
	console.Set("info", warnFn) // info same as warn

	return nil
}

func (js *TypeScriptExecutor) Language() Language { return TypeScript }
func (js *TypeScriptExecutor) IsAvailable() bool {
	// Otto is embedded, so it's always available once built
	return true
}
func (js *TypeScriptExecutor) Cleanup() error {
	// No cleanup needed for Otto
	return nil
}
