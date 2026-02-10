package telemetry

import (
	"testing"
)

// TestInitLogger tests logger initialization
func TestInitLogger(t *testing.T) {
	logger := InitLogger()

	if logger.stdout == nil {
		t.Error("InitLogger() stdout is nil")
	}

	if logger.stderr == nil {
		t.Error("InitLogger() stderr is nil")
	}
}

// TestGetLogger tests retrieving global logger
func TestGetLogger(t *testing.T) {
	logger := GetLogger()

	if logger == nil {
		t.Error("GetLogger() returned nil")
	}
}

// TestSetLogLevel tests setting log level
func TestSetLogLevel(t *testing.T) {
	// Just verify it doesn't panic
	SetLogLevel("debug")
	SetLogLevel("info")
	SetLogLevel("warn")
	SetLogLevel("error")
}

// TestGetMetrics tests metrics retrieval
func TestGetMetrics(t *testing.T) {
	metrics := GetMetrics()

	// Verify metrics structure
	if metrics.CommandsExecuted < 0 {
		t.Error("GetMetrics() CommandsExecuted is negative")
	}

	if metrics.ErrorCount < 0 {
		t.Error("GetMetrics() ErrorCount is negative")
	}
}

// TestResetMetrics tests metrics reset
func TestResetMetrics(t *testing.T) {
	// Record some events
	RecordCommand()
	RecordError()

	// Reset
	ResetMetrics()

	// Verify metrics are reset
	metrics := GetMetrics()
	if metrics.CommandsExecuted != 0 {
		t.Errorf("ResetMetrics() CommandsExecuted = %d, want 0", metrics.CommandsExecuted)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("ResetMetrics() ErrorCount = %d, want 0", metrics.ErrorCount)
	}
}
