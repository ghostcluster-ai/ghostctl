package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/shell"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ghostctl for vCluster management",
	Long: `Initialize ghostctl for managing vClusters.

This command:
  - Validates connectivity to the Kubernetes host cluster
  - Checks that vcluster CLI is installed
  - Creates the ghostcluster namespace (if it doesn't exist)
  - Sets up local metadata store

Example:
  ghostctl init`,
	RunE: runInitCmd,
}

func runInitCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	namespace := "ghostcluster"

	logger.Info("Initializing ghostctl")

	// Check if vcluster CLI is available
	logger.Info("Checking for vcluster CLI")
	if !shell.CommandExists("vcluster") {
		logger.Error("vcluster CLI not found")
		return fmt.Errorf("vcluster CLI not found in PATH. Please install vCluster first:\n" +
			"  https://www.vcluster.com/docs/getting-started/setup")
	}
	fmt.Println("✓ vcluster CLI found")

	// Check kubectl connectivity
	logger.Info("Checking Kubernetes connectivity")
	result, err := shell.ExecuteCommand("kubectl", "cluster-info")
	if err != nil || result.ExitCode != 0 {
		logger.Error("Failed to connect to Kubernetes cluster")
		return fmt.Errorf("failed to connect to Kubernetes cluster. Make sure kubeconfig is set correctly")
	}
	fmt.Println("✓ Connected to Kubernetes cluster")

	// Create namespace if it doesn't exist
	logger.Info("Ensuring namespace exists", "namespace", namespace)
	result, _ = shell.ExecuteCommand("kubectl", "get", "namespace", namespace)
	if result.ExitCode != 0 {
		// Namespace doesn't exist, create it
		logger.Info("Creating namespace", "namespace", namespace)
		result, err = shell.ExecuteCommand("kubectl", "create", "namespace", namespace)
		if err != nil || result.ExitCode != 0 {
			logger.Error("Failed to create namespace")
			return fmt.Errorf("failed to create namespace %q", namespace)
		}
		fmt.Printf("✓ Created namespace: %s\n", namespace)
	} else {
		fmt.Printf("✓ Namespace already exists: %s\n", namespace)
	}

	// Initialize metadata store
	logger.Info("Initializing metadata store")
	metaStore, err := metadata.NewStore()
	if err != nil {
		logger.Error("Failed to initialize metadata store", "error", err)
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}
	fmt.Println("✓ Metadata store initialized")

	// List the directories
	clusters, _ := metaStore.List()
	fmt.Printf("\n✓ ghostctl initialized successfully\n")
	fmt.Printf("  Config directory: ~/.ghost\n")
	fmt.Printf("  Active clusters: %d\n", len(clusters))
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  ghostctl up my-cluster --ttl 1h  # Create a cluster\n")
	fmt.Printf("  ghostctl connect my-cluster      # Connect to cluster\n")
	fmt.Printf("  Kubeconfigs: ~/.ghost/kubeconfigs\n")
	fmt.Printf("  Metadata: ~/.ghost/clusters.json\n\n")
	fmt.Printf("Current clusters: %d\n", len(clusters))
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Create your first cluster: ghostctl up my-cluster")
	fmt.Println("  2. List active clusters: ghostctl list")
	fmt.Println("  3. Check cluster status: ghostctl status my-cluster")

	return nil
}
