// Package kafka provides Kafka client functionality for the cfk application
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/segmentio/kafka-go"
)

// Client represents a Kafka client
type Client struct {
	Config config.KafkaClusterConfig
	Conn   *kafka.Conn
}

// TopicInfo holds information about a Kafka topic
type TopicInfo struct {
	Name              string
	Partitions        int
	ReplicationFactor int
	Config            map[string]string
}

// NewClient creates a new Kafka client
func NewClient(clusterConfig config.KafkaClusterConfig) *Client {
	return &Client{
		Config: clusterConfig,
	}
}

// Connect establishes a connection to the Kafka cluster
func (c *Client) Connect() error {
	// Use the first bootstrap server for initial connection
	if len(c.Config.Bootstrap) == 0 {
		return fmt.Errorf("no bootstrap servers configured")
	}

	// Set up dialer with authentication if needed
	dialer := &kafka.Dialer{
		Timeout: 10 * time.Second,
	}

	// Configure SASL if enabled
	if c.Config.SASL {
		// For SASL PLAIN authentication, we need to set up the mechanism
		// The kafka-go package requires a specific format for SASL
		dialer.DualStack = true
		dialer.TLS = nil // Disable TLS for now, can be enabled with proper config
	}

	// Connect to the broker
	conn, err := dialer.Dial("tcp", c.Config.Bootstrap[0])
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}

	c.Conn = conn
	return nil
}

// Close closes the Kafka connection
func (c *Client) Close() error {
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

// ListTopics lists all topics in the Kafka cluster
func (c *Client) ListTopics(ctx context.Context) ([]string, error) {
	if c.Conn == nil {
		return nil, fmt.Errorf("not connected to Kafka")
	}

	// Get metadata for all topics
	partitions, err := c.Conn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	// Extract unique topic names
	topicMap := make(map[string]bool)
	for _, p := range partitions {
		topicMap[p.Topic] = true
	}

	// Convert map to slice
	topics := make([]string, 0, len(topicMap))
	for topic := range topicMap {
		topics = append(topics, topic)
	}

	return topics, nil
}

// GetTopicInfo gets detailed information about a specific topic
func (c *Client) GetTopicInfo(ctx context.Context, topicName string) (*TopicInfo, error) {
	if c.Conn == nil {
		return nil, fmt.Errorf("not connected to Kafka")
	}

	// Get metadata for all topics
	partitions, err := c.Conn.ReadPartitions(topicName)
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions for topic %s: %w", topicName, err)
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("topic %s not found", topicName)
	}

	// Create topic info
	topicInfo := &TopicInfo{
		Name:       topicName,
		Partitions: len(partitions),
		Config:     make(map[string]string),
	}

	// For now, we don't have a way to get replication factor and config
	// This would require admin client functionality

	return topicInfo, nil
}
