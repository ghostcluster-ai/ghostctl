package cmd

import (
	"fmt"
	"time"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [cluster-name]",
	Short: "Create a new ephemeral vCluster",
	Long: `Create a new virtual Kubernetes cluster for your experiment, PR, or notebook.

The cluster will be provisioned with specified resources and will automatically
be destroyed after the TTL (time-to-live) expires.

Examples:
  ghostctl up                              # Create default cluster with default template
  ghostctl up my-lab --template gpt4       # Create named cluster with specific template
  ghostctl up ml-exp --gpu 1 --ttl 4h      # Create cluster with GPU, 4 hour TTL
  ghostctl up pr-123 --from-pr 123 --gpu 2 # Create from PR context with 2 GPUs`,
	RunE: runUpCmd,
}

var (
	clusterTemplate string
	gpuCount        int
	gpuType         string
	ttl             string
	memory          string
	cpu             string
	fromPR          string
	wait            bool
	waitTimeout     string
	dryRun          bool
)

func init() {
	upCmd.Flags().StringVar(
		&clusterTemplate, "template", "default",
		"cluster template to use (default: default)",
	)
	upCmd.Flags().IntVar(
		&gpuCount, "gpu", 0,
		"number of GPUs to allocate",
	)
	upCmd.Flags().StringVar(
		&gpuType, "gpu-type", "nvidia-t4",
		"type of GPU (e.g., nvidia-t4, nvidia-a100)",
	)
	upCmd.Flags().StringVar(
		&ttl, "ttl", "1h",
		"time-to-live for the cluster (default: 1h, examples: 30m, 2h, 1d)",
	)
	upCmd.Flags().StringVar(
		&memory, "memory", "4Gi",
		"memory allocation for the cluster",
	)
	upCmd.Flags().StringVar(
		&cpu, "cpu", "2",
		"CPU allocation for the cluster",
	)
	upCmd.Flags().StringVar(
		&fromPR, "from-pr", "",
		"create cluster from PR context (PR number)",
	)
	upCmd.Flags().BoolVar(
		&wait, "wait", true,
		"wait for cluster to be ready",
	)
	upCmd.Flags().StringVar(
		&waitTimeout, "wait-timeout", "5m",
		"timeout for waiting for cluster readiness",
	)
	upCmd.Flags().BoolVar(
		&dryRun, "dry-run", false,
		"simulate cluster creation without actually creating it",
	)
}

func runUpCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	// Get cluster name or generate one
	clusterName := "ghostctl"
	if len(args) > 0 {
		clusterName = args[0]
	}

	logger.Info("Creating new cluster",
		"name", clusterName,
		"template", clusterTemplate,
		"gpu", gpuCount,
		"ttl", ttl,
	)

	if dryRun {
		fmt.Println("DRY RUN: Would create cluster with the following configuration:")
		fmt.Printf("  Name: %s\n", clusterName)
		fmt.Printf("  Template: %s\n", clusterTemplate)
		fmt.Printf("  GPU Count: %d\n", gpuCount)
		fmt.Printf("  GPU Type: %s\n", gpuType)
		fmt.Printf("  TTL: %s\n", ttl)
		fmt.Printf("  Memory: %s\n", memory)
		fmt.Printf("  CPU: %s\n", cpu)
		return nil
	}

	// Initialize cluster manager
	cm := cluster.NewClusterManager()

	// Prepare cluster config
	config := &cluster.Config{
		Name:      clusterName,
		Template:  clusterTemplate,
		GPU:       gpuCount,
		GPUType:   gpuType,
		TTL:       ttl,
		Memory:    memory,
		CPU:       cpu,
		FromPR:    fromPR,
		Namespace: "ghostcluster",
	}

	// Create cluster
	logger.Info("Provisioning cluster")
	if err := cm.CreateCluster(config); err != nil {
		logger.Error("Failed to create cluster", "error", err)
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	// Wait for cluster to be ready if requested
	if wait {
		logger.Info("Waiting for cluster to be ready", "timeout", waitTimeout)
		timeout, _ := time.ParseDuration(waitTimeout)
		if err := cm.WaitForCluster(clusterName, timeout); err != nil {
			logger.Error("Cluster failed to become ready", "error", err)
			return fmt.Errorf("cluster not ready: %w", err)
		}
	}

	logger.Info("✓ Cluster created successfully", "name", clusterName)
	fmt.Printf("\nCluster '%s' is ready!\n", clusterName)
	fmt.Println("\nUseful commands:")
	fmt.Printf("  ghostctl status %s                    # Check cluster status\n", clusterName)
	fmt.Printf("  ghostctl exec %s 'kubectl get pods'   # Run command in cluster\n", clusterName)
	fmt.Printf("  ghostctl logs %s -f                   # Stream cluster logs\n", clusterName)
	fmt.Printf("  ghostctl down %s                      # Destroy cluster\n", clusterName)

	return nil
}
