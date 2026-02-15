package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/ghostcluster-ai/ghostctl/internal/config"
	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down <cluster-name>",
	Short: "Destroy an ephemeral vCluster",
	Long: `Destroy and remove a virtual Kubernetes cluster.

This command will delete the vCluster from the Kubernetes host cluster
and clean up local metadata and kubeconfig files.

Examples:
  ghostctl down my-cluster              # Destroy cluster with confirmation
  ghostctl down my-cluster --force      # Force destroy without confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: runDownCmd,
}

var (
	force bool
)

func init() {
	downCmd.Flags().BoolVar(
		&force, "force", false,
		"force destroy without confirmation",
	)
}

func runDownCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return fmt.Errorf("failed to load config: %w", err)
	}

	namespace := cfg.Namespace
	if namespace == "" {
		namespace = vcluster.DefaultNamespace
	}

	// Initialize metadata store
	var meta *metadata.ClusterMetadata
	metaStore, err := metadata.NewStore()
	if err != nil {
		logger.Warn("Failed to initialize metadata store", "error", err)
	}

	// Check if cluster exists in metadata
	if metaStore != nil {
		meta, err = metaStore.Get(clusterName)
		if err != nil {
			logger.Info("Local metadata for cluster not found; proceeding with live deletion", "name", clusterName)
		} else if meta.Namespace != "" {
			namespace = meta.Namespace
		}
	}

	logger.Info("Destroying vCluster", "name", clusterName, "namespace", namespace)

	// Confirm deletion
	if !force {
		fmt.Printf("Are you sure you want to destroy cluster '%s'? This cannot be undone. (y/n): ", clusterName)
		reader := bufio.NewReader(cmd.InOrStdin())
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			logger.Info("Cluster destruction cancelled")
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the vCluster
	logger.Info("Deleting vCluster from Kubernetes")
	if err := vcluster.Delete(clusterName, namespace); err != nil {
		logger.Error("Failed to delete vCluster", "error", err)
		return fmt.Errorf("failed to delete vCluster: %w", err)
	}

	// Delete kubeconfig file
	logger.Info("Cleaning up kubeconfig")
	kubeMgr, err := kubeconfig.NewManager()
	if err != nil {
		logger.Warn("Failed to create kubeconfig manager, skipping cleanup", "error", err)
	} else {
		_ = kubeMgr.Delete(clusterName)
	}

	// Remove from metadata store
	if metaStore != nil {
		if err := metaStore.Remove(clusterName); err != nil {
			logger.Error("Failed to remove cluster metadata", "error", err)
			// Don't fail here, cluster was deleted from k8s
		}
	}

	logger.Info("✓ vCluster destroyed successfully", "name", clusterName)
	fmt.Printf("✓ Cluster '%s' has been destroyed\n", clusterName)

	return nil
}
