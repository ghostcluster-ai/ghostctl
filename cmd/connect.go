package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <cluster-name>",
	Short: "Connect to a vCluster",
	Long: `Connect to a vCluster by merging its kubeconfig and switching context.

This merges the vCluster kubeconfig into your ~/.kube/config and switches
the current context, so kubectl commands will work with the vCluster.

Examples:
  ghostctl connect my-cluster      # Connect to my-cluster`,
	Args: cobra.ExactArgs(1),
	RunE: runConnectCmd,
}

func runConnectCmd(cmd *cobra.Command, args []string) error {
	clusterName := args[0]

	// Initialize metadata store
	metaStore, err := metadata.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}

	// Get cluster metadata
	meta, err := metaStore.Get(clusterName)
	if err != nil {
		return fmt.Errorf("cluster %q not found in local registry", clusterName)
	}

	// Initialize kubeconfig manager
	kubeMgr, err := kubeconfig.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create kubeconfig manager: %w", err)
	}

	// Save current context (for disconnect)
	if err := kubeMgr.SaveCurrentContext(); err != nil {
		// Non-fatal, but warn user
		fmt.Printf("Warning: Could not save current context: %v\n", err)
	}

	// Merge vCluster kubeconfig into ~/.kube/config
	contextName, err := kubeMgr.MergeIntoDefault(clusterName, meta.Namespace)
	if err != nil {
		return fmt.Errorf("failed to merge kubeconfig: %w", err)
	}

	// Switch to vCluster context
	if err := kubeMgr.SwitchContext(contextName); err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	fmt.Printf("✓ Connected to cluster: %s\n", clusterName)
	fmt.Printf("Context switched to: %s\n", contextName)

	return nil
}
