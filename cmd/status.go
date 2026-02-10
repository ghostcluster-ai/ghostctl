package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <cluster-name>",
	Short: "Display cluster status and resource usage",
	Long: `Show detailed status information about a specific cluster including:
  - Current status (running, creating, terminating, failed)
  - Resource allocation and usage (CPU, memory, GPU)
  - GPU utilization
  - Pod status
  - Time-to-live (TTL) remaining
  - Estimated cost

Examples:
  ghostctl status my-cluster              # Show cluster status
  ghostctl status my-cluster --watch      # Watch status updates in real-time
  ghostctl status my-cluster --verbose    # Show detailed information`,
	Args: cobra.ExactArgs(1),
	RunE: runStatusCmd,
}

var (
	watch    bool
	detailed bool
)

func init() {
	statusCmd.Flags().BoolVar(
		&watch, "watch", false,
		"watch cluster status in real-time",
	)
	statusCmd.Flags().BoolVar(
		&detailed, "detailed", false,
		"show detailed status information",
	)
}

func runStatusCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]

	logger.Info("Fetching cluster status", "name", clusterName)

	// Initialize cluster manager
	cm := cluster.NewClusterManager()

	// Get cluster status
	status, err := cm.GetClusterStatus(clusterName)
	if err != nil {
		logger.Error("Failed to get cluster status", "error", err)
		return fmt.Errorf("failed to get cluster status: %w", err)
	}

	// Display status
	displayStatus(status, detailed)

	return nil
}

func displayStatus(status *cluster.ClusterStatus, detailed bool) {
	fmt.Printf("Cluster: %s\n", status.Name)
	fmt.Printf("Status: %s\n", status.Status)
	fmt.Printf("Created: %s\n", status.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("TTL Remaining: %s\n", status.TTLRemaining)
	fmt.Printf("\nResources:\n")
	fmt.Printf("  CPU Requested: %s | CPU Used: %s (%.1f%%)\n",
		status.CPURequested,
		status.CPUUsed,
		status.CPUUsagePercent,
	)
	fmt.Printf("  Memory Requested: %s | Memory Used: %s (%.1f%%)\n",
		status.MemoryRequested,
		status.MemoryUsed,
		status.MemoryUsagePercent,
	)

	if status.GPUCount > 0 {
		fmt.Printf("  GPU Count: %d\n", status.GPUCount)
		fmt.Printf("  GPU Type: %s\n", status.GPUType)
		fmt.Printf("  GPU Utilization: %.1f%%\n", status.GPUUtilization)
	}

	fmt.Printf("\nPods: %d running, %d pending, %d failed\n",
		status.RunningPods,
		status.PendingPods,
		status.FailedPods,
	)

	fmt.Printf("\nEstimated Cost:\n")
	fmt.Printf("  Hourly: $%.2f\n", status.HourlyCost)
	fmt.Printf("  Total (projected): $%.2f\n", status.EstimatedTotalCost)

	if detailed {
		fmt.Printf("\nDetailed Information:\n")
		fmt.Printf("  Version: %s\n", status.Version)
		fmt.Printf("  Kubernetes Version: %s\n", status.KubernetesVersion)
		fmt.Printf("  Nodes: %d\n", status.NodeCount)
	}
}
