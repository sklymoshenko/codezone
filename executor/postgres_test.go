// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package executor

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Test configuration for PostgreSQL (can be overridden with env vars)
func getTestPostgreSQLConfig() *PostgreSQLConfig {
	config := &PostgreSQLConfig{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefaultInt("POSTGRES_PORT", 5433),
		Database: getEnvOrDefault("POSTGRES_DB", "testdb"),
		Username: getEnvOrDefault("POSTGRES_USER", "testuser"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "testpassword"),
		SSLMode:  "disable",
	}

	return config
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Add this helper function for integer environment variables
func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper function to check if PostgreSQL is available for testing
func isPostgreSQLAvailable() bool {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())
	config := getTestPostgreSQLConfig()
	executor.SetConfig(config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create pool with the config (this also sets the config)
	err := executor.CreatePgPool(ctx, config)
	if err != nil {
		log.Printf("PostgreSQL: Connection failed - pool creation error: %v", err)
		return false
	}

	return executor.TestConnection(ctx, config) == nil
}

func TestPostgreSQLExecutor_Basic(t *testing.T) {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())

	t.Run("should return correct language", func(t *testing.T) {
		if executor.Language() != PostgreSQL {
			t.Errorf("Expected language PostgreSQL, got %s", executor.Language())
		}
	})

	t.Run("should not be available without config", func(t *testing.T) {
		if executor.IsAvailable() {
			t.Error("Expected executor to not be available without configuration")
		}
	})

	t.Run("should be available with valid config", func(t *testing.T) {
		config := getTestPostgreSQLConfig()
		executor.SetConfig(config)

		if !executor.IsAvailable() {
			t.Error("Expected executor to be available with configuration")
		}
	})

	t.Run("should handle cleanup", func(t *testing.T) {
		err := executor.Cleanup()
		if err != nil {
			t.Errorf("Cleanup should not return error, got %v", err)
		}
	})
}

func TestPostgreSQLExecutor_Configuration(t *testing.T) {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())

	t.Run("should set and use configuration", func(t *testing.T) {
		config := getTestPostgreSQLConfig()

		executor.SetConfig(config)

		if !executor.IsAvailable() {
			t.Error("Expected executor to be available after setting config")
		}

		// Test connection string building (via buildConnectionString method)
		expectedConnStr := "host=localhost port=5433 dbname=testdb user=testuser password=testpassword sslmode=disable"
		actualConnStr := executor.buildConnectionString()

		if actualConnStr != expectedConnStr {
			t.Errorf("Expected connection string %q, got %q", expectedConnStr, actualConnStr)
		}
	})

	t.Run("should handle missing SSL mode", func(t *testing.T) {
		config := getTestPostgreSQLConfig()

		executor.SetConfig(config)
		connStr := executor.buildConnectionString()

		if !contains(connStr, "sslmode=disable") {
			t.Errorf("Expected default sslmode=disable in connection string, got %q", connStr)
		}
	})
}

func TestPostgreSQLExecutor_QueryTypeDetection(t *testing.T) {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())

	testCases := []struct {
		name     string
		query    string
		expected string
	}{
		{"Simple SELECT", "SELECT * FROM users", "SELECT"},
		{"SELECT with whitespace", "  select name from users  ", "SELECT"},
		{"WITH query", "WITH cte AS (SELECT 1) SELECT * FROM cte", "WITH"},
		{"INSERT query", "INSERT INTO users (name) VALUES ('John')", "INSERT"},
		{"UPDATE query", "UPDATE users SET name = 'Jane'", "UPDATE"},
		{"DELETE query", "DELETE FROM users WHERE id = 1", "DELETE"},
		{"CREATE query", "CREATE TABLE test (id INT)", "CREATE"},
		{"DROP query", "DROP TABLE test", "DROP"},
		{"ALTER query", "ALTER TABLE users ADD COLUMN email VARCHAR(100)", "ALTER"},
		{"Unknown query", "EXPLAIN SELECT * FROM users", "OTHER"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := executor.detectQueryType(tc.query)
			if result != tc.expected {
				t.Errorf("Expected query type %q, got %q for query %q", tc.expected, result, tc.query)
			}
		})
	}
}

func TestPostgreSQLExecutor_IsSelectQuery(t *testing.T) {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())

	testCases := []struct {
		queryType string
		expected  bool
	}{
		{"SELECT", true},
		{"select", true},
		{"  SELECT  ", true},
		{"WITH", true},
		{"with", true},
		{"INSERT", false},
		{"UPDATE", false},
		{"DELETE", false},
		{"CREATE", false},
		{"DROP", false},
		{"ALTER", false},
		{"OTHER", false},
	}

	for _, tc := range testCases {
		t.Run(tc.queryType, func(t *testing.T) {
			result := executor.isSelectQuery(tc.queryType)
			if result != tc.expected {
				t.Errorf("Expected isSelectQuery(%q) = %v, got %v", tc.queryType, tc.expected, result)
			}
		})
	}
}

