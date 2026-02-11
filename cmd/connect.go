package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var (
	connectPathOnly bool
	connectSet      bool
)

var connectCmd = &cobra.Command{
	Use:   "connect <cluster-name>",
	Short: "Get connection info for a vCluster",
	Long: `Display the kubeconfig for a vCluster.

By default, outputs the export command that you can eval:
  eval $(ghostctl connect my-cluster)

Or use 'source' to apply in current shell:
  source <(ghostctl connect my-cluster)

Use --path-only to get just the kubeconfig path:
  ghostctl connect my-cluster --path-only

Use --set to update shell configuration files globally:
  ghostctl connect my-cluster --set

Note: For running commands with cluster context, use 'ghostctl exec' instead:
  ghostctl exec test -- kubectl get pods`,
	Args: cobra.ExactArgs(1),
	RunE: runConnectCmd,
}

func init() {
	connectCmd.Flags().BoolVar(
		&connectPathOnly, "path-only", false,
		"print only the kubeconfig path",
	)
	connectCmd.Flags().BoolVar(
		&connectSet, "set", false,
		"update shell config files (~/.bashrc, ~/.zshrc) with KUBECONFIG export",
	)
}

func runConnectCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]

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

	logger.Info("Getting kubeconfig for cluster", "name", clusterName)

	// Ensure kubeconfig exists (regenerate if necessary)
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

	// If --set flag is used, update shell config files
	if connectSet {
		return setGlobalKubeconfig(clusterName, kubePath, logger)
	}

	// Print based on flag
	if connectPathOnly {
		fmt.Println(kubePath)
	} else {
		fmt.Printf("export KUBECONFIG=%s\n", kubePath)
	}

	return nil
}

func setGlobalKubeconfig(clusterName, kubePath string, logger interface{ Info(msg string, args ...interface{}) }) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	exportLine := fmt.Sprintf("export KUBECONFIG=%s", kubePath)
	marker := "# ghostctl kubeconfig export"
	newExport := fmt.Sprintf("%s\n%s", marker, exportLine)

	// Map of shell config files and their export syntax
	shellConfigs := map[string]string{
		filepath.Join(homeDir, ".bashrc"):              newExport, // bash on Linux
		filepath.Join(homeDir, ".bash_profile"):        newExport, // bash on macOS
		filepath.Join(homeDir, ".zshrc"):               newExport, // zsh
		filepath.Join(homeDir, ".kshrc"):               newExport, // ksh
		filepath.Join(homeDir, ".tcshrc"):              "setenv KUBECONFIG " + kubePath, // tcsh
		filepath.Join(homeDir, ".config/fish/config.fish"): "set -gx KUBECONFIG " + kubePath, // fish
	}

	updatedFiles := []string{}

	for configFile, exportCmd := range shellConfigs {
		// Check if file exists
		fileInfo, err := os.Stat(configFile)
		if err != nil {
			// File doesn't exist, skip it
			continue
		}

		// Read existing content
		content, err := os.ReadFile(configFile)
		if err != nil {
			continue
		}

		contentStr := string(content)

		// Check if already has ghostctl export
		if contains(contentStr, marker) {
			continue // Already set
		}

		// Append the export
		if contentStr != "" && !endsWithNewline(contentStr) {
			contentStr += "\n"
		}
		contentStr += exportCmd + "\n"

		// Write back
		if err := os.WriteFile(configFile, []byte(contentStr), fileInfo.Mode()); err != nil {
			return fmt.Errorf("failed to update %s: %w", configFile, err)
		}

		updatedFiles = append(updatedFiles, configFile)
		if logger != nil {
			logger.Info("Updated", "file", configFile)
		}
	}

	if len(updatedFiles) == 0 {
		fmt.Printf("⚠ No shell config files found. Update manually with:\n")
		fmt.Printf("  export KUBECONFIG=%s\n", kubePath)
		return nil
	}

	fmt.Printf("✓ KUBECONFIG for '%s' set in %d shell config file(s)\n", clusterName, len(updatedFiles))
	for _, f := range updatedFiles {
		fmt.Printf("  - %s\n", f)
	}
	fmt.Printf("\nRestart your terminal or run: source ~/.bashrc  (or appropriate shell config)\n")

	return nil
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

func endsWithNewline(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '\n'
}
