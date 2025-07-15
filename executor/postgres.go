// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package executor

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQLExecutor implements PostgreSQL execution using pgx
type PostgreSQLExecutor struct {
	options ExecutorOptions
	pool    *pgxpool.Pool
	config  *PostgreSQLConfig
	mu      sync.Mutex
}

// NewPostgreSQLExecutor creates a new PostgreSQL executor
func NewPostgreSQLExecutor(opts ExecutorOptions) *PostgreSQLExecutor {
	return &PostgreSQLExecutor{
		options: opts,
	}
}

// Execute runs SQL queries against PostgreSQL database
func (p *PostgreSQLExecutor) Execute(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	start := time.Now()

	// Add default timeout if context doesn't have one
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Use mutex to ensure thread safety
	p.mu.Lock()
	defer p.mu.Unlock()

	result := &ExecutionResult{
		Language: PostgreSQL,
	}

	// Check if PostgreSQL connection is available
	if !p.isAvailableInternal() {
		result.Error = "PostgreSQL connection is not configured or unavailable"
		result.ExitCode = ExitCodePostgresNotAvailable
		return result, nil
	}

	// Clean and prepare SQL code first (before attempting connection)
	sqlCode := p.prepareSQLCode(code)
	if strings.TrimSpace(sqlCode) == "" {
		result.Error = "No SQL query provided"
		result.ExitCode = ExitCodePostgresQueryError
		return result, nil
	}

	// Ensure we have a connection pool
	if err := p.ensureConnection(ctx); err != nil {
		result.Error = fmt.Sprintf("Failed to connect to PostgreSQL: %v", err)
		result.ExitCode = ExitCodePostgresConnFailed
		return result, nil
	}

	// Execute the SQL query
	sqlResult, err := p.executeSQL(ctx, sqlCode)
	if err != nil {
		// Handle timeout
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "Query execution timed out"
			result.ExitCode = 124
		} else {
			result.Error = fmt.Sprintf("SQL execution error: %v", err)
			result.ExitCode = ExitCodePostgresQueryError
		}
		return result, nil
	}

	// Set results
	result.SQLResult = sqlResult
	result.Output = p.formatQueryOutput(sqlResult)
	result.ExitCode = 0

	// Calculate duration
	result.Duration = time.Since(start)
	result.DurationString = formatDuration(result.Duration)

	return result, nil
}

// ensureConnection establishes connection pool if not already connected
func (p *PostgreSQLExecutor) ensureConnection(ctx context.Context) error {
	if p.pool != nil {
		// Test existing connection
		if err := p.pool.Ping(ctx); err == nil {
			return nil
		}
		// Close bad connection
		p.pool.Close()
		p.pool = nil
	}

	if p.config == nil {
		return fmt.Errorf("no PostgreSQL configuration provided")
	}

	// Build connection string
	connStr := p.buildConnectionString()

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection configuration: %w", err)
	}

	// Set reasonable pool settings
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.pool = pool
	return nil
}

// buildConnectionString creates PostgreSQL connection string from config
func (p *PostgreSQLExecutor) buildConnectionString() string {
	sslMode := p.config.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}

	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.Database,
		p.config.Username,
		p.config.Password,
		sslMode,
	)
}

// executeSQL executes SQL query and returns structured result
func (p *PostgreSQLExecutor) executeSQL(ctx context.Context, sqlCode string) (*SQLQueryResult, error) {
	queryStart := time.Now()

	// Determine query type
	queryType := p.detectQueryType(sqlCode)

	result := &SQLQueryResult{
		QueryType:     queryType,
		ExecutionTime: 0,
	}

	if p.isSelectQuery(queryType) {
		// Handle SELECT queries
		rows, err := p.pool.Query(ctx, sqlCode)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Get column descriptions
		fieldDescriptions := rows.FieldDescriptions()
		columns := make([]string, len(fieldDescriptions))
		for i, fd := range fieldDescriptions {
			columns[i] = string(fd.Name)
		}
		result.Columns = columns

		// Collect rows
		var allRows [][]interface{}
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				return nil, err
			}

			// Convert values for JSON serialization
			row := make([]interface{}, len(values))
			for i, val := range values {
				row[i] = p.convertValue(val)
			}
			allRows = append(allRows, row)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		result.Rows = allRows
		result.RowsAffected = int64(len(allRows))
	} else {
		// Handle non-SELECT queries (INSERT, UPDATE, DELETE, etc.)
		commandTag, err := p.pool.Exec(ctx, sqlCode)
		if err != nil {
			return nil, err
		}

		result.RowsAffected = commandTag.RowsAffected()
		result.Columns = []string{"Rows Affected"}
		result.Rows = [][]interface{}{{result.RowsAffected}}
	}

	result.ExecutionTime = time.Since(queryStart)
	return result, nil
}

