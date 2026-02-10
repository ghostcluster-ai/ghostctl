package cmd

import (
	"fmt"
	"strings"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <cluster-name> <command> [args...]",
	Short: "Execute commands in a cluster",
	Long: `Execute kubectl or arbitrary commands within a cluster.

This command provides a convenient way to run commands inside a vCluster without
manually connecting to it. All commands are executed as if you had kubectl configured
to point to that cluster.

Examples:
  ghostctl exec my-cluster 'kubectl get pods'              # Run kubectl command
  ghostctl exec my-cluster 'kubectl apply -f app.yaml'     # Deploy application
  ghostctl exec my-cluster 'helm install myapp myrepo/app' # Use Helm
  ghostctl exec my-cluster bash -c "for i in {1..5}; do sleep 1; done"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runExecCmd,
}

var (
	execNamespace string
	execPod       string
	execContainer string
	execStdin     bool
	execTTY       bool
)

func init() {
	execCmd.Flags().StringVar(
		&execNamespace, "namespace", "default",
		"namespace for command execution",
	)
	execCmd.Flags().StringVar(
		&execPod, "pod", "",
		"specific pod to execute command in",
	)
	execCmd.Flags().StringVar(
		&execContainer, "container", "",
		"specific container within the pod",
	)
	execCmd.Flags().BoolVar(
		&execStdin, "stdin", false,
		"keep stdin open",
	)
	execCmd.Flags().BoolVar(
		&execTTY, "tty", false,
		"allocate a pseudo-TTY",
	)
}

func runExecCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]
	cmdToExecute := strings.Join(args[1:], " ")

	logger.Info("Executing command in cluster",
		"cluster", clusterName,
		"command", cmdToExecute,
	)

	// Initialize cluster manager
	cm := cluster.NewClusterManager()

	// Prepare execution options
	opts := &cluster.ExecOptions{
		Namespace: execNamespace,
		Pod:       execPod,
		Container: execContainer,
		Stdin:     execStdin,
		TTY:       execTTY,
	}

	// Execute command
	result, err := cm.ExecuteCommand(clusterName, cmdToExecute, opts)
	if err != nil {
		logger.Error("Failed to execute command", "error", err)
		return fmt.Errorf("failed to execute command: %w", err)
	}

	// Display output
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}

	if result.Stderr != "" {
		fmt.Fprintf(fmt.Stderr(), "%s", result.Stderr)
	}

	if result.ExitCode != 0 {
		logger.Error("Command failed with non-zero exit code", "code", result.ExitCode)
		return fmt.Errorf("command exited with code %d", result.ExitCode)
	}

	return nil
}
