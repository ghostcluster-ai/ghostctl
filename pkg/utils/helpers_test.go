package utils

import (
	"testing"
)

// TestParseDuration tests the ParseDuration function
func TestParseDuration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantValid bool
	}{
		{"valid 1h", "1h", false, true},
		{"valid 30m", "30m", false, true},
		{"valid 10s", "10s", false, true},
		{"invalid format", "invalid", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == 0 {
				t.Errorf("ParseDuration() got 0, expected non-zero duration")
			}
		})
	}
}

// TestFormatBytes tests the FormatBytes function
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"zero bytes", 0, "0 B"},
		{"512 bytes", 512, "512 B"},
		{"1 KB", 1024, "1.00 KB"},
		{"1 MB", 1024 * 1024, "1.00 MB"},
		{"1 GB", 1024 * 1024 * 1024, "1.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBytes(tt.size)
			if got != tt.want {
				t.Errorf("FormatBytes(%d) = %s, want %s", tt.size, got, tt.want)
			}
		})
	}
}

// TestValidateClusterName tests cluster name validation
func TestValidateClusterName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-cluster", false},
		{"valid name with numbers", "cluster-123", false},
		{"too long", string(make([]byte, 64)), true},
		{"empty name", "", true},
		{"uppercase not allowed", "MyCluster", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClusterName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClusterName() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
