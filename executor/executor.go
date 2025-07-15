package executor

import (
	"context"
	"fmt"
	"sync"
)

// ExecutionManager manages multiple language executors
type ExecutionManager struct {
	executors map[Language]Executor
	options   ExecutorOptions
	mu        sync.RWMutex
}

// NewExecutionManager creates a new execution manager
func NewExecutionManager(opts ExecutorOptions) *ExecutionManager {
	manager := &ExecutionManager{
		executors: make(map[Language]Executor),
		options:   opts,
	}

	// Initialize all supported executors
	manager.executors[JavaScript] = NewJavaScriptExecutor(opts)
	manager.executors[Go] = NewGoExecutor(opts)

	return manager
}

// Execute runs code in the specified language
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

	if !exists || !executor.IsAvailable() {
		return nil, fmt.Errorf("executor for %s is not available", config.Language)
	}

	return executor.Execute(ctx, config.Code, config.Input)
}

// GetSupportedLanguages returns a list of all supported languages.
func (em *ExecutionManager) GetSupportedLanguages() []Language {
	em.mu.RLock()
	defer em.mu.RUnlock()
	languages := make([]Language, 0, len(em.executors))
	for lang := range em.executors {
		languages = append(languages, lang)
	}
	return languages
}

// Cleanup releases all executor resources
func (em *ExecutionManager) Cleanup() {
	em.mu.Lock()
	defer em.mu.Unlock()
	for _, executor := range em.executors {
		executor.Cleanup()
	}
}

// RefreshExecutor safely recreates an executor for a clean state.
func (em *ExecutionManager) RefreshExecutor(lang Language) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if oldExecutor, ok := em.executors[lang]; ok {
		oldExecutor.Cleanup()
	}

	// Re-create the specific executor
	switch lang {
	case JavaScript:
		em.executors[JavaScript] = NewJavaScriptExecutor(em.options)
	case Go:
		em.executors[Go] = NewGoExecutor(em.options)
	default:
		return fmt.Errorf("cannot refresh unsupported language: %s", lang)
	}

	return nil
}
