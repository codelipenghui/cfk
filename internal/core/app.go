// Package core provides the main application logic for the cfk application
package core

import (
	"context"
	"fmt"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/cfk-dev/cfk/internal/kafka"
)

// App represents the main application
type App struct {
	Config      *config.AppConfig
	KafkaClient *kafka.Client
	CurrentView string
}

// NewApp creates a new application instance
func NewApp(cfg *config.AppConfig) *App {
	return &App{
		Config:      cfg,
		CurrentView: "clusters",
	}
}

// ConnectToCluster connects to a Kafka cluster
func (a *App) ConnectToCluster(clusterName string) error {
	// Find the cluster config
	var clusterConfig config.KafkaClusterConfig
	found := false

	for _, cluster := range a.Config.Clusters {
		if cluster.Name == clusterName {
			clusterConfig = cluster
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("cluster %s not found in configuration", clusterName)
	}

	// Create and connect Kafka client
	a.KafkaClient = kafka.NewClient(clusterConfig)
	if err := a.KafkaClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to cluster %s: %w", clusterName, err)
	}

	return nil
}

// ListTopics lists all topics in the connected Kafka cluster
func (a *App) ListTopics(ctx context.Context) ([]string, error) {
	if a.KafkaClient == nil {
		return nil, fmt.Errorf("not connected to any Kafka cluster")
	}

	return a.KafkaClient.ListTopics(ctx)
}

// GetTopicInfo gets detailed information about a specific topic
func (a *App) GetTopicInfo(ctx context.Context, topicName string) (*kafka.TopicInfo, error) {
	if a.KafkaClient == nil {
		return nil, fmt.Errorf("not connected to any Kafka cluster")
	}

	return a.KafkaClient.GetTopicInfo(ctx, topicName)
}

// AddCluster adds a new Kafka cluster configuration
func (a *App) AddCluster(cluster config.KafkaClusterConfig) error {
	// Check if cluster with same name already exists
	for _, c := range a.Config.Clusters {
		if c.Name == cluster.Name {
			return fmt.Errorf("cluster with name %s already exists", cluster.Name)
		}
	}

	// Add the new cluster
	a.Config.Clusters = append(a.Config.Clusters, cluster)

	// Save the updated configuration
	homeDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := homeDir + "/config.yaml"
	if err := config.SaveAppConfig(a.Config, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// UpdateCluster updates an existing Kafka cluster configuration
func (a *App) UpdateCluster(cluster config.KafkaClusterConfig) error {
	// Find the cluster to update
	found := false
	for i, c := range a.Config.Clusters {
		if c.Name == cluster.Name {
			// Update the cluster
			a.Config.Clusters[i] = cluster
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("cluster with name %s not found", cluster.Name)
	}

	// Save the updated configuration
	homeDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := homeDir + "/config.yaml"
	if err := config.SaveAppConfig(a.Config, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// RemoveCluster removes a Kafka cluster configuration
func (a *App) RemoveCluster(clusterName string) error {
	// Find the cluster to remove
	found := false
	var updatedClusters []config.KafkaClusterConfig
	
	for _, c := range a.Config.Clusters {
		if c.Name == clusterName {
			found = true
		} else {
			updatedClusters = append(updatedClusters, c)
		}
	}

	if !found {
		return fmt.Errorf("cluster with name %s not found", clusterName)
	}

	// Update the clusters list
	a.Config.Clusters = updatedClusters

	// Save the updated configuration
	homeDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := homeDir + "/config.yaml"
	if err := config.SaveAppConfig(a.Config, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// CreateTopic creates a new topic in the connected Kafka cluster
func (a *App) CreateTopic(ctx context.Context, topicName string, numPartitions int, replicationFactor int) error {
	if a.KafkaClient == nil {
		return fmt.Errorf("not connected to any Kafka cluster")
	}

	return a.KafkaClient.CreateTopic(ctx, topicName, numPartitions, replicationFactor)
}

// DeleteTopic deletes a topic from the connected Kafka cluster
func (a *App) DeleteTopic(ctx context.Context, topicName string) error {
	if a.KafkaClient == nil {
		return fmt.Errorf("not connected to any Kafka cluster")
	}

	return a.KafkaClient.DeleteTopic(ctx, topicName)
}

// UpdateTopicPartitions updates the number of partitions for a topic
func (a *App) UpdateTopicPartitions(ctx context.Context, topicName string, numPartitions int) error {
	if a.KafkaClient == nil {
		return fmt.Errorf("not connected to any Kafka cluster")
	}

	return a.KafkaClient.UpdateTopicPartitions(ctx, topicName, numPartitions)
}
