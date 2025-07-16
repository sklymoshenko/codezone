// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"codezone-wails/executor"
)

// App struct
type App struct {
	ctx     context.Context
	execMgr *executor.ExecutionManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	opts := executor.DefaultExecutorOptions()
	// Allow a generous timeout for potentially long-running code.
	opts.Timeout = 15 * time.Second
	opts.MemoryMB = 128 // 128MB memory limit per execution context.

	return &App{
		execMgr: executor.NewExecutionManager(opts),
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// onBeforeClose is called just before the application shuts down.
// It's the ideal place to clean up resources.
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
	log.Println("Application: Starting shutdown process...")

	if a.execMgr != nil {
		// Explicitly disconnect PostgreSQL if connected
		pgExecutor, ok := a.execMgr.GetExecutor(executor.PostgreSQL).(*executor.PostgreSQLExecutor)
		if ok && pgExecutor.IsConnected() {
			log.Println("Application: Disconnecting from PostgreSQL before shutdown...")
			if err := pgExecutor.Cleanup(); err != nil {
				log.Printf("Application: Error during PostgreSQL cleanup: %v", err)
			} else {
				log.Println("Application: PostgreSQL disconnected successfully")
			}
		}

		// Clean up all executors
		log.Println("Application: Cleaning up all executors...")
		a.execMgr.Cleanup()
	}

	log.Println("Application: Shutdown process completed")
	return false
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetGoVersion() string {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "Error getting Go version"
	}
	version := strings.TrimSpace(string(output))
	parts := strings.Fields(version)
	if len(parts) >= 3 {
		// Expected format: "go version go1.22.4 linux/amd64"
		// Extract "go1.22.4" and remove "go" prefix
		return "go v" + strings.TrimPrefix(parts[2], "go")
	}
	return "Unknown Go version"
}

// ExecuteCode executes code using the persistent execution manager.
func (a *App) ExecuteCode(config executor.ExecutionConfig) (*executor.ExecutionResult, error) {
	return a.execMgr.Execute(config)
}

// GetSupportedLanguages returns available languages.
func (a *App) GetSupportedLanguages() []executor.Language {
	return a.execMgr.GetSupportedLanguages()
}

// RefreshExecutor creates a new, clean execution environment for a language.
func (a *App) RefreshExecutor(lang executor.Language) error {
	return a.execMgr.RefreshExecutor(lang)
}

// HadleConnection creates pool and tests PostgreSQL connection
func (a *App) HadleConnection(config *executor.PostgreSQLConfig) (bool, error) {
	log.Printf("PostgreSQL: Attempting connection to %s:%d/%s as user %s",
		config.Host, config.Port, config.Database, config.Username)

	if a.execMgr == nil {
		log.Println("PostgreSQL: Connection failed - execution manager not initialized")
		return false, fmt.Errorf("execution manager not initialized")
	}

	// Get the PostgreSQL executor
	pgExecutor, ok := a.execMgr.GetExecutor(executor.PostgreSQL).(*executor.PostgreSQLExecutor)
	if !ok {
		log.Println("PostgreSQL: Connection failed - PostgreSQL executor not available")
		return false, fmt.Errorf("PostgreSQL executor not available")
	}

	// Create pool with the config (this also sets the config)
	err := pgExecutor.CreatePgPool(a.ctx, config)
	if err != nil {
		log.Printf("PostgreSQL: Connection failed - pool creation error: %v", err)
		return false, err
	}

	// Test the pool
	err = pgExecutor.TestConnection(a.ctx, config)
	if err != nil {
		log.Printf("PostgreSQL: Connection failed - connection test error: %v", err)
		return false, err
	}

	log.Printf("PostgreSQL: Successfully connected to %s:%d/%s",
		config.Host, config.Port, config.Database)
	return true, nil
}

// SetPostgreSQLConfig sets the PostgreSQL connection configuration
func (a *App) SetPostgreSQLConfig(config *executor.PostgreSQLConfig) error {
	if a.execMgr == nil {
		return fmt.Errorf("execution manager not initialized")
	}

	// Get the PostgreSQL executor
	pgExecutor, ok := a.execMgr.GetExecutor(executor.PostgreSQL).(*executor.PostgreSQLExecutor)
	if !ok {
		return fmt.Errorf("PostgreSQL executor not available")
	}

	// Set the configuration (pool should already be created by TestPostgreSQLConnection)
	pgExecutor.SetConfig(config)
	return nil
}

// GetPostgreSQLConnectionStatus returns the current PostgreSQL connection status
func (a *App) GetPostgreSQLConnectionStatus() (bool, error) {
	if a.execMgr == nil {
		return false, fmt.Errorf("execution manager not initialized")
	}

	// Get the PostgreSQL executor
	pgExecutor, ok := a.execMgr.GetExecutor(executor.PostgreSQL).(*executor.PostgreSQLExecutor)
	if !ok {
		return false, fmt.Errorf("PostgreSQL executor not available")
	}

	// Return the actual connection status
	return pgExecutor.IsConnected(), nil
}

// DisconnectPostgreSQL disconnects from PostgreSQL database
func (a *App) DisconnectPostgreSQL() error {
	log.Println("PostgreSQL: Attempting to disconnect from database")

	if a.execMgr == nil {
		log.Println("PostgreSQL: Disconnection failed - execution manager not initialized")
		return fmt.Errorf("execution manager not initialized")
	}

	// Get the PostgreSQL executor
	pgExecutor, ok := a.execMgr.GetExecutor(executor.PostgreSQL).(*executor.PostgreSQLExecutor)
	if !ok {
		log.Println("PostgreSQL: Disconnection failed - PostgreSQL executor not available")
		return fmt.Errorf("PostgreSQL executor not available")
	}

	// Cleanup the connection
	err := pgExecutor.Cleanup()
	if err != nil {
		log.Printf("PostgreSQL: Disconnection failed - cleanup error: %v", err)
		return err
	}

	log.Println("PostgreSQL: Successfully disconnected from database")
	return nil
}
