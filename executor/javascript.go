package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"rogchap.com/v8go"
)

// formatDuration formats a duration with max 3 decimal places for cleaner display
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.3gμs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.3gms", float64(d.Nanoseconds())/1000000)
	}
	return fmt.Sprintf("%.3gs", d.Seconds())
}

// JavaScriptExecutor implements JavaScript execution using V8
type JavaScriptExecutor struct {
	options ExecutorOptions
	mu      sync.Mutex // Protect V8 operations
}

// NewJavaScriptExecutor creates a new V8-based executor
func NewJavaScriptExecutor(opts ExecutorOptions) *JavaScriptExecutor {
	return &JavaScriptExecutor{
		options: opts,
	}
}

// Execute runs JavaScript code using V8
func (js *JavaScriptExecutor) Execute(ctx context.Context, code string, input string) (*ExecutionResult, error) {
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
		Language: JavaScript,
	}

	// Create a new isolate and context for each execution
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	global := v8go.NewObjectTemplate(iso)
	ctx_v8 := v8go.NewContext(iso, global)
	defer ctx_v8.Close()

	// Set up console logging
	outputs := make([]string, 0, 10)
	errors := make([]string, 0, 5)
	if err := js.setupConsole(ctx_v8, &outputs, &errors); err != nil {
		result.Error = fmt.Sprintf("Failed to setup console: %v", err)
		result.ExitCode = 1
		return result, nil
	}

	// Execute with timeout using a channel
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

	// Wait for execution or timeout
	select {
	case <-done:
		// Execution completed
		if execErr != nil {
			result.Error = execErr.Error()
			result.ExitCode = 1
		} else {
			// Include the final expression result if it's meaningful
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
		// Timeout occurred
		result.Error = "Execution timed out"
		result.ExitCode = 124
	}

	duration := time.Since(start)
	result.Duration = duration
	result.DurationString = formatDuration(duration)
	return result, nil
}

// setupConsole sets up console.log, console.error, etc.
func (js *JavaScriptExecutor) setupConsole(ctx *v8go.Context, outputs *[]string, errors *[]string) error {
	// Create console object
	console := v8go.NewObjectTemplate(ctx.Isolate())

	// console.log
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
	console.Set("info", warnFn) // info same as warn

	// Set console in global scope
	global := ctx.Global()
	consoleObj, err := console.NewInstance(ctx)
	if err != nil {
		return fmt.Errorf("failed to create console object: %w", err)
	}

	return global.Set("console", consoleObj)
}

func (js *JavaScriptExecutor) Language() Language { return JavaScript }
func (js *JavaScriptExecutor) IsAvailable() bool {
	// V8 is embedded, so it's always available once built
	return true
}
func (js *JavaScriptExecutor) Cleanup() error {
	// No cleanup needed for V8
	return nil
}
