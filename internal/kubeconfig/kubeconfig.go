package kubeconfig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// MergeIntoDefault merges a vCluster kubeconfig into the default ~/.kube/config
// and returns the context name that was added
func (km *Manager) MergeIntoDefault(clusterName, namespace string) (string, error) {
	// Get vCluster kubeconfig content
	vclusterConfig, err := vcluster.GetKubeconfig(clusterName, namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get vCluster kubeconfig: %w", err)
	}

	// Write to temp file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	tmpFile := filepath.Join(homeDir, ".ghost", fmt.Sprintf(".tmp_%s.yaml", clusterName))
	if err := os.WriteFile(tmpFile, []byte(vclusterConfig), 0600); err != nil {
		return "", fmt.Errorf("failed to write temp kubeconfig: %w", err)
	}
	defer os.Remove(tmpFile)

	// Use kubectl to merge
	kubeconfigPath := getDefaultKubeconfigPath()
	contextName := fmt.Sprintf("vcluster_%s", clusterName)
	
	// Set KUBECONFIG to merge
	cmd := exec.Command("kubectl", "config", "view", "--flatten", "--merge")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s:%s", kubeconfigPath, tmpFile))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to merge kubeconfig: %w - %s", err, string(output))
	}

	// Write merged config back
	if err := os.WriteFile(kubeconfigPath, output, 0600); err != nil {
		return "", fmt.Errorf("failed to write merged kubeconfig: %w", err)
	}

	// Rename the context to a predictable name (non-fatal if it fails)
	_ = renameContext(tmpFile, contextName)

	return contextName, nil
}

// SwitchContext switches the current context in ~/.kube/config
func (km *Manager) SwitchContext(contextName string) error {
	cmd := exec.Command("kubectl", "config", "use-context", contextName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to switch context: %w - %s", err, string(output))
	}
	return nil
}

// GetCurrentContext returns the current context name
func (km *Manager) GetCurrentContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get current context: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// SaveCurrentContext saves the current context for later restoration
func (km *Manager) SaveCurrentContext() error {
	currentContext, err := km.GetCurrentContext()
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	contextFile := filepath.Join(homeDir, ".ghost", ".saved_context")
	return os.WriteFile(contextFile, []byte(currentContext), 0600)
}

// RestoreSavedContext restores the previously saved context
func (km *Manager) RestoreSavedContext() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	contextFile := filepath.Join(homeDir, ".ghost", ".saved_context")
	data, err := os.ReadFile(contextFile)
	if err != nil {
		return "", fmt.Errorf("no saved context found")
	}

	savedContext := strings.TrimSpace(string(data))
	if err := km.SwitchContext(savedContext); err != nil {
		return "", err
	}

	// Clean up saved context file
	os.Remove(contextFile)

	return savedContext, nil
}

// RemoveContext removes a context from ~/.kube/config
func (km *Manager) RemoveContext(contextName string) error {
	cmd := exec.Command("kubectl", "config", "delete-context", contextName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove context: %w - %s", err, string(output))
	}
	return nil
}

func getDefaultKubeconfigPath() string {
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		return strings.Split(kc, ":")[0]
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".kube", "config")
}

func renameContext(kubeconfigFile, newName string) error {
	// This is optional - extracts original context name and renames it
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name", "--kubeconfig", kubeconfigFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	
	oldName := strings.TrimSpace(string(output))
	if oldName == "" {
		return fmt.Errorf("no context found")
	}

	cmd = exec.Command("kubectl", "config", "rename-context", oldName, newName)
	_, err = cmd.CombinedOutput()
	return err
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
