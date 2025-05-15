// Package tui provides the terminal user interface for the cfk application
package tui

import (
	"fmt"

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
	err          error
	selectedItem string
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

	return Model{
		config:      cfg,
		app:         app,
		state:       "clusters",
		clusterList: clusterList,
		topicList:   topicList,
		topicTable:  topicTable,
		viewport:    viewport,
		clusterForm: clusterForm,
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
		
		// Update list heights
		h := m.height - 6 // Adjust for header and footer
		m.clusterList.SetHeight(h)
		m.topicList.SetHeight(h)
		
		// Update viewport
		m.viewport.Width = m.width - 4
		m.viewport.Height = m.height - 8
		
		return m, nil
		
	case tea.KeyMsg:
		// Handle global key events
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state != "add_cluster" && m.state != "edit_cluster" {
				return m, tea.Quit
			}
		case "enter":
			if m.state == "clusters" {
				// When a cluster is selected, connect to it and switch to topics view
				if i, ok := m.clusterList.SelectedItem().(Item); ok {
					m.selectedItem = i.Title()
					// Connect to the selected cluster
					return m, func() tea.Msg {
						if err := m.app.ConnectToCluster(m.selectedItem); err != nil {
							return ErrorMsg{err}
						}
						m.state = "topics"
						// After connecting, update the topic list
						return UpdateTopicListCmd(m.app)()
					}
				}
			}
		case "backspace", "esc":
			// Go back to the previous view
			if m.state == "topics" {
				m.state = "clusters"
				return m, nil
			} else if m.state == "topic_details" {
				m.state = "topics"
				return m, nil
			} else if m.state == "messages" {
				m.state = "topic_details"
				return m, nil
			}
		case "a":
			// Add a new cluster or topic depending on the current state
			if m.state == "clusters" {
				m.state = "add_cluster"
				m.clusterForm = NewClusterForm(m.width, m.height, nil)
				return m, m.clusterForm.Init()
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
			}
		case "e":
			// Edit the selected cluster or topic
			if m.state == "clusters" {
				if i, ok := m.clusterList.SelectedItem().(Item); ok {
					clusterName := i.Title()
					// Find the cluster config
					var clusterConfig *config.KafkaClusterConfig
					for _, c := range m.config.Clusters {
						if c.Name == clusterName {
							clusterConfig = &c
							break
						}
					}
					
					if clusterConfig != nil {
						m.state = "edit_cluster"
						m.clusterForm = NewClusterForm(m.width, m.height, clusterConfig)
						return m, m.clusterForm.Init()
					}
				}
			}
		}
	case ItemsUpdatedMsg:
		// Update the list with the new items
		if m.state == "topics" {
			m.topicList.SetItems(msg.Items)
		} else if m.state == "clusters" {
			m.clusterList.SetItems(msg.Items)
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
	default: // clusters
		m.clusterList, cmd = m.clusterList.Update(msg)
		return m, cmd
	}
}

// View renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err)
	}

	switch m.state {
	case "topics":
		return fmt.Sprintf("Connected to: %s\n\n%s\n\nPress 'esc' to go back, 'q' to quit", 
			m.selectedItem, m.topicList.View())
	case "topic_details":
		return m.topicTable.View() + "\n\nPress 'esc' to go back, 'q' to quit"
	case "messages":
		return m.viewport.View() + "\n\nPress 'esc' to go back, 'q' to quit"
	case "add_cluster", "edit_cluster":
		return m.clusterForm.View()
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
