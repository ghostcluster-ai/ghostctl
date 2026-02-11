package cmd

import (
	"fmt"
	"strings"

	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/shell"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <cluster-name>",
	Short: "Display vCluster status",
	Long: `Show status information about a virtual Kubernetes cluster.

This displays whether the vCluster is running and accessible, creation time,
and time-to-live information.

Examples:
  ghostctl status my-cluster        # Show cluster status
  ghostctl status my-cluster -v     # Show detailed error information`,
	Args: cobra.ExactArgs(1),
	RunE: runStatusCmd,
}

func runStatusCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]

	logger.Info("Fetching cluster status", "name", clusterName)

	// Initialize metadata store
	metaStore, err := metadata.NewStore()
	if err != nil {
		logger.Error("Failed to initialize metadata store", "error", err)
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}

	// Get cluster metadata
	meta, err := metaStore.Get(clusterName)
	if err != nil {
		logger.Error("Cluster not found", "name", clusterName)
		return fmt.Errorf("cluster %q not found in local registry", clusterName)
	}

	// Check if vCluster is accessible
	var status string
	isReachable := false

	// Check if pod is running
	if err := vcluster.Status(clusterName, meta.Namespace); err == nil {
		status = "running"
		
		// Now verify we can actually reach the API server
		kubeMgr, err := kubeconfig.NewManager()
		if err == nil {
			kubePath, err := kubeMgr.Get(clusterName, meta.Namespace)
			if err == nil {
				// Try a simple kubectl command to verify API access
				env := []string{"KUBECONFIG=" + kubePath}
				result, _ := shell.ExecuteCommandWithEnv(env, "kubectl", "get", "--raw", "/healthz")
				if result.ExitCode == 0 && strings.Contains(result.Stdout, "ok") {
					isReachable = true
				}
			}
		}
	} else {
		status = "offline"
	}

	// Display status
	displayStatus(clusterName, meta, status, isReachable)

	return nil
}

func displayStatus(name string, meta *metadata.ClusterMetadata, status string, isReachable bool) {
	fmt.Printf("Cluster: %s\n", name)
	fmt.Printf("Namespace: %s\n", meta.Namespace)
	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Created: %s\n", meta.CreatedAt.Format("2006-01-02 15:04:05"))

	// Display TTL information if set
	if meta.TTL != "" {
		fmt.Printf("TTL: %s\n", meta.TTL)
		// Note: Computing actual TTL remaining would require parsing the TTL string
		// For now, just display that it's set
	}

	// Display kubeconfig path
	kubePath, _ := metadata.GetClusterPath(name)
	fmt.Printf("Kubeconfig: %s\n", kubePath)

	// Display connectivity status
	if isReachable {
		fmt.Printf("\n✓ vCluster is accessible\n")
	} else {
		fmt.Printf("\n✗ vCluster is not accessible\n")
	}

	fmt.Printf("\nTo connect, run:\n")
	fmt.Printf("  ghostctl connect %s\n", name)
}
