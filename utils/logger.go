package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	// Logger is the global logger instance
	Logger *log.Logger
	// LogFile is the file handle for the log file
	LogFile *os.File
	// LogLevel controls the verbosity of logging
	LogLevel = LevelInfo
)

// Log levels
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// InitLogger initializes the logger with the specified file path
func InitLogger(logFilePath string) error {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logFilePath)
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file (create if not exists, append if exists)
	LogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Initialize logger
	Logger = log.New(LogFile, "", log.Ldate|log.Ltime)
	return nil
}

// Close closes the log file
func Close() {
	if LogFile != nil {
		LogFile.Close()
	}
}

// SetLogLevel sets the log level
func SetLogLevel(level int) {
	LogLevel = level
}

// logWithLevel logs a message with the specified level and additional caller info
func logWithLevel(level int, levelStr string, format string, v ...interface{}) {
	if level < LogLevel {
		return
	}

	// Get caller info
	_, file, line, _ := runtime.Caller(2) // Skip this function and the calling log function

	// Format the log message with timestamp, level, file, line, and message
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, v...)

	logMessage := fmt.Sprintf("[%s] [%s] [%s:%d] %s",
		timestamp,
		levelStr,
		filepath.Base(file), // Only use the filename, not the full path
		line,
		message,
	)

	// Log to the file
	if Logger != nil {
		Logger.Println(logMessage)
	}

	// Also print to stdout for immediate feedback
	fmt.Println(logMessage)
}

// Debug logs a debug message with caller information
func Debug(format string, v ...interface{}) {
	logWithLevel(LevelDebug, "DEBUG", format, v...)
}

// Info logs an info message with caller information
func Info(format string, v ...interface{}) {
	logWithLevel(LevelInfo, "INFO", format, v...)
}

// Warn logs a warning message with caller information
func Warn(format string, v ...interface{}) {
	logWithLevel(LevelWarn, "WARN", format, v...)
}

// Error logs an error message with caller information
func Error(format string, v ...interface{}) {
	logWithLevel(LevelError, "ERROR", format, v...)
}

// Fatal logs a fatal message with caller information and exits
func Fatal(format string, v ...interface{}) {
	logWithLevel(LevelFatal, "FATAL", format, v...)
	os.Exit(1)
}

// Trace is a helper function to trace function entry and exit
// Usage: defer Trace()()
func Trace() func() {
	_, file, line, _ := runtime.Caller(1)
	fnName := filepath.Base(file)

	// Log entry
	logWithLevel(LevelDebug, "TRACE", "Entering %s:%d", fnName, line)

	// Return a function to log exit
	return func() {
		_, file, line, _ := runtime.Caller(1)
		fnName := filepath.Base(file)
		logWithLevel(LevelDebug, "TRACE", "Exiting %s:%d", fnName, line)
	}
}

// TraceSQL logs SQL operations
func TraceSQL(query string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fnName := filepath.Base(file)

	// Log the SQL query and arguments
	logWithLevel(LevelDebug, "SQL", "[%s:%d] Query: %s, Args: %v", fnName, line, query, args)
}
