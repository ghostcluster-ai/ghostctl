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

var execCmd = &cobra.Command{
	Use:   "exec <cluster-name> -- <command> [args...]",
	Short: "Execute commands in a vCluster",
	Long: `Execute commands (e.g. kubectl) against a vCluster.

This command runs a command with the vCluster's kubeconfig set as KUBECONFIG,
allowing you to interact with the virtual cluster as if you had it configured
in your local kubectl.

Examples:
  ghostctl exec my-cluster -- kubectl get pods
  ghostctl exec my-cluster -- kubectl apply -f deployment.yaml
  ghostctl exec my-cluster -- helm list
  ghostctl exec my-cluster -- bash -c "kubectl port-forward svc/app 8080:80"`,
	RunE: runExecCmd,
}

func runExecCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	if len(args) < 2 {
		return fmt.Errorf("usage: exec <cluster-name> -- <command> [args...]")
	}

	clusterName := args[0]
	// args[1] should be "--", skip it
	commandArgs := args[2:]

	if len(commandArgs) == 0 {
		return fmt.Errorf("no command specified after --")
	}

	logger.Info("Executing command in vCluster",
		"cluster", clusterName,
		"command", commandArgs[0],
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

	// Set up environment with kubeconfig
	env := os.Environ()
	env = append(env, "KUBECONFIG="+kubePath)

	// Execute command with streaming output
	exitCode, err := shell.ExecuteCommandStreamingWithEnv(env, commandArgs[0], commandArgs[1:]...)
	if err != nil {
		logger.Error("Failed to execute command", "error", err)
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("command exited with code %d", exitCode)
	}

	return nil
}
