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

	// Debug code removed

	// Initialize application core
	app := core.NewApp(appConfig)
	
	// Pass the app to the TUI
	if err := tui.Start(appConfig, app); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
