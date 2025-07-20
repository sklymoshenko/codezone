//go:build unix

// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.
// This file uses v8go (BSD-3-Clause licensed by Roger Peppe)

package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"rogchap.com/v8go"
)

type TypeScriptExecutor struct {
	options ExecutorOptions
	mu      sync.Mutex
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

	iso := v8go.NewIsolate()
	defer iso.Dispose()

	global := v8go.NewObjectTemplate(iso)
	ctx_v8 := v8go.NewContext(iso, global)
	defer ctx_v8.Close()

	outputs := make([]string, 0, 10)
	errors := make([]string, 0, 5)
	if err := js.setupConsole(ctx_v8, &outputs, &errors); err != nil {
		result.Error = fmt.Sprintf("Failed to setup console: %v", err)
		result.ExitCode = 1
		return result, nil
	}

	done := make(chan struct{})
	var execErr error
	var value *v8go.Value

	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				execErr = fmt.Errorf("panic during execution: %v", r)
			}
		}()

		value, execErr = ctx_v8.RunScript(code, "user_code.js")
	}()

	select {
	case <-done:
		if execErr != nil {
			result.Error = execErr.Error()
			result.ExitCode = 1
		} else {
			if value != nil && !value.IsUndefined() && !value.IsNull() {
				outputs = append(outputs, value.String())
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

	duration := time.Since(start)
	result.Duration = duration
	result.DurationString = formatDuration(duration)
	return result, nil
}

func (js *TypeScriptExecutor) setupConsole(ctx *v8go.Context, outputs *[]string, errors *[]string) error {
	console := v8go.NewObjectTemplate(ctx.Isolate())

	logFn := v8go.NewFunctionTemplate(ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := make([]string, len(info.Args()))
		for i := 0; i < len(info.Args()); i++ {
			args[i] = info.Args()[i].String()
		}
		*outputs = append(*outputs, strings.Join(args, " "))
		return v8go.Undefined(ctx.Isolate())
	})
	console.Set("log", logFn)

	// console.error
	errorFn := v8go.NewFunctionTemplate(ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := make([]string, len(info.Args()))
		for i := 0; i < len(info.Args()); i++ {
			args[i] = info.Args()[i].String()
		}
		*errors = append(*errors, strings.Join(args, " "))
		return v8go.Undefined(ctx.Isolate())
	})
	console.Set("error", errorFn)

	// console.warn (treat as output)
	warnFn := v8go.NewFunctionTemplate(ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := make([]string, len(info.Args()))
		for i := 0; i < len(info.Args()); i++ {
			args[i] = info.Args()[i].String()
		}
		*outputs = append(*outputs, strings.Join(args, " "))
		return v8go.Undefined(ctx.Isolate())
	})
	console.Set("warn", warnFn)
	console.Set("info", warnFn)

	global := ctx.Global()
	consoleObj, err := console.NewInstance(ctx)
	if err != nil {
		return fmt.Errorf("failed to create console object: %w", err)
	}

	return global.Set("console", consoleObj)
}

func (js *TypeScriptExecutor) Language() Language { return TypeScript }
func (js *TypeScriptExecutor) IsAvailable() bool {
	return true
}
func (js *TypeScriptExecutor) Cleanup() error {
	return nil
}
