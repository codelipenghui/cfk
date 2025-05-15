package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/cfk-dev/cfk/internal/kafka"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Command represents a command that can be executed by the TUI
type Command func() tea.Msg

// CommandMsg is a message containing a command to be executed
type CommandMsg struct {
	Cmd Command
}

// ErrorMsg is a message containing an error
type ErrorMsg struct {
	Err error
}

// TopicsLoadedMsg is a message containing loaded topics
type TopicsLoadedMsg struct {
	Topics []string
}

// ConnectedMsg is a message indicating a successful connection
type ConnectedMsg struct {
	ClusterName string
}

// LoadTopicsCmd returns a command that loads topics from Kafka
func LoadTopicsCmd(client *kafka.Client) Command {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		topics, err := client.ListTopics(ctx)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to list topics: %w", err)}
		}

		return TopicsLoadedMsg{Topics: topics}
	}
}

// ConnectToClusterCmd returns a command that connects to a Kafka cluster
func ConnectToClusterCmd(clusterConfig config.KafkaClusterConfig) Command {
	return func() tea.Msg {
		client := kafka.NewClient(clusterConfig)
		err := client.Connect()
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to connect to cluster %s: %w", clusterConfig.Name, err)}
		}

		return ConnectedMsg{ClusterName: clusterConfig.Name}
	}
}

// UpdateTopicListCmd returns a command that updates the topic list
func UpdateTopicListCmd(topics []string) Command {
	return func() tea.Msg {
		items := make([]list.Item, len(topics))
		for i, topic := range topics {
			items[i] = NewTopicItem(topic, nil)
		}

		return list.NewItems(items)
	}
}

// UpdateClusterListCmd returns a command that updates the cluster list
func UpdateClusterListCmd(clusters []config.KafkaClusterConfig) Command {
	return func() tea.Msg {
		items := make([]list.Item, len(clusters))
		for i, cluster := range clusters {
			items[i] = NewClusterItem(cluster.Name, cluster.Bootstrap)
		}

		return list.NewItems(items)
	}
}
