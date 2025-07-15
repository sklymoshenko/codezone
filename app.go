// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package main

import (
	"context"
	"fmt"
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
	if a.execMgr != nil {
		a.execMgr.Cleanup()
	}
	return false
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
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

// TestPostgreSQLConnection tests a PostgreSQL connection configuration
func (a *App) TestPostgreSQLConnection(config *executor.PostgreSQLConfig) (bool, error) {
	if a.execMgr == nil {
		return false, fmt.Errorf("execution manager not initialized")
	}

	// Get the PostgreSQL executor
	pgExecutor, ok := a.execMgr.GetExecutor(executor.PostgreSQL).(*executor.PostgreSQLExecutor)
	if !ok {
		return false, fmt.Errorf("PostgreSQL executor not available")
	}

	// Test the connection
	err := pgExecutor.TestConnection(a.ctx, config)
	return err == nil, err
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

	// Set the configuration
	pgExecutor.SetConfig(config)
	return nil
}
