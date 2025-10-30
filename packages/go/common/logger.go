package common

import (
	"fmt"
	"time"
)

// Logger provides basic logging functionality
type Logger struct {
	prefix string
}

// NewLogger creates a new logger instance
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

// Info logs an info message
func (l *Logger) Info(message string) {
	fmt.Printf("[%s] [INFO] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), l.prefix, message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	fmt.Printf("[%s] [ERROR] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), l.prefix, message)
}
