package cmd

import (
	"fmt"
	"time"

	"github.com/ghostcluster-ai/ghostctl/internal/cluster"
	"github.com/ghostcluster-ai/ghostctl/internal/kubeconfig"
	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
	"github.com/ghostcluster-ai/ghostctl/internal/templates"
	"github.com/ghostcluster-ai/ghostctl/internal/vcluster"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [cluster-name]",
	Short: "Create a new ephemeral vCluster",
	Long: `Create a new virtual Kubernetes cluster in the host Kubernetes cluster.

The vCluster will be created in the ghostcluster namespace. Use 'ghostctl connect'
to switch to the cluster context and interact with it using kubectl.

You can use templates to apply predefined resource configurations, and override
individual settings with CLI flags.

Examples:
  ghostctl up my-cluster                         # Create with default template
  ghostctl up my-cluster --ttl 2h                # Create with 2 hour TTL
  ghostctl up ml-job --template gpu              # Use GPU template
  ghostctl up ml-job --template gpu --gpu 2      # Override GPU count
  ghostctl up test --template minimal --ttl 30m  # Minimal resources, 30m TTL
  ghostctl connect my-cluster                    # Connect to the cluster`,
	RunE: runUpCmd,
}

var (
	upTTL      string
	upTemplate string
	upCPU      string
	upMemory   string
	upStorage  string
	upGPU      int
	upGPUType  string
)

func init() {
	upCmd.Flags().StringVar(&upTTL, "ttl", "", "Time-to-live for the cluster (e.g., 30m, 2h, 1d)")
	upCmd.Flags().StringVar(&upTemplate, "template", "default", "Template to use (default, gpu, minimal, large)")
	upCmd.Flags().StringVar(&upCPU, "cpu", "", "CPU allocation (overrides template)")
	upCmd.Flags().StringVar(&upMemory, "memory", "", "Memory allocation (overrides template)")
	upCmd.Flags().StringVar(&upStorage, "storage", "", "Storage allocation (overrides template)")
	upCmd.Flags().IntVar(&upGPU, "gpu", 0, "Number of GPUs (overrides template)")
	upCmd.Flags().StringVar(&upGPUType, "gpu-type", "", "GPU type (overrides template)")
}

func runUpCmd(cmd *cobra.Command, args []string) error {
	logger := telemetry.GetLogger()

	// Get cluster name from args
	if len(args) == 0 {
		return fmt.Errorf("cluster name is required")
	}
	clusterName := args[0]

	// Initialize metadata store
	metaStore, err := metadata.NewStore()
	if err != nil {
		logger.Error("Failed to initialize metadata store", "error", err)
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}

	// Check if cluster already exists
	if metaStore.Exists(clusterName) {
		return fmt.Errorf("cluster %q already exists", clusterName)
	}

	// Load template and build create options
	opts, err := buildCreateOptions(cmd, clusterName, logger)
	if err != nil {
		return err
	}

	logger.Info("Creating new vCluster",
		"name", clusterName,
		"template", upTemplate,
		"ttl", opts.TTL,
		"cpu", opts.CPU,
		"memory", opts.Memory,
		"gpu", opts.GPU,
	)

	namespace := vcluster.DefaultNamespace

	// Create the vCluster
	logger.Info("Creating vCluster in Kubernetes")
	if err := vcluster.Create(clusterName, namespace); err != nil {
		logger.Error("Failed to create vCluster", "error", err)
		return err
	}

	// Wait for vCluster to be ready
	logger.Info("Waiting for vCluster to be ready")
	waitTimeout := 5 * time.Minute
	if err := vcluster.IsReady(clusterName, namespace, waitTimeout); err != nil {
		logger.Error("vCluster failed to become ready", "error", err)
		return err
	}

	// Retrieve and store kubeconfig
	logger.Info("Retrieving kubeconfig")
	kubeconfigPath, err := kubeconfig.NewManager()
	if err != nil {
		logger.Error("Failed to create kubeconfig manager", "error", err)
		return err
	}

	_, err = kubeconfigPath.Fresh(clusterName, namespace)
	if err != nil {
		logger.Error("Failed to retrieve kubeconfig", "error", err)
		return err
	}

	// Store metadata
	kubePath, _ := metadata.GetClusterPath(clusterName)
	meta := &metadata.ClusterMetadata{
		Name:           clusterName,
		Namespace:      namespace,
		CreatedAt:      time.Now(),
		TTL:            opts.TTL,
		KubeconfigPath: kubePath,
		HostCluster:    "current",
		Template:       upTemplate,
		CPU:            opts.CPU,
		Memory:         opts.Memory,
		GPU:            opts.GPU,
		Labels:         opts.Labels,
	}

	if err := metaStore.Add(meta); err != nil {
		logger.Error("Failed to store cluster metadata", "error", err)
		return fmt.Errorf("failed to store cluster metadata: %w", err)
	}

	logger.Info("✓ vCluster created successfully", "name", clusterName)
	
	// Display creation summary
	displayCreationSummary(clusterName, opts)

	return nil
}

