package cmd

import (
	"fmt"
	"time"

	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [cluster-name]",
	Short: "Create a new ephemeral vCluster",
	Long: `Create a new virtual Kubernetes cluster in the host Kubernetes cluster.

The vCluster will be created in the ghostcluster namespace and can be accessed
using the generated kubeconfig.

Examples:
  ghostctl up my-cluster              # Create cluster with auto-generated name
  ghostctl up my-cluster --ttl 2h     # Create with 2 hour TTL`,
	RunE: runUpCmd,
}

var (
	ttl string
)

func init() {
	upCmd.Flags().StringVar(
		&ttl, "ttl", "",
		"time-to-live for the cluster (optional, examples: 30m, 2h, 1d)",
	)
}

func runUpCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	// Get cluster name from args
	if len(args) == 0 {
		return fmt.Errorf("cluster name is required")
	}
	clusterName := args[0]

	logger.Info("Creating new vCluster",
		"name", clusterName,
		"ttl", ttl,
	)

	// Initialize metadata store
	metaStore, err := metadata.NewStore()
	if err != nil {
		logger.Error("Failed to initialize metadata store", "error", err)
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}

	// Check if cluster already exists
	if metaStore.Exists(clusterName) {
		return fmt.Errorf("cluster %q already exists", clusterName)
	}

	namespace := vcluster.DefaultNamespace

	// Create the vCluster
	logger.Info("Creating vCluster in Kubernetes")
	if err := vcluster.Create(clusterName, namespace); err != nil {
		logger.Error("Failed to create vCluster", "error", err)
		return err
	}

	// Wait for vCluster to be ready
	logger.Info("Waiting for vCluster to be ready")
	waitTimeout := 5 * time.Minute
	if err := vcluster.IsReady(clusterName, namespace, waitTimeout); err != nil {
		logger.Error("vCluster failed to become ready", "error", err)
		return err
	}

	// Retrieve and store kubeconfig
	logger.Info("Retrieving kubeconfig")
	kubeconfigPath, err := kubeconfig.NewManager()
	if err != nil {
		logger.Error("Failed to create kubeconfig manager", "error", err)
		return err
	}

	_, err = kubeconfigPath.Fresh(clusterName, namespace)
	if err != nil {
		logger.Error("Failed to retrieve kubeconfig", "error", err)
		return err
	}

	// Store metadata
	kubePath, _ := metadata.GetClusterPath(clusterName)
	meta := &metadata.ClusterMetadata{
		Name:           clusterName,
		Namespace:      namespace,
		CreatedAt:      time.Now(),
		TTL:            ttl,
		KubeconfigPath: kubePath,
		HostCluster:    "current",
	}

	if err := metaStore.Add(meta); err != nil {
		logger.Error("Failed to store cluster metadata", "error", err)
		return fmt.Errorf("failed to store cluster metadata: %w", err)
	}

	logger.Info("✓ vCluster created successfully", "name", clusterName)
	fmt.Printf("\n✓ Cluster '%s' is ready!\n", clusterName)
	fmt.Println("\nUseful commands:")
	fmt.Printf("  ghostctl status %s               # Check cluster status\n", clusterName)
	fmt.Printf("  ghostctl connect %s              # Show connection command\n", clusterName)
	fmt.Printf("  ghostctl exec %s -- kubectl ... # Run command in cluster\n", clusterName)
	fmt.Printf("  ghostctl down %s                 # Destroy cluster\n", clusterName)

	return nil
}
