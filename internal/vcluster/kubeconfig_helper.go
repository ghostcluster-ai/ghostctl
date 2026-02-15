package vcluster

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
)

// ClusterRef identifies a vCluster and its namespace.
type ClusterRef struct {
	Name      string
	Namespace string
}

// KubeconfigManager provides access to vCluster kubeconfig files.
type KubeconfigManager interface {
	GetOrCreateKubeconfig(ref ClusterRef) (string, error)
	KubeconfigPath(ref ClusterRef) string
}

type vclusterHelper struct {
	BaseDir   string
	Namespace string
}

// NewKubeconfigManager creates a kubeconfig manager rooted at baseDir.
// If baseDir is empty, ~/.ghost is used. If namespace is empty, the default is used.
func NewKubeconfigManager(baseDir, namespace string) (KubeconfigManager, error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, metadata.DefaultDir)
	}

	if namespace == "" {
		namespace = DefaultNamespace
	}

	return &vclusterHelper{
		BaseDir:   baseDir,
		Namespace: namespace,
	}, nil
}

func (h *vclusterHelper) KubeconfigPath(ref ClusterRef) string {
	return filepath.Join(h.BaseDir, metadata.KubeconfigsDirName, ref.Name+".yaml")
}

func (h *vclusterHelper) GetOrCreateKubeconfig(ref ClusterRef) (string, error) {
	if strings.TrimSpace(ref.Name) == "" {
		return "", fmt.Errorf("cluster name is required")
	}

	path := h.KubeconfigPath(ref)
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return path, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", fmt.Errorf("failed to create kubeconfig directory: %w", err)
	}

	namespace := h.namespaceFor(ref)
	kubeconfig, err := GetKubeconfig(ref.Name, namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig for vCluster %q in namespace %q: %w", ref.Name, namespace, err)
	}

	if err := os.WriteFile(path, []byte(kubeconfig), 0600); err != nil {
		return "", fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	return path, nil
}

func (h *vclusterHelper) namespaceFor(ref ClusterRef) string {
	if ref.Namespace != "" {
		return ref.Namespace
	}
	if h.Namespace != "" {
		return h.Namespace
	}
	return DefaultNamespace
}
