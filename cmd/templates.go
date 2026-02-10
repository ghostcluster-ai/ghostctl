package cmd

import (
	"fmt"

	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage cluster templates (upcoming feature)",
	Long: `Cluster templates allow you to define standard configurations for vClusters.

This feature is coming soon. For now, vClusters are created with basic default
configuration. Custom templates will be supported in a future release.

Examples:
  ghostctl templates list      # List available templates (coming soon)
  ghostctl templates create    # Create a custom template (coming soon)`,
	RunE: runTemplatesCmd,
}

func runTemplatesCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	logger.Info("Templates feature coming soon")

	fmt.Println("Cluster templates are coming soon!")
	fmt.Println("\nFor now, use: ghostctl up <name> to create a basic vCluster")
	fmt.Println("\nIn the future, you'll be able to:")
	fmt.Println("  - Define custom templates with resource specifications")
	fmt.Println("  - Share templates across teams")
	fmt.Println("  - Version control template definitions")

	return nil
}
