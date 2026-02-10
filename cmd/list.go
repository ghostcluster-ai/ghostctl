package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active vClusters",
	Long: `List all active virtual Kubernetes clusters managed by ghostctl.

This command displays cluster names, status, creation time, and TTL information
from the local metadata store.

Examples:
  ghostctl list                  # List all clusters
  ghostctl list --output json    # Output as JSON
  ghostctl list --output yaml    # Output as YAML`,
	RunE: runListCmd,
}

var (
	outputFormat string
)

func init() {
	listCmd.Flags().StringVar(
		&outputFormat, "output", "table",
		"output format (table, json, yaml)",
	)
}

func runListCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	logger.Info("Listing vClusters")

	// Initialize metadata store
	metaStore, err := metadata.NewStore()
	if err != nil {
		logger.Error("Failed to initialize metadata store", "error", err)
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}

	// List clusters from metadata
	clusters, err := metaStore.List()
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

func displayClustersTable(clusters []*metadata.ClusterMetadata) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()

	// Header
	_, _ = fmt.Fprintln(w, "NAME\tNAMESPACE\tSTATUS\tCREATED\tTTL")

	// Rows
	for _, c := range clusters {
		status := "unknown"
		// Check if cluster is actually running
		if err := vcluster.Status(c.Name, c.Namespace); err == nil {
			status = "running"
		} else {
			status = "offline"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			c.Name,
			c.Namespace,
			status,
			c.CreatedAt.Format("2006-01-02 15:04"),
			c.TTL,
		)
	}
}

func displayClustersJSON(clusters []*metadata.ClusterMetadata) error {
	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal clusters to JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func displayClustersYAML(clusters []*metadata.ClusterMetadata) error {
	data, err := yaml.Marshal(clusters)
	if err != nil {
		return fmt.Errorf("failed to marshal clusters to YAML: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
