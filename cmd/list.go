package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active vClusters",
	Long: `List all active virtual Kubernetes clusters managed by Ghostcluster.

This command displays cluster name, status, resource usage, and TTL information.

Examples:
  ghostctl list                    # List all clusters
  ghostctl list --namespace dev    # List clusters in specific namespace
  ghostctl list --sort ttl         # Sort by time-to-live
  ghostctl list --output json      # Output as JSON`,
	RunE: runListCmd,
}

var (
	listNamespace string
	sortBy        string
	outputFormat  string
	allNamespaces bool
)

func init() {
	listCmd.Flags().StringVar(
		&listNamespace, "namespace", "ghostcluster",
		"namespace to list clusters from",
	)
	listCmd.Flags().BoolVar(
		&allNamespaces, "all-namespaces", false,
		"list clusters from all namespaces",
	)
	listCmd.Flags().StringVar(
		&sortBy, "sort", "name",
		"sort by (name, status, ttl, created)",
	)
	listCmd.Flags().StringVar(
		&outputFormat, "output", "table",
		"output format (table, json, yaml)",
	)
}

func runListCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	if allNamespaces {
		logger.Debug("Listing clusters from all namespaces")
	} else {
		logger.Debug("Listing clusters from namespace", "namespace", listNamespace)
	}

	// Initialize cluster manager
	cm := cluster.NewClusterManager()

	// List clusters
	var clusters []*cluster.ClusterInfo
	var err error
	if allNamespaces {
		clusters, err = cm.ListClustersAllNamespaces()
	} else {
		clusters, err = cm.ListClusters(listNamespace)
	}

	if err != nil {
		logger.Error("Failed to list clusters", "error", err)
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	if len(clusters) == 0 {
		fmt.Println("No active clusters found")
		return nil
	}

	// Display clusters based on output format
	switch outputFormat {
	case "json":
		return displayClustersJSON(clusters)
	case "yaml":
		return displayClustersYAML(clusters)
	default:
		displayClustersTable(clusters)
	}

	return nil
}

func displayClustersTable(clusters []*cluster.ClusterInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }() // nolint:errcheck

	// Header
	_, _ = fmt.Fprintln(w, "NAME\tNAMESPACE\tSTATUS\tGPU\tMEMORY\tTTL\tCREATED\tESTIMATED COST") // nolint:errcheck

	// Rows
	for _, c := range clusters {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t$%0.2f\n", // nolint:errcheck
			c.Name,
			c.Namespace,
			c.Status,
			c.GPUCount,
			c.Memory,
			c.TTL,
			c.CreatedAt.Format("2006-01-02 15:04"),
			c.EstimatedCost,
		)
	}
}

func displayClustersJSON(clusters []*cluster.ClusterInfo) error {
	// JSON formatting would be implemented here
	_, _ = fmt.Fprintln(os.Stdout, "JSON output format not yet implemented") // nolint:errcheck
	return nil
}

func displayClustersYAML(clusters []*cluster.ClusterInfo) error {
	// YAML formatting would be implemented here
	fmt.Println("YAML output format not yet implemented")
	return nil
}
