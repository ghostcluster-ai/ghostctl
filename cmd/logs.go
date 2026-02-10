package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <cluster-name> [pod-name]",
	Short: "Stream logs from a cluster",
	Long: `View and stream logs from a cluster or a specific pod.

By default, streams logs from all pods in the cluster. You can filter by:
  - Pod name (exact or prefix match)
  - Namespace
  - Label selectors
  - Previous logs (from crashed containers)

Examples:
  ghostctl logs my-cluster                # Stream logs from all pods
  ghostctl logs my-cluster my-pod         # Stream logs from specific pod
  ghostctl logs my-cluster --namespace kube-system # Logs from system namespace
  ghostctl logs my-cluster my-pod --tail 100       # Last 100 lines
  ghostctl logs my-cluster --follow=false --since 1h # Logs from last hour`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLogsCmd,
}

var (
	logsNamespace string
	logsPodLabel  string
	logsContainer string
	follow        bool
	tail          int
	since         string
	timestamps    bool
	previous      bool
	allContainers bool
)

func init() {
	logsCmd.Flags().StringVar(
		&logsNamespace, "namespace", "",
		"namespace to get logs from (empty = all namespaces)",
	)
	logsCmd.Flags().StringVar(
		&logsPodLabel, "labels", "",
		"label selector for pods (e.g., app=myapp)",
	)
	logsCmd.Flags().StringVar(
		&logsContainer, "container", "",
		"specific container name",
	)
	logsCmd.Flags().BoolVarP(
		&follow, "follow", "f", true,
		"follow log stream",
	)
	logsCmd.Flags().IntVar(
		&tail, "tail", 10,
		"number of recent log lines to display",
	)
	logsCmd.Flags().StringVar(
		&since, "since", "",
		"show logs since time (e.g., 1h, 30m, 10s)",
	)
	logsCmd.Flags().BoolVar(
		&timestamps, "timestamps", false,
		"include timestamps in log output",
	)
	logsCmd.Flags().BoolVar(
		&previous, "previous", false,
		"show logs from previous container instance",
	)
	logsCmd.Flags().BoolVar(
		&allContainers, "all-containers", false,
		"show logs from all containers",
	)
}

func runLogsCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]

	var podName string
	if len(args) > 1 {
		podName = args[1]
	}

	logger.Info("Fetching cluster logs",
		"cluster", clusterName,
		"pod", podName,
		"follow", follow,
	)

	// Initialize cluster manager
	cm := cluster.NewClusterManager()

	// Prepare log options
	opts := &cluster.LogOptions{
		PodName:       podName,
		Namespace:     logsNamespace,
		Container:     logsContainer,
		Follow:        follow,
		Tail:          int64(tail),
		Since:         since,
		Timestamps:    timestamps,
		Previous:      previous,
		AllContainers: allContainers,
	}

	// Get logs
	logStream, err := cm.GetLogs(clusterName, opts)
	if err != nil {
		logger.Error("Failed to get logs", "error", err)
		return fmt.Errorf("failed to get logs: %w", err)
	}
	defer func() { _ = logStream.Close() }() // nolint:errcheck

	// Stream logs to stdout
	if _, err := io.Copy(os.Stdout, logStream); err != nil {
		logger.Error("Failed to stream logs", "error", err)
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	return nil
}
