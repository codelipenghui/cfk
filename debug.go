package main

import (
	"fmt"
	"os"

	"github.com/cfk-dev/cfk/internal/config"
)

func main() {
	// Load configuration
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Print the clusters
	fmt.Println("Clusters in config:")
	for i, cluster := range appConfig.Clusters {
		fmt.Printf("%d. Name: %s, Bootstrap: %v\n", i+1, cluster.Name, cluster.Bootstrap)
	}
}