package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ghostcluster-ai/ghostctl/internal/config"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:                "exec <cluster-name> -- <command> [args...]",
	Short:              "Execute commands in a vCluster",
	DisableFlagParsing: true,
	Long: `Execute commands (e.g. kubectl) against a vCluster.

This command runs a command with the vCluster's kubeconfig set as KUBECONFIG,
allowing you to interact with the virtual cluster without switching contexts.

Note: With the new connect/disconnect workflow, you can also:
  ghostctl connect <cluster-name>
  kubectl <command>  # Works directly
  ghostctl disconnect

Examples:
  ghostctl exec my-cluster -- kubectl get pods
  ghostctl exec my-cluster -- kubectl apply -f deployment.yaml
  ghostctl exec my-cluster -- helm list`,
	RunE: runExecCmd,
}

func runExecCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	// Better error messages for common mistakes
	if len(args) == 0 {
		return fmt.Errorf("missing cluster name\n\nUsage: ghostctl exec <cluster-name> -- <command> [args...]\n\nExample:\n  ghostctl exec test -- kubectl get pods\n\nNote: You can also use 'ghostctl connect test' to switch contexts globally.")
	}

	// Find the "--" separator
	dashIndex := -1
	for i, arg := range args {
		if arg == "--" {
			dashIndex = i
			break
		}
	}

	if dashIndex == -1 {
		return fmt.Errorf("missing '--' separator\n\nUsage: ghostctl exec <cluster-name> -- <command> [args...]\n\nExample:\n  ghostctl exec test -- kubectl get pods\n\nNote: You can also use 'ghostctl connect test' to switch contexts globally.")
	}

	if dashIndex == 0 {
		return fmt.Errorf("missing cluster name\n\nUsage: ghostctl exec <cluster-name> -- <command> [args...]\n\nExample:\n  ghostctl exec test -- kubectl get pods")
	}

	clusterName := args[0]
	commandArgs := args[dashIndex+1:]

	if len(commandArgs) == 0 {
		return fmt.Errorf("no command specified after --")
	}

	logger.Info("Executing command in vCluster",
		"cluster", clusterName,
		"command", commandArgs[0],
	)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return fmt.Errorf("failed to load config: %w", err)
	}

	namespace := cfg.Namespace
	if namespace == "" {
		namespace = vcluster.DefaultNamespace
	}

	kubeMgr, err := vcluster.NewKubeconfigManager("", namespace)
	if err != nil {
		logger.Error("Failed to create kubeconfig manager", "error", err)
		return fmt.Errorf("failed to create kubeconfig manager: %w", err)
	}

	ref := vcluster.ClusterRef{Name: clusterName, Namespace: namespace}
	kubePath, err := kubeMgr.GetOrCreateKubeconfig(ref)
	if err != nil {
		logger.Error("Failed to get kubeconfig", "error", err)
		if strings.Contains(err.Error(), "vcluster CLI not found") {
			return err
		}
		return fmt.Errorf("virtual cluster %q not found in namespace %q; ensure it exists or run 'ghostctl up %s' first: %w", clusterName, namespace, clusterName, err)
	}

	// Set up environment with kubeconfig
	env := os.Environ()
	env = append(env, "KUBECONFIG="+kubePath)

	// Execute command with streaming output
	runCmd := exec.Command(commandArgs[0], commandArgs[1:]...)
	runCmd.Env = env
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin

	if err := runCmd.Run(); err != nil {
		logger.Error("Failed to execute command", "error", err)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}
