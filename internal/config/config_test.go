package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDefaultConfig tests the default configuration
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.APIServer == "" {
		t.Error("DefaultConfig() APIServer is empty")
	}

	if cfg.Namespace == "" {
		t.Error("DefaultConfig() Namespace is empty")
	}
}

// TestValidate tests configuration validation
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			"valid config",
			&Config{APIServer: "localhost:8080", Namespace: "default"},
			false,
		},
		{
			"missing apiserver",
			&Config{Namespace: "default"},
			true,
		},
		{
			"missing namespace",
			&Config{APIServer: "localhost:8080"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetConfigPath tests the config path resolution
func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() err = %v", err)
	}

	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	if !filepath.IsAbs(path) {
		t.Errorf("GetConfigPath() returned non-absolute path: %s", path)
	}

	// Verify it contains the expected directory name
	expectedDir := filepath.Join(os.Getenv("HOME"), ConfigDirName)
	if !strings.Contains(path, expectedDir) {
		t.Errorf("GetConfigPath() path doesn't contain expected directory: %s in %s", expectedDir, path)
	}
}
