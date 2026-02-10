package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool

	// Version information (injected at build time)
	Version   string = "dev"
	Commit    string = "unknown"
	BuildTime string = "unknown"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ghostctl",
	Short: "Manage ephemeral Kubernetes clusters with vCluster",
	Long: `ghostctl is a CLI tool for managing ephemeral Kubernetes clusters using vCluster.

Create, manage, and destroy virtual Kubernetes clusters for experiments, PRs, and notebooks.

Examples:
  ghostctl init                           # Initialize Ghostcluster controller
  ghostctl up --template default --ttl 1h # Create a new cluster
  ghostctl list                           # List active clusters
  ghostctl status <cluster-name>          # Check cluster status
  ghostctl down <cluster-name>            # Destroy a cluster`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildTime),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			telemetry.SetLogLevel("debug")
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "",
		"config file (default is $HOME/.ghost/config.yaml)",
	)
	RootCmd.PersistentFlags().BoolVarP(
		&verbose, "verbose", "v", false,
		"enable verbose logging",
	)

	// Add subcommands
	RootCmd.AddCommand(
		initCmd,
		upCmd,
		downCmd,
		listCmd,
		statusCmd,
		logsCmd,
		execCmd,
		templatesCmd,
	)
}

func initConfig() {
	// Config initialization happens in the config package
	// This is called by Cobra before any command runs
}

func Execute() error {
	return RootCmd.Execute()
}
