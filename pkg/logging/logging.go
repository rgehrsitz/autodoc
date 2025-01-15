// autodoc/pkg/logging/logging.go

package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Logger represents a custom logger with additional functionality
type Logger struct {
	*log.Logger
	logFile *os.File
}

// NewLogger creates a new logger instance that writes to both file and stdout
func NewLogger(logPath string) (*Logger, error) {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer logger
	logger := log.New(f, "", log.LstdFlags)

	return &Logger{
		Logger:  logger,
		logFile: f,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}