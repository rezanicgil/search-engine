// logger.go - Structured logging utilities
// Provides consistent logging interface across the application

package logger

import (
	"log"
	"os"
)

var (
	// InfoLogger logs informational messages
	InfoLogger *log.Logger
	// ErrorLogger logs error messages
	ErrorLogger *log.Logger
	// DebugLogger logs debug messages (only in debug mode)
	DebugLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info logs an informational message
func Info(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

// Debug logs a debug message (only in debug mode)
func Debug(format string, v ...interface{}) {
	if os.Getenv("GIN_MODE") == "debug" {
		DebugLogger.Printf(format, v...)
	}
}

// Fatal logs a fatal error and exits
func Fatal(format string, v ...interface{}) {
	ErrorLogger.Fatalf(format, v...)
}
