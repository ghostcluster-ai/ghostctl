package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down <cluster-name>",
	Short: "Destroy an ephemeral cluster",
	Long: `Destroy and remove an ephemeral vCluster.

This command will:
  - Stop all running workloads in the cluster
  - Remove persistent volumes and data
  - Delete the cluster from the host
  - Clean up networking and RBAC resources

Examples:
  ghostctl down my-cluster                 # Destroy cluster
  ghostctl down my-cluster --force         # Force destroy without confirmation
  ghostctl down my-cluster --drain-timeout 2m # Wait 2 minutes for graceful termination`,
	Args: cobra.ExactArgs(1),
	RunE: runDownCmd,
}

var (
	force         bool
	drainTimeout  string
	deleteStorage bool
)

func init() {
	downCmd.Flags().BoolVar(
		&force, "force", false,
		"force destroy without confirmation",
	)
	downCmd.Flags().StringVar(
		&drainTimeout, "drain-timeout", "1m",
		"timeout for graceful pod termination",
	)
	downCmd.Flags().BoolVar(
		&deleteStorage, "delete-storage", true,
		"delete persistent volumes with the cluster",
	)
}

func runDownCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]

	logger.Info("Destroying cluster", "name", clusterName)

	if !force {
		fmt.Printf("Are you sure you want to destroy cluster '%s'? (y/n): ", clusterName)
		var response string
		_, _ = fmt.Scanln(&response) // nolint:errcheck
		if response != "y" && response != "yes" {
			logger.Info("Cluster destruction cancelled")
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Initialize cluster manager
	cm := cluster.NewClusterManager()

	// Delete cluster
	logger.Info("Deleting cluster resources", "name", clusterName)
	deleteOpts := &cluster.DeleteOptions{
		DrainTimeout:  drainTimeout,
		DeleteStorage: deleteStorage,
	}

	if err := cm.DeleteCluster(clusterName, deleteOpts); err != nil {
		logger.Error("Failed to delete cluster", "error", err)
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	logger.Info("✓ Cluster destroyed successfully", "name", clusterName)
	fmt.Printf("✓ Cluster '%s' has been destroyed\n", clusterName)

	return nil
}
