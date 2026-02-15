package cmd

import (
	"fmt"
	"strings"

	"github.com/ghostcluster-ai/ghostctl/internal/config"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <cluster-name>",
	Short: "Print kubeconfig for a vCluster",
	Long: `Print the kubeconfig export for a vCluster.

This fetches a real vCluster kubeconfig (if needed) and prints an export
line you can eval to target the virtual cluster with kubectl.

Examples:
  ghostctl connect my-cluster      # Connect to my-cluster`,
	Args: cobra.ExactArgs(1),
	RunE: runConnectCmd,
}

func runConnectCmd(cmd *cobra.Command, args []string) error {
	clusterName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	namespace := cfg.Namespace
	if namespace == "" {
		namespace = vcluster.DefaultNamespace
	}

	kubeMgr, err := vcluster.NewKubeconfigManager("", namespace)
	if err != nil {
		return fmt.Errorf("failed to create kubeconfig manager: %w", err)
	}

	ref := vcluster.ClusterRef{Name: clusterName, Namespace: namespace}
	kubePath, err := kubeMgr.GetOrCreateKubeconfig(ref)
	if err != nil {
		if strings.Contains(err.Error(), "vcluster CLI not found") {
			return err
		}
		return fmt.Errorf("virtual cluster %q not found in namespace %q; ensure it exists or run 'ghostctl up %s' first: %w", clusterName, namespace, clusterName, err)
	}

	fmt.Printf("# Use this virtual cluster with kubectl:\n")
	fmt.Printf("export KUBECONFIG=%s\n", kubePath)

	return nil
}
