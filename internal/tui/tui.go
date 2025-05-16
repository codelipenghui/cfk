// Package tui provides the terminal user interface for the cfk application
package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cfk-dev/cfk/internal/config"
	"github.com/cfk-dev/cfk/internal/core"
	"github.com/cfk-dev/cfk/internal/kafka"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the TUI application model
type Model struct {
	config       *config.AppConfig
	app          *core.App
	kafkaClient  *kafka.Client
	state        string
	clusterList  list.Model
	topicList    list.Model
	topicTable   table.Model
	viewport     viewport.Model
	clusterForm  ClusterForm
	topicForm    TopicForm
	err          error
	selectedItem string
	selectedCluster string
	width        int
	height       int
}

// Initialize the TUI model
func NewModel(cfg *config.AppConfig, app *core.App) Model {
	// Default width and height
	width := 80
	height := 24

	// Create cluster list
	clusterList := list.New([]list.Item{}, NewCustomDelegate(), 0, 0)
	clusterList.Title = "Kafka Clusters"

	// Create topic list
	topicList := list.New([]list.Item{}, NewCustomDelegate(), 0, 0)
	topicList.Title = "Topics"

	// Create topic table
	columns := []table.Column{
		{Title: "Topic", Width: 30},
		{Title: "Partitions", Width: 10},
		{Title: "Replicas", Width: 10},
	}
	rows := []table.Row{}
	topicTable := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Create viewport for message viewing
	viewport := viewport.New(80, 20)
	viewport.Style = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

	// Create cluster form
	clusterForm := NewClusterForm(width, height, nil)

	// Create topic form
	topicForm := NewTopicForm(width, height, "", 0)

	return Model{
		config:      cfg,
		app:         app,
		state:       "clusters",
		clusterList: clusterList,
		topicList:   topicList,
		topicTable:  topicTable,
		viewport:    viewport,
		clusterForm: clusterForm,
		topicForm:   topicForm,
		width:       width,
		height:      height,
	}
}

