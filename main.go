package main

import (
	"os"

	"github.com/ghostcluster-ai/ghostctl/cmd"
	"github.com/ghostcluster-ai/ghostctl/internal/telemetry"
)

func main() {
	// Initialize logging
	telemetry.InitLogger()

	// Execute root command
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
