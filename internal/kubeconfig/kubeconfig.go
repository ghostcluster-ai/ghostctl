package kubeconfig

import (
	"fmt"
	"os"
	"time"

	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
)

// Manager manages kubeconfig files for vClusters
type Manager struct {
	metaStore *metadata.Store
}

// NewManager creates a new kubeconfig manager
func NewManager() (*Manager, error) {
	store, err := metadata.NewStore()
	if err != nil {
		return nil, err
	}

	return &Manager{
		metaStore: store,
	}, nil
}

// EnsureExists ensures a kubeconfig file exists for a cluster
// If it doesn't exist, it retrieves it from the vCluster
func (km *Manager) EnsureExists(clusterName, namespace string) (string, error) {
	// Get path where kubeconfig should be stored
	path, err := metadata.GetClusterPath(clusterName)
	if err != nil {
		return "", err
	}

	// Check if file already exists and is recent (less than 1 hour old)
	if fileInfo, err := os.Stat(path); err == nil {
		if time.Since(fileInfo.ModTime()) < time.Hour {
			return path, nil
		}
	}

	// Retrieve kubeconfig from vCluster
	kubeconfig, err := vcluster.GetKubeconfig(clusterName, namespace)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve kubeconfig: %w", err)
	}

	// Write kubeconfig to file
	if err := os.WriteFile(path, []byte(kubeconfig), 0600); err != nil {
		return "", fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	return path, nil
}

// Get gets the kubeconfig path for a cluster
// It ensures the kubeconfig file exists, regenerating if necessary
func (km *Manager) Get(clusterName, namespace string) (string, error) {
	return km.EnsureExists(clusterName, namespace)
}

// Delete deletes the kubeconfig file for a cluster
func (km *Manager) Delete(clusterName string) error {
	path, err := metadata.GetClusterPath(clusterName)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete kubeconfig: %w", err)
	}

	return nil
}

// Exists checks if a kubeconfig file exists
func (km *Manager) Exists(clusterName string) bool {
	path, err := metadata.GetClusterPath(clusterName)
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// GetPath returns the path to a cluster's kubeconfig without checking existence
func (km *Manager) GetPath(clusterName string) (string, error) {
	return metadata.GetClusterPath(clusterName)
}

// Fresh regenerates a kubeconfig from the vCluster
func (km *Manager) Fresh(clusterName, namespace string) (string, error) {
	kubeconfig, err := vcluster.GetKubeconfig(clusterName, namespace)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve kubeconfig: %w", err)
	}

	path, err := metadata.GetClusterPath(clusterName)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, []byte(kubeconfig), 0600); err != nil {
		return "", fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	return path, nil
}
