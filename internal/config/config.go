package config

import (
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

const (
	ConfigFileName = "config.yaml"
	ConfigDirName  = ".ghost"
)

// Config represents the ghostctl configuration structure
type Config struct {
	APIServer       string            `yaml:"apiServer"`
	AuthToken       string            `yaml:"authToken"`
	DefaultTemplate string            `yaml:"defaultTemplate"`
	DefaultTTL      string            `yaml:"defaultTTL"`
	Namespace       string            `yaml:"namespace"`
	LogLevel        string            `yaml:"logLevel"`
	CloudProvider   string            `yaml:"cloudProvider"`
	ProjectID       string            `yaml:"projectID"`
	Metadata        map[string]string `yaml:"metadata"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ConfigDirName)
	configPath := filepath.Join(configDir, ConfigFileName)
	return configPath, nil
}

// Load loads the configuration from file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		APIServer:       "localhost:8080",
		DefaultTemplate: "default",
		DefaultTTL:      "1h",
		Namespace:       "ghostcluster",
		LogLevel:        "info",
		CloudProvider:   "local",
		Metadata:        make(map[string]string),
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.APIServer == "" {
		return fmt.Errorf("apiServer is required")
	}

	if c.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	return nil
}
