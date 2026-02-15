package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ghostcluster-ai/ghostctl/internal/config"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <cluster-name>",
	Short: "Display vCluster status",
	Long: `Show status information about a virtual Kubernetes cluster.

This displays whether the vCluster is running and accessible, creation time,
and time-to-live information.

Examples:
  ghostctl status my-cluster        # Show cluster status
  ghostctl status my-cluster -v     # Show detailed error information`,
	Args: cobra.ExactArgs(1),
	RunE: runStatusCmd,
}

func runStatusCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()
	clusterName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	namespace := cfg.Namespace
	if namespace == "" {
		namespace = vcluster.DefaultNamespace
	}

	logger.Info("Fetching cluster status", "name", clusterName)

	baseDir := resolveGhostDir()

	// Initialize metadata store
	var meta *metadata.ClusterMetadata
	metaStore, err := metadata.NewStore()
	if err != nil {
		logger.Warn("Failed to initialize metadata store", "error", err)
	} else {
		meta, err = metaStore.Get(clusterName)
		if err != nil {
			logger.Info("Local metadata for cluster not found; using live vCluster status", "name", clusterName, "path", baseDir)
		} else if meta.Namespace != "" {
			namespace = meta.Namespace
		}
	}

	ref := vcluster.ClusterRef{Name: clusterName, Namespace: namespace}

	// Check if vCluster exists
	exists := false
	reachable := false
	status := "not found"

	if err := vcluster.Status(clusterName, namespace); err == nil {
		exists = true
		status = "offline"
	} else if strings.Contains(err.Error(), "vcluster CLI not found") {
		return err
	} else if !strings.Contains(err.Error(), "vCluster not found") {
		status = "unknown"
	}

	var kubePath string
	kubeMgr, err := vcluster.NewKubeconfigManager("", namespace)
	if err != nil {
		logger.Warn("Failed to create kubeconfig manager", "error", err)
	} else {
		kubePath = kubeMgr.KubeconfigPath(ref)
	}

	if exists && kubeMgr != nil {
		path, err := kubeMgr.GetOrCreateKubeconfig(ref)
		if err == nil {
			kubePath = path
			if err := checkKubeconfigReachable(path); err == nil {
				reachable = true
				status = "running"
			} else {
				status = "unreachable"
			}
		}
	}

	// Display status
	displayStatus(clusterName, meta, namespace, kubePath, status, exists, reachable)

	return nil
}

func displayStatus(name string, meta *metadata.ClusterMetadata, namespace, kubePath, status string, exists, reachable bool) {
	fmt.Printf("Cluster: %s\n", name)
	fmt.Printf("Namespace: %s\n", namespace)
	fmt.Printf("Status: %s\n", status)

	if meta != nil {
		fmt.Printf("Created: %s\n", meta.CreatedAt.Format("2006-01-02 15:04:05"))
		if meta.TTL != "" {
			fmt.Printf("TTL: %s\n", meta.TTL)
		}
	} else {
		fmt.Printf("Created: unknown\n")
		fmt.Printf("TTL: unknown\n")
	}

	if kubePath != "" {
		fmt.Printf("Kubeconfig: %s\n", kubePath)
	} else {
		fmt.Printf("Kubeconfig: unknown\n")
	}

	if exists {
		if reachable {
			fmt.Printf("\n✓ vCluster is accessible\n")
		} else {
			fmt.Printf("\n✗ vCluster exists but is not accessible\n")
		}
	} else {
		fmt.Printf("\n✗ vCluster not found in host cluster\n")
	}

	fmt.Printf("\nTo connect, run:\n")
	fmt.Printf("  ghostctl connect %s\n", name)
}

func checkKubeconfigReachable(kubeconfigPath string) error {
	cmd := exec.Command("kubectl", "--kubeconfig", kubeconfigPath, "get", "ns")
	return cmd.Run()
}

func resolveGhostDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return metadata.DefaultDir
	}
	return filepath.Join(home, metadata.DefaultDir)
}
