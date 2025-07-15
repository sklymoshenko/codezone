package executor

import (
	"context"
	"time"
)

// Language represents supported programming languages
type Language string

const (
	JavaScript Language = "javascript"
	Go         Language = "go"
)

// Special exit codes for different error types
const (
	ExitCodeGoNotInstalled = 150 // Go compiler not found/installed
)

// ExecutionConfig holds configuration for code execution
type ExecutionConfig struct {
	Code     string        `json:"code"`
	Language Language      `json:"language"`
	Timeout  time.Duration `json:"timeout"`
	Input    string        `json:"input,omitempty"`
}

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	Output         string        `json:"output"`
	Error          string        `json:"error"`
	ExitCode       int           `json:"exitCode"`
	Duration       time.Duration `json:"duration"`
	DurationString string        `json:"durationString"`
	Language       Language      `json:"language"`
}

// Executor interface for different language executors
type Executor interface {
	Execute(ctx context.Context, code string, input string) (*ExecutionResult, error)
	Language() Language
	IsAvailable() bool
	Cleanup() error
}

// ExecutorOptions holds options for creating executors
type ExecutorOptions struct {
	Timeout    time.Duration
	MemoryMB   int
	MaxOutputs int
}

// DefaultExecutorOptions returns sensible default options
func DefaultExecutorOptions() ExecutorOptions {
	return ExecutorOptions{
		Timeout:    10 * time.Second,
		MemoryMB:   50,
		MaxOutputs: 1000,
	}
}
