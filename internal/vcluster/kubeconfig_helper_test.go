package vcluster

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKubeconfigPath(t *testing.T) {
	baseDir := t.TempDir()
	mgr, err := NewKubeconfigManager(baseDir, "ghostcluster")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ref := ClusterRef{Name: "pr-123"}
	got := mgr.KubeconfigPath(ref)
	want := filepath.Join(baseDir, "kubeconfigs", "pr-123.yaml")
	if got != want {
		t.Fatalf("expected kubeconfig path %q, got %q", want, got)
	}
}

func TestGetOrCreateKubeconfigUsesExistingFile(t *testing.T) {
	baseDir := t.TempDir()
	mgr, err := NewKubeconfigManager(baseDir, "ghostcluster")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ref := ClusterRef{Name: "pr-456"}
	path := filepath.Join(baseDir, "kubeconfigs", "pr-456.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	content := []byte("apiVersion: v1\nkind: Config\n")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("failed to write kubeconfig: %v", err)
	}

	got, err := mgr.GetOrCreateKubeconfig(ref)
	if err != nil {
		t.Fatalf("GetOrCreateKubeconfig error: %v", err)
	}
	if got != path {
		t.Fatalf("expected path %q, got %q", path, got)
	}
}

func TestGetOrCreateKubeconfigRequiresName(t *testing.T) {
	baseDir := t.TempDir()
	mgr, err := NewKubeconfigManager(baseDir, "ghostcluster")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = mgr.GetOrCreateKubeconfig(ClusterRef{Name: ""})
	if err == nil {
		t.Fatalf("expected error for empty cluster name")
	}
}