func TestPostgreSQLExecutor_PrepareSQLCode(t *testing.T) {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Simple query",
			"SELECT * FROM users",
			"SELECT * FROM users",
		},
		{
			"Query with comments",
			`-- This is a comment
			SELECT * FROM users -- inline comment
			WHERE id = 1`,
			"SELECT * FROM users\nWHERE id = 1",
		},
		{
			"Query with empty lines",
			`
			SELECT * FROM users
			
			WHERE id = 1
			`,
			"SELECT * FROM users\nWHERE id = 1",
		},
		{
			"Only comments",
			`-- Comment 1
			-- Comment 2`,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := executor.prepareSQLCode(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestPostgreSQLExecutor_ConvertValue(t *testing.T) {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())

	testCases := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"nil value", nil, nil},
		{"string value", "hello", "hello"},
		{"int value", 42, 42},
		{"byte slice", []byte("hello"), "hello"},
		{"time value", time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), "2023-01-01T12:00:00Z"},
		{"bool value", true, true},
		{"float value", 3.14, 3.14},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := executor.convertValue(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestPostgreSQLExecutor_ExecuteWithoutConnection(t *testing.T) {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())

	t.Run("should fail without configuration", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		result, err := executor.Execute(ctx, "SELECT 1", "")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != ExitCodePostgresNotAvailable {
			t.Errorf("Expected exit code %d, got %d", ExitCodePostgresNotAvailable, result.ExitCode)
		}

		if result.Error == "" {
			t.Error("Expected error message")
		}

		if result.Language != PostgreSQL {
			t.Errorf("Expected language PostgreSQL, got %s", result.Language)
		}
	})

	t.Run("should fail with empty query", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		config := getTestPostgreSQLConfig()
		executor.SetConfig(config)

		result, err := executor.Execute(ctx, "", "")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != ExitCodePostgresQueryError {
			t.Errorf("Expected exit code %d, got %d", ExitCodePostgresQueryError, result.ExitCode)
		}
	})

	t.Run("should fail with only comments", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		config := getTestPostgreSQLConfig()
		executor.SetConfig(config)

		result, err := executor.Execute(ctx, "-- Just a comment", "")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != ExitCodePostgresQueryError {
			t.Errorf("Expected exit code %d, got %d", ExitCodePostgresQueryError, result.ExitCode)
		}
	})
}

// Integration tests - require live PostgreSQL
func TestPostgreSQLExecutor_Integration(t *testing.T) {
	if !isPostgreSQLAvailable() {
		t.Skip("PostgreSQL not available for integration testing. Set POSTGRES_HOST, POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD env vars to run these tests.")
	}

	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())
	config := getTestPostgreSQLConfig()
	executor.SetConfig(config)

	t.Run("should execute simple SELECT query", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := executor.Execute(ctx, "SELECT 1 as test_column", "")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got %d. Error: %s", result.ExitCode, result.Error)
		}

		if result.SQLResult == nil {
			t.Fatal("Expected SQLResult to be set")
		}

		if result.SQLResult.QueryType != "SELECT" {
			t.Errorf("Expected query type SELECT, got %s", result.SQLResult.QueryType)
		}

		if len(result.SQLResult.Columns) != 1 || result.SQLResult.Columns[0] != "test_column" {
			t.Errorf("Expected columns [test_column], got %v", result.SQLResult.Columns)
		}

		if len(result.SQLResult.Rows) != 1 {
			t.Errorf("Expected 1 row, got %d", len(result.SQLResult.Rows))
		}

		if result.Duration <= 0 {
			t.Error("Expected positive duration")
		}
	})

	t.Run("should handle query timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Use a query that might take longer than 1ms
		result, err := executor.Execute(ctx, "SELECT pg_sleep(1)", "")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != 124 { // Timeout exit code
			t.Errorf("Expected timeout exit code 124, got %d", result.ExitCode)
		}

		if !contains(result.Error, "timed out") {
			t.Errorf("Expected timeout error, got %s", result.Error)
		}
	})

	t.Run("should handle invalid SQL", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := executor.Execute(ctx, "INVALID SQL SYNTAX", "")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ExitCode != ExitCodePostgresQueryError {
			t.Errorf("Expected exit code %d, got %d", ExitCodePostgresQueryError, result.ExitCode)
		}

		if result.Error == "" {
			t.Error("Expected error message for invalid SQL")
		}
	})
}

func TestPostgreSQLExecutor_ConnectionTesting(t *testing.T) {
	executor := NewPostgreSQLExecutor(DefaultExecutorOptions())

	t.Run("should test valid connection", func(t *testing.T) {
		if !isPostgreSQLAvailable() {
			t.Skip("PostgreSQL not available for connection testing")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		config := getTestPostgreSQLConfig()

		err := executor.CreatePgPool(context.Background(), config)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		err = executor.TestConnection(ctx, config)
		if err != nil {
			t.Errorf("Expected successful connection test, got error: %v", err)
		}
	})

	t.Run("should fail with invalid connection", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		config := &PostgreSQLConfig{
			Host:     "nonexistent-host",
			Port:     9999,
			Database: "nonexistent",
			Username: "nonexistent",
			Password: "wrong",
			SSLMode:  "disable",
		}

		err := executor.CreatePgPool(context.Background(), config)
		if err == nil {
			t.Error("Expected connection test to fail with invalid config")
		}

		err = executor.TestConnection(ctx, config)

		if err == nil {
			t.Error("Expected connection test to fail with invalid config, got nil")
		}
	})

	t.Run("should fail with nil config", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := executor.CreatePgPool(context.Background(), nil)
		if err == nil {
			t.Error("Expected connection test to fail with nil config")
		}

		err = executor.TestConnection(ctx, nil)
		if err == nil {
			t.Error("Expected connection test to fail with nil config")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
