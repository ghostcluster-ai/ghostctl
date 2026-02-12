package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/ghostcluster-ai/ghostctl/internal/templates"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	templateFilter   string
	templateFormat   string
	templateExtended bool
)

var templatesCmd = &cobra.Command{
	Use:   "templates [name]",
	Short: "List and inspect cluster templates",
	Long: `List available cluster templates or inspect a specific template.

Templates define standard configurations for vClusters with predefined
resource allocations, GPU settings, and time-to-live values.

Examples:
  ghostctl templates                    # List all templates
  ghostctl templates gpu                # Show details for 'gpu' template
  ghostctl templates --filter gpu       # Filter templates by keyword
  ghostctl templates --format json      # Output as JSON
  ghostctl templates gpu --extended     # Show full template details`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTemplatesCmd,
}

func init() {
	templatesCmd.Flags().StringVar(&templateFilter, "filter", "", "Filter templates by name or feature")
	templatesCmd.Flags().StringVar(&templateFormat, "format", "table", "Output format: table, json, yaml")
	templatesCmd.Flags().BoolVar(&templateExtended, "extended", false, "Show extended template information")
}

func runTemplatesCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	// Get templates directory
	templatesDir := templates.GetTemplatesDir()
	store := templates.NewFileStore(templatesDir)

	// If a specific template name is provided, show details
	if len(args) == 1 {
		return showTemplateDetails(store, args[0], logger)
	}

	// List all templates
	return listTemplates(store, logger)
}

func listTemplates(store templates.Store, logger *telemetry.Logger) error {
	logger.Info("Listing templates")

	templateList, err := store.List()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fmt.Println("No templates directory found.")
			fmt.Println("\nTemplates are expected in one of these locations:")
			fmt.Printf("  - %s\n", templates.GetTemplatesDir())
			fmt.Println("\nCreate template YAML files to get started.")
			return nil
		}
		return fmt.Errorf("failed to list templates: %w", err)
	}

	// Apply filter if specified
	if templateFilter != "" {
		filtered := []templates.Template{}
		for _, tmpl := range templateList {
			if strings.Contains(strings.ToLower(tmpl.Name), strings.ToLower(templateFilter)) ||
				strings.Contains(strings.ToLower(tmpl.Description), strings.ToLower(templateFilter)) {
				filtered = append(filtered, tmpl)
			}
		}
		templateList = filtered
	}

	if len(templateList) == 0 {
		if templateFilter != "" {
			fmt.Printf("No templates found matching filter: %s\n", templateFilter)
		} else {
			fmt.Println("No templates found.")
		}
		return nil
	}

	// Output in requested format
	switch templateFormat {
	case "json":
		return outputJSON(templateList)
	case "yaml":
		return outputYAML(templateList)
	case "table":
		return outputTable(templateList)
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", templateFormat)
	}
}

func showTemplateDetails(store templates.Store, name string, logger *telemetry.Logger) error {
	logger.Info("Showing template details", "name", name)

	tmpl, err := store.Get(name)
	if err != nil {
		return err
	}

	// Output in requested format
	switch templateFormat {
	case "json":
		data, err := json.MarshalIndent(tmpl, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(tmpl)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	default:
		// Default: human-readable details
		fmt.Printf("Template: %s\n", tmpl.Name)
		fmt.Printf("Description: %s\n\n", tmpl.Description)

		fmt.Println("Resources:")
		if tmpl.CPU != "" {
			fmt.Printf("  CPU:     %s\n", tmpl.CPU)
		}
		if tmpl.Memory != "" {
			fmt.Printf("  Memory:  %s\n", tmpl.Memory)
		}
		if tmpl.Storage != "" {
			fmt.Printf("  Storage: %s\n", tmpl.Storage)
		}
		if tmpl.GPU > 0 {
			fmt.Printf("  GPU:     %d", tmpl.GPU)
			if tmpl.GPUType != "" {
				fmt.Printf(" (%s)", tmpl.GPUType)
			}
			fmt.Println()
		}
		if tmpl.TTL != "" {
			fmt.Printf("  TTL:     %s\n", tmpl.TTL)
		}

		if len(tmpl.Labels) > 0 && templateExtended {
			fmt.Println("\nLabels:")
			for k, v := range tmpl.Labels {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		fmt.Printf("\nUsage:\n")
		fmt.Printf("  ghostctl up my-cluster --template %s\n", tmpl.Name)
		if tmpl.GPU > 0 {
			fmt.Printf("  ghostctl up my-cluster --template %s --gpu 2  # Override GPU count\n", tmpl.Name)
		}
	}

	return nil
}

func outputTable(templateList []templates.Template) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	if templateExtended {
		fmt.Fprintln(w, "NAME\tDESCRIPTION\tCPU\tMEMORY\tSTORAGE\tGPU\tGPU TYPE\tTTL")
	} else {
		fmt.Fprintln(w, "NAME\tDESCRIPTION\tCPU\tMEMORY\tGPU\tTTL")
	}

	// Rows
	for _, tmpl := range templateList {
		cpu := tmpl.CPU
		if cpu == "" {
			cpu = "-"
		}
		memory := tmpl.Memory
		if memory == "" {
			memory = "-"
		}
		storage := tmpl.Storage
		if storage == "" {
			storage = "-"
		}
		gpu := "-"
		if tmpl.GPU > 0 {
			gpu = fmt.Sprintf("%d", tmpl.GPU)
		}
		gpuType := tmpl.GPUType
		if gpuType == "" {
			gpuType = "-"
		}
		ttl := tmpl.TTL
		if ttl == "" {
			ttl = "-"
		}

		if templateExtended {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				tmpl.Name, tmpl.Description, cpu, memory, storage, gpu, gpuType, ttl)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				tmpl.Name, tmpl.Description, cpu, memory, gpu, ttl)
		}
	}

	return nil
}

func outputJSON(templateList []templates.Template) error {
	data, err := json.MarshalIndent(templateList, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputYAML(templateList []templates.Template) error {
	data, err := yaml.Marshal(map[string]interface{}{
		"templates": templateList,
	})
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}
