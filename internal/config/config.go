// Package config provides configuration management for the cfk application
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// AppConfig holds the application configuration
type AppConfig struct {
	Clusters []KafkaClusterConfig `mapstructure:"clusters"`
	UI       UIConfig             `mapstructure:"ui"`
}

// KafkaClusterConfig holds configuration for a Kafka cluster
type KafkaClusterConfig struct {
	Name      string   `mapstructure:"name"`
	Bootstrap []string `mapstructure:"bootstrap_servers"`
	Username  string   `mapstructure:"username,omitempty"`
	Password  string   `mapstructure:"password,omitempty"`
	SSL       bool     `mapstructure:"ssl"`
	SASL      bool     `mapstructure:"sasl"`
	SASLType  string   `mapstructure:"sasl_type,omitempty"` // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
}

// UIConfig holds UI-related configuration
type UIConfig struct {
	Theme            string `mapstructure:"theme"`
	RefreshInterval  int    `mapstructure:"refresh_interval"`
	MaxMessagesShown int    `mapstructure:"max_messages_shown"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Clusters: []KafkaClusterConfig{},
		UI: UIConfig{
			Theme:            "default",
			RefreshInterval:  5,
			MaxMessagesShown: 100,
		},
	}
}

// LoadAppConfig loads the application configuration from file
func LoadAppConfig() (*AppConfig, error) {
	// Set default configuration
	config := DefaultConfig()

	// Setup viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Look for config in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	cfkDir := filepath.Join(homeDir, ".cfk")
	v.AddConfigPath(cfkDir)

	// Create config directory if it doesn't exist
	if _, err := os.Stat(cfkDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cfkDir, 0755); err != nil {
			return nil, fmt.Errorf("could not create config directory: %w", err)
		}
	}

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we'll create it with defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configPath := filepath.Join(cfkDir, "config.yaml")
			if err := SaveAppConfig(config, configPath); err != nil {
				return nil, fmt.Errorf("could not create default config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("could not read config: %w", err)
		}
	} else {
		// Config file found, unmarshal
		if err := v.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("could not parse config: %w", err)
		}
	}

	return config, nil
}

// SaveAppConfig saves the application configuration to the specified path
func SaveAppConfig(config *AppConfig, path string) error {
	v := viper.New()
	v.SetConfigFile(path)

	// Set config values
	v.Set("clusters", config.Clusters)
	v.Set("ui", config.UI)

	// Write config to file
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}

	return nil
}
