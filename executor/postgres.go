// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package executor

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

type PostgreSQLExecutor struct {
	options ExecutorOptions
	pool    *pgxpool.Pool
	config  *PostgreSQLConfig
	mu      sync.Mutex
}

func NewPostgreSQLExecutor(opts ExecutorOptions) *PostgreSQLExecutor {
	return &PostgreSQLExecutor{
		options: opts,
	}
}

func (p *PostgreSQLExecutor) Execute(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	start := time.Now()

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	result := &ExecutionResult{
		Language: PostgreSQL,
	}

	if !p.isAvailableInternal() {
		result.Error = "PostgreSQL connection is not configured or unavailable"
		result.ExitCode = ExitCodePostgresNotAvailable
		return result, nil
	}

	sqlCode := p.prepareSQLCode(code)
	if strings.TrimSpace(sqlCode) == "" {
		result.Error = "No SQL query provided"
		result.ExitCode = ExitCodePostgresQueryError
		return result, nil
	}

	if err := p.ensureConnection(ctx); err != nil {
		result.Error = fmt.Sprintf("Failed to connect to PostgreSQL: %v", err)
		result.ExitCode = ExitCodePostgresConnFailed
		return result, nil
	}

	sqlResult, err := p.executeSQL(ctx, sqlCode)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "Query execution timed out"
			result.ExitCode = 124
		} else {
			result.Error = fmt.Sprintf("SQL execution error: %v", err)
			result.ExitCode = ExitCodePostgresQueryError
		}
		return result, nil
	}

	result.SQLResult = sqlResult
	result.Output = p.formatQueryOutput(sqlResult)
	result.ExitCode = 0

	result.Duration = time.Since(start)
	result.DurationString = formatDuration(result.Duration)

	return result, nil
}

func (p *PostgreSQLExecutor) ensureConnection(ctx context.Context) error {
	if p.pool != nil {
		log.Println("PostgreSQL Executor: Testing existing connection pool")
		if err := p.pool.Ping(ctx); err == nil {
			log.Println("PostgreSQL Executor: Existing connection pool is healthy")
			return nil
		}
		log.Println("PostgreSQL Executor: Existing connection pool is unhealthy, closing it")
		p.pool.Close()
		p.pool = nil
	}

	if p.config == nil {
		log.Println("PostgreSQL Executor: ensureConnection failed - no configuration provided")
		return fmt.Errorf("no PostgreSQL configuration provided")
	}

	log.Printf("PostgreSQL Executor: Building connection string for %s:%d/%s",
		p.config.Host, p.config.Port, p.config.Database)

	connStr := p.buildConnectionString()

	log.Println("PostgreSQL Executor: Parsing connection configuration")

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Printf("PostgreSQL Executor: Invalid connection configuration: %v", err)
		return fmt.Errorf("invalid connection configuration: %w", err)
	}

	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	log.Printf("PostgreSQL Executor: Creating connection pool (MaxConns: %d, MinConns: %d)",
		poolConfig.MaxConns, poolConfig.MinConns)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Printf("PostgreSQL Executor: Failed to create connection pool: %v", err)
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	log.Println("PostgreSQL Executor: Testing new connection pool with ping")
	if err := pool.Ping(ctx); err != nil {
		log.Printf("PostgreSQL Executor: Failed to ping database: %v", err)
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("PostgreSQL Executor: Connection pool created and tested successfully")
	p.pool = pool
	return nil
}

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

func (p *PostgreSQLExecutor) executeSQL(ctx context.Context, sqlCode string) (*SQLQueryResult, error) {
	queryStart := time.Now()

	queryType := p.detectQueryType(sqlCode)

	result := &SQLQueryResult{
		QueryType:     queryType,
		ExecutionTime: 0,
	}

	if p.isSelectQuery(queryType) {
		rows, err := p.pool.Query(ctx, sqlCode)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		fieldDescriptions := rows.FieldDescriptions()
		columns := make([]string, len(fieldDescriptions))
		for i, fd := range fieldDescriptions {
			columns[i] = string(fd.Name)
		}
		result.Columns = columns

		var allRows [][]interface{}
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				return nil, err
			}

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

func (p *PostgreSQLExecutor) isSelectQuery(queryType string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(queryType))
	return strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "WITH")
}

