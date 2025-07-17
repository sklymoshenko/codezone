// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package executor

import (
	"context"
	"time"
)

// Language represents supported programming languages
type Language string

const (
	JavaScript Language = "javascript"
	TypeScript Language = "typescript"
	Go         Language = "go"
	PostgreSQL Language = "postgres"
)

// Special exit codes for different error types
const (
	ExitCodeGoNotInstalled       = 150 // Go compiler not found/installed
	ExitCodePostgresNotAvailable = 151 // PostgreSQL executor not available
	ExitCodePostgresConnFailed   = 152 // PostgreSQL connection failed
	ExitCodePostgresQueryError   = 153 // PostgreSQL query execution error
	ExitCodeNodeNotAvailable     = 160 // Node.js not available
)

// ExecutionConfig holds configuration for code execution
type ExecutionConfig struct {
	Code           string            `json:"code"`
	Language       Language          `json:"language"`
	Timeout        time.Duration     `json:"timeout"`
	Input          string            `json:"input,omitempty"`
	PostgreSQLConn *PostgreSQLConfig `json:"postgresqlConn,omitempty"`
}

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	Output         string          `json:"output"`
	Error          string          `json:"error"`
	ExitCode       int             `json:"exitCode"`
	Duration       time.Duration   `json:"duration"`
	DurationString string          `json:"durationString"`
	Language       Language        `json:"language"`
	SQLResult      *SQLQueryResult `json:"sqlResult,omitempty"`
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

type PostgreSQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"sslMode"`
}

type SQLQueryResult struct {
	QueryType     string          `json:"queryType"`    // SELECT, INSERT, UPDATE, etc.
	Columns       []string        `json:"columns"`      // Column names
	Rows          [][]interface{} `json:"rows"`         // Data rows
	RowsAffected  int64           `json:"rowsAffected"` // For non-SELECT queries
	ExecutionTime time.Duration   `json:"executionTime"`
}
