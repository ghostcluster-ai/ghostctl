package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStoreList(t *testing.T) {
	// Create temporary directory with test templates
	tmpDir := t.TempDir()

	// Create individual template file
	defaultTemplate := `name: default
description: Balanced resources for general workloads
cpu: "2"
memory: 4Gi
storage: 20Gi
ttl: 1h
`
	if err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultTemplate), 0644); err != nil {
		t.Fatal(err)
	}

	// Create multi-template file
	multiTemplate := `templates:
  - name: gpu
    description: GPU-accelerated workload
    cpu: "4"
    memory: 16Gi
    gpu: 1
    gpuType: nvidia-t4
    ttl: 2h
  - name: minimal
    description: Minimal resources
    cpu: "1"
    memory: 2Gi
`
	if err := os.WriteFile(filepath.Join(tmpDir, "templates.yaml"), []byte(multiTemplate), 0644); err != nil {
		t.Fatal(err)
	}

	// Test List
	store := NewFileStore(tmpDir)
	templates, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(templates) != 3 {
		t.Fatalf("Expected 3 templates, got %d", len(templates))
	}

	// Test Get
	defaultTmpl, err := store.Get("default")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if defaultTmpl.CPU != "2" {
		t.Errorf("Expected CPU=2, got %s", defaultTmpl.CPU)
	}

	gpuTmpl, err := store.Get("gpu")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if gpuTmpl.GPU != 1 {
		t.Errorf("Expected GPU=1, got %d", gpuTmpl.GPU)
	}

	// Test non-existent template
	_, err = store.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}
}

func TestFileStoreEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	store := NewFileStore(tmpDir)
	_, err := store.List()
	if err == nil {
		t.Error("Expected error for empty templates directory")
	}
}

func TestFileStoreNonExistentDir(t *testing.T) {
	store := NewFileStore("/nonexistent/path")
	_, err := store.List()
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
