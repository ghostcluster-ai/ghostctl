package telemetry

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// LogLevel defines logging levels
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger handles logging for the application
type Logger struct {
	level  LogLevel
	mu     sync.Mutex
	stdout *log.Logger
	stderr *log.Logger
}

var (
	instance *Logger
	once     sync.Once
)

// InitLogger initializes the global logger
func InitLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			level:  InfoLevel,
			stdout: log.New(os.Stdout, "", log.LstdFlags),
			stderr: log.New(os.Stderr, "", log.LstdFlags),
		}

		// Read log level from environment
		if levelStr := os.Getenv("GHOSTCTL_LOG_LEVEL"); levelStr != "" {
			instance.SetLevelString(levelStr)
		}
	})

	return instance
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if instance == nil {
		return InitLogger()
	}
	return instance
}

// SetLogLevel sets the logging level
func SetLogLevel(level string) {
	if instance != nil {
		instance.SetLevelString(level)
	}
}

// SetLevelString sets the logging level from string
func (l *Logger) SetLevelString(levelStr string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch levelStr {
	case "debug":
		l.level = DebugLevel
	case "info":
		l.level = InfoLevel
	case "warn":
		l.level = WarnLevel
	case "error":
		l.level = ErrorLevel
	}
}

// Debug logs debug-level messages
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.level <= DebugLevel {
		l.printf("[DEBUG] %s %s", msg, formatFields(fields))
	}
}

// Info logs info-level messages
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.level <= InfoLevel {
		l.printf("[INFO] %s %s", msg, formatFields(fields))
	}
}

// Warn logs warning-level messages
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.level <= WarnLevel {
		l.printf("[WARN] %s %s", msg, formatFields(fields))
	}
}

// Error logs error-level messages
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.level <= ErrorLevel {
		l.eprintf("[ERROR] %s %s", msg, formatFields(fields))
	}
}

func (l *Logger) printf(format string, args ...interface{}) {
	l.stdout.Printf(format+"\n", args...)
}

func (l *Logger) eprintf(format string, args ...interface{}) {
	l.stderr.Printf(format+"\n", args...)
}

// formatFields formats key-value pairs for logging
func formatFields(fields []interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	result := ""
	for i := 0; i < len(fields)-1; i += 2 {
		key := fields[i]
		val := fields[i+1]
		result += fmt.Sprintf("%s=%v ", key, val)
	}

	return result
}

// Metrics represents application metrics
type Metrics struct {
	CommandsExecuted int64
	ErrorCount       int64
	WarningCount     int64
	RequestCount     int64
}

var metrics = &Metrics{}
var metricsMu sync.RWMutex

// RecordCommand records a command execution
func RecordCommand() {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	metrics.CommandsExecuted++
}

// RecordError records an error
func RecordError() {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	metrics.ErrorCount++
}

// RecordWarning records a warning
func RecordWarning() {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	metrics.WarningCount++
}

// RecordRequest records an API request
func RecordRequest() {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	metrics.RequestCount++
}

// GetMetrics returns current metrics
func GetMetrics() Metrics {
	metricsMu.RLock()
	defer metricsMu.RUnlock()
	return *metrics
}

// ResetMetrics resets all metrics
func ResetMetrics() {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	metrics = &Metrics{}
}
