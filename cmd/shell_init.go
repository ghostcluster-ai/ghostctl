package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var shellInitCmd = &cobra.Command{
	Use:   "shell-init [bash|zsh|fish]",
	Short: "Generate shell configuration for ghostctl",
	Long: `Generate shell configuration to make ghostctl commands more convenient.

This outputs shell function and alias definitions that you can source to enhance
ghostctl usage. Automatically detect your shell or specify one:

  ghostctl shell-init bash
  ghostctl shell-init zsh  
  ghostctl shell-init fish

Add to your shell config (~/.bashrc, ~/.zshrc, ~/.config/fish/config.fish):

  eval "$(ghostctl shell-init)"`,
	RunE: runShellInitCmd,
}

func runShellInitCmd(cmd *cobra.Command, args []string) error {
	shell := "bash" // default
	if len(args) > 0 {
		shell = args[0]
	}

	switch shell {
	case "bash", "zsh":
		fmt.Println(bashShellInit)
	case "fish":
		fmt.Println(fishShellInit)
	default:
		return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
	}

	return nil
}

const bashShellInit = `# ghostctl shell integration for bash/zsh
# This allows: ghostctl connect test (sets KUBECONFIG in current shell)
# And also: ghostctl disconnect (restores previous KUBECONFIG)

_ghostctl_connect() {
    local cluster="$1"
    if [[ -z "$cluster" ]]; then
        echo "Usage: ghostctl connect <cluster-name>"
        return 1
    fi
    
    # Get the export statement and evaluate it
    local export_stmt=$(command ghostctl connect "$cluster" 2>/dev/null)
    local exit_code=$?
    if [[ $exit_code -eq 0 ]]; then
        eval "$export_stmt"
        echo "✓ Connected to cluster: $cluster"
    else
        echo "Error: Could not connect to cluster: $cluster"
        return $exit_code
    fi
}

_ghostctl_connect_set() {
    local cluster="$1"
    if [[ -z "$cluster" ]]; then
        echo "Usage: ghostctl connect <cluster-name> --set"
        return 1
    fi
    
    # Get the export statement from --set and evaluate it
    local export_stmt=$(command ghostctl connect "$cluster" --set 2>/dev/null)
    local exit_code=$?
    if [[ $exit_code -eq 0 ]]; then
        eval "$export_stmt"
        echo "✓ Connected to cluster: $cluster"
    else
        echo "Error: Could not connect to cluster: $cluster"
        return $exit_code
    fi
}

_ghostctl_disconnect() {
    # Get the restore statement and evaluate it
    local restore_stmt=$(command ghostctl disconnect 2>/dev/null)
    local exit_code=$?
    if [[ $exit_code -eq 0 ]]; then
        eval "$restore_stmt"
        echo "✓ Disconnected from cluster"
    else
        echo "No previous kubeconfig to restore"
        return $exit_code
    fi
}

# Create a wrapper for ghostctl
ghostctl() {
    case "$1" in
        connect)
            if [[ $# -lt 2 ]]; then
                echo "Usage: ghostctl connect <cluster-name> [--set|--path-only]"
                return 1
            fi
            # Check if --set or --path-only flags are used
            if [[ "$3" == "--set" ]]; then
                _ghostctl_connect_set "$2"
            elif [[ "$3" == "--path-only" ]] || [[ "$2" == "--path-only" ]]; then
                # For --path-only, just pass through to real command
                command ghostctl "$@"
            else
                # Regular connect with eval
                _ghostctl_connect "$2"
            fi
            ;;
        disconnect)
            _ghostctl_disconnect
            ;;
        *)
            # Pass through all other commands to the real ghostctl
            command ghostctl "$@"
            ;;
    esac
}
`

const fishShellInit = `# ghostctl shell integration for fish
# This allows: ghostctl connect test (sets KUBECONFIG in current shell)
# And also: ghostctl disconnect (restores previous KUBECONFIG)

function _ghostctl_connect
    set cluster $argv[1]
    if test -z "$cluster"
        echo "Usage: ghostctl connect <cluster-name>"
        return 1
    end
    
    # Get the export statement and evaluate it
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
    # Get the restore statement and evaluate it
    set restore_stmt (command ghostctl disconnect 2>/dev/null)
    if test $status -eq 0
        eval "$restore_stmt"
        echo "✓ Disconnected from cluster"
    else
        echo "No previous kubeconfig to restore"
        return 1
    end
end

# Create a wrapper for ghostctl
function ghostctl
    if test "$argv[1]" = "connect" -a -n "$argv[2]"
        # If it's "ghostctl connect <cluster>", use our wrapper
        set argv $argv[2..-1]
        _ghostctl_connect $argv
    else if test "$argv[1]" = "disconnect"
        # If it's "ghostctl disconnect", use our wrapper
        _ghostctl_disconnect
    else
        # Otherwise, use the real ghostctl command
        command ghostctl $argv
    end
end
`