// buildCreateOptions loads template and applies CLI flag overrides
func buildCreateOptions(cmd *cobra.Command, clusterName string, logger *telemetry.Logger) (*cluster.CreateOptions, error) {
	opts := &cluster.CreateOptions{
		Name:      clusterName,
		Namespace: vcluster.DefaultNamespace,
		Labels:    make(map[string]string),
	}

	// Load template if specified
	if upTemplate != "" {
		templatesDir := templates.GetTemplatesDir()
		store := templates.NewFileStore(templatesDir)

		tmpl, err := store.Get(upTemplate)
		if err != nil {
			// Template not found - show helpful message
			logger.Warn("Template not found, using defaults", "template", upTemplate)
			fmt.Printf("Warning: Template %q not found. Using default values.\n", upTemplate)
			fmt.Println("Run 'ghostctl templates' to see available templates.")
		} else {
			// Apply template defaults
			logger.Info("Loaded template", "name", tmpl.Name)
			opts.CPU = tmpl.CPU
			opts.Memory = tmpl.Memory
			opts.Storage = tmpl.Storage
			opts.GPU = tmpl.GPU
			opts.GPUType = tmpl.GPUType
			opts.TTL = tmpl.TTL
			if tmpl.Labels != nil {
				opts.Labels = tmpl.Labels
			}
		}
	}

	// Apply CLI flag overrides (flags take precedence over template)
	if cmd.Flags().Changed("cpu") {
		opts.CPU = upCPU
	}
	if cmd.Flags().Changed("memory") {
		opts.Memory = upMemory
	}
	if cmd.Flags().Changed("storage") {
		opts.Storage = upStorage
	}
	if cmd.Flags().Changed("gpu") {
		opts.GPU = upGPU
	}
	if cmd.Flags().Changed("gpu-type") {
		opts.GPUType = upGPUType
	}
	if cmd.Flags().Changed("ttl") {
		opts.TTL = upTTL
	}

	return opts, nil
}

// displayCreationSummary shows a summary of the created cluster
func displayCreationSummary(clusterName string, opts *cluster.CreateOptions) {
	fmt.Printf("\n✓ Cluster '%s' is ready!\n", clusterName)
	
	if upTemplate != "" {
		fmt.Printf("\nTemplate: %s\n", upTemplate)
	}
	
	fmt.Println("\nResources:")
	if opts.CPU != "" {
		fmt.Printf("  CPU:     %s\n", opts.CPU)
	}
	if opts.Memory != "" {
		fmt.Printf("  Memory:  %s\n", opts.Memory)
	}
	if opts.Storage != "" {
		fmt.Printf("  Storage: %s\n", opts.Storage)
	}
	if opts.GPU > 0 {
		fmt.Printf("  GPU:     %d", opts.GPU)
		if opts.GPUType != "" {
			fmt.Printf(" (%s)", opts.GPUType)
		}
		fmt.Println()
	}
	if opts.TTL != "" {
		fmt.Printf("  TTL:     %s\n", opts.TTL)
	}

	fmt.Println("\nUseful commands:")
	fmt.Printf("  ghostctl status %s               # Check cluster status\n", clusterName)
	fmt.Printf("  ghostctl connect %s              # Switch to this cluster\n", clusterName)
	fmt.Printf("  ghostctl exec %s -- kubectl ... # Run command in cluster\n", clusterName)
	fmt.Printf("  ghostctl disconnect              # Return to parent cluster\n")
	fmt.Printf("  ghostctl down %s                 # Destroy cluster\n", clusterName)
}
