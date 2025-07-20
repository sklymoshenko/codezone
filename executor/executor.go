// Copyright (c) 2024-2025 Stanislav Klymoshenko
// Licensed under the MIT License. See LICENSE file for details.

package executor

import (
	"context"
	"fmt"
	"sync"
)

type ExecutionManager struct {
	executors map[Language]Executor
	options   ExecutorOptions
	mu        sync.RWMutex
}

func NewExecutionManager(opts ExecutorOptions) *ExecutionManager {
	manager := &ExecutionManager{
		executors: make(map[Language]Executor),
		options:   opts,
	}

	manager.executors[TypeScript] = NewTypeScriptExecutor(opts)
	manager.executors[Go] = NewGoExecutor(opts)
	manager.executors[PostgreSQL] = NewPostgreSQLExecutor(opts)

	return manager
}

func (em *ExecutionManager) Execute(config ExecutionConfig) (*ExecutionResult, error) {
	ctx := context.Background()
	var cancel context.CancelFunc
	if config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	em.mu.RLock()
	executor, exists := em.executors[config.Language]
	em.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("executor for %s is not available", config.Language)
	}

	if config.Language == PostgreSQL {
		if pgExecutor, ok := executor.(*PostgreSQLExecutor); ok {
			if config.PostgreSQLConn != nil {
				pgExecutor.SetConfig(config.PostgreSQLConn)
			}
		}
	}

	if !executor.IsAvailable() {
		return nil, fmt.Errorf("executor for %s is not available", config.Language)
	}

	return executor.Execute(ctx, config.Code, config.Input)
}

func (em *ExecutionManager) GetSupportedLanguages() []Language {
	em.mu.RLock()
	defer em.mu.RUnlock()
	languages := make([]Language, 0, len(em.executors))
	for lang := range em.executors {
		languages = append(languages, lang)
	}
	return languages
}

func (em *ExecutionManager) GetExecutor(lang Language) Executor {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.executors[lang]
}

func (em *ExecutionManager) Cleanup() {
	em.mu.Lock()
	defer em.mu.Unlock()
	for _, executor := range em.executors {
		executor.Cleanup()
	}
}

func (em *ExecutionManager) RefreshExecutor(lang Language) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if oldExecutor, ok := em.executors[lang]; ok {
		oldExecutor.Cleanup()
	}

	switch lang {
	case TypeScript:
		em.executors[TypeScript] = NewTypeScriptExecutor(em.options)
	case Go:
		em.executors[Go] = NewGoExecutor(em.options)
	default:
		return fmt.Errorf("cannot refresh unsupported language: %s", lang)
	}

	return nil
}
