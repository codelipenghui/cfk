package main

import (
	"fmt"
	"os"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/cfk-dev/cfk/internal/core"
	"github.com/cfk-dev/cfk/internal/tui"
)

func main() {
	fmt.Println("Welcome to cfk - Console for Kafka!")

	// Load configuration
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize application core
	app := core.NewApp(appConfig)

	// Initialize and start TUI
	if err := tui.Start(appConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
