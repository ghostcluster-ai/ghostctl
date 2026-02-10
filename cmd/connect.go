package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <cluster-name>",
	Short: "Show connection command for a vCluster",
	Long: `Display the command to connect to a vCluster using kubectl.

This command prints an export statement that you can run in your shell
to set KUBECONFIG to point to the virtual cluster.

Examples:
  ghostctl connect my-cluster              # Show connection command
  eval $(ghostctl connect my-cluster)      # Connect in current shell`,
	Args: cobra.ExactArgs(1),
	RunE: runConnectCmd,
}

func runConnectCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]

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

	logger.Info("Getting kubeconfig for cluster", "name", clusterName)

	// Ensure kubeconfig exists (regenerate if necessary)
	kubeMgr, err := kubeconfig.NewManager()
	if err != nil {
		logger.Error("Failed to create kubeconfig manager", "error", err)
		return fmt.Errorf("failed to create kubeconfig manager: %w", err)
	}

	kubePath, err := kubeMgr.Get(clusterName, meta.Namespace)
	if err != nil {
		logger.Error("Failed to get kubeconfig", "error", err)
		return fmt.Errorf("failed to get kubeconfig for cluster %q: %w", clusterName, err)
	}

	// Print the export command
	fmt.Printf("export KUBECONFIG=%s\n", kubePath)

	return nil
}
