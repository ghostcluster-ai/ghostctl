package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from vCluster and restore parent kubeconfig",
	Long: `Restore the parent kubeconfig before connecting to any vCluster.

This command restores the kubeconfig that was active before you first
connected to a vCluster, regardless of how many nested clusters you've
connected to.

Examples:
  ghostctl disconnect              # Restore parent KUBECONFIG`,
	RunE: runDisconnectCmd,
}

func runDisconnectCmd(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	ghostDir := filepath.Join(homeDir, ".ghost")
	rootKubeconfigPath := filepath.Join(ghostDir, ".root_kubeconfig")

	// Try to read the root kubeconfig
	rootKubeconfig, err := os.ReadFile(rootKubeconfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not connected to any vCluster - nothing to disconnect from")
		}
		return fmt.Errorf("failed to read root kubeconfig: %w", err)
	}

	rootConfig := string(rootKubeconfig)

	// Output the unset or export statement
	if rootConfig == "unset" {
		fmt.Println("unset KUBECONFIG")
	} else {
		fmt.Printf("export KUBECONFIG=%s\n", rootConfig)
	}

	// Clean up root marker so next connection saves a new root
	os.Remove(rootKubeconfigPath)

	return nil
}
