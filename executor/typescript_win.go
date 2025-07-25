//go:build windows

// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
)

type TypeScriptExecutor struct {
	options       ExecutorOptions
	mu            sync.Mutex
	nodeAvailable *bool
}

func NewTypeScriptExecutor(opts ExecutorOptions) *TypeScriptExecutor {
	return &TypeScriptExecutor{
		options: opts,
	}
}

func (js *TypeScriptExecutor) Execute(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	start := time.Now()

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	js.mu.Lock()
	defer js.mu.Unlock()

	result := &ExecutionResult{
		Language: TypeScript,
	}

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

	if js.isNodeAvailable() {
		nodeResult := js.executeWithNode(ctx, code)
		nodeResult.Duration = time.Since(start)
		nodeResult.DurationString = formatDuration(nodeResult.Duration)
		return nodeResult, nil
	}

	gojaResult := js.executeWithGoja(ctx, code)

	if gojaResult.ExitCode != 0 && js.isGojaUnsupportedFeatureError(gojaResult.Error) {
		result.Error = "Goja failed to execute the code."
		result.ExitCode = ExitCodeNodeNotAvailable
		result.Duration = time.Since(start)
		result.DurationString = formatDuration(result.Duration)
		return result, nil
	}

	gojaResult.Duration = time.Since(start)
	gojaResult.DurationString = formatDuration(gojaResult.Duration)
	return gojaResult, nil
}

func (js *TypeScriptExecutor) executeWithGoja(ctx context.Context, code string) *ExecutionResult {
	result := &ExecutionResult{
		Language: TypeScript,
	}

	vm := goja.New()

	outputs := make([]string, 0, 10)
	errors := make([]string, 0, 5)
	if err := js.setupConsole(vm, &outputs, &errors); err != nil {
		result.Error = fmt.Sprintf("Failed to setup console: %v", err)
		result.ExitCode = 1
		return result
	}

	done := make(chan struct{})
	var execErr error
	var value goja.Value

	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				execErr = fmt.Errorf("panic during execution: %v", r)
			}
		}()

		value, execErr = vm.RunString(code)
	}()

	select {
	case <-done:
		if execErr != nil {
			result.Error = execErr.Error()
			result.ExitCode = 1
		} else {
			if value != nil {
				if str := value.String(); str != "undefined" && str != "null" {
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
		result.Error = "Execution timed out"
		result.ExitCode = 124
	}

	return result
}

func (js *TypeScriptExecutor) executeWithNode(ctx context.Context, code string) *ExecutionResult {
	result := &ExecutionResult{
		Language: TypeScript,
	}

	if !js.isNodeAvailable() {
		result.Error = "Node.js is not available for fallback execution"
		result.ExitCode = ExitCodeNodeNotAvailable
		return result
	}

	tempFile, err := createTempFile(code)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create temp file: %v", err)
		result.ExitCode = 158
		return result
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	cmd := exec.CommandContext(ctx, "node", tempFile.Name())

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

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

func (js *TypeScriptExecutor) isGojaUnsupportedFeatureError(errorMsg string) bool {
	fallbackPatterns := []string{
		"Unexpected token",
		"SyntaxError",
		"ReferenceError",
		"TypeError",
	}

	for _, pattern := range fallbackPatterns {
		if strings.Contains(strings.ToLower(errorMsg), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

func (js *TypeScriptExecutor) isNodeAvailable() bool {
	if js.nodeAvailable != nil {
		return *js.nodeAvailable
	}

	if isTestBuild() {
		available := false
		js.nodeAvailable = &available
		return false
	}

	cmd := exec.Command("node", "--version")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	available := cmd.Run() == nil
	js.nodeAvailable = &available
	return available
}

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

func isTestBuild() bool {
	return os.Getenv("CODEZONE_TEST_MODE") == "true" ||
		strings.Contains(os.Args[0], "test") ||
		strings.Contains(os.Args[0], "_test")
}

func (js *TypeScriptExecutor) setupConsole(vm *goja.Runtime, outputs *[]string, errors *[]string) error {
	console := vm.NewObject()

	logFn := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		args := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.String()
		}
		*outputs = append(*outputs, strings.Join(args, " "))
		return goja.Undefined()
	})
	console.Set("log", logFn)

	// console.error
	errorFn := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		args := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.String()
		}
		*errors = append(*errors, strings.Join(args, " "))
		return goja.Undefined()
	})
	console.Set("error", errorFn)

	// console.warn (treat as output)
	warnFn := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		args := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.String()
		}
		*outputs = append(*outputs, strings.Join(args, " "))
		return goja.Undefined()
	})
	console.Set("warn", warnFn)
	console.Set("info", warnFn)

	// Set console in global scope
	vm.Set("console", console)

	return nil
}

func (js *TypeScriptExecutor) Language() Language { return TypeScript }
func (js *TypeScriptExecutor) IsAvailable() bool {
	return true
}
func (js *TypeScriptExecutor) Cleanup() error {
	return nil
}
