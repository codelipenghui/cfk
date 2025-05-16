package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/cfk-dev/cfk/internal/core"
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
	err error
}

// TopicsLoadedMsg is a message containing loaded topics
type TopicsLoadedMsg struct {
	Topics []string
}

// ConnectedMsg is a message indicating a successful connection
type ConnectedMsg struct {
	ClusterName string
}

// ItemsUpdatedMsg is a message containing updated list items
type ItemsUpdatedMsg struct {
	Items []list.Item
}

// LoadTopicsCmd returns a command that loads topics from Kafka
func LoadTopicsCmd(client *kafka.Client) Command {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		topics, err := client.ListTopics(ctx)
		if err != nil {
			return ErrorMsg{err: fmt.Errorf("failed to list topics: %w", err)}
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
			return ErrorMsg{err: fmt.Errorf("failed to connect to cluster %s: %w", clusterConfig.Name, err)}
		}

		return ConnectedMsg{ClusterName: clusterConfig.Name}
	}
}

// UpdateTopicListCmd returns a command that updates the topic list
func UpdateTopicListCmd(app *core.App) Command {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		topics, err := app.ListTopics(ctx)
		if err != nil {
			return ErrorMsg{err: fmt.Errorf("failed to list topics: %w", err)}
		}

		items := make([]list.Item, len(topics))
		for i, topic := range topics {
			// Get topic info to show partitions
			topicInfo, err := app.GetTopicInfo(ctx, topic)
			if err != nil {
				// If we can't get info, still show the topic but without partition info
				items[i] = NewTopicItem(topic, nil)
			} else {
				items[i] = NewTopicItem(topic, topicInfo)
			}
		}

		return ItemsUpdatedMsg{Items: items}
	}
}

// UpdateClusterListCmd returns a command that updates the cluster list
func UpdateClusterListCmd(clusters []config.KafkaClusterConfig) Command {
	return func() tea.Msg {
		items := make([]list.Item, len(clusters))
		for i, cluster := range clusters {
			items[i] = NewClusterItem(cluster.Name, cluster.Bootstrap)
		}

		return ItemsUpdatedMsg{Items: items}
	}
}
