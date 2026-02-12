package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Template represents a cluster configuration template
type Template struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Labels      map[string]string `yaml:"labels,omitempty"`

	CPU     string `yaml:"cpu,omitempty"`     // e.g. "2"
	Memory  string `yaml:"memory,omitempty"`  // e.g. "4Gi"
	Storage string `yaml:"storage,omitempty"` // e.g. "20Gi"
	GPU     int    `yaml:"gpu,omitempty"`
	GPUType string `yaml:"gpuType,omitempty"` // e.g. "nvidia-t4"
	TTL     string `yaml:"ttl,omitempty"`     // e.g. "1h"
}

// Store interface for template management
type Store interface {
	List() ([]Template, error)
	Get(name string) (*Template, error)
}

// FileStore implements Store by reading templates from the filesystem
type FileStore struct {
	BaseDir string
}

// TemplatesFile represents a multi-template YAML file
type TemplatesFile struct {
	Templates []Template `yaml:"templates"`
}

// NewFileStore creates a new FileStore
func NewFileStore(baseDir string) *FileStore {
	return &FileStore{
		BaseDir: baseDir,
	}
}

// List returns all available templates
func (s *FileStore) List() ([]Template, error) {
	// Check if base directory exists
	if _, err := os.Stat(s.BaseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("templates directory not found at %s", s.BaseDir)
	}

	var templates []Template

	// Check for templates.yaml (multi-template file)
	templatesFile := filepath.Join(s.BaseDir, "templates.yaml")
	if _, err := os.Stat(templatesFile); err == nil {
		data, err := os.ReadFile(templatesFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read templates.yaml: %w", err)
		}

		var tf TemplatesFile
		if err := yaml.Unmarshal(data, &tf); err != nil {
			return nil, fmt.Errorf("failed to parse templates.yaml: %w", err)
		}

		templates = append(templates, tf.Templates...)
	}

	// Also scan for individual template files (*.yaml, excluding templates.yaml)
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") || entry.Name() == "templates.yaml" {
			continue
		}

		path := filepath.Join(s.BaseDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip files we can't read
		}

		var tmpl Template
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			continue // Skip malformed files
		}

		// If name is not set, derive from filename
		if tmpl.Name == "" {
			tmpl.Name = strings.TrimSuffix(entry.Name(), ".yaml")
		}

		templates = append(templates, tmpl)
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found in %s", s.BaseDir)
	}

	return templates, nil
}

// Get returns a specific template by name
func (s *FileStore) Get(name string) (*Template, error) {
	templates, err := s.List()
	if err != nil {
		return nil, err
	}

	for _, tmpl := range templates {
		if tmpl.Name == name {
			return &tmpl, nil
		}
	}

	return nil, fmt.Errorf("template %q not found; run 'ghostctl templates' to see available templates", name)
}

// GetTemplatesDir returns the default templates directory
func GetTemplatesDir() string {
	// Check if running from source (templates/ directory exists relative to binary)
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		
		// Try relative to executable (for development)
		templatesDir := filepath.Join(execDir, "..", "templates")
		if _, err := os.Stat(templatesDir); err == nil {
			if abs, err := filepath.Abs(templatesDir); err == nil {
				return abs
			}
		}
		
		// Try same directory as executable
		templatesDir = filepath.Join(execDir, "templates")
		if _, err := os.Stat(templatesDir); err == nil {
			return templatesDir
		}
	}

	// Try current working directory (for development)
	if cwd, err := os.Getwd(); err == nil {
		templatesDir := filepath.Join(cwd, "templates")
		if _, err := os.Stat(templatesDir); err == nil {
			return templatesDir
		}
	}

	// Default fallback locations
	// For installed binaries, could be /usr/local/share/ghostctl/templates
	possibleDirs := []string{
		"/usr/local/share/ghostctl/templates",
		"/opt/homebrew/share/ghostctl/templates",
		filepath.Join(os.Getenv("HOME"), ".ghost", "templates"),
	}

	for _, dir := range possibleDirs {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	// Return default path even if it doesn't exist (will error appropriately later)
	return filepath.Join(os.Getenv("HOME"), ".ghost", "templates")
}
