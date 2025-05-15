package tui

import (
	"fmt"
	"os"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/charmbracelet/bubbles/list"
)

// DebugClusterList prints debug information about the cluster list
func DebugClusterList(clusters []config.KafkaClusterConfig) {
	fmt.Println("DEBUG: Clusters in config:")
	for i, cluster := range clusters {
		fmt.Printf("%d. Name: %s, Bootstrap: %v\n", i+1, cluster.Name, cluster.Bootstrap)
	}

	// Create list items
	items := make([]list.Item, len(clusters))
	for i, cluster := range clusters {
		items[i] = NewClusterItem(cluster.Name, cluster.Bootstrap)
	}

	fmt.Println("\nDEBUG: List items created:")
	for idx, item := range items {
		if i, ok := item.(Item); ok {
			fmt.Printf("%d. Title: %s, Description: %s\n", idx+1, i.Title(), i.Description())
		} else {
			fmt.Printf("%d. Item is not of type Item\n", idx+1)
		}
	}

	// Write to a debug file
	f, err := os.Create("/tmp/cfk_debug.txt")
	if err != nil {
		fmt.Println("Error creating debug file:", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "DEBUG: Clusters in config:\n")
	for i, cluster := range clusters {
		fmt.Fprintf(f, "%d. Name: %s, Bootstrap: %v\n", i+1, cluster.Name, cluster.Bootstrap)
	}

	fmt.Fprintf(f, "\nDEBUG: List items created:\n")
	for idx, item := range items {
		if i, ok := item.(Item); ok {
			fmt.Fprintf(f, "%d. Title: %s, Description: %s\n", idx+1, i.Title(), i.Description())
		} else {
			fmt.Fprintf(f, "%d. Item is not of type Item\n", idx+1)
		}
	}
}