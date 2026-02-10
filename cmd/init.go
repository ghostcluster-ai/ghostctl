package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Ghostcluster controller in the host cluster",
	Long: `Initialize the Ghostcluster controller in your Kubernetes host cluster.

This command:
  - Validates connectivity to the host cluster
  - Creates necessary namespaces
  - Installs vCluster controller components
  - Sets up RBAC for cluster management

Example:
  ghostctl init --host-cluster my-cluster --namespace ghostcluster`,
	RunE: runInitCmd,
}

var (
	hostCluster    string
	namespace      string
	gcpProject     string
	awsRegion      string
	skipValidation bool
)

func init() {
	initCmd.Flags().StringVar(
		&hostCluster, "host-cluster", "local",
		"name of the host Kubernetes cluster",
	)
	initCmd.Flags().StringVar(
		&namespace, "namespace", "ghostcluster",
		"namespace to install Ghostcluster controller",
	)
	initCmd.Flags().StringVar(
		&gcpProject, "gcp-project", "",
		"GCP project ID for cluster provisioning",
	)
	initCmd.Flags().StringVar(
		&awsRegion, "aws-region", "us-west-2",
		"AWS region for cluster provisioning",
	)
	initCmd.Flags().BoolVar(
		&skipValidation, "skip-validation", false,
		"skip validation checks",
	)
}

func runInitCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	logger.Info("Initializing Ghostcluster controller",
		"hostCluster", hostCluster,
		"namespace", namespace,
	)

	// Initialize cluster manager
	cm := cluster.NewClusterManager()

	// Validate kubeconfig
	if !skipValidation {
		logger.Debug("Validating Kubernetes connection")
		if err := cm.ValidateConnection(); err != nil {
			logger.Error("Failed to connect to host cluster", "error", err)
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// Create namespace
	logger.Info("Creating namespace", "namespace", namespace)
	if err := cm.CreateNamespace(namespace); err != nil {
		logger.Error("Failed to create namespace", "error", err)
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Install vCluster controller
	logger.Info("Installing vCluster controller components")
	if err := cm.InstallController(namespace); err != nil {
		logger.Error("Failed to install controller", "error", err)
		return fmt.Errorf("failed to install controller: %w", err)
	}

	// Configure cloud provider if specified
	if gcpProject != "" {
		logger.Info("Configuring GCP project", "project", gcpProject)
		if err := cm.ConfigureGCP(namespace, gcpProject); err != nil {
			logger.Error("Failed to configure GCP", "error", err)
			return fmt.Errorf("failed to configure GCP: %w", err)
		}
	}

	logger.Info("✓ Ghostcluster controller initialized successfully")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Verify installation: ghostctl status")
	fmt.Println("  2. Create your first cluster: ghostctl up --template default")
	fmt.Println("  3. View active clusters: ghostctl list")

	return nil
}
