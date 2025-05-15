package tui

import (
	"fmt"

	"github.com/cfk-dev/cfk/internal/kafka"
)

// Item represents an item in a list
type Item struct {
	itemTitle       string
	itemDescription string
	Data            interface{}
}

// FilterValue implements the list.Item interface
func (i Item) FilterValue() string {
	return i.itemTitle
}

// Title returns the item's title
func (i Item) Title() string {
	return i.itemTitle
}

// Description returns the item's description
func (i Item) Description() string {
	return i.itemDescription
}

// NewTopicItem creates a new item for a topic
func NewTopicItem(topicName string, info *kafka.TopicInfo) Item {
	description := "Unknown details"
	if info != nil {
		description = fmt.Sprintf("%d partitions", info.Partitions)
	}

	return Item{
		itemTitle:       topicName,
		itemDescription: description,
		Data:            info,
	}
}

// NewClusterItem creates a new item for a cluster
func NewClusterItem(clusterName string, bootstrapServers []string) Item {
	description := "No bootstrap servers"
	if len(bootstrapServers) > 0 {
		description = bootstrapServers[0]
		if len(bootstrapServers) > 1 {
			description += fmt.Sprintf(" (+%d more)", len(bootstrapServers)-1)
		}
	}

	return Item{
		itemTitle:       clusterName,
		itemDescription: description,
	}
}
