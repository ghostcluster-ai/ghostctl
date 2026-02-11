package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
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

	// Ensure kubeconfig exists (regenerate if necessary)
	kubeMgr, err := kubeconfig.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create kubeconfig manager: %w", err)
	}

	kubePath, err := kubeMgr.Get(clusterName, meta.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig for cluster %q: %w", clusterName, err)
	}

	// Save current KUBECONFIG as root (parent) on first connection
	saveRootKubeconfig()

	// If --set flag is used, update shell config files
	if connectSet {
		if err := setGlobalKubeconfig(clusterName, kubePath, nil); err != nil {
			return err
		}
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

	// Detect current shell and suggest appropriate source command
	shellSourceCmd := getShellSourceCommand(updatedFiles)
	if shellSourceCmd != "" {
		fmt.Printf("\n# Apply now in current shell:\n")
		fmt.Printf("%s\n", shellSourceCmd)
	} else {
		fmt.Printf("\nRestart your terminal or run: source ~/.bashrc  (or appropriate shell config)\n")
	}

	return nil
}

// getShellSourceCommand returns the appropriate source command for the current shell
func getShellSourceCommand(updatedFiles []string) string {
	shell := os.Getenv("SHELL")
	
	// Map shells to their config files and source command format
	shellMap := map[string]string{
		"bash":   "source ~/.bashrc",
		"zsh":    "source ~/.zshrc",
		"ksh":    "source ~/.kshrc",
		"tcsh":   "source ~/.tcshrc",
		"fish":   "source ~/.config/fish/config.fish",
	}

	// Extract shell name from path (e.g., /bin/zsh -> zsh)
	shellName := filepath.Base(shell)
	
	if cmd, ok := shellMap[shellName]; ok {
		return cmd
	}

	// Fallback to first updated file
	if len(updatedFiles) > 0 {
		return fmt.Sprintf("source %s", updatedFiles[0])
	}

	return ""
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

func endsWithNewline(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '\n'
}

// saveRootKubeconfig saves the current KUBECONFIG as the root (parent) on first connection
// Subsequent connects don't overwrite it, allowing disconnect to always return to root
func saveRootKubeconfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	ghostDir := filepath.Join(homeDir, ".ghost")
	rootKubeconfigPath := filepath.Join(ghostDir, ".root_kubeconfig")

	// Check if root is already saved
	if _, err := os.Stat(rootKubeconfigPath); err == nil {
		// Root already saved, don't overwrite
		return nil
	}

	// Get current KUBECONFIG (this is the parent/root)
	currentKubeconfig := os.Getenv("KUBECONFIG")
	if currentKubeconfig == "" {
		currentKubeconfig = "unset"
	}

	// Write to file
	if err := os.WriteFile(rootKubeconfigPath, []byte(currentKubeconfig), 0600); err != nil {
		// Silently fail - this is not critical
		return nil
	}

	return nil
}
