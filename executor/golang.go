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

type GoExecutor struct {
	options ExecutorOptions
	mu      sync.Mutex
}

func NewGoExecutor(opts ExecutorOptions) *GoExecutor {
	return &GoExecutor{
		options: opts,
	}
}

func (g *GoExecutor) Execute(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	start := time.Now()

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	result := &ExecutionResult{
		Language: Go,
	}

	if !g.IsAvailable() {
		result.Error = "Go is not installed. Please install Go from https://golang.org/dl/ or install this package using your system's package manager"
		result.ExitCode = ExitCodeGoNotInstalled
		return result, nil
	}

	tempDir, err := os.MkdirTemp("", "codezone-go-*")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create temp directory: %v", err)
		result.ExitCode = 1
		return result, nil
	}
	defer os.RemoveAll(tempDir)

	goCode := g.prepareGoCode(code)

	tempFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(tempFile, []byte(goCode), 0644); err != nil {
		result.Error = fmt.Sprintf("Failed to write temp file: %v", err)
		result.ExitCode = 1
		return result, nil
	}

	cmd := exec.CommandContext(ctx, "go", "run", tempFile)

	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}

	cmd.Dir = tempDir

	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	result.Output = strings.TrimSpace(stdout.String())

	if err != nil {
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

func (g *GoExecutor) prepareGoCode(code string) string {
	if strings.Contains(code, "package ") {
		return code
	}

	hasMain := strings.Contains(code, "func main(")

	if hasMain {
		return fmt.Sprintf("package main\n\n%s", code)
	}

	return fmt.Sprintf(`package main

import "fmt"

func main() {
%s
}`, g.indentCode(code))
}

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

func (g *GoExecutor) cleanGoError(errorText string) string {
	lines := strings.Split(errorText, "\n")
	var cleanLines []string

	for _, line := range lines {
		if strings.Contains(line, "/tmp/codezone-go-") {
			parts := strings.Split(line, "/")
			for i, part := range parts {
				if strings.HasPrefix(part, "codezone-go-") {
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

func (g *GoExecutor) Language() Language {
	return Go
}

func (g *GoExecutor) IsAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

func (g *GoExecutor) Cleanup() error {
	return nil
}