// detectQueryType determines the type of SQL query
func (p *PostgreSQLExecutor) detectQueryType(sqlCode string) string {
	trimmed := strings.TrimSpace(strings.ToUpper(sqlCode))

	switch {
	case strings.HasPrefix(trimmed, "SELECT"):
		return "SELECT"
	case strings.HasPrefix(trimmed, "INSERT"):
		return "INSERT"
	case strings.HasPrefix(trimmed, "UPDATE"):
		return "UPDATE"
	case strings.HasPrefix(trimmed, "DELETE"):
		return "DELETE"
	case strings.HasPrefix(trimmed, "CREATE"):
		return "CREATE"
	case strings.HasPrefix(trimmed, "DROP"):
		return "DROP"
	case strings.HasPrefix(trimmed, "ALTER"):
		return "ALTER"
	case strings.HasPrefix(trimmed, "WITH"):
		return "WITH"
	default:
		return "OTHER"
	}
}

// isSelectQuery checks if the query type returns rows (SELECT or WITH queries)
func (p *PostgreSQLExecutor) isSelectQuery(queryType string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(queryType))
	return strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "WITH")
}

// convertValue converts PostgreSQL values to JSON-serializable types
func (p *PostgreSQLExecutor) convertValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case []byte:
		return string(v)
	case time.Time:
		return v.Format(time.RFC3339)
	case pgx.Rows:
		return "[nested result]"
	default:
		// Handle arrays and other complex types
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() != reflect.Uint8 {
			// Convert array to string representation
			return fmt.Sprintf("%v", val)
		}
		return val
	}
}

// prepareSQLCode cleans and prepares SQL code for execution
func (p *PostgreSQLExecutor) prepareSQLCode(code string) string {
	// Remove common SQL comments and clean up
	lines := strings.Split(code, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and SQL comments
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		// Remove inline comments
		if idx := strings.Index(line, "--"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
			if line == "" {
				continue
			}
		}

		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, "\n")
}

// formatQueryOutput creates a human-readable output from SQL results
func (p *PostgreSQLExecutor) formatQueryOutput(sqlResult *SQLQueryResult) string {
	if sqlResult == nil {
		return "No results"
	}

	var output strings.Builder

	output.WriteString(fmt.Sprintf("Query Type: %s\n", sqlResult.QueryType))
	output.WriteString(fmt.Sprintf("Execution Time: %s\n", formatDuration(sqlResult.ExecutionTime)))

	if p.isSelectQuery(sqlResult.QueryType) {
		output.WriteString(fmt.Sprintf("Rows Returned: %d\n\n", len(sqlResult.Rows)))

		if len(sqlResult.Rows) > 0 && len(sqlResult.Columns) > 0 {
			// Simple table format for console output
			output.WriteString(strings.Join(sqlResult.Columns, " | "))
			output.WriteString("\n")
			output.WriteString(strings.Repeat("-", len(strings.Join(sqlResult.Columns, " | "))))
			output.WriteString("\n")

			// Show first 100 rows to avoid overwhelming output
			maxRows := len(sqlResult.Rows)
			if maxRows > 100 {
				maxRows = 100
			}

			for i := 0; i < maxRows; i++ {
				row := sqlResult.Rows[i]
				stringRow := make([]string, len(row))
				for j, val := range row {
					if val == nil {
						stringRow[j] = "NULL"
					} else {
						stringRow[j] = fmt.Sprintf("%v", val)
					}
				}
				output.WriteString(strings.Join(stringRow, " | "))
				output.WriteString("\n")
			}

			if len(sqlResult.Rows) > 100 {
				output.WriteString(fmt.Sprintf("... and %d more rows\n", len(sqlResult.Rows)-100))
			}
		}
	} else {
		output.WriteString(fmt.Sprintf("Rows Affected: %d\n", sqlResult.RowsAffected))
	}

	return output.String()
}

// SetConfig sets the PostgreSQL connection configuration
func (p *PostgreSQLExecutor) SetConfig(config *PostgreSQLConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config

	// Close existing connection if config changed
	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}
}

// TestConnection tests the PostgreSQL connection without executing queries
func (p *PostgreSQLExecutor) TestConnection(ctx context.Context, config *PostgreSQLConfig) error {
	if config == nil {
		return fmt.Errorf("no configuration provided")
	}

	// Temporarily set config for testing
	oldConfig := p.config
	p.config = config
	defer func() { p.config = oldConfig }()

	// Test connection
	err := p.ensureConnection(ctx)
	if p.pool != nil && p.config == config {
		// Only close if we're testing a different config
		p.pool.Close()
		p.pool = nil
	}

	return err
}

// Language returns the language identifier
func (p *PostgreSQLExecutor) Language() Language {
	return PostgreSQL
}

// IsAvailable checks if PostgreSQL connection is configured
func (p *PostgreSQLExecutor) IsAvailable() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.isAvailableInternal()
}

// isAvailableInternal checks availability without locking (internal use only)
func (p *PostgreSQLExecutor) isAvailableInternal() bool {
	return p.config != nil &&
		p.config.Host != "" &&
		p.config.Database != "" &&
		p.config.Username != ""
}

// Cleanup closes database connections and performs cleanup
func (p *PostgreSQLExecutor) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}

	return nil
}