func (p *PostgreSQLExecutor) convertValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case uuid.UUID:
		return v.String()
	case []byte:
		return string(v)
	case time.Time:
		return v.Format(time.RFC3339)
	case pgx.Rows:
		return "[nested result]"
	default:
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() != reflect.Uint8 {
			return fmt.Sprintf("%v", val)
		}
		return val
	}
}

func (p *PostgreSQLExecutor) prepareSQLCode(code string) string {
	lines := strings.Split(code, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

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
			output.WriteString(strings.Join(sqlResult.Columns, " | "))
			output.WriteString("\n")
			output.WriteString(strings.Repeat("-", len(strings.Join(sqlResult.Columns, " | "))))
			output.WriteString("\n")

			for i := 0; i < len(sqlResult.Rows); i++ {
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

func (p *PostgreSQLExecutor) SetConfig(config *PostgreSQLConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config

	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}
}

func (p *PostgreSQLExecutor) CreatePgPool(ctx context.Context, config *PostgreSQLConfig) error {
	if config == nil {
		log.Println("PostgreSQL Executor: CreatePgPool failed - no configuration provided")
		return fmt.Errorf("no configuration provided")
	}

	log.Printf("PostgreSQL Executor: Creating connection pool for %s:%d/%s",
		config.Host, config.Port, config.Database)

	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config

	if p.pool != nil {
		log.Println("PostgreSQL Executor: Closing existing connection pool")
		p.pool.Close()
		p.pool = nil
	}

	err := p.ensureConnection(ctx)
	if err != nil {
		log.Printf("PostgreSQL Executor: Pool creation failed: %v", err)
		return err
	}

	log.Printf("PostgreSQL Executor: Successfully created connection pool for %s:%d/%s",
		config.Host, config.Port, config.Database)
	return nil
}

func (p *PostgreSQLExecutor) TestConnection(ctx context.Context, config *PostgreSQLConfig) error {
	if config == nil {
		log.Println("PostgreSQL Executor: TestConnection failed - no configuration provided")
		return fmt.Errorf("no configuration provided")
	}

	log.Printf("PostgreSQL Executor: Testing connection to %s:%d/%s",
		config.Host, config.Port, config.Database)

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool == nil {
		log.Println("PostgreSQL Executor: TestConnection failed - no connection pool available")
		return fmt.Errorf("no connection pool available - call CreatePgPool first")
	}

	err := p.pool.Ping(ctx)
	if err != nil {
		log.Printf("PostgreSQL Executor: Connection test failed: %v", err)
		return err
	}

	log.Printf("PostgreSQL Executor: Connection test successful for %s:%d/%s",
		config.Host, config.Port, config.Database)
	return nil
}

func (p *PostgreSQLExecutor) Language() Language {
	return PostgreSQL
}

func (p *PostgreSQLExecutor) IsAvailable() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.isAvailableInternal()
}

func (p *PostgreSQLExecutor) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isAvailableInternal() {
		return false
	}

	if p.pool == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := p.pool.Ping(ctx)
	if err != nil {
		log.Printf("PostgreSQL Executor: Connection status check failed: %v", err)
		return false
	}

	return true
}

func (p *PostgreSQLExecutor) isAvailableInternal() bool {
	return p.config != nil &&
		p.config.Host != "" &&
		p.config.Database != "" &&
		p.config.Username != ""
}

func (p *PostgreSQLExecutor) Cleanup() error {
	log.Println("PostgreSQL Executor: Starting cleanup process")

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		log.Println("PostgreSQL Executor: Closing connection pool")
		p.pool.Close()
		p.pool = nil
		log.Println("PostgreSQL Executor: Connection pool closed successfully")
	} else {
		log.Println("PostgreSQL Executor: No active connection pool to close")
	}

	return nil
}