// Init initializes the TUI model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles TUI events and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update window size
		m.width = msg.Width
		m.height = msg.Height

		// Update components with new size
		m.clusterForm = NewClusterForm(m.width, m.height, nil)
		m.topicForm = NewTopicForm(m.width, m.height, "", 0)

		// Update list heights
		h := m.height - 6 // Adjust for header and footer
		m.clusterList.SetHeight(h)
		m.topicList.SetHeight(h)

		// Update viewport
		m.viewport.Width = m.width - 4
		m.viewport.Height = m.height - 8

		return m, nil

	case tea.KeyMsg:
		// Debug log to file
		f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		fmt.Fprintf(f, "Key pressed: '%s' in state: %s\n", msg.String(), m.state)

		// Handle global key events
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state != "add_cluster" && m.state != "edit_cluster" &&
			   m.state != "add_topic" && m.state != "edit_topic" {
				return m, tea.Quit
			}
		case "enter":
			if m.state == "clusters" {
				// When a cluster is selected, connect to it and switch to topics view
				if i, ok := m.clusterList.SelectedItem().(Item); ok {
					m.selectedItem = i.Title()
					m.selectedCluster = i.Title()
					// Connect to the selected cluster
					return m, func() tea.Msg {
						// Debug log to file
						f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						defer f.Close()
						fmt.Fprintf(f, "Connecting to cluster: %s\n", m.selectedItem)

						if err := m.app.ConnectToCluster(m.selectedItem); err != nil {
							fmt.Fprintf(f, "Error connecting: %v\n", err)
							return ErrorMsg{err}
						}

						fmt.Fprintf(f, "Connected successfully, setting state to topics\n")
						m.state = "topics"

						// Update the topic list title
						m.topicList.Title = "Topics in " + m.selectedCluster
						fmt.Fprintf(f, "Updated title to: %s\n", m.topicList.Title)

						// After connecting, update the topic list
						return UpdateTopicListCmd(m.app)()
					}
				}
			} else if m.state == "topics" {
				// When a topic is selected, show topic details
				if i, ok := m.topicList.SelectedItem().(Item); ok {
					m.selectedItem = i.Title()
					// Get topic details
					return m, func() tea.Msg {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						topicInfo, err := m.app.GetTopicInfo(ctx, m.selectedItem)
						if err != nil {
							return ErrorMsg{err}
						}

						// Update the topic table with the details
						rows := []table.Row{
							{m.selectedItem, fmt.Sprintf("%d", topicInfo.Partitions), "1"},
						}
						m.topicTable.SetRows(rows)

						m.state = "topic_details"
						return nil
					}
				}
			}
		case "backspace", "esc":
			// Debug log to file
			f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			defer f.Close()
			fmt.Fprintf(f, "Processing backspace/esc in state: %s\n", m.state)

			// Go back to the previous view
			if m.state == "topics" {
				fmt.Fprintf(f, "Changing state from topics to clusters\n")
				m.state = "clusters"
				return m, nil
			} else if m.state == "topic_details" {
				fmt.Fprintf(f, "Changing state from topic_details to topics\n")
				m.state = "topics"
				return m, nil
			} else if m.state == "messages" {
				fmt.Fprintf(f, "Changing state from messages to topic_details\n")
				m.state = "topic_details"
				return m, nil
			}
		case "b":
			// Debug log to file
			f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			defer f.Close()
			fmt.Fprintf(f, "Processing 'b' key in state: %s\n", m.state)

			// Go directly back to clusters view from any view
			if m.state == "topics" || m.state == "topic_details" || m.state == "messages" {
				fmt.Fprintf(f, "Changing state to clusters from %s\n", m.state)
				m.state = "clusters"
				return m, nil
			}
		case "a":
			// Add a new cluster
			if m.state == "clusters" {
				m.state = "add_cluster"
				m.clusterForm = NewClusterForm(m.width, m.height, nil)
				return m, m.clusterForm.Init()
			}
		case "n":
			// Add a new topic
			// Debug log to file
			f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			defer f.Close()
			fmt.Fprintf(f, "'n' key pressed in state: %s\n", m.state)

			if m.state == "topics" {
				fmt.Fprintf(f, "Creating new topic form\n")
				m.state = "add_topic"
				m.topicForm = NewTopicForm(m.width, m.height, "", 0)
				cmd := m.topicForm.Init()
				fmt.Fprintf(f, "Topic form initialized\n")
				return m, cmd
			}
		case "d":
			// Delete the selected cluster or topic
			if m.state == "clusters" {
				if i, ok := m.clusterList.SelectedItem().(Item); ok {
					clusterName := i.Title()
					return m, func() tea.Msg {
						if err := m.app.RemoveCluster(clusterName); err != nil {
							return ErrorMsg{err}
						}
						// Update the cluster list
						return UpdateClusterListCmd(m.config.Clusters)()
					}
				}
			} else if m.state == "topics" {
				if i, ok := m.topicList.SelectedItem().(Item); ok {
					topicName := i.Title()
					return m, func() tea.Msg {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						if err := m.app.DeleteTopic(ctx, topicName); err != nil {
							return ErrorMsg{err}
						}

						// Update the topic list
						return UpdateTopicListCmd(m.app)()
					}
				}
			}
		case "e":
			// Debug log to file
			f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			defer f.Close()
			fmt.Fprintf(f, "'e' key pressed in state: %s\n", m.state)

			// Edit the selected cluster or topic
			if m.state == "clusters" {
				// Edit cluster
				fmt.Fprintf(f, "Editing cluster\n")
				if i, ok := m.clusterList.SelectedItem().(Item); ok {
					clusterName := i.Title()
					fmt.Fprintf(f, "Selected cluster: %s\n", clusterName)

					// Find the cluster config
					var clusterConfig *config.KafkaClusterConfig
					for _, c := range m.config.Clusters {
						if c.Name == clusterName {
							clusterConfig = &c
							break
						}
					}

					if clusterConfig != nil {
						fmt.Fprintf(f, "Creating cluster form and setting state to edit_cluster\n")
						// Create the form and set the state
						m.clusterForm = NewClusterForm(m.width, m.height, clusterConfig)
						m.state = "edit_cluster"
						return m, m.clusterForm.Init()
					} else {
						fmt.Fprintf(f, "Cluster config not found\n")
					}
				} else {
					fmt.Fprintf(f, "Could not get selected cluster item\n")
				}
			} else if m.state == "topics" {
				// Edit topic
				fmt.Fprintf(f, "Editing topic\n")
				if i, ok := m.topicList.SelectedItem().(Item); ok {
					topicName := i.Title()
					fmt.Fprintf(f, "Selected topic: %s\n", topicName)

					// IMPORTANT: Set the state to edit_topic BEFORE getting topic info
					// This prevents the ItemsUpdatedMsg handler from changing it back
					fmt.Fprintf(f, "Setting state to edit_topic\n")
					m.state = "edit_topic"

					// Get topic info for editing
					return m, func() tea.Msg {
						f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						defer f.Close()
						fmt.Fprintf(f, "Getting topic info for %s\n", topicName)

						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						topicInfo, err := m.app.GetTopicInfo(ctx, topicName)
						if err != nil {
							fmt.Fprintf(f, "Error getting topic info: %v\n", err)
							return ErrorMsg{err}
						}

						fmt.Fprintf(f, "Got topic info, partitions: %d\n", topicInfo.Partitions)
						// Create the form
						m.topicForm = NewTopicForm(m.width, m.height, topicName, topicInfo.Partitions)
						fmt.Fprintf(f, "Created topic form\n")

						// Double-check that the state is still edit_topic
						if m.state != "edit_topic" {
							fmt.Fprintf(f, "WARNING: State changed from edit_topic to %s, forcing back to edit_topic\n", m.state)
							m.state = "edit_topic"
						}

						// Make sure the topic name is set in the form
						m.topicForm.topicName = topicName
						m.topicForm.isEdit = true
						fmt.Fprintf(f, "Explicitly set topicName='%s' and isEdit=true\n", topicName)

						// Initialize the form
						cmd := m.topicForm.Init()()
						fmt.Fprintf(f, "Initialized topic form\n")
						return cmd
					}
				} else {
					fmt.Fprintf(f, "Could not get selected topic item\n")
				}
			} else {
				fmt.Fprintf(f, "Unhandled state for 'e' key: %s\n", m.state)
			}
		}
	case ItemsUpdatedMsg:
		// Debug log to file
		f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		fmt.Fprintf(f, "ItemsUpdatedMsg received with state: %s, items count: %d\n", m.state, len(msg.Items))

		// Check if these are topics (they'll have names like _schemas or __consumer_offsets)
		isTopic := false
		if len(msg.Items) > 0 {
			if i, ok := msg.Items[0].(Item); ok {
				title := i.Title()
				if title == "_schemas" || title == "__consumer_offsets" {
					isTopic = true
					fmt.Fprintf(f, "Detected topic items based on names\n")
				}
			}
		}

		// Update the list with the new items
		if m.state == "topics" || (isTopic && m.state != "edit_topic" && m.state != "add_topic") {
			fmt.Fprintf(f, "Setting items for topic list\n")
			// Only set the state if we're not in a form
			if m.state != "edit_topic" && m.state != "add_topic" {
				m.state = "topics"
			}
			m.topicList.Title = "Topics in " + m.selectedCluster // Ensure title is set
			m.topicList.SetItems(msg.Items)
		} else if m.state == "clusters" {
			fmt.Fprintf(f, "Setting items for cluster list\n")
			m.clusterList.SetItems(msg.Items)
		} else {
			fmt.Fprintf(f, "Unknown state: %s\n", m.state)
		}
		return m, nil
	case ErrorMsg:
		// Handle errors
		m.err = msg.err
		return m, nil
	case ClusterAddedMsg:
		// Add the new cluster to the config
		if err := m.app.AddCluster(msg.Cluster); err != nil {
			m.err = err
			return m, nil
		}

		// Return to clusters view and update the list
		m.state = "clusters"
		return m, func() tea.Msg { return UpdateClusterListCmd(m.config.Clusters)() }
	case ClusterUpdatedMsg:
		// Update the cluster in the config
		if err := m.app.UpdateCluster(msg.Cluster); err != nil {
			m.err = err
			return m, nil
		}

		// Return to clusters view and update the list
		m.state = "clusters"
		return m, func() tea.Msg { return UpdateClusterListCmd(m.config.Clusters)() }
	case ClusterFormCancelledMsg:
		// Return to clusters view
		m.state = "clusters"
		return m, nil
	case TopicAddedMsg:
		// Add the new topic
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := m.app.CreateTopic(ctx, msg.Name, msg.Partitions, msg.ReplicationFactor); err != nil {
				return ErrorMsg{err}
			}

			// Return to topics view and update the list
			m.state = "topics"
			return UpdateTopicListCmd(m.app)()
		}
	case TopicUpdatedMsg:
		// Update the topic
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Currently we can only update partitions
			if err := m.app.UpdateTopicPartitions(ctx, msg.OldName, msg.Partitions); err != nil {
				return ErrorMsg{err}
			}

			// Return to topics view and update the list
			m.state = "topics"
			return UpdateTopicListCmd(m.app)()
		}
	case TopicFormCancelledMsg:
		// Return to topics view
		m.state = "topics"
		return m, nil
	}

	// Update components based on current state
	switch m.state {
	case "topics":
		m.topicList, cmd = m.topicList.Update(msg)
		return m, cmd
	case "topic_details":
		m.topicTable, cmd = m.topicTable.Update(msg)
		return m, cmd
	case "messages":
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	case "add_cluster", "edit_cluster":
		// Update the cluster form
		newForm, cmd := m.clusterForm.Update(msg)
		m.clusterForm = newForm
		return m, cmd
	case "add_topic", "edit_topic":
		// Debug log to file
		f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		fmt.Fprintf(f, "Updating topic form with message type: %T\n", msg)

		// Update the topic form
		newForm, cmd := m.topicForm.Update(msg)
		m.topicForm = newForm
		return m, cmd
	default: // clusters
		m.clusterList, cmd = m.clusterList.Update(msg)
		return m, cmd
	}
}

