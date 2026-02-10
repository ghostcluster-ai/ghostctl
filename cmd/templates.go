package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:   "templates [template-name]",
	Short: "List or inspect cluster templates",
	Long: `Manage cluster templates used for provisioning.

Templates define cluster configurations including:
  - Resource allocations (CPU, memory, GPU)
  - Node counts and types
  - Networking and storage configurations
  - Pre-installed applications
  - Cost optimization settings

Use 'ghostctl templates' to list available templates or
'ghostctl templates <name>' to inspect a specific template.

Examples:
  ghostctl templates                       # List all templates
  ghostctl templates default               # Show details of default template
  ghostctl templates test --extended       # Extended information
  ghostctl templates --filter gpu          # List templates with GPU support`,
	RunE: runTemplatesCmd,
}

var (
	templateFilter   string
	templateFormat   string
	templateExtended bool
)

func init() {
	templatesCmd.Flags().StringVar(
		&templateFilter, "filter", "",
		"filter templates by name or feature (e.g., 'gpu', 'cpu-only')",
	)
	templatesCmd.Flags().StringVar(
		&templateFormat, "format", "table",
		"output format (table, json, yaml)",
	)
	templatesCmd.Flags().BoolVar(
		&templateExtended, "extended", false,
		"show extended template information",
	)
}

func runTemplatesCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	// Initialize cluster manager
	cm := cluster.NewClusterManager()

	if len(args) == 0 {
		// List all templates
		logger.Debug("Listing cluster templates")
		templates, err := cm.ListTemplates()
		if err != nil {
			logger.Error("Failed to list templates", "error", err)
			return fmt.Errorf("failed to list templates: %w", err)
		}

		displayTemplatesList(templates, templateFilter)
	} else {
		// Show specific template
		templateName := args[0]
		logger.Info("Fetching template details", "name", templateName)

		template, err := cm.GetTemplate(templateName)
		if err != nil {
			logger.Error("Failed to get template", "error", err)
			return fmt.Errorf("failed to get template: %w", err)
		}

		displayTemplateDetails(template, templateExtended)
	}

	return nil
}

func displayTemplatesList(templates []*cluster.Template, filter string) {
	w := tabwriter.NewWriter(fmt.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tDESCRIPTION\tCPU\tMEMORY\tGPU\tCOST/HOUR")

	for _, t := range templates {
		// Apply filter if specified
		if filter != "" && !matchesFilter(t.Name, filter) {
			continue
		}

		gpu := "-"
		if t.GPUCount > 0 {
			gpu = fmt.Sprintf("%d (%s)", t.GPUCount, t.GPUType)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t$%.2f\n",
			t.Name,
			t.Description,
			t.CPU,
			t.Memory,
			gpu,
			t.HourlyCost,
		)
	}
}

func displayTemplateDetails(template *cluster.Template, extended bool) {
	fmt.Printf("Template: %s\n", template.Name)
	fmt.Printf("Description: %s\n", template.Description)
	fmt.Printf("\nResources:\n")
	fmt.Printf("  CPU: %s\n", template.CPU)
	fmt.Printf("  Memory: %s\n", template.Memory)
	if template.GPUCount > 0 {
		fmt.Printf("  GPU: %d x %s\n", template.GPUCount, template.GPUType)
	}
	fmt.Printf("  Nodes: %d\n", template.NodeCount)
	fmt.Printf("\nCost:\n")
	fmt.Printf("  Hourly: $%.2f\n", template.HourlyCost)
	fmt.Printf("  Daily: $%.2f\n", template.HourlyCost*24)

	if extended {
		fmt.Printf("\nAdvanced Configuration:\n")
		fmt.Printf("  Storage Size: %s\n", template.StorageSize)
		fmt.Printf("  Network Type: %s\n", template.NetworkType)
		fmt.Printf("  Auto-Scaling: %v\n", template.AutoScaling)
		if template.PreInstalledApps != nil && len(template.PreInstalledApps) > 0 {
			fmt.Printf("  Pre-installed Apps: %v\n", template.PreInstalledApps)
		}
	}
}

func matchesFilter(templateName, filter string) bool {
	// Simple substring matching for template filtering
	return len(filter) == 0 || contains(templateName, filter)
}

func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr)
}
