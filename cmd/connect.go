package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

For automatic shell integration, use --set (one-time setup):
  ghostctl connect my-cluster --set
  source ~/.zshrc  (or ~/.bashrc)

After setup, simply use:
  ghostctl connect my-cluster

Use --path-only to get just the kubeconfig path:
  ghostctl connect my-cluster --path-only`,
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
		"add shell integration to your shell config (~/.bashrc, ~/.zshrc, etc.)",
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

	// Detect current shell
	shell := os.Getenv("SHELL")
	shellName := filepath.Base(shell)

	var configFile string
	var shellInitContent string

	switch shellName {
	case "bash":
		// Check if on macOS or Linux
		bashProfile := filepath.Join(homeDir, ".bash_profile")
		bashrc := filepath.Join(homeDir, ".bashrc")
		if _, err := os.Stat(bashProfile); err == nil {
			configFile = bashProfile
		} else {
			configFile = bashrc
		}
		shellInitContent = getBashShellInit()
	case "zsh":
		configFile = filepath.Join(homeDir, ".zshrc")
		shellInitContent = getBashShellInit() // bash/zsh use same syntax
	case "fish":
		configFile = filepath.Join(homeDir, ".config/fish/config.fish")
		shellInitContent = getFishShellInit()
	default:
		return fmt.Errorf("unsupported shell: %s", shellName)
	}

	// Read existing content
	content := ""
	if data, err := os.ReadFile(configFile); err == nil {
		content = string(data)
	}

	marker := "# ghostctl shell integration"
	updated := false

	// Check if shell integration is already present
	if !contains(content, marker) {
		// Add shell integration
		if content != "" && !endsWithNewline(content) {
			content += "\n"
		}
		content += "\n" + shellInitContent + "\n"
		updated = true
	}

	// Write back if updated
	if updated {
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to update %s: %w", configFile, err)
		}
	}

	fmt.Printf("✓ Shell integration added to: %s\n", configFile)
	fmt.Printf("\n# To activate in current shell, run:\n")
	fmt.Printf("source %s\n", configFile)
	fmt.Printf("\n# After sourcing, you can simply use:\n")
	fmt.Printf("ghostctl connect %s\n", clusterName)

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (strings.Contains(s, substr))
}

func endsWithNewline(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '\n'
}

// getBashShellInit returns the shell integration content for bash/zsh
func getBashShellInit() string {
	return `# ghostctl shell integration for bash/zsh
# This allows: ghostctl connect <cluster> (sets KUBECONFIG in current shell)

_ghostctl_connect() {
    local cluster="$1"
    if [[ -z "$cluster" ]]; then
        echo "Usage: ghostctl connect <cluster-name>"
        return 1
    fi
    
    local export_stmt=$(command ghostctl connect "$cluster" 2>/dev/null)
    if [[ $? -eq 0 ]]; then
        eval "$export_stmt"
        echo "✓ Connected to cluster: $cluster"
    else
        echo "Error: Could not connect to cluster: $cluster"
        return 1
    fi
}

_ghostctl_disconnect() {
    local restore_stmt=$(command ghostctl disconnect 2>/dev/null)
    if [[ $? -eq 0 ]]; then
        eval "$restore_stmt"
        echo "✓ Disconnected from cluster"
    else
        echo "No previous kubeconfig to restore"
        return 1
    fi
}

ghostctl() {
    if [[ "$1" == "connect" ]] && [[ ! "$2" =~ ^- ]]; then
        shift
        _ghostctl_connect "$@"
    elif [[ "$1" == "disconnect" ]]; then
        _ghostctl_disconnect
    else
        command ghostctl "$@"
    fi
}`
}

// getFishShellInit returns the shell integration content for fish
func getFishShellInit() string {
	return `# ghostctl shell integration for fish
# This allows: ghostctl connect <cluster> (sets KUBECONFIG in current shell)

function _ghostctl_connect
    set cluster $argv[1]
    if test -z "$cluster"
        echo "Usage: ghostctl connect <cluster-name>"
        return 1
    end
    
    set export_stmt (command ghostctl connect "$cluster" 2>/dev/null)
    if test $status -eq 0
        eval "$export_stmt"
        echo "✓ Connected to cluster: $cluster"
    else
        echo "Error: Could not connect to cluster: $cluster"
        return 1
    end
end

function _ghostctl_disconnect
    set restore_stmt (command ghostctl disconnect 2>/dev/null)
    if test $status -eq 0
        eval "$restore_stmt"
        echo "✓ Disconnected from cluster"
    else
        echo "No previous kubeconfig to restore"
        return 1
    end
end

function ghostctl
    if test "$argv[1]" = "connect" -a -n "$argv[2]"
        set argv $argv[2..-1]
        _ghostctl_connect $argv
    else if test "$argv[1]" = "disconnect"
        _ghostctl_disconnect
    else
        command ghostctl $argv
    end
end`
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