// View renders the TUI
func (m Model) View() string {
	// Debug log to file
	f, _ := os.OpenFile("/tmp/cfk_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	fmt.Fprintf(f, "View called with state: %s\n", m.state)

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err)
	}



	fmt.Fprintf(f, "View switch statement with state: %s\n", m.state)
	switch m.state {
	case "topics":
		helpText := "\nPress 'n' to add new topic, 'e' to edit, 'd' to delete, 'enter' to view details, 'b' or 'esc' to go back to clusters, 'q' to quit"
		return fmt.Sprintf("Connected to cluster: %s\n\n%s\n%s",
			m.selectedCluster, m.topicList.View(), helpText)
	case "topic_details":
		return m.topicTable.View() + "\n\nPress 'esc' to go back to topics, 'b' to go back to clusters, 'q' to quit"
	case "messages":
		return m.viewport.View() + "\n\nPress 'esc' to go back, 'q' to quit"
	case "add_cluster", "edit_cluster":
		return m.clusterForm.View()
	case "add_topic", "edit_topic":
		fmt.Fprintf(f, "Rendering %s form\n", m.state)
		// Just return the form view
		return m.topicForm.View()
	default: // clusters
		helpText := "\nPress 'a' to add, 'e' to edit, 'd' to delete, 'enter' to connect, 'q' to quit"
		if m.clusterList.Items() == nil || len(m.clusterList.Items()) == 0 {
			return fmt.Sprintf("%s\n\nNo clusters configured. %s", m.clusterList.View(), helpText)
		}
		return fmt.Sprintf("%s\n%s", m.clusterList.View(), helpText)
	}
}

// Start starts the TUI application
func Start(cfg *config.AppConfig, app *core.App) error {
	model := NewModel(cfg, app)

	// Initialize the cluster list
	initialCmd := UpdateClusterListCmd(cfg.Clusters)

	p := tea.NewProgram(model, tea.WithAltScreen())

	// Send the initial command to update the cluster list
	go func() {
		p.Send(initialCmd())
	}()

	_, err := p.Run()
	return err
}
