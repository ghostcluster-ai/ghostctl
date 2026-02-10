package vcluster

import (
	"fmt"
	"strings"
	"time"

	"github.com/ghostcluster-ai/ghostctl/internal/shell"
)

const (
	DefaultNamespace = "ghostcluster"
)

// VCluster represents a vCluster instance
type VCluster struct {
	Name      string
	Namespace string
}

// Create creates a new vCluster using the vcluster CLI
func Create(name, namespace string) error {
	if !shell.CommandExists("vcluster") {
		return fmt.Errorf("vcluster CLI not found in PATH. Please install vCluster: https://www.vcluster.com/docs/getting-started/setup")
	}

	args := []string{
		"create", name,
		"-n", namespace,
		"--connect=false",
		"--update-current=false",
	}

	result, err := shell.ExecuteCommand("vcluster", args...)
	if err != nil {
		return fmt.Errorf("failed to create vCluster: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("vCluster creation failed (exit code %d): %s", result.ExitCode, result.Stdout)
	}

	return nil
}

// Delete deletes a vCluster using the vcluster CLI
func Delete(name, namespace string) error {
	if !shell.CommandExists("vcluster") {
		return fmt.Errorf("vcluster CLI not found in PATH")
	}

	args := []string{
		"delete", name,
		"-n", namespace,
	}

	result, err := shell.ExecuteCommand("vcluster", args...)
	if err != nil {
		return fmt.Errorf("failed to delete vCluster: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("vCluster deletion failed (exit code %d): %s", result.ExitCode, result.Stdout)
	}

	return nil
}

// Status gets the status of a vCluster
func Status(name, namespace string) error {
	if !shell.CommandExists("vcluster") {
		return fmt.Errorf("vcluster CLI not found in PATH")
	}

	args := []string{
		"status", name,
		"-n", namespace,
	}

	result, err := shell.ExecuteCommand("vcluster", args...)
	if err != nil {
		return fmt.Errorf("failed to get vCluster status: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("vCluster not ready or not found (exit code %d)", result.ExitCode)
	}

	return nil
}

// GetKubeconfig retrieves the kubeconfig for a vCluster
func GetKubeconfig(name, namespace string) (string, error) {
	if !shell.CommandExists("vcluster") {
		return "", fmt.Errorf("vcluster CLI not found in PATH")
	}

	args := []string{
		"connect", name,
		"-n", namespace,
		"--update-current=false",
		"--print",
	}

	result, err := shell.ExecuteCommand("vcluster", args...)
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("failed to get kubeconfig (exit code %d): %s", result.ExitCode, result.Stdout)
	}

	return result.Stdout, nil
}

// IsReady waits for a vCluster to be ready with polling
func IsReady(name, namespace string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for vCluster %s to be ready", name)
		}

		err := Status(name, namespace)
		if err == nil {
			return nil
		}

		select {
		case <-ticker.C:
			// Continue polling
		}
	}
}

// List lists all vClusters in a namespace
func List(namespace string) ([]string, error) {
	if !shell.CommandExists("vcluster") {
		return nil, fmt.Errorf("vcluster CLI not found in PATH")
	}

	args := []string{
		"list",
		"-n", namespace,
	}

	result, err := shell.ExecuteCommand("vcluster", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list vClusters: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("failed to list vClusters (exit code %d): %s", result.ExitCode, result.Stdout)
	}

	// Parse output - skip header and extract cluster names
	var clusters []string
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "NAME") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			clusters = append(clusters, fields[0])
		}
	}

	return clusters, nil
}
