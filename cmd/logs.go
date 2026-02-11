package cmd

import (
	"fmt"
	"os"

	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/shell"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <cluster-name> [pod-name]",
	Short: "Stream logs from a vCluster pod",
	Long: `View and stream logs from pods in a vCluster.

Note: You can also use 'ghostctl connect' to switch contexts and use kubectl directly:
  ghostctl connect my-cluster
  kubectl logs my-pod -f
  ghostctl disconnect

Examples:
  ghostctl logs my-cluster                               # Show available pods
  ghostctl logs my-cluster my-pod -f                     # Stream pod logs
  ghostctl logs my-cluster my-pod --tail 100             # Show last 100 lines
  ghostctl logs my-cluster my-pod -n default             # From specific namespace`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLogsCmd,
}

var (
	logsNamespace string
	follow        bool
	tail          int
	since         string
)

func init() {
	logsCmd.Flags().StringVarP(
		&logsNamespace, "namespace", "n", "default",
		"namespace to get logs from",
	)
	logsCmd.Flags().BoolVarP(
		&follow, "follow", "f", false,
		"follow log stream",
	)
	logsCmd.Flags().IntVar(
		&tail, "tail", -1,
		"number of recent log lines to display",
	)
	logsCmd.Flags().StringVar(
		&since, "since", "",
		"show logs since time (e.g., 1h, 30m, 10s)",
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
	)

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

	// Get kubeconfig
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

	// Build kubectl logs command
	args = []string{"logs", "-n", logsNamespace}
	if podName != "" {
		args = append(args, podName)
	} else {
		// If no pod specified, list pods instead
		args = []string{"get", "pods", "-n", logsNamespace}
	}

	if follow && podName != "" {
		args = append(args, "-f")
	}

	if tail != -1 {
		args = append(args, "--tail="+fmt.Sprint(tail))
	}

	if since != "" {
		args = append(args, "--since="+since)
	}

	// Set up environment with kubeconfig
	env := os.Environ()
	env = append(env, "KUBECONFIG="+kubePath)

	// Execute kubectl with streaming output
	exitCode, err := shell.ExecuteCommandStreamingWithEnv(env, "kubectl", args...)
	if err != nil {
		logger.Error("Failed to get logs", "error", err)
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("kubectl exited with code %d", exitCode)
	}

	return nil
}
