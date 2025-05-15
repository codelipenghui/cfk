package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetConfigDir returns the path to the cfk configuration directory
func GetConfigDir() (string, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	// Create .cfk directory path
	cfkDir := filepath.Join(homeDir, ".cfk")

	// Create directory if it doesn't exist
	if _, err := os.Stat(cfkDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cfkDir, 0755); err != nil {
			return "", fmt.Errorf("could not create config directory: %w", err)
		}
	}

	return cfkDir, nil
}
