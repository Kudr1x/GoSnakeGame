// Package logger provides structured logging for the application.
package logger

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var (
	log *zap.Logger
	mu  sync.RWMutex
)

// Init initializes the global logger.
func Init(development bool) error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	if development {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}

	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	return nil
}

// Get returns the global logger instance.
func Get() *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()

	if log == nil {
		return zap.NewNop()
	}

	return log
}

// Sync flushes any buffered log entries.
func Sync() error {
	mu.RLock()
	defer mu.RUnlock()

	if log != nil {
		if err := log.Sync(); err != nil {
			return fmt.Errorf("failed to sync logger: %w", err)
		}
	}

	return nil
}
