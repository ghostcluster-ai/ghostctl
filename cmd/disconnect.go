package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/spf13/cobra"
)

var (
	disconnectCleanup bool
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from vCluster and restore parent context",
	Long: `Restore the parent kubeconfig context before connecting to a vCluster.

This restores the kubectl context that was active before you first
connected to a vCluster.

Examples:
  ghostctl disconnect              # Restore parent context
  ghostctl disconnect --cleanup    # Also remove vCluster context from kubeconfig`,
	RunE: runDisconnectCmd,
}

func init() {
	disconnectCmd.Flags().BoolVar(
		&disconnectCleanup, "cleanup", false,
		"remove the vCluster context from kubeconfig",
	)
}

func runDisconnectCmd(cmd *cobra.Command, args []string) error {
	// Initialize kubeconfig manager
	kubeMgr, err := kubeconfig.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create kubeconfig manager: %w", err)
	}

	// Get current context before restoring (for cleanup)
	currentContext := ""
	if disconnectCleanup {
		currentContext, _ = kubeMgr.GetCurrentContext()
	}

	// Restore saved context
	savedContext, err := kubeMgr.RestoreSavedContext()
	if err != nil {
		return fmt.Errorf("failed to restore context: %w", err)
	}

	fmt.Printf("✓ Disconnected from vCluster\n")
	fmt.Printf("Context restored to: %s\n", savedContext)

	// Optionally cleanup the vCluster context
	if disconnectCleanup && currentContext != "" {
		if err := kubeMgr.RemoveContext(currentContext); err != nil {
			fmt.Printf("Warning: Could not remove context %s: %v\n", currentContext, err)
		} else {
			fmt.Printf("✓ Removed context: %s\n", currentContext)
		}
	}

	return nil
}
